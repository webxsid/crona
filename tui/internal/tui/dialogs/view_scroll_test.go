package dialogs

import (
	"fmt"
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func longViewEntity(height int) controllerpkg.State {
	lines := []string{"Description"}
	for i := 1; i <= 60; i++ {
		lines = append(lines, fmt.Sprintf("line %02d", i))
	}
	state := controllerpkg.OpenViewEntity(controllerpkg.State{}, "Issue", "Long issue", "ID 1", strings.Join(lines, "\n"))
	return sized(state, 120, height)
}

func TestViewEntityScrollsALongBody(t *testing.T) {
	state := longViewEntity(30)
	top := ansi.Strip(Render(testTheme(), state))
	if !strings.Contains(top, "line 01") || strings.Contains(top, "line 60") ||
		!strings.Contains(top, "↓ more") || !strings.Contains(top, "↑/↓ scroll") {
		t.Fatalf("expected the top of the body with a scroll hint, got\n%s", top)
	}
	for range 5 {
		state, _, _ = controllerpkg.Update(state, controllerpkg.UpdateContext{}, "", tea.KeyMsg{Type: tea.KeyDown})
	}
	if moved := ansi.Strip(Render(testTheme(), state)); !strings.Contains(moved, "↑ more") {
		t.Fatalf("expected scrolling down to reveal ↑ more, got\n%s", moved)
	}
	state, _, _ = controllerpkg.Update(state, controllerpkg.UpdateContext{}, "", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	state = ClampViewScroll(testTheme(), state)
	bottom := ansi.Strip(Render(testTheme(), state))
	if !strings.Contains(bottom, "line 60") || strings.Contains(bottom, "↓ more") {
		t.Fatalf("expected G to reach the end, got\n%s", bottom)
	}
	end := state.ViewScroll
	state, _, _ = controllerpkg.Update(state, controllerpkg.UpdateContext{}, "", tea.KeyMsg{Type: tea.KeyUp})
	if state.ViewScroll != end-1 {
		t.Fatalf("expected one up from the clamped end to move, got %d from %d", state.ViewScroll, end)
	}
}

func TestShortViewEntityHasNoScrollHint(t *testing.T) {
	state := sized(controllerpkg.OpenViewEntity(controllerpkg.State{}, "Issue", "Short", "ID 1", "Description\nhi"), 120, 40)
	if out := ansi.Strip(Render(testTheme(), state)); strings.Contains(out, "scroll") || strings.Contains(out, "more") {
		t.Fatalf("expected no scroll chrome on a short body, got\n%s", out)
	}
}
