package updates

import (
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func Render(theme types.Theme, state types.ContentState) string {
	view := viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("updates", false, renderView),
		},
	}
	layout := viewui.NewLayout(theme, state, view.Render)
	return layout.RenderView()
}
