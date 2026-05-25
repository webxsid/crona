package ui

import (
	"strings"

	viewtypes "crona/tui/internal/tui/views/types"

	tea "github.com/charmbracelet/bubbletea"
)

type Pane interface {
	ID() string
	Focusable() bool
	Focus(bool)
	Resize(width, height int)
	Render(theme viewtypes.Theme, state viewtypes.ContentState) string
	HandleKey(msg tea.KeyMsg, state viewtypes.ContentState) (viewtypes.ContentState, tea.Cmd, bool)
}

type RenderPane struct {
	PaneBase
	CanFocus   bool
	RenderFunc func(viewtypes.Theme, viewtypes.ContentState) string
	KeyFunc    func(tea.KeyMsg, viewtypes.ContentState) (viewtypes.ContentState, tea.Cmd, bool)
}

func NewRenderPane(
	id string,
	focusable bool,
	renderFunc func(viewtypes.Theme, viewtypes.ContentState) string,
) *RenderPane {
	return &RenderPane{
		PaneBase:   PaneBase{ID: id},
		CanFocus:   focusable,
		RenderFunc: renderFunc,
	}
}

func (p *RenderPane) ID() string {
	return p.PaneBase.ID
}

func (p *RenderPane) Focusable() bool {
	return p.CanFocus
}

func (p *RenderPane) Focus(focused bool) {
	p.PaneBase.Focus(focused)
}

func (p *RenderPane) Resize(width, height int) {
	p.PaneBase.Resize(width, height)
}

func (p *RenderPane) Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	if p.RenderFunc == nil {
		return ""
	}
	return p.RenderFunc(theme, state)
}

func (p *RenderPane) HandleKey(
	msg tea.KeyMsg,
	state viewtypes.ContentState,
) (viewtypes.ContentState, tea.Cmd, bool) {
	if p.KeyFunc == nil {
		return state, nil, false
	}
	return p.KeyFunc(msg, state)
}

type View struct {
	ID           string
	Panes        []Pane
	FocusedIndex int
	HandleGlobal func(tea.KeyMsg, viewtypes.ContentState) (viewtypes.ContentState, tea.Cmd, bool)
	RenderFunc   func(viewtypes.Theme, viewtypes.ContentState) string
}

func (v *View) SetFocusedIndex(index int) {
	if len(v.Panes) == 0 {
		v.FocusedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(v.Panes) {
		index = len(v.Panes) - 1
	}
	for i, pane := range v.Panes {
		if pane == nil {
			continue
		}
		pane.Focus(i == index)
	}
	v.FocusedIndex = index
}

func (v *View) Resize(width, height int) {
	for _, pane := range v.Panes {
		if pane == nil {
			continue
		}
		pane.Resize(width, height)
	}
}

func (v View) FocusedPane() Pane {
	if len(v.Panes) == 0 || v.FocusedIndex < 0 || v.FocusedIndex >= len(v.Panes) {
		return nil
	}
	return v.Panes[v.FocusedIndex]
}

func (v View) Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	if v.RenderFunc != nil {
		return v.RenderFunc(theme, state)
	}
	parts := make([]string, 0, len(v.Panes))
	for _, pane := range v.Panes {
		if pane == nil {
			continue
		}
		parts = append(parts, pane.Render(theme, state))
	}
	return strings.Join(parts, "\n")
}

func (v View) HandleKey(
	msg tea.KeyMsg,
	state viewtypes.ContentState,
) (viewtypes.ContentState, tea.Cmd) {
	if v.HandleGlobal != nil {
		if next, cmd, handled := v.HandleGlobal(msg, state); handled {
			return next, cmd
		}
	}
	if pane := v.FocusedPane(); pane != nil {
		if next, cmd, handled := pane.HandleKey(msg, state); handled {
			return next, cmd
		}
	}
	return state, nil
}
