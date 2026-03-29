package app

import (
	"time"

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
	return *m
}

func (m *Model) setStatus(message string, isError bool) tea.Cmd {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	duration := statusInfoDuration
	if isError {
		duration = statusErrorDuration
	}
	return commands.ClearStatusAfter(m.statusSeq, duration)
}
