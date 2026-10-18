package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

var datedViews = []uistate.View{
	uistate.ViewDaily,
	uistate.ViewSummary,
	uistate.ViewMomentum,
	uistate.ViewWellbeing,
	uistate.ViewRollup,
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
		LoadRollupSummaries: func(string, string) tea.Cmd { return nil },
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

func TestRollupWindowPresets(t *testing.T) {
	reloads := 0
	var loaded [2]string
	deps := dateDeps(&reloads)
	deps.LoadRollupSummaries = func(start, end string) tea.Cmd {
		loaded = [2]string{start, end}
		return nil
	}
	s := State{ActiveView: uistate.ViewRollup, ActivePane: uistate.DefaultPane(uistate.ViewRollup), Date: "2026-07-30"}
	for _, want := range []int{14, 30, 90, 90} {
		s, _ = Handle(s, runeKey('+'), deps)
		if s.RollupWindowDays != want {
			t.Fatalf("expected + to step the rollup window to %d, got %d", want, s.RollupWindowDays)
		}
	}
	if loaded != [2]string{"2026-05-02", "2026-07-30"} {
		t.Fatalf("expected a 90-day rollup ending on the shared date, got %v", loaded)
	}
	s, _ = Handle(s, runeKey('-'), deps)
	if s.RollupWindowDays != 30 {
		t.Fatalf("expected - to step back to 30, got %d", s.RollupWindowDays)
	}
	s, _ = Handle(s, runeKey('0'), deps)
	if s.RollupWindowDays != 7 || loaded != [2]string{"2026-07-24", "2026-07-30"} {
		t.Fatalf("expected 0 to reset to a 7-day rollup, got %d %v", s.RollupWindowDays, loaded)
	}
}
