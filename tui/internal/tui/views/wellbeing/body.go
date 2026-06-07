package wellbeing

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func summaryBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	if compact {
		lines := append([]string{}, compactCards(theme, state)...)
			if heatmap := Heatmap(types.ViewSizeCompact, theme, state); len(heatmap) > 0 {
				lines = append(lines, theme.StyleHeader.Render("Activity"))
				lines = append(lines, heatmap...)
			}
			if status := wellbeingStatusLines(theme, state, width, true); len(status) > 0 {
				lines = append(lines, "")
				lines = append(lines, status...)
			}
		return lines
	}

	lines := append([]string{}, cards(theme, state, width)...)
	if heatmap := Heatmap(types.ViewSizeStandard, theme, state); len(heatmap) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Recent Activity"))
		lines = append(lines, heatmap...)
	}
	if status := wellbeingStatusLines(theme, state, width, false); len(status) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Today"), status[0])
	}
	return lines
}

func detailsBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	lines := []string{}
	if section := wellbeingCheckInDetailLines(theme, state, width, compact); len(section) > 0 {
		lines = append(lines, section...)
	}

	size := types.ViewSizeStandard
	if compact {
		size = types.ViewSizeCompact
	}
	if section := Accountability(size, theme, state); len(section) > 0 {
		lines = appendSection(lines, section)
	}
	if section := RiskSnapshot(size, theme, state); len(section) > 0 {
		lines = appendSection(lines, section)
	}
	if burnout := latestBurnout(state); burnout != nil {
		risks, recoveries := burnoutContributorLines(burnout)
		if len(risks) > 0 {
			lines = appendSection(
				lines,
				append([]string{theme.StyleHeader.Render("Top Risk Drivers")}, risks...),
			)
		}
		if len(recoveries) > 0 {
			lines = appendSection(
				lines,
				append([]string{theme.StyleHeader.Render("Top Recovery Drivers")}, recoveries...),
			)
		}
	}
	if len(lines) == 0 {
		return []string{theme.StyleDim.Render("No detailed wellbeing context for this date")}
	}
	return lines
}

func wellbeingStatusLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		return []string{theme.StyleDim.Render("No check-in recorded for this date")}
	}

	parts := []string{
		theme.StyleHeader.Render("Check-in"),
		theme.StyleNormal.Render("recorded"),
	}
	if !countsForCheckInStreak(state.DailyCheckIn) {
		parts = append(parts, theme.StyleError.Render("backfilled"))
	}
	if state.DailyCheckIn.Notes != nil && strings.TrimSpace(*state.DailyCheckIn.Notes) != "" {
		parts = append(
			parts,
			theme.StyleDim.Render(
				viewhelpers.Truncate(strings.TrimSpace(*state.DailyCheckIn.Notes), max(18, width-28)),
			),
		)
	}
	line := strings.Join(parts, "  ")
	limit := width - 8
	if compact {
		limit = width - 6
	}
	return []string{viewhelpers.Truncate(line, max(20, limit))}
}

func wellbeingCheckInDetailLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		return []string{
			theme.StyleHeader.Render("Check-in"),
			theme.StyleDim.Render("No check-in recorded for this date"),
		}
	}

	lines := []string{
		theme.StyleHeader.Render("Check-in"),
		fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Mood"), state.DailyCheckIn.Mood),
		fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Energy"), state.DailyCheckIn.Energy),
	}
	if state.DailyCheckIn.SleepHours != nil {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %s",
				theme.StyleHeader.Render("Sleep"),
				helperpkg.FormatCompactDurationHours(*state.DailyCheckIn.SleepHours),
			),
		)
	}
	if state.DailyCheckIn.SleepScore != nil {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %d/100",
				theme.StyleHeader.Render("Sleep Score"),
				*state.DailyCheckIn.SleepScore,
			),
		)
	}
	if state.DailyCheckIn.ScreenTimeMinutes != nil {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %s",
				theme.StyleHeader.Render("Screen Time"),
				helperpkg.FormatCompactDurationMinutes(*state.DailyCheckIn.ScreenTimeMinutes),
			),
		)
	}
	if state.DailyCheckIn.Notes != nil && strings.TrimSpace(*state.DailyCheckIn.Notes) != "" {
		noteWidth := width - 8
		if compact {
			noteWidth = width - 6
		}
		lines = append(
			lines,
			"",
			theme.StyleHeader.Render("Notes"),
			viewhelpers.Truncate(strings.TrimSpace(*state.DailyCheckIn.Notes), max(20, noteWidth)),
		)
	}
	if !countsForCheckInStreak(state.DailyCheckIn) {
		lines = append(
			lines,
			"",
			theme.StyleError.Render("Backfilled check-in"),
			theme.StyleDim.Render("Recorded later, so it does not count toward the same-day streak."),
		)
	}
	return lines
}

