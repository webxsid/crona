package rollup

import (
	"fmt"
	"math"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderSummary(theme types.Theme, state types.ContentState, width, height int) string {
	startDateText := helperpkg.FormatDisplayDate(state.RollupStartDate, state.Settings)
	endDateText := helperpkg.FormatDisplayDate(state.RollupEndDate, state.Settings)
	summaryInnerW := max(24, width-8)

	lines := []string{
		theme.StylePaneTitle.Render("Rollup Dashboard"),
		fmt.Sprintf("%s  %s -> %s", theme.StyleHeader.Render("Range"), theme.StyleNormal.Render(startDateText), theme.StyleNormal.Render(endDateText)),
		theme.StyleDim.Render("[S/E] calendar   [h/l] start   [,/.] end   [g] weekly"),
		"",
		renderWindowLine(theme, state),
		renderFocusLine(theme, state),
		renderProgressLine(theme, state),
		renderEstimateBiasLine(theme, state),
	}
	lines = viewchrome.ClipSummaryLines(theme, lines, height)
	if viewcalendar.ShouldRender(summaryInnerW) {
		calendarLines := viewcalendar.Render(theme, viewcalendar.Selection{
			AnchorDate: state.RollupEndDate,
			RangeStart: state.RollupStartDate,
			RangeEnd:   state.RollupEndDate,
			MaxLines:   len(lines),
		})
		if len(calendarLines) > 0 {
			lines = viewcalendar.MergeBeside(lines, calendarLines, summaryInnerW, 3)
		}
	}
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width - 2).Height(max(1, height-2)).Render(viewhelpers.StringsJoin(lines))
}

func renderWindowLine(theme types.Theme, state types.ContentState) string {
	if state.DashboardWindow == nil {
		return "Plan  planned -  done -  missed -  carry -"
	}
	return fmt.Sprintf("%s  %s %d  %s %d  %s %d  %s %d",
		theme.StyleHeader.Render("Plan"),
		lipgloss.NewStyle().Foreground(theme.ColorBlue).Render("planned"), state.DashboardWindow.PlannedCount,
		lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("done"), state.DashboardWindow.CompletedCount,
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render("missed"), state.DashboardWindow.MissedCount,
		lipgloss.NewStyle().Foreground(theme.ColorYellow).Render("carry"), state.DashboardWindow.CarryOverCount,
	)
}

func renderFocusLine(theme types.Theme, state types.ContentState) string {
	if state.WeeklyFocusScore == nil {
		return "Focus  score -  status -  worked -"
	}
	level := strings.ToLower(string(state.WeeklyFocusScore.Level))
	return fmt.Sprintf("%s  score %s  status %s  worked %s",
		theme.StyleHeader.Render("Focus"),
		levelStyle(theme, level).Render(fmt.Sprintf("%d/100", state.WeeklyFocusScore.Score)),
		levelStyle(theme, level).Render(level),
		theme.StyleNormal.Render(viewhelpers.FormatClock(state.WeeklyFocusScore.WorkedSeconds)),
	)
}

func renderProgressLine(theme types.Theme, state types.ContentState) string {
	if state.GoalProgress == nil {
		return "Progress  estimated -  worked -  status -"
	}
	status := "-"
	if len(state.GoalProgress.Rows) > 0 {
		status = strings.ReplaceAll(string(state.GoalProgress.Rows[0].Status), "_", " ")
	}
	return fmt.Sprintf("%s  estimated %s  worked %s  status %s",
		theme.StyleHeader.Render("Progress"),
		theme.StyleNormal.Render(helperpkg.FormatCompactDurationMinutes(state.GoalProgress.TotalEstimateMinutes)),
		theme.StyleNormal.Render(viewhelpers.FormatClock(state.GoalProgress.TotalActualSeconds)),
		progressStyle(theme, state.GoalProgress.EstimateBias, status).Render(status),
	)
}

func renderEstimateBiasLine(theme types.Theme, state types.ContentState) string {
	if state.GoalProgress == nil || state.GoalProgress.EstimatedItems == 0 {
		return "Estimate Bias  no estimated work in this range"
	}
	delta := fmt.Sprintf("%+s", helperpkg.FormatCompactDurationMinutes(int(math.Round(state.GoalProgress.AverageDeltaMinutes))))
	percent := fmt.Sprintf("%+.0f%%", state.GoalProgress.AverageDeltaPercent)
	bias := strings.TrimSpace(state.GoalProgress.EstimateBias)
	if bias == "" {
		bias = "balanced"
	}
	return fmt.Sprintf("%s  avg %s  drift %s  bias %s  items %d",
		theme.StyleHeader.Render("Estimate Bias"),
		biasStyle(theme, bias).Render(delta),
		biasStyle(theme, bias).Render(percent),
		biasStyle(theme, bias).Render(bias),
		state.GoalProgress.EstimatedItems,
	)
}
