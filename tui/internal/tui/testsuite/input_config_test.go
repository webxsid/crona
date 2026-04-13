package testsuite

import (
	"testing"

	inputpkg "crona/tui/internal/tui/input"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigPaneChangeDirectoryShortcutUsesC(t *testing.T) {
	called := false
	state := inputpkg.State{
		ActiveView: uistate.ViewConfig,
		ActivePane: uistate.PaneConfig,
	}
	next, _ := inputpkg.Handle(state, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}, inputpkg.Deps{
		ConfigChangeSelected: func(_ *inputpkg.State) tea.Cmd {
			called = true
			return nil
		},
	})
	if next.ActiveView != uistate.ViewConfig || next.ActivePane != uistate.PaneConfig {
		t.Fatalf("unexpected state after c: view=%q pane=%q", next.ActiveView, next.ActivePane)
	}
	if !called {
		t.Fatal("expected c to invoke ConfigChangeSelected")
	}
}
