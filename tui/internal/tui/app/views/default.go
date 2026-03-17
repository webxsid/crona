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

	inner := height - len(lines) - 2
	if inner < 1 {
		inner = 1
	}

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
		title := issue.Title + issueDueSuffix(issue.TodoForDate)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, truncate(title, titleW), statusW, truncate(plainIssueStatus(string(issue.Status)), statusW), estimateW, estimate, repoW, truncate(issue.RepoName, repoW), streamW, truncate(issue.StreamName, streamW))

		selected := paneActive && pos == localCur
		lines = append(lines, renderDefaultIssueRow(theme, row, width, selected, active, string(issue.Status)))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
	}
	return renderPaneBox(theme, paneActive, width, height, strings.Join(lines, "\n"))
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
