package model

import (
	"fmt"
	"strings"
	"time"

	"crona/tui/internal/tui/terminaltitle"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) terminalTitle() string {
	if m.timer != nil && m.timer.State != "idle" {
		return strings.Join(compactTitleParts([]string{
			"Crona",
			m.terminalSessionTitle(),
			fmt.Sprintf("%s %s", compactElapsed(m.timer.ElapsedSeconds+m.elapsed), terminalTimerState(m)),
		}), " · ")
	}
	return strings.Join(compactTitleParts([]string{
		"Crona",
		m.terminalContextTitle(),
		terminalViewTitle(m.view),
	}), " · ")
}

func (m Model) terminalSessionTitle() string {
	if m.timer == nil || m.timer.IssueID == nil {
		return "Focus session"
	}
	for _, issue := range m.allIssues {
		if issue.ID == *m.timer.IssueID && strings.TrimSpace(issue.Title) != "" {
			return truncateTitle(strings.TrimSpace(issue.Title), 42)
		}
	}
	return fmt.Sprintf("Issue #%d", *m.timer.IssueID)
}

func (m Model) terminalContextTitle() string {
	parts := []string{}
	if m.context != nil {
		if m.context.RepoName != nil && strings.TrimSpace(*m.context.RepoName) != "" {
			parts = append(parts, truncateTitle(strings.TrimSpace(*m.context.RepoName), 24))
		}
		if m.context.StreamName != nil && strings.TrimSpace(*m.context.StreamName) != "" {
			parts = append(parts, truncateTitle(strings.TrimSpace(*m.context.StreamName), 24))
		}
	}
	return strings.Join(parts, " / ")
}

func terminalTimerState(m Model) string {
	if m.timer == nil {
		return ""
	}
	if m.timer.SegmentType != nil && strings.TrimSpace(string(*m.timer.SegmentType)) != "" && *m.timer.SegmentType != "work" {
		return strings.ToUpper(string(*m.timer.SegmentType))
	}
	if m.timer.State == "paused" {
		return "PAUSED"
	}
	return "WORK"
}

func terminalViewTitle(view View) string {
	switch view {
	case ViewAway:
		return "Away"
	case ViewDefault:
		return "Issues"
	case ViewDaily:
		return "Daily"
	case ViewRollup:
		return "Rollup"
	case ViewMeta:
		return "Meta"
	case ViewSessionHistory:
		return "History"
	case ViewHabitHistory:
		return "Habit History"
	case ViewSessionActive:
		return "Session"
	case ViewScratch:
		return "Scratchpads"
	case ViewOps:
		return "Ops"
	case ViewWellbeing:
		return "Wellbeing"
	case ViewReports:
		return "Reports"
	case ViewConfig:
		return "Config"
	case ViewSettings:
		return "Settings"
	case ViewAlerts:
		return "Alerts"
	case ViewUpdates:
		return "Updates"
	case ViewSupport:
		return "Support"
	default:
		return strings.TrimSpace(string(view))
	}
}

func (m Model) withTerminalTitle(cmd tea.Cmd) (Model, tea.Cmd) {
	if !m.terminalTitleEnabled {
		return m, cmd
	}
	title := terminaltitle.Sanitize(m.terminalTitle())
	if title == "" || title == m.lastTerminalTitle {
		return m, cmd
	}
	m.lastTerminalTitle = title
	return m, batchCmds(terminaltitle.Command(title), cmd)
}

func compactElapsed(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	duration := time.Duration(totalSeconds) * time.Second
	totalMinutes := int(duration.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func compactTitleParts(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if text := strings.TrimSpace(part); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func truncateTitle(value string, maxRunes int) string {
	if maxRunes < 4 {
		maxRunes = 4
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes-3]) + "..."
}
