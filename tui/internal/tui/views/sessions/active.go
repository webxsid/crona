package sessions

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func renderActiveView(theme types.Theme, state types.ContentState) string {
	var activeIssue *api.IssueWithMeta
	if state.Timer != nil && state.Timer.IssueID != nil {
		activeIssue = sessionmeta.IssueMetaByID(state.AllIssues, *state.Timer.IssueID)
	}
	total := state.Timer.ElapsedSeconds + state.Elapsed
	elapsed := viewhelpers.FormatClockText(total)
	seg := "work"
	if state.Timer.SegmentType != nil {
		seg = string(*state.Timer.SegmentType)
	}
	stateColor := theme.ColorGreen
	timerTitle := "Focus Session"
	timerHint := "p=pause  x=end  z=stash  i=context"
	if state.Timer.State == "paused" {
		stateColor = theme.ColorYellow
		timerTitle = "Paused For"
		timerHint = "r=resume  x=end  z=stash  i=context"
		seg = "paused"
	}
	leftW := state.Width - 4
	totalH := state.Height
	totalH = max(totalH, 1)
	timerH, issueH := viewhelpers.SplitVertical(totalH, 8, 8, max(8, totalH/2))
	clockWidth := max(12, leftW-4)
	clockHeight := max(7, timerH-8)
	timerText := sessionmeta.RenderResponsiveClock(elapsed, clockWidth, clockHeight, stateColor, theme.ColorDim)
	timerText = lipgloss.NewStyle().
		Width(clockWidth).
		AlignHorizontal(lipgloss.Center).
		Render(timerText)
	priorWorkedSeconds, completedSessions := sessionmeta.SummarizeCompletedSessions(state.IssueSessions)
	metadataLines := []string{
		theme.StyleDim.Render(fmt.Sprintf("%s  ·  Completed sessions: %d", strings.ToUpper(seg), completedSessions)),
	}
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		metadataLines = append(metadataLines, theme.StyleDim.Render(sessionmeta.FormatEstimateProgress(priorWorkedSeconds+total, *activeIssue.EstimateMinutes)))
	}
	centerWidth := max(1, leftW-4)
	timerBody := strings.Join([]string{
		timerTitle,
		"",
		lipgloss.NewStyle().
			Width(centerWidth).
			AlignHorizontal(lipgloss.Center).
			Render(lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText)),
		"",
		centerLines(metadataLines, centerWidth),
	}, "\n")
	timerSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(stateColor).
		Padding(1, 2).
		Width(leftW).
		AlignVertical(lipgloss.Top).
		Height(max(1, timerH-2)).
		Render(timerBody)
	issueSection := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCyan).
		Padding(1, 2).
		Width(leftW).
		Height(max(1, issueH-2)).
		Render(strings.Join(sessionIssueCompactLines(theme, activeIssue, leftW, state.Height, timerHint), "\n"))
	return lipgloss.JoinVertical(lipgloss.Left, timerSection, issueSection)
}

func centerLines(lines []string, width int) string {
	centered := make([]string, 0, len(lines))
	for _, line := range lines {
		centered = append(centered, lipgloss.NewStyle().Width(max(1, width)).AlignHorizontal(lipgloss.Center).Render(line))
	}
	return strings.Join(centered, "\n")
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
