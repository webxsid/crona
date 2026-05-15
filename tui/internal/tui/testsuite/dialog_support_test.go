package testsuite

import (
	"testing"

	sharedtypes "crona/shared/types"
	dialogs "crona/tui/internal/tui/dialogs/controller"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewJumpDialogUsesMnemonicKeys(t *testing.T) {
	state := dialogs.OpenViewJump(dialogs.State{}, []uistate.View{uistate.ViewUpdates, uistate.ViewSupport})
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-04", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got kind %q", next.Kind)
	}
	if action == nil || action.Kind != "jump_view" || action.TargetView != "updates" {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestViewJumpDialogHidesUnavailableViews(t *testing.T) {
	state := dialogs.OpenViewJump(dialogs.State{}, []uistate.View{uistate.ViewDaily, uistate.ViewSupport})
	for _, item := range state.ChoiceItems {
		if item == "[a] Away" || item == "[n] Session" || item == "[u] Updates" {
			t.Fatalf("unexpected unavailable item %q", item)
		}
	}

	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-04", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "view_jump" {
		t.Fatalf("expected dialog to remain open, got kind %q", next.Kind)
	}
	if action != nil {
		t.Fatalf("expected no action for unavailable key, got %+v", action)
	}
}

func TestViewJumpDialogRestrictsChoicesDuringActiveSession(t *testing.T) {
	state := dialogs.OpenViewJump(dialogs.State{}, []uistate.View{
		uistate.ViewSessionActive,
		uistate.ViewSessionHistory,
		uistate.ViewScratch,
	})
	for _, item := range state.ChoiceItems {
		switch item {
		case "[n] Session", "[y] History", "[x] Scratchpads":
		default:
			t.Fatalf("unexpected item %q in active-session jump menu", item)
		}
	}
}

func TestBetaSupportDialogUsesMnemonicKeys(t *testing.T) {
	state := dialogs.OpenBetaSupport(dialogs.State{})
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-04", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got kind %q", next.Kind)
	}
	if action == nil || action.Kind != "generate_support_bundle" {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestDirectoryDetailDialogRoutesChangeAndResetKeys(t *testing.T) {
	tests := []struct {
		name       string
		viewName   string
		changeKind string
		resetKind  string
	}{
		{
			name:       "reports",
			viewName:   "Reports directory",
			changeKind: "open_export_reports_dir_dialog",
			resetKind:  "reset_export_reports_dir",
		},
		{
			name:       "ics",
			viewName:   "ICS export directory",
			changeKind: "open_export_ics_dir_dialog",
			resetKind:  "reset_export_ics_dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" change", func(t *testing.T) {
			state := dialogs.OpenViewEntity(dialogs.State{}, "Directory", tt.viewName, "", "")
			next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
			if status != "" {
				t.Fatalf("unexpected status %q", status)
			}
			if next.Kind != "" {
				t.Fatalf("expected dialog to close, got kind %q", next.Kind)
			}
			if action == nil || action.Kind != tt.changeKind {
				t.Fatalf("unexpected action %+v", action)
			}
		})

		t.Run(tt.name+" reset", func(t *testing.T) {
			state := dialogs.OpenViewEntity(dialogs.State{}, "Directory", tt.viewName, "", "")
			next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
			if status != "" {
				t.Fatalf("unexpected status %q", status)
			}
			if next.Kind != "" {
				t.Fatalf("expected dialog to close, got kind %q", next.Kind)
			}
			if action == nil || action.Kind != tt.resetKind {
				t.Fatalf("unexpected action %+v", action)
			}
		})
	}
}

func TestViewEntityDialogRoutesEditorKeyWhenPathIsAvailable(t *testing.T) {
	state := dialogs.OpenViewEntityWithPath(dialogs.State{}, "Template", "Daily report template", "", "Press e to open in $EDITOR.", "/tmp/report.hbs")
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got kind %q", next.Kind)
	}
	if action == nil || action.Kind != "open_view_entity_editor" || action.Path != "/tmp/report.hbs" {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestViewEntityDialogIgnoresEditorKeyWithoutPath(t *testing.T) {
	state := dialogs.OpenViewEntity(dialogs.State{}, "Directory", "Reports directory", "", "Press c to change the directory.")
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "view_entity" {
		t.Fatalf("expected dialog to remain open, got kind %q", next.Kind)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
}

func TestStashConflictDialogOffersResumeAndContinue(t *testing.T) {
	state := dialogs.OpenStashConflict(dialogs.State{}, sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{{
			ID:        "stash-1",
			CreatedAt: "2026-04-10T09:00:00Z",
		}},
	})
	if state.Kind != "stash_conflict" {
		t.Fatalf("expected single-stash conflict dialog, got %q", state.Kind)
	}
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "apply_stash" || action.ID != "stash-1" {
		t.Fatalf("unexpected action %+v", action)
	}

	state = dialogs.OpenStashConflict(dialogs.State{}, sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{{
			ID:        "stash-1",
			CreatedAt: "2026-04-10T09:00:00Z",
		}},
	})
	state.RepoID = 7
	state.StreamID = 8
	state.IssueID = 42
	next, action, status = dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "continue_focus_fresh" || action.RepoID != 7 || action.StreamID != 8 || action.IssueID != 42 {
		t.Fatalf("unexpected continue action %+v", action)
	}
}

func TestStashConflictPickDialogOpensSelectedStash(t *testing.T) {
	state := dialogs.OpenStashConflict(dialogs.State{}, sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{
			{ID: "stash-1", CreatedAt: "2026-04-10T09:00:00Z"},
			{ID: "stash-2", CreatedAt: "2026-04-10T08:00:00Z"},
		},
	})
	if state.Kind != "stash_conflict_pick" {
		t.Fatalf("expected stash picker, got %q", state.Kind)
	}
	next, action, status := dialogs.Update(state, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if status != "" || action != nil {
		t.Fatalf("unexpected move result action=%+v status=%q", action, status)
	}
	next, action, status = dialogs.Update(next, dialogs.UpdateContext{}, "2026-04-10", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no immediate action, got %+v", action)
	}
	if next.Kind != "stash_conflict" || next.DeleteID != "stash-2" {
		t.Fatalf("expected selected stash conflict dialog, got kind=%q id=%q", next.Kind, next.DeleteID)
	}
}
