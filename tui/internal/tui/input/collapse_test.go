package input

import (
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	uistate "crona.work/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAltNumbersFoldTodayPanesAndSaveThem(t *testing.T) {
	var savedKey sharedtypes.CoreSettingsKey
	var saved any
	deps := Deps{
		DefaultPane: uistate.DefaultPane,
		NextPane:    uistate.NextPane,
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, _, _ int64, _ string) tea.Cmd {
			savedKey, saved = key, value
			return nil
		},
	}
	state := State{ActiveView: uistate.ViewDaily, ActivePane: uistate.PaneHabits, Settings: &api.CoreSettings{}}
	next, _ := Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}, Alt: true}, deps)
	if !uistate.PaneCollapsed(next.Settings.CollapsedPanes, uistate.CollapseDailyHabits) {
		t.Fatalf("expected alt+1 to fold Habits, got %v", next.Settings.CollapsedPanes)
	}
	if savedKey != sharedtypes.CoreSettingsKeyCollapsedPanes || len(saved.([]string)) != 1 {
		t.Fatalf("expected the fold saved, got %q %v", savedKey, saved)
	}
	if next.ActivePane != uistate.PaneIssues {
		t.Fatalf("expected focus to leave the folded Habits pane, got %q", next.ActivePane)
	}
	if state.Settings.CollapsedPanes != nil {
		t.Fatal("expected the original settings left untouched")
	}
	if tabbed, _ := Handle(next, tea.KeyMsg{Type: tea.KeyTab}, deps); tabbed.ActivePane != uistate.PaneIssues {
		t.Fatalf("expected tab to skip folded Habits, got %q", tabbed.ActivePane)
	}
	again, _ := Handle(next, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}, Alt: true}, deps)
	if uistate.PaneCollapsed(again.Settings.CollapsedPanes, uistate.CollapseDailyHabits) {
		t.Fatal("expected a second alt+1 to unfold Habits")
	}
}
