// Package summary renders the Day summary: a day's focus stats, plan accountability,
// check-in and streaks, shown in Daily's detail pane or its Day dialog.
package summary

import (
	"fmt"
	"sort"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DayLines is the Day summary at width, one stat per row: the day's focus stats, plan
// accountability,
// check-in and streaks. It waits on the summary snapshot for the shown date.
func DayLines(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate {
		return []string{theme.StyleDim.Render("Loading summary…")}
	}
	lines := strings.Split(renderStats(theme, snapshot, plannedStatusCounts(state), width), "\n")
	lines = append(lines, "")
	lines = append(lines, renderPlan(theme, snapshot.Plan, snapshot.Metrics, plannedStatusCounts(state), width)...)
	lines = append(lines, renderTopPenalties(theme, snapshot.Plan, snapshot.Issues, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWellbeing(theme, snapshot.CheckIn, snapshot.Rollup, state.Settings, width)...)
	lines = append(lines, "")
	return append(lines, renderMomentum(theme, state, snapshot.Streaks, width)...)
}

// labelWidth is the label column every summary row lines its value up after.
const labelWidth = 10

// Row is one stat: a dim label column, then its value cut to the room left.
func Row(theme types.Theme, label, value string, width int) string {
	return theme.StyleDim.Render(fmt.Sprintf("%-*s", labelWidth, label)) +
		viewhelpers.TruncateANSI(value, max(1, width-labelWidth))
}

func renderStats(theme types.Theme, snapshot *api.SummarySnapshot, statusCounts map[string]int, width int) string {
	if snapshot == nil {
		return ""
	}
	lines := []string{theme.StylePaneTitle.Render("Stats")}
	if snapshot.Focus == nil || snapshot.Metrics == nil {
		return strings.Join(append(lines, theme.StyleDim.Render("Stats unavailable.")), "\n")
	}
	focus := snapshot.Focus
	metrics := snapshot.Metrics
	duration := helperpkg.FormatCompactDurationSeconds
	lines = headed(lines[0], focusVerdict(theme, focus), width)
	target := max(1, focus.TargetWorkedSeconds)
	workedFraction := float64(metrics.WorkedSeconds) / float64(target)
	workedColor := theme.ColorCyan
	if workedFraction >= 1 {
		workedColor = theme.ColorGreen
	}
	arcs := renderGauges(theme, []gauge{
		{
			label: "Focus", value: fmt.Sprintf("%d", focus.Score),
			fraction: float64(focus.Score) / 100, color: focusColor(theme, focus.Level),
		},
		{
			label: "Worked", value: duration(metrics.WorkedSeconds),
			fraction: workedFraction, color: workedColor,
		},
	}, gaugeCols*2+gaugeGap)
	sessions := "sessions"
	if metrics.SessionCount == 1 {
		sessions = "session"
	}
	rest := theme.StyleNormal.Render(duration(metrics.RestSeconds))
	if metrics.RestSeconds == 0 && metrics.WorkedSeconds >= 45*60 {
		rest = lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(duration(metrics.RestSeconds)) +
			theme.StyleDim.Render("  aim ~1 break an hour")
	}
	facts := []string{
		factRow(theme, "Worked", workedText(metrics.WorkedSeconds, focus.TargetWorkedSeconds)+
			theme.StyleDim.Render(fmt.Sprintf(" · %d %s", metrics.SessionCount, sessions))),
		factRow(theme, "Rest", rest),
	}
	lines = append(lines, besideArcs(arcs, facts, gaugeCols*2+gaugeGap, width)...)
	return strings.Join(lines, "\n")
}

func renderPlan(
	theme types.Theme,
	plan *sharedtypes.DailyPlan,
	metrics *api.DailyMetricsDay,
	statusCounts map[string]int,
	width int,
) []string {
	title := theme.StylePaneTitle.Render("Accountability")
	hasPlan := plan != nil && strings.TrimSpace(plan.Date) != ""
	var lines, facts []string
	if hasPlan {
		s := plan.Summary
		lines = headed(title, planVerdict(theme, s), width)
		facts = []string{
			factRow(theme, "Slipping", riskDots(theme, s)),
			factRow(theme, "Delayed", delayText(theme, s)),
			factRow(theme, "Resolved", resolvedChips(theme, s)),
		}
	} else {
		lines = []string{title}
		facts = []string{theme.StyleDim.Render("No committed plan.")}
	}
	if metrics == nil {
		return append(lines, facts...)
	}
	// The arcs carry the done counts, so the indicator rows below list only the rest.
	issueFraction := float64(metrics.CompletedIssues) / float64(max(1, metrics.TotalIssues))
	habitFraction := float64(metrics.HabitCompletedCount) / float64(max(1, metrics.HabitDueCount))
	arcs := renderGauges(theme, []gauge{
		{
			label: "Issues", value: fmt.Sprintf("%d/%d", metrics.CompletedIssues, metrics.TotalIssues),
			fraction: issueFraction, color: doneColor(theme, issueFraction),
		},
		{
			label: "Habits", value: fmt.Sprintf("%d/%d", metrics.HabitCompletedCount, metrics.HabitDueCount),
			fraction: habitFraction, color: doneColor(theme, habitFraction),
		},
	}, gaugeCols*2+gaugeGap)
	lines = append(lines, besideArcs(arcs, facts, gaugeCols*2+gaugeGap, width)...)
	open := map[string]int{}
	for status, n := range statusCounts {
		if status != "done" {
			open[status] = n
		}
	}
	if legend := IssueStatusLegend(theme, open); legend != "" {
		lines = append(lines, Row(theme, "Issues", legend, width))
	}
	if legend := HabitLegend(theme, 0, metrics.HabitFailedCount, metrics.HabitDueCount-metrics.HabitCompletedCount); legend != "" {
		lines = append(lines, Row(theme, "Habits", legend, width))
	}
	return lines
}

// headed puts a verdict beside the section title, or on its own line when both don't fit.
func headed(title, verdict string, width int) []string {
	if line := title + "   " + verdict; ansi.StringWidth(line) <= width {
		return []string{line}
	}
	return []string{title, ansi.Truncate(verdict, width, "…")}
}

func workedText(worked, target int) string {
	if target <= 0 {
		return helperpkg.FormatCompactDurationSeconds(worked)
	}
	return fmt.Sprintf("%s of %s", helperpkg.FormatCompactDurationSeconds(worked), helperpkg.FormatCompactDurationSeconds(target))
}

// besideArcs puts fact lines to the right of arc gauges, or under them when the pane is narrow.
func besideArcs(arcs, facts []string, arcWidth, width int) []string {
	if width < arcWidth+gaugeGap+28 {
		return append(append([]string{}, arcs...), facts...)
	}
	lines := []string{}
	for i := range max(len(arcs), len(facts)) {
		left := strings.Repeat(" ", arcWidth)
		if i < len(arcs) {
			left = padTo(arcs[i], arcWidth)
		}
		right := ""
		if i < len(facts) {
			right = facts[i]
		}
		lines = append(lines, ansi.Truncate(left+strings.Repeat(" ", gaugeGap+1)+right, width, "…"))
	}
	return lines
}

// focusVerdict turns the focus level and reason into words, coloured by level.
func focusVerdict(theme types.Theme, focus *sharedtypes.FocusScoreSummary) string {
	reasons := map[sharedtypes.FocusScoreReason]string{
		sharedtypes.FocusScoreReasonNoActivity:   "no focus logged yet",
		sharedtypes.FocusScoreReasonUnderTarget:  "under today's target",
		sharedtypes.FocusScoreReasonNeedsBreaks:  "take a break",
		sharedtypes.FocusScoreReasonBalanced:     "work and rest balanced",
		sharedtypes.FocusScoreReasonOverextended: "past target, ease off",
	}
	head := map[sharedtypes.FocusScoreLevel]string{
		sharedtypes.FocusScoreLevelLow:          "▲ Low focus",
		sharedtypes.FocusScoreLevelSteady:       "◆ Steady",
		sharedtypes.FocusScoreLevelStrong:       "✓ Strong",
		sharedtypes.FocusScoreLevelOverextended: "▲ Overextended",
	}[focus.Level]
	if head == "" {
		head = "● Focus"
	}
	out := lipgloss.NewStyle().Foreground(focusColor(theme, focus.Level)).Bold(true).Render(head)
	if reason := reasons[focus.Reason]; reason != "" {
		out += theme.StyleDim.Render(" · " + reason)
	}
	return out
}

func focusColor(theme types.Theme, level sharedtypes.FocusScoreLevel) lipgloss.Color {
	switch level {
	case sharedtypes.FocusScoreLevelStrong:
		return theme.ColorGreen
	case sharedtypes.FocusScoreLevelSteady:
		return theme.ColorYellow
	case sharedtypes.FocusScoreLevelOverextended:
		return theme.ColorOrange
	}
	return theme.ColorRed
}

// planVerdict sums the plan up in words: at risk once anything is high risk, slipping while
// work is delayed, otherwise on track (or done once nothing is left open).
func planVerdict(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	var reasons []string
	if s.HighRiskIssueCount > 0 {
		reasons = append(reasons, fmt.Sprintf("%d high risk", s.HighRiskIssueCount))
	}
	if s.DelayedIssueCount > 0 {
		reasons = append(reasons, fmt.Sprintf("%d delayed", s.DelayedIssueCount))
	}
	reason := theme.StyleDim.Render(" · " + strings.Join(reasons, ", "))
	switch {
	case s.HighRiskIssueCount > 0:
		return lipgloss.NewStyle().Foreground(theme.ColorRed).Bold(true).Render("▲ At risk") + reason
	case s.DelayedIssueCount > 0:
		return lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render("◆ Slipping") + reason
	case s.PlannedCount > 0 && s.CompletedCount == s.PlannedCount:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("✓ All done")
	}
	return lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("● On track")
}

func factRow(theme types.Theme, label, value string) string {
	return theme.StyleDim.Render(fmt.Sprintf("%-10s", label)) + value
}

// riskDots draws one dot per planned task, high-risk ones filled red first.
func riskDots(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	if s.HighRiskIssueCount == 0 {
		return theme.StyleDim.Render("none high risk")
	}
	total := max(s.PlannedCount, s.HighRiskIssueCount)
	shown := min(total, 12)
	hot := min(shown, s.HighRiskIssueCount)
	dots := lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("●", hot)) +
		theme.StyleDim.Render(strings.Repeat("○", shown-hot))
	return dots + "  " + fmt.Sprintf("%d high risk", s.HighRiskIssueCount)
}

func delayText(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	if s.DelayedIssueCount == 0 {
		return theme.StyleDim.Render("nothing late")
	}
	color := theme.ColorYellow
	if s.MaxDelayDays >= 3 {
		color = theme.ColorOrange
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%d", s.DelayedIssueCount)) +
		theme.StyleDim.Render(fmt.Sprintf(" · avg %.1fd · max %dd", s.AvgDelayDays, s.MaxDelayDays))
}

// resolvedChips lists only the outcomes that happened, or says there were no misses.
func resolvedChips(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	var chips []string
	for _, c := range []struct {
		glyph, label string
		count        int
		color        lipgloss.Color
	}{
		{"✗", "failed", s.FailedCount, theme.ColorRed},
		{"⊘", "abandoned", s.AbandonedCount, theme.ColorOrange},
		{"↷", "carried", s.CarriedCount, theme.ColorLavender},
		{"↺", "to roll back", s.PendingRollbackCount, theme.ColorYellow},
	} {
		if c.count > 0 {
			chips = append(chips, outcome(theme, c.glyph, c.count, c.label, c.color))
		}
	}
	if len(chips) == 0 {
		return theme.StyleDim.Render("no misses yet")
	}
	return strings.Join(chips, theme.StyleDim.Render("  "))
}

// doneColor is green once every planned task is done, yellow past half and lavender before.
func doneColor(theme types.Theme, fraction float64) lipgloss.Color {
	switch {
	case fraction >= 1:
		return theme.ColorGreen
	case fraction >= 0.5:
		return theme.ColorYellow
	}
	return theme.ColorLavender
}

func outcome(theme types.Theme, glyph string, count int, label string, color lipgloss.Color) string {
	text := fmt.Sprintf("%s %d %s", glyph, count, label)
	if count == 0 {
		return theme.StyleDim.Render(text)
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %d", glyph, count)) +
		theme.StyleNormal.Render(" "+label)
}

// renderTopPenalties lists the up to three live plan entries adding most to the accountability
// score, so the issues worth doing first stand out.
func renderTopPenalties(
	theme types.Theme,
	plan *sharedtypes.DailyPlan,
	issues *sharedtypes.DailyIssueSummary,
	width int,
) []string {
	if plan == nil {
		return nil
	}
	titles := map[int64]string{}
	if issues != nil {
		for _, issue := range issues.Issues {
			titles[issue.ID] = issue.Title
		}
	}
	entries := []sharedtypes.DailyPlanEntry{}
	for _, entry := range plan.Entries {
		if entry.Penalty != nil && !entry.Penalty.Historical && entry.FailScore > 0 {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil
	}
	sort.SliceStable(
		entries,
		func(i, j int) bool { return entries[i].FailScore > entries[j].FailScore },
	)
	lines := []string{}
	for i, entry := range entries[:min(3, len(entries))] {
		label := ""
		if i == 0 {
			label = "Top"
		}
		title := titles[entry.IssueID]
		if title == "" {
			title = fmt.Sprintf("#%d", entry.IssueID)
		}
		mark := "  "
		color := theme.ColorYellow
		if entry.Penalty.HighRisk {
			mark = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("▲ ")
			color = theme.ColorOrange
		}
		why := "due today"
		if entry.CurrentDelayedDays > 0 {
			why = fmt.Sprintf("late %dd", entry.CurrentDelayedDays)
		}
		if entry.Penalty.Contribution != nil {
			why += fmt.Sprintf(" · %.0f%%", entry.Penalty.Contribution.AccountabilityPct*100)
		}
		why = lipgloss.NewStyle().Foreground(color).Render(why)
		room := max(8, width-labelWidth-2-ansi.StringWidth(why)-2)
		title = ansi.Truncate(title, room, "…")
		pad := strings.Repeat(" ", max(1, room-ansi.StringWidth(title)+2))
		lines = append(lines, Row(theme, label, mark+title+pad+why, width))
	}
	return lines
}

func renderWellbeing(
	theme types.Theme,
	checkIn *sharedtypes.DailyCheckIn,
	rollup *sharedtypes.MetricsRollup,
	settings *sharedtypes.CoreSettings,
	width int,
) []string {
	lines := []string{theme.StylePaneTitle.Render("Wellbeing")}
	if checkIn == nil || strings.TrimSpace(checkIn.Date) == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded."))
		if rollup != nil {
			if rollup.AverageMood != nil {
				lines = append(
					lines,
					Row(theme, "7d mood", fmt.Sprintf("%.1f/5", *rollup.AverageMood), width),
				)
			}
			if rollup.AverageEnergy != nil {
				lines = append(
					lines,
					Row(theme, "7d energy", fmt.Sprintf("%.1f/5", *rollup.AverageEnergy), width),
				)
			}
			if rollup.AverageSleepHours != nil {
				lines = append(
					lines,
					Row(
						theme,
						"7d sleep",
						helperpkg.FormatCompactDurationHours(*rollup.AverageSleepHours),
						width,
					),
				)
			}
		}
		return lines
	}
	var avgMood, avgEnergy, avgSleep, avgScreen *float64
	if rollup != nil {
		avgMood, avgEnergy = rollup.AverageMood, rollup.AverageEnergy
		avgSleep, avgScreen = rollup.AverageSleepHours, rollup.AverageScreenTimeMins
	}
	oneDecimal := func(v float64) string { return fmt.Sprintf("%.1f", v) }
	gauges := []gauge{
		{
			label: "Mood", value: fmt.Sprintf("%d/5", checkIn.Mood), fraction: float64(checkIn.Mood) / 5,
			color: scaleColor(theme, checkIn.Mood),
			trend: trend(theme, float64(checkIn.Mood), avgMood, 0.3, true, oneDecimal),
		},
		{
			label: "Energy", value: fmt.Sprintf("%d/5", checkIn.Energy), fraction: float64(checkIn.Energy) / 5,
			color: scaleColor(theme, checkIn.Energy),
			trend: trend(theme, float64(checkIn.Energy), avgEnergy, 0.3, true, oneDecimal),
		},
	}
	sleepTarget, screenLimit := wellbeingTargets(settings)
	if checkIn.SleepHours != nil {
		minutes := *checkIn.SleepHours * 60
		gauges = append(gauges, gauge{
			label: "Sleep", value: helperpkg.FormatCompactDurationHours(*checkIn.SleepHours),
			fraction: minutes / float64(sleepTarget),
			color:    targetColor(theme, minutes/float64(sleepTarget)),
			trend:    trend(theme, *checkIn.SleepHours, avgSleep, 0.25, true, helperpkg.FormatCompactDurationHours),
		})
	}
	if checkIn.ScreenTimeMinutes != nil {
		// The screen gauge is the limit left, so like the others a fuller arc is better.
		left := screenLimit - *checkIn.ScreenTimeMinutes
		value := helperpkg.FormatCompactDurationMinutes(max(0, left))
		if left < 0 {
			value = "+" + helperpkg.FormatCompactDurationMinutes(-left)
		}
		var avgLeft *float64
		if avgScreen != nil {
			v := float64(screenLimit) - *avgScreen
			avgLeft = &v
		}
		gauges = append(gauges, gauge{
			label: "Screen left", value: value,
			fraction: float64(left) / float64(screenLimit),
			color:    limitColor(theme, left, screenLimit),
			trend: trend(theme, float64(left), avgLeft, 15, true, func(v float64) string {
				if v < 0 {
					return "+" + helperpkg.FormatCompactDurationMinutes(int(-v))
				}
				return helperpkg.FormatCompactDurationMinutes(int(v))
			}),
		})
	}
	lines = append(lines, renderGauges(theme, gauges, width)...)
	if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
		lines = append(lines, "")
		label, quote := "Notes", "❝ "
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*checkIn.Notes), max(1, width-labelWidth-2), ""), "\n") {
			lines = append(lines, Row(theme, label, theme.StyleDim.Render(quote)+theme.StyleDim.Italic(true).Render(line), width))
			label, quote = "", "  "
		}
	}
	return lines
}

