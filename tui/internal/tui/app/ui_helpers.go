package app

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/app/views"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	statusInfoDuration  = 3 * time.Second
	statusErrorDuration = 5 * time.Second
)

func (m Model) paneActions() []string {
	timerState := ""
	if m.timer != nil {
		timerState = m.timer.State
	}
	return views.PaneActions(viewTheme(), views.ActionsState{
		View:           string(m.view),
		Pane:           string(m.pane),
		ScratchpadOpen: m.scratchpadOpen,
		TimerState:     timerState,
		IsDevMode:      m.isDevMode(),
	})
}

func issueDueLabel(todoForDate *string) string {
	if todoForDate == nil {
		return ""
	}
	date := strings.TrimSpace(*todoForDate)
	if date == "" {
		return ""
	}
	if date == time.Now().Format("2006-01-02") {
		return "today"
	}
	return "due " + date
}

func deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func firstNonEmpty(a, b *string) string {
	if a != nil && *a != "" {
		return *a
	}
	return deref(b)
}

func truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

func sessionHistorySummary(entry api.SessionHistoryEntry) string {
	if entry.ParsedNotes != nil {
		if message := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionCommit]); message != "" {
			return message
		}
		if note := strings.TrimSpace(entry.ParsedNotes[sharedtypes.SessionNoteSectionNotes]); note != "" {
			return note
		}
	}
	if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
		return strings.TrimSpace(*entry.Notes)
	}
	return fmt.Sprintf("Issue #%d", entry.IssueID)
}

func (m Model) activeIssueWithMeta() *api.IssueWithMeta {
	var issueID int64
	if m.timer != nil && m.timer.IssueID != nil {
		issueID = *m.timer.IssueID
	} else if m.context != nil && m.context.IssueID != nil {
		issueID = *m.context.IssueID
	}
	if issueID == 0 {
		return nil
	}
	return m.issueMetaByID(issueID)
}

func (m Model) issueMetaByID(issueID int64) *api.IssueWithMeta {
	for i := range m.allIssues {
		if m.allIssues[i].ID == issueID {
			return &m.allIssues[i]
		}
	}
	return nil
}

func (m Model) defaultScopedIssues() []api.IssueWithMeta {
	if m.context == nil {
		return m.allIssues
	}
	if m.context.StreamID != nil {
		out := make([]api.IssueWithMeta, 0, len(m.allIssues))
		for _, issue := range m.allIssues {
			if issue.StreamID == *m.context.StreamID {
				out = append(out, issue)
			}
		}
		return out
	}
	if m.context.RepoID != nil {
		out := make([]api.IssueWithMeta, 0, len(m.allIssues))
		for _, issue := range m.allIssues {
			if issue.RepoID == *m.context.RepoID {
				out = append(out, issue)
			}
		}
		return out
	}
	return m.allIssues
}

func (m Model) filteredDueHabits() []api.HabitDailyItem {
	if m.context == nil {
		return m.dueHabits
	}
	if m.context.StreamID != nil {
		out := make([]api.HabitDailyItem, 0, len(m.dueHabits))
		for _, habit := range m.dueHabits {
			if habit.StreamID == *m.context.StreamID {
				out = append(out, habit)
			}
		}
		return out
	}
	if m.context.RepoID != nil {
		out := make([]api.HabitDailyItem, 0, len(m.dueHabits))
		for _, habit := range m.dueHabits {
			if habit.RepoID == *m.context.RepoID {
				out = append(out, habit)
			}
		}
		return out
	}
	return m.dueHabits
}

func (m Model) dailyScopedIssues() []api.Issue {
	if m.dailySummary == nil {
		return nil
	}
	issues := m.dailySummary.Issues
	if m.context == nil {
		return issues
	}
	if m.context.StreamID != nil {
		out := make([]api.Issue, 0, len(issues))
		for _, issue := range issues {
			if issue.StreamID == *m.context.StreamID {
				out = append(out, issue)
			}
		}
		return out
	}
	if m.context.RepoID != nil {
		out := make([]api.Issue, 0, len(issues))
		for _, issue := range issues {
			meta := m.issueMetaByID(issue.ID)
			if meta != nil && meta.RepoID == *m.context.RepoID {
				out = append(out, issue)
			}
		}
		return out
	}
	return issues
}

func (m *Model) withStatus(message string, isError bool) Model {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	return *m
}

