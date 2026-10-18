package momentumview

import (
	"math"
	"strings"

	sharedtypes "crona/shared/types"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

const (
	sparkCells = 8
	markWidth  = 6
)

var sparkLevels = []rune("▁▂▃▄▅▆▇█")

// sparkline draws the last buckets as bars scaled by count against target, coloured by status.
func sparkline(theme types.Theme, series []sharedtypes.MomentumSeriesPoint, enabled bool) string {
	series = series[max(0, len(series)-sparkCells):]
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", sparkCells-len(series)))
	for _, point := range series {
		status := momentumStatusForPoint(point, enabled)
		if point.Skipped {
			b.WriteString(theme.StyleDim.Render("·"))
			continue
		}
		ratio := 1.0
		if point.Target > 0 {
			ratio = min(1, float64(point.Count)/float64(point.Target))
		}
		cell := string(sparkLevels[int(math.Round(ratio*float64(len(sparkLevels)-1)))])
		b.WriteString(momentumStatusStyle(theme, status).Render(cell))
	}
	return b.String()
}

// latestStatus is the status of the momentum's latest bucket, or "" with no buckets.
func latestStatus(card sharedtypes.MomentumCard) string {
	if !card.Definition.Enabled {
		return "paused"
	}
	if len(card.Series) == 0 {
		return ""
	}
	return momentumStatusForPoint(card.Series[len(card.Series)-1], true)
}

// statusMarks are the one-glyph forms of bucket statuses; paused stays a word.
var statusMarks = map[string]string{"met": "✓", "near": "◐", "missed": "✗", "skipped": "·", "paused": "paused"}

func statusMark(theme types.Theme, card sharedtypes.MomentumCard) string {
	status := latestStatus(card)
	mark := statusMarks[status]
	if mark == "" {
		return ""
	}
	return momentumStatusStyle(theme, status).Render(mark)
}

func streakLabel(card sharedtypes.MomentumCard, value int) string {
	return momentumhelpers.FormatLength(value, momentumhelpers.Unit(card.Definition.Period))
}

// listRow is a momentum's name, sparkline, current streak and latest mark, exactly width cells.
func listRow(theme types.Theme, card sharedtypes.MomentumCard, width int) string {
	nameWidth := max(1, width-sparkCells-markWidth-9)
	name := padRight(ansi.Truncate(card.Definition.Name, nameWidth, "…"), nameWidth)
	streak := streakLabel(card, card.Current)
	streak = strings.Repeat(" ", max(0, 4-len(streak))) + streak
	mark := statusMark(theme, card)
	mark += strings.Repeat(" ", max(0, markWidth-ansi.StringWidth(mark)))
	if !card.Definition.Enabled {
		name, streak = theme.StyleDim.Render(name), theme.StyleDim.Render(streak)
	}
	row := name + "  " + sparkline(theme, card.Series, card.Definition.Enabled) + "  " + streak + " " + mark
	row = ansi.Truncate(row, width, "")
	return row + strings.Repeat(" ", max(0, width-ansi.StringWidth(row)))
}
