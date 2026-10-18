package settings

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("settings", true, renderView),
		},
	}
}

// renderView lays Settings out as a grouped list beside the selected setting's help.
func renderView(theme types.Theme, state types.ContentState) string {
	timezone := ""
	if state.Health != nil {
		timezone = state.Health.Timezone
	}
	rows := settingsmeta.RowsWithTimezone(state.Settings, timezone)
	indices := settingsmeta.FilteredIndices(state.Filters["settings"], state.Settings)
	cursor := min(max(0, state.Cursors["settings"]), max(0, len(indices)-1))
	listWidth, detailWidth := viewchrome.LogPaneWidths(state.Width)

	var body []string
	switch {
	case state.Settings == nil:
		body = []string{theme.StyleDim.Render("Loading settings…")}
	case len(indices) == 0:
		body = []string{theme.StyleDim.Render("No settings match the current filter")}
	default:
		labels := make([]string, 0, len(indices))
		for _, idx := range indices {
			labels = append(labels, rows[idx].Label)
		}
		column := viewchrome.LabelColumn(labels, 24)
		groupRows := make([]viewchrome.GroupRow, 0, len(indices))
		for _, idx := range indices {
			row := rows[idx]
			groupRows = append(groupRows, viewchrome.GroupRow{
				Group: row.Section, Danger: row.Section == "Danger",
				Line: viewchrome.LabelCell(row.Label, column) + "  " + row.Value,
			})
		}
		body = viewchrome.GroupList(theme, groupRows, cursor, listWidth-1, state.Height-1)
	}
	list := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Settings", Meta: fmt.Sprint(len(indices)), Focused: true,
		Width: listWidth, Height: state.Height, Lines: body,
	})
	if detailWidth == 0 {
		return list
	}
	title, lines := "Detail", []string{}
	if len(indices) > 0 && state.Settings != nil {
		row := rows[indices[cursor]]
		title = row.Label
		lines = viewchrome.HelpLines(theme, settingsHelp[row.Label], strings.TrimSpace(row.Value), detailWidth-2)
	}
	detail := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: title, Width: detailWidth, Height: state.Height, Lines: lines,
	})
	return viewchrome.JoinColumns(theme, state.Height, list, detail)
}
