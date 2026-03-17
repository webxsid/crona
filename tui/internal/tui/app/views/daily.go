package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderDailyView(theme Theme, state ContentState) string {
	summaryH, listH := splitVertical(state.Height, 10, 8, state.Height/3)
	leftW, rightW := splitHorizontal(state.Width, 24, 24, state.Width*3/5)
	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		renderDailyIssues(theme, state, leftW, listH),
		renderDailyHabits(theme, state, rightW, listH),
	)
	return lipgloss.JoinVertical(lipgloss.Left, renderDailySummary(theme, state, state.Width, summaryH), lists)
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
		issueSummary,
		renderDailyIssueStatusBar(theme, issueStatusCounts, issueBarWidth),
		theme.StyleDim.Render(renderDailyIssueLegend(issueStatusCounts)),
		"",
		stringsJoin(habitVisual),
	}
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorDim).Padding(1, 2).Width(width - 2).Height(max(7, height-2)).Render(stringsJoin(lines))
}

func renderDailyIssues(theme Theme, state ContentState, width, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	issues := make([]apiIssue, 0, len(state.DailyIssues))
	for _, issue := range state.DailyIssues {
		issues = append(issues, newAPIIssue(issue.ID, issue.Title, issue.Status, issue.EstimateMinutes, issue.TodoForDate))
	}
	indices := filteredIssueIndices(issues, state.Filters["issues"])
	total := len(indices)
	inner := height - 5
	if inner < 1 {
		inner = 1
	}
	actions := paneActionsForState(theme, state, active)
	lines := []string{theme.StylePaneTitle.Render("Planned Tasks [1]"), theme.StyleHeader.Render(defaultScopeLabel(state.Context)), renderPaneActionLine(theme, state.Filters["issues"], width-6, actions)}
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
		title := issue.Title + issueDueSuffix(issue.TodoForDate)
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
	inner := height - 5
	if inner < 1 {
		inner = 1
	}
	actions := paneActionsForState(theme, state, active)
	lines := []string{theme.StylePaneTitle.Render("Habits Due [2]"), theme.StyleHeader.Render(defaultScopeLabel(state.Context)), renderPaneActionLine(theme, state.Filters["habits"], width-6, actions)}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No due habits for this date"))
		return renderPaneBox(theme, active, width, height, stringsJoin(lines))
	}
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
