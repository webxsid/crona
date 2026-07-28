package summary

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
)

const wideSummaryWidth = 104

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, renderView).RenderView()
}

func renderView(theme types.Theme, state types.ContentState) string {
	innerWidth := max(24, state.Width-6)
	lines := renderHeader(theme, state, innerWidth)
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate {
		lines = append(lines, "", theme.StyleDim.Render("Loading summary…"))
		return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, strings.Join(lines, "\n"))
	}

	lines = append(lines, "", renderKPIBand(theme, state, innerWidth))
	if state.Width >= wideSummaryWidth {
		leftWidth, rightWidth := viewhelpers.SplitHorizontal(innerWidth, 42, 32, innerWidth*3/5)
		left := strings.Join(renderAgenda(theme, state, leftWidth), "\n")
		right := strings.Join(renderSignals(theme, state, rightWidth), "\n")
		lines = append(lines, "", lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(left),
			lipgloss.NewStyle().Width(rightWidth).Render(right),
		))
	} else {
		lines = append(lines, "")
		lines = append(lines, renderAgenda(theme, state, innerWidth)...)
		lines = append(lines, "")
		lines = append(lines, renderSignals(theme, state, innerWidth)...)
	}
	content := scrollSummaryLines(
		lines,
		state.Cursors["summary"],
		max(1, state.Height-2),
		theme.StyleDim.Render("↓ more"),
	)
	return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, strings.Join(content, "\n"))
}

func PaneLineCount(state types.ContentState) int {
	if state.SummarySnapshot == nil {
		return 1
	}
	return 18 + len(state.SummarySnapshot.Habits) + issueCount(state.SummarySnapshot.Issues)
}

func scrollSummaryLines(lines []string, offset, height int, moreLabel string) []string {
	if len(lines) <= height {
		return lines
	}
	maxOffset := max(0, len(lines)-height+1)
	offset = clamp(offset, 0, maxOffset)
	end := min(len(lines), offset+height-1)
	out := append([]string(nil), lines[offset:end]...)
	if end < len(lines) {
		out = append(out, moreLabel)
	}
	return out
}

func renderHeader(theme types.Theme, state types.ContentState, width int) []string {
	date := helperpkg.FormatDisplayDate(state.SummaryDate, state.Settings)
	title := fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Summary"), theme.StyleHeader.Render(date))
	actions := viewchrome.RenderPaneActionLine(
		theme,
		viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane}),
		width,
	)
	scope := contextText(state)
	status := dateState(state.SummaryDate)
	meta := theme.StyleDim.Render(status + "  ·  " + scope)
	lines := []string{viewhelpers.TruncateANSI(title, width)}
	if strings.TrimSpace(actions) != "" {
		lines = append(lines, actions)
	}
	lines = append(lines, viewhelpers.TruncateANSI(meta, width))
	if status == "Today" {
		if timer := summaryTimerLine(theme, state, width); timer != "" {
			lines = append(lines, timer)
		}
	}
	return lines
}

