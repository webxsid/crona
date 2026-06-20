package rollup

import (
	"fmt"
	"math"
	"strings"
	"time"

	"crona/tui/internal/api"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

type focusCell struct {
	bar     bool
	line    bool
	lineTop bool
}

func renderFocusVisual(theme types.Theme, state types.ContentState, width int, availableHeight int) []string {
	window := state.RollupMetricsRange
	if len(window) == 0 || width < 28 {
		return []string{theme.StyleHeader.Render("Focus") + "  " + theme.StyleDim.Render("No focus data for this range.")}
	}
	lines := []string{
		theme.StyleHeader.Render("Focus"),
		theme.StyleDim.Render(renderFocusSubtitle(state, width)),
		renderFocusLegend(theme, state),
	}
	graphHeight := max(1, availableHeight-len(lines))
	lines = append(lines, renderFocusGraph(theme, window, width, graphHeight)...)
	return lines
}

func renderFocusGraph(
	theme types.Theme,
	days []api.DailyMetricsDay,
	width int,
	height int,
) []string {
	if len(days) == 0 || width < 24 || height < 6 {
		return []string{theme.StyleDim.Render("No focus data")}
	}

	yLabelWidth := 5
	chartWidth := max(12, width-yLabelWidth-2)
	chartHeight := max(4, height-2)

	estimated := make([]float64, 0, len(days))
	actual := make([]float64, 0, len(days))
	maxValue := 1.0
	for _, day := range days {
		est := float64(day.TotalEstimatedMinutes) / 60.0
		act := float64(day.WorkedSeconds) / 3600.0
		estimated = append(estimated, est)
		actual = append(actual, act)
		maxValue = math.Max(maxValue, math.Max(est, act))
	}
	maxAxis := math.Ceil(maxValue)
	if maxAxis < 1 {
		maxAxis = 1
	}

	grid := make([][]focusCell, chartHeight)
	for y := range grid {
		grid[y] = make([]focusCell, chartWidth)
	}

	xPositions := focusXPositions(len(days), chartWidth)
	for i, value := range actual {
		x := xPositions[i]
		row := focusValueRow(value, maxAxis, chartHeight)
		for y := row; y < chartHeight; y++ {
			grid[y][x].bar = true
		}
	}
	for i, value := range estimated {
		x := xPositions[i]
		row := focusValueRow(value, maxAxis, chartHeight)
		grid[row][x].line = true
		grid[row][x].lineTop = true
		if x > 0 {
			grid[row][x-1].line = true
		}
	}

	tickRows, tickLabels := focusTicks(maxAxis, chartHeight)
	lines := make([]string, 0, chartHeight+2)
	barRamp := viewhelpers.GradientRamp(theme.ColorDullGreen, theme.ColorCyan, chartHeight)
	lineStyle := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true)
	axisStyle := lipgloss.NewStyle().Foreground(theme.ColorDim)

	for row := range chartHeight {
		label := "    "
		if tick, ok := tickLabels[row]; ok {
			label = fmt.Sprintf("%4s", tick)
		}
		var b strings.Builder
		for col := range chartWidth {
			cell := grid[row][col]
			switch {
			case cell.lineTop:
				b.WriteString(lineStyle.Render("─"))
			case cell.line:
				b.WriteString(lineStyle.Render("╌"))
			case cell.bar:
				b.WriteString(
					lipgloss.NewStyle().
						Foreground(viewhelpers.GradientColorAt(barRamp, row)).
						Render("█"),
				)
			case tickRows[row]:
				b.WriteString(axisStyle.Render("┈"))
			default:
				b.WriteRune(' ')
			}
		}
		lines = append(lines, fmt.Sprintf("%s │%s", theme.StyleDim.Render(label), b.String()))
	}
	lines = append(lines, fmt.Sprintf("%s └%s", strings.Repeat(" ", yLabelWidth), strings.Repeat("─", chartWidth)))
	lines = append(lines, renderFocusXAxis(theme, days, yLabelWidth+2, chartWidth, xPositions))
	return lines
}

func focusXPositions(points int, chartWidth int) []int {
	out := make([]int, points)
	if points == 1 {
		out[0] = min(chartWidth-1, chartWidth/2)
		return out
	}
	for i := range points {
		out[i] = int(math.Round(float64(i) * float64(chartWidth-1) / float64(points-1)))
	}
	return out
}

