package reports

import (
	"fmt"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
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
	active := state.Pane == "export_reports"
	cur := state.Cursors["export_reports"]
	items := reportItems(state.ExportReports)
	indices := viewhelpers.FilteredStrings(items, state.Filters["export_reports"])
	total := len(indices)
	lines := []string{
		theme.StylePaneTitle.Render("Reports"),
		viewchrome.RenderPaneActionLine(
			theme,
			viewchrome.ContextualActions(
				theme,
				viewchrome.ActionsState{View: state.View, Pane: state.Pane},
			),
			state.Width-6,
		),
	}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export configuration..."))
		return viewchrome.RenderPaneBox(
			theme,
			active,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}
	lines = append(lines, theme.StyleDim.Render("Dir: "+state.ExportAssets.ReportsDir))
	if total == 0 {
		lines = append(lines, "", theme.StyleDim.Render("No exported reports found"))
		return viewchrome.RenderPaneBox(
			theme,
			active,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}
	inner := viewchrome.RemainingPaneHeight(state.Height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(
			lines,
			viewchrome.RenderPaneRowStyled(
				theme,
				i,
				cur,
				active,
				items[indices[i]],
				nil,
				state.Width,
			),
		)
	}
	if remaining := total - end; remaining > 0 {
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
