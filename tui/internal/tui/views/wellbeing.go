package views

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderWellbeingView(theme Theme, state ContentState) string {
	if state.Height < 30 {
		return renderWellbeingSmallScreenView(theme, state)
	}
	if state.Height < 37 {
		return renderWellbeingCompactView(theme, state)
	}
	topH, bottomH := splitVertical(state.Height, 11, 8, state.Height/2)
	return lipgloss.JoinVertical(lipgloss.Left,
		renderWellbeingSummary(theme, state, state.Width, topH),
		renderWellbeingTrends(theme, state, state.Width, bottomH),
	)
}

func renderWellbeingSmallScreenView(theme Theme, state ContentState) string {
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(state.WellbeingDate)),
		renderActionLine(theme, state.Width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane, RestModeActive: state.RestModeActive, AwayModeActive: state.AwayModeActive})),
	}
	lines = append(lines, wellbeingCompactCards(theme, state)...)
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in for selected date"))
	}
	lines = append(lines, wellbeingCompactAccountabilityLines(theme, state)...)
	lines = append(lines, wellbeingCompactRiskSnapshotLines(theme, state)...)
	if strips := wellbeingTrendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, theme.StyleHeader.Render("Signals"))
		lines = append(lines, strips...)
	}
	if activity := wellbeingCompactHeatmap(theme, state); len(activity) > 0 {
		lines = append(lines, theme.StyleHeader.Render("Activity"))
		lines = append(lines, activity...)
	}
	return renderPaneBox(theme, false, state.Width, state.Height, stringsJoin(lines))
}

func renderWellbeingCompactView(theme Theme, state ContentState) string {
	topH := max(10, state.Height*11/20)
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
		renderActionLine(theme, width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane, RestModeActive: state.RestModeActive, AwayModeActive: state.AwayModeActive})),
		"",
	}
	lines = append(lines, wellbeingCards(theme, state, width)...)
	if heatmap := wellbeingHeatmap(theme, state, width); len(heatmap) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Recent Activity"))
		lines = append(lines, heatmap...)
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
	lines = append(lines, wellbeingRiskSnapshotLines(theme, state)...)
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func renderWellbeingCompactSummary(theme Theme, state ContentState, width, height int) string {
	dateText := state.WellbeingDate
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Wellbeing"), theme.StyleHeader.Render(dateText)),
		renderActionLine(theme, width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane, RestModeActive: state.RestModeActive, AwayModeActive: state.AwayModeActive})),
	}
	lines = append(lines, wellbeingCompactCards(theme, state)...)
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded for this date"))
	}
	lines = append(lines, wellbeingCompactAccountabilityLines(theme, state)...)
	lines = append(lines, wellbeingCompactRiskSnapshotLines(theme, state)...)
	if heatmap := wellbeingCompactHeatmap(theme, state); len(heatmap) > 0 {
		lines = append(lines, theme.StyleHeader.Render("Activity"))
		lines = append(lines, heatmap...)
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
	lines = append(lines, wellbeingTrendCards(theme, state)...)
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
	if strips := wellbeingTrendStrips(theme, state); len(strips) > 0 {
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
	if strips := wellbeingTrendStrips(theme, state); len(strips) > 0 {
		lines = append(lines, theme.StyleDim.Render(truncate(strips[0], width-6)))
	}
	if burnout := latestBurnout(state); burnout != nil {
		risks, _ := burnoutContributorLines(burnout)
		if len(risks) > 0 {
			lines = append(lines, theme.StyleDim.Render(truncate(risks[0], width-6)))
		}
	}
	return renderPaneBox(theme, false, width, height, stringsJoin(lines))
}

func wellbeingCards(theme Theme, state ContentState, width int) []string {
	cards := []string{
		wellbeingCard(theme, "Mood", wellbeingMoodLabel(state)),
		wellbeingCard(theme, "Energy", wellbeingEnergyLabel(state)),
		wellbeingCard(theme, "Sleep", wellbeingSleepLabel(state)),
		wellbeingCard(theme, "Worked", wellbeingWorkedLabel(state)),
		wellbeingCard(theme, "Streaks", wellbeingStreakLabel(state)),
	}
	return wrapJoinedCards(cards, max(24, width-6))
}

func wellbeingCompactCards(theme Theme, state ContentState) []string {
	return []string{
		fmt.Sprintf("%s  %s   %s  %s", theme.StyleHeader.Render("Mood"), wellbeingMoodLabel(state), theme.StyleHeader.Render("Energy"), wellbeingEnergyLabel(state)),
		fmt.Sprintf("%s  %s   %s  %s", theme.StyleHeader.Render("Sleep"), wellbeingSleepLabel(state), theme.StyleHeader.Render("Worked"), wellbeingWorkedLabel(state)),
	}
}

func wellbeingTrendCards(theme Theme, state ContentState) []string {
	return []string{
		fmt.Sprintf("%s  %d   %s  %d", theme.StyleHeader.Render("Days"), state.MetricsRollup.Days, theme.StyleHeader.Render("Check-ins"), state.MetricsRollup.CheckInDays),
		fmt.Sprintf("%s  %d   %s  %s", theme.StyleHeader.Render("Focus Days"), state.MetricsRollup.FocusDays, theme.StyleHeader.Render("Worked"), formatClock(state.MetricsRollup.WorkedSeconds)),
		fmt.Sprintf("%s  %s   %s  %d", theme.StyleHeader.Render("Rest"), formatClock(state.MetricsRollup.RestSeconds), theme.StyleHeader.Render("Sessions"), state.MetricsRollup.SessionCount),
	}
}

func wellbeingCard(theme Theme, title, value string) string {
	return theme.StyleHeader.Render(title) + " " + theme.StyleNormal.Render(value)
}

func wrapJoinedCards(cards []string, maxWidth int) []string {
	if len(cards) == 0 {
		return nil
	}
	lines := []string{}
	current := ""
	for _, card := range cards {
		if current == "" {
			current = card
			continue
		}
		candidate := current + "   " + card
		if lipgloss.Width(candidate) > maxWidth {
			lines = append(lines, current)
			current = card
			continue
		}
		current = candidate
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func wellbeingMoodLabel(state ContentState) string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		if state.MetricsRollup != nil && state.MetricsRollup.AverageMood != nil {
			return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageMood)
		}
		return "-"
	}
	return fmt.Sprintf("%d/5", state.DailyCheckIn.Mood)
}

