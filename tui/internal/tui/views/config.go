package views

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
)

type configVisibleRow struct {
	Header       bool
	Text         string
	SelectableAt int
}

func renderConfigView(theme Theme, state ContentState) string {
	active := state.Pane == "config"
	cur := state.Cursors["config"]
	items := configItems(state.ExportAssets)
	indices := filteredStrings(items, state.Filters["config"])
	total := len(indices)
	actionLine := renderPaneActionLine(theme, state.Filters["config"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render("Report Config"), actionLine}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export assets..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No config items match the current filter"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
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
	inner := remainingPaneHeight(state.Height, lines)
	start, end := listWindow(selectedVisibleIdx, len(visibleRows), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		row := visibleRows[i]
		if row.Header {
			lines = append(lines, theme.StyleHeader.Render(strings.ToUpper(row.Text)))
			continue
		}
		lines = append(lines, renderPaneRowStyled(theme, row.SelectableAt, cur, active, row.Text, nil, state.Width))
	}
	if remaining := len(visibleRows) - end; remaining > 0 {
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

func groupedConfigRows(indices []int, selected int, sectionByIndex map[int]string, items []string) ([]configVisibleRow, int) {
	rows := make([]configVisibleRow, 0, len(indices)+4)
	lastSection := ""
	selectedVisible := 0
	for i, idx := range indices {
		section := sectionByIndex[idx]
		if section != "" && section != lastSection {
			rows = append(rows, configVisibleRow{Header: true, Text: section, SelectableAt: -1})
			lastSection = section
		}
		rows = append(rows, configVisibleRow{Text: items[idx], SelectableAt: i})
		if i == selected {
			selectedVisible = len(rows) - 1
		}
	}
	return rows, selectedVisible
}
