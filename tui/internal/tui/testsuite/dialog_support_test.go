package testsuite

import (
	"testing"

	"crona/tui/internal/api"
	dialogs "crona/tui/internal/tui/dialogs/controller"
	uistate "crona/tui/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestViewJumpDialogUsesMnemonicKeys(t *testing.T) {
	state := dialogs.OpenViewJump(
		dialogs.State{},
		[]uistate.View{uistate.ViewUpdates, uistate.ViewSupport},
	)
	next, action, status := dialogs.Update(
		state,
		dialogs.UpdateContext{},
		"2026-04-04",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}},
	)
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
	state := dialogs.OpenViewJump(
		dialogs.State{},
		[]uistate.View{uistate.ViewDaily, uistate.ViewSupport},
	)
	for _, item := range state.ChoiceItems {
		if item == "[a] Away" || item == "[n] Session" || item == "[t] Updates" {
			t.Fatalf("unexpected unavailable item %q", item)
		}
	}

	next, action, status := dialogs.Update(
		state,
		dialogs.UpdateContext{},
		"2026-04-04",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
	)
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
	})
	for _, item := range state.ChoiceItems {
		switch item {
		case "[n] Session", "[y] History":
		default:
			t.Fatalf("unexpected item %q in active-session jump menu", item)
		}
	}
}

func TestCheckoutDialogLabelsShowPlaceholdersWhenInputsAreEmpty(t *testing.T) {
	repoInput := textinput.New()
	streamInput := textinput.New()
	repoLabel, streamLabel := dialogs.CheckoutDialogLabels(
		[]textinput.Model{repoInput, streamInput},
		-1,
		-1,
		[]api.Repo{{ID: 1, Name: "Work"}},
		nil,
		[]api.Stream{{ID: 2, RepoID: 1, Name: "main"}},
		nil,
	)
	if repoLabel != "Select a repo" {
		t.Fatalf("expected empty repo input to show placeholder, got %q", repoLabel)
	}
	if streamLabel != "Select a stream" {
		t.Fatalf("expected empty stream input to show placeholder, got %q", streamLabel)
	}
}

func TestBetaSupportDialogUsesMnemonicKeys(t *testing.T) {
	state := dialogs.OpenBetaSupport(dialogs.State{})
	next, action, status := dialogs.Update(
		state,
		dialogs.UpdateContext{},
		"2026-04-04",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}},
	)
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
			next, action, status := dialogs.Update(
				state,
				dialogs.UpdateContext{},
				"2026-04-10",
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
			)
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
			next, action, status := dialogs.Update(
				state,
				dialogs.UpdateContext{},
				"2026-04-10",
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			)
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
	state := dialogs.OpenViewEntityWithPath(
		dialogs.State{},
		"Template",
		"Daily report template",
		"",
		"Press e to open in $EDITOR.",
		"/tmp/report.hbs",
	)
	next, action, status := dialogs.Update(
		state,
		dialogs.UpdateContext{},
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got kind %q", next.Kind)
	}
	if action == nil || action.Kind != "open_view_entity_editor" ||
		action.Path != "/tmp/report.hbs" {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestViewEntityDialogIgnoresEditorKeyWithoutPath(t *testing.T) {
	state := dialogs.OpenViewEntity(
		dialogs.State{},
		"Directory",
		"Reports directory",
		"",
		"Press c to change the directory.",
	)
	next, action, status := dialogs.Update(
		state,
		dialogs.UpdateContext{},
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
	)
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
