package dialogs

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditIssueDueDateGSetsToday(t *testing.T) {
	state := OpenEditIssue(State{}, 12, 34, "Fix alerts", nil, nil, nil)
	state.FocusIdx = 3
	state = SyncDialogFocus(state)

	next, action, status := updateEditIssue(state, "2026-04-16", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if action != nil || status != "" {
		t.Fatalf("expected no action/status, got action=%+v status=%q", action, status)
	}
	if got := next.Inputs[2].Value(); got != "2026-04-16" {
		t.Fatalf("expected due date to be set to current date, got %q", got)
	}
}

func TestEditIssueDueDateHintIncludesTodayShortcut(t *testing.T) {
	state := OpenEditIssue(State{}, 12, 34, "Fix alerts", nil, nil, nil)
	state.FocusIdx = 3

	hint := issueDialogHint(state, "save")
	if !strings.Contains(hint, "[g] today") {
		t.Fatalf("expected edit issue due-date hint to include today shortcut, got %q", hint)
	}
}
