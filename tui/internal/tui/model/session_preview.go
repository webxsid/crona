package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"crona/tui/internal/tui/commands"
)

const sessionPreviewDelay = 150 * time.Millisecond

// selectedSessionID is the ID of the session Log's Sessions list has selected, or "".
func (m Model) selectedSessionID() string {
	if m.view != ViewSessionHistory {
		return ""
	}
	if entry, ok := m.selectedSessionHistoryEntry(); ok {
		return entry.ID
	}
	return ""
}

// sessionPreviewTick schedules a preview fetch when the selection moved to a new session,
// or the detail pane just appeared, on a screen wide enough to show it.
func (m Model) sessionPreviewTick(before string, wasWide bool) tea.Cmd {
	id := m.selectedSessionID()
	if id == "" || m.width < 100 || (id == before && wasWide) {
		return nil
	}
	return tea.Tick(sessionPreviewDelay, func(time.Time) tea.Msg { return commands.SessionPreviewTickMsg{ID: id} })
}

func (m Model) handleSessionPreview(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case commands.SessionPreviewTickMsg:
		if msg.ID == "" || msg.ID != m.selectedSessionID() {
			return m, nil, true
		}
		return m, commands.LoadSessionPreview(m.client, msg.ID), true
	case commands.SessionPreviewLoadedMsg:
		if msg.Detail != nil && msg.Detail.ID == m.selectedSessionID() {
			m.sessionPreview = msg.Detail
			m.sessionPreviewFailed = ""
		}
		return m, nil, true
	case commands.SessionPreviewFailedMsg:
		if msg.ID == m.selectedSessionID() {
			m.sessionPreviewFailed = msg.ID
		}
		return m, nil, true
	}
	return m, nil, false
}
