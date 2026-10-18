package review

import (
	"fmt"
	"strings"

	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// StripRows sums the window: work over wellbeing, column groups dropped right-first.
func StripRows(theme types.Theme, state types.ContentState, width int) []string {
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
	rows := []string{"", ""}
	for _, group := range groups {
		columnWidth := max(ansi.StringWidth(group[0]), ansi.StringWidth(group[1]))
		gap := ""
		if rows[0] != "" {
			gap = "   "
		}
		if ansi.StringWidth(rows[0])+len(gap)+columnWidth > width {
			break
		}
		for i := range rows {
			rows[i] += gap + group[i] + strings.Repeat(" ", columnWidth-ansi.StringWidth(group[i]))
		}
	}
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], width, "")
		rows[i] += strings.Repeat(" ", width-ansi.StringWidth(rows[i]))
	}
	return rows
}

func scoreBar(theme types.Theme, score int) string {
	filled := max(0, min(10, score/10))
	return theme.StyleHeader.Render(strings.Repeat("█", filled)) + theme.StyleDim.Render(strings.Repeat("░", 10-filled))
}
