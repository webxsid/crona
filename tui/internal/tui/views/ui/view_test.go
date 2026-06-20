package ui

import (
	"strings"
	"testing"

	viewtypes "crona/tui/internal/tui/views/types"

	tea "github.com/charmbracelet/bubbletea"
)

type fakePane struct {
	id       string
	focused  bool
	resizedW int
	resizedH int
	rendered string
	keySeen  string
}

func (p *fakePane) ID() string               { return p.id }
func (p *fakePane) Focusable() bool          { return true }
func (p *fakePane) Focus(focused bool)       { p.focused = focused }
func (p *fakePane) Resize(width, height int) { p.resizedW, p.resizedH = width, height }
func (p *fakePane) Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	return p.rendered
}

func (p *fakePane) HandleKey(
	msg tea.KeyMsg,
	state viewtypes.ContentState,
) (viewtypes.ContentState, tea.Cmd, bool) {
	p.keySeen = msg.String()
	return state, nil, true
}

func TestViewDelegatesToFocusedPane(t *testing.T) {
	pane := &fakePane{id: "issues", rendered: "pane"}
	view := View{
		ID:           "default",
		Panes:        []Pane{pane},
		FocusedIndex: 0,
	}

	rendered := view.Render(viewtypes.Theme{}, viewtypes.ContentState{})
	if rendered != "pane" {
		t.Fatalf("expected pane render to be returned, got %q", rendered)
	}

	_, _ = view.HandleKey(
		tea.KeyMsg{Type: tea.KeyDown},
		viewtypes.ContentState{},
	)
	if pane.keySeen != "down" {
		t.Fatalf("expected focused pane to receive key, got %q", pane.keySeen)
	}
	view.SetFocusedIndex(0)
	if !pane.focused {
		t.Fatalf("expected focused pane to be marked focused")
	}
	view.Resize(80, 20)
	if pane.resizedW != 80 || pane.resizedH != 20 {
		t.Fatalf("expected resize to propagate, got %dx%d", pane.resizedW, pane.resizedH)
	}
}

func TestPaneBaseControlLineSwitchesModes(t *testing.T) {
	base := PaneBase{}
	active := base.ControlLine(viewtypes.Theme{}, 40, true, []string{"[x] close"}, false)
	if !strings.Contains(active, "[x] close") {
		t.Fatalf("expected active control line to render actions, got %q", active)
	}
	inactive := base.ControlLine(viewtypes.Theme{}, 40, false, []string{"[x] close"}, true)
	if inactive != "" {
		t.Fatalf("expected inactive control line to be empty, got %q", inactive)
	}
}
