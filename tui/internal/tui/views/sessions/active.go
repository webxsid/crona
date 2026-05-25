package sessions

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
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
	if state.Timer.State == "ready" {
		total = 0
	}
	elapsed := viewhelpers.FormatClockText(total)
	seg := "work"
	if state.Timer.SegmentType != nil {
		seg = string(*state.Timer.SegmentType)
	} else if state.Timer.ReadySegmentType != nil {
		seg = string(*state.Timer.ReadySegmentType)
	}
	timerTitle := "Focus Session"
	timerHint := "[p] pause  [x] end  [z] stash  [i] context"
	structured := state.Settings != nil && state.Settings.TimerMode == "structured" &&
		state.Settings.BreaksEnabled
	nextLabel := sessionActionSegmentLabel(state.Timer)
	stateColor := activeTimerColor(theme, state.Timer)
	if state.Timer.State == "ready" {
		timerTitle = "Ready For"
		timerHint = "[r] start " + nextLabel + "  [x] end  [z] stash  [i] context"
	} else if structured {
		timerHint = "[r] start " + nextLabel + "  [x] end  [z] stash  [i] context"
		if state.Timer.SegmentType != nil {
			switch *state.Timer.SegmentType {
			case sharedtypes.SessionSegmentWork, sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
				timerHint = "[m] pause  " + timerHint
			}
		}
	} else if state.Timer.State == "paused" {
		stateColor = theme.ColorYellow
		timerTitle = "Paused For"
		timerHint = "[r] resume  [x] end  [z] stash  [i] context"
		seg = "paused"
	}
	leftW := state.Width - 4
	totalH := state.Height
	totalH = max(totalH, 1)
	timerH, issueH := viewhelpers.SplitVertical(totalH, 8, 8, max(8, totalH/2))
	clockWidth := max(12, leftW-4)
	clockHeight := max(7, timerH-8)
	timerText := sessionmeta.RenderResponsiveClock(
		elapsed,
		clockWidth,
		clockHeight,
		stateColor,
		theme.ColorDim,
	)
	timerText = lipgloss.NewStyle().
		Width(clockWidth).
		AlignHorizontal(lipgloss.Center).
		Render(timerText)
	timingLabel := sessionTimingLabel(state)
	priorWorkedSeconds, completedSessions := sessionmeta.SummarizeCompletedSessions(
		state.IssueSessions,
	)
	metadataLines := []string{
		theme.StyleDim.Render(
			fmt.Sprintf("%s  ·  Completed sessions: %d", strings.ToUpper(seg), completedSessions),
		),
	}
	if activeIssue != nil && activeIssue.EstimateMinutes != nil {
		metadataLines = append(
			metadataLines,
			theme.StyleDim.Render(
				sessionmeta.FormatEstimateProgress(
					priorWorkedSeconds+total,
					*activeIssue.EstimateMinutes,
				),
			),
		)
	}
	centerWidth := max(1, leftW-4)
	progressLine := ""
	if structured {
		progressLine = structuredProgressBar(theme, state, centerWidth)
	}
	timerParts := []string{
		timerTitle,
		"",
		lipgloss.NewStyle().
			Width(centerWidth).
			AlignHorizontal(lipgloss.Center).
			Render(lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timerText)),
	}
	if timingLabel != "" {
		timerParts = append(timerParts, "", lipgloss.NewStyle().
			Width(centerWidth).
			AlignHorizontal(lipgloss.Center).
			Render(lipgloss.NewStyle().Foreground(stateColor).Bold(true).Render(timingLabel)))
	}
	if progressLine != "" {
		timerParts = append(timerParts, "", progressLine)
	}
	timerParts = append(timerParts, "", centerLines(metadataLines, centerWidth))
	timerBody := strings.Join(timerParts, "\n")
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

func sessionTimingLabel(state types.ContentState) string {
	if state.Timer == nil {
		return ""
	}
	segment := state.Timer.SegmentType
	if state.Timer.State == "ready" {
		segment = state.Timer.ReadySegmentType
		if segment == nil {
			segment = state.Timer.NextSegmentType
		}
		if segment == nil {
			return ""
		}
		return segmentReadyLabel(*segment) + " ready"
	}
	if segment == nil || state.Settings == nil {
		return ""
	}
	durationSeconds, ok := segmentDurationSeconds(state.Settings, *segment)
	if !ok || durationSeconds <= 0 {
		return ""
	}
	elapsed := state.Timer.ElapsedSeconds + state.Elapsed
	elapsed = max(0, elapsed)
	elapsed = min(elapsed, durationSeconds)
	remaining := durationSeconds - elapsed
	remainingLabel := formatRemainingMinutes(remaining)
	if remainingLabel == "" {
		return ""
	}
	if *segment == sharedtypes.SessionSegmentWork && timerNextBreakSegment(state.Timer) {
		return remainingLabel + " until break"
	}
	return remainingLabel + " left"
}

func timerNextBreakSegment(timer *api.TimerState) bool {
	if timer == nil || timer.NextSegmentType == nil {
		return false
	}
	switch *timer.NextSegmentType {
	case sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
		return true
	default:
		return false
	}
}

