package app

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		close(m.eventStop)
		return m, tea.Quit
	case "K":
		return m, cmdShutdownKernel(m.client)
	case "f6":
		if m.isDevMode() {
			return m, cmdSeedDevData(m.client)
		}
	case "f7":
		if m.isDevMode() {
			return m, cmdClearDevData(m.client)
		}
	case "]":
		if m.timer != nil && m.timer.State != "idle" {
			if m.view == ViewSessionActive {
				m.view = ViewScratch
			} else {
				m.view = ViewSessionActive
			}
			m.pane = viewDefaultPane[m.view]
			return m, nil
		}
		m.view = nextView(m.view, 1)
		m.pane = viewDefaultPane[m.view]
		return m, nil
	case "[":
		if m.timer != nil && m.timer.State != "idle" {
			if m.view == ViewScratch {
				m.view = ViewSessionActive
			} else {
				m.view = ViewScratch
			}
			m.pane = viewDefaultPane[m.view]
			return m, nil
		}
		m.view = nextView(m.view, -1)
		m.pane = viewDefaultPane[m.view]
		return m, nil
	case "tab":
		if m.filterEditing {
			return m, nil
		}
		if m.view == ViewDefault && m.pane == PaneIssues {
			m.cycleDefaultIssueSection(1)
			return m, nil
		}
		m.pane = nextPane(m.view, m.pane, 1)
		return m, nil
	case "shift+tab":
		if m.filterEditing {
			return m, nil
		}
		if m.view == ViewDefault && m.pane == PaneIssues {
			m.cycleDefaultIssueSection(-1)
			return m, nil
		}
		m.pane = nextPane(m.view, m.pane, -1)
		return m, nil
	case "R":
		logger.Info("Kernel restart requested")
		return m, nil
	}

	switch key {
	case "j", "down":
		max := m.listLen(m.pane)
		if max > 0 && m.cursor[m.pane] < max-1 {
			m.cursor[m.pane]++
		}
		return m, nil
	case "k", "up":
		if m.cursor[m.pane] > 0 {
			m.cursor[m.pane]--
		}
		return m, nil
	}

	if m.view == ViewMeta {
		switch key {
		case "1":
			m.pane = PaneRepos
			return m, nil
		case "2":
			m.pane = PaneStreams
			return m, nil
		case "3":
			m.pane = PaneIssues
			return m, nil
		case "4":
			m.pane = PaneHabits
			return m, nil
		}
	}
	if m.view == ViewDefault && key == "1" {
		m.pane = PaneIssues
		m.setDefaultIssueSection(DefaultIssueSectionOpen)
		return m, nil
	}
	if m.view == ViewDefault && key == "2" {
		m.pane = PaneIssues
		m.setDefaultIssueSection(DefaultIssueSectionCompleted)
		return m, nil
	}
	if m.view == ViewDaily {
		switch key {
		case "1":
			m.pane = PaneIssues
			return m, nil
		case "2":
			m.pane = PaneHabits
			return m, nil
		}
	}
	if m.view == ViewScratch && key == "1" {
		m.pane = PaneScratchpads
		return m, nil
	}
	if m.view == ViewSettings {
		switch key {
		case "1":
			m.pane = PaneSettings
			return m, nil
		case "left", "h":
			return m.adjustSelectedSetting(-1)
		case "right", "l", "enter", " ":
			return m.adjustSelectedSetting(1)
		}
	}

	switch key {
	case "f":
		return m.startFocusFromSelection()
	case "s":
		if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch) {
			return m, m.setStatus("End or stash the active session before changing issue status", true)
		}
		if _, _, status, _, ok := m.selectedIssue(); ok {
			return m.openIssueStatusDialog(status), nil
		}
		return m, nil
	case "A":
		return m.abandonSelectedIssue()
	case "y":
		return m.toggleSelectedIssueToday()
	case "D":
		return m.openSelectedIssueTodoDateDialog(), nil
	case ",":
		if m.view == ViewDaily {
			m.dashboardDate = shiftISODate(m.currentDashboardDate(), -1)
			return m, tea.Batch(loadDailySummary(m.client, m.dashboardDate), loadDueHabits(m.client, m.dashboardDate))
		}
		if m.view == ViewWellbeing {
			m.wellbeingDate = shiftISODate(m.currentWellbeingDate(), -1)
			return m, loadWellbeing(m.client, m.wellbeingDate)
		}
	case ".":
		if m.view == ViewDaily {
			m.dashboardDate = shiftISODate(m.currentDashboardDate(), 1)
			return m, tea.Batch(loadDailySummary(m.client, m.dashboardDate), loadDueHabits(m.client, m.dashboardDate))
		}
		if m.view == ViewWellbeing {
			m.wellbeingDate = shiftISODate(m.currentWellbeingDate(), 1)
			return m, loadWellbeing(m.client, m.wellbeingDate)
		}
	case "g":
		if m.view == ViewDaily {
			m.dashboardDate = ""
			return m, tea.Batch(loadDailySummary(m.client, ""), loadDueHabits(m.client, m.currentDashboardDate()))
		}
		if m.view == ViewWellbeing {
			m.wellbeingDate = ""
			return m, loadWellbeing(m.client, m.currentWellbeingDate())
		}
	case "c":
		if (m.view == ViewDefault && m.pane == PaneIssues) || m.view == ViewDaily {
			return m.openCheckoutContextDialog(), nil
		}
		return m.checkout()
	case "p":
		if m.view == ViewSessionActive && m.timer != nil && m.timer.State == "running" {
			return m, cmdPauseFocusSession(m.client)
		}
	case "r":
		if m.view == ViewSessionActive && m.timer != nil && m.timer.State == "paused" {
			return m, cmdResumeFocusSession(m.client)
		}
	case "x":
		if m.view == ViewSessionActive && m.timer != nil && m.timer.State != "idle" {
			return m.openSessionMessageDialog("end_session"), nil
		}
	case "z":
		if m.view == ViewSessionActive && m.timer != nil && m.timer.State != "idle" {
			return m.openSessionMessageDialog("stash_session"), nil
		}
	case "Z":
		if (m.timer == nil || m.timer.State == "idle") && m.dialog == "" {
			return m.openStashListDialog(), loadStashes(m.client)
		}
	case "+", "=":
		if m.pane == PaneOps {
			m.opsLimitPinned = true
			m.opsLimit += 10
			if m.opsLimit < 10 {
				m.opsLimit = 10
			}
			return m, loadOps(m.client, m.currentOpsLimit())
		}
	case "-":
		if m.pane == PaneOps {
			m.opsLimitPinned = true
			m.opsLimit -= 10
			if m.opsLimit < 10 {
				m.opsLimit = 10
			}
			m.clampFiltered(PaneOps)
			return m, loadOps(m.client, m.currentOpsLimit())
		}
	case "/":
		if m.pane != PaneOps && m.pane != PaneIssues && m.pane != PaneHabits && m.pane != PaneRepos && m.pane != PaneStreams && m.pane != PaneScratchpads && m.pane != PaneSessions {
			return m, nil
		}
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			return m, nil
		}
		m.startFilterEdit(m.pane)
		return m, nil
	case "?":
		m.helpOpen = true
		return m, nil
	case "a":
		if m.view == ViewWellbeing {
			return m.openCreateCheckInDialog(), nil
		}
		return m.handleCreateAction()
	case "e":
		if m.view == ViewDaily && m.pane == PaneHabits {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				return m.openHabitCompletionDialog(habit.ID, m.currentDashboardDate(), habit.DurationMinutes, habit.Notes), nil
			}
			return m, nil
		}
		if m.view == ViewWellbeing {
			if m.dailyCheckIn == nil {
				return m.openCreateCheckInDialog(), nil
			}
			return m.openEditCheckInDialog(), nil
		}
		if m.view == ViewSessionHistory && m.pane == PaneSessions {
			if entry, ok := m.selectedSessionHistoryEntry(); ok {
				m.sessionDetailOpen = true
				m.sessionDetailY = 0
				return m, loadSessionDetail(m.client, entry.ID)
			}
		}
		if m.dialog == "" {
			if next, ok := m.openSelectedEditDialog(); ok {
				return next, nil
			}
		}
	case "F":
		if m.view == ViewDaily && m.pane == PaneHabits {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				if habit.Status == sharedtypes.HabitCompletionStatusFailed {
					return m, cmdUncompleteHabit(m.client, habit.ID, m.currentDashboardDate())
				}
				return m, cmdSetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusFailed, habit.DurationMinutes, habit.Notes)
			}
			return m, nil
		}
	case "d":
		if m.view == ViewWellbeing {
			if m.dailyCheckIn == nil {
				return m, m.setStatus("No check-in to delete for this date", true)
			}
			return m.openConfirmDeleteEntity("checkin", m.currentWellbeingDate(), "this check-in"), nil
		}
		if m.pane == PaneScratchpads {
			rawIdx := m.filteredIndexAtCursor(PaneScratchpads)
			if rawIdx >= 0 && rawIdx < len(m.scratchpads) {
				return m.openConfirmDelete(m.scratchpads[rawIdx].ID), nil
			}
		}
		if m.dialog == "" {
			if next, ok := m.openSelectedDeleteDialog(); ok {
				return next, nil
			}
		}
	case "enter", "o":
		if m.view == ViewSessionHistory && m.pane == PaneSessions {
			if entry, ok := m.selectedSessionHistoryEntry(); ok {
				m.sessionDetailOpen = true
				m.sessionDetailY = 0
				return m, loadSessionDetail(m.client, entry.ID)
			}
			return m, nil
		}
		if m.dialog == "" {
			if next, ok := m.openSelectedViewDialog(); ok {
				return next, nil
			}
		}
		if m.pane == PaneScratchpads {
			rawIdx := m.filteredIndexAtCursor(PaneScratchpads)
			if rawIdx < 0 || rawIdx >= len(m.scratchpads) {
				return m, nil
			}
			m.scratchpadOpen = true
			m.setActiveScratchpadByIndex(rawIdx)
			return m, cmdOpenScratchpad(m.client, m.scratchpads, rawIdx)
		}
	}

	switch key {
	case "x", " ":
		if m.view == ViewDaily && m.pane == PaneHabits {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				if habit.Status == sharedtypes.HabitCompletionStatusCompleted {
					return m, cmdUncompleteHabit(m.client, habit.ID, m.currentDashboardDate())
				}
				return m, cmdSetHabitStatus(m.client, habit.ID, m.currentDashboardDate(), sharedtypes.HabitCompletionStatusCompleted, nil, nil)
			}
		}
	}

	return m, nil
}

