package layout

import (
	"strings"
	"testing"

	uistate "crona/tui/internal/tui/state"
	viewtypes "crona/tui/internal/tui/views/types"
)

func TestRenderSidebarIncludesHabitHistoryView(t *testing.T) {
	rendered := Render(State{
		Width:  140,
		Height: 44,
		View:   uistate.ViewDaily,
		Pane:   uistate.PaneIssues,
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewDaily),
			Pane:   string(uistate.PaneIssues),
			Width:  118,
			Height: 44,
			Cursors: map[string]int{
				string(uistate.PaneIssues):       0,
				string(uistate.PaneHabits):       0,
				string(uistate.PaneHabitHistory): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneIssues):       "",
				string(uistate.PaneHabits):       "",
				string(uistate.PaneHabitHistory): "",
			},
		},
	})

	if !strings.Contains(rendered, "Habit History") {
		t.Fatalf("expected sidebar to include Habit History, got %q", rendered)
	}
}
