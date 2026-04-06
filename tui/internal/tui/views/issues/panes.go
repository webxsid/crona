package issues

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	issuecore "crona/tui/internal/tui/views/issuecore"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderIssuePane(theme types.Theme, state types.ContentState, title, subtitle string, indices []int, offset int, showFilter bool, height int, emptyText string, sectionActive bool) string {
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
		theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	}
	if paneActive {
		lines = append(lines, viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})))
	} else if showFilter {
		lines = append(lines, viewchrome.RenderFilterLine(theme, state.Filters["issues"], width-6))
	} else {
		lines = append(lines, theme.StyleDim.Render(""))
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, "Issue", statusW, "Status", estimateW, "Estimate", repoW, "Repo", streamW, "Stream")
	lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(header, width-6)))
	inner := viewchrome.RemainingPaneHeight(height, lines)

	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return viewchrome.RenderPaneBox(theme, paneActive, width, height, viewhelpers.StringsJoin(lines))
	}

	start, end := viewchrome.ListWindow(max(0, localCur), len(indices), inner)
	if !paneActive && localCur >= len(indices) {
		start, end = viewchrome.ListWindow(len(indices)-1, len(indices), inner)
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
		title := issue.Title + issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt)
		row := fmt.Sprintf("%-2s %-*s %-*s %-*s %-*s %-*s", "", titleW, viewhelpers.Truncate(title, titleW), statusW, viewhelpers.Truncate(issuecore.PlainIssueStatus(string(issue.Status)), statusW), estimateW, estimate, repoW, viewhelpers.Truncate(issue.RepoName, repoW), streamW, viewhelpers.Truncate(issue.StreamName, streamW))

		selected := paneActive && pos == localCur
		lines = append(lines, renderIssueRow(theme, row, width, selected, active, string(issue.Status)))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, paneActive, width, height, viewhelpers.StringsJoin(lines))
}

func renderCompactIssuePane(theme types.Theme, state types.ContentState, title, subtitle string, indices []int, offset int, emptyText string, activeSection bool, height int) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"] - offset
	paneActive := active && activeSection
	lines := []string{
		theme.StylePaneTitle.Render(title),
		theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	}
	if paneActive {
		lines = append(lines, viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})))
	} else {
		lines = append(lines, viewchrome.RenderFilterLine(theme, state.Filters["issues"], state.Width-6))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return viewchrome.RenderPaneBox(theme, paneActive, state.Width, height, viewhelpers.StringsJoin(lines))
	}
	start, end := viewchrome.ListWindow(max(0, cur), len(indices), inner)
	for pos := start; pos < end; pos++ {
		lines = append(lines, renderCompactIssueRow(theme, state.Width, paneActive && pos == cur, active, state.DefaultIssues[indices[pos]]))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("... %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, paneActive, state.Width, height, viewhelpers.StringsJoin(lines))
}

func renderCompactIssueRow(theme types.Theme, width int, selected, active bool, issue api.IssueWithMeta) string {
	parts := []string{viewhelpers.Truncate(issue.Title+issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt), max(18, width/2))}
	parts = append(parts, viewhelpers.Truncate(issuecore.PlainIssueStatus(string(issue.Status)), 11))
	if issue.EstimateMinutes != nil {
		parts = append(parts, fmt.Sprintf("%dm", *issue.EstimateMinutes))
	}
	row := strings.Join(parts, "  ")
	contentStyle := issuecore.IssueStatusStyle(theme, string(issue.Status))
	if contentStyle != nil {
		row = contentStyle.Render(row)
	}
	row = viewhelpers.Truncate(row, width-6)
	if selected && active {
		return theme.StyleCursor.Render("▶ " + row)
	}
	if selected {
		return theme.StyleSelected.Render("  " + row)
	}
	return theme.StyleNormal.Render("  " + row)
}

func renderCompactFooter(theme types.Theme, label string, indices []int, state types.ContentState, height int) string {
	lines := []string{
		theme.StylePaneTitle.Render(label),
		theme.StyleDim.Render(fmt.Sprintf("%d issues", len(indices))),
	}
	if len(indices) > 0 && height > 3 {
		lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(state.DefaultIssues[indices[0]].Title, state.Width-6)))
	}
	return viewchrome.RenderPaneBox(theme, false, state.Width, height, viewhelpers.StringsJoin(lines))
}

func renderIssueRow(theme types.Theme, row string, width int, selected, active bool, status string) string {
	contentStyle := issuecore.IssueStatusStyle(theme, status)
	line := viewhelpers.Truncate(strings.TrimPrefix(row, "  "), width-6)
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