func renderKPIBand(theme types.Theme, state types.ContentState, width int) string {
	snapshot := state.SummarySnapshot
	issueTotal, resolved, worked, estimated := 0, 0, 0, 0
	if snapshot.Issues != nil {
		issueTotal = snapshot.Issues.TotalIssues
		resolved = snapshot.Issues.CompletedIssues + snapshot.Issues.AbandonedIssues
		worked = snapshot.Issues.WorkedSeconds
		estimated = snapshot.Issues.TotalEstimatedMinutes * 60
	}
	habitDone, habitFailed := habitCounts(snapshot.Habits)
	accountability := 0.0
	if snapshot.Plan != nil {
		accountability = snapshot.Plan.Summary.AccountabilityScore
	}
	cards := []kpi{
		{label: "Focus", value: helperpkg.FormatCompactDurationSeconds(worked), current: worked, total: estimated, color: theme.ColorCyan},
		{label: "Issues", value: fmt.Sprintf("%d/%d resolved", resolved, issueTotal), current: resolved, total: issueTotal, color: theme.ColorGreen},
		{label: "Habits", value: fmt.Sprintf("%d/%d done", habitDone, len(snapshot.Habits)), current: habitDone, total: len(snapshot.Habits), failed: habitFailed, color: theme.ColorGreen},
		{label: "Plan", value: fmt.Sprintf("%.0f%% accountable", accountability), current: int(accountability + .5), total: 100, color: theme.ColorMagenta},
	}
	if width < 72 {
		out := make([]string, 0, len(cards)*2+1)
		out = append(out, theme.StyleHeader.Render("At a glance"))
		for _, card := range cards {
			out = append(out, renderKPILine(theme, card, width))
		}
		return strings.Join(out, "\n")
	}
	cellWidth := max(16, width/len(cards))
	rendered := make([]string, 0, len(cards))
	for _, card := range cards {
		rendered = append(rendered, renderKPICard(theme, card, cellWidth))
	}
	return theme.StyleHeader.Render("At a glance") + "\n" +
		lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

type kpi struct {
	label   string
	value   string
	current int
	total   int
	failed  int
	color   lipgloss.Color
}

func renderKPICard(theme types.Theme, card kpi, width int) string {
	inner := max(8, width-3)
	body := strings.Join([]string{
		theme.StyleHeader.Render(card.label),
		theme.StyleNormal.Render(card.value),
		renderBar(theme, card.current, card.failed, card.total, inner, card.color),
	}, "\n")
	return lipgloss.NewStyle().Width(width).PaddingRight(1).Render(body)
}

func renderKPILine(theme types.Theme, card kpi, width int) string {
	label := fmt.Sprintf("%-7s %-18s", card.label, card.value)
	barWidth := max(6, width-lipgloss.Width(label)-2)
	return viewhelpers.TruncateANSI(
		theme.StyleHeader.Render(fmt.Sprintf("%-7s", card.label))+" "+
			theme.StyleNormal.Render(fmt.Sprintf("%-18s", card.value))+" "+
			renderBar(theme, card.current, card.failed, card.total, barWidth, card.color),
		width,
	)
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

func renderAgenda(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	lines := []string{theme.StyleHeader.Render("Agenda")}
	issues := []sharedtypes.Issue(nil)
	if snapshot.Issues != nil {
		issues = snapshot.Issues.Issues
	}
	lines = append(lines, theme.StylePaneTitle.Render(fmt.Sprintf("Issues  %d", len(issues))))
	if len(issues) == 0 {
		lines = append(lines, theme.StyleDim.Render("No issues scheduled."))
	} else {
		for _, issue := range issues {
			status, style := issueStatus(theme, issue.Status)
			meta := issueMeta(issue)
			prefix := style.Render(status)
			available := max(8, width-lipgloss.Width(prefix)-lipgloss.Width(meta)-4)
			line := prefix + " " + theme.StyleNormal.Render(viewhelpers.Truncate(issue.Title, available))
			if meta != "" {
				line += "  " + theme.StyleDim.Render(meta)
			}
			lines = append(lines, viewhelpers.TruncateANSI(line, width))
		}
	}
	lines = append(lines, "", theme.StylePaneTitle.Render(fmt.Sprintf("Habits  %d", len(snapshot.Habits))))
	if len(snapshot.Habits) == 0 {
		lines = append(lines, theme.StyleDim.Render("No habits due."))
	} else {
		for _, habit := range snapshot.Habits {
			status, style := habitStatus(theme, habit)
			meta := habitMeta(habit)
			prefix := style.Render(status)
			available := max(8, width-lipgloss.Width(prefix)-lipgloss.Width(meta)-4)
			line := prefix + " " + theme.StyleNormal.Render(viewhelpers.Truncate(habit.Name, available))
			if meta != "" {
				line += "  " + theme.StyleDim.Render(meta)
			}
			lines = append(lines, viewhelpers.TruncateANSI(line, width))
		}
	}
	return lines
}

func renderSignals(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	lines := []string{theme.StyleHeader.Render("Signals")}
	lines = append(lines, renderPlan(theme, snapshot.Plan, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWellbeing(theme, snapshot.CheckIn, snapshot.Rollup, width)...)
	lines = append(lines, "")
	lines = append(lines, renderStreaks(theme, snapshot.Streaks, width)...)
	return lines
}

func renderPlan(theme types.Theme, plan *sharedtypes.DailyPlan, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Accountability")}
	if plan == nil || strings.TrimSpace(plan.Date) == "" {
		return append(lines, theme.StyleDim.Render("No committed plan."))
	}
	s := plan.Summary
	lines = append(lines,
		fmt.Sprintf("%s  %s", theme.StyleNormal.Render(fmt.Sprintf("%.0f%% score", s.AccountabilityScore)), theme.StyleDim.Render(fmt.Sprintf("pressure %.1f", s.BacklogPressure))),
		renderBar(theme, s.CompletedCount, s.FailedCount+s.AbandonedCount, max(1, s.PlannedCount), max(8, width-2), theme.ColorGreen),
		theme.StyleDim.Render(fmt.Sprintf("done %d  pending %d  failed %d  abandoned %d", s.CompletedCount, s.PendingRollbackCount, s.FailedCount, s.AbandonedCount)),
	)
	return lines
}

func renderWellbeing(theme types.Theme, checkIn *sharedtypes.DailyCheckIn, rollup *sharedtypes.MetricsRollup, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Wellbeing")}
	if checkIn == nil || strings.TrimSpace(checkIn.Date) == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded."))
		if rollup != nil && (rollup.AverageMood != nil || rollup.AverageEnergy != nil) {
			lines = append(lines, theme.StyleDim.Render(averageSignals(rollup)))
		}
		return lines
	}
	signals := []string{fmt.Sprintf("mood %d/5", checkIn.Mood), fmt.Sprintf("energy %d/5", checkIn.Energy)}
	if checkIn.SleepHours != nil {
		signals = append(signals, "sleep "+helperpkg.FormatCompactDurationHours(*checkIn.SleepHours))
	}
	if checkIn.ScreenTimeMinutes != nil {
		signals = append(signals, "screen "+helperpkg.FormatCompactDurationMinutes(*checkIn.ScreenTimeMinutes))
	}
	lines = append(lines, viewhelpers.TruncateANSI(theme.StyleNormal.Render(strings.Join(signals, "  ·  ")), width))
	if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
		lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(strings.TrimSpace(*checkIn.Notes), width)))
	}
	return lines
}

