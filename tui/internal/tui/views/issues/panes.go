package issues

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	"crona/tui/internal/tui/views/daily"
	issuecore "crona/tui/internal/tui/views/issuecore"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func issueIndices(state types.ContentState) []int {
	open, closed := issuecore.SplitDefaultIssueIndices(state.DefaultIssues, state.Filters["issues"], state.Settings)
	if state.DefaultIssueSection == "completed" {
		return closed
	}
	return open
}

func renderIssues(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "issues"
	indices := issueIndices(state)
	cur := state.Cursors["issues"]
	open, done := theme.StyleSelectedInverse.Render("Open"), theme.StyleDim.Render("Done")
	if state.DefaultIssueSection == "completed" {
		open, done = theme.StyleDim.Render("Open"), theme.StyleSelectedInverse.Render("Done")
	}
	selector := open + "  " + done
	lines := []string{selector}
	if active {
		lines = append(lines, newIssueLine(theme, state, width-2))
	}
	if active && height < 8 {
		room := width - 2 - ansi.StringWidth(selector) - 3
		if keys := newIssueLine(theme, state, room); keys != "" {
			selector += "   " + keys
		}
		lines = []string{selector}
	}
	if len(indices) == 0 {
		empty := "No open issues in this scope"
		if state.DefaultIssueSection == "completed" {
			empty = "No done or abandoned issues in this scope"
		}
		lines = append(lines, theme.StyleDim.Render(empty))
	} else {
		budget := height - 1 - len(lines)
		slots := max(1, (budget+1)/3)
		start, end := viewchrome.ListWindow(cur, len(indices), slots)
		if start > 0 {
			budget--
		}
		if end < len(indices) {
			budget--
		}
		start, end = viewchrome.ListWindow(cur, len(indices), max(1, (budget+1)/3))
		if start > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
		}
		for pos := start; pos < end; pos++ {
			if pos > start {
				lines = append(lines, "")
			}
			issue := state.DefaultIssues[indices[pos]]
			title := issue.Title
			if penalty := daily.PenaltyMarkerText(state, issue.ID); penalty != "" {
				title += "  " + penalty
			}
			date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings)
			room := max(1, width-4)
			if date != "" {
				field := "  [" + date + "]"
				title = ansi.Truncate(title, max(1, room-ansi.StringWidth(field)), "…") + field
			}
			marker := "  "
			style := theme.StyleNormal
			if statusStyle := issuecore.IssueStatusStyle(theme, string(issue.Status)); statusStyle != nil {
				style = *statusStyle
			}
			if pos == cur && active {
				marker = viewchrome.SelectionCursor + " "
				style = theme.StyleCursor
			}
			lines = append(lines, style.Render(marker+ansi.Truncate(title, room, "…")))
			if pos == cur && active {
				lines = append(lines, daily.TaskActionLine(theme, state, width))
				continue
			}
			worked := issue.WorkedSeconds
			if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); worked <= 0 && meta != nil {
				worked = meta.WorkedSeconds
			}
			meta := strings.Join([]string{issuecore.IssueContextLabel(issue.RepoName, issue.StreamName), issuecore.IssueWorkedEstimateCompactLabel(worked, issue.EstimateMinutes), issuecore.PlainIssueStatus(string(issue.Status))}, " | ")
			lines = append(lines, theme.StyleDim.Render("  "+ansi.Truncate(meta, max(1, width-4), "…")))
		}
		if end < len(indices) {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", len(indices)-end)))
		}
	}
	meta := fmt.Sprintf("%d", len(indices))
	if active {
		meta += "  [←/→]"
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{Title: "Issues", Meta: meta, Focused: active, Width: width, Height: height, Lines: lines})
}
