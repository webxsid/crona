package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderDailyView(theme Theme, state ContentState) string {
	summaryH, listH := splitVertical(state.Height, 10, 8, state.Height/3)
	sections := []string{}
	if summaryH >= 3 {
		sections = append(sections, renderDailySummary(theme, state, state.Width, summaryH))
	}
	if state.Width < 56 {
		issuesH, habitsH := splitVertical(listH, 8, 8, listH/2)
		sections = append(sections,
			renderDailyIssues(theme, state, state.Width, issuesH),
			renderDailyHabits(theme, state, state.Width, habitsH),
		)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	leftW, rightW := splitHorizontal(state.Width, 24, 24, state.Width*3/5)
	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		renderDailyIssues(theme, state, leftW, listH),
		renderDailyHabits(theme, state, rightW, listH),
	)
	sections = append(sections, lists)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderDailySummary(theme Theme, state ContentState, width, height int) string {
	dateText := currentDashboardDate(state)
	totalIssues, totalEstimate, completedCount, abandonedCount := 0, 0, 0, 0
	totalHabits, completedHabits, failedHabits, habitMinutes := len(state.DueHabits), 0, 0, 0
	habitTargetMinutes := 0
	issueStatusCounts := map[string]int{}
	for _, habit := range state.DueHabits {
		switch habit.Status {
		case "completed":
			completedHabits++
		case "failed":
			failedHabits++
		}
		if habit.DurationMinutes != nil {
			habitMinutes += *habit.DurationMinutes
		}
		if habit.TargetMinutes != nil {
			habitTargetMinutes += *habit.TargetMinutes
		}
	}
	for _, issue := range state.DailyIssues {
		totalIssues++
		issueStatusCounts[string(issue.Status)]++
		if issue.EstimateMinutes != nil {
			totalEstimate += *issue.EstimateMinutes
		}
		switch issue.Status {
		case "done":
			completedCount++
		case "abandoned":
			abandonedCount++
		}
	}
	if state.DailySummary != nil {
		dateText = state.DailySummary.Date
	}
	resolvedCount := completedCount + abandonedCount
	scopeText := defaultScopeLabel(state.Context)
	summaryInnerW := max(24, width-8)
	issueBarWidth := max(16, summaryInnerW-4)
	issueSummary := fmt.Sprintf(
		"%s  %s  %s",
		theme.StyleHeader.Render("Issues"),
		theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
		theme.StyleDim.Render(fmt.Sprintf("estimate %dm", totalEstimate)),
	)
	habitSummary := fmt.Sprintf(
		"%s  %s",
		theme.StyleHeader.Render("Habits"),
		theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
	)
	habitMeta := theme.StyleDim.Render(fmt.Sprintf("logged %dm", habitMinutes))
	if habitTargetMinutes > 0 {
		habitMeta = theme.StyleDim.Render(fmt.Sprintf("logged %dm / target %dm", habitMinutes, habitTargetMinutes))
	}
	habitVisual := []string{
		habitSummary,
		renderDailyHabitBar(theme, completedHabits, failedHabits, totalHabits, issueBarWidth),
		habitMeta,
		theme.StyleDim.Render(fmt.Sprintf("failed %d   remaining %d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
	}
	lines := []string{
		theme.StylePaneTitle.Render("Daily Dashboard"),
		theme.StyleHeader.Render(fmt.Sprintf("For %s", dateText)),
		theme.StyleDim.Render(scopeText),
		theme.StyleDim.Render("[,] prev   [.] next   [g] today"),
		"",
	}
	switch {
	case state.Height < 37:
		lines = []string{
			renderCompactMetadataRow(summaryInnerW,
				theme.StylePaneTitle.Render("Daily Dashboard"),
				theme.StyleHeader.Render(dateText),
			),
			renderCompactMetadataRow(summaryInnerW,
				theme.StyleDim.Render(scopeText),
				theme.StyleDim.Render("[,] [.] [g]"),
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", resolvedCount, totalIssues)),
					theme.StyleDim.Render(fmt.Sprintf("%dm", totalEstimate)),
					theme.StyleDim.Render(compactIssueLegend(issueStatusCounts)),
				},
				func(barWidth int) string { return renderDailyIssueStatusBar(theme, issueStatusCounts, barWidth) },
				tinySummaryBarWidth,
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d", completedHabits, totalHabits)),
					theme.StyleDim.Render(compactHabitProgress(habitMinutes, habitTargetMinutes)),
					theme.StyleDim.Render(fmt.Sprintf("f%d r%d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
				},
				func(barWidth int) string {
					return renderDailyHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				tinySummaryBarWidth,
			),
		}
	case state.Height < 48:
		lines = append(lines,
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
					theme.StyleDim.Render(fmt.Sprintf("estimate %dm", totalEstimate)),
				},
				func(barWidth int) string { return renderDailyIssueStatusBar(theme, issueStatusCounts, barWidth) },
				compactSummaryBarWidth,
			),
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string {
					return renderDailyHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				compactSummaryBarWidth,
			),
		)
	case state.Height < 55:
		lines = append(lines,
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Issues"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d resolved", resolvedCount, totalIssues)),
					theme.StyleDim.Render(fmt.Sprintf("estimate %dm", totalEstimate)),
				},
				func(barWidth int) string { return renderDailyIssueStatusBar(theme, issueStatusCounts, barWidth) },
				compactSummaryBarWidth,
			),
			theme.StyleDim.Render(renderDailyIssueLegend(issueStatusCounts)),
			"",
			renderCompactSummaryRow(summaryInnerW,
				[]string{
					theme.StyleHeader.Render("Habits"),
					theme.StyleNormal.Render(fmt.Sprintf("%d/%d completed", completedHabits, totalHabits)),
					habitMeta,
				},
				func(barWidth int) string {
					return renderDailyHabitBar(theme, completedHabits, failedHabits, totalHabits, barWidth)
				},
				compactSummaryBarWidth,
			),
			theme.StyleDim.Render(fmt.Sprintf("failed %d   remaining %d", failedHabits, max(0, totalHabits-completedHabits-failedHabits))),
		)
	default:
		lines = append(lines,
			issueSummary,
			renderDailyIssueStatusBar(theme, issueStatusCounts, issueBarWidth),
			theme.StyleDim.Render(renderDailyIssueLegend(issueStatusCounts)),
			"",
		)
		lines = append(lines, habitVisual...)
	}
	lines = clipDailySummaryLines(theme, lines, height)
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width - 2).Height(max(1, height-2)).Render(stringsJoin(lines))
}

