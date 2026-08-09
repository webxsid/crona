package controller

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestEditIssueDueDateGSetsToday(t *testing.T) {
	state := OpenEditIssue(State{}, 12, 34, "Fix alerts", nil, nil, nil)
	state.FocusIdx = 3
	state = SyncDialogFocus(state)

	next, action, status := updateEditIssue(
		state,
		"2026-04-16",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}},
	)
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

	next, action, status := updateCreateIssueMeta(
		state,
		"2026-04-20",
		tea.KeyMsg{Type: tea.KeyCtrlE},
	)
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

	next, action, status := updateCreateIssueDefault(
		state,
		UpdateContext{},
		"2026-04-20",
		tea.KeyMsg{Type: tea.KeyCtrlE},
	)
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

func TestValidateDueDateRejectsPastAndConfiguredRestDates(t *testing.T) {
	state := State{
		DueDateToday:        "2026-04-20",
		DueDateRestWeekdays: []int{6},
		DueDateRestDates:    []string{"2026-04-23"},
	}
	if err := ValidateDueDate(state, "2026-04-19"); err == nil || err.Error() != "due dates cannot be in the past" {
		t.Fatalf("expected past-date validation error, got %v", err)
	}
	if err := ValidateDueDate(state, "2026-04-23"); err == nil || err.Error() != "due dates cannot fall on configured rest days" {
		t.Fatalf("expected explicit rest-date validation error, got %v", err)
	}
	if err := ValidateDueDate(state, "2026-04-25"); err == nil || err.Error() != "due dates cannot fall on configured rest days" {
		t.Fatalf("expected weekday rest-date validation error, got %v", err)
	}
	if err := ValidateDueDate(state, "2026-04-24"); err != nil {
		t.Fatalf("expected valid future due date, got %v", err)
	}
}

func TestDueDatePickerRejectsInvalidSelection(t *testing.T) {
	state := State{
		Kind:            "pick_date",
		Parent:          "edit_issue",
		DateCursorValue: "2026-04-19",
		DueDateToday:    "2026-04-20",
	}
	next, action, status := updateDatePicker(state, UpdateContext{}, "2026-04-20", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || status != "" {
		t.Fatalf("expected invalid selection to emit no action/status, got action=%+v status=%q", action, status)
	}
	if next.ErrorMessage != "due dates cannot be in the past" {
		t.Fatalf("expected picker validation error, got %q", next.ErrorMessage)
	}
}

func TestDirectIssueDueDatePickerRejectsInvalidSelection(t *testing.T) {
	state := State{
		Kind:            "pick_date",
		IssueID:         42,
		DateCursorValue: "2026-04-19",
		DueDateToday:    "2026-04-20",
	}
	next, action, status := updateDatePicker(state, UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || status != "" {
		t.Fatalf("expected invalid direct selection to emit no action/status, got action=%+v status=%q", action, status)
	}
	if next.ErrorMessage != "due dates cannot be in the past" {
		t.Fatalf("expected direct picker validation error, got %q", next.ErrorMessage)
	}
}

func TestDueDatePolicyDoesNotApplyToOtherCalendars(t *testing.T) {
	for _, parent := range []string{"rollup_start", "rollup_end", "edit_rest_protection"} {
		state := State{Parent: parent, DueDateToday: "2026-04-20"}
		if isIssueDueDatePicker(state) {
			t.Fatalf("expected %q calendar to remain outside due-date policy", parent)
		}
	}
}

func TestDueDateCalendarStylesOnlyIssuePolicyDates(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(previousProfile)
	theme := Theme{
		ColorRed:    lipgloss.Color("1"),
		ColorDim:    lipgloss.Color("8"),
		StyleNormal: lipgloss.NewStyle(),
		StyleDim:    lipgloss.NewStyle(),
		StyleCursor: lipgloss.NewStyle(),
	}
	issue := State{
		IssueID:          42,
		DueDateToday:     "2026-04-20",
		DueDateRestDates: []string{"2026-04-21"},
		DateCursorValue:  "2026-04-20",
		DateMonthValue:   "2026-04-01",
	}
	issueGrid := PopulateDatePresentation(theme, issue, "2026-04-20").DateGrid
	redCell := lipgloss.NewStyle().Foreground(theme.ColorRed).Render(" 21 ")
	if !strings.Contains(issueGrid, redCell) {
		t.Fatalf("expected issue rest date to use red style, got %q", issueGrid)
	}
	rollup := issue
	rollup.IssueID = 0
	rollup.Parent = "rollup_start"
	rollupGrid := PopulateDatePresentation(theme, rollup, "2026-04-20").DateGrid
	if strings.Contains(rollupGrid, redCell) {
		t.Fatalf("expected rollup calendar not to use due-date rest styling, got %q", rollupGrid)
	}
}
