package momentumview

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

const (
	// pairedHeaderWidth is the narrowest detail that sets the settings side by side.
	pairedHeaderWidth = 60
	// minChartRows is the shortest chart the medium layout keeps above its buckets.
	minChartRows = 8
	// maxBucketLabel fits a week's label, "[25] Jun 22-28".
	maxBucketLabel = 14
)

// DetailLines is the momentum's settings, then its chart, which fills the room left; with
// buckets, the chart takes 55% of that room and the bucket table the rest.
func DetailLines(theme types.Theme, card sharedtypes.MomentumCard, width, height int, buckets bool) []string {
	lines := append(headerLines(theme, card, width), "")
	room := height - len(lines)
	chartRows := room
	if buckets {
		chartRows = max(minChartRows, room*55/100)
		if room-chartRows < 3 {
			chartRows, buckets = room, false
		}
	}
	lines = append(lines, chartLines(theme, card, width, chartRows)...)
	if buckets {
		lines = append(lines, "", theme.StylePaneTitle.Render("Buckets"))
		lines = append(lines, BucketLines(theme, card, width)...)
	}
	return lines[:min(len(lines), max(0, height))]
}

func headerLines(theme types.Theme, card sharedtypes.MomentumCard, width int) []string {
	def := card.Definition
	status := latestStatus(card)
	statusText := "–"
	if status != "" {
		statusText = momentumStatusStyle(theme, status).Render(status)
	}
	left := [][2]string{
		{"Cadence", momentumhelpers.CadenceLabel(def.Period)},
		{"Target", momentumTargetSummary(def)},
		{"Streak", streakLabel(card, card.Current) + " now · best " + streakLabel(card, card.Longest)},
	}
	right := [][2]string{
		{momentumCardTargetLabel(def), momentumTargetSummaryNames(card)},
		{"Status", statusText},
	}
	lines := []string{}
	if width < pairedHeaderWidth {
		for _, row := range append(left, right...) {
			lines = append(lines, viewsummary.Row(theme, row[0], row[1], width))
		}
	} else {
		// The left column takes what its values need, up to 3/5, and the right the rest.
		leftWidth := 0
		for _, row := range left {
			leftWidth = max(leftWidth, ansi.StringWidth(viewsummary.Row(theme, row[0], row[1], width)))
		}
		leftWidth = min(leftWidth, width*3/5)
		for i, row := range left {
			line := viewsummary.Row(theme, row[0], row[1], leftWidth)
			if i < len(right) {
				line += strings.Repeat(" ", max(0, leftWidth-ansi.StringWidth(line))+2) + viewsummary.Row(theme, right[i][0], right[i][1], width-leftWidth-2)
			}
			lines = append(lines, line)
		}
	}
	if def.Description != nil && strings.TrimSpace(*def.Description) != "" {
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*def.Description), max(1, width), ""), "\n") {
			lines = append(lines, theme.StyleDim.Render(line))
		}
	}
	return lines
}

// chartLines is the chart and the latest-bucket footnote at exactly height lines.
func chartLines(theme types.Theme, card sharedtypes.MomentumCard, width, height int) []string {
	if height <= 0 {
		return nil
	}
	def := card.Definition
	footnote := theme.StyleDim.Render(viewhelpers.Truncate(momentumSeriesFootnote(def, card.Series, def.Enabled), width))
	rows := height - 1
	var chart []string
	switch {
	case rows <= 0:
	case sharedtypes.NormalizeHabitStreakPeriod(def.Period) == sharedtypes.HabitStreakPeriodDay && momentumDailySeriesUsesSquares(def, card.Series):
		chart = strings.Split(renderMomentumDailySquares(theme, card.Series, width, def.Enabled), "\n")
	default:
		chart = renderMomentumBars(theme, def, card.Series, width, rows, def.Enabled)
	}
	lines := make([]string, 0, height)
	for _, line := range chart[:min(len(chart), max(0, rows))] {
		lines = append(lines, ansi.Truncate(line, width, ""))
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return append(lines, footnote)
}

// BucketLines lists the momentum's buckets newest first: date, count of target, status,
// a bar on the shared scale, and notes on protected days and prorated targets.
func BucketLines(theme types.Theme, card sharedtypes.MomentumCard, width int) []string {
	def, series := card.Definition, card.Series
	scale := momentumSeriesScale(series)
	labelWidth, ratioWidth := 6, 5
	for _, point := range series {
		labelWidth = max(labelWidth, min(maxBucketLabel, ansi.StringWidth(momentumDisplayLabel(point))))
		ratioWidth = max(ratioWidth, ansi.StringWidth(momentumRatioText(point, def)))
	}
	// The label, ratio, status and gaps leave the bar the rest, 4 to 16 cells.
	barWidth := max(4, min(16, width-labelWidth-ratioWidth-7-6))
	lines := make([]string, 0, len(series))
	for i := len(series) - 1; i >= 0; i-- {
		point := series[i]
		status := momentumStatusForPoint(point, def.Enabled)
		row := padRight(viewhelpers.Truncate(momentumDisplayLabel(point), labelWidth), labelWidth) + "  " +
			padRight(momentumRatioText(point, def), ratioWidth) + "  " +
			momentumStatusStyle(theme, status).Render(padRight(status, 7)) + "  " +
			momentumBucketBar(theme, point, barWidth, def.Enabled, scale)
		if notes := bucketNotes(point, def); notes != "" {
			row += "  " + theme.StyleDim.Render(notes)
		}
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	return lines
}

func bucketNotes(point sharedtypes.MomentumSeriesPoint, def sharedtypes.HabitStreakDefinition) string {
	notes := []string{}
	if point.ProtectedDayCount > 0 {
		notes = append(notes, fmt.Sprintf("protected %dd", point.ProtectedDayCount))
	}
	if point.OriginalTarget > 0 && point.OriginalTarget != point.Target {
		notes = append(notes, "prorated from "+momentumValueDisplay(point.OriginalTarget, def))
	}
	return strings.Join(notes, " · ")
}
