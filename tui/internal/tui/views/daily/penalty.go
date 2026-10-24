package daily

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// issuePlanEntry is the shown day's plan entry for an issue, if it has one.
func issuePlanEntry(state types.ContentState, issueID int64) *api.DailyPlanEntry {
	if state.DailyPlan == nil {
		return nil
	}
	for i := range state.DailyPlan.Entries {
		if state.DailyPlan.Entries[i].IssueID == issueID {
			return &state.DailyPlan.Entries[i]
		}
	}
	return nil
}

// penaltyLines explains the issue's delay penalty: how late it is, how the score is made up,
// its share of today's totals and what doing or postponing it would change. The kernel does
// all the scoring; this only lays it out.
func penaltyLines(theme types.Theme, state types.ContentState, issueID int64, width int) []string {
	entry := issuePlanEntry(state, issueID)
	if entry == nil || entry.Penalty == nil {
		return nil
	}
	penalty := entry.Penalty
	late := entry.CurrentDelayedDays > 0 || entry.PostponeCount > 1
	if !late {
		if penalty.Historical {
			return nil
		}
		return []string{"", theme.StyleDim.Render("On schedule · no penalty")}
	}
	warn := lipgloss.NewStyle().Foreground(theme.ColorOrange)
	label := func(text string) string { return theme.StyleDim.Render(padCells(text, 12)) }
	num := func(value float64) string { return fmt.Sprintf("%.2f", value) }
	signed := func(value float64) string { return fmt.Sprintf("%+.2f", value) }
	pct := func(value float64) string { return fmt.Sprintf("%.0f%%", value*100) }

	title := "Delay penalty"
	if penalty.Historical {
		title = "Delay penalty (final)"
	}
	header := theme.StyleDim.Bold(true).Render(title)
	if penalty.HighRisk && !penalty.Historical {
		header += "  " + warn.Render("▲ high risk")
	}
	body := theme.StyleNormal
	if penalty.Historical {
		body = theme.StyleDim
	}

	planned := fmt.Sprintf(
		"%s → %s  (%s late, %s",
		helperpkg.FormatDisplayDate(entry.BaselineDate, state.Settings),
		helperpkg.FormatDisplayDate(entry.CurrentPlannedDate, state.Settings),
		plural(entry.CurrentDelayedDays, "day"),
		plural(entry.PostponeCount, "postpone"),
	)
	if penalty.ProtectedDaysSkipped > 0 {
		planned += ", " + plural(penalty.ProtectedDaysSkipped, "away day") + " skipped"
	}
	planned += ")"
	score := fmt.Sprintf(
		"%s = %s delay + %s frequency",
		num(entry.FailScore), num(penalty.DelayComponent), num(penalty.FrequencyComponent),
	)
	scoreLine := label("Penalty") + body.Render(score) +
		theme.StyleDim.Render("  threshold "+num(penalty.HighRiskThreshold))

	lines := []string{"", header, label("Planned") + body.Render(planned), scoreLine}
	if penalty.Historical {
		return truncateLines(lines, width)
	}
	if share := penalty.Contribution; share != nil && state.DailyPlan != nil {
		summary := state.DailyPlan.Summary
		lines = append(lines,
			label("Today")+body.Render(fmt.Sprintf(
				"%s accountability (%s of %s)",
				signed(share.Accountability), pct(share.AccountabilityPct), num(summary.AccountabilityScore),
			)),
			label("")+body.Render(fmt.Sprintf(
				"%s backlog pressure (%s of %s)",
				signed(share.BacklogPressure), pct(share.BacklogPressurePct), num(summary.BacklogPressure),
			)),
		)
	}
	if done := penalty.IfDoneToday; done != nil {
		lines = append(lines, label("Done today")+body.Render(fmt.Sprintf(
			"accountability %s · backlog %s", signed(done.Accountability), signed(done.BacklogPressure),
		)))
	}
	if push := penalty.IfPostponedOneDay; push != nil {
		text := body.Render(fmt.Sprintf("penalty %s (%s)", num(push.FailScore), signed(push.AccountabilityDelta)))
		if push.BecomesHighRisk {
			text += "  " + warn.Render("becomes high risk")
		}
		lines = append(lines, label("Push 1 day")+text)
	}
	return truncateLines(lines, width)
}

// PenaltyMarkerText is the compact high-risk marker for issue rows, unstyled so it takes the
// row's style, or "" when the issue's shown-day entry is not high risk.
func PenaltyMarkerText(state types.ContentState, issueID int64) string {
	entry := issuePlanEntry(state, issueID)
	if entry == nil || entry.Penalty == nil || entry.Penalty.Historical || !entry.Penalty.HighRisk {
		return ""
	}
	return fmt.Sprintf("▲%.1f", entry.FailScore)
}

// RescheduleHint is the cost of pushing an issue one more day, shown when rescheduling it,
// or "" when the issue has no live penalty projection in the plan.
func RescheduleHint(plan *api.DailyPlan, issueID int64) string {
	if plan == nil {
		return ""
	}
	for _, entry := range plan.Entries {
		if entry.IssueID != issueID || entry.Penalty == nil || entry.Penalty.IfPostponedOneDay == nil {
			continue
		}
		push := entry.Penalty.IfPostponedOneDay
		hint := fmt.Sprintf("Push 1 day: penalty %.2f (%+.2f)", push.FailScore, push.AccountabilityDelta)
		if push.BecomesHighRisk {
			hint += " · becomes high risk"
		}
		return hint
	}
	return ""
}

func plural(count int, unit string) string {
	if count == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", count, unit)
}

func truncateLines(lines []string, width int) []string {
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = viewhelpers.TruncateANSI(line, width)
		}
	}
	return lines
}