func segmentReadyLabel(segment sharedtypes.SessionSegmentType) string {
	switch segment {
	case sharedtypes.SessionSegmentWork:
		return "work"
	case sharedtypes.SessionSegmentShortBreak, sharedtypes.SessionSegmentLongBreak:
		return "break"
	case sharedtypes.SessionSegmentRest:
		return "rest"
	default:
		return "next"
	}
}

func segmentDurationSeconds(
	settings *api.CoreSettings,
	segment sharedtypes.SessionSegmentType,
) (int, bool) {
	if settings == nil {
		return 0, false
	}
	switch segment {
	case sharedtypes.SessionSegmentWork:
		return settings.WorkDurationMinutes * 60, settings.WorkDurationMinutes > 0
	case sharedtypes.SessionSegmentShortBreak:
		return settings.ShortBreakMinutes * 60, settings.ShortBreakMinutes > 0
	case sharedtypes.SessionSegmentLongBreak:
		return settings.LongBreakMinutes * 60, settings.LongBreakMinutes > 0
	default:
		return 0, false
	}
}

func formatRemainingMinutes(seconds int) string {
	seconds = max(0, seconds)
	minutes := seconds / 60
	if seconds%60 != 0 {
		minutes++
	}
	if minutes < 1 {
		minutes = 1
	}
	if minutes == 1 {
		return "1 min"
	}
	return fmt.Sprintf("%d mins", minutes)
}

func sessionActionSegmentLabel(timer *api.TimerState) string {
	if timer == nil {
		return "next"
	}
	if timer.ReadySegmentType != nil {
		return segmentActionLabel(string(*timer.ReadySegmentType))
	}
	if timer.NextSegmentType != nil {
		return segmentActionLabel(string(*timer.NextSegmentType))
	}
	return "next"
}

func segmentActionLabel(segment string) string {
	switch strings.TrimSpace(segment) {
	case "short_break":
		return "short break"
	case "long_break":
		return "long break"
	case "work":
		return "work"
	default:
		return "next"
	}
}

func activeTimerColor(theme types.Theme, timer *api.TimerState) lipgloss.Color {
	if timer == nil {
		return theme.ColorGreen
	}
	if timer.State == "ready" {
		return theme.ColorYellow
	}
	segment := ""
	if timer.SegmentType != nil {
		segment = string(*timer.SegmentType)
	}
	switch segment {
	case "short_break":
		return theme.ColorCyan
	case "long_break":
		return theme.ColorMagenta
	case "rest":
		return theme.ColorYellow
	default:
		if timer.State == "paused" {
			return theme.ColorYellow
		}
		return theme.ColorGreen
	}
}

func structuredProgressBar(theme types.Theme, state types.ContentState, width int) string {
	if state.Timer == nil || state.Settings == nil || width < 8 {
		return ""
	}
	durationSeconds := 0
	if state.Timer.SegmentType != nil {
		switch *state.Timer.SegmentType {
		case sharedtypes.SessionSegmentWork:
			durationSeconds = state.Settings.WorkDurationMinutes * 60
		case sharedtypes.SessionSegmentShortBreak:
			durationSeconds = state.Settings.ShortBreakMinutes * 60
		case sharedtypes.SessionSegmentLongBreak:
			durationSeconds = state.Settings.LongBreakMinutes * 60
		}
	}
	if durationSeconds <= 0 || state.Timer.State == "ready" {
		return ""
	}
	elapsed := state.Timer.ElapsedSeconds + state.Elapsed
	elapsed = max(0, elapsed)
	elapsed = min(elapsed, durationSeconds)
	filled := 0
	if durationSeconds > 0 {
		filled = elapsed * width / durationSeconds
	}
	if elapsed > 0 && filled == 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	color := activeTimerColor(theme, state.Timer)
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		theme.StyleDim.Render(strings.Repeat("█", width-filled))
	return lipgloss.NewStyle().Width(width).Render(bar)
}

func centerLines(lines []string, width int) string {
	centered := make([]string, 0, len(lines))
	for _, line := range lines {
		centered = append(
			centered,
			lipgloss.NewStyle().Width(max(1, width)).AlignHorizontal(lipgloss.Center).Render(line),
		)
	}
	return strings.Join(centered, "\n")
}

func sessionIssueCompactLines(
	theme types.Theme,
	activeIssue *api.IssueWithMeta,
	width, height int,
	timerHint string,
) []string {
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
		lines = append(
			lines,
			"",
			theme.StyleDim.Render(
				"Estimate "+helperpkg.FormatCompactDurationMinutes(*activeIssue.EstimateMinutes),
			),
		)
	}
	if activeIssue.Description != nil && strings.TrimSpace(*activeIssue.Description) != "" {
		lines = append(
			lines,
			theme.StyleDim.Render(
				"Desc  "+viewhelpers.Truncate(collapseSpace(*activeIssue.Description), maxLine),
			),
		)
	}
	if activeIssue.Notes != nil && strings.TrimSpace(*activeIssue.Notes) != "" {
		lines = append(
			lines,
			theme.StyleDim.Render(
				"Notes "+viewhelpers.Truncate(collapseSpace(*activeIssue.Notes), maxLine),
			),
		)
	}
	lines = append(lines, "", theme.StyleDim.Render(timerHint))
	return lines
}
