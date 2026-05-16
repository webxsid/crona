package wellbeing

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func summaryBodyLines(theme types.Theme, state types.ContentState, width int, compact bool) []string {
	lines := []string{}
	if compact {
		lines = append(lines, compactCards(theme, state)...)
		if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
			lines = append(lines, theme.StyleDim.Render("No check-in recorded for this date"))
		} else if !countsForCheckInStreak(state.DailyCheckIn) {
			lines = append(lines, theme.StyleError.Render("Backfilled check-in"))
		}
		lines = append(lines, Accountability(types.ViewSizeCompact, theme, state)...)
		lines = append(lines, RiskSnapshot(types.ViewSizeCompact, theme, state)...)
		if heatmap := Heatmap(types.ViewSizeCompact, theme, state); len(heatmap) > 0 {
			lines = append(lines, theme.StyleHeader.Render("Activity"))
			lines = append(lines, heatmap...)
		}
		return lines
	}

	lines = append(lines, cards(theme, state, width)...)
	if heatmap := Heatmap(types.ViewSizeStandard, theme, state); len(heatmap) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Recent Activity"))
		lines = append(lines, heatmap...)
	}
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded for this date"))
	} else {
		lines = append(lines,
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Mood"), state.DailyCheckIn.Mood),
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Energy"), state.DailyCheckIn.Energy),
		)
		if state.DailyCheckIn.SleepHours != nil {
			lines = append(lines, fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Sleep"), helperpkg.FormatCompactDurationHours(*state.DailyCheckIn.SleepHours)))
		}
		if state.DailyCheckIn.SleepScore != nil {
			lines = append(lines, fmt.Sprintf("%s  %d/100", theme.StyleHeader.Render("Sleep Score"), *state.DailyCheckIn.SleepScore))
		}
		if state.DailyCheckIn.ScreenTimeMinutes != nil {
			lines = append(lines, fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Screen Time"), helperpkg.FormatCompactDurationMinutes(*state.DailyCheckIn.ScreenTimeMinutes)))
		}
		if state.DailyCheckIn.Notes != nil && *state.DailyCheckIn.Notes != "" {
			lines = append(lines, "", theme.StyleHeader.Render("Notes"), viewhelpers.Truncate(*state.DailyCheckIn.Notes, max(20, width-8)))
		}
		if !countsForCheckInStreak(state.DailyCheckIn) {
			lines = append(lines, "", theme.StyleError.Render("Backfilled check-in"), theme.StyleDim.Render("Recorded later, so it does not count toward the same-day streak."))
		}
	}
	lines = append(lines, Accountability(types.ViewSizeStandard, theme, state)...)
	lines = append(lines, RiskSnapshot(types.ViewSizeStandard, theme, state)...)
	return lines
}

func trendsBodyLines(theme types.Theme, state types.ContentState, width int, compact bool) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact, true)
}

func metricsBodyLines(theme types.Theme, state types.ContentState, width int, compact bool) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact, false)
}

