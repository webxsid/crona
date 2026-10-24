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

func TestSlashOpensFilterAndEscClearsIt(t *testing.T) {
	opened := false
	var cleared *string
	deps := dateDeps(new(int))
	deps.OpenPaneFilter = func(*State) bool { opened = true; return true }
	deps.SetPaneFilter = func(_ uistate.Pane, value string) tea.Cmd { cleared = &value; return nil }
	state := State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneIssues, Filters: map[uistate.Pane]string{}}
	Handle(state, runeKey('/'), deps)
	if !opened {
		t.Fatal("expected / to open the filter")
	}
	state.Filters[uistate.PaneIssues] = "+urgent"
	Handle(state, tea.KeyMsg{Type: tea.KeyEsc}, deps)
	if cleared == nil || *cleared != "" {
		t.Fatal("expected esc to clear the active filter")
	}
}
