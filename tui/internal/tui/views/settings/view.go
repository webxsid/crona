package settings

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderView(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "settings"
	cur := state.Cursors["settings"]
	rows := settingsmeta.Rows(state.Settings)
	indices := settingsmeta.FilteredIndices(state.Filters["settings"], state.Settings)
	total := len(indices)
	lines := []string{theme.StylePaneTitle.Render("Settings"), viewchrome.RenderActionLine(theme, state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})), ""}
	if state.Settings == nil {
		lines = append(lines, theme.StyleDim.Render("Loading settings..."))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No settings match the current filter"))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	visibleRows, selectedVisibleIdx := settingsmeta.GroupedVisibleRows(indices, cur, func(idx int) string { return rows[idx].Section }, func(idx int) string {
		return fmt.Sprintf("%-24s %s", rows[idx].Label, rows[idx].Value)
	})
	inner := state.Height - 5
	if inner < 1 {
		inner = 1
	}
	start, end := viewchrome.ListWindow(selectedVisibleIdx, len(visibleRows), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		row := visibleRows[i]
		if row.Header {
			if row.Danger {
				lines = append(lines, theme.StyleError.Render(strings.ToUpper(row.Text)))
			} else {
				lines = append(lines, theme.StyleHeader.Render(strings.ToUpper(row.Text)))
			}
			continue
		}
		switch {
		case row.SelectableAt == cur && active:
			if row.Danger {
				lines = append(lines, theme.StyleError.Render("▶ "+row.Text))
			} else {
				lines = append(lines, theme.StyleCursor.Render("▶ "+row.Text))
			}
		case row.SelectableAt == cur:
			if row.Danger {
				lines = append(lines, theme.StyleError.Render("  "+row.Text))
			} else {
				lines = append(lines, theme.StyleSelected.Render("  "+row.Text))
			}
		default:
			if row.Danger {
				lines = append(lines, theme.StyleError.Render("  "+row.Text))
			} else {
				lines = append(lines, theme.StyleNormal.Render("  "+row.Text))
			}
		}
	}
	if remaining := len(visibleRows) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}
