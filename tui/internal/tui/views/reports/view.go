package reports

import (
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
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
