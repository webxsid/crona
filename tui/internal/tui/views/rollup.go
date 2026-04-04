package views

import (
	"fmt"
	"math"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"

	"github.com/charmbracelet/lipgloss"
)

func renderRollupView(theme Theme, state ContentState) string {
	headerH, detailH := splitVertical(state.Height, 10, 8, 12)
	header := renderRollupSummary(theme, state, state.Width, headerH)
	detail := renderRollupDetails(theme, state, state.Width, detailH)
	return lipgloss.JoinVertical(lipgloss.Left, header, detail)
}

func renderRollupSummary(theme Theme, state ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Rollup Dashboard"),
		fmt.Sprintf("%s  %s -> %s", theme.StyleHeader.Render("Range"), theme.StyleNormal.Render(state.RollupStartDate), theme.StyleNormal.Render(state.RollupEndDate)),
		theme.StyleDim.Render("[S/E] calendar   [h/l] start   [,/.] end   [g] weekly"),
		"",
		renderRollupWindowLine(theme, state),
		renderRollupFocusLine(theme, state),
		renderRollupProgressLine(theme, state),
		renderRollupEstimateBiasLine(theme, state),
	}
	lines = clipDailySummaryLines(theme, lines, height)
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width - 2).Height(max(1, height-2)).Render(stringsJoin(lines))
}

func renderRollupWindowLine(theme Theme, state ContentState) string {
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

func renderRollupFocusLine(theme Theme, state ContentState) string {
	if state.WeeklyFocusScore == nil {
		return "Focus  score -  status -  worked -"
	}
	level := strings.ToLower(string(state.WeeklyFocusScore.Level))
	return fmt.Sprintf("%s  score %s  status %s  worked %s",
		theme.StyleHeader.Render("Focus"),
		rollupLevelStyle(theme, level).Render(fmt.Sprintf("%d/100", state.WeeklyFocusScore.Score)),
		rollupLevelStyle(theme, level).Render(level),
		theme.StyleNormal.Render(formatClock(state.WeeklyFocusScore.WorkedSeconds)),
	)
}

func renderRollupProgressLine(theme Theme, state ContentState) string {
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
		theme.StyleNormal.Render(formatClock(state.GoalProgress.TotalActualSeconds)),
		rollupProgressStyle(theme, state.GoalProgress.EstimateBias, status).Render(status),
	)
}

func renderRollupEstimateBiasLine(theme Theme, state ContentState) string {
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
		rollupBiasStyle(theme, bias).Render(delta),
		rollupBiasStyle(theme, bias).Render(percent),
		rollupBiasStyle(theme, bias).Render(bias),
		state.GoalProgress.EstimatedItems,
	)
}

func renderRollupDetails(theme Theme, state ContentState, width, height int) string {
	leftW, rightW := splitHorizontal(width, 34, 34, width/2)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		renderRollupWindowDays(theme, state, leftW, height),
		renderRollupDistributionPane(theme, state, rightW, height),
	)
}

func renderRollupWindowDays(theme Theme, state ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Daily Status"),
		theme.StyleDim.Render("[enter] details"),
	}
	if state.DashboardWindow == nil || len(state.DashboardWindow.Days) == 0 {
		lines = append(lines, theme.StyleDim.Render("No daily rollup data for this range"))
		return renderPaneBox(theme, state.View == "rollup", width, height, stringsJoin(lines))
	}
	active := state.View == "rollup" && state.Pane == "rollup_days"
	cur := state.Cursors["rollup_days"]
	inner := remainingPaneHeight(height, lines)
	start, end := listWindow(cur, len(state.DashboardWindow.Days), inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("  ↑ %d more", start)))
	}
	for idx := start; idx < end; idx++ {
		day := state.DashboardWindow.Days[idx]
		status := rollupStatusStyle(theme, string(day.Status)).Render(prettyWindowStatus(string(day.Status)))
		row := fmt.Sprintf("%s  %-11s  p%d d%d f%d c%d", day.Date, status, day.PlannedCount, day.CompletedCount, day.FailedCount, day.CarryOverCount)
		switch {
		case active && idx == cur:
			lines = append(lines, theme.StyleCursor.Render("▶ "+row))
		case idx == cur:
			lines = append(lines, theme.StyleSelected.Render("  "+row))
		default:
			lines = append(lines, "  "+row)
		}
	}
	if remaining := len(state.DashboardWindow.Days) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("  ↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, width, height, stringsJoin(lines))
}

func renderRollupDistributionPane(theme Theme, state ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Breakdown"),
		theme.StyleDim.Render("Top time allocations"),
	}
	lines = append(lines, renderDistributionSection(theme, "Repos", state.RepoDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Streams", state.StreamDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Issues", state.IssueDistribution)...)
	lines = append(lines, renderDistributionSection(theme, "Segments", state.SegmentDistribution)...)
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func renderDistributionSection(theme Theme, title string, summary *api.TimeDistributionSummary) []string {
	lines := []string{theme.StyleHeader.Render(title)}
	if summary == nil || len(summary.Rows) == 0 {
		return append(lines, theme.StyleDim.Render("  No data"))
	}
	limit := min(3, len(summary.Rows))
	for i := 0; i < limit; i++ {
		row := summary.Rows[i]
		lines = append(lines, fmt.Sprintf("  %s  %d%%  %s", truncate(row.Label, 18), int(row.Percent+0.5), formatClock(row.WorkedSeconds)))
	}
	return lines
}

func rollupStatusStyle(theme Theme, status string) lipgloss.Style {
	switch status {
	case "done":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "missed":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	case "carry_over":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "planned":
		return lipgloss.NewStyle().Foreground(theme.ColorBlue)
	case "mixed":
		return lipgloss.NewStyle().Foreground(theme.ColorMagenta)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorDim)
	}
}

func rollupLevelStyle(theme Theme, level string) lipgloss.Style {
	switch level {
	case "strong":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "steady":
		return lipgloss.NewStyle().Foreground(theme.ColorCyan)
	case "overextended":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	}
}

func rollupBiasStyle(theme Theme, bias string) lipgloss.Style {
	switch bias {
	case "under":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "over":
		return lipgloss.NewStyle().Foreground(theme.ColorCyan)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	}
}

func rollupProgressStyle(theme Theme, bias string, status string) lipgloss.Style {
	switch strings.TrimSpace(status) {
	case "on track":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case "at risk":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "over":
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	case "unestimated":
		return rollupBiasStyle(theme, bias)
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorWhite)
	}
}

func prettyWindowStatus(status string) string {
	switch status {
	case "done":
		return "done"
	case "missed":
		return "missed"
	case "carry_over":
		return "carry over"
	case "planned":
		return "planned"
	case "mixed":
		return "mixed"
	default:
		return "no activity"
	}
}
