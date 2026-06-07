package layout

import (
	"strings"
	"testing"

	dialogstate "crona/tui/internal/tui/dialogs/controller"
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

func TestRenderProtectedSidebarIncludesSettingsView(t *testing.T) {
	rendered := Render(State{
		Width:         140,
		Height:        44,
		View:          uistate.ViewSettings,
		Pane:          uistate.PaneIssues,
		ProtectedMode: true,
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewSettings),
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

	if !strings.Contains(rendered, "Settings") {
		t.Fatalf("expected protected sidebar to include Settings, got %q", rendered)
	}
}

func TestRenderDialogUsesBlankScreen(t *testing.T) {
	rendered := Render(State{
		Width:      140,
		Height:     44,
		View:       uistate.ViewDaily,
		Pane:       uistate.PaneIssues,
		DialogOpen: true,
		DialogState: dialogstate.State{
			Kind:      "view_entity",
			Width:     140,
			ViewTitle: "Keys",
			ViewName:  "Keyboard Shortcuts",
			ViewMeta:  "Hint Press ? or esc to close",
			ViewBody:  "Global\n[v] switch views",
		},
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewDaily),
			Pane:   string(uistate.PaneIssues),
			Width:  118,
			Height: 44,
			Cursors: map[string]int{
				string(uistate.PaneIssues): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneIssues): "",
			},
		},
	})

	if strings.Contains(rendered, "Habit History") || strings.Contains(rendered, "Daily") {
		t.Fatalf("expected dialog render to exclude base screen content, got %q", rendered)
	}
	if !strings.Contains(rendered, "Keyboard Shortcuts") {
		t.Fatalf("expected dialog render to show the modal content, got %q", rendered)
	}
}
