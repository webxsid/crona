package summary

import (
	"math"
	"strings"

	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	// gaugeCols and gaugeRows size one arc: braille dots are square, so a half circle of
	// radius r dots spans r cells across and r/4 rows up.
	gaugeCols = 13
	gaugeRows = 3
	gaugeGap  = 2
)

// gauge is one metric drawn as a half-circle arc meter.
type gauge struct {
	label, value, trend string
	fraction            float64
	color               lipgloss.Color
}

// brailleBit is the dot bit for a dot column (0-1) and row (0-3) within a braille cell.
var brailleBit = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// gaugeArc draws the arc: dots from the left end clockwise to fraction take color, the rest
// the dim track, and the value sits centred in the arc's open bottom row.
func gaugeArc(theme types.Theme, g gauge) []string {
	fraction := math.Max(0, math.Min(1, g.fraction))
	width, height := float64(gaugeCols*2), float64(gaugeRows*4)
	cx, cy := width/2, height
	outer := math.Min(cx, cy) - 0.4
	inner := outer - 2.8
	fill := lipgloss.NewStyle().Foreground(g.color)
	lines := make([]string, gaugeRows)
	for row := range gaugeRows {
		var b strings.Builder
		for col := range gaugeCols {
			var bits rune
			filled := false
			for dx := range 2 {
				for dy := range 4 {
					px, py := float64(col*2+dx)+0.5, float64(row*4+dy)+0.5
					d := math.Hypot(px-cx, cy-py)
					if d < inner || d > outer {
						continue
					}
					bits |= brailleBit[dx][dy]
					if (math.Pi-math.Atan2(cy-py, px-cx))/math.Pi <= fraction {
						filled = true
					}
				}
			}
			switch {
			case bits == 0:
				b.WriteString(" ")
			case filled:
				b.WriteString(fill.Render(string(0x2800 + bits)))
			default:
				b.WriteString(theme.StyleDim.Render(string(0x2800 + bits)))
			}
		}
		lines[row] = b.String()
	}
	// The value goes in the bottom row between the arc's feet.
	value := theme.StyleNormal.Bold(true).Render(g.value)
	pad := (gaugeCols - ansi.StringWidth(value)) / 2
	bottom := lines[gaugeRows-1]
	lines[gaugeRows-1] = ansi.Cut(bottom, 0, pad) + value +
		ansi.Cut(bottom, pad+ansi.StringWidth(value), gaugeCols)
	return lines
}

// renderGauges lays gauges out side by side, as many per row as width holds: the arc, then
// the label and the trend centred under it.
func renderGauges(theme types.Theme, gauges []gauge, width int) []string {
	perRow := max(1, (width+gaugeGap)/(gaugeCols+gaugeGap))
	center := func(s string) string {
		w := ansi.StringWidth(s)
		if w >= gaugeCols {
			return ansi.Truncate(s, gaugeCols, "")
		}
		left := (gaugeCols - w) / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", gaugeCols-w-left)
	}
	var lines []string
	for start := 0; start < len(gauges); start += perRow {
		group := gauges[start:min(len(gauges), start+perRow)]
		block := make([][]string, gaugeRows+2)
		for _, g := range group {
			arc := gaugeArc(theme, g)
			for i := range gaugeRows {
				block[i] = append(block[i], arc[i])
			}
			block[gaugeRows] = append(block[gaugeRows], center(theme.StyleDim.Render(g.label)))
			block[gaugeRows+1] = append(block[gaugeRows+1], center(g.trend))
		}
		if start > 0 {
			lines = append(lines, "")
		}
		for _, cells := range block {
			line := strings.TrimRight(strings.Join(cells, strings.Repeat(" ", gaugeGap)), " ")
			if strings.TrimSpace(ansi.Strip(line)) == "" {
				continue // no trends without a 7-day average
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// Arc is one half-circle gauge for callers outside the summary.
type Arc struct {
	Label, Value string
	Fraction     float64
	Color        lipgloss.Color
}

// Arcs renders arcs side by side, as the summary's sections do.
func Arcs(theme types.Theme, arcs []Arc, width int) []string {
	gauges := make([]gauge, 0, len(arcs))
	for _, a := range arcs {
		gauges = append(gauges, gauge{label: a.Label, value: a.Value, fraction: a.Fraction, color: a.Color})
	}
	return renderGauges(theme, gauges, width)
}

// ArcsWidth is the cells n arcs take side by side.
func ArcsWidth(n int) int {
	return max(1, n)*gaugeCols + max(0, n-1)*gaugeGap
}

// BesideArcs puts fact lines to the right of arcs, or under them when the width is narrow.
func BesideArcs(arcs, facts []string, arcWidth, width int) []string {
	return besideArcs(arcs, facts, arcWidth, width)
}
