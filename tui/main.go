package main

import (
	"fmt"
	"os"

	"crona/shared/config"
	"crona/tui/internal/api"
	"crona/tui/internal/kernel"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui"
	"crona/tui/internal/tui/terminaltitle"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	_ = config.Load()

	info, err := kernel.Ensure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start local engine: %v\n", err)
		logger.Errorf("Local engine start failed: %v", err)
		os.Exit(1)
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

	model := tui.New(info.Transport, kernel.Endpoint(info), info.ScratchDir, info.Env, executablePath, done)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		logger.Errorf("TUI exited with error: %v", err)
		os.Exit(1)
	}

	logger.Info("Crona TUI exited")
}
