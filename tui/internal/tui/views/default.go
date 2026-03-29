package views

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderDefaultView(theme Theme, state ContentState) string {
	openIndices, completedIndices := SplitDefaultIssueIndices(state.DefaultIssues, state.Filters["issues"], state.Settings)
	if state.Height < 37 {
		return renderDefaultCompactView(theme, state, openIndices, completedIndices)
	}
	summaryH := 5
	if state.Height < 20 {
		summaryH = 4
	}
	remainingHeight := max(8, state.Height-summaryH)
	priorityH, completedH := splitVertical(remainingHeight, 8, 6, remainingHeight*2/3)

	summary := renderDefaultSummary(theme, state, summaryH)
	priorityPane := renderDefaultIssuePane(theme, state, "Active Issues [1]", "Due work and open issues", openIndices, 0, true, priorityH, "No open issues match the current filter", state.DefaultIssueSection != "completed")
	completedPane := renderDefaultIssuePane(theme, state, "Completed Issues [2]", "Done and abandoned, ready to revisit", completedIndices, len(openIndices), false, completedH, "No done or abandoned issues", state.DefaultIssueSection == "completed")

	return lipgloss.JoinVertical(lipgloss.Left, summary, priorityPane, completedPane)
}

func renderDefaultCompactView(theme Theme, state ContentState, openIndices, completedIndices []int) string {
	summaryH := 4
	footerH := 4
	mainH := max(8, state.Height-summaryH-footerH)
	mainTitle := "Active Issues [1]"
	mainSubtitle := "Due work and open issues"
	mainIndices := openIndices
	mainOffset := 0
	mainEmpty := "No open issues match the current filter"
	mainActive := state.DefaultIssueSection != "completed"
	footerTitle := "Closed"
	footerIndices := completedIndices

	if state.DefaultIssueSection == "completed" {
		mainTitle = "Completed Issues [2]"
		mainSubtitle = "Done and abandoned, ready to revisit"
		mainIndices = completedIndices
		mainOffset = len(openIndices)
		mainEmpty = "No done or abandoned issues"
		mainActive = true
		footerTitle = "Open"
		footerIndices = openIndices
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		renderDefaultCompactSummary(theme, state, summaryH),
		renderDefaultCompactIssuePane(theme, state, mainTitle, mainSubtitle, mainIndices, mainOffset, mainEmpty, mainActive, mainH),
		renderDefaultCompactFooter(theme, footerTitle, footerIndices, state, footerH),
	)
}

func renderDefaultSummary(theme Theme, state ContentState, height int) string {
	today := time.Now().Format("2006-01-02")
	dueNow, openCount, closedCount := 0, 0, 0
	for _, issue := range state.DefaultIssues {
		if isClosedIssueStatus(issue.Status) {
			closedCount++
			continue
		}
		openCount++
		if due := normalizedDueDate(issue.TodoForDate); due != "" && due <= today {
			dueNow++
		}
	}
	leftW, midW := splitHorizontal(state.Width, 20, 20, state.Width/3)
	midW, rightW := splitHorizontal(state.Width-leftW, 20, 20, (state.Width-leftW)/2)

	cards := []string{
		renderDefaultStatCard(theme, "Due Now", fmt.Sprintf("%d", dueNow), "today + overdue", theme.ColorYellow, leftW, height),
		renderDefaultStatCard(theme, "Open", fmt.Sprintf("%d", openCount), "ready to work", theme.ColorCyan, midW, height),
		renderDefaultStatCard(theme, "Closed", fmt.Sprintf("%d", closedCount), "done + abandoned", theme.ColorSubtle, rightW, height),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func renderDefaultCompactSummary(theme Theme, state ContentState, height int) string {
	today := time.Now().Format("2006-01-02")
	dueNow, openCount, closedCount := 0, 0, 0
	for _, issue := range state.DefaultIssues {
		if isClosedIssueStatus(issue.Status) {
			closedCount++
			continue
		}
		openCount++
		if due := normalizedDueDate(issue.TodoForDate); due != "" && due <= today {
			dueNow++
		}
	}
	lines := []string{
		fmt.Sprintf("%s  %s  %s", lipStyle(theme, theme.ColorYellow).Render(fmt.Sprintf("Due %d", dueNow)), theme.StyleHeader.Render(fmt.Sprintf("Open %d", openCount)), theme.StyleDim.Render(fmt.Sprintf("Closed %d", closedCount))),
		theme.StyleDim.Render("due now   open work   done + abandoned"),
	}
	return renderPaneBox(theme, false, state.Width, height, stringsJoin(lines))
}

func renderDefaultStatCard(theme Theme, label, value, hint string, border lipgloss.Color, width, height int) string {
	body := []string{
		theme.StyleDim.Render(label),
		lipStyle(theme, border).Render(value),
		theme.StyleDim.Render(hint),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width - 2).
		Height(height - 2).
		Render(strings.Join(body, "\n"))
}

func renderDefaultIssuePane(theme Theme, state ContentState, title, subtitle string, indices []int, offset int, showFilter bool, height int, emptyText string, sectionActive bool) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	localCur := cur - offset
	paneActive := active && sectionActive
	width := state.Width

	repoW := max(10, width/8)
	streamW := max(10, width/8)
	statusW := 11
	estimateW := 8
	titleW := width - repoW - streamW - statusW - estimateW - 14
	if titleW < 14 {
		titleW = 14
	}

	lines := []string{
		theme.StylePaneTitle.Render(title),
		theme.StyleHeader.Render(defaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	}
	if paneActive {
		lines = append(lines, renderPaneActionLine(theme, state.Filters["issues"], width-6, paneActionsForState(theme, state, true)))
	} else if showFilter {
		lines = append(lines, renderFilterLine(theme, state.Filters["issues"], width-6))
	} else {
		lines = append(lines, theme.StyleDim.Render(""))
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, "Issue", statusW, "Status", estimateW, "Estimate", repoW, "Repo", streamW, "Stream")
	lines = append(lines, theme.StyleDim.Render(truncate(header, width-6)))
	inner := remainingPaneHeight(height, lines)

	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return renderPaneBox(theme, paneActive, width, height, strings.Join(lines, "\n"))
	}

	start, end := listWindow(max(0, localCur), len(indices), inner)
	if !paneActive && localCur >= len(indices) {
		start, end = listWindow(len(indices)-1, len(indices), inner)
	}
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↑ %d more", start)))
	}
	for pos := start; pos < end; pos++ {
		issue := state.DefaultIssues[indices[pos]]
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = fmt.Sprintf("%dm", *issue.EstimateMinutes)
		}
		title := issue.Title + issueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, truncate(title, titleW), statusW, truncate(plainIssueStatus(string(issue.Status)), statusW), estimateW, estimate, repoW, truncate(issue.RepoName, repoW), streamW, truncate(issue.StreamName, streamW))

		selected := paneActive && pos == localCur
		lines = append(lines, renderDefaultIssueRow(theme, row, width, selected, active, string(issue.Status)))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
	}
	return renderPaneBox(theme, paneActive, width, height, strings.Join(lines, "\n"))
}

