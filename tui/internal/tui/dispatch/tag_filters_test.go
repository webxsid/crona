package dispatch

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPaneFilterSavesOnlyItsTagPart(t *testing.T) {
	var saved map[string]string
	deps := MessageDeps{SaveTagFilters: func(f map[string]string) tea.Cmd { saved = f; return nil }}
	state := MessageState{Settings: &api.CoreSettings{}}
	state, _, _ = HandleMessage(state, commands.PaneFilterSetMsg{Pane: "issues", Value: "+urgent login"}, deps)
	if state.Filters[uistate.PaneIssues] != "+urgent login" || saved["issues"] != "+urgent" {
		t.Fatalf("expected filter applied and tag part saved, got %q / %v", state.Filters[uistate.PaneIssues], saved)
	}
	saved = nil
	state.Settings.TagFilters = map[string]string{"issues": "+urgent"}
	state, _, _ = HandleMessage(state, commands.PaneFilterSetMsg{Pane: "issues", Value: "+urgent other"}, deps)
	if saved != nil {
		t.Fatalf("expected no save when only the text changed, got %v", saved)
	}
}

func TestSavedTagFiltersRestoreOnce(t *testing.T) {
	settings := &api.CoreSettings{TagFilters: map[string]string{"habits": "+health -gone"}}
	state := MessageState{Tags: []api.Tag{{Name: "health"}}}
	state, _, _ = HandleMessage(state, commands.SettingsLoadedMsg{Settings: settings}, MessageDeps{})
	if state.Filters[uistate.PaneHabits] != "+health" || !state.TagFiltersRestored {
		t.Fatalf("expected the saved filter restored without the deleted tag, got %q", state.Filters[uistate.PaneHabits])
	}
	state.Filters[uistate.PaneHabits] = ""
	state, _, _ = HandleMessage(state, commands.SettingsLoadedMsg{Settings: settings}, MessageDeps{})
	if state.Filters[uistate.PaneHabits] != "" {
		t.Fatal("expected later settings loads not to restore again")
	}
}
