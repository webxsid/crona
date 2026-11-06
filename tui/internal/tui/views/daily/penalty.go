package daily

import (
	"fmt"
	"strings"

	"crona.work/tui/internal/api"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

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

// penaltyLines explains the issue's delay penalty in plain terms: how late it is, a bar of the
// penalty against the high-risk limit, its share of today's penalty, and what finishing it
// today or pushing it again would do. The kernel does all the scoring; this only lays it out.
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
	good := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	bad := lipgloss.NewStyle().Foreground(theme.ColorRed)
	label := func(text string) string { return theme.StyleDim.Render(padCells(text, 14)) }
	num := func(value float64) string { return fmt.Sprintf("%.2f", value) }

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

	lateParts := []string{plural(entry.CurrentDelayedDays, "day")}
	if entry.PostponeCount > 1 {
		lateParts = append(lateParts, fmt.Sprintf("pushed %d×", entry.PostponeCount))
	}
	if penalty.ProtectedDaysSkipped > 0 {
		lateParts = append(lateParts, plural(penalty.ProtectedDaysSkipped, "away day")+" skipped")
	}
	lines := []string{
		"", header,
		label("Late") + body.Render(strings.Join(lateParts, " · ")),
		label("Penalty") + penaltyBar(theme, entry.FailScore, penalty.HighRiskThreshold, penalty.HighRisk && !penalty.Historical) +
			" " + body.Render(num(entry.FailScore)) + theme.StyleDim.Render("  limit "+num(penalty.HighRiskThreshold)),
	}
	if penalty.Historical {
		return truncateLines(lines, width)
	}
	if share := penalty.Contribution; share != nil {
		lines = append(lines, label("Share")+body.Render(
			fmt.Sprintf("%.0f%% of today's penalty", share.AccountabilityPct*100),
		))
	}
	if done := penalty.IfDoneToday; done != nil && done.Accountability < 0 {
		lines = append(lines, label("Finish today")+good.Render("clears "+num(-done.Accountability)))
	}
	if push := penalty.IfPostponedOneDay; push != nil {
		text := bad.Render(fmt.Sprintf("rises to %s (+%s)", num(push.FailScore), num(push.AccountabilityDelta)))
		if push.BecomesHighRisk {
			text += "  " + warn.Render("▲ high risk")
		}
		lines = append(lines, label("Push a day")+text)
	}
	return truncateLines(lines, width)
}

// penaltyView is the issue's delay penalty laid out for its details: an arc while it is live,
// a verdict for the fact column, and the rows under the arcs.
type penaltyView struct {
	arc     *viewsummary.Arc
	verdict string
	lines   []string
}

// issuePenaltyView splits a live penalty into an arc of the score against twice the
// high-risk limit, a verdict (risk and share of today), and the delay and what-if rows. A
// settled penalty keeps its plain rows.
func issuePenaltyView(theme types.Theme, state types.ContentState, issueID int64, width int) penaltyView {
	entry := issuePlanEntry(state, issueID)
	if entry == nil || entry.Penalty == nil {
		return penaltyView{}
	}
	penalty := entry.Penalty
	late := entry.CurrentDelayedDays > 0 || entry.PostponeCount > 1
	if !late {
		if penalty.Historical {
			return penaltyView{}
		}
		return penaltyView{verdict: theme.StyleDim.Render("on schedule · no penalty")}
	}
	if penalty.Historical {
		return penaltyView{lines: penaltyLines(theme, state, issueID, width)}
	}
	warn := lipgloss.NewStyle().Foreground(theme.ColorOrange)
	num := func(value float64) string { return fmt.Sprintf("%.2f", value) }
	limit := penalty.HighRiskThreshold
	color := theme.ColorYellow
	verdict := lipgloss.NewStyle().Foreground(theme.ColorYellow).Render("◆ late")
	if penalty.HighRisk {
		color = theme.ColorOrange
		verdict = warn.Render("▲ high risk")
	}
	if share := penalty.Contribution; share != nil {
		verdict += theme.StyleDim.Render(fmt.Sprintf(" · %.0f%% of today", share.AccountabilityPct*100))
	}
	label := "Penalty"
	if limit > 0 {
		label += "/" + strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", limit), "0"), ".")
	}
	arc := &viewsummary.Arc{
		Label: label, Value: num(entry.FailScore),
		Fraction: entry.FailScore / max(limit*2, 0.01), Color: color,
	}
	row := func(name, value string) string { return theme.StyleDim.Render(padCells(name, 11)) + value }
	lateParts := []string{plural(entry.CurrentDelayedDays, "day") + " late"}
	if entry.PostponeCount > 1 {
		lateParts = append(lateParts, fmt.Sprintf("pushed %d×", entry.PostponeCount))
	}
	if penalty.ProtectedDaysSkipped > 0 {
		lateParts = append(lateParts, plural(penalty.ProtectedDaysSkipped, "away day")+" skipped")
	}
	lines := []string{"", row("Delay", theme.StyleNormal.Render(strings.Join(lateParts, " · ")))}
	if done := penalty.IfDoneToday; done != nil && done.Accountability < 0 {
		lines = append(lines, row("If today", lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("✓ clears "+num(-done.Accountability))))
	}
	if push := penalty.IfPostponedOneDay; push != nil {
		text := lipgloss.NewStyle().Foreground(theme.ColorRed).Render(
			fmt.Sprintf("▲ rises to %s (+%s)", num(push.FailScore), num(push.AccountabilityDelta)))
		if push.BecomesHighRisk {
			text += "  " + warn.Render("turns high risk")
		}
		lines = append(lines, row("If pushed", text))
	}
	return penaltyView{arc: arc, verdict: verdict, lines: truncateLines(lines, width)}
}

// penaltyBarWidth is the penalty bar's cells.
const penaltyBarWidth = 16

// penaltyBar fills to score on a scale of twice the high-risk limit (or the score, if
// higher), with a │ at the limit; it is orange when high risk, yellow otherwise.
func penaltyBar(theme types.Theme, score, limit float64, highRisk bool) string {
	scale := max(limit*2, score, 0.01)
	filled := min(penaltyBarWidth, int(score/scale*penaltyBarWidth+0.5))
	mark := -1
	if limit > 0 {
		mark = min(penaltyBarWidth-1, int(limit/scale*penaltyBarWidth+0.5))
	}
	fill := lipgloss.NewStyle().Foreground(theme.ColorYellow)
	if highRisk {
		fill = lipgloss.NewStyle().Foreground(theme.ColorOrange)
	}
	var out strings.Builder
	for i := range penaltyBarWidth {
		switch {
		case i == mark:
			out.WriteString(theme.StyleNormal.Render("│"))
		case i < filled:
			out.WriteString(fill.Render("█"))
		default:
			out.WriteString(theme.StyleDim.Render("░"))
		}
	}
	return out.String()
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
		if entry.IssueID != issueID || entry.Penalty == nil ||
			entry.Penalty.IfPostponedOneDay == nil {
			continue
		}
		push := entry.Penalty.IfPostponedOneDay
		hint := fmt.Sprintf(
			"Push a day: penalty rises to %.2f (%+.2f)",
			push.FailScore,
			push.AccountabilityDelta,
		)
		if push.BecomesHighRisk {
			hint += " ▲ high risk"
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
