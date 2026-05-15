package model

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/tui/commands"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	helperpkg "crona/tui/internal/tui/helpers"
	selectionpkg "crona/tui/internal/tui/selection"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) checkout() (Model, tea.Cmd) {
	snapshot := m.selectionSnapshot()
	switch m.pane {
	case PaneRepos:
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneRepos)
		if rawIdx >= 0 && rawIdx < len(m.repos) {
			repo := m.repos[rawIdx]
			return m, commands.CheckoutRepo(m.client, repo.ID)
		}
	case PaneStreams:
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneStreams)
		if rawIdx >= 0 && rawIdx < len(m.streams) {
			stream := m.streams[rawIdx]
			return m, commands.CheckoutStream(m.client, stream.ID)
		}
	}
	return m, nil
}

func (m Model) handleInputCreateAction() Model {
	if m.view == ViewDefault && m.pane == PaneIssues {
		return m.openCreateIssueDefaultDialog()
	}
	if m.view == ViewDaily {
		switch m.pane {
		case PaneIssues:
			return m.openCreateIssueDefaultDialog()
		case PaneHabits:
			repoName := "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName := "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName)
		}
	}
	if m.view == ViewMeta {
		switch m.pane {
		case PaneRepos:
			return m.openCreateRepoDialog()
		case PaneStreams:
			repoID, repoName, ok := m.selectedMetaRepo()
			if !ok {
				m.setStatus("Select or checkout a repo first", true)
				return m
			}
			return m.openCreateStreamDialog(repoID, repoName)
		case PaneIssues:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if !ok {
				m.setStatus("Select or checkout a stream first", true)
				return m
			}
			return m.openCreateIssueMetaDialog(streamID, streamName, repoName)
		case PaneHabits:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if ok {
				return m.openCreateHabitDialog(streamID, streamName, repoName)
			}
			repoName = "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName = "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName)
		}
	}
	if m.view == ViewWellbeing {
		return m.openCreateCheckInDialog()
	}
	if m.pane == PaneScratchpads {
		return m.openCreateScratchpad()
	}
	return m
}

