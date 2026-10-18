package momentumview

import (
	"fmt"
	"strings"
	"time"

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
	// minBucketBar is the shortest bar a row keeps before it drops its notes.
	minBucketBar = 12
	isoDate      = "2006-01-02"
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
		chart = calendarLines(theme, card.Series, width, rows, def.Enabled)
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

// BucketLines is a summary line, then the momentum's buckets newest first: a status mark,
// the bucket, its count and a bar that fills the width. The target is stated once when it
// is shared, so only a bucket with its own target repeats it; daily buckets break by week.
func BucketLines(theme types.Theme, card sharedtypes.MomentumCard, width int) []string {
	def, series := card.Definition, card.Series
	shared, hasShared := momentumSharedTarget(series)
	daily := sharedtypes.NormalizeHabitStreakPeriod(def.Period) == sharedtypes.HabitStreakPeriodDay
	lines := []string{theme.StyleDim.Render(ansi.Truncate(bucketSummary(def, series, shared, hasShared), width, "…"))}

	labels := make([]string, len(series))
	counts := make([]string, len(series))
	labelWidth, countWidth := 0, 0
	for i, point := range series {
		labels[i] = bucketLabel(point, daily)
		counts[i] = bucketCount(point, def, shared, hasShared)
		labelWidth = max(labelWidth, min(maxBucketLabel, ansi.StringWidth(labels[i])))
		countWidth = max(countWidth, ansi.StringWidth(counts[i]))
	}
	scale := momentumSeriesScale(series)
	week := ""
	for i := len(series) - 1; i >= 0; i-- {
		point := series[i]
		if daily {
			if day, err := time.Parse(isoDate, point.BucketKey); err == nil {
				_, w := day.ISOWeek()
				if label := fmt.Sprintf("W%d", w); label != week {
					week = label
					rule := "── " + label + " "
					lines = append(lines, theme.StyleDim.Render(rule+strings.Repeat("─", max(0, width-ansi.StringWidth(rule)))))
				}
			}
		}
		status := momentumStatusForPoint(point, def.Enabled)
		row := momentumStatusStyle(theme, status).Render(statusMarks[status]) + " " +
			padRight(ansi.Truncate(labels[i], labelWidth, "…"), labelWidth) + "  " +
			padRight(counts[i], countWidth) + "  "
		room := max(1, width-ansi.StringWidth(row))
		notes := bucketNotes(point, def)
		switch {
		case point.Skipped:
			row += theme.StyleDim.Render(padRight(notes, room))
		case notes != "" && room-len(notes)-2 >= minBucketBar:
			row += momentumBucketBar(theme, point, room-len(notes)-2, def.Enabled, scale) + "  " + theme.StyleDim.Render(notes)
		default:
			row += momentumBucketBar(theme, point, room, def.Enabled, scale)
		}
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	return lines
}

// bucketSummary is the table's opening line: the shared target, buckets met, and the mean.
func bucketSummary(def sharedtypes.HabitStreakDefinition, series []sharedtypes.MomentumSeriesPoint, shared int, hasShared bool) string {
	met, counted, total := 0, 0, 0
	for _, point := range series {
		if point.Skipped {
			continue
		}
		counted++
		total += point.Count
		if momentumStatusForPoint(point, true) == "met" {
			met++
		}
	}
	parts := []string{}
	if !def.Enabled {
		parts = append(parts, "paused")
	}
	if hasShared {
		parts = append(parts, "target "+momentumValueDisplay(shared, def))
	}
	if def.Enabled {
		parts = append(parts, fmt.Sprintf("met %d of %d", met, counted))
	}
	if counted > 0 {
		parts = append(parts, "avg "+momentumValueDisplay(total/counted, def))
	}
	return strings.Join(parts, " · ")
}

// bucketLabel is a daily bucket's weekday and day, or the bucket's own label.
func bucketLabel(point sharedtypes.MomentumSeriesPoint, daily bool) string {
	if daily {
		if day, err := time.Parse(isoDate, point.BucketKey); err == nil {
			return day.Format("Mon 2")
		}
	}
	return momentumDisplayLabel(point)
}

// bucketCount is the bucket's count, with its target only when it differs from the shared one.
func bucketCount(point sharedtypes.MomentumSeriesPoint, def sharedtypes.HabitStreakDefinition, shared int, hasShared bool) string {
	if point.Skipped {
		return "skip"
	}
	if hasShared && point.Target == shared {
		return momentumValueDisplay(point.Count, def)
	}
	return momentumRatioText(point, def)
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
