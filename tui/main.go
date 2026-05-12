package main

import (
	"fmt"
	"os"

	"crona/shared/config"
	sharedposthog "crona/shared/posthog"
	versionpkg "crona/shared/version"
	"crona/tui/internal/api"
	"crona/tui/internal/kernel"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui"
	"crona/tui/internal/tui/terminaltitle"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	appEnv := config.Load()
	var telemetry sharedposthog.Client
	info, err := kernel.Ensure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start local engine: %v\n", err)
		logger.Errorf("Local engine start failed: %v", err)
		os.Exit(1)
	}

	settingsClient := api.NewClient(info.Transport, kernel.Endpoint(info), info.ScratchDir)
	telemetryConfig := sharedposthog.LoadConfig("tui")
	telemetryConfig.Version = versionpkg.Current()
	telemetryConfig.Mode = appEnv.Mode
	telemetryConfig.UsageEnabled = false
	telemetryConfig.ErrorReportingEnabled = false
	if settings, settingsErr := settingsClient.GetSettings(); settingsErr != nil {
		logger.Errorf("Telemetry settings load failed: %v", settingsErr)
	} else {
		telemetryConfig.UsageEnabled = telemetryConfig.Enabled && telemetryUsageEnabled(settings)
		telemetryConfig.ErrorReportingEnabled = telemetryConfig.Enabled && telemetryErrorReportingEnabled(settings)
	}
	telemetry, err = sharedposthog.New(telemetryConfig)
	if err != nil {
		logger.Errorf("Telemetry init failed: %v", err)
	} else {
		defer func() {
			if closeErr := telemetry.Close(); closeErr != nil {
				logger.Errorf("Telemetry close failed: %v", closeErr)
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				_ = telemetry.ReportError("panic", panicError(r), sharedposthog.Properties{
					"entrypoint": "tui",
					"operation":  "main",
				})
				_ = telemetry.Flush()
				panic(r)
			}
		}()
		if captureErr := captureTUIStarted(telemetry); captureErr != nil {
			logger.Errorf("Telemetry capture failed: %v", captureErr)
		}
	}

	logger.Info("Crona TUI starting")
	logger.Infof("Connected to local engine at %s", kernel.EndpointLabel(info))
	_ = terminaltitle.Write(os.Stdout, "Crona")
	defer func() { _ = terminaltitle.Reset(os.Stdout) }()

	done := make(chan struct{})
	eventStream := api.Subscribe(info.Transport, kernel.Endpoint(info), done)
	tui.SetEventChannel(eventStream)

	executablePath, _ := os.Executable()
	if err := kernel.WriteTUIRuntimeState(executablePath); err != nil {
		logger.Errorf("WriteTUIRuntimeState failed: %v", err)
	}

	model := tui.New(info.Transport, kernel.Endpoint(info), info.ScratchDir, info.Env, executablePath, done, telemetry)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		if telemetry != nil {
			_ = telemetry.ReportError("handled", err, sharedposthog.Properties{
				"entrypoint": "tui",
				"operation":  "program_run",
			})
			_ = telemetry.Flush()
		}
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		logger.Errorf("TUI exited with error: %v", err)
		os.Exit(1)
	}

	logger.Info("Crona TUI exited")
}