func (m Model) handleInputOpenEditor() (Model, tea.Cmd, bool) {
	snapshot := m.selectionSnapshot()
	if m.view == ViewConfig && m.pane == PaneConfig {
		if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok && item.Editable && strings.TrimSpace(item.Path) != "" {
			return m, dialogruntime.OpenEditor(item.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(snapshot); ok && strings.TrimSpace(report.Path) != "" {
			return m, dialogruntime.OpenEditor(report.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return m, nil, true
	}
	if m.view == ViewDaily && m.pane == PaneHabits {
		if next, ok := m.openSelectedEditDialog(); ok {
			return next, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewWellbeing {
		if m.dailyCheckIn == nil {
			m = m.openCreateCheckInDialog()
		} else {
			m = m.openEditCheckInDialog()
		}
		return m, nil, true
	}
	if m.view == ViewSessionHistory && m.pane == PaneSessions {
		if entry, ok := m.selectedSessionHistoryEntry(); ok {
			m.sessionDetailOpen = true
			m.sessionDetailY = 0
			return m, commands.LoadSessionDetail(m.client, entry.ID), true
		}
		return m, nil, true
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedEditDialog(); ok {
			return next, nil, true
		}
	}
	return m, nil, false
}

func (m Model) handleInputSetHabitFailed() (tea.Cmd, bool) {
	if m.view != ViewDaily || m.pane != PaneHabits {
		return nil, false
	}
	if habit, ok := m.selectedDailyHabitRecord(); ok {
		if habit.Status == sharedtypes.HabitCompletionStatusFailed {
			return commands.UncompleteHabit(m.client, habit.ID, m.currentDashboardDate()), true
		}
		return commands.SetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusFailed, habit.DurationMinutes, habit.Notes), true
	}
	return nil, true
}

func (m Model) handleInputDeleteSelection() (Model, tea.Cmd, bool) {
	snapshot := m.selectionSnapshot()
	if m.view == ViewWellbeing {
		if m.dailyCheckIn == nil {
			return m, m.setStatus("No check-in to delete for this date", true), true
		}
		m = m.openConfirmDeleteEntity("checkin", m.currentWellbeingDate(), "this check-in")
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(snapshot); ok {
			m = m.openConfirmDeleteEntity("report", report.Path, report.Name)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.pane == PaneScratchpads {
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneScratchpads)
		if rawIdx >= 0 && rawIdx < len(m.scratchpads) {
			m = m.openConfirmDelete(m.scratchpads[rawIdx].ID)
			return m, nil, true
		}
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedDeleteDialog(); ok {
			return next, nil, true
		}
	}
	return m, nil, false
}

func (m Model) handleInputOpenSelection() (tea.Cmd, bool) {
	snapshot := m.selectionSnapshot()
	if m.view == ViewUpdates && m.updateStatus != nil {
		return commands.OpenExternalURL(m.updateStatus.ReleaseURL), true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(snapshot); ok && strings.TrimSpace(report.Path) != "" {
			return dialogruntime.OpenDefaultViewer(report.Path, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
		}
		return nil, true
	}
	if m.pane == PaneScratchpads && m.scratchpadOpen && strings.TrimSpace(m.scratchpadFilePath) != "" {
		return dialogruntime.OpenDefaultViewer(m.scratchpadFilePath, func(err error) tea.Msg { return commands.ErrMsg{Err: err} }), true
	}
	return nil, false
}

func (m Model) handleInputEnter() (Model, tea.Cmd, bool) {
	snapshot := m.selectionSnapshot()
	if m.view == ViewConfig && m.pane == PaneConfig {
		if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok {
			if item.Editable && strings.TrimSpace(item.Path) != "" {
				m = m.openViewEntityDialogWithPath(item.DetailTitle, item.Label, item.DetailMeta, item.DetailBody, item.Path)
			} else {
				m = m.openViewEntityDialog(item.DetailTitle, item.Label, item.DetailMeta, item.DetailBody)
			}
			return m, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewReports && m.pane == PaneExportReports {
		if report, ok := selectionpkg.SelectedExportReport(snapshot); ok {
			meta := fmt.Sprintf("Kind %s   Format %s   Modified %s", report.Kind, report.Format, report.ModifiedAt)
			scope := strings.TrimSpace(report.ScopeLabel)
			if scope == "" {
				scope = "-"
			}
			dateLabel := strings.TrimSpace(report.DateLabel)
			if dateLabel == "" {
				dateLabel = report.Date
			}
			body := "Scope\n" + scope + "\n\nDate\n" + dateLabel + "\n\nPath\n" + report.Path + "\n\nSize\n" + fmt.Sprintf("%d bytes", report.SizeBytes)
			m = m.openViewEntityDialog("Export Report", report.Name, meta, body)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.view == ViewSessionHistory && m.pane == PaneSessions {
		if entry, ok := m.selectedSessionHistoryEntry(); ok {
			m.sessionDetailOpen = true
			m.sessionDetailY = 0
			return m, commands.LoadSessionDetail(m.client, entry.ID), true
		}
		return m, nil, true
	}
	if m.view == ViewHabitHistory && m.pane == PaneHabitHistory {
		if entry, ok := m.selectedHabitHistoryEntry(); ok {
			durationText := formatHabitCompletionDuration(entry.DurationMinutes)
			contextLabel := strings.TrimSpace(entry.RepoName)
			if strings.TrimSpace(entry.StreamName) != "" {
				if contextLabel != "" {
					contextLabel += " > "
				}
				contextLabel += strings.TrimSpace(entry.StreamName)
			}
			if contextLabel == "" {
				contextLabel = "-"
			}
			habitLabel := strings.TrimSpace(entry.HabitName)
			if habitLabel == "" {
				habitLabel = fmt.Sprintf("Habit %d", entry.HabitID)
			}
			meta := strings.Join([]string{
				fmt.Sprintf("Habit %s", habitLabel),
				fmt.Sprintf("Context %s", contextLabel),
				fmt.Sprintf("Status %s", strings.ReplaceAll(string(entry.Status), "_", " ")),
				fmt.Sprintf("Duration %s", durationText),
			}, "   ")
			body := strings.Join([]string{
				"Habit",
				habitLabel,
				"",
				"Context",
				contextLabel,
				"",
				"Date",
				entry.Date,
				"",
				"Status",
				strings.ReplaceAll(string(entry.Status), "_", " "),
				"",
				"Duration",
				durationText,
				"",
				"Notes",
				derefHabitNotes(entry.Notes),
				"",
				"Snapshot",
				strings.TrimSpace(strings.Join([]string{
					derefHabitSnapshot(entry.SnapshotName),
					derefHabitSnapshot(entry.SnapshotDesc),
				}, " / ")),
			}, "\n")
			m = m.openViewEntityDialog("Habit History Entry", habitLabel, meta, body)
			return m, nil, true
		}
		return m, nil, true
	}
	if m.dialog == "" {
		if next, ok := m.openSelectedViewDialog(); ok {
			return next, nil, true
		}
	}
	if m.pane == PaneScratchpads {
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneScratchpads)
		if rawIdx < 0 || rawIdx >= len(m.scratchpads) {
			return m, nil, true
		}
		m.scratchpadOpen = true
		m.scratchpadMeta = helperpkg.ScratchpadMetaAt(m.scratchpads, rawIdx)
		return m, commands.OpenScratchpad(m.client, m.scratchpads, rawIdx), true
	}
	return m, nil, false
}

func formatHabitCompletionDuration(value *int) string {
	if value == nil {
		return "-"
	}
	return helperpkg.FormatCompactDurationMinutes(*value)
}

func derefHabitNotes(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "-"
	}
	return strings.TrimSpace(*value)
}

func derefHabitSnapshot(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "-"
	}
	return strings.TrimSpace(*value)
}

func (m Model) handleInputToggleHabitCompleted() (tea.Cmd, bool) {
	if m.view == ViewDaily && m.pane == PaneHabits {
		if habit, ok := m.selectedDailyHabitRecord(); ok {
			if habit.Status == sharedtypes.HabitCompletionStatusCompleted {
				return commands.UncompleteHabit(m.client, habit.ID, m.currentDashboardDate()), true
			}
			return commands.SetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusCompleted, nil, nil), true
		}
	}
	return nil, false
}

func (m Model) handleInputConfigReset() (tea.Cmd, bool) {
	if m.view != ViewConfig || m.exportAssets == nil {
		return nil, false
	}
	snapshot := m.selectionSnapshot()
	if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok {
		if item.Label == "Reports directory" && m.exportAssets.ReportsDirCustomized {
			return commands.SetExportReportsDir(m.client, ""), true
		}
		if item.Label == "ICS export directory" && m.exportAssets.ICSDirCustomized {
			return commands.SetExportICSDir(m.client, ""), true
		}
		if item.Resettable {
			return commands.ResetExportTemplate(m.client, item.ReportKind, item.AssetKind), true
		}
	}
	return nil, false
}

func (m Model) handleInputStartFocusFromSelection() (tea.Model, tea.Cmd) {
	snapshot := m.selectionSnapshot()
	if m.view == ViewSessionHistory && (m.timer == nil || m.timer.State == "idle") {
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneSessions)
		if rawIdx < 0 || rawIdx >= len(m.sessionHistory) {
			return m, nil
		}
		issueID := m.sessionHistory[rawIdx].IssueID
		meta := selectionpkg.IssueMetaByID(snapshot, issueID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, commands.StartFocusSession(m.client, meta.RepoID, meta.StreamID, issueID)
	}
	if m.view == ViewSessionActive {
		if m.timer == nil || m.timer.State == "idle" {
			if m.context == nil || m.context.RepoID == nil || m.context.StreamID == nil || m.context.IssueID == nil {
				return m, m.setStatus("No active issue in context", true)
			}
			meta := selectionpkg.ActiveIssue(snapshot)
			if meta == nil {
				return m, m.setStatus("Active issue metadata unavailable", true)
			}
			return m, commands.StartFocusSession(m.client, *m.context.RepoID, *m.context.StreamID, *m.context.IssueID)
		}
		return m, nil
	}
	if m.pane != PaneIssues {
		return m, nil
	}
	if m.view == ViewDefault {
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneIssues)
		issues := selectionpkg.DefaultScopedIssues(snapshot)
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		return m, commands.StartFocusSession(m.client, issue.RepoID, issue.StreamID, issue.ID)
	}
	if m.view == ViewDaily {
		rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneIssues)
		issues := selectionpkg.DailyScopedIssues(snapshot)
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		meta := selectionpkg.IssueMetaByID(snapshot, issue.ID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, commands.StartFocusSession(m.client, meta.RepoID, issue.StreamID, issue.ID)
	}
	rawIdx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneIssues)
	if rawIdx < 0 || rawIdx >= len(m.issues) {
		return m, nil
	}
	if m.context == nil || m.context.RepoID == nil {
		return m, m.setStatus("No repo in context for selected issue", true)
	}
	issue := m.issues[rawIdx]
	return m, commands.StartFocusSession(m.client, *m.context.RepoID, issue.StreamID, issue.ID)
}
