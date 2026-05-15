package controller

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

func TestEditIssueDueDateF2FallbackOpensCalendar(t *testing.T) {
	state := OpenEditIssue(State{}, 12, 34, "Fix alerts", nil, nil, nil)
	state.FocusIdx = 3
	state.Inputs[2].SetValue("2026-04-18")
	state = SyncDialogFocus(state)

	next, action, status := updateEditIssue(state, "2026-04-20", tea.KeyMsg{Type: tea.KeyCtrlE})
	if action != nil || status != "" {
		t.Fatalf("expected no action/status, got action=%+v status=%q", action, status)
	}
	if next.Kind != "pick_date" {
		t.Fatalf("expected pick_date dialog, got %q", next.Kind)
	}
	if next.Parent != "edit_issue" {
		t.Fatalf("expected edit_issue parent, got %q", next.Parent)
	}
	if next.DateCursorValue != "2026-04-18" {
		t.Fatalf("expected calendar to open on existing due date, got %q", next.DateCursorValue)
	}
}

func TestCreateIssueMetaDueDateF2FallbackOpensCalendar(t *testing.T) {
	state := OpenCreateIssueMeta(State{}, 34, "main", "Crona")
	state.FocusIdx = 3
	state.Inputs[2].SetValue("2026-04-19")
	state = SyncDialogFocus(state)

	next, action, status := updateCreateIssueMeta(state, "2026-04-20", tea.KeyMsg{Type: tea.KeyCtrlE})
	if action != nil || status != "" {
		t.Fatalf("expected no action/status, got action=%+v status=%q", action, status)
	}
	if next.Kind != "pick_date" {
		t.Fatalf("expected pick_date dialog, got %q", next.Kind)
	}
	if next.Parent != "create_issue_meta" {
		t.Fatalf("expected create_issue_meta parent, got %q", next.Parent)
	}
	if next.DateCursorValue != "2026-04-19" {
		t.Fatalf("expected calendar to open on existing due date, got %q", next.DateCursorValue)
	}
}

func TestCreateIssueDefaultDueDateF2FallbackOpensCalendar(t *testing.T) {
	state := OpenCreateIssueDefault(State{})
	state.FocusIdx = 5
	state.Inputs[4].SetValue("2026-04-21")
	state = SyncDialogFocus(state)

	next, action, status := updateCreateIssueDefault(state, UpdateContext{}, "2026-04-20", tea.KeyMsg{Type: tea.KeyCtrlE})
	if action != nil || status != "" {
		t.Fatalf("expected no action/status, got action=%+v status=%q", action, status)
	}
	if next.Kind != "pick_date" {
		t.Fatalf("expected pick_date dialog, got %q", next.Kind)
	}
	if next.Parent != "create_issue_default" {
		t.Fatalf("expected create_issue_default parent, got %q", next.Parent)
	}
	if next.DateCursorValue != "2026-04-21" {
		t.Fatalf("expected calendar to open on existing due date, got %q", next.DateCursorValue)
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