func (m Model) adjustSelectedSetting(dir int) (tea.Model, tea.Cmd) {
	if m.view != ViewSettings || m.settings == nil {
		return m, nil
	}
	rawIdx := m.filteredIndexAtCursor(PaneSettings)
	if rawIdx < 0 {
		return m, nil
	}

	repoID := int64(0)
	if m.context != nil && m.context.RepoID != nil {
		repoID = *m.context.RepoID
	}
	streamID := int64(0)
	if m.context != nil && m.context.StreamID != nil {
		streamID = *m.context.StreamID
	}

	switch rawIdx {
	case 0:
		next := sharedtypes.TimerModeStructured
		if m.settings.TimerMode == sharedtypes.TimerModeStructured {
			next = sharedtypes.TimerModeStopwatch
		}
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyTimerMode, next, repoID, streamID, m.dashboardDate)
	case 1:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyBreaksEnabled, !m.settings.BreaksEnabled, repoID, streamID, m.dashboardDate)
	case 2:
		next := clampMin(m.settings.WorkDurationMinutes+dir*5, 5)
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyWorkDurationMinutes, next, repoID, streamID, m.dashboardDate)
	case 3:
		next := clampMin(m.settings.ShortBreakMinutes+dir, 1)
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyShortBreakMinutes, next, repoID, streamID, m.dashboardDate)
	case 4:
		next := clampMin(m.settings.LongBreakMinutes+dir*5, 5)
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyLongBreakMinutes, next, repoID, streamID, m.dashboardDate)
	case 5:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyLongBreakEnabled, !m.settings.LongBreakEnabled, repoID, streamID, m.dashboardDate)
	case 6:
		next := clampMin(m.settings.CyclesBeforeLongBreak+dir, 1)
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak, next, repoID, streamID, m.dashboardDate)
	case 7:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyAutoStartBreaks, !m.settings.AutoStartBreaks, repoID, streamID, m.dashboardDate)
	case 8:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyAutoStartWork, !m.settings.AutoStartWork, repoID, streamID, m.dashboardDate)
	case 9:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyRepoSort, nextRepoSort(m.settings.RepoSort, dir), repoID, streamID, m.dashboardDate)
	case 10:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyStreamSort, nextStreamSort(m.settings.StreamSort, dir), repoID, streamID, m.dashboardDate)
	case 11:
		return m, cmdPatchSetting(m.client, sharedtypes.CoreSettingsKeyIssueSort, nextIssueSort(m.settings.IssueSort, dir), repoID, streamID, m.dashboardDate)
	default:
		return m, nil
	}
}

