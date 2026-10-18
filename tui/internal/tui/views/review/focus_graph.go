package review

import (
	"fmt"
	"math"
	"strings"
	"time"

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

// renderFocusBars draws each day's logged hours as a solid bar in its slot and the day's
// estimate as a dashed tick across it, at exactly rows lines with hour ticks in the gutter.
func renderFocusBars(theme types.Theme, days []api.DailyMetricsDay, starts []int, barWidth, width, rows int) []string {
	busiest := 0.0
	logged := make([]float64, len(days))
	for i, day := range days {
		logged[i] = float64(day.WorkedSeconds) / 3600
		busiest = max(busiest, logged[i], float64(day.TotalEstimatedMinutes)/60)
	}
	top, step := hourAxis(busiest)
	grid := charts.BarGrid(logged, starts, barWidth, width, rows, top)
	ramp := viewhelpers.GradientRamp(theme.ColorDullGreen, theme.ColorCyan, rows)
	for row := range grid {
		style := lipgloss.NewStyle().Foreground(viewhelpers.GradientColorAt(ramp, row))
		for col, cell := range grid[row] {
			if cell != " " {
				grid[row][col] = style.Render(cell)
			}
		}
	}
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render("╌")
	for i, day := range days {
		if i >= len(starts) || day.TotalEstimatedMinutes <= 0 {
			continue
		}
		eighths := max(1, int(math.Round(float64(day.TotalEstimatedMinutes)/60/top*float64(rows*8))))
		row := max(0, rows-1-(eighths-1)/8)
		for col := max(0, starts[i]-1); col < min(width, starts[i]+barWidth+1); col++ {
			grid[row][col] = estimate
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
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render("╌")
	actual := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("█")
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

func compactHoursFromMinutes(minutes int) string {
	return fmt.Sprintf("%.1fh", float64(minutes)/60.0)
}

func compactHoursFromSeconds(seconds int) string {
	return fmt.Sprintf("%.1fh", float64(seconds)/3600.0)
}

func shortFocusDate(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("Jan 2")
}
