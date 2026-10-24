package input

import (
	"testing"

	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDailyRCarriesOnlyOnOverdue(t *testing.T) {
	carried := 0
	deps := dateDeps(new(int))
	deps.CarrySelectedIssue = func(*State) tea.Cmd { carried++; return nil }
	state := State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues, DailyTaskSection: uistate.DailyTaskSectionPlanned}
	Handle(state, runeKey('r'), deps)
	if carried != 0 {
		t.Fatal("expected r to do nothing outside Overdue")
	}
	state.DailyTaskSection = uistate.DailyTaskSectionOverdue
	Handle(state, runeKey('r'), deps)
	if carried != 1 {
		t.Fatal("expected r on Overdue to carry the selected issue")
	}
}
