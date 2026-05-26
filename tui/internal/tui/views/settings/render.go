package settings

import types "crona/tui/internal/tui/views/types"
import viewui "crona/tui/internal/tui/views/ui"

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}
