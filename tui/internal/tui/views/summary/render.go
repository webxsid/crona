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
	lines := strings.Split(renderStats(theme, snapshot, width), "\n")
	lines = append(lines, "")
	lines = append(lines, renderPlan(theme, snapshot.Plan, width)...)
	lines = append(lines, renderTopPenalties(theme, snapshot.Plan, snapshot.Issues, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWellbeing(theme, snapshot.CheckIn, snapshot.Rollup, width)...)
	lines = append(lines, "")
	return append(lines, renderStreaks(theme, snapshot.Streaks, width)...)
}

// labelWidth is the label column every summary row lines its value up after.
const labelWidth = 10

// Row is one stat: a dim label column, then its value cut to the room left.
func Row(theme types.Theme, label, value string, width int) string {
	return theme.StyleDim.Render(fmt.Sprintf("%-*s", labelWidth, label)) +
		viewhelpers.TruncateANSI(value, max(1, width-labelWidth))
}

func renderStats(theme types.Theme, snapshot *api.SummarySnapshot, width int) string {
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
	)
	return strings.Join(lines, "\n")
}

func renderPlan(theme types.Theme, plan *sharedtypes.DailyPlan, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Accountability")}
	if plan == nil || strings.TrimSpace(plan.Date) == "" {
		return append(lines, theme.StyleDim.Render("No committed plan."))
	}
	s := plan.Summary
	return append(
		lines,
		Row(
			theme,
			"Score",
			fmt.Sprintf(
				"%.0f%%",
				s.AccountabilityScore,
			)+theme.StyleDim.Render(
				fmt.Sprintf("  pressure %.1f", s.BacklogPressure),
			),
			width,
		),
		Row(theme, "Done", fmt.Sprintf("%d of %d", s.CompletedCount, s.PlannedCount), width),
		strings.Repeat(
			" ",
			labelWidth,
		)+renderBar(
			theme,
			s.CompletedCount,
			s.FailedCount+s.AbandonedCount,
			max(1, s.PlannedCount),
			max(8, width-labelWidth),
			theme.ColorGreen,
		),
		Row(
			theme,
			"Missed",
			fmt.Sprintf("%d failed · %d abandoned", s.FailedCount, s.AbandonedCount),
			width,
		),
		Row(theme, "Pending", fmt.Sprintf("%d to roll back", s.PendingRollbackCount), width),
	)
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
			share = theme.StyleDim.Render(
				fmt.Sprintf(" %.0f%%", entry.Penalty.Contribution.AccountabilityPct*100),
			)
		}
		lines = append(lines, Row(theme, label, score+share+"  "+title, width))
	}
	return lines
}

func renderWellbeing(
	theme types.Theme,
	checkIn *sharedtypes.DailyCheckIn,
	rollup *sharedtypes.MetricsRollup,
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
	lines = append(lines,
		Row(theme, "Mood", fmt.Sprintf("%d/5", checkIn.Mood), width),
		Row(theme, "Energy", fmt.Sprintf("%d/5", checkIn.Energy), width),
	)
	if checkIn.SleepHours != nil {
		lines = append(
			lines,
			Row(theme, "Sleep", helperpkg.FormatCompactDurationHours(*checkIn.SleepHours), width),
		)
	}
	if checkIn.ScreenTimeMinutes != nil {
		lines = append(
			lines,
			Row(
				theme,
				"Screen",
				helperpkg.FormatCompactDurationMinutes(*checkIn.ScreenTimeMinutes),
				width,
			),
		)
	}
	if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
		label := "Notes"
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*checkIn.Notes), max(1, width-labelWidth), ""), "\n") {
			lines = append(lines, Row(theme, label, theme.StyleDim.Render(line), width))
			label = ""
		}
	}
	return lines
}

func renderStreaks(theme types.Theme, streaks *sharedtypes.StreakSummary, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Momentum")}
	if streaks == nil {
		return append(lines, theme.StyleDim.Render("No streak data."))
	}
	return append(lines,
		Row(theme, "Focus", days(streaks.CurrentFocusDays), width),
		Row(theme, "Check-ins", days(streaks.CurrentCheckInDays), width),
		Row(theme, "Habits", days(streaks.CurrentHabitDays), width),
	)
}

func days(n int) string {
	if n == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", n)
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
