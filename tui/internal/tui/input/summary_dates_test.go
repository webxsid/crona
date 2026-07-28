package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSummaryDateNavigationDoesNotMutateDashboardDate(t *testing.T) {
	state := State{
		ActiveView:    uistate.ViewSummary,
		DashboardDate: "2026-07-01",
		SummaryDate:   "2026-07-28",
	}
	var loaded string
	deps := Deps{
		CurrentSummaryDate: func(s State) string { return s.SummaryDate },
		LoadSummarySnapshot: func(date string) tea.Cmd {
			loaded = date
			return nil
		},
	}
	nextModel, _, handled := handleShiftSummaryDate(state, deps, -1)
	if !handled {
		t.Fatal("expected summary date shift to be handled")
	}
	next := nextModel.(State)
	if next.SummaryDate != "2026-07-27" || loaded != "2026-07-27" {
		t.Fatalf("unexpected summary shift: state=%q loaded=%q", next.SummaryDate, loaded)
	}
	if next.DashboardDate != "2026-07-01" {
		t.Fatalf("summary navigation changed dashboard date to %q", next.DashboardDate)
	}
}
