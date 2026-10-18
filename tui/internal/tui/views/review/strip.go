package review

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

// StripRows sums the window: work over wellbeing, column groups dropped right-first.
func StripRows(theme types.Theme, state types.ContentState, width int) []string {
	days := state.ReviewWindowDays
	if days < 1 {
		days = 7
	}
	return viewchrome.WithWindowHint(theme, days, width, func(width int) []string { return stripRows(theme, state, width) })
}

func stripRows(theme types.Theme, state types.ContentState, width int) []string {
	if width <= 0 {
		return []string{"", ""}
	}
	label := func(s string) string { return theme.StyleHeader.Render(s) }
	worked, focus, plan, outcome := "–", "–", "–", "done –  failed –"
	if rollup := state.MetricsRollup; rollup != nil {
		worked = viewhelpers.FormatCompactDurationSeconds(rollup.WorkedSeconds)
	}
	if state.WeeklyFocusScore != nil {
		focus = fmt.Sprintf("%d  %s", state.WeeklyFocusScore.Score, scoreBar(theme, state.WeeklyFocusScore.Score))
	}
	if window := state.DashboardWindow; window != nil {
		if window.PlannedCount > 0 {
			plan = fmt.Sprintf("%d%%", window.CompletedCount*100/window.PlannedCount)
		}
		outcome = fmt.Sprintf("done %d  failed %d", window.CompletedCount, window.FailedCount)
	}
	mood, energy, sleep, checkIns, risk := "–", "–", "–", "–", "–"
	if rollup := state.MetricsRollup; rollup != nil {
		if rollup.AverageMood != nil {
			mood = fmt.Sprintf("%.1f", *rollup.AverageMood)
		}
		if rollup.AverageEnergy != nil {
			energy = fmt.Sprintf("%.1f", *rollup.AverageEnergy)
		}
		if rollup.AverageSleepHours != nil {
			sleep = fmt.Sprintf("%.1fh", *rollup.AverageSleepHours)
		}
		checkIns = fmt.Sprintf("%d/%d", rollup.CheckInDays, rollup.Days)
		if rollup.LatestBurnout != nil {
			risk = string(rollup.LatestBurnout.Level)
		}
	}
	groups := [][2]string{
		{label("Worked") + " " + worked, label("Mood") + " " + mood},
		{label("Focus") + " " + focus, label("Energy") + " " + energy + "  " + label("Sleep") + " " + sleep},
		{label("Plan") + " " + plan, label("Check-ins") + " " + checkIns},
		{outcome, label("Risk") + " " + risk},
	}
	return viewchrome.StripRows(groups, width)
}

func scoreBar(theme types.Theme, score int) string {
	filled := max(0, min(10, score/10))
	return theme.StyleHeader.Render(strings.Repeat("█", filled)) + theme.StyleDim.Render(strings.Repeat("░", 10-filled))
}
