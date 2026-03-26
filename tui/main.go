package main

import (
	"fmt"
	"os"

	"crona/shared/config"
	"crona/tui/internal/api"
	"crona/tui/internal/kernel"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	_ = config.Load()

	logger.Info("Crona TUI starting")

	info, err := kernel.Ensure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start kernel: %v\n", err)
		logger.Errorf("Kernel start failed: %v", err)
		os.Exit(1)
	}

	logger.Infof("Connected to kernel at %s", kernel.EndpointLabel(info))

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
