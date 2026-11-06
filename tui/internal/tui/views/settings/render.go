package settings

import types "crona.work/tui/internal/tui/views/types"
import viewui "crona.work/tui/internal/tui/views/ui"

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}