func renderCompactSummaryRow(width int, segments []string, renderBar func(int) string, sizeBar func(int, int) int) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if strings.TrimSpace(segment) != "" {
			parts = append(parts, segment)
		}
	}
	text := strings.Join(parts, "  ")
	if sizeBar == nil {
		sizeBar = compactSummaryBarWidth
	}
	barWidth := sizeBar(width, lipgloss.Width(text))
	if barWidth < 1 || renderBar == nil {
		return truncate(text, width)
	}
	bar := renderBar(barWidth)
	if strings.TrimSpace(bar) == "" {
		return truncate(text, width)
	}
	bar = ansiAwareTruncate(bar, barWidth)
	row := text + "  " + bar
	return truncate(row, width)
}

func renderCompactMetadataRow(width int, left, right string) string {
	row := left
	if strings.TrimSpace(right) == "" {
		return truncate(row, width)
	}
	remaining := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if remaining < 1 {
		return truncate(left+"  "+right, width)
	}
	return left + strings.Repeat(" ", remaining+2) + right
}

func compactSummaryBarWidth(totalWidth, textWidth int) int {
	remaining := totalWidth - textWidth - 2
	if remaining < 8 {
		return 0
	}
	if remaining > 18 {
		return 18
	}
	return remaining
}

func tinySummaryBarWidth(totalWidth, textWidth int) int {
	remaining := totalWidth - textWidth - 2
	if remaining < 6 {
		return 0
	}
	if remaining > 12 {
		return 12
	}
	return remaining
}

