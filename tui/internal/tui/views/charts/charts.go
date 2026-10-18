// Package charts draws braille chart pieces shared by the Insights views.
package charts

import (
	"math"
	"strings"
)

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

// Line draws values as a braille line across rows lines of width cells: each value sits at
// its column in xs, scaled between lo and hi; neighbours are joined and a NaN leaves a gap.
func Line(values []float64, xs []int, lo, hi float64, width, rows int) []string {
	if width < 1 || rows < 1 {
		return nil
	}
	pw, ph := width*2, rows*4
	pixels := make([][]bool, ph)
	for y := range pixels {
		pixels[y] = make([]bool, pw)
	}
	span := hi - lo
	if span <= 0 {
		span = 1
	}
	point := func(i int) (int, int) {
		norm := math.Min(1, math.Max(0, (values[i]-lo)/span))
		return min(pw-1, xs[i]*2), int(math.Round((1 - norm) * float64(ph-1)))
	}
	for i := range values {
		if math.IsNaN(values[i]) {
			continue
		}
		x, y := point(i)
		set(pixels, x, y)
		set(pixels, x+1, y)
		if i+1 < len(values) && !math.IsNaN(values[i+1]) {
			nx, ny := point(i + 1)
			line(pixels, x+1, y, nx, ny)
		}
	}
	out := make([]string, rows)
	for row := range rows {
		var b strings.Builder
		for col := range width {
			bits := rune(0)
			for dy := range 4 {
				for dx := range 2 {
					if pixels[row*4+dy][col*2+dx] {
						bits |= brailleBit(dx, dy)
					}
				}
			}
			if bits == 0 {
				b.WriteRune(' ')
			} else {
				b.WriteRune(0x2800 + bits)
			}
		}
		out[row] = b.String()
	}
	return out
}

// brailleBit is the dot for column dx (0 left) and row dy (0 top) of a cell.
func brailleBit(dx, dy int) rune {
	if dy == 3 {
		return []rune{0x40, 0x80}[dx]
	}
	return rune(1<<dy) << (3 * dx)
}

func set(pixels [][]bool, x, y int) {
	if y >= 0 && y < len(pixels) && x >= 0 && x < len(pixels[y]) {
		pixels[y][x] = true
	}
}

// line sets the pixels between two points (Bresenham).
func line(pixels [][]bool, x0, y0, x1, y1 int) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx + dy
	for {
		set(pixels, x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(v int) int { return max(v, -v) }

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
