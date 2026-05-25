package ui

import viewtypes "crona/tui/internal/tui/views/types"

type Layout struct {
	Theme      viewtypes.Theme
	State      viewtypes.ContentState
	RenderFunc func(viewtypes.Theme, viewtypes.ContentState) string
}

func NewLayout(
	theme viewtypes.Theme,
	state viewtypes.ContentState,
	renderFunc func(viewtypes.Theme, viewtypes.ContentState) string,
) Layout {
	return Layout{
		Theme:      theme,
		State:      state,
		RenderFunc: renderFunc,
	}
}

func (l Layout) RenderView() string {
	if l.RenderFunc == nil {
		return ""
	}
	return l.RenderFunc(l.Theme, l.State)
}