func trendsBodyLinesWithStreaks(theme types.Theme, state types.ContentState, width int, compact bool, includeStreaks bool) []string {
	if state.MetricsRollup == nil {
		return []string{theme.StyleDim.Render("Loading metrics...")}
	}

	lines := []string{}
	if compact {
		lines = append(lines,
			fmt.Sprintf("%s  %d  %s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days, theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
			fmt.Sprintf("%s  %d  %s  %s", theme.StyleHeader.Render("Focus"), state.MetricsRollup.FocusDays, theme.StyleHeader.Render("Worked"), viewhelpers.FormatClockText(state.MetricsRollup.WorkedSeconds)),
		)
		if state.MetricsRollup.HabitDueCount > 0 || state.MetricsRollup.HabitCompletedCount > 0 || state.MetricsRollup.HabitFailedCount > 0 {
			lines = append(lines, fmt.Sprintf("%s  %d  %s  %d  %s  %d", theme.StyleHeader.Render("Habits due"), state.MetricsRollup.HabitDueCount, theme.StyleHeader.Render("Done"), state.MetricsRollup.HabitCompletedCount, theme.StyleHeader.Render("Failed"), state.MetricsRollup.HabitFailedCount))
		}
		if state.MetricsRollup.AverageMood != nil || state.MetricsRollup.AverageEnergy != nil {
			avgMood := "-"
			avgEnergy := "-"
			if state.MetricsRollup.AverageMood != nil {
				avgMood = fmt.Sprintf("%.1f", *state.MetricsRollup.AverageMood)
			}
			if state.MetricsRollup.AverageEnergy != nil {
				avgEnergy = fmt.Sprintf("%.1f", *state.MetricsRollup.AverageEnergy)
			}
			lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", theme.StyleHeader.Render("Mood"), avgMood, theme.StyleHeader.Render("Energy"), avgEnergy))
		}
		if includeStreaks {
			lines = append(lines, streaksBodyLines(theme, state, width, true)...)
		}
		if canvas := trendCanvas(theme, state, width-6, state.Height); len(canvas) > 0 {
			lines = append(lines, canvas...)
		} else if strips := trendStrips(theme, state); len(strips) > 0 {
			lines = append(lines, theme.StyleHeader.Render(fmt.Sprintf("Signals (%s)", wellbeingWindowLabel(state))))
			lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(strips[0], width-6)))
		}
		if burnout := latestBurnout(state); burnout != nil {
			risks, _ := burnoutContributorLines(burnout)
			if len(risks) > 0 {
				lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(risks[0], width-6)))
			}
		}
		return lines
	}

	lines = append(lines, trendCards(theme, state)...)
	if state.MetricsRollup.AverageMood != nil {
		lines = append(lines, fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Mood"), *state.MetricsRollup.AverageMood))
	}
	if state.MetricsRollup.AverageEnergy != nil {
		lines = append(lines, fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Energy"), *state.MetricsRollup.AverageEnergy))
	}
	if state.MetricsRollup.HabitDueCount > 0 || state.MetricsRollup.HabitCompletedCount > 0 || state.MetricsRollup.HabitFailedCount > 0 {
		lines = append(lines, fmt.Sprintf("%s  %d  %s  %d  %s  %d", theme.StyleHeader.Render("Habits due"), state.MetricsRollup.HabitDueCount, theme.StyleHeader.Render("Done"), state.MetricsRollup.HabitCompletedCount, theme.StyleHeader.Render("Failed"), state.MetricsRollup.HabitFailedCount))
	}
	if includeStreaks {
		lines = append(lines, "")
		lines = append(lines, streaksBodyLines(theme, state, width, false)...)
	}
	if canvas := trendCanvas(theme, state, width-6, state.Height); len(canvas) > 0 {
		lines = append(lines, "")
		lines = append(lines, canvas...)
	} else if strips := trendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render(fmt.Sprintf("Signals (%s)", wellbeingWindowLabel(state))))
		lines = append(lines, strips...)
	}
	if burnout := latestBurnout(state); burnout != nil {
		risks, recoveries := burnoutContributorLines(burnout)
		if len(risks) > 0 {
			lines = append(lines, "", theme.StyleHeader.Render("Top Risk Drivers"))
			lines = append(lines, risks...)
		}
		if len(recoveries) > 0 {
			lines = append(lines, theme.StyleHeader.Render("Top Recovery Drivers"))
			lines = append(lines, recoveries...)
		}
	}
	return lines
}