func renderStreaks(theme types.Theme, streaks *sharedtypes.StreakSummary, width int) []string {
	lines := []string{theme.StylePaneTitle.Render("Momentum")}
	if streaks == nil {
		return append(lines, theme.StyleDim.Render("No streak data."))
	}
	text := fmt.Sprintf("focus %d  ·  check-ins %d  ·  habits %d", streaks.CurrentFocusDays, streaks.CurrentCheckInDays, streaks.CurrentHabitDays)
	return append(lines, viewhelpers.TruncateANSI(theme.StyleNormal.Render(text), width))
}

func contextText(state types.ContentState) string {
	if state.Context == nil {
		return "Workspace"
	}
	parts := make([]string, 0, 3)
	for _, value := range []*string{state.Context.RepoName, state.Context.StreamName, state.Context.IssueTitle} {
		if value != nil && strings.TrimSpace(*value) != "" {
			parts = append(parts, strings.TrimSpace(*value))
		}
	}
	if len(parts) == 0 {
		return "Workspace"
	}
	return strings.Join(parts, " › ")
}

func summaryTimerLine(theme types.Theme, state types.ContentState, width int) string {
	if state.Timer == nil || state.Timer.State == "idle" || state.Timer.State == "ready" {
		return ""
	}
	segment := "session"
	if state.Timer.SegmentType != nil {
		segment = string(*state.Timer.SegmentType)
	}
	text := fmt.Sprintf("%s  %s  %s", strings.ToUpper(state.Timer.State), segment, helperpkg.FormatCompactDurationSeconds(state.Elapsed))
	return viewhelpers.TruncateANSI(theme.StyleHeader.Render(text), width)
}

