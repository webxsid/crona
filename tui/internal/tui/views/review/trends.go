package review

import (
	"fmt"
	"math"
	"strings"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/views/charts"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

type metricSeries struct {
	name   string
	color  lipgloss.Color
	values []float64
	min    float64
	max    float64
	format func(float64) string
}

func metricTrendValues(days []api.DailyMetricsDay, fn func(api.DailyMetricsDay) float64) []float64 {
	values := make([]float64, 0, len(days))
	for _, day := range days {
		values = append(values, fn(day))
	}
	return values
}

// metricMoodValue is NaN without a check-in, so the chart leaves a gap.
func metricMoodValue(day api.DailyMetricsDay) float64 {
	if day.CheckIn == nil {
		return math.NaN()
	}
	return float64(day.CheckIn.Mood)
}

func metricEnergyValue(day api.DailyMetricsDay) float64 {
	if day.CheckIn == nil {
		return math.NaN()
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

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

const (
	// trendGutter fits the widest axis label, "100%".
	trendGutter = 4
	// minMetricRows is the shortest wellbeing chart; below it the last chart drops.
	minMetricRows = 2
)

// renderTrends stacks the focus bars and the mood, energy and recovery bars over one shared
// date row. Every chart places a day at the same column, focus takes about 2/5 of the height,
// and the wellbeing charts split the rest, dropping from the bottom while too short.
func renderTrends(theme types.Theme, state types.ContentState, width, height int) string {
	room := max(1, height-1)
	chartWidth := max(4, width-2-trendGutter-2)
	days := state.RollupMetricsRange
	starts, barWidth := charts.Slots(max(1, len(days)), chartWidth)
	centers := make([]int, len(starts))
	for i, start := range starts {
		centers[i] = start + barWidth/2
	}

	metrics := trendSeries(theme, state)
	focusRows := max(4, room*2/5)
	metricRows := 0
	for ; len(metrics) > 0; metrics = metrics[:len(metrics)-1] {
		if metricRows = (room - 3 - focusRows - 2*len(metrics)) / len(metrics); metricRows >= minMetricRows {
			break
		}
	}
	focusRows = room - 3 - len(metrics)*(2+metricRows)

	lines := []string{theme.StyleHeader.Render("Focus") + "  " + renderFocusLegend(theme, state)}
	lines = append(lines, renderFocusBars(theme, days, starts, barWidth, chartWidth, focusRows)...)
	for _, metric := range metrics {
		lines = append(lines, "", metricHeader(theme, metric))
		lines = append(
			lines,
			metricLines(theme, metric, starts, barWidth, chartWidth, metricRows)...)
	}
	lines = append(lines, strings.Repeat(" ", trendGutter)+" └"+strings.Repeat("─", chartWidth))
	lines = append(lines, trendDates(theme, days, centers, chartWidth))
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Trends", Meta: viewchrome.KeyHint(theme, "S", "time"), Width: width, Height: height, Lines: lines,
	})
}

// gutterLine is a chart row behind its axis label.
func gutterLine(theme types.Theme, label, row string) string {
	return theme.StyleDim.Render(fmt.Sprintf("%*s", trendGutter, label)) + " │" + row
}

// metricHeader is the metric's latest value, the window's average and how the latest sits
// against it: ▲ above, ▼ below, ▬ within a sixteenth of the scale.
func metricHeader(theme types.Theme, metric metricSeries) string {
	latest, total, count := math.NaN(), 0.0, 0
	for _, value := range metric.values {
		if !math.IsNaN(value) {
			latest, total, count = value, total+value, count+1
		}
	}
	header := theme.StyleHeader.Render(fmt.Sprintf("%-9s", metric.name))
	if count == 0 {
		return header + theme.StyleDim.Render("–")
	}
	avg := total / float64(count)
	mark, band := "▬", (metric.max-metric.min)/16
	switch {
	case latest > avg+band:
		mark = "▲"
	case latest < avg-band:
		mark = "▼"
	}
	return header + metric.format(
		latest,
	) + "  " + theme.StyleDim.Render(
		"avg "+metric.format(avg),
	) + "  " + mark
}

// metricLines is the metric's days as short bars in the focus columns, rising from zero to
// the scale's top; a day without a value is a dim dot.
func metricLines(
	theme types.Theme,
	metric metricSeries,
	starts []int,
	barWidth, width, rows int,
) []string {
	values := metric.values[:min(len(metric.values), len(starts))]
	color := lipgloss.NewStyle().Foreground(metric.color)
	grid := charts.BarGrid(values, starts, barWidth, width, rows, metric.max)
	out := make([]string, 0, rows)
	for row := range grid {
		var b strings.Builder
		for _, cell := range grid[row] {
			switch cell {
			case " ":
				b.WriteString(cell)
			case "·":
				b.WriteString(theme.StyleDim.Render(cell))
			default:
				b.WriteString(color.Render(cell))
			}
		}
		label := ""
		if row == 0 {
			label = metric.format(metric.max)
		}
		out = append(out, gutterLine(theme, label, b.String()))
	}
	return out
}

// trendDates labels the first, middle and last days under their columns.
func trendDates(theme types.Theme, days []api.DailyMetricsDay, centers []int, width int) string {
	runes := []rune(strings.Repeat(" ", width))
	for _, i := range []int{0, len(days) / 2, len(days) - 1} {
		if i < 0 || i >= len(days) || i >= len(centers) {
			continue
		}
		label := []rune(shortFocusDate(days[i].Date))
		start := max(0, min(centers[i]-len(label)/2, width-len(label)))
		copy(runes[start:], label)
	}
	return strings.Repeat(" ", trendGutter+2) + theme.StyleDim.Render(string(runes))
}

func trendSeries(theme types.Theme, state types.ContentState) []metricSeries {
	return []metricSeries{
		{
			name: "Mood", color: theme.ColorCyan, min: 1, max: 5,
			values: metricTrendValues(state.MetricsRange, metricMoodValue),
			format: func(value float64) string { return fmt.Sprintf("%.1f", value) },
		},
		{
			name: "Energy", color: theme.ColorYellow, min: 1, max: 5,
			values: metricTrendValues(state.MetricsRange, metricEnergyValue),
			format: func(value float64) string { return fmt.Sprintf("%.1f", value) },
		},
		{
			name: "Recovery", color: theme.ColorMagenta, min: 0, max: 1,
			values: metricTrendValues(state.MetricsRange, metricRecoveryValue),
			format: func(value float64) string { return fmt.Sprintf("%.0f%%", value*100) },
		},
	}
}