func streaksBodyLines(theme types.Theme, state types.ContentState, width int, compact bool) []string {
	if state.Streaks == nil {
		return []string{theme.StyleDim.Render("Loading momentum...")}
	}
	if compact {
		lines := []string{
			compactMomentumRow(theme, "Check-ins", "", sharedtypes.HabitStreakPeriodDay, state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, width),
			compactMomentumRow(theme, "Focus", "", sharedtypes.HabitStreakPeriodDay, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays, width),
		}
		for _, streak := range state.Streaks.CustomHabitStreaks {
			lines = append(lines, compactMomentumRow(theme, streak.Name, habitStreakCadenceLabel(streak.Period), streak.Period, streak.Current, streak.Longest, width))
		}
		return lines
	}
	lines := []string{}
	lines = append(lines, momentumRow(theme, "Check-ins", "", sharedtypes.HabitStreakPeriodDay, state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, width)...)
	lines = append(lines, momentumRow(theme, "Focus", "", sharedtypes.HabitStreakPeriodDay, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays, width)...)
	lines = append(lines, "", theme.StyleHeader.Render(viewhelpers.Truncate("Custom Momentum", width-6)))
	if len(state.Streaks.CustomHabitStreaks) == 0 {
		lines = append(lines, theme.StyleDim.Render("No custom momentum yet"))
	}
	for _, streak := range state.Streaks.CustomHabitStreaks {
		lines = append(lines, momentumRow(theme, streak.Name, habitStreakCadenceLabel(streak.Period), streak.Period, streak.Current, streak.Longest, width)...)
	}
	return lines
}

func momentumRow(theme types.Theme, name, cadence string, period sharedtypes.HabitStreakPeriod, current, longest, width int) []string {
	label := viewhelpers.Truncate(name, max(8, width-10))
	if cadence != "" {
		label = fmt.Sprintf("%s  %s", label, theme.StyleDim.Render("["+cadence+"]"))
	}
	unit := momentumUnit(period)
	return []string{
		theme.StyleHeader.Render(label),
		fmt.Sprintf("%s  %s current · %s best", momentumLadder(theme, period, current, longest), formatMomentumLength(current, unit), formatMomentumLength(longest, unit)),
	}
}

func compactMomentumRow(theme types.Theme, name, cadence string, period sharedtypes.HabitStreakPeriod, current, longest, width int) string {
	label := name
	if cadence != "" {
		label = fmt.Sprintf("%s [%s]", label, cadence)
	}
	return viewhelpers.Truncate(fmt.Sprintf("%s  %s  %s", label, momentumLadder(theme, period, current, longest), formatMomentumLength(current, momentumUnit(period))), max(12, width-6))
}

func momentumLadder(theme types.Theme, period sharedtypes.HabitStreakPeriod, current, longest int) string {
	thresholds := momentumThresholds(period)
	filled := momentumTierCount(period, current)
	ladder := strings.Repeat("▰", filled) + strings.Repeat("▱", len(thresholds)-filled)
	style := lipgloss.NewStyle().Foreground(theme.ColorDim)
	switch {
	case current > 0 && current == longest:
		style = lipgloss.NewStyle().Foreground(theme.ColorCyan)
	case filled >= 4:
		style = lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case filled > 0:
		style = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	}
	return style.Render(ladder)
}

func momentumTierCount(period sharedtypes.HabitStreakPeriod, current int) int {
	if current <= 0 {
		return 0
	}
	filled := 0
	for _, threshold := range momentumThresholds(period) {
		if current < threshold {
			break
		}
		filled++
	}
	return filled
}

func momentumThresholds(period sharedtypes.HabitStreakPeriod) []int {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return []int{1, 2, 4, 8, 13, 26, 52}
	case sharedtypes.HabitStreakPeriodMonth:
		return []int{1, 2, 3, 6, 12, 24}
	default:
		return []int{1, 3, 7, 14, 30, 60, 100}
	}
}

func momentumUnit(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "w"
	case sharedtypes.HabitStreakPeriodMonth:
		return "mo"
	default:
		return "d"
	}
}

func formatMomentumLength(value int, unit string) string {
	return fmt.Sprintf("%d%s", value, unit)
}

func habitStreakCadenceLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "weekly"
	case sharedtypes.HabitStreakPeriodMonth:
		return "monthly"
	default:
		return "daily"
	}
}
