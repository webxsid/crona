package main

import (
	"fmt"
	"os"

	"crona/tui/internal/api"
	"crona/tui/internal/kernel"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logger.Info("Crona TUI starting")

	info, err := kernel.Ensure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start kernel: %v\n", err)
		logger.Errorf("Kernel start failed: %v", err)
		os.Exit(1)
	}

	logger.Infof("Connected to kernel at %s", info.BaseURL)

	// Start SSE subscription — runs in its own goroutine.
	// Events are forwarded into the Bubbletea program via p.Send().
	done := make(chan struct{})
	sseEvents := api.Subscribe(info.BaseURL, info.Token, done)
	tui.SetSSEChannel(sseEvents)

	model := tui.New(info.BaseURL, info.Token, info.ScratchDir, done)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		logger.Errorf("TUI exited with error: %v", err)
		os.Exit(1)
	}

	logger.Info("Crona TUI exited")
}
