package views

import (
	"fmt"
	"path/filepath"

	"crona/tui/internal/api"
)

func renderReportsView(theme Theme, state ContentState) string {
	active := state.Pane == "export_reports"
	cur := state.Cursors["export_reports"]
	items := exportReportItems(state.ExportReports)
	indices := filteredStrings(items, state.Filters["export_reports"])
	total := len(indices)
	actionLine := renderPaneActionLine(theme, state.Filters["export_reports"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render("Reports"), actionLine}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export configuration..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	lines = append(lines, theme.StyleDim.Render("Dir: "+state.ExportAssets.ReportsDir))
	if total == 0 {
		lines = append(lines, "")
		lines = append(lines, theme.StyleDim.Render("No exported reports found"))
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

func exportReportItems(reports []api.ExportReportFile) []string {
	items := make([]string, 0, len(reports))
	for _, report := range reports {
		scope := report.ScopeLabel
		if scope == "" {
			scope = "-"
		}
		dateLabel := report.DateLabel
		if dateLabel == "" {
			dateLabel = report.Date
		}
		items = append(items, fmt.Sprintf("[%s] %s    %s    [%s] %s    %d B", report.Kind, scope, dateLabel, report.Format, filepath.Base(report.Path), report.SizeBytes))
	}
	return items
}
