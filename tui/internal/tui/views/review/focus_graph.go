package review

import (
	"fmt"
	"math"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/views/charts"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// hourSteps are the hour tick spacings the focus axis picks from.
var hourSteps = []float64{1, 2, 3, 4, 6, 8, 12, 24}

// hourAxis rounds the busiest day up to the axis top, with at most 4 even steps above 0.
func hourAxis(hours float64) (top, step float64) {
	for _, step = range hourSteps {
		if math.Ceil(hours/step) <= 4 {
			break
		}
	}
	return max(1, math.Ceil(hours/step)) * step, step
}

// renderFocusBars draws each day's logged hours as a braille bar in its slot and the day's
// estimate as a dotted tick across it, at exactly rows lines with hour ticks in the gutter.
func renderFocusBars(theme types.Theme, days []api.DailyMetricsDay, starts []int, barWidth, width, rows int) []string {
	busiest := 0.0
	for _, day := range days {
		busiest = max(busiest, float64(day.WorkedSeconds)/3600, float64(day.TotalEstimatedMinutes)/60)
	}
	top, step := hourAxis(busiest)
	dots := func(hours float64) int { return int(math.Round(hours / top * float64(rows*4))) }

	grid := make([][]string, rows)
	for row := range grid {
		grid[row] = make([]string, width)
		for col := range grid[row] {
			grid[row][col] = " "
		}
	}
	ramp := viewhelpers.GradientRamp(theme.ColorDullGreen, theme.ColorCyan, rows)
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true)
	for i, day := range days {
		if i >= len(starts) {
			break
		}
		height := min(rows*4, dots(float64(day.WorkedSeconds)/3600))
		for row := range rows {
			if fill := height - (rows-1-row)*4; fill > 0 {
				style := lipgloss.NewStyle().Foreground(viewhelpers.GradientColorAt(ramp, row))
				for col := starts[i]; col < starts[i]+barWidth; col++ {
					grid[row][col] = style.Render(charts.Column(fill))
				}
			}
		}
		if day.TotalEstimatedMinutes > 0 {
			level := min(rows*4, max(1, dots(float64(day.TotalEstimatedMinutes)/60)))
			row := rows - 1 - (level-1)/4
			for col := max(0, starts[i]-1); col < min(width, starts[i]+barWidth+1); col++ {
				grid[row][col] = estimate.Render(charts.DotRow((level - 1) % 4))
			}
		}
	}

	labels := map[int]string{}
	for value := 0.0; value <= top; value += step {
		labels[rows-1-int(math.Round(value/top*float64(rows-1)))] = fmt.Sprintf("%gh", value)
	}
	lines := make([]string, 0, rows)
	for row := range rows {
		lines = append(lines, gutterLine(theme, labels[row], strings.Join(grid[row], "")))
	}
	return lines
}

func renderFocusLegend(theme types.Theme, state types.ContentState) string {
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render(charts.DotRow(1))
	actual := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render(charts.Column(4))
	estimateText := "estimated -"
	loggedText := "logged -"
	if state.GoalProgress != nil {
		estimateText = "estimated " + compactHoursFromMinutes(
			state.GoalProgress.TotalEstimateMinutes,
		)
		loggedText = "logged " + compactHoursFromSeconds(state.GoalProgress.TotalActualSeconds)
	}
	return fmt.Sprintf("%s %s   %s %s", estimate, estimateText, actual, loggedText)
}

func renderFocusSubtitle(state types.ContentState, width int) string {
	line := "logged bars with estimated reference line"
	focusText := ""
	if state.WeeklyFocusScore != nil {
		focusText = fmt.Sprintf(
			"   focus %d/100   %s",
			state.WeeklyFocusScore.Score,
			focusReasonText(state.WeeklyFocusScore.Reason),
		)
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
			focusText,
			signedCompactHoursFromSeconds(diffSeconds),
			status,
		)
	} else if focusText != "" {
		line = "logged bars with estimate line" + focusText
	}
	return truncateFocusLine(line, width)
}

func focusReasonText(reason sharedtypes.FocusScoreReason) string {
	switch reason {
	case sharedtypes.FocusScoreReasonNoActivity:
		return "ready to begin"
	case sharedtypes.FocusScoreReasonUnderTarget:
		return "build momentum"
	case sharedtypes.FocusScoreReasonNeedsBreaks:
		return "take a break"
	case sharedtypes.FocusScoreReasonOverextended:
		return "overextended"
	default:
		return "steady rhythm"
	}
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
