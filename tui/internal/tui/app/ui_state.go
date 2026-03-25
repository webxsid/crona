package app

import (
	"fmt"
	"time"

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
		View:                   string(m.view),
		Pane:                   string(m.pane),
		ScratchpadOpen:         m.scratchpadOpen,
		TimerState:             timerState,
		IsDevMode:              m.isDevMode(),
		UpdateVisible:          viewsShouldShowUpdate(m.updateStatus),
		UpdateInstallAvailable: m.selfUpdateInstallAvailable(),
	})
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

func (m Model) activeTimerIssueID() *int64 {
	if m.timer != nil && m.timer.IssueID != nil && *m.timer.IssueID > 0 {
		return m.timer.IssueID
	}
	return nil
}

func (m Model) sessionHistoryScopeIssueID() *int64 {
	if issueID := m.activeTimerIssueID(); issueID != nil {
		return issueID
	}
	return nil
}

func (m Model) isSessionHistoryScopedToActiveIssue() bool {
	return m.sessionHistoryScopeIssueID() != nil
}

func (m Model) sessionHistoryTitle() string {
	if issueID := m.sessionHistoryScopeIssueID(); issueID != nil {
		if issue := m.issueMetaByID(*issueID); issue != nil {
			return fmt.Sprintf("Session History For #%d %s", issue.ID, issue.Title)
		}
		return fmt.Sprintf("Session History For Issue #%d", *issueID)
	}
	return "Session History"
}

func (m Model) sessionHistorySubtitle() string {
	if issueID := m.sessionHistoryScopeIssueID(); issueID != nil {
		if issue := m.issueMetaByID(*issueID); issue != nil {
			return fmt.Sprintf("Previous sessions for the active issue in [%s/%s]", issue.RepoName, issue.StreamName)
		}
		return "Previous sessions for the active issue"
	}
	return "Recent sessions across the workspace"
}

func (m Model) nextActiveSessionView(dir int) View {
	views := []View{ViewSessionActive, ViewSessionHistory, ViewScratch}
	current := ViewSessionActive
	for _, candidate := range views {
		if m.view == candidate {
			current = candidate
			break
		}
	}
	for i, candidate := range views {
		if candidate == current {
			return views[(i+dir+len(views))%len(views)]
		}
	}
	return ViewSessionActive
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

func (m Model) selectedConfigItem() (configItem, bool) {
	if m.view != ViewConfig || m.pane != PaneConfig {
		return configItem{}, false
	}
	rawIdx := m.filteredIndexAtCursor(PaneConfig)
	items := m.configItems()
	if rawIdx < 0 || rawIdx >= len(items) {
		return configItem{}, false
	}
	return items[rawIdx], true
}

func (m Model) selectedExportReport() (api.ExportReportFile, bool) {
	if m.view != ViewReports || m.pane != PaneExportReports {
		return api.ExportReportFile{}, false
	}
	rawIdx := m.filteredIndexAtCursor(PaneExportReports)
	if rawIdx < 0 || rawIdx >= len(m.exportReports) {
		return api.ExportReportFile{}, false
	}
	return m.exportReports[rawIdx], true
}
