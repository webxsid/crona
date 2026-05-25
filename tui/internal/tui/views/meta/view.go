package meta

import viewui "crona/tui/internal/tui/views/ui"

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("meta", true, renderView),
		},
		RenderFunc: renderView,
	}
}