func clampMin(value, min int) int {
	if value < min {
		return min
	}
	return value
}

func nextRepoSort(current sharedtypes.RepoSort, dir int) sharedtypes.RepoSort {
	options := []sharedtypes.RepoSort{
		sharedtypes.RepoSortAlphabeticalAsc,
		sharedtypes.RepoSortAlphabeticalDesc,
		sharedtypes.RepoSortChronologicalAsc,
		sharedtypes.RepoSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextStreamSort(current sharedtypes.StreamSort, dir int) sharedtypes.StreamSort {
	options := []sharedtypes.StreamSort{
		sharedtypes.StreamSortAlphabeticalAsc,
		sharedtypes.StreamSortAlphabeticalDesc,
		sharedtypes.StreamSortChronologicalAsc,
		sharedtypes.StreamSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextIssueSort(current sharedtypes.IssueSort, dir int) sharedtypes.IssueSort {
	options := []sharedtypes.IssueSort{
		sharedtypes.IssueSortPriority,
		sharedtypes.IssueSortDueDateAsc,
		sharedtypes.IssueSortDueDateDesc,
		sharedtypes.IssueSortAlphabeticalAsc,
		sharedtypes.IssueSortAlphabeticalDesc,
		sharedtypes.IssueSortChronologicalAsc,
		sharedtypes.IssueSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextIndex[T comparable](current T, options []T, dir int) int {
	if len(options) == 0 {
		return 0
	}
	index := 0
	for i, option := range options {
		if option == current {
			index = i
			break
		}
	}
	index += dir
	if index < 0 {
		index = len(options) - 1
	}
	if index >= len(options) {
		index = 0
	}
	return index
}

func (m Model) handleCreateAction() (tea.Model, tea.Cmd) {
	if m.view == ViewDefault && m.pane == PaneIssues {
		return m.openCreateIssueDefaultDialog(), nil
	}
	if m.view == ViewDaily {
		switch m.pane {
		case PaneIssues:
			return m.openCreateIssueDefaultDialog(), nil
		case PaneHabits:
			repoName := "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName := "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName), nil
		}
	}
	if m.view == ViewMeta {
		switch m.pane {
		case PaneRepos:
			return m.openCreateRepoDialog(), nil
		case PaneStreams:
			repoID, repoName, ok := m.selectedMetaRepo()
			if !ok {
				return m, m.setStatus("Select or checkout a repo first", true)
			}
			return m.openCreateStreamDialog(repoID, repoName), nil
		case PaneIssues:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if !ok {
				return m, m.setStatus("Select or checkout a stream first", true)
			}
			return m.openCreateIssueMetaDialog(streamID, streamName, repoName), nil
		case PaneHabits:
			streamID, streamName, repoName, ok := m.selectedMetaStream()
			if ok {
				return m.openCreateHabitDialog(streamID, streamName, repoName), nil
			}
			repoName = "-"
			if m.context != nil && m.context.RepoName != nil && *m.context.RepoName != "" {
				repoName = *m.context.RepoName
			}
			streamName = "-"
			if m.context != nil && m.context.StreamName != nil && *m.context.StreamName != "" {
				streamName = *m.context.StreamName
			}
			return m.openCreateHabitDialog(0, streamName, repoName), nil
		}
	}
	if m.pane == PaneScratchpads {
		return m.openCreateScratchpad(), nil
	}
	return m, nil
}

func (m Model) startFocusFromSelection() (tea.Model, tea.Cmd) {
	if m.view == ViewSessionHistory && (m.timer == nil || m.timer.State == "idle") {
		rawIdx := m.filteredIndexAtCursor(PaneSessions)
		if rawIdx < 0 || rawIdx >= len(m.sessionHistory) {
			return m, nil
		}
		issueID := m.sessionHistory[rawIdx].IssueID
		meta := m.issueMetaByID(issueID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, cmdStartFocusSession(m.client, meta.RepoID, meta.StreamID, issueID)
	}
	if m.view == ViewSessionActive {
		if m.timer == nil || m.timer.State == "idle" {
			if m.context == nil || m.context.RepoID == nil || m.context.StreamID == nil || m.context.IssueID == nil {
				return m, m.setStatus("No active issue in context", true)
			}
			meta := m.activeIssueWithMeta()
			if meta == nil {
				return m, m.setStatus("Active issue metadata unavailable", true)
			}
			return m, cmdStartFocusSession(m.client, *m.context.RepoID, *m.context.StreamID, *m.context.IssueID)
		}
		return m, nil
	}
	if m.pane != PaneIssues {
		return m, nil
	}
	if m.view == ViewDefault {
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		issues := m.defaultScopedIssues()
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		return m, cmdStartFocusSession(m.client, issue.RepoID, issue.StreamID, issue.ID)
	}
	if m.view == ViewDaily {
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		issues := m.dailyScopedIssues()
		if rawIdx < 0 || rawIdx >= len(issues) {
			return m, nil
		}
		issue := issues[rawIdx]
		meta := m.issueMetaByID(issue.ID)
		if meta == nil {
			return m, m.setStatus("Issue metadata unavailable", true)
		}
		return m, cmdStartFocusSession(m.client, meta.RepoID, issue.StreamID, issue.ID)
	}
	rawIdx := m.filteredIndexAtCursor(PaneIssues)
	if rawIdx < 0 || rawIdx >= len(m.issues) {
		return m, nil
	}
	if m.context == nil || m.context.RepoID == nil {
		return m, m.setStatus("No repo in context for selected issue", true)
	}
	issue := m.issues[rawIdx]
	return m, cmdStartFocusSession(m.client, *m.context.RepoID, issue.StreamID, issue.ID)
}
