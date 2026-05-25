package sessions

import (
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("sessions", true, renderView),
		},
	}
}

func renderView(theme types.Theme, state types.ContentState) string {
	if state.View == "session_history" {
		return renderHistoryView(theme, state)
	}
	if state.Timer == nil || state.Timer.State == "idle" {
		return renderHistoryView(theme, state)
	}
	return renderActiveView(theme, state)
}
