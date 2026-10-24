package reports

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// renderLog lays the exported reports out as a strip over a day list and a detail pane.
func renderLog(theme types.Theme, state types.ContentState) string {
	indices := viewhelpers.FilteredStrings(
		reportItems(state.ExportReports),
		state.Filters["export_reports"],
	)
	reports := make([]api.ExportReportFile, 0, len(indices))
	for _, index := range indices {
		reports = append(reports, state.ExportReports[index])
	}
	cursor := min(max(0, state.Cursors["export_reports"]), max(0, len(reports)-1))
	strip := reportsStrip(theme, state, state.Width)
	height := max(1, state.Height-len(strip)-1)
	listWidth, detailWidth := viewchrome.LogPaneWidths(state.Width)

	var body []string
	switch {
	case state.ExportAssets == nil:
		body = []string{theme.StyleDim.Render("Loading…")}
	case len(reports) == 0:
		body = []string{
			theme.StyleDim.Render("No exported reports yet"),
			theme.StyleDim.Render("Export from Insights to see reports here"),
		}
	default:
		rows := make([]viewchrome.DayRow, 0, len(reports))
		counts := map[string]int{}
		for _, report := range reports {
			date := localTime(report.ModifiedAt).Format("2006-01-02")
			counts[date]++
			parts := []string{fmt.Sprintf("%-8s", report.Kind)}
			if report.ScopeLabel != "" {
				parts = append(parts, report.ScopeLabel)
			}
			parts = append(parts, reportPeriod(report), theme.StyleDim.Render(report.Format))
			rows = append(rows, viewchrome.DayRow{Date: date, Line: strings.Join(parts, "  ")})
		}
		body = viewchrome.DayList(
			theme,
			rows,
			cursor,
			listWidth-1,
			height-1,
			func(date string) string { return fmt.Sprint(counts[date]) },
		)
	}
	panes := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Reports", Meta: fmt.Sprint(len(reports)), Focused: true,
		Width: listWidth, Height: height, Lines: body,
	})
	if detailWidth > 0 {
		title, meta, lines := "Detail", "", []string{}
		if state.ExportAssets != nil && len(reports) > 0 {
			report := reports[cursor]
			title = filepath.Base(report.Path)
			meta = "[o] open  [e] edit  [D] delete"
			lines = reportDetailLines(theme, report, detailWidth-2)
		}
		detail := viewchrome.RenderSection(theme, viewchrome.Section{
			Title: title, Meta: meta, Width: detailWidth, Height: height, Lines: lines,
		})
		panes = viewchrome.JoinColumns(theme, height, panes, detail)
	}
	return strings.Join(append(strip, "", panes), "\n")
}

func reportsStrip(theme types.Theme, state types.ContentState, width int) []string {
	newest := ""
	for _, report := range state.ExportReports {
		if report.ModifiedAt > newest {
			newest = report.ModifiedAt
		}
	}
	top := fmt.Sprintf("Reports %d", len(state.ExportReports))
	if newest != "" {
		top += "   newest " + localTime(newest).Format("2006-01-02 15:04")
	}
	dir := "–"
	if state.ExportAssets != nil {
		dir = state.ExportAssets.ReportsDir
	}
	bottom := "Dir " + dir
	if ansi.StringWidth(bottom) > width {
		bottom = "Dir …" + ansi.TruncateLeft(dir, ansi.StringWidth(dir)-max(0, width-5), "")
	}
	pad := func(s string) string {
		s = ansi.Truncate(s, width, "")
		return s + strings.Repeat(" ", max(0, width-ansi.StringWidth(s)))
	}
	return []string{pad(top), theme.StyleDim.Render(pad(bottom))}
}

func reportDetailLines(theme types.Theme, report api.ExportReportFile, width int) []string {
	fields := [][2]string{
		{"Kind", string(report.Kind)},
		{"Period", reportPeriod(report)},
		{"Format", report.Format},
		{"Size", humanSize(report.SizeBytes)},
		{"Created", localTime(report.ModifiedAt).Format("2006-01-02 15:04")},
	}
	if report.ScopeLabel != "" {
		fields = append([][2]string{{"Scope", report.ScopeLabel}}, fields...)
	}
	lines := []string{}
	for _, field := range fields {
		lines = append(
			lines,
			theme.StyleDim.Render(
				fmt.Sprintf("%-8s", field[0]),
			)+" "+ansi.Truncate(
				field[1],
				max(1, width-9),
				"…",
			),
		)
	}
	lines = append(lines, "", theme.StyleDim.Render("Path"))
	return append(lines, strings.Split(ansi.Hardwrap(report.Path, max(1, width), false), "\n")...)
}

func reportPeriod(report api.ExportReportFile) string {
	switch {
	case report.DateLabel != "":
		return report.DateLabel
	case report.Date != "":
		return report.Date
	case report.StartDate != "":
		return report.StartDate + " – " + report.EndDate
	default:
		return "–"
	}
}

func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1000:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1000)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func localTime(timestamp string) time.Time {
	at, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return time.Time{}
	}
	return at.Local()
}
