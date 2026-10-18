package model

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
)

func previewModel(width int) Model {
	return Model{
		view: ViewSessionHistory, pane: PaneSessions, width: width, height: 40,
		cursor: map[Pane]int{PaneSessions: 0}, filters: map[Pane]string{PaneSessions: ""},
		sessionHistory: []api.SessionHistoryEntry{
			{Session: api.Session{ID: "a", StartTime: "2026-10-17T10:00:00Z"}},
			{Session: api.Session{ID: "b", StartTime: "2026-10-16T10:00:00Z"}},
		},
	}
}

// messages runs cmd and flattens batches into the messages they yield.
func messages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		out := []tea.Msg{}
		for _, inner := range batch {
			out = append(out, messages(inner)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

func hasTick(cmd tea.Cmd, id string) bool {
	for _, msg := range messages(cmd) {
		if tick, ok := msg.(commands.SessionPreviewTickMsg); ok && tick.ID == id {
			return true
		}
	}
	return false
}

func TestSessionPreviewDebouncesSelection(t *testing.T) {
	_, cmd := previewModel(160).Update(tea.KeyMsg{Type: tea.KeyDown})
	if !hasTick(cmd, "b") {
		t.Fatalf("expected a preview tick for the newly selected session")
	}
	_, cmd = previewModel(60).Update(tea.KeyMsg{Type: tea.KeyDown})
	if hasTick(cmd, "b") {
		t.Fatalf("expected no preview tick on a narrow screen")
	}
}

func TestSessionPreviewTickFetchesOnlyIfStillSelected(t *testing.T) {
	m := previewModel(160)
	if _, cmd := m.update(commands.SessionPreviewTickMsg{ID: "a"}); cmd == nil {
		t.Fatalf("expected a fetch for the still-selected session")
	}
	if _, cmd := m.update(commands.SessionPreviewTickMsg{ID: "b"}); cmd != nil {
		t.Fatalf("expected no fetch once the selection moved on")
	}
}

func TestSessionPreviewIgnoresStale(t *testing.T) {
	m := previewModel(160)
	m.sessionPreviewFailed = "a"
	next, _ := m.update(commands.SessionPreviewLoadedMsg{Detail: &api.SessionDetail{SessionHistoryEntry: api.SessionHistoryEntry{Session: api.Session{ID: "b"}}}})
	if next.(Model).sessionPreview != nil {
		t.Fatalf("expected a preview for another session to be ignored")
	}
	next, _ = m.update(commands.SessionPreviewLoadedMsg{Detail: &api.SessionDetail{SessionHistoryEntry: api.SessionHistoryEntry{Session: api.Session{ID: "a"}}}})
	if got := next.(Model); got.sessionPreview == nil || got.sessionPreviewFailed != "" {
		t.Fatalf("expected the selected preview stored and the failure cleared")
	}
}

func TestSessionPreviewFailureFlags(t *testing.T) {
	next, _ := previewModel(160).update(commands.SessionPreviewFailedMsg{ID: "a", Err: errors.New("boom")})
	m := next.(Model)
	if !m.viewContentState(m.width, m.height, m.selectionSnapshot(), nil).SessionPreviewFailed {
		t.Fatalf("expected the failure passed to the view")
	}
	next, _ = previewModel(160).update(commands.SessionPreviewFailedMsg{ID: "b", Err: errors.New("boom")})
	m = next.(Model)
	if m.viewContentState(m.width, m.height, m.selectionSnapshot(), nil).SessionPreviewFailed {
		t.Fatalf("expected another session's failure ignored")
	}
}

func TestSessionDialogPathUntouched(t *testing.T) {
	next, _ := previewModel(160).update(commands.SessionPreviewLoadedMsg{Detail: &api.SessionDetail{SessionHistoryEntry: api.SessionHistoryEntry{Session: api.Session{ID: "a"}}}})
	if next.(Model).sessionDetailOpen || next.(Model).sessionDetail != nil {
		t.Fatalf("expected the preview kept apart from the details dialog")
	}
	next, cmd := previewModel(160).update(tea.KeyMsg{Type: tea.KeyEnter})
	if !next.(Model).sessionDetailOpen || cmd == nil {
		t.Fatalf("expected enter to still open the details dialog")
	}
}

func TestSessionPreviewLoadsAfterWidening(t *testing.T) {
	if previewModel(160).sessionPreviewTick("a", false) == nil {
		t.Fatalf("expected a preview tick once the detail pane appears")
	}
	if previewModel(160).sessionPreviewTick("a", true) != nil {
		t.Fatalf("expected no tick when nothing changed")
	}
}
