package wellbeing

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderSmallScreen(theme types.Theme, state types.ContentState) string {
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(state.WellbeingDate)),
		viewchrome.RenderActionLine(theme, state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})),
	}
	lines = append(lines, compactCards(theme, state)...)
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in for selected date"))
	} else if !countsForCheckInStreak(state.DailyCheckIn) {
		lines = append(lines, theme.StyleError.Render("Backfilled check-in"))
	}
	lines = append(lines, Accountability(types.ViewSizeCompact, theme, state)...)
	lines = append(lines, RiskSnapshot(types.ViewSizeCompact, theme, state)...)
	if strips := trendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, theme.StyleHeader.Render("Signals"))
		lines = append(lines, strips...)
	}
	if activity := Heatmap(types.ViewSizeCompact, theme, state); len(activity) > 0 {
		lines = append(lines, theme.StyleHeader.Render("Activity"))
		lines = append(lines, activity...)
	}
	return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}

func renderCompact(theme types.Theme, state types.ContentState) string {
	topH := max(10, state.Height*11/20)
	if topH > state.Height-6 {
		topH = state.Height - 6
	}
	bottomH := max(6, state.Height-topH)
	return lipglossJoinCompact(theme, state, topH, bottomH)
}

func lipglossJoinCompact(theme types.Theme, state types.ContentState, topH, bottomH int) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderCompactSummary(theme, state, state.Width, topH),
		renderCompactTrends(theme, state, state.Width, bottomH),
	)
}

func renderSummary(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Wellbeing"),
		theme.StylePaneTitle.Render(fmt.Sprintf("date: %s", state.WellbeingDate)),
		viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})),
		"",
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
	return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
}

func renderCompactSummary(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(state.WellbeingDate)),
		viewchrome.RenderActionLine(theme, width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{
			View:           state.View,
			Pane:           state.Pane,
			RestModeActive: state.RestModeActive,
			AwayModeActive: state.AwayModeActive,
		})),
	}
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
	return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
}

func renderTrends(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if state.MetricsRollup == nil {
		lines = append(lines, theme.StyleDim.Render("Loading metrics..."))
		return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
	}
	lines = append(lines, trendCards(theme, state)...)
	if state.MetricsRollup.AverageMood != nil {
		lines = append(lines, fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Mood"), *state.MetricsRollup.AverageMood))
	}
	if state.MetricsRollup.AverageEnergy != nil {
		lines = append(lines, fmt.Sprintf("%s  %.1f", theme.StyleHeader.Render("Avg Energy"), *state.MetricsRollup.AverageEnergy))
	}
	if state.Streaks != nil {
		lines = append(lines, "",
			fmt.Sprintf("%s  %d current / %d longest", theme.StyleHeader.Render("Same-Day Check-In Streak"), state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays),
			fmt.Sprintf("%s  %d current / %d longest", theme.StyleHeader.Render("Focus Streak"), state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays),
		)
	}
	if strips := trendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Signals (7d)"))
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
	return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
}

func renderCompactTrends(theme types.Theme, state types.ContentState, width, height int) string {
	lines := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if state.MetricsRollup == nil {
		lines = append(lines, theme.StyleDim.Render("Loading metrics..."))
		return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
	}
	lines = append(lines,
		fmt.Sprintf("%s  %d  %s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days, theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
		fmt.Sprintf("%s  %d  %s  %s", theme.StyleHeader.Render("Focus"), state.MetricsRollup.FocusDays, theme.StyleHeader.Render("Worked"), viewhelpers.FormatClock(state.MetricsRollup.WorkedSeconds)),
	)
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
	if state.Streaks != nil {
		lines = append(lines, fmt.Sprintf("Check-in %d/%d  Focus %d/%d", state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays))
	}
	if strips := trendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(strips[0], width-6)))
	}
	if burnout := latestBurnout(state); burnout != nil {
		risks, _ := burnoutContributorLines(burnout)
		if len(risks) > 0 {
			lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(risks[0], width-6)))
		}
	}
	return viewchrome.RenderPaneBox(theme, false, width, height, viewhelpers.StringsJoin(lines))
}