func renderDefaultCompactIssuePane(theme Theme, state ContentState, title, subtitle string, indices []int, offset int, emptyText string, activeSection bool, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"] - offset
	paneActive := active && activeSection
	lines := []string{
		theme.StylePaneTitle.Render(title),
		theme.StyleHeader.Render(defaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	}
	if paneActive {
		lines = append(lines, renderPaneActionLine(theme, state.Filters["issues"], state.Width-6, paneActionsForState(theme, state, true)))
	} else {
		lines = append(lines, renderFilterLine(theme, state.Filters["issues"], state.Width-6))
	}
	inner := remainingPaneHeight(height, lines)
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return renderPaneBox(theme, paneActive, state.Width, height, stringsJoin(lines))
	}
	start, end := listWindow(max(0, cur), len(indices), inner)
	for pos := start; pos < end; pos++ {
		issue := state.DefaultIssues[indices[pos]]
		lines = append(lines, renderDefaultCompactIssueRow(theme, state.Width, paneActive && pos == cur, active, issue))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("... %d more", remaining)))
	}
	return renderPaneBox(theme, paneActive, state.Width, height, stringsJoin(lines))
}

func renderDefaultCompactIssueRow(theme Theme, width int, selected, active bool, issue api.IssueWithMeta) string {
	parts := []string{truncate(issue.Title+issueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt), max(18, width/2))}
	parts = append(parts, truncate(plainIssueStatus(string(issue.Status)), 11))
	if issue.EstimateMinutes != nil {
		parts = append(parts, fmt.Sprintf("%dm", *issue.EstimateMinutes))
	}
	row := strings.Join(parts, "  ")
	contentStyle := issueStatusStyle(theme, string(issue.Status))
	if contentStyle != nil {
		row = contentStyle.Render(row)
	}
	row = truncate(row, width-6)
	if selected && active {
		return theme.StyleCursor.Render("▶ " + row)
	}
	if selected {
		return theme.StyleSelected.Render("  " + row)
	}
	return theme.StyleNormal.Render("  " + row)
}

func renderDefaultCompactFooter(theme Theme, label string, indices []int, state ContentState, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render(label),
		theme.StyleDim.Render(fmt.Sprintf("%d issues", len(indices))),
	}
	if len(indices) > 0 && height > 3 {
		issue := state.DefaultIssues[indices[0]]
		lines = append(lines, theme.StyleDim.Render(truncate(issue.Title, state.Width-6)))
	}
	return renderPaneBox(theme, false, state.Width, height, stringsJoin(lines))
}

func defaultScopeLabel(ctx *api.ActiveContext) string {
	if ctx == nil {
		return "Scope: All"
	}
	repoName := ""
	if ctx.RepoName != nil {
		repoName = strings.TrimSpace(*ctx.RepoName)
	}
	streamName := ""
	if ctx.StreamName != nil {
		streamName = strings.TrimSpace(*ctx.StreamName)
	}
	switch {
	case repoName != "" && streamName != "":
		return "Scope: " + repoName + " > " + streamName
	case repoName != "":
		return "Scope: " + repoName
	case streamName != "":
		return "Scope: " + streamName
	}
	return "Scope: All"
}

func renderDefaultIssueRow(theme Theme, row string, width int, selected, active bool, status string) string {
	contentStyle := issueStatusStyle(theme, status)
	line := truncate(strings.TrimPrefix(row, "  "), width-6)
	if contentStyle != nil {
		line = contentStyle.Render(line)
	}
	if selected && active {
		return theme.StyleCursor.Render("▶ " + line)
	}
	if selected {
		return theme.StyleSelected.Render("  " + line)
	}
	return theme.StyleNormal.Render("  " + line)
}