func appendSection(lines, section []string) []string {
	if len(section) == 0 {
		return lines
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	return append(lines, section...)
}

func trendsBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact)
}

func metricsBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact)
}

func trendsBodyLinesWithStreaks(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	if state.MetricsRollup == nil {
		return []string{theme.StyleDim.Render("Loading metrics...")}
	}

	lines := []string{}
	if compact {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %d  %s  %d",
				theme.StyleHeader.Render("Days"),
				state.MetricsRollup.Days,
				theme.StyleHeader.Render("Check-ins"),
				state.MetricsRollup.CheckInDays,
			),
			fmt.Sprintf(
				"%s  %d  %s  %s",
				theme.StyleHeader.Render("Focus"),
				state.MetricsRollup.FocusDays,
				theme.StyleHeader.Render("Worked"),
				viewhelpers.FormatClockText(state.MetricsRollup.WorkedSeconds),
			),
		)
		if state.MetricsRollup.HabitDueCount > 0 || state.MetricsRollup.HabitCompletedCount > 0 ||
			state.MetricsRollup.HabitFailedCount > 0 {
			lines = append(
				lines,
				fmt.Sprintf(
					"%s  %d  %s  %d  %s  %d",
					theme.StyleHeader.Render("Habits due"),
					state.MetricsRollup.HabitDueCount,
					theme.StyleHeader.Render("Done"),
					state.MetricsRollup.HabitCompletedCount,
					theme.StyleHeader.Render("Failed"),
					state.MetricsRollup.HabitFailedCount,
				),
			)
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
			lines = append(
				lines,
				fmt.Sprintf(
					"%s  %s  %s  %s",
					theme.StyleHeader.Render("Mood"),
					avgMood,
					theme.StyleHeader.Render("Energy"),
					avgEnergy,
				),
			)
		}
		if canvas := trendCanvas(theme, state, width-6, state.Height); len(canvas) > 0 {
			lines = append(lines, canvas...)
		} else if strips := trendStrips(theme, state); len(strips) > 0 {
			lines = append(lines, theme.StyleHeader.Render(fmt.Sprintf("Signals (%s)", wellbeingWindowLabel(state))))
			lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(strips[0], width-6)))
		}
		return lines
	}

	lines = append(lines, trendCards(theme, state)...)
	if state.MetricsRollup.AverageMood != nil {
		lines = append(
			lines,
			fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Mood"), *state.MetricsRollup.AverageMood),
		)
	}
	if state.MetricsRollup.AverageEnergy != nil {
		lines = append(
			lines,
			fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Energy"), *state.MetricsRollup.AverageEnergy),
		)
	}
	if state.MetricsRollup.HabitDueCount > 0 || state.MetricsRollup.HabitCompletedCount > 0 ||
		state.MetricsRollup.HabitFailedCount > 0 {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %d  %s  %d  %s  %d",
				theme.StyleHeader.Render("Habits due"),
				state.MetricsRollup.HabitDueCount,
				theme.StyleHeader.Render("Done"),
				state.MetricsRollup.HabitCompletedCount,
				theme.StyleHeader.Render("Failed"),
				state.MetricsRollup.HabitFailedCount,
			),
		)
	}
	if canvas := trendCanvas(theme, state, width-6, state.Height); len(canvas) > 0 {
		lines = append(lines, "")
		lines = append(lines, canvas...)
	} else if strips := trendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render(fmt.Sprintf("Signals (%s)", wellbeingWindowLabel(state))))
		lines = append(lines, strips...)
	}
	return lines
}
