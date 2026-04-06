package config

import (
	"fmt"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
)

type visibleRow struct {
	Header       bool
	Text         string
	SelectableAt int
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

func groupedConfigRows(indices []int, selected int, sectionByIndex map[int]string, items []string) ([]visibleRow, int) {
	rows := make([]visibleRow, 0, len(indices)+4)
	lastSection := ""
	selectedVisible := 0
	for i, idx := range indices {
		section := sectionByIndex[idx]
		if section != "" && section != lastSection {
			rows = append(rows, visibleRow{Header: true, Text: section, SelectableAt: -1})
			lastSection = section
		}
		rows = append(rows, visibleRow{Text: items[idx], SelectableAt: i})
		if i == selected {
			selectedVisible = len(rows) - 1
		}
	}
	return rows, selectedVisible
}
