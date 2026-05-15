package model

import (
	"strings"
	"time"

	"crona/tui/internal/logger"
	commands "crona/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	statusInfoDuration  = 3 * time.Second
	statusErrorDuration = 5 * time.Second
)

func (m *Model) withStatus(message string, isError bool) Model {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	logStatusError(message, isError)
	return *m
}

func (m *Model) setStatus(message string, isError bool) tea.Cmd {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	logStatusError(message, isError)
	duration := statusInfoDuration
	if isError {
		duration = statusErrorDuration
	}
	return commands.ClearStatusAfter(m.statusSeq, duration)
}

func logStatusError(message string, isError bool) {
	if !isError {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" || strings.HasPrefix(message, "Error: ") {
		return
	}
	logger.Error("UI error: " + message)
}
