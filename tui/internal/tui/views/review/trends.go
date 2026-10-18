package review

import (
	"fmt"
	"math"
	"strings"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

type metricSeries struct {
	name         string
	color        lipgloss.Color
	values       []float64
	min          float64
	max          float64
	formatLatest func(float64) string
}

func metricTrendValues(days []api.DailyMetricsDay, fn func(api.DailyMetricsDay) float64) []float64 {
	values := make([]float64, 0, len(days))
	for _, day := range days {
		values = append(values, fn(day))
	}
	return values
}

func metricMoodValue(day api.DailyMetricsDay) float64 {
	if day.CheckIn == nil {
		return 0
	}
	return float64(day.CheckIn.Mood)
}

func metricEnergyValue(day api.DailyMetricsDay) float64 {
	if day.CheckIn == nil {
		return 0
	}
	return float64(day.CheckIn.Energy)
}

func metricRecoveryValue(day api.DailyMetricsDay) float64 {
	sleep := 0.0
	if day.CheckIn != nil {
		if day.CheckIn.SleepHours != nil {
			sleep = clamp01(*day.CheckIn.SleepHours / 8.0)
		} else if day.CheckIn.SleepScore != nil {
			sleep = clamp01(float64(*day.CheckIn.SleepScore) / 100.0)
		}
	}
	breakRatio := 0.0
	if day.WorkedSeconds > 0 {
		breakRatio = float64(day.RestSeconds) / float64(day.WorkedSeconds)
	}
	return clamp01((sleep + clamp01(breakRatio/0.2)) / 2.0)
}

func renderMetricCanvas(
	theme types.Theme,
	metric metricSeries,
	labelWidth, chartWidth, rowHeight int,
) []string {
	if len(metric.values) == 0 {
		return []string{theme.StyleHeader.Render(metric.name), theme.StyleDim.Render("no data")}
	}
	label := theme.StyleHeader.Render(metric.name)
	latest := metric.values[len(metric.values)-1]
	latestLabel := "-"
	if metric.formatLatest != nil {
		latestLabel = metric.formatLatest(latest)
	}
	meta := theme.StyleDim.Render(latestLabel)
	graph := renderBrailleCanvas(metric.values, metric.min, metric.max, chartWidth, rowHeight)
	if len(graph) == 0 {
		return []string{fmt.Sprintf("%s  %s", label, meta), theme.StyleDim.Render("no data")}
	}
	pad := strings.Repeat(" ", max(0, labelWidth-len(metric.name)))
	out := make([]string, 0, len(graph)+1)
	out = append(out, fmt.Sprintf("%s%s  %s", label, pad, meta))
	color := lipgloss.NewStyle().Foreground(metric.color).Bold(true)
	for _, line := range graph {
		out = append(out, strings.Repeat(" ", labelWidth)+"  "+color.Render(line))
	}
	return out
}

func renderBrailleCanvas(values []float64, minValue, maxValue float64, width, height int) []string {
	if width < 2 || height < 1 || len(values) == 0 {
		return nil
	}
	pixelWidth := width * 2
	pixelHeight := height * 4
	pixels := make([][]bool, pixelHeight)
	for y := range pixels {
		pixels[y] = make([]bool, pixelWidth)
	}
	span := maxValue - minValue
	if span <= 0 {
		span = 1
	}
	prevX := -1
	prevY := -1
	for x := range pixelWidth {
		idx := 0
		if len(values) > 1 {
			idx = int(math.Round(float64(x) * float64(len(values)-1) / float64(pixelWidth-1)))
		}

		idx = max(0, min(idx, len(values)-1))

		if idx >= len(values) {
			idx = len(values) - 1
		}
		norm := clamp01((values[idx] - minValue) / span)
		y := int(math.Round((1 - norm) * float64(pixelHeight-1)))
		y = max(0, min(y, pixelHeight-1))
		if prevX >= 0 {
			drawBrailleLine(pixels, prevX, prevY, x, y)
		}
		setBraillePixel(pixels, x, y)
		prevX = x
		prevY = y
	}
	lines := make([]string, height)
	for cellY := range height {
		var b strings.Builder
		for cellX := range width {
			bits := brailleBitsForCell(pixels, cellX, cellY)
			if bits == 0 {
				b.WriteRune(' ')
				continue
			}
			b.WriteRune(rune(0x2800) + rune(bits))
		}
		line := strings.TrimRight(b.String(), " ")
		if line == "" {
			line = " "
		}
		lines[cellY] = line
	}
	return lines
}

func drawBrailleLine(pixels [][]bool, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		setBraillePixel(pixels, x0, y0)
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

func setBraillePixel(pixels [][]bool, x, y int) {
	if y < 0 || y >= len(pixels) {
		return
	}
	if x < 0 || x >= len(pixels[y]) {
		return
	}
	pixels[y][x] = true
}

func brailleBitsForCell(pixels [][]bool, cellX, cellY int) uint8 {
	baseX := cellX * 2
	baseY := cellY * 4
	bits := uint8(0)
	if braillePixel(pixels, baseX, baseY) {
		bits |= 0x01
	}
	if braillePixel(pixels, baseX, baseY+1) {
		bits |= 0x02
	}
	if braillePixel(pixels, baseX, baseY+2) {
		bits |= 0x04
	}
	if braillePixel(pixels, baseX+1, baseY) {
		bits |= 0x08
	}
	if braillePixel(pixels, baseX+1, baseY+1) {
		bits |= 0x10
	}
	if braillePixel(pixels, baseX+1, baseY+2) {
		bits |= 0x20
	}
	if braillePixel(pixels, baseX, baseY+3) {
		bits |= 0x40
	}
	if braillePixel(pixels, baseX+1, baseY+3) {
		bits |= 0x80
	}
	return bits
}

func braillePixel(pixels [][]bool, x, y int) bool {
	if y < 0 || y >= len(pixels) {
		return false
	}
	if x < 0 || x >= len(pixels[y]) {
		return false
	}
	return pixels[y][x]
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// focusBlockHeight is the focus graph's rows under its label; metricRowHeight is each
// braille chart's height.
const (
	focusBlockHeight = 10
	metricRowHeight  = 3
)

// renderTrends stacks the focus graph and the mood, energy and recovery charts, dropping
// charts from the bottom while they do not fit (the first always stays).
func renderTrends(theme types.Theme, state types.ContentState, width, height int) string {
	bodyWidth := max(1, width-2)
	blocks := [][]string{append([]string{theme.StyleHeader.Render("Focus") + "  " + renderFocusLegend(theme, state)},
		renderFocusGraph(theme, state.RollupMetricsRange, bodyWidth, focusBlockHeight)...)}
	labelWidth := len("Recovery")
	chartWidth := max(2, bodyWidth-labelWidth-2)
	for _, metric := range trendSeries(theme, state) {
		blocks = append(blocks, renderMetricCanvas(theme, metric, labelWidth, chartWidth, metricRowHeight))
	}
	lines := blocks[0]
	for _, block := range blocks[1:] {
		if len(lines)+1+len(block) > height-1 {
			break
		}
		lines = append(append(lines, ""), block...)
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Trends", Meta: viewchrome.KeyHint(theme, "S", "time"), Width: width, Height: height, Lines: lines,
	})
}

func trendSeries(theme types.Theme, state types.ContentState) []metricSeries {
	return []metricSeries{
		{
			name: "Mood", color: theme.ColorCyan, min: 1, max: 5,
			values:       metricTrendValues(state.MetricsRange, metricMoodValue),
			formatLatest: func(value float64) string { return fmt.Sprintf("%.1f/5", value) },
		},
		{
			name: "Energy", color: theme.ColorYellow, min: 1, max: 5,
			values:       metricTrendValues(state.MetricsRange, metricEnergyValue),
			formatLatest: func(value float64) string { return fmt.Sprintf("%.1f/5", value) },
		},
		{
			name: "Recovery", color: theme.ColorMagenta, min: 0, max: 1,
			values:       metricTrendValues(state.MetricsRange, metricRecoveryValue),
			formatLatest: func(value float64) string { return fmt.Sprintf("%.0f%%", value*100) },
		},
	}
}
