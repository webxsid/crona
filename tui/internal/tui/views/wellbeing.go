package views

import (
	"fmt"
	"sort"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderWellbeingView(theme Theme, state ContentState) string {
	if state.Height < 37 {
		return renderWellbeingCompactView(theme, state)
	}
	topH, bottomH := splitVertical(state.Height, 11, 8, state.Height/2)
	return lipgloss.JoinVertical(lipgloss.Left,
		renderWellbeingSummary(theme, state, state.Width, topH),
		renderWellbeingTrends(theme, state, state.Width, bottomH),
	)
}

func renderWellbeingCompactView(theme Theme, state ContentState) string {
	topH := max(10, state.Height*3/5)
	if topH > state.Height-6 {
		topH = state.Height - 6
	}
	bottomH := max(6, state.Height-topH)
	return lipgloss.JoinVertical(lipgloss.Left,
		renderWellbeingCompactSummary(theme, state, state.Width, topH),
		renderWellbeingCompactTrends(theme, state, state.Width, bottomH),
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
	lines = append(lines, wellbeingAccountabilityLines(theme, state)...)
	if burnout := latestBurnout(state); burnout != nil {
		lines = append(lines, "",
			fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)),
			theme.StyleDim.Render(burnoutSummary(burnout)),
		)
	}
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func renderWellbeingCompactSummary(theme Theme, state ContentState, width, height int) string {
	dateText := state.WellbeingDate
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(dateText)),
		renderActionLine(theme, width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})),
	}
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded for this date"))
	} else {
		lines = append(lines,
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Mood"), state.DailyCheckIn.Mood),
			fmt.Sprintf("%s  %d/5", theme.StyleHeader.Render("Energy"), state.DailyCheckIn.Energy),
		)
		if state.DailyCheckIn.SleepHours != nil {
			lines = append(lines, fmt.Sprintf("%s  %.1fh", theme.StyleHeader.Render("Sleep"), *state.DailyCheckIn.SleepHours))
		} else if state.DailyCheckIn.SleepScore != nil {
			lines = append(lines, fmt.Sprintf("%s  %d/100", theme.StyleHeader.Render("Sleep"), *state.DailyCheckIn.SleepScore))
		}
	}
	lines = append(lines, wellbeingCompactAccountabilityLines(theme, state)...)
	if burnout := latestBurnout(state); burnout != nil {
		lines = append(lines,
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

func renderWellbeingCompactTrends(theme Theme, state ContentState, width, height int) string {
	lines := []string{theme.StylePaneTitle.Render("Metrics Window")}
	if state.MetricsRollup == nil {
		lines = append(lines, theme.StyleDim.Render("Loading metrics..."))
		return renderPaneBox(theme, false, width, height, stringsJoin(lines))
	}
	lines = append(lines,
		fmt.Sprintf("%s  %d  %s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days, theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
		fmt.Sprintf("%s  %d  %s  %s", theme.StyleHeader.Render("Focus"), state.MetricsRollup.FocusDays, theme.StyleHeader.Render("Worked"), formatClock(state.MetricsRollup.WorkedSeconds)),
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
		lines = append(lines,
			fmt.Sprintf("Check-in %d/%d  Focus %d/%d", state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays),
		)
	}
	if burnout := latestBurnout(state); burnout != nil {
		factors := burnoutFactorLines(burnout)
		if len(factors) > 0 {
			lines = append(lines, theme.StyleDim.Render(truncate(factors[0], width-6)))
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

func wellbeingAccountabilityLines(theme Theme, state ContentState) []string {
	planned, completed, failed, abandoned, pending := dailyPlanCounts(state.DailyPlan)
	score, pressure, delayed, highRisk := dailyPlanSignals(state.DailyPlan)
	if planned == 0 && completed == 0 && failed == 0 && abandoned == 0 && pending == 0 && delayed == 0 && score == 0 {
		return nil
	}
	lines := []string{
		"",
		fmt.Sprintf("%s  planned %d   completed %d   failed %d", theme.StyleHeader.Render("Accountability"), planned, completed, failed),
		theme.StyleDim.Render(fmt.Sprintf("pending rollback %d   abandoned %d", pending, abandoned)),
	}
	if score > 0 || delayed > 0 || highRisk > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("score %.1f   backlog %.1f   delayed %d   high risk %d", score, pressure, delayed, highRisk)))
	}
	for _, item := range recentPlanFailureLines(state.DailyPlan, 2) {
		lines = append(lines, theme.StyleDim.Render(item))
	}
	return lines
}

func wellbeingCompactAccountabilityLines(theme Theme, state ContentState) []string {
	planned, completed, failed, _, pending := dailyPlanCounts(state.DailyPlan)
	score, _, delayed, _ := dailyPlanSignals(state.DailyPlan)
	if planned == 0 && completed == 0 && failed == 0 && pending == 0 && delayed == 0 && score == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("%s  p%d c%d f%d r%d s%.1f d%d", theme.StyleHeader.Render("Accountability"), planned, completed, failed, pending, score, delayed),
	}
}

func dailyPlanCounts(plan *api.DailyPlan) (planned, completed, failed, abandoned, pending int) {
	if plan == nil {
		return 0, 0, 0, 0, 0
	}
	if plan.Summary.PlannedCount > 0 || plan.Summary.CompletedCount > 0 || plan.Summary.FailedCount > 0 || plan.Summary.AbandonedCount > 0 || plan.Summary.PendingRollbackCount > 0 {
		return plan.Summary.PlannedCount, plan.Summary.CompletedCount, plan.Summary.FailedCount, plan.Summary.AbandonedCount, plan.Summary.PendingRollbackCount
	}
	for _, entry := range plan.Entries {
		planned++
		if entry.PendingFailureAt != nil {
			pending++
		}
		switch entry.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "abandoned":
			abandoned++
		}
	}
	return planned, completed, failed, abandoned, pending
}

func dailyPlanSignals(plan *api.DailyPlan) (score, pressure float64, delayed, highRisk int) {
	if plan == nil {
		return 0, 0, 0, 0
	}
	return plan.Summary.AccountabilityScore, plan.Summary.BacklogPressure, plan.Summary.DelayedIssueCount, plan.Summary.HighRiskIssueCount
}

func recentPlanFailureLines(plan *api.DailyPlan, limit int) []string {
	if plan == nil || limit <= 0 {
		return nil
	}
	lines := make([]string, 0, limit)
	for _, entry := range plan.Entries {
		if entry.Status == "failed" && entry.FailureReason != nil {
			lines = append(lines, fmt.Sprintf("- issue #%d marked failed (%s)", entry.IssueID, *entry.FailureReason))
			if len(lines) >= limit {
				break
			}
		}
	}
	return lines
}
