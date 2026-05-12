package testsuite

import (
	"testing"

	inputpkg "crona/tui/internal/tui/input"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMetaRepoPaneCUsesCheckout(t *testing.T) {
	calledCheckout := false
	openedDialog := false
	state := inputpkg.State{
		ActiveView: uistate.ViewMeta,
		ActivePane: uistate.PaneRepos,
	}
	next, _ := inputpkg.Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}, inputpkg.Deps{
		Checkout: func(_ *inputpkg.State) tea.Cmd {
			calledCheckout = true
			return nil
		},
		OpenCheckoutContextDialog: func(_ *inputpkg.State) bool {
			openedDialog = true
			return true
		},
	})
	if next.ActiveView != uistate.ViewMeta || next.ActivePane != uistate.PaneRepos {
		t.Fatalf("unexpected state after repo c: view=%q pane=%q", next.ActiveView, next.ActivePane)
	}
	if !calledCheckout {
		t.Fatal("expected repo-pane c to invoke checkout")
	}
	if openedDialog {
		t.Fatal("expected repo-pane c to avoid opening checkout dialog")
	}
}

func TestMetaStreamPaneCUsesCheckout(t *testing.T) {
	calledCheckout := false
	openedDialog := false
	state := inputpkg.State{
		ActiveView: uistate.ViewMeta,
		ActivePane: uistate.PaneStreams,
	}
	next, _ := inputpkg.Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}, inputpkg.Deps{
		Checkout: func(_ *inputpkg.State) tea.Cmd {
			calledCheckout = true
			return nil
		},
		OpenCheckoutContextDialog: func(_ *inputpkg.State) bool {
			openedDialog = true
			return true
		},
	})
	if next.ActiveView != uistate.ViewMeta || next.ActivePane != uistate.PaneStreams {
		t.Fatalf("unexpected state after stream c: view=%q pane=%q", next.ActiveView, next.ActivePane)
	}
	if !calledCheckout {
		t.Fatal("expected stream-pane c to invoke checkout")
	}
	if openedDialog {
		t.Fatal("expected stream-pane c to avoid opening checkout dialog")
	}
}

func TestMetaIssuePaneCStillOpensDialog(t *testing.T) {
	calledCheckout := false
	openedDialog := false
	state := inputpkg.State{
		ActiveView: uistate.ViewMeta,
		ActivePane: uistate.PaneIssues,
	}
	next, _ := inputpkg.Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}, inputpkg.Deps{
		Checkout: func(_ *inputpkg.State) tea.Cmd {
			calledCheckout = true
			return nil
		},
		OpenCheckoutContextDialog: func(_ *inputpkg.State) bool {
			openedDialog = true
			return true
		},
	})
	if next.ActiveView != uistate.ViewMeta || next.ActivePane != uistate.PaneIssues {
		t.Fatalf("unexpected state after issue c: view=%q pane=%q", next.ActiveView, next.ActivePane)
	}
	if calledCheckout {
		t.Fatal("expected issue-pane c to avoid direct checkout")
	}
	if !openedDialog {
		t.Fatal("expected issue-pane c to open checkout dialog")
	}
}
