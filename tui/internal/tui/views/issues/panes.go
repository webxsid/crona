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
	viewui "crona/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

func renderIssuePane(
	theme types.Theme,
	state types.ContentState,
	title, subtitle string,
	indices []int,
	offset int,
	showFilter bool,
	height int,
	emptyText string,
	sectionActive bool,
) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	localCur := cur - offset
	paneActive := active && sectionActive
	base := viewui.PaneBase{Focused: paneActive, Width: state.Width, Height: height, Cursor: cur}
	width := state.Width
	layout := issuecore.IssueTableLayoutForWidth(width)

	lines := base.HeaderLines(
		base.TitleLine(theme, title),
		theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	)
	lines = append(
		lines,
		base.ControlLine(
			theme,
			state.Filters["issues"],
			width-6,
			paneActive,
			viewchrome.PaneActionsForState(theme, state, paneActive),
			showFilter,
		),
	)
	inner := viewchrome.RemainingPaneHeight(height, lines)

	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return base.Render(theme, viewhelpers.StringsJoin(lines))
	}

	start, end := viewchrome.ListWindow(max(0, localCur), len(indices), inner)
	if !paneActive && localCur >= len(indices) {
		start, end = viewchrome.ListWindow(len(indices)-1, len(indices), inner)
	}
	if start > 0 {
		lines = append(lines, base.MoreAbove(theme, start))
	}
	tableRows := make([]table.Row, 0, end-start)
	for pos := start; pos < end; pos++ {
		issue := state.DefaultIssues[indices[pos]]
		workedSeconds := issue.WorkedSeconds
		if workedSeconds <= 0 {
			if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil && meta.WorkedSeconds > 0 {
				workedSeconds = meta.WorkedSeconds
			}
		}
		title := issue.Title
		title += issuecore.IssueDueSuffix(
			issue.Status,
			issue.TodoForDate,
			issue.CompletedAt,
			issue.AbandonedAt,
			state.Settings,
		)
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
			cursor = viewchrome.SelectionCursor
		}
		tableRows = append(tableRows, issuecore.IssueTableRow(cursor, issuecore.IssueTableData{
			Issue:    title,
			Status:   issuecore.PlainIssueStatus(string(issue.Status)),
			Estimate: issuecore.IssueEstimateLabel(issue.EstimateMinutes),
			Worked:   issuecore.IssueWorkedLabel(workedSeconds),
			Repo:     issue.RepoName,
			Stream:   issue.StreamName,
			Context:  issuecore.IssueContextLabel(issue.RepoName, issue.StreamName),
			Effort:   issuecore.IssueWorkedEstimateCompactLabel(workedSeconds, issue.EstimateMinutes),
		}, rowStyle))
	}
	tablePane := viewui.TablePane{
		Columns:     issuecore.IssueTableColumns(layout),
		Rows:        tableRows,
		HeaderStyle: theme.StyleDim,
	}
	lines = append(lines, tablePane.View())
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, base.MoreBelow(theme, remaining))
	}
	return base.Render(theme, viewhelpers.StringsJoin(lines))
}

func renderCompactIssuePane(
	theme types.Theme,
	state types.ContentState,
	title, subtitle string,
	indices []int,
	offset int,
	emptyText string,
	activeSection bool,
	height int,
) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"] - offset
	paneActive := active && activeSection
	base := viewui.PaneBase{Focused: paneActive, Width: state.Width, Height: height, Cursor: cur}
	lines := base.HeaderLines(
		base.TitleLine(theme, title),
		theme.StyleHeader.Render(contextmeta.DefaultScopeLabel(state.Context)),
		theme.StyleDim.Render(subtitle),
	)
	lines = append(
		lines,
		base.ControlLine(
			theme,
			state.Filters["issues"],
			state.Width-6,
			paneActive,
			viewchrome.PaneActionsForState(theme, state, paneActive),
			true,
		),
	)
	inner := viewchrome.RemainingPaneHeight(height, lines)
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return base.Render(theme, viewhelpers.StringsJoin(lines))
	}
	start, end := viewchrome.ListWindow(max(0, cur), len(indices), inner)
	for pos := start; pos < end; pos++ {
		lines = append(
			lines,
			renderCompactIssueRow(
				theme,
				state.Width,
				paneActive && pos == cur,
				active,
				state.DefaultIssues[indices[pos]],
				state.Settings,
			),
		)
	}
	if remaining := len(indices) - end; remaining > 0 {
		lines = append(lines, base.MoreBelow(theme, remaining))
	}
	return base.Render(theme, viewhelpers.StringsJoin(lines))
}

func renderCompactIssueRow(
	theme types.Theme,
	width int,
	selected, active bool,
	issue api.IssueWithMeta,
	settings *api.CoreSettings,
) string {
	const compactIssueRowGap = "   "
	title := issue.Title + issuecore.IssueDueSuffix(
		issue.Status,
		issue.TodoForDate,
		issue.CompletedAt,
		issue.AbandonedAt,
		settings,
	)
	parts := []string{
		viewhelpers.Truncate(title, max(18, width/2-1)),
		viewhelpers.Truncate(issuecore.PlainIssueStatus(string(issue.Status)), 11),
		viewhelpers.Truncate(issuecore.IssueContextLabel(issue.RepoName, issue.StreamName), max(14, width/4-1)),
		viewhelpers.Truncate(issuecore.IssueWorkedEstimateCompactLabel(issue.WorkedSeconds, issue.EstimateMinutes), max(14, width/4-1)),
	}
	row := strings.Join(parts, compactIssueRowGap)
	contentStyle := issuecore.IssueStatusStyle(theme, string(issue.Status))
	if contentStyle != nil {
		row = contentStyle.Render(row)
	}
	row = viewhelpers.Truncate(row, width-4)
	if selected && active {
		return theme.StyleCursor.Render(viewchrome.SelectionCursor + " " + row)
	}
	if selected {
		return theme.StyleSelected.Render("  " + row)
	}
	return theme.StyleNormal.Render("  " + row)
}

func renderCompactFooter(
	theme types.Theme,
	label string,
	indices []int,
	state types.ContentState,
	height int,
) string {
	lines := []string{
		theme.StylePaneTitle.Render(label),
		theme.StyleDim.Render(fmt.Sprintf("%d issues", len(indices))),
	}
	if len(indices) > 0 && height > 3 {
		lines = append(
			lines,
			theme.StyleDim.Render(
				viewhelpers.Truncate(state.DefaultIssues[indices[0]].Title, state.Width-6),
			),
		)
	}
	return viewchrome.RenderPaneBox(
		theme,
		false,
		state.Width,
		height,
		viewhelpers.StringsJoin(lines),
	)
}
