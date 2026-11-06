package reports

import (
	types "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("export_reports", true, renderView),
		},
	}
}

func renderView(theme types.Theme, state types.ContentState) string {
	return renderLog(theme, state)
}