// DayLine is the day summary in one line for narrow screens, where the detail pane that
// holds it is hidden: focus, worked against target, rest, sessions and accountability. Parts
// drop from the right to fit; it is empty until the summary has loaded.
func DayLine(theme types.Theme, state types.ContentState, width int) string {
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate || snapshot.Focus == nil || snapshot.Metrics == nil {
		return ""
	}
	duration := helperpkg.FormatCompactDurationSeconds
	focus, metrics := snapshot.Focus, snapshot.Metrics
	part := func(label, value string) string {
		return theme.StyleDim.Render(label+" ") + value
	}
	level := strings.ReplaceAll(string(focus.Level), "_", " ")
	parts := []string{
		part("Focus", theme.StyleHeader.Render(fmt.Sprint(focus.Score))+" "+theme.StyleNormal.Render(level)),
		part("Worked", duration(metrics.WorkedSeconds)+theme.StyleDim.Render("/"+duration(focus.TargetWorkedSeconds))),
		part("Rest", duration(metrics.RestSeconds)),
		part("Sessions", fmt.Sprint(metrics.SessionCount)),
	}
	if plan := snapshot.Plan; plan != nil && strings.TrimSpace(plan.Date) != "" {
		parts = append(parts, part("Score", fmt.Sprintf("%.0f%%", plan.Summary.AccountabilityScore)))
	}
	sep := theme.StyleDim.Render("  ·  ")
	for len(parts) > 1 && ansi.StringWidth(" "+strings.Join(parts, sep)) > width {
		parts = parts[:len(parts)-1]
	}
	return viewhelpers.TruncateANSI(" "+strings.Join(parts, sep), width)
}

