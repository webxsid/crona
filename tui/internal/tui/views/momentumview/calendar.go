package momentumview

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"
)

const (
	// calendarLabelWidth fits a week label, "W42", and its gap.
	calendarLabelWidth = 5
	// maxCalendarCellWidth stops a few weeks from turning into slabs on very wide panes.
	maxCalendarCellWidth = 14
)

var calendarWeekdays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// calendarWeek is one ISO week's days, Monday first; nil days fall outside the series.
type calendarWeek struct {
	label string
	days  [7]*sharedtypes.MomentumSeriesPoint
}

// calendarLines draws a binary daily momentum as weeks by weekday at exactly height lines:
// a weekday header, square cells sized to the room, the latest weeks kept, each weekday's
// hits right under the grid, then blank lines.
func calendarLines(theme types.Theme, series []sharedtypes.MomentumSeriesPoint, width, height int, enabled bool) []string {
	weeks := calendarWeeks(series)
	if len(weeks) == 0 || height < 3 {
		return padLines([]string{theme.StyleDim.Render("no data")}, height)
	}
	cellWidth := max(2, min(maxCalendarCellWidth, (width-calendarLabelWidth)/7-1))
	room := height - 2
	// Terminal cells are about twice as tall as wide, so half the width keeps cells square.
	cellHeight := max(1, min(cellWidth/2, room/len(weeks)-1))
	block := cellHeight
	if cellHeight > 1 {
		block++ // a blank line between weeks
	}
	weeks = weeks[max(0, len(weeks)-max(1, room/block)):]

	column := func(s string) string { return padRight(s, cellWidth) + " " }
	header := strings.Repeat(" ", calendarLabelWidth)
	for _, day := range calendarWeekdays {
		header += column(day)
	}
	lines := []string{theme.StyleDim.Render(strings.TrimRight(header, " "))}
	for _, week := range weeks {
		for row := range cellHeight {
			label := ""
			if row == 0 {
				label = week.label
			}
			line := theme.StyleDim.Render(padRight(label, calendarLabelWidth))
			for _, point := range week.days {
				line += calendarCell(theme, point, cellWidth, enabled) + " "
			}
			lines = append(lines, strings.TrimRight(line, " "))
		}
		if block > cellHeight {
			lines = append(lines, "")
		}
	}
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	lines = lines[:min(len(lines), height-1)]

	hit := padRight("hit", calendarLabelWidth)
	for weekday := range 7 {
		done, total := 0, 0
		for _, week := range calendarWeeks(series) {
			if point := week.days[weekday]; point != nil && !point.Skipped {
				total++
				if momentumStatusForPoint(*point, true) == "met" {
					done++
				}
			}
		}
		hit += column(fmt.Sprintf("%d/%d", done, total))
	}
	return padLines(append(lines, theme.StyleDim.Render(strings.TrimRight(hit, " "))), height)
}

func calendarCell(theme types.Theme, point *sharedtypes.MomentumSeriesPoint, width int, enabled bool) string {
	switch {
	case point == nil:
		return strings.Repeat(" ", width)
	case point.Skipped:
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	status := momentumStatusForPoint(*point, enabled)
	glyph := "█"
	if status != "met" {
		glyph = "░"
	}
	return momentumStatusStyle(theme, status).Render(strings.Repeat(glyph, width))
}

// calendarWeeks groups the series by ISO week, oldest first.
func calendarWeeks(series []sharedtypes.MomentumSeriesPoint) []calendarWeek {
	weeks := []calendarWeek{}
	for i := range series {
		day, err := time.Parse(isoDate, series[i].BucketKey)
		if err != nil {
			continue
		}
		_, w := day.ISOWeek()
		label := fmt.Sprintf("W%d", w)
		if len(weeks) == 0 || weeks[len(weeks)-1].label != label {
			weeks = append(weeks, calendarWeek{label: label})
		}
		weeks[len(weeks)-1].days[(int(day.Weekday())+6)%7] = &series[i]
	}
	return weeks
}

func padLines(lines []string, height int) []string {
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}
