package sessions

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderActiveView(theme types.Theme, state types.ContentState) string {
	var activeIssue *api.IssueWithMeta
	if state.Timer != nil && state.Timer.IssueID != nil {
		activeIssue = sessionmeta.IssueMetaByID(state.AllIssues, *state.Timer.IssueID)
	}
	total := state.Timer.ElapsedSeconds + state.Elapsed
	elapsed := viewhelpers.FormatClock(total)
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
	timerH, issueH := viewhelpers.SplitVertical(totalH, 10, 7, max(10, totalH*3/5))
	timerText := sessionmeta.RenderBigClock(elapsed)
	priorWorkedSeconds, completedSessions := sessionmeta.SummarizeCompletedSessions(state.IssueSessions)
	progress := theme.StyleDim.Render(fmt.Sprintf("Completed sessions: %d", completedSessions))
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		progress += "\n" + theme.StyleDim.Render(sessionmeta.FormatEstimateProgress(priorWorkedSeconds+total, *activeIssue.EstimateMinutes))
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

func sessionIssueCompactLines(theme types.Theme, activeIssue *api.IssueWithMeta, width, height int, timerHint string) []string {
	lines := []string{"Active Issue", ""}
	if activeIssue == nil {
		lines = append(lines, "No issue selected", "", theme.StyleDim.Render(timerHint))
		return lines
	}
	maxLine := max(20, width-12)
	lines = append(lines,
		fmt.Sprintf("[%s/%s]", activeIssue.RepoName, activeIssue.StreamName),
		viewhelpers.Truncate(activeIssue.Title, maxLine),
	)
	if activeIssue.EstimateMinutes != nil && *activeIssue.EstimateMinutes > 0 {
		lines = append(lines, theme.StyleDim.Render("Estimate "+helperpkg.FormatCompactDurationMinutes(*activeIssue.EstimateMinutes)))
	}
	if activeIssue.Description != nil && strings.TrimSpace(*activeIssue.Description) != "" {
		lines = append(lines, theme.StyleDim.Render("Desc  "+viewhelpers.Truncate(collapseSpace(*activeIssue.Description), maxLine)))
	}
	if activeIssue.Notes != nil && strings.TrimSpace(*activeIssue.Notes) != "" {
		lines = append(lines, theme.StyleDim.Render("Notes "+viewhelpers.Truncate(collapseSpace(*activeIssue.Notes), maxLine)))
	}
	if height >= 34 {
		lines = append(lines, "", theme.StyleDim.Render("[i] open full context"))
	}
	lines = append(lines, theme.StyleDim.Render(timerHint))
	return lines
}
