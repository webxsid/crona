// Package charts draws braille chart pieces shared by the Insights views.
package charts

import "math"

// dotRows are a cell's dot pairs from the bottom up; each pair spans both dot columns.
var dotRows = []rune{0x40 | 0x80, 0x04 | 0x20, 0x02 | 0x10, 0x01 | 0x08}

// maxBarWidth caps a bar's cells, so a handful of bars still reads as bars.
const maxBarWidth = 6

// Column is a cell filled with dots rows from the bottom, 0 to 4.
func Column(dots int) string {
	bits := rune(0)
	for i := range min(4, max(0, dots)) {
		bits |= dotRows[i]
	}
	return string(0x2800 + bits)
}

// DotRow is a cell with only its within-th dot row set, counted from the bottom, 0 to 3;
// a run of them draws a fine horizontal line.
func DotRow(within int) string {
	return string(0x2800 + dotRows[min(3, max(0, within))])
}

// Slots centres n bars in equal slots across width: each bar's first column, and the width
// every bar shares, about 3/5 of a slot.
func Slots(n, width int) ([]int, int) {
	slot := float64(width) / float64(max(1, n))
	barWidth := max(1, min(maxBarWidth, int(slot*3/5)))
	starts := make([]int, n)
	for i := range starts {
		starts[i] = max(0, min(width-barWidth, int(float64(i)*slot+(slot-float64(barWidth))/2)))
	}
	return starts, barWidth
}

var blocks = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// Block is a cell filled from the bottom in eighths, 0 to 8.
func Block(eighths int) string {
	return blocks[min(8, max(0, eighths))]
}

// BarGrid draws values as solid bars, top row first, rows lines of width cells: each value
// fills its slot from starts with barWidth cells, scaled so top fills every row. A NaN
// value leaves a dot on the bottom row; empty cells are spaces.
func BarGrid(values []float64, starts []int, barWidth, width, rows int, top float64) [][]string {
	grid := make([][]string, rows)
	for row := range grid {
		grid[row] = make([]string, width)
		for col := range grid[row] {
			grid[row][col] = " "
		}
	}
	if rows < 1 || top <= 0 {
		return grid
	}
	for i, value := range values {
		if i >= len(starts) {
			break
		}
		if math.IsNaN(value) {
			grid[rows-1][min(width-1, starts[i]+(barWidth-1)/2)] = "·"
			continue
		}
		// A known value always shows: zero draws a one-eighth floor line.
		eighths := max(1, min(rows*8, int(math.Round(value/top*float64(rows*8)))))
		for row := range rows {
			fill := eighths - (rows-1-row)*8
			if fill <= 0 {
				continue
			}
			for col := starts[i]; col < min(width, starts[i]+barWidth); col++ {
				grid[row][col] = Block(fill)
			}
		}
	}
	return grid
}
