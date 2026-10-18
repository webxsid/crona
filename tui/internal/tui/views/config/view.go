package config

import (
	"strings"

	configitems "crona/tui/internal/tui/configitems"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("config", true, renderView),
		},
		RenderFunc: renderView,
	}
}

// renderView lays Config out as a grouped list beside the selected item's detail.
func renderView(theme types.Theme, state types.ContentState) string {
	indices := viewhelpers.FilteredStrings(configItems(state.ExportAssets), state.Filters["config"])
	cursor := min(max(0, state.Cursors["config"]), max(0, len(indices)-1))
	listWidth, detailWidth := viewchrome.LogPaneWidths(state.Width)
	items, groups := []configitems.Item{}, []string{}
	for _, section := range configitems.BuildSections(state.ExportAssets) {
		for _, item := range section.Items {
			items = append(items, item)
			groups = append(groups, section.Title)
		}
	}

	var body []string
	switch {
	case state.ExportAssets == nil:
		body = []string{theme.StyleDim.Render("Loading…")}
	case len(indices) == 0:
		body = []string{theme.StyleDim.Render("No config items match the current filter")}
	default:
		labels := make([]string, 0, len(indices))
		for _, idx := range indices {
			labels = append(labels, items[idx].Label)
		}
		column := viewchrome.LabelColumn(labels, 24)
		rows := make([]viewchrome.GroupRow, 0, len(indices))
		for _, idx := range indices {
			rows = append(rows, viewchrome.GroupRow{
				Group: groups[idx], Line: viewchrome.LabelCell(items[idx].Label, column) + "  " + items[idx].Value,
			})
		}
		body = viewchrome.GroupList(theme, rows, cursor, listWidth-1, state.Height-1)
	}
	list := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Config", Meta: viewchrome.KeyHint(theme, "R", "rescan tools"), Focused: true,
		Width: listWidth, Height: state.Height, Lines: body,
	})
	if detailWidth == 0 {
		return list
	}
	title, lines := "Detail", []string{}
	if state.ExportAssets != nil && len(indices) > 0 {
		item := items[indices[cursor]]
		title = item.Label
		lines = viewchrome.HelpLines(theme, viewchrome.RowHelp{
			Description: strings.TrimSpace(item.DetailBody), Keys: itemKeys(item),
		}, "", detailWidth-2)
	}
	detail := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: title, Width: detailWidth, Height: state.Height, Lines: lines,
	})
	return viewchrome.JoinColumns(theme, state.Height, list, detail)
}

func itemKeys(item configitems.Item) []string {
	keys := []string{}
	if item.Editable {
		keys = append(keys, "[e] edit or open")
	}
	if item.Mutable {
		keys = append(keys, "[c/space] change")
	}
	if item.Resettable {
		keys = append(keys, "[r] reset")
	}
	return keys
}