func wellbeingEnergyLabel(state ContentState) string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		if state.MetricsRollup != nil && state.MetricsRollup.AverageEnergy != nil {
			return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageEnergy)
		}
		return "-"
	}
	return fmt.Sprintf("%d/5", state.DailyCheckIn.Energy)
}

func wellbeingSleepLabel(state ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.SleepHours != nil {
		return fmt.Sprintf("%.1fh", *state.DailyCheckIn.SleepHours)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageSleepHours != nil {
		return fmt.Sprintf("avg %.1fh", *state.MetricsRollup.AverageSleepHours)
	}
	return "-"
}

func wellbeingWorkedLabel(state ContentState) string {
	if state.MetricsRollup == nil {
		return "-"
	}
	return formatClock(state.MetricsRollup.WorkedSeconds)
}

func wellbeingStreakLabel(state ContentState) string {
	if state.Streaks == nil {
		return "-"
	}
	return fmt.Sprintf("C%d/%d F%d/%d", state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays)
}

func wellbeingHeatmap(theme Theme, state ContentState, width int) []string {
	if len(state.MetricsRange) == 0 || width < 48 {
		return nil
	}
	rows := wellbeingHeatmapRows(state.MetricsRange)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, theme.StyleDim.Render(row))
	}
	return out
}

func wellbeingCompactHeatmap(theme Theme, state ContentState) []string {
	if len(state.MetricsRange) < 7 {
		return nil
	}
	glyphs := " .:-=+*#"
	window := state.MetricsRange
	if len(window) > 7 {
		window = window[len(window)-7:]
	}
	row := ""
	for _, day := range window {
		row += string(heatmapGlyph(glyphs, day)) + " "
	}
	row = strings.TrimSpace(row)
	if row == "" {
		return nil
	}
	return []string{
		theme.StyleDim.Render("Mon Tue Wed Thu Fri Sat Sun"),
		theme.StyleDim.Render(strings.ReplaceAll(row, " ", "   ")),
		theme.StyleDim.Render("low " + strings.TrimSpace(glyphs) + " high"),
	}
}

func wellbeingHeatmapRows(days []api.DailyMetricsDay) []string {
	const columns = 7
	glyphs := " .:-=+*#"
	rows := []string{"Mon Tue Wed Thu Fri Sat Sun"}
	line := ""
	weekdayCount := 0
	for _, day := range days {
		if weekdayCount == columns {
			rows = append(rows, strings.TrimSpace(line))
			line = ""
			weekdayCount = 0
		}
		line += string(heatmapGlyph(glyphs, day)) + "   "
		weekdayCount++
	}
	if strings.TrimSpace(line) != "" {
		rows = append(rows, strings.TrimSpace(line))
	}
	rows = append(rows, "Scale  low "+strings.TrimSpace(glyphs)+" high")
	return rows
}

func heatmapGlyph(glyphs string, day api.DailyMetricsDay) byte {
	score := 0
	if day.CheckIn != nil {
		score += 1
	}
	if day.SessionCount > 0 {
		score += min(4, day.SessionCount)
	}
	if day.WorkedSeconds >= 1800 {
		score += 1
	}
	if score < 0 {
		score = 0
	}
	if score >= len(glyphs) {
		score = len(glyphs) - 1
	}
	return glyphs[score]
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
		return "Risk is elevated. Reduce load and prioritize recovery today."
	case "guarded":
		return "Signals are mixed. Keep scope tight and protect breaks."
	default:
		return "Current signals look stable. Maintain your recovery rhythm."
	}
}