func compactIssueLegend(counts map[string]int) string {
	order := []string{"done", "abandoned", "blocked", "in_progress", "in_review", "ready", "planned", "backlog"}
	labels := map[string]string{
		"done":        "d",
		"abandoned":   "a",
		"blocked":     "b",
		"in_progress": "ip",
		"in_review":   "ir",
		"ready":       "r",
		"planned":     "p",
		"backlog":     "bk",
	}
	parts := make([]string, 0, len(order))
	for _, status := range order {
		if count := counts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s%d", labels[status], count))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func compactHabitProgress(loggedMinutes, targetMinutes int) string {
	if targetMinutes > 0 {
		return fmt.Sprintf("%d/%dm", loggedMinutes, targetMinutes)
	}
	return fmt.Sprintf("%dm", loggedMinutes)
}

func ansiAwareTruncate(s string, width int) string {
	if width < 1 {
		return ""
	}
	return ansi.Truncate(s, width, "")
}

func clipDailySummaryLines(theme Theme, lines []string, height int) []string {
	maxLines := height - 4
	flattened := make([]string, 0, len(lines))
	for _, line := range lines {
		flattened = append(flattened, strings.Split(line, "\n")...)
	}
	if maxLines < 1 || len(flattened) <= maxLines {
		return flattened
	}
	if maxLines == 1 {
		return []string{theme.StyleDim.Render("...")}
	}
	clipped := append([]string{}, flattened[:maxLines-1]...)
	return append(clipped, theme.StyleDim.Render("..."))
}

func renderDailyIssues(theme Theme, state ContentState, width, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	issues := make([]apiIssue, 0, len(state.DailyIssues))
	for _, issue := range state.DailyIssues {
		issues = append(issues, newAPIIssue(issue.ID, issue.Title, issue.Status, issue.EstimateMinutes, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt))
	}
	indices := filteredIssueIndices(issues, state.Filters["issues"])
	total := len(indices)
	actions := paneActionsForState(theme, state, active)
	actionLine := renderPaneActionLine(theme, state.Filters["issues"], width-6, actions)
	lines := []string{theme.StylePaneTitle.Render("Planned Tasks [1]"), theme.StyleHeader.Render(defaultScopeLabel(state.Context)), actionLine}
	if len(issues) == 0 || total == 0 {
		lines = append(lines, theme.StyleDim.Render("No planned tasks for this date"))
		return renderPaneBox(theme, active, width, height, stringsJoin(lines))
	}
	statusW := 11
	estimateW := 8
	repoW := max(10, width/7)
	streamW := max(10, width/7)
	titleW := width - statusW - estimateW - repoW - streamW - 16
	if titleW < 14 {
		titleW = 14
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, "Issue", statusW, "Status", estimateW, "Estimate", repoW, "Repo", streamW, "Stream")
	lines = append(lines, theme.StyleDim.Render(truncate(header, width-6)))
	inner := remainingPaneHeight(height, lines)
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		issue := issues[indices[i]]
		meta := issueMetaByID(state.AllIssues, issue.ID)
		repoName, streamName := "-", "-"
		if meta != nil {
			repoName = meta.RepoName
			streamName = meta.StreamName
		}
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = fmt.Sprintf("%dm", *issue.EstimateMinutes)
		}
		title := issue.Title + issueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, truncate(title, titleW), statusW, truncate(plainIssueStatus(string(issue.Status)), statusW), estimateW, estimate, repoW, truncate(repoName, repoW), streamW, truncate(streamName, streamW))
		lines = append(lines, renderPaneRowStyled(theme, i, cur, active, row, issueStatusStyle(theme, string(issue.Status)), width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, width, height, stringsJoin(lines))
}

