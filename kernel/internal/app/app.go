package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	"crona/kernel/internal/export"
	"crona/kernel/internal/ipc"
	"crona/kernel/internal/notify"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/store"
	"crona/kernel/internal/updatecheck"
	"crona/shared/config"
	"crona/shared/localipc"
	sharedposthog "crona/shared/posthog"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
	versionpkg "crona/shared/version"
)

func Run(ctx context.Context) (runErr error) {
	appEnv := config.Load()

	if err := runtime.MigrateLegacyBaseDir(appEnv.Mode); err != nil {
		return fmt.Errorf("migrate runtime dir: %w", err)
	}

	paths, err := runtime.ResolvePaths()
	if err != nil {
		return fmt.Errorf("resolve runtime paths: %w", err)
	}
	if err := runtime.EnsurePaths(paths); err != nil {
		return fmt.Errorf("ensure runtime paths: %w", err)
	}

	logger := runtime.NewLogger(paths)
	startedAt := time.Now().UTC().Format(time.RFC3339)

	dbStore, err := store.Open(paths.DBPath)
	if err != nil {
		return fmt.Errorf("open sqlite store: %w", err)
	}
	defer func() {
		if err := dbStore.Close(); err != nil {
			logger.Error("close sqlite store", err)
		}
	}()

	if err := dbStore.Ping(ctx); err != nil {
		return fmt.Errorf("ping sqlite store: %w", err)
	}
	if err := store.InitSchema(ctx, dbStore.DB()); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}

	bus := events.NewBus()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	registry := store.NewRegistry(dbStore.DB())

	commandCtx := core.NewContext(
		dbStore,
		registry,
		"local",
		hostnameOr("device-1"),
		paths.ScratchDir,
		func() string { return time.Now().UTC().Format(time.RFC3339) },
		bus,
	)
	if err := commandCtx.InitDefaults(runCtx); err != nil {
		return fmt.Errorf("init command defaults: %w", err)
	}
	telemetryConfig := sharedposthog.LoadConfig("kernel")
	telemetryConfig.Version = versionpkg.Current()
	telemetryConfig.Mode = appEnv.Mode
	telemetryConfig.RuntimeDir = paths.BaseDir
	telemetryConfig.UsageEnabled = false
	telemetryConfig.ErrorReportingEnabled = false
	if settings, settingsErr := commandCtx.CoreSettings.Get(runCtx, commandCtx.UserID); settingsErr != nil {
		logger.Error("load telemetry settings", settingsErr)
	} else {
		telemetryConfig.UsageEnabled = telemetryConfig.Enabled && telemetryUsageEnabled(settings)
		telemetryConfig.ErrorReportingEnabled = telemetryConfig.Enabled && telemetryErrorReportingEnabled(settings)
	}
	telemetry, err := sharedposthog.New(telemetryConfig)
	if err != nil {
		logger.Error("init telemetry", err)
	} else {
		defer func() {
			if closeErr := telemetry.Close(); closeErr != nil {
				logger.Error("close telemetry", closeErr)
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				_ = telemetry.ReportError("panic", panicError(r), sharedposthog.Properties{
					"entrypoint": "daemon",
					"operation":  "run",
				})
				_ = telemetry.Flush()
				panic(r)
			}
		}()
		defer func() {
			if runErr != nil {
				_ = telemetry.ReportError("handled", runErr, sharedposthog.Properties{
					"entrypoint": "daemon",
					"operation":  "run",
				})
			}
		}()
	}
	alerts := notify.Start(runCtx, commandCtx, bus, logger, paths)
	updater := updatecheck.Start(runCtx, commandCtx, bus, logger, paths, appEnv.Mode)
	if _, err := export.EnsureAssets(paths); err != nil {
		return fmt.Errorf("ensure export assets: %w", err)
	}

	info := sharedtypes.KernelInfo{
		PID:             os.Getpid(),
		Transport:       paths.Transport,
		Endpoint:        paths.Endpoint,
		SocketPath:      paths.SocketPath,
		ProtocolVersion: protocol.Version,
		Token:           "",
		StartedAt:       startedAt,
		ScratchDir:      paths.ScratchDir,
		Env:             appEnv.Mode,
		ExecutablePath:  executablePath(),
		RunningChannel:  versionpkg.RunningChannel(),
		RunningIsBeta:   versionpkg.IsBetaRelease(),
	}

	server := ipc.NewServer(paths.Transport, paths.Endpoint, NewHandler(startedAt, info, dbStore.Ping, commandCtx, bus, cancel, appEnv.Mode, paths, updater, alerts, telemetry), logger)
	timer := corecommands.GetTimerService(commandCtx)
	if err := timer.RecoverBoundary(runCtx); err != nil {
		return fmt.Errorf("recover timer boundary: %w", err)
	}
	if err := server.Start(); err != nil {
		return fmt.Errorf("start ipc server: %w", err)
	}

	if err := runtime.WriteKernelInfo(paths, info); err != nil {
		return fmt.Errorf("write kernel info: %w", err)
	}
	if telemetry != nil && telemetry.UsageEnabled() {
		if captureErr := captureDaemonStarted(telemetry, paths.Transport, info.RunningChannel); captureErr != nil {
			logger.Error("capture daemon started", captureErr)
		}
	}
	defer func() {
		if err := runtime.ClearKernelInfo(paths); err != nil {
			logger.Error("clear kernel info", err)
		}
	}()
	defer func() {
		if err := server.Close(); err != nil {
			logger.Error("close ipc server", err)
		}
	}()

	logger.Info("kernel listening on " + localipc.Label(paths.Transport, paths.Endpoint))

	<-runCtx.Done()
	if telemetry != nil && telemetry.UsageEnabled() {
		if captureErr := captureDaemonStopped(telemetry, paths.Transport, info.RunningChannel); captureErr != nil {
			logger.Error("capture daemon stopped", captureErr)
		}
	}
	logger.Info("kernel shutting down")
	return nil
}

func hostnameOr(fallback string) string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return fallback
	}
	return name
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return path
}