func burnoutContributorLines(burnout *api.BurnoutIndicator) (risks []string, recoveries []string) {
	type factor struct {
		name  string
		score float64
	}
	positive := make([]factor, 0, len(burnout.Factors))
	negative := make([]factor, 0, len(burnout.Factors))
	for name, score := range burnout.Factors {
		if score >= 0 {
			positive = append(positive, factor{name: name, score: score})
			continue
		}
		negative = append(negative, factor{name: name, score: score})
	}
	sort.Slice(positive, func(i, j int) bool { return positive[i].score > positive[j].score })
	sort.Slice(negative, func(i, j int) bool { return negative[i].score < negative[j].score })
	riskLimit := min(3, len(positive))
	recoveryLimit := min(2, len(negative))
	risks = make([]string, 0, riskLimit)
	recoveries = make([]string, 0, recoveryLimit)
	for i := 0; i < riskLimit; i++ {
		risks = append(risks, fmt.Sprintf("- %s: +%d", prettifyBurnoutFactor(positive[i].name), int(math.Round(positive[i].score*100))))
	}
	for i := 0; i < recoveryLimit; i++ {
		recoveries = append(recoveries, fmt.Sprintf("- %s: -%d", prettifyBurnoutFactor(negative[i].name), int(math.Round(math.Abs(negative[i].score)*100))))
	}
	return risks, recoveries
}

func prettifyBurnoutFactor(name string) string {
	switch name {
	case "workloadPressure":
		return "Workload pressure"
	case "breakDebt":
		return "Break debt"
	case "moodEnergyDrag":
		return "Mood/energy drag"
	case "sleepDebt":
		return "Sleep debt"
	case "recoveryConsistency":
		return "Recovery consistency"
	case "recoveryBreaks":
		return "Break recovery"
	case "loadStability":
		return "Load stability"
	default:
		return name
	}
}

func wellbeingRiskSnapshotLines(theme Theme, state ContentState) []string {
	burnout := latestBurnout(state)
	if burnout == nil && state.DailyPlan == nil {
		return nil
	}
	score, pressure, delayed, _ := dailyPlanSignals(state.DailyPlan)
	lines := []string{"", theme.StyleHeader.Render("Risk Snapshot")}
	if burnout != nil {
		lines = append(lines, fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)))
		lines = append(lines, theme.StyleDim.Render(burnoutSummary(burnout)))
	}
	lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("accountability %.1f   backlog %.1f   delayed %d", score, pressure, delayed)))
	return lines
}

func wellbeingCompactRiskSnapshotLines(theme Theme, state ContentState) []string {
	burnout := latestBurnout(state)
	if burnout == nil {
		return nil
	}
	return []string{
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)),
		theme.StyleDim.Render(burnoutSummary(burnout)),
	}
}

func wellbeingTrendStrips(theme Theme, state ContentState) []string {
	if len(state.MetricsRange) == 0 {
		return nil
	}
	window := state.MetricsRange
	if len(window) > 7 {
		window = window[len(window)-7:]
	}
	mood := make([]float64, 0, len(window))
	energy := make([]float64, 0, len(window))
	work := make([]float64, 0, len(window))
	recovery := make([]float64, 0, len(window))
	for _, day := range window {
		if day.CheckIn != nil {
			mood = append(mood, float64(day.CheckIn.Mood))
			energy = append(energy, float64(day.CheckIn.Energy))
		} else {
			mood = append(mood, 0)
			energy = append(energy, 0)
		}
		work = append(work, float64(day.WorkedSeconds)/3600.0)
		sleep := 0.0
		if day.CheckIn != nil {
			if day.CheckIn.SleepHours != nil {
				sleep = clamp01(*day.CheckIn.SleepHours / 8.0)
			} else if day.CheckIn.SleepScore != nil {
				sleep = clamp01(float64(*day.CheckIn.SleepScore) / 100.0)
			}
		}
		breakRatio := 0.0
		if day.WorkedSeconds > 0 {
			breakRatio = float64(day.RestSeconds) / float64(day.WorkedSeconds)
		}
		recovery = append(recovery, clamp01((sleep+clamp01(breakRatio/0.2))/2.0))
	}
	return []string{
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Mood"), sparkline(mood, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Energy"), sparkline(energy, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Work"), sparkline(work, 0.0, 8.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Recovery"), sparkline(recovery, 0.0, 1.0)),
	}
}

func sparkline(values []float64, minValue, maxValue float64) string {
	if len(values) == 0 {
		return "-"
	}
	glyphs := []rune("▁▂▃▄▅▆▇█")
	span := maxValue - minValue
	if span <= 0 {
		span = 1
	}
	out := make([]rune, 0, len(values))
	for _, value := range values {
		norm := clamp01((value - minValue) / span)
		idx := int(math.Round(norm * float64(len(glyphs)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(glyphs) {
			idx = len(glyphs) - 1
		}
		out = append(out, glyphs[idx])
	}
	return string(out)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
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
		fmt.Sprintf("%s  Planned %d  Completed %d  Failed %d  Pending %d", theme.StyleHeader.Render("Accountability"), planned, completed, failed, pending),
		theme.StyleDim.Render(fmt.Sprintf("Score %.1f  Delayed %d", score, delayed)),
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
