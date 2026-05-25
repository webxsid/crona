package issues

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	contextmeta "crona/tui/internal/tui/views/contextmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

func renderIssuePane(theme types.Theme, state types.ContentState, title, subtitle string, indices []int, offset int, showFilter bool, height int, emptyText string, sectionActive bool) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	localCur := cur - offset
	paneActive := active && sectionActive
	width := state.Width

	repoW := max(10, width/8)
	streamW := max(10, width/8)
	statusW := 14
	estimateW := 11
	spentW := 11
	titleW := width - repoW - streamW - statusW - estimateW - spentW - 18
	titleW = max(14, titleW)

	lines := []string{
		theme.StylePaneTitle.Render(title),
		theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	}
	if paneActive {
		lines = append(lines, viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], width-6, viewchrome.PaneActionsForState(theme, state, paneActive)))
	} else if showFilter {
		lines = append(lines, viewchrome.RenderFilterLine(theme, state.Filters["issues"], width-6))
	} else {
		lines = append(lines, theme.StyleDim.Render(""))
	}
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
	tableRows := make([]table.Row, 0, end-start)
	for pos := start; pos < end; pos++ {
		issue := state.DefaultIssues[indices[pos]]
		estimate := "-"
		if issue.EstimateMinutes != nil {
			estimate = fmt.Sprintf("%dm", *issue.EstimateMinutes)
		}
		spent := "-"
		if issue.WorkedSeconds > 0 {
			spent = viewhelpers.FormatCompactDurationSeconds(issue.WorkedSeconds)
		} else if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil && meta.WorkedSeconds > 0 {
			spent = viewhelpers.FormatCompactDurationSeconds(meta.WorkedSeconds)
		}
		title := issue.Title
		title += issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings)
		selected := paneActive && pos == localCur
		rowStyle := lipgloss.NewStyle()
		if statusStyle := issuecore.IssueStatusStyle(theme, string(issue.Status)); statusStyle != nil {
			rowStyle = *statusStyle
		}
		if selected {
			rowStyle = rowStyle.Bold(true)
		}
		cursor := " "
		if selected {
			cursor = "▶"
		}
		tableRows = append(tableRows, issuecore.IssueTableRow(cursor, title, issuecore.PlainIssueStatus(string(issue.Status)), estimate, spent, issue.RepoName, issue.StreamName, rowStyle))
	}
	lines = append(lines, issuecore.IssueTableView(issuecore.IssueTableColumns(titleW, statusW, estimateW, spentW, repoW, streamW), tableRows, theme.StyleDim))
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
		lines = append(lines, viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], state.Width-6, viewchrome.PaneActionsForState(theme, state, paneActive)))
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
		lines = append(lines, renderCompactIssueRow(theme, state.Width, paneActive && pos == cur, active, state.DefaultIssues[indices[pos]], state.Settings))
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("... %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, paneActive, state.Width, height, viewhelpers.StringsJoin(lines))
}

func renderCompactIssueRow(theme types.Theme, width int, selected, active bool, issue api.IssueWithMeta, settings *api.CoreSettings) string {
	title := issue.Title + issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, settings)
	parts := []string{viewhelpers.Truncate(title, max(18, width/2))}
	parts = append(parts, viewhelpers.Truncate(issuecore.PlainIssueStatus(string(issue.Status)), 11))
	if issue.EstimateMinutes != nil {
		parts = append(parts, fmt.Sprintf("%dm", *issue.EstimateMinutes))
	}
	if issue.WorkedSeconds > 0 {
		parts = append(parts, viewhelpers.FormatCompactDurationSeconds(issue.WorkedSeconds))
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
