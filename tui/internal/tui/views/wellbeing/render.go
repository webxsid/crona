package wellbeing

import (
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, renderView).RenderView()
}

func renderView(theme types.Theme, state types.ContentState) string {
	if state.Height < 30 {
		return renderSmallScreen(theme, state)
	}
	if state.Width >= 96 {
		return renderSplit(theme, state)
	}
	return renderCompact(theme, state)
}