// plannedStatusCounts counts the shown day's planned tasks by status, for the status bar.
func plannedStatusCounts(state types.ContentState) map[string]int {
	counts := map[string]int{}
	for _, issue := range state.DailyPlannedIssues {
		counts[string(issue.Status)]++
	}
	return counts
}

// wellbeingTargets are the sleep target and screen-time limit from settings, in minutes,
// falling back to the defaults.
func wellbeingTargets(settings *sharedtypes.CoreSettings) (sleep, screen int) {
	sleep, screen = sharedtypes.DefaultSleepTargetMinutes, sharedtypes.DefaultScreenTimeLimitMinutes
	if settings != nil && settings.SleepTargetMinutes > 0 {
		sleep = settings.SleepTargetMinutes
	}
	if settings != nil && settings.ScreenTimeLimitMins > 0 {
		screen = settings.ScreenTimeLimitMins
	}
	return sleep, screen
}

// targetColor colours progress toward a target: green from the target, yellow within an
// eighth below it, red further off.
func targetColor(theme types.Theme, fraction float64) lipgloss.Color {
	switch {
	case fraction >= 1:
		return theme.ColorGreen
	case fraction >= 0.875:
		return theme.ColorYellow
	}
	return theme.ColorRed
}

// scaleColor colours a 1-5 check-in value: red low, yellow middling, green high.
func scaleColor(theme types.Theme, value int) lipgloss.Color {
	switch {
	case value <= 2:
		return theme.ColorRed
	case value == 3:
		return theme.ColorYellow
	}
	return theme.ColorGreen
}

// trend compares today with the 7-day average: an arrow for the direction, green when it
// moved the good way, red when the bad way, then the average, dim. Within tolerance it shows
// only the average.
func trend(theme types.Theme, today float64, average *float64, tolerance float64, higherIsBetter bool, format func(float64) string) string {
	if average == nil {
		return ""
	}
	avg := theme.StyleDim.Render("7d " + format(*average))
	diff := today - *average
	if diff > -tolerance && diff < tolerance {
		return theme.StyleDim.Render("=") + " " + avg
	}
	arrow, good := "▲", diff > 0
	if diff < 0 {
		arrow = "▼"
	}
	if !higherIsBetter {
		good = !good
	}
	color := theme.ColorRed
	if good {
		color = theme.ColorGreen
	}
	return lipgloss.NewStyle().Foreground(color).Render(arrow) + " " + avg
}

// limitColor colours what is left of a limit: green with a quarter or more left, yellow
// below that, red once over.
func limitColor(theme types.Theme, left, limit int) lipgloss.Color {
	switch {
	case left < 0:
		return theme.ColorRed
	case left*4 < limit:
		return theme.ColorYellow
	}
	return theme.ColorGreen
}
