package controller

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestSummaryCalendarUsesScoreBoxesAndAwaySemantics(t *testing.T) {
	stateWidth := 135
	state := OpenSummaryCalendar(State{}, "2026-08-18")
	state.Width = stateWidth
	state.DueDateAwayDates = []string{"2026-08-17"}
	state.CalendarScores = map[string]api.FocusScoreRangeDay{
		"2026-08-18": {Date: "2026-08-18", Score: 82, HasData: true},
	}
	state = PopulateSummaryCalendarPresentation(Theme{}, state, "2026-08-18")
	if !strings.Contains(state.DateGrid, "82") || !strings.Contains(state.DateGrid, "Away") {
		t.Fatalf("expected score and away cells in calendar: %q", state.DateGrid)
	}
	if strings.Count(state.DateGrid, "╭") < 6 {
		t.Fatalf("expected boxed date cells, got: %q", state.DateGrid)
	}
	rows := strings.Split(state.DateGrid, "\n")
	if len(rows) < 25 || !strings.Contains(rows[2], "27") || !strings.Contains(rows[2], " 2") {
		t.Fatalf("expected seven date cells on the same rendered row, got: %q", state.DateGrid)
	}
}

func TestSummaryCalendarFollowsWeekStartAndShowsWeekNumbers(t *testing.T) {
	state := OpenSummaryCalendar(State{}, "2026-08-18")
	state.Width = 135
	state.DueDateWeekStart = sharedtypes.WeekStartSunday
	state = PopulateSummaryCalendarPresentation(Theme{}, state, "2026-08-18")
	rows := strings.Split(ansi.Strip(state.DateGrid), "\n")
	if fields := strings.Fields(rows[0]); fields[0] != "Su" {
		t.Fatalf("expected Sunday first, got %q", rows[0])
	}
	// August 2026 starts on a Saturday, so a Sunday grid opens on July 26.
	if !strings.Contains(rows[2], "W31") || !strings.Contains(rows[2], "26") ||
		strings.Index(rows[2], "26") > strings.Index(rows[2], "27") {
		t.Fatalf("expected the first week to start Sunday July 26 with its number, got %q", rows[2])
	}
}

func TestSummaryCalendarSelectsDate(t *testing.T) {
	state := OpenSummaryCalendar(State{}, "2026-08-18")
	state.DateCursorValue = "2026-08-17"
	next, action, _ := updateSummaryCalendar(state, "2026-08-18", tea.KeyMsg{Type: tea.KeyEnter})
	if next.Kind != "" || action == nil || action.Kind != "set_summary_date" ||
		action.DueDate == nil ||
		*action.DueDate != "2026-08-17" {
		t.Fatalf("unexpected calendar selection: state=%+v action=%+v", next, action)
	}
}

func TestSummaryCalendarRejectsFutureDate(t *testing.T) {
	state := OpenSummaryCalendar(State{}, "2026-08-18")
	state.DateCursorValue = "2026-08-19"
	next, action, _ := updateSummaryCalendar(state, "2026-08-18", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || next.Kind != "summary_calendar" || next.ErrorMessage == "" {
		t.Fatalf("expected future date rejection: state=%+v action=%+v", next, action)
	}
}

func TestSummaryCalendarCursorStaysInsideValidMonthAndToday(t *testing.T) {
	state := OpenSummaryCalendar(State{}, "2026-08-01")
	state = summaryCalendarMove(state, "2026-08-18", -1)
	if state.DateCursorValue != "2026-08-01" {
		t.Fatalf("expected previous-month spillover to be skipped, got %q", state.DateCursorValue)
	}
	state.DateCursorValue = "2026-08-18"
	state = summaryCalendarMove(state, "2026-08-18", 1)
	if state.DateCursorValue != "2026-08-18" {
		t.Fatalf("expected future date to be skipped, got %q", state.DateCursorValue)
	}
}
