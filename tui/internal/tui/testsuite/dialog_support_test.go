package testsuite

import (
	"testing"

	dialogs "crona/tui/internal/tui/dialogs"
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
