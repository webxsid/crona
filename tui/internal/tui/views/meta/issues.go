package meta

import (
	"fmt"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	types "crona/tui/internal/tui/views/types"
)

func renderIssues(theme types.Theme, state types.ContentState, width, height int, emptyText string) string {
	active := state.Pane == "issues"
	cur := state.Cursors["issues"]
	issues := make([]issuecore.APIIssue, 0, len(state.Issues))
	for _, issue := range state.Issues {
		issues = append(issues, issuecore.NewAPIIssue(issue.ID, issue.Title, issue.Status, issue.EstimateMinutes, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt))
	}
	indices := issuecore.FilteredIssueIndices(issues, state.Filters["issues"])
	total := len(indices)
	lines := []string{
		theme.StylePaneTitle.Render("Issues [3]"),
		viewchrome.RenderPaneActionLine(theme, state.Filters["issues"], width-6, func() []string {
			if !active {
				return nil
			}
			return paneActions(theme, state, "issues")
		}()),
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render(emptyText))
		return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
	}
	inner := viewchrome.RemainingPaneHeight(height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		issue := issues[indices[i]]
		text := fmt.Sprintf("[%s] %s%s", issuecore.PlainIssueStatus(string(issue.Status)), issue.Title, issuecore.IssueDueSuffix(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt))
		lines = append(lines, viewchrome.RenderPaneRowStyled(theme, i, cur, active, text, issuecore.IssueStatusStyle(theme, string(issue.Status)), width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}
