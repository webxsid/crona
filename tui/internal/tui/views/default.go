package views

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"

	"github.com/charmbracelet/lipgloss"
)

func renderDefaultView(theme Theme, state ContentState) string {
	openIndices, completedIndices := SplitDefaultIssueIndices(state.DefaultIssues, state.Filters["issues"], state.Settings)
	if state.Height < 37 {
		return renderDefaultCompactView(theme, state, openIndices, completedIndices)
	}
	summaryH := 9
	if state.Height < 44 {
		summaryH = 7
	}
	remainingHeight := max(8, state.Height-summaryH-1)
	primaryPreferred := remainingHeight / 2
	if state.DefaultIssueSection == "completed" {
		completedH, priorityH := splitVertical(remainingHeight, 6, 8, remainingHeight*2/3)
		if completedH > 6 {
			completedH--
			priorityH++
		}
		summary := renderDefaultSummary(theme, state, summaryH)
		priorityPane := renderDefaultIssuePane(theme, state, "Active Issues [1]", "Due work and open issues", openIndices, 0, true, priorityH, "No open issues match the current filter", false)
		completedPane := renderDefaultIssuePane(theme, state, "Completed Issues [2]", "Done and abandoned, ready to revisit", completedIndices, len(openIndices), false, completedH, "No done or abandoned issues", true)
		return lipgloss.JoinVertical(lipgloss.Left, summary, priorityPane, completedPane)
	}
	priorityPreferred := remainingHeight * 2 / 3
	if priorityPreferred < primaryPreferred {
		priorityPreferred = primaryPreferred
	}
	priorityH, completedH := splitVertical(remainingHeight, 8, 6, priorityPreferred)
	if completedH > 6 {
		completedH--
		priorityH++
	}

	summary := renderDefaultSummary(theme, state, summaryH)
	priorityPane := renderDefaultIssuePane(theme, state, "Active Issues [1]", "Due work and open issues", openIndices, 0, true, priorityH, "No open issues match the current filter", state.DefaultIssueSection != "completed")
	completedPane := renderDefaultIssuePane(theme, state, "Completed Issues [2]", "Done and abandoned, ready to revisit", completedIndices, len(openIndices), false, completedH, "No done or abandoned issues", state.DefaultIssueSection == "completed")

	return lipgloss.JoinVertical(lipgloss.Left, summary, priorityPane, completedPane)
}

func renderDefaultCompactView(theme Theme, state ContentState, openIndices, completedIndices []int) string {
	summaryH := 6
	footerH := 4
	mainH := max(8, state.Height-summaryH-footerH-1)
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
	topW, bottomW := splitVertical(height, 3, 3, height/2)
	row1Left, row1Right := splitHorizontal(state.Width, 28, 28, state.Width/2)
	row2Left, row2Right := splitHorizontal(state.Width, 28, 28, state.Width/2)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderDefaultStatCard(theme, "Open", defaultOpenSummary(state), "active workload", theme.ColorYellow, row1Left, topW),
		renderDefaultStatCard(theme, "Closed", defaultClosedSummary(state), "done + abandoned", theme.ColorCyan, row1Right, topW),
	)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		renderDefaultStatCard(theme, "Due", defaultDueSummary(state), "today vs overdue", theme.ColorGreen, row2Left, bottomW),
		renderDefaultStatCard(theme, "Estimate", defaultEstimateSummary(state), "current issue load", theme.ColorMagenta, row2Right, bottomW),
	)
	return lipgloss.JoinVertical(lipgloss.Left, row1, row2)
}

func renderDefaultCompactSummary(theme Theme, state ContentState, height int) string {
	lines := []string{
		fmt.Sprintf("%s  %s", theme.StylePaneTitle.Render("Default Dashboard"), theme.StyleHeader.Render(defaultScopeLabel(state.Context))),
		truncate(defaultOpenSummary(state), state.Width-6),
		truncate(defaultClosedSummary(state), state.Width-6),
		truncate(defaultDueSummary(state), state.Width-6),
		truncate(defaultEstimateSummary(state), state.Width-6),
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

func defaultOpenSummary(state ContentState) string {
	open, inProgress, blocked := 0, 0, 0
	for _, issue := range state.DefaultIssues {
		switch issue.Status {
		case "done", "abandoned":
			continue
		case "in_progress":
			inProgress++
		case "blocked":
			blocked++
		}
		open++
	}
	return fmt.Sprintf("open %d  in progress %d  blocked %d", open, inProgress, blocked)
}

func defaultClosedSummary(state ContentState) string {
	done, abandoned := 0, 0
	for _, issue := range state.DefaultIssues {
		switch issue.Status {
		case "done":
			done++
		case "abandoned":
			abandoned++
		}
	}
	return fmt.Sprintf("done %d  abandoned %d", done, abandoned)
}

func defaultDueSummary(state ContentState) string {
	today := 0
	overdue := 0
	now := time.Now().Format("2006-01-02")
	for _, issue := range state.DefaultIssues {
		if issue.TodoForDate == nil || issue.Status == "done" || issue.Status == "abandoned" {
			continue
		}
		if *issue.TodoForDate == now {
			today++
		}
		if *issue.TodoForDate < now {
			overdue++
		}
	}
	return fmt.Sprintf("today %d  overdue %d", today, overdue)
}

func defaultEstimateSummary(state ContentState) string {
	estimated, scoped := 0, 0
	for _, issue := range state.DefaultIssues {
		if issue.Status == "done" || issue.Status == "abandoned" || issue.EstimateMinutes == nil {
			continue
		}
		estimated += *issue.EstimateMinutes
		scoped++
	}
	return fmt.Sprintf("estimated %s  scoped %d", helperpkg.FormatCompactDurationMinutes(estimated), scoped)
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
