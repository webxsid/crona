package momentumview

import (
	"fmt"
	"math"
	"strings"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// brailleRows are a cell's dot pairs from the bottom up; each pair spans both dot columns.
var brailleRows = []rune{0x40 | 0x80, 0x04 | 0x20, 0x02 | 0x10, 0x01 | 0x08}

// brailleColumn is a cell filled with dots rows from the bottom, 0 to 4.
func brailleColumn(dots int) string {
	bits := rune(0)
	for i := range min(4, max(0, dots)) {
		bits |= brailleRows[i]
	}
	return string(0x2800 + bits)
}

// statusRank orders bucket statuses by how much they need attention.
var statusRank = map[string]int{"skipped": 0, "met": 1, "near": 2, "missed": 3}

// bucketGroup is one chart column: the buckets it covers, their mean, and the worst status.
type bucketGroup struct {
	count, target int
	status        string
	first, last   int
}

// groupBuckets fits the series into at most cols columns, merging neighbours when crowded.
func groupBuckets(series []sharedtypes.MomentumSeriesPoint, cols int, enabled bool) []bucketGroup {
	n := len(series)
	cols = max(1, min(cols, n))
	groups := make([]bucketGroup, 0, cols)
	for c := range cols {
		lo, hi := c*n/cols, (c+1)*n/cols
		group := bucketGroup{first: lo, last: hi - 1, status: "skipped"}
		counted := 0
		for _, point := range series[lo:hi] {
			if point.Skipped {
				continue
			}
			if status := momentumStatusForPoint(point, enabled); counted == 0 || statusRank[status] > statusRank[group.status] {
				group.status = status
			}
			group.count += point.Count
			group.target += point.Target
			counted++
		}
		if counted > 0 {
			group.count = int(math.Round(float64(group.count) / float64(counted)))
			group.target = int(math.Round(float64(group.target) / float64(counted)))
		}
		if !enabled {
			group.status = "paused"
		}
		groups = append(groups, group)
	}
	return groups
}

// renderMomentumBars is a braille bar chart at exactly height lines: bars, axis, then dates.
// Bars rise in quarter-cell steps, coloured by status; the target is a dotted line.
func renderMomentumBars(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	width, height int,
	enabled bool,
) []string {
	if len(series) == 0 || width < 1 || height < 1 {
		return []string{theme.StyleDim.Render("no data")}
	}
	chartRows := max(1, height-2)
	axisMax := momentumSeriesScale(series)
	_, tickLabels := momentumChartTicks(axisMax, chartRows, def)
	labelWidth := 0
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) == sharedtypes.MomentumTargetKindContext {
		axis := momentumContextAxis(axisMax)
		axisMax = axis.MaxAxis
		_, tickLabels = momentumContextTickRows(axisMax, chartRows, axis.Step, axis.Format)
	}
	for _, label := range tickLabels {
		labelWidth = max(labelWidth, len(label))
	}
	chartWidth := max(4, width-labelWidth-2)
	groups := groupBuckets(series, chartWidth, enabled)
	starts, barWidth := barColumns(len(groups), chartWidth)
	dots := func(value int) int {
		return int(math.Round(float64(value) / float64(max(1, axisMax)) * float64(chartRows*4)))
	}

	grid := make([][]string, chartRows)
	for row := range grid {
		grid[row] = make([]string, chartWidth)
		for col := range grid[row] {
			grid[row][col] = " "
		}
	}
	if target, ok := momentumSharedTarget(series); ok && target > 0 {
		level := min(chartRows*4, max(1, dots(target)))
		row, within := chartRows-1-(level-1)/4, (level-1)%4
		line := lipgloss.NewStyle().Foreground(theme.ColorYellow).Render(string(0x2800 + brailleRows[within]))
		for col := range grid[row] {
			grid[row][col] = line
		}
	}
	seriesX := make([]int, len(series))
	for i, group := range groups {
		for idx := group.first; idx <= group.last; idx++ {
			seriesX[idx] = starts[i] + barWidth/2
		}
		style := momentumStatusStyle(theme, group.status)
		if !enabled {
			style = theme.StyleDim
		}
		height := min(chartRows*4, dots(group.count))
		for row := range chartRows {
			fill := height - (chartRows-1-row)*4
			if fill <= 0 {
				continue
			}
			for col := starts[i]; col < starts[i]+barWidth; col++ {
				grid[row][col] = style.Render(brailleColumn(fill))
			}
		}
	}

	lines := make([]string, 0, chartRows+2)
	for row := range chartRows {
		label := fmt.Sprintf("%*s", labelWidth, tickLabels[row])
		lines = append(lines, theme.StyleDim.Render(label)+" │"+strings.Join(grid[row], ""))
	}
	if height < 3 {
		return lines[:min(len(lines), height)]
	}
	lines = append(lines, strings.Repeat(" ", labelWidth)+" └"+strings.Repeat("─", chartWidth))
	return append(lines, momentumChartXAxis(theme, series, labelWidth+2, chartWidth, seriesX))
}

// maxBarWidth caps a bar's cells, so a handful of buckets still reads as bars.
const maxBarWidth = 6

// barColumns centres n bars in equal slots across width: each bar's first column, and the
// width every bar shares, about 3/5 of a slot.
func barColumns(n, width int) ([]int, int) {
	slot := float64(width) / float64(max(1, n))
	barWidth := max(1, min(maxBarWidth, int(slot*3/5)))
	starts := make([]int, n)
	for i := range starts {
		starts[i] = min(width-barWidth, int(float64(i)*slot+(slot-float64(barWidth))/2))
	}
	return starts, barWidth
}
