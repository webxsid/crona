package wellbeing

import (
	"fmt"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewmomentum "crona/tui/internal/tui/views/momentum"
	types "crona/tui/internal/tui/views/types"
)

func summaryBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
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
		if state.DailyCheckIn.Notes != nil && *state.DailyCheckIn.Notes != "" {
			lines = append(
				lines,
				"",
				theme.StyleHeader.Render("Notes"),
				viewhelpers.Truncate(*state.DailyCheckIn.Notes, max(20, width-8)),
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
	}
	lines = append(lines, Accountability(types.ViewSizeStandard, theme, state)...)
	lines = append(lines, RiskSnapshot(types.ViewSizeStandard, theme, state)...)
	return lines
}

func trendsBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact, false)
}

func metricsBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	return trendsBodyLinesWithStreaks(theme, state, width, compact, false)
}

func trendsBodyLinesWithStreaks(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
	includeStreaks bool,
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
				lines = append(
					lines,
					theme.StyleDim.Render(viewhelpers.Truncate(risks[0], width-6)),
				)
			}
		}
		return lines
	}

	lines = append(lines, trendCards(theme, state)...)
	if state.MetricsRollup.AverageMood != nil {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %.1f",
				theme.StyleHeader.Render("Avg Mood"),
				*state.MetricsRollup.AverageMood,
			),
		)
	}
	if state.MetricsRollup.AverageEnergy != nil {
		lines = append(
			lines,
			fmt.Sprintf(
				"%s  %.1f",
				theme.StyleHeader.Render("Avg Energy"),
				*state.MetricsRollup.AverageEnergy,
			),
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

func streaksBodyLines(
	theme types.Theme,
	state types.ContentState,
	width int,
	compact bool,
) []string {
	if state.Streaks == nil {
		return []string{theme.StyleDim.Render("Loading momentum...")}
	}
	if compact {
		lines := []string{
			viewmomentum.CompactRow(
				theme,
				"Check-ins",
				"",
				sharedtypes.HabitStreakPeriodDay,
				state.Streaks.CurrentCheckInDays,
				state.Streaks.LongestCheckInDays,
				width,
			),
			viewmomentum.CompactRow(
				theme,
				"Focus",
				"",
				sharedtypes.HabitStreakPeriodDay,
				state.Streaks.CurrentFocusDays,
				state.Streaks.LongestFocusDays,
				width,
			),
		}
		for _, streak := range state.Streaks.CustomHabitStreaks {
			lines = append(
				lines,
				viewmomentum.CompactRow(
					theme,
					streak.Name,
					viewmomentum.CadenceLabel(streak.Period),
					streak.Period,
					streak.Current,
					streak.Longest,
					width,
				),
			)
		}
		return lines
	}
	lines := []string{}
	lines = append(
		lines,
		viewmomentum.Row(
			theme,
			"Check-ins",
			"",
			sharedtypes.HabitStreakPeriodDay,
			state.Streaks.CurrentCheckInDays,
			state.Streaks.LongestCheckInDays,
			width,
		)...)
	lines = append(
		lines,
		viewmomentum.Row(
			theme,
			"Focus",
			"",
			sharedtypes.HabitStreakPeriodDay,
			state.Streaks.CurrentFocusDays,
			state.Streaks.LongestFocusDays,
			width,
		)...)
	lines = append(
		lines,
		"",
		theme.StyleHeader.Render(viewhelpers.Truncate("Custom Momentum", width-6)),
	)
	if len(state.Streaks.CustomHabitStreaks) == 0 {
		lines = append(lines, theme.StyleDim.Render("No custom momentum yet"))
	}
	for _, streak := range state.Streaks.CustomHabitStreaks {
		lines = append(
			lines,
			viewmomentum.Row(
				theme,
				streak.Name,
				viewmomentum.CadenceLabel(streak.Period),
				streak.Period,
				streak.Current,
				streak.Longest,
				width,
			)...)
	}
	return lines
}
