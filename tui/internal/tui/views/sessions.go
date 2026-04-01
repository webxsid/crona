package views

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

func renderSessionView(theme Theme, state ContentState) string {
	if state.View == "session_history" {
		return renderSessionHistory(theme, state)
	}
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
	timerHint := "p=pause  x=end  z=stash  i=context  [ ]=session/history/scratch"
	if state.Timer.State == "paused" {
		stateColor = theme.ColorYellow
		timerTitle = "Paused For"
		timerHint = "r=resume  x=end  z=stash  i=context  [ ]=session/history/scratch"
		seg = "paused"
	}
	leftW := state.Width - 4
	totalH := max(12, state.Height)
	timerH, issueH := splitVertical(totalH, 10, 7, max(10, totalH*3/5))
	timerText := renderBigClock(elapsed)
	priorWorkedSeconds, completedSessions := summarizeCompletedSessions(state.IssueSessions)
	progress := theme.StyleDim.Render(fmt.Sprintf("Completed sessions: %d", completedSessions))
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		progress += "\n" + theme.StyleDim.Render(formatEstimateProgress(priorWorkedSeconds+total, *activeIssue.EstimateMinutes))
	}
	timerSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(stateColor).
		Padding(1, 2).
		Width(leftW).
		Height(max(1, timerH-2)).
		Render(fmt.Sprintf("%s\n\n%s\n\n%s%s", timerTitle, lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText), theme.StyleDim.Render(strings.ToUpper(seg)), "\n\n"+progress))
	issueSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Padding(1, 2).
		Width(leftW).
		Height(max(1, issueH-2)).
		Render(strings.Join(sessionIssueCompactLines(theme, activeIssue, leftW, state.Height, timerHint), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, timerSection, issueSection)
}

func renderSessionHistory(theme Theme, state ContentState) string {
	active := state.Pane == "sessions"
	cur := state.Cursors["sessions"]
	indices := filteredSessionIndices(state.SessionHistory, state.Filters["sessions"])
	total := len(indices)
	title := state.SessionHistoryTitle
	if strings.TrimSpace(title) == "" {
		title = "Session History"
	}
	subtitle := state.SessionHistoryMeta
	if strings.TrimSpace(subtitle) == "" {
		subtitle = "Recent sessions across the workspace"
	}
	actionLine := renderPaneActionLine(theme, state.Filters["sessions"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render(title), theme.StyleDim.Render(subtitle), actionLine}
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
	inner := remainingPaneHeight(state.Height, lines)
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

func sessionIssueCompactLines(theme Theme, activeIssue *api.IssueWithMeta, width, height int, timerHint string) []string {
	lines := []string{"Active Issue", ""}
	if activeIssue == nil {
		lines = append(lines, "No issue selected", "", theme.StyleDim.Render(timerHint))
		return lines
	}
	maxLine := max(20, width-12)
	lines = append(lines,
		fmt.Sprintf("[%s/%s]", activeIssue.RepoName, activeIssue.StreamName),
		truncate(activeIssue.Title, maxLine),
	)
	if activeIssue.EstimateMinutes != nil && *activeIssue.EstimateMinutes > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("Estimate %dm", *activeIssue.EstimateMinutes)))
	}
	if activeIssue.Description != nil && strings.TrimSpace(*activeIssue.Description) != "" {
		lines = append(lines, theme.StyleDim.Render("Desc  "+truncate(collapseSpace(*activeIssue.Description), maxLine)))
	}
	if activeIssue.Notes != nil && strings.TrimSpace(*activeIssue.Notes) != "" {
		lines = append(lines, theme.StyleDim.Render("Notes "+truncate(collapseSpace(*activeIssue.Notes), maxLine)))
	}
	if height >= 34 {
		lines = append(lines, "", theme.StyleDim.Render("[i] open full context"))
	}
	lines = append(lines, theme.StyleDim.Render(timerHint))
	return lines
}

func collapseSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
