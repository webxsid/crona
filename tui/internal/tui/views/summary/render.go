// Package summary renders the Day summary: a day's focus stats, plan accountability,
// check-in and streaks, shown in Daily's detail pane or its Day dialog.
package summary

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// DayLines is the Day summary at width: the day's focus stats, plan accountability,
// check-in and streaks. It waits on the summary snapshot for the shown date.
func DayLines(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate {
		return []string{theme.StyleDim.Render("Loading summary…")}
	}
	lines := strings.Split(renderStats(theme, snapshot, width), "\n")
	lines = append(lines, "")
	lines = append(lines, renderPlan(theme, snapshot.Plan, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWellbeing(theme, snapshot.CheckIn, snapshot.Rollup, width)...)
	lines = append(lines, "")
	return append(lines, renderStreaks(theme, snapshot.Streaks, width)...)
}

func renderStats(theme types.Theme, snapshot *api.SummarySnapshot, width int) string {
	if snapshot == nil {
		return ""
	}
	lines := []string{theme.StyleHeader.Render("Stats")}
	if snapshot.Focus == nil || snapshot.Metrics == nil {
		return strings.Join(append(lines, theme.StyleDim.Render("Stats unavailable.")), "\n")
	}
	focus := snapshot.Focus
	metrics := snapshot.Metrics
	reason := strings.ReplaceAll(string(focus.Reason), "_", " ")
	level := strings.ReplaceAll(string(focus.Level), "_", " ")
	if level != "" {
		level = strings.ToUpper(level[:1]) + level[1:]
	}
	lines = append(lines,
		fmt.Sprintf("%s  %s  %s", theme.StylePaneTitle.Render(fmt.Sprintf("Focus %d/100", focus.Score)), theme.StyleNormal.Render(level), theme.StyleDim.Render(reason)),
		fmt.Sprintf("worked %s  ·  rest %s  ·  sessions %d  ·  target %s",
			helperpkg.FormatCompactDurationSeconds(metrics.WorkedSeconds),
			helperpkg.FormatCompactDurationSeconds(metrics.RestSeconds),
			metrics.SessionCount,
			helperpkg.FormatCompactDurationSeconds(focus.TargetWorkedSeconds)),
		fmt.Sprintf("issues  completed %d  ·  abandoned %d  ·  planned %d",
			metrics.CompletedIssues, metrics.AbandonedIssues, metrics.TotalIssues),
		fmt.Sprintf("habits  done %d  ·  due %d  ·  failed %d",
			metrics.HabitCompletedCount, metrics.HabitDueCount, metrics.HabitFailedCount),
		renderBar(theme, metrics.WorkedSeconds, 0, max(1, focus.TargetWorkedSeconds), max(8, width-2), theme.ColorCyan),
	)
	return strings.Join(lines, "\n")
}

func renderPlan(theme types.Theme, plan *sharedtypes.DailyPlan, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Accountability")}
	if plan == nil || strings.TrimSpace(plan.Date) == "" {
		return append(lines, theme.StyleDim.Render("No committed plan."))
	}
	s := plan.Summary
	lines = append(lines,
		fmt.Sprintf("%s  %s", theme.StyleNormal.Render(fmt.Sprintf("%.0f%% score", s.AccountabilityScore)), theme.StyleDim.Render(fmt.Sprintf("pressure %.1f", s.BacklogPressure))),
		renderBar(theme, s.CompletedCount, s.FailedCount+s.AbandonedCount, max(1, s.PlannedCount), max(8, width-2), theme.ColorGreen),
		theme.StyleDim.Render(fmt.Sprintf("done %d  pending %d  failed %d  abandoned %d", s.CompletedCount, s.PendingRollbackCount, s.FailedCount, s.AbandonedCount)),
	)
	return lines
}

func renderWellbeing(theme types.Theme, checkIn *sharedtypes.DailyCheckIn, rollup *sharedtypes.MetricsRollup, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Wellbeing")}
	if checkIn == nil || strings.TrimSpace(checkIn.Date) == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded."))
		if rollup != nil && (rollup.AverageMood != nil || rollup.AverageEnergy != nil) {
			lines = append(lines, theme.StyleDim.Render(averageSignals(rollup)))
		}
		return lines
	}
	signals := []string{fmt.Sprintf("mood %d/5", checkIn.Mood), fmt.Sprintf("energy %d/5", checkIn.Energy)}
	if checkIn.SleepHours != nil {
		signals = append(signals, "sleep "+helperpkg.FormatCompactDurationHours(*checkIn.SleepHours))
	}
	if checkIn.ScreenTimeMinutes != nil {
		signals = append(signals, "screen "+helperpkg.FormatCompactDurationMinutes(*checkIn.ScreenTimeMinutes))
	}
	lines = append(lines, viewhelpers.TruncateANSI(theme.StyleNormal.Render(strings.Join(signals, "  ·  ")), width))
	if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
		lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(strings.TrimSpace(*checkIn.Notes), width)))
	}
	return lines
}

func renderStreaks(theme types.Theme, streaks *sharedtypes.StreakSummary, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Momentum")}
	if streaks == nil {
		return append(lines, theme.StyleDim.Render("No streak data."))
	}
	text := fmt.Sprintf("focus %d  ·  check-ins %d  ·  habits %d", streaks.CurrentFocusDays, streaks.CurrentCheckInDays, streaks.CurrentHabitDays)
	return append(lines, viewhelpers.TruncateANSI(theme.StyleNormal.Render(text), width))
}

func averageSignals(rollup *sharedtypes.MetricsRollup) string {
	parts := make([]string, 0, 3)
	if rollup.AverageMood != nil {
		parts = append(parts, fmt.Sprintf("7d mood %.1f/5", *rollup.AverageMood))
	}
	if rollup.AverageEnergy != nil {
		parts = append(parts, fmt.Sprintf("energy %.1f/5", *rollup.AverageEnergy))
	}
	if rollup.AverageSleepHours != nil {
		parts = append(parts, "sleep "+helperpkg.FormatCompactDurationHours(*rollup.AverageSleepHours))
	}
	return strings.Join(parts, "  ·  ")
}

func renderBar(theme types.Theme, current, failed, total, width int, color lipgloss.Color) string {
	if width < 1 {
		return ""
	}
	if total < 1 {
		return theme.StyleDim.Render(strings.Repeat("░", width))
	}
	filled := clamp(current*width/total, 0, width)
	failedWidth := clamp(failed*width/total, 0, width-filled)
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("█", failedWidth)) +
		theme.StyleDim.Render(strings.Repeat("░", width-filled-failedWidth))
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
