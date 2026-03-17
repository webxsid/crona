package views

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderSessionView(theme Theme, state ContentState) string {
	if state.Timer == nil || state.Timer.State == "idle" {
		return renderSessionHistory(theme, state)
	}
	var activeIssue *api.IssueWithMeta
	if state.Timer.IssueID != nil {
		activeIssue = issueMetaByID(state.AllIssues, *state.Timer.IssueID)
	}
	total := state.Timer.ElapsedSeconds + state.Elapsed
	elapsed := formatClock(total)
	seg := "work"
	if state.Timer.SegmentType != nil {
		seg = string(*state.Timer.SegmentType)
	}
	stateColor := theme.ColorGreen
	timerTitle := "Focus Session"
	timerHint := "p=pause  x=end  z=stash  ]=scratchpads"
	if state.Timer.State == "paused" {
		stateColor = theme.ColorYellow
		timerTitle = "Paused For"
		timerHint = "r=resume  x=end  z=stash  ]=scratchpads"
		seg = "paused"
	}
	issueBox := "No issue selected"
	if activeIssue != nil {
		issueBox = fmt.Sprintf("[%s/%s]\n%s", activeIssue.RepoName, activeIssue.StreamName, activeIssue.Title)
	}
	leftW := state.Width - 4
	timerText := renderBigClock(elapsed)
	priorWorkedSeconds, completedSessions := summarizeCompletedSessions(state.IssueSessions)
	progress := theme.StyleDim.Render(fmt.Sprintf("Completed sessions: %d", completedSessions))
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		progress += "\n" + theme.StyleDim.Render(formatEstimateProgress(priorWorkedSeconds+total, *activeIssue.EstimateMinutes))
	}
	timerSection := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(stateColor).Padding(1, 2).Width(leftW).Render(fmt.Sprintf("%s\n\n%s\n\n%s%s", timerTitle, lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText), theme.StyleDim.Render(strings.ToUpper(seg)), "\n\n"+progress))
	issueSection := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorCyan).Padding(1, 2).Width(leftW).Render("Active Issue\n\n" + issueBox + "\n\n" + theme.StyleDim.Render(timerHint))
	return lipgloss.JoinVertical(lipgloss.Left, issueSection, timerSection)
}

func renderSessionHistory(theme Theme, state ContentState) string {
	active := state.Pane == "sessions"
	cur := state.Cursors["sessions"]
	indices := filteredSessionIndices(state.SessionHistory, state.Filters["sessions"])
	total := len(indices)
	inner := state.Height - 6
	if inner < 1 {
		inner = 1
	}
	lines := []string{theme.StylePaneTitle.Render("Session History"), theme.StyleDim.Render("Recent sessions across the workspace"), renderPaneActionLine(theme, state.Filters["sessions"], state.Width-6, paneActionsForState(theme, state, active))}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No sessions recorded"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	dateW, durW := 16, 10
	issueW := state.Width - dateW - durW - 12
	if issueW < 18 {
		issueW = 18
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %s", "", dateW, "Ended", durW, "Duration", "Notes")
	lines = append(lines, theme.StyleDim.Render(truncate(header, state.Width-6)))
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render("..."))
	}
	for pos := start; pos < end; pos++ {
		entry := state.SessionHistory[indices[pos]]
		ended := entry.StartTime
		if entry.EndTime != nil && *entry.EndTime != "" {
			ended = *entry.EndTime
		}
		ended = formatSessionTimestamp(ended)
		duration := formatSessionDuration(entry.DurationSeconds, entry.StartTime, entry.EndTime)
		note := sessionHistorySummary(entry)
		row := fmt.Sprintf("%-2s %-*s %-*s %s", "", dateW, ended, durW, duration, truncate(note, issueW))
		if pos == cur && active {
			lines = append(lines, theme.StyleCursor.Render("▶ "+truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		} else if pos == cur {
			lines = append(lines, theme.StyleSelected.Render("  "+truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		} else {
			lines = append(lines, theme.StyleNormal.Render("  "+truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		}
	}
	if end < total {
		lines = append(lines, theme.StyleDim.Render("..."))
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}
