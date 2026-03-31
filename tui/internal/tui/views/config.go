package views

import (
	"fmt"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
)

func renderConfigView(theme Theme, state ContentState) string {
	active := state.Pane == "config"
	cur := state.Cursors["config"]
	items := configItems(state.ExportAssets)
	indices := filteredStrings(items, state.Filters["config"])
	total := len(indices)
	actionLine := renderPaneActionLine(theme, state.Filters["config"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render("Config"), actionLine}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export assets..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No config items match the current filter"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	inner := remainingPaneHeight(state.Height, lines)
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(lines, renderPaneRowStyled(theme, i, cur, active, items[indices[i]], nil, state.Width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}

func configItems(status *api.ExportAssetStatus) []string {
	if status == nil {
		return nil
	}
	built := configitems.Build(status)
	items := make([]string, 0, len(built))
	for _, item := range built {
		items = append(items, fmt.Sprintf("%-24s %s", item.Label, item.Value))
	}
	return items
}