func issueStatus(theme types.Theme, status sharedtypes.IssueStatus) (string, lipgloss.Style) {
	switch status {
	case sharedtypes.IssueStatusDone:
		return "✓", lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case sharedtypes.IssueStatusAbandoned:
		return "×", lipgloss.NewStyle().Foreground(theme.ColorRed)
	case sharedtypes.IssueStatusBlocked:
		return "!", lipgloss.NewStyle().Foreground(theme.ColorRed)
	case sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusInReview:
		return "●", lipgloss.NewStyle().Foreground(theme.ColorCyan)
	default:
		return "○", lipgloss.NewStyle().Foreground(theme.ColorYellow)
	}
}

func habitStatus(theme types.Theme, habit sharedtypes.HabitDailyItem) (string, lipgloss.Style) {
	if habit.Status == sharedtypes.HabitCompletionStatusFailed {
		return "×", lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
	if habit.Completed {
		return "✓", lipgloss.NewStyle().Foreground(theme.ColorGreen)
	}
	return "○", lipgloss.NewStyle().Foreground(theme.ColorYellow)
}

func issueMeta(issue sharedtypes.Issue) string {
	parts := make([]string, 0, 2)
	if issue.EstimateMinutes != nil {
		parts = append(parts, helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes))
	}
	if issue.WorkedSeconds > 0 {
		parts = append(parts, helperpkg.FormatCompactDurationSeconds(issue.WorkedSeconds)+" worked")
	}
	return strings.Join(parts, " · ")
}

func habitMeta(habit sharedtypes.HabitDailyItem) string {
	target := habit.TargetMinutes
	if habit.SnapshotTarget != nil {
		target = habit.SnapshotTarget
	}
	switch {
	case habit.DurationMinutes != nil && target != nil:
		return fmt.Sprintf("%s/%s", helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes), helperpkg.FormatCompactDurationMinutes(*target))
	case habit.DurationMinutes != nil:
		return helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes)
	case target != nil:
		return "target " + helperpkg.FormatCompactDurationMinutes(*target)
	default:
		return ""
	}
}

func habitCounts(habits []sharedtypes.HabitDailyItem) (completed, failed int) {
	for _, habit := range habits {
		switch {
		case habit.Status == sharedtypes.HabitCompletionStatusFailed:
			failed++
		case habit.Completed:
			completed++
		}
	}
	return completed, failed
}

func issueCount(summary *sharedtypes.DailyIssueSummary) int {
	if summary == nil {
		return 0
	}
	return len(summary.Issues)
}

func averageSignals(rollup *sharedtypes.MetricsRollup) string {
	parts := make([]string, 0, 3)
	if rollup.AverageMood != nil {
		parts = append(parts, fmt.Sprintf("7d mood %.1f/5", *rollup.AverageMood))
	}
	if rollup.AverageEnergy != nil {
		parts = append(parts, fmt.Sprintf("energy %.1f/5", *rollup.AverageEnergy))
	}
	if rollup.AverageSleepHours != nil {
		parts = append(parts, "sleep "+helperpkg.FormatCompactDurationHours(*rollup.AverageSleepHours))
	}
	return strings.Join(parts, "  ·  ")
}

func dateState(date string) string {
	today := time.Now().Format("2006-01-02")
	switch {
	case date == today:
		return "Today"
	case date < today:
		return "Historical"
	default:
		return "Upcoming"
	}
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
