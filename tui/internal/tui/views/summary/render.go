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
	lines = append(lines, renderPlan(theme, snapshot.Plan, width)...)
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
	reason := strings.ReplaceAll(string(focus.Reason), "_", " ")
	level := strings.ReplaceAll(string(focus.Level), "_", " ")
	if level != "" {
		level = strings.ToUpper(level[:1]) + level[1:]
	}
	duration := helperpkg.FormatCompactDurationSeconds
	lines = append(
		lines,
		Row(
			theme,
			"Focus",
			theme.StyleHeader.Render(
				fmt.Sprintf("%d/100", focus.Score),
			)+"  "+theme.StyleNormal.Render(
				level,
			),
			width,
		),
	)
	if reason != "" && !strings.EqualFold(reason, level) {
		lines = append(lines, Row(theme, "", theme.StyleDim.Render(reason), width))
	}
	lines = append(
		lines,
		Row(
			theme,
			"Worked",
			fmt.Sprintf(
				"%s of %s",
				duration(metrics.WorkedSeconds),
				duration(focus.TargetWorkedSeconds),
			),
			width,
		),
		strings.Repeat(
			" ",
			labelWidth,
		)+renderBar(
			theme,
			metrics.WorkedSeconds,
			0,
			max(1, focus.TargetWorkedSeconds),
			max(8, width-labelWidth),
			theme.ColorCyan,
		),
		Row(theme, "Rest", duration(metrics.RestSeconds), width),
		Row(theme, "Sessions", fmt.Sprint(metrics.SessionCount), width),
		Row(
			theme,
			"Issues",
			fmt.Sprintf(
				"%d of %d done · %d abandoned",
				metrics.CompletedIssues,
				metrics.TotalIssues,
				metrics.AbandonedIssues,
			),
			width,
		),
		strings.Repeat(" ", labelWidth)+IssueStatusBar(theme, statusCounts, max(8, width-labelWidth)),
		Row(
			theme,
			"Habits",
			fmt.Sprintf(
				"%d of %d done · %d failed",
				metrics.HabitCompletedCount,
				metrics.HabitDueCount,
				metrics.HabitFailedCount,
			),
			width,
		),
		strings.Repeat(" ", labelWidth)+HabitBar(
			theme,
			metrics.HabitCompletedCount,
			metrics.HabitFailedCount,
			metrics.HabitDueCount,
			max(8, width-labelWidth),
		),
	)
	return strings.Join(lines, "\n")
}

func renderPlan(theme types.Theme, plan *sharedtypes.DailyPlan, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Accountability")}
	if plan == nil || strings.TrimSpace(plan.Date) == "" {
		return append(lines, theme.StyleDim.Render("No committed plan."))
	}
	s := plan.Summary
	barWidth := max(8, width-labelWidth)
	scoreColor := theme.ColorGreen
	switch {
	case s.AccountabilityScore < 50:
		scoreColor = theme.ColorRed
	case s.AccountabilityScore < 80:
		scoreColor = theme.ColorYellow
	}
	score := lipgloss.NewStyle().Foreground(scoreColor).Bold(true).Render(fmt.Sprintf("%.0f%%", s.AccountabilityScore))
	return append(
		lines,
		Row(theme, "Score", score+theme.StyleDim.Render(fmt.Sprintf("  pressure %.1f", s.BacklogPressure)), width),
		strings.Repeat(" ", labelWidth)+renderBar(theme, int(s.AccountabilityScore), 0, 100, barWidth, scoreColor),
		Row(theme, "Done", fmt.Sprintf("%d of %d", s.CompletedCount, s.PlannedCount), width),
		strings.Repeat(" ", labelWidth)+planBar(theme, s, barWidth),
		Row(theme, "Missed", outcome(theme, "✗", s.FailedCount, "failed", theme.ColorRed)+
			theme.StyleDim.Render("  ·  ")+outcome(theme, "⊘", s.AbandonedCount, "abandoned", theme.ColorOrange), width),
		Row(theme, "Carried", outcome(theme, "↷", s.CarriedCount, "to a later day", theme.ColorLavender), width),
		Row(theme, "Pending", outcome(theme, "↺", s.PendingRollbackCount, "to roll back", theme.ColorYellow), width),
	)
}

// planBar splits the planned tasks into done, failed, abandoned and carried segments over a
// track for those still open.
func planBar(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary, width int) string {
	total := max(1, s.PlannedCount)
	used := 0
	var b strings.Builder
	for _, part := range []struct {
		count int
		color lipgloss.Color
	}{
		{s.CompletedCount, theme.ColorGreen},
		{s.FailedCount, theme.ColorRed},
		{s.AbandonedCount, theme.ColorOrange},
		{s.CarriedCount, theme.ColorLavender},
	} {
		cells := min(width-used, part.count*width/total)
		if part.count > 0 && cells == 0 && used < width {
			cells = 1
		}
		b.WriteString(lipgloss.NewStyle().Foreground(part.color).Render(strings.Repeat("█", cells)))
		used += cells
	}
	return b.String() + theme.StyleDim.Render(strings.Repeat("░", width-used))
}

// outcome is a glyph, count and label, coloured when the count is not zero and dim when it is.
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
		score := fmt.Sprintf("%.2f", entry.FailScore)
		if entry.Penalty.HighRisk {
			score = lipgloss.NewStyle().Foreground(theme.ColorOrange).Render("▲" + score)
		}
		share := ""
		if entry.Penalty.Contribution != nil {
			pct := entry.Penalty.Contribution.AccountabilityPct
			color := theme.ColorYellow
			if entry.Penalty.HighRisk {
				color = theme.ColorOrange
			}
			// A small bar shows the issue's share of the day's penalty.
			share = "  " + renderBar(theme, int(pct*100), 0, 100, 6, color) +
				theme.StyleDim.Render(fmt.Sprintf(" %3.0f%%", pct*100))
		}
		lines = append(lines, Row(theme, label, score+share+"  "+title, width))
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

func renderBar(theme types.Theme, current, failed, total, width int, color lipgloss.Color) string {
	if width < 1 {
		return ""
	}
	if total < 1 {
		return theme.StyleDim.Render(strings.Repeat("░", width))
	}
	filled := clamp(current*width/total, 0, width)
	failedWidth := clamp(failed*width/total, 0, width-filled)
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("█", failedWidth)) +
		theme.StyleDim.Render(strings.Repeat("░", width-filled-failedWidth))
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
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