func (m *Model) setStatus(message string, isError bool) tea.Cmd {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	duration := statusInfoDuration
	if isError {
		duration = statusErrorDuration
	}
	return clearStatusAfter(m.statusSeq, duration)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sessionCommit(detail *api.SessionDetail) string {
	if detail == nil || detail.ParsedNotes == nil {
		return ""
	}
	return strings.TrimSpace(detail.ParsedNotes[sharedtypes.SessionNoteSectionCommit])
}

func (m Model) sessionDetailContentLines() []string {
	if m.sessionDetail == nil {
		return []string{"Loading session detail...", "", "[esc] close"}
	}

	detail := m.sessionDetail
	ended := "-"
	if detail.EndTime != nil && strings.TrimSpace(*detail.EndTime) != "" {
		ended = *detail.EndTime
	}
	duration := formatSessionDurationText(detail.DurationSeconds, detail.StartTime, detail.EndTime)
	lines := []string{
		fmt.Sprintf("Repo: %s", detail.RepoName),
		fmt.Sprintf("Stream: %s", detail.StreamName),
		fmt.Sprintf("Issue: #%d %s", detail.IssueID, detail.IssueTitle),
		fmt.Sprintf("Started: %s", detail.StartTime),
		fmt.Sprintf("Ended: %s", ended),
		fmt.Sprintf("Duration: %s", duration),
		"",
		fmt.Sprintf("Work: %s", formatClockText(detail.WorkSummary.WorkSeconds)),
		fmt.Sprintf("Rest: %s", formatClockText(detail.WorkSummary.RestSeconds)),
		fmt.Sprintf("Segments: %d work / %d rest", detail.WorkSummary.WorkSegments, detail.WorkSummary.RestSegments),
	}

	sectionOrder := []sharedtypes.SessionNoteSection{
		sharedtypes.SessionNoteSectionCommit,
		sharedtypes.SessionNoteSectionContext,
		sharedtypes.SessionNoteSectionWork,
		sharedtypes.SessionNoteSectionNotes,
	}
	labels := map[sharedtypes.SessionNoteSection]string{
		sharedtypes.SessionNoteSectionCommit:  "Commit",
		sharedtypes.SessionNoteSectionContext: "Context",
		sharedtypes.SessionNoteSectionWork:    "Work Summary",
		sharedtypes.SessionNoteSectionNotes:   "Notes",
	}
	for _, section := range sectionOrder {
		value := ""
		if detail.ParsedNotes != nil {
			value = strings.TrimSpace(detail.ParsedNotes[section])
		}
		if value == "" {
			continue
		}
		lines = append(lines, "", labels[section]+":")
		lines = append(lines, strings.Split(value, "\n")...)
	}
	return lines
}

func (m Model) sessionDetailViewportHeight() int {
	if m.height < 16 {
		return max(6, m.height-8)
	}
	return min(18, m.height-8)
}

func formatClockText(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func formatSessionDurationText(durationSeconds *int, start string, end *string) string {
	if durationSeconds != nil {
		return formatClockText(*durationSeconds)
	}
	if end != nil && *end != "" {
		st, se := time.Parse(time.RFC3339, start)
		et, ee := time.Parse(time.RFC3339, *end)
		if se == nil && ee == nil {
			return formatClockText(int(et.Sub(st).Seconds()))
		}
	}
	return "-"
}

func (m Model) sessionDetailMaxOffset() int {
	boxWidth := min(max(52, m.width-10), 96)
	innerWidth := boxWidth - 4
	lines := m.sessionDetailContentLines()
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapText(line, innerWidth)...)
	}
	return max(0, len(wrapped)-m.sessionDetailViewportHeight())
}

func (m Model) updateSessionDetailOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "o", "enter":
		m.sessionDetailOpen = false
		m.sessionDetail = nil
		m.sessionDetailY = 0
		return m, nil
	case "j", "down":
		if m.sessionDetailY < m.sessionDetailMaxOffset() {
			m.sessionDetailY++
		}
		return m, nil
	case "k", "up":
		if m.sessionDetailY > 0 {
			m.sessionDetailY--
		}
		return m, nil
	case "e":
		if m.sessionDetail == nil {
			return m, nil
		}
		return m.openAmendSessionDialog(m.sessionDetail.ID, sessionCommit(m.sessionDetail)), nil
	}
	return m, nil
}
