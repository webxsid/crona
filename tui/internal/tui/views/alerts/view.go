package alerts

import (
	"fmt"

	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("alerts", true, renderView),
		},
	}
}

func renderView(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "alerts"
	cur := state.Cursors["alerts"]
	rows := alertsmeta.Rows(state.Settings, state.AlertStatus, state.AlertReminders)
	indices := alertsmeta.FilteredIndices(
		state.Filters["alerts"],
		state.Settings,
		state.AlertStatus,
		state.AlertReminders,
	)
	lines := []string{
		theme.StylePaneTitle.Render("Alerts"),
		viewchrome.RenderActionLine(
			theme,
			state.Width-6,
			viewchrome.ContextualActions(
				theme,
				viewchrome.ActionsState{View: state.View, Pane: state.Pane},
			),
		),
		"",
	}
	if state.Settings == nil {
		lines = append(lines, theme.StyleDim.Render("Loading alert settings..."))
		return viewchrome.RenderPaneBox(
			theme,
			active,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render("No alert rows match the current filter"))
		return viewchrome.RenderPaneBox(
			theme,
			active,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}
	maxSelectable := alertsmeta.FilteredSelectableCount(
		state.Filters["alerts"],
		state.Settings,
		state.AlertStatus,
		state.AlertReminders,
	)
	if maxSelectable == 0 {
		lines = append(
			lines,
			theme.StyleDim.Render("No selectable alert rows match the current filter"),
		)
		return viewchrome.RenderPaneBox(
			theme,
			active,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}
	if cur >= maxSelectable {
		cur = maxSelectable - 1
	}
	cur = max(cur, 0)
	visibleRows, selectedVisible := alertsmeta.GroupedVisibleRows(indices, rows, cur)
	inner := state.Height - 5
	inner = max(inner, 1)
	start, end := viewchrome.ListWindow(selectedVisible, len(visibleRows), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		row := visibleRows[i]
		if row.Header {
			lines = append(lines, theme.StyleHeader.Render(row.Text))
			continue
		}
		switch {
		case row.SelectableAt == cur && active:
			lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+row.Text))
		case row.SelectableAt == cur:
			lines = append(lines, theme.StyleSelected.Render("  "+row.Text))
		default:
			lines = append(lines, theme.StyleNormal.Render("  "+row.Text))
		}
	}
	if remaining := len(visibleRows) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(
		theme,
		active,
		state.Width,
		state.Height,
		viewhelpers.StringsJoin(lines),
	)
}
