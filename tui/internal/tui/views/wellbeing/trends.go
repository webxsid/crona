package wellbeing

import (
	"fmt"
	"math"
	"strings"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func trendStrips(theme types.Theme, state types.ContentState) []string {
	if len(state.MetricsRange) == 0 {
		return nil
	}
	window := state.MetricsRange
	if limit := wellbeingWindowDays(state); len(window) > limit {
		window = window[len(window)-limit:]
	}
	mood := make([]float64, 0, len(window))
	energy := make([]float64, 0, len(window))
	work := make([]float64, 0, len(window))
	recovery := make([]float64, 0, len(window))
	for _, day := range window {
		if day.CheckIn != nil {
			mood = append(mood, float64(day.CheckIn.Mood))
			energy = append(energy, float64(day.CheckIn.Energy))
		} else {
			mood = append(mood, 0)
			energy = append(energy, 0)
		}
		work = append(work, float64(day.WorkedSeconds)/3600.0)
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
		recovery = append(recovery, clamp01((sleep+clamp01(breakRatio/0.2))/2.0))
	}
	return []string{
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Mood"), sparkline(mood, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Energy"), sparkline(energy, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Work"), sparkline(work, 0.0, 8.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Recovery"), sparkline(recovery, 0.0, 1.0)),
	}
}

func trendCanvas(theme types.Theme, state types.ContentState, width, height int) []string {
	if len(state.MetricsRange) == 0 || width < 48 {
		return nil
	}
	labelWidth := max(10, len("Recovery"))
	chartWidth := width - labelWidth - 2
	if chartWidth < 18 {
		return nil
	}
	rowHeight := 1
	if width >= 110 && height >= 18 {
		rowHeight = 2
	}
	series := []metricSeries{
		{
			name:        "Mood",
			color:       theme.ColorCyan,
			values:      metricTrendValues(state.MetricsRange, func(day api.DailyMetricsDay) float64 { return metricMoodValue(day) }),
			min:         1,
			max:         5,
			formatLatest: func(value float64) string { return fmt.Sprintf("%.1f/5", value) },
		},
		{
			name:        "Energy",
			color:       theme.ColorYellow,
			values:      metricTrendValues(state.MetricsRange, func(day api.DailyMetricsDay) float64 { return metricEnergyValue(day) }),
			min:         1,
			max:         5,
			formatLatest: func(value float64) string { return fmt.Sprintf("%.1f/5", value) },
		},
		{
			name:        "Work",
			color:       theme.ColorGreen,
			values:      metricTrendValues(state.MetricsRange, func(day api.DailyMetricsDay) float64 { return float64(day.WorkedSeconds) / 3600.0 }),
			min:         0,
			max:         8,
			formatLatest: func(value float64) string { return fmt.Sprintf("%.1fh", value) },
		},
		{
			name:        "Recovery",
			color:       theme.ColorMagenta,
			values:      metricTrendValues(state.MetricsRange, func(day api.DailyMetricsDay) float64 { return metricRecoveryValue(day) }),
			min:         0,
			max:         1,
			formatLatest: func(value float64) string { return fmt.Sprintf("%.0f%%", value*100) },
		},
	}
	lines := []string{
		theme.StyleHeader.Render(fmt.Sprintf("Signals (%s)", wellbeingWindowLabel(state))),
	}
	for i, metric := range series {
		block := renderMetricCanvas(theme, metric, labelWidth, chartWidth, rowHeight)
		if len(block) == 0 {
			continue
		}
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, block...)
	}
	return lines
}

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

func renderMetricCanvas(theme types.Theme, metric metricSeries, labelWidth, chartWidth, rowHeight int) []string {
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
	for x := 0; x < pixelWidth; x++ {
		idx := 0
		if len(values) > 1 {
			idx = int(math.Round(float64(x) * float64(len(values)-1) / float64(pixelWidth-1)))
		}
		if idx < 0 {
			idx = 0
		}
		if idx >= len(values) {
			idx = len(values) - 1
		}
		norm := clamp01((values[idx] - minValue) / span)
		y := int(math.Round((1 - norm) * float64(pixelHeight-1)))
		if y < 0 {
			y = 0
		}
		if y >= pixelHeight {
			y = pixelHeight - 1
		}
		if prevX >= 0 {
			drawBrailleLine(pixels, prevX, prevY, x, y)
		}
		setBraillePixel(pixels, x, y)
		prevX = x
		prevY = y
	}
	lines := make([]string, height)
	for cellY := 0; cellY < height; cellY++ {
		var b strings.Builder
		for cellX := 0; cellX < width; cellX++ {
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

func wellbeingWindowLabel(state types.ContentState) string {
	days := state.WellbeingWindowDays
	if days < 1 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	return fmt.Sprintf("%dd", days)
}

func sparkline(values []float64, minValue, maxValue float64) string {
	if len(values) == 0 {
		return "-"
	}
	glyphs := []rune("▁▂▃▄▅▆▇█")
	span := maxValue - minValue
	if span <= 0 {
		span = 1
	}
	out := make([]rune, 0, len(values))
	for _, value := range values {
		norm := clamp01((value - minValue) / span)
		idx := int(math.Round(norm * float64(len(glyphs)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(glyphs) {
			idx = len(glyphs) - 1
		}
		out = append(out, glyphs[idx])
	}
	return string(out)
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

func wellbeingWindowDays(state types.ContentState) int {
	days := state.WellbeingWindowDays
	if days < 1 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	return days
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func countsForCheckInStreak(checkIn *api.DailyCheckIn) bool {
	if checkIn == nil {
		return false
	}
	return len(checkIn.CreatedAt) >= 10 && checkIn.CreatedAt[:10] == checkIn.Date
}