func renderDailyHabits(theme Theme, state ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	indices := filteredStrings(habitDailyItems(state.DueHabits), state.Filters["habits"])
	total := len(indices)
	actions := paneActionsForState(theme, state, active)
	actionLine := renderPaneActionLine(theme, state.Filters["habits"], width-6, actions)
	lines := []string{theme.StylePaneTitle.Render("Habits Due [2]"), theme.StyleHeader.Render(defaultScopeLabel(state.Context)), actionLine}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No due habits for this date"))
		return renderPaneBox(theme, active, width, height, stringsJoin(lines))
	}
	inner := remainingPaneHeight(height, lines)
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		habit := state.DueHabits[indices[i]]
		status := "[ ]"
		style := &theme.StyleNormal
		switch habit.Status {
		case "completed":
			status = "[x]"
			s := lipgloss.NewStyle().Foreground(theme.ColorGreen)
			style = &s
		case "failed":
			status = "[!]"
			s := lipgloss.NewStyle().Foreground(theme.ColorRed)
			style = &s
		}
		duration := ""
		if habit.DurationMinutes != nil {
			duration = fmt.Sprintf("  %dm", *habit.DurationMinutes)
		} else if habit.TargetMinutes != nil {
			duration = fmt.Sprintf("  target %dm", *habit.TargetMinutes)
		}
		row := fmt.Sprintf("%s %s%s", status, habit.Name, duration)
		lines = append(lines, renderPaneRowStyled(theme, i, cur, active, row, style, width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, width, height, stringsJoin(lines))
}

func renderDailyIssueStatusBar(theme Theme, counts map[string]int, width int) string {
	if width < 8 {
		width = 8
	}
	order := []struct {
		status string
		color  lipgloss.Color
	}{
		{status: "done", color: theme.ColorGreen},
		{status: "abandoned", color: theme.ColorRed},
		{status: "blocked", color: theme.ColorRed},
		{status: "in_progress", color: theme.ColorYellow},
		{status: "in_review", color: theme.ColorMagenta},
		{status: "ready", color: theme.ColorCyan},
		{status: "planned", color: theme.ColorBlue},
		{status: "backlog", color: theme.ColorSubtle},
	}
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	segments := make([]string, 0, len(order))
	used := 0
	remainingStatuses := 0
	for _, item := range order {
		if counts[item.status] > 0 {
			remainingStatuses++
		}
	}
	for _, item := range order {
		count := counts[item.status]
		if count <= 0 {
			continue
		}
		segmentWidth := (count * width) / total
		if segmentWidth == 0 {
			segmentWidth = 1
		}
		if used+segmentWidth > width {
			segmentWidth = width - used
		}
		if remainingStatuses == 1 {
			segmentWidth = width - used
		}
		if segmentWidth <= 0 {
			continue
		}
		segments = append(segments, lipgloss.NewStyle().Foreground(item.color).Render(strings.Repeat("█", segmentWidth)))
		used += segmentWidth
		remainingStatuses--
	}
	if used < width {
		segments = append(segments, theme.StyleDim.Render(strings.Repeat("█", width-used)))
	}
	return strings.Join(segments, "")
}

func renderDailyIssueLegend(counts map[string]int) string {
	labels := []struct {
		status string
		label  string
	}{
		{status: "done", label: "done"},
		{status: "abandoned", label: "abandoned"},
		{status: "blocked", label: "blocked"},
		{status: "in_progress", label: "active"},
		{status: "in_review", label: "review"},
		{status: "ready", label: "ready"},
		{status: "planned", label: "planned"},
		{status: "backlog", label: "backlog"},
	}
	parts := make([]string, 0, len(labels))
	for _, item := range labels {
		if counts[item.status] <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %d", item.label, counts[item.status]))
	}
	if len(parts) == 0 {
		return "no issues scheduled"
	}
	return strings.Join(parts, "   ")
}

func renderDailyHabitBar(theme Theme, completed, failed, total, width int) string {
	if width < 8 {
		width = 8
	}
	if total <= 0 {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	completedWidth := (completed * width) / total
	failedWidth := (failed * width) / total
	if completed > 0 && completedWidth == 0 {
		completedWidth = 1
	}
	if failed > 0 && failedWidth == 0 {
		failedWidth = 1
	}
	if completedWidth+failedWidth > width {
		failedWidth = max(0, width-completedWidth)
	}
	remainingWidth := width - completedWidth - failedWidth
	return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(strings.Repeat("█", completedWidth)) +
		lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("█", failedWidth)) +
		theme.StyleDim.Render(strings.Repeat("█", remainingWidth))
}
