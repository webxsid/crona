package layout

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
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

func TestRenderMomentumDetailDialogShowsModalContent(t *testing.T) {
	rendered := Render(State{
		Width:      140,
		Height:     44,
		View:       uistate.ViewMomentum,
		Pane:       uistate.PaneMomentumCards,
		DialogOpen: true,
		DialogState: dialogstate.State{
			Kind:      "view_momentum_detail",
			Width:     140,
			ViewTitle: "Momentum Detail",
			ViewName:  "Delivery Context All",
			ViewMeta:  "Kind Context   Match All",
			ViewBody:  "Current Bucket\nStatus Met",
			HabitStreakDraft: sharedtypes.HabitStreakDefinition{
				ID:   "momentum-1",
				Name: "Delivery Context All",
			},
			ChoiceItems:   []string{"2026-06-19  Deep work"},
			ChoiceDetails: []string{"Work/app   1h12m"},
		},
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewMomentum),
			Pane:   string(uistate.PaneMomentumCards),
			Width:  118,
			Height: 44,
			Cursors: map[string]int{
				string(uistate.PaneMomentumCards): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneMomentumCards): "",
			},
		},
	})

	for _, want := range []string{"Momentum Detail", "Delivery Context All", "Contributors", "Deep work"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected momentum detail render to contain %q, got %q", want, rendered)
		}
	}
}
