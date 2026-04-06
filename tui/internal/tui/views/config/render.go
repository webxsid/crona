package config

import (
	"fmt"
	"strings"

	configitems "crona/tui/internal/tui/configitems"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func Render(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "config"
	cur := state.Cursors["config"]
	items := configItems(state.ExportAssets)
	indices := viewhelpers.FilteredStrings(items, state.Filters["config"])
	total := len(indices)
	lines := []string{
		theme.StylePaneTitle.Render("Report Config"),
		viewchrome.RenderPaneActionLine(theme, state.Filters["config"], state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})),
	}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export assets..."))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No config items match the current filter"))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	sections := configitems.BuildSections(state.ExportAssets)
	sectionByIndex := make(map[int]string, len(items))
	rawIdx := 0
	for _, section := range sections {
		for range section.Items {
			sectionByIndex[rawIdx] = section.Title
			rawIdx++
		}
	}
	visibleRows, selectedVisibleIdx := groupedConfigRows(indices, cur, sectionByIndex, items)
	inner := viewchrome.RemainingPaneHeight(state.Height, lines)
	start, end := viewchrome.ListWindow(selectedVisibleIdx, len(visibleRows), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		row := visibleRows[i]
		if row.Header {
			lines = append(lines, theme.StyleHeader.Render(strings.ToUpper(row.Text)))
			continue
		}
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, row.SelectableAt, cur, active, row.Text, nil, state.Width))
	}
	if remaining := len(visibleRows) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}