func focusValueRow(value, maxAxis float64, chartHeight int) int {
	if maxAxis <= 0 {
		return chartHeight - 1
	}
	row := chartHeight - 1 - int(math.Round(clampFocus(value/maxAxis)*float64(chartHeight-1)))
	if row < 0 {
		return 0
	}
	if row >= chartHeight {
		return chartHeight - 1
	}
	return row
}

func focusTicks(maxAxis float64, chartHeight int) (map[int]bool, map[int]string) {
	rows := map[int]bool{}
	labels := map[int]string{}
	steps := min(chartHeight, 5)
	steps = max(steps, 2)
	seen := map[int]bool{}
	for i := 0; i < steps; i++ {
		value := maxAxis * (1 - float64(i)/float64(steps-1))
		row := int(math.Round(float64(i) * float64(chartHeight-1) / float64(steps-1)))

		row = min(chartHeight-1, max(0, row))
		if seen[row] {
			continue
		}
		seen[row] = true
		rows[row] = true
		labels[row] = fmt.Sprintf("%.0fh", value)
	}
	return rows, labels
}

func renderFocusLegend(theme types.Theme, state types.ContentState) string {
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render("╌")
	actual := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("█")
	estimateText := "estimated -"
	loggedText := "logged -"
	if state.GoalProgress != nil {
		estimateText = "estimated " + compactHoursFromMinutes(state.GoalProgress.TotalEstimateMinutes)
		loggedText = "logged " + compactHoursFromSeconds(state.GoalProgress.TotalActualSeconds)
	}
	return fmt.Sprintf("%s %s   %s %s", estimate, estimateText, actual, loggedText)
}

func renderFocusXAxis(
	theme types.Theme,
	days []api.DailyMetricsDay,
	leftPad int,
	chartWidth int,
	xPositions []int,
) string {
	if len(days) == 0 {
		return ""
	}
	runes := make([]rune, chartWidth)
	for i := range runes {
		runes[i] = ' '
	}
	samples := []int{0}
	if len(days) > 1 {
		samples = append(samples, len(days)/3, (2*len(days))/3, len(days)-1)
	}
	seen := map[int]bool{}
	for _, idx := range samples {
		if idx < 0 || idx >= len(days) || seen[idx] {
			continue
		}
		seen[idx] = true
		label := shortFocusDate(days[idx].Date)
		x := xPositions[idx]
		start := max(0, min(x-len(label)/2, chartWidth-len(label)))
		for i, r := range label {
			if start+i >= 0 && start+i < len(runes) {
				runes[start+i] = r
			}
		}
	}
	return strings.Repeat(" ", leftPad) + theme.StyleDim.Render(string(runes))
}

func renderFocusSubtitle(state types.ContentState, width int) string {
	line := "logged bars with estimated reference line"
	scoreText := ""
	if state.WeeklyFocusScore != nil {
		scoreText = fmt.Sprintf("   focus %d/100", state.WeeklyFocusScore.Score)
	}
	if state.GoalProgress != nil {
		diffSeconds := state.GoalProgress.TotalActualSeconds - (state.GoalProgress.TotalEstimateMinutes * 60)
		status := "on target"
		switch {
		case diffSeconds > 0:
			status = "over estimate"
		case diffSeconds < 0:
			status = "under estimate"
		}
		line = fmt.Sprintf(
			"logged bars with estimate line%s   delta %s   %s",
			scoreText,
			signedCompactHoursFromSeconds(diffSeconds),
			status,
		)
	} else if scoreText != "" {
		line = "logged bars with estimate line" + scoreText
	}
	return truncateFocusLine(line, width)
}

func compactHoursFromMinutes(minutes int) string {
	return fmt.Sprintf("%.1fh", float64(minutes)/60.0)
}

func compactHoursFromSeconds(seconds int) string {
	return fmt.Sprintf("%.1fh", float64(seconds)/3600.0)
}

func signedCompactHoursFromSeconds(seconds int) string {
	sign := ""
	if seconds > 0 {
		sign = "+"
	}
	return sign + compactHoursFromSeconds(seconds)
}

func shortFocusDate(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("Jan 2")
}

func clampFocus(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func truncateFocusLine(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	if width <= 1 {
		return string(runes[:max(0, min(len(runes), width))])
	}
	return string(runes[:max(0, min(len(runes), width-1))]) + "…"
}
