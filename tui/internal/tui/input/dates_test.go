package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

var datedViews = []uistate.View{
	uistate.ViewDaily,
	uistate.ViewMomentum,
	uistate.ViewReview,
}

func dateDeps(reloads *int) Deps {
	return Deps{
		DefaultPane:        uistate.DefaultPane,
		CurrentDate:        func(s State) string { return s.Date },
		CurrentLogicalDate: func(State) string { return "2026-07-30" },
		ReloadDateSurfaces: func(State) tea.Cmd {
			*reloads++
			return nil
		},
		LoadReviewWindow: func(string, string) tea.Cmd { return nil },
	}
}

func TestShiftDateMovesSharedDateOnEveryDatedView(t *testing.T) {
	for _, view := range datedViews {
		reloads := 0
		start := State{ActiveView: view, ActivePane: uistate.DefaultPane(view), Date: "2026-07-28"}
		back, _ := Handle(start, runeKey(','), dateDeps(&reloads))
		forward, _ := Handle(start, runeKey('.'), dateDeps(&reloads))
		if back.Date != "2026-07-27" || forward.Date != "2026-07-29" {
			t.Fatalf("%s: expected ,/. to move the shared date, got %q / %q", view, back.Date, forward.Date)
		}
		if reloads != 2 {
			t.Fatalf("%s: expected every move to reload the date surfaces, got %d reloads", view, reloads)
		}
	}
}

func TestResetDateGoesToLogicalToday(t *testing.T) {
	for _, view := range datedViews {
		reloads := 0
		next, _ := Handle(State{ActiveView: view, ActivePane: uistate.DefaultPane(view), Date: "2026-07-01"}, runeKey('g'), dateDeps(&reloads))
		if next.Date != "2026-07-30" || reloads != 1 {
			t.Fatalf("%s: expected g to reset to the logical today and reload, got %q (%d reloads)", view, next.Date, reloads)
		}
	}
}

func TestDailyCOpensCalendar(t *testing.T) {
	opened := false
	deps := dateDeps(new(int))
	deps.OpenSummaryCalendar = func(*State) bool { opened = true; return true }
	Handle(State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues}, runeKey('C'), deps)
	if !opened {
		t.Fatal("expected C on Daily to open the calendar")
	}
}
