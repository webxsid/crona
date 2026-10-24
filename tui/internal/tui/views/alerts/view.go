package alerts

import (
	"fmt"
	"strings"

	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("alerts", true, renderView),
		},
		RenderFunc: renderView,
	}
}

// renderView lays Alerts out as a backend strip over the editable rows and their help.
func renderView(theme types.Theme, state types.ContentState) string {
	filter := state.Filters["alerts"]
	rows := alertsmeta.FilteredSelectableRows(
		filter,
		state.Settings,
		state.AlertStatus,
		state.AlertReminders,
	)
	cursor := min(max(0, state.Cursors["alerts"]), max(0, len(rows)-1))
	strip := viewchrome.StripRows(
		backendGroups(alertsmeta.Rows(state.Settings, state.AlertStatus, state.AlertReminders)),
		state.Width,
	)
	height := max(1, state.Height-len(strip)-1)
	listWidth, detailWidth := viewchrome.LogPaneWidths(state.Width)

	var body []string
	switch {
	case state.Settings == nil:
		body = []string{theme.StyleDim.Render("Loading alert settings…")}
	case len(rows) == 0:
		body = []string{theme.StyleDim.Render("No alert rows match the current filter")}
	default:
		labels := make([]string, 0, len(rows))
		for _, row := range rows {
			labels = append(labels, row.Label)
		}
		column := viewchrome.LabelColumn(labels, 24)
		groupRows := make([]viewchrome.GroupRow, 0, len(rows))
		for _, row := range rows {
			groupRows = append(groupRows, viewchrome.GroupRow{
				Group: row.Section, Line: viewchrome.LabelCell(row.Label, column) + "  " + row.Value,
			})
		}
		body = viewchrome.GroupList(theme, groupRows, cursor, listWidth-1, height-1)
	}
	panes := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Alerts", Meta: fmt.Sprint(len(rows)), Focused: true,
		Width: listWidth, Height: height, Lines: body,
	})
	if detailWidth > 0 {
		title, lines := "Detail", []string{}
		if len(rows) > 0 {
			row := rows[cursor]
			title = row.Label
			lines = viewchrome.HelpLines(theme, helpFor(row), row.Value, detailWidth-2)
		}
		detail := viewchrome.RenderSection(theme, viewchrome.Section{
			Title: title, Width: detailWidth, Height: height, Lines: lines,
		})
		panes = viewchrome.JoinColumns(theme, height, panes, detail)
	}
	return strings.Join(append(strip, "", panes), "\n")
}

// backendGroups pairs the read-only Backend rows into strip columns, in row order.
func backendGroups(rows []alertsmeta.Row) [][2]string {
	cells := []string{}
	for _, row := range rows {
		if row.Section == "Backend" {
			cells = append(cells, row.Label+"  "+row.Value)
		}
	}
	groups := [][2]string{}
	for i := 0; i < len(cells); i += 2 {
		group := [2]string{cells[i], ""}
		if i+1 < len(cells) {
			group[1] = cells[i+1]
		}
		groups = append(groups, group)
	}
	return groups
}
