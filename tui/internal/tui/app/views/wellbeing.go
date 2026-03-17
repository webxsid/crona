package views

import (
	"fmt"
	"sort"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderWellbeingView(theme Theme, state ContentState) string {
	topH, bottomH := splitVertical(state.Height, 11, 8, state.Height/2)
	return lipgloss.JoinVertical(lipgloss.Left,
		renderWellbeingSummary(theme, state, state.Width, topH),
		renderWellbeingTrends(theme, state, state.Width, bottomH),
	)
}

func renderWellbeingSummary(theme Theme, state ContentState, width, height int) string {
	dateText := state.WellbeingDate
	lines := []string{
		theme.StylePaneTitle.Render("Wellbeing"),
		theme.StylePaneTitle.Render(fmt.Sprintf("date: %s", dateText)),
		renderActionLine(theme, width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})),
		"",
	}
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines,
			theme.StyleDim.Render("No check-in recorded for this date"),
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Mood"), state.DailyCheckIn.Mood),
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Energy"), state.DailyCheckIn.Energy),
		)
		if state.DailyCheckIn.SleepHours != nil {
			lines = append(lines, fmt.Sprintf("%s  %.1fh", theme.StyleHeader.Render("Sleep"), *state.DailyCheckIn.SleepHours))
		}
		if state.DailyCheckIn.SleepScore != nil {
			lines = append(lines, fmt.Sprintf("%s  %d/100", theme.StyleHeader.Render("Sleep Score"), *state.DailyCheckIn.SleepScore))
		}
		if state.DailyCheckIn.ScreenTimeMinutes != nil {
			lines = append(lines, fmt.Sprintf("%s  %dm", theme.StyleHeader.Render("Screen Time"), *state.DailyCheckIn.ScreenTimeMinutes))
		}
		if state.DailyCheckIn.Notes != nil && *state.DailyCheckIn.Notes != "" {
			lines = append(lines, "", theme.StyleHeader.Render("Notes"), truncate(*state.DailyCheckIn.Notes, max(20, width-8)))
		}
		if !countsForCheckInStreak(state.DailyCheckIn) {
			lines = append(lines, "", theme.StyleDim.Render("This check-in was backfilled later, so it does not count toward the same-day streak."))
		}
	}
	if burnout := latestBurnout(state); burnout != nil {
		lines = append(lines, "",
			fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)),
			theme.StyleDim.Render(burnoutSummary(burnout)),
		)
	}
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func renderWellbeingTrends(theme Theme, state ContentState, width, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render("Metrics Window"),
	}
	if state.MetricsRollup == nil {
		lines = append(lines, theme.StyleDim.Render("Loading metrics..."))
		return renderPaneBox(theme, false, width, height, stringsJoin(lines))
	}
	lines = append(lines,
		fmt.Sprintf("%s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days),
		fmt.Sprintf("%s  %d", theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
		fmt.Sprintf("%s  %d", theme.StyleHeader.Render("Focus Days"), state.MetricsRollup.FocusDays),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Worked"), formatClock(state.MetricsRollup.WorkedSeconds)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Rest"), formatClock(state.MetricsRollup.RestSeconds)),
	)
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
	if burnout := latestBurnout(state); burnout != nil {
		lines = append(lines, "",
			theme.StyleHeader.Render("Top Burnout Factors"),
		)
		for _, factor := range burnoutFactorLines(burnout) {
			lines = append(lines, factor)
		}
	}
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func latestBurnout(state ContentState) *api.BurnoutIndicator {
	if state.MetricsRollup == nil || state.MetricsRollup.LatestBurnout == nil {
		return nil
	}
	return state.MetricsRollup.LatestBurnout
}

func countsForCheckInStreak(checkIn *api.DailyCheckIn) bool {
	if checkIn == nil {
		return false
	}
	return len(checkIn.CreatedAt) >= 10 && checkIn.CreatedAt[:10] == checkIn.Date
}

func burnoutBadge(theme Theme, burnout *api.BurnoutIndicator) string {
	style := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	switch burnout.Level {
	case "guarded":
		style = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "high":
		style = lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
	return style.Render(fmt.Sprintf("%d %s", burnout.Score, strings.ToUpper(string(burnout.Level))))
}

func burnoutSummary(burnout *api.BurnoutIndicator) string {
	switch burnout.Level {
	case "high":
		return "Recent work and wellbeing signals are trending hot."
	case "guarded":
		return "Signals are mixed. Pace and recovery are worth watching."
	default:
		return "Current signals look stable."
	}
}

func burnoutFactorLines(burnout *api.BurnoutIndicator) []string {
	type factor struct {
		name  string
		score float64
	}
	factors := make([]factor, 0, len(burnout.Factors))
	for name, score := range burnout.Factors {
		factors = append(factors, factor{name: name, score: score})
	}
	sort.Slice(factors, func(i, j int) bool { return factors[i].score > factors[j].score })
	limit := min(3, len(factors))
	lines := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		lines = append(lines, fmt.Sprintf("- %s: %d%%", prettifyBurnoutFactor(factors[i].name), int(factors[i].score*100)))
	}
	return lines
}

func prettifyBurnoutFactor(name string) string {
	switch name {
	case "sessionDensity":
		return "Session density"
	case "breakCompliance":
		return "Break compliance"
	case "moodTrend":
		return "Mood trend"
	case "energyTrend":
		return "Energy trend"
	case "sleepRisk":
		return "Sleep risk"
	default:
		return name
	}
}
