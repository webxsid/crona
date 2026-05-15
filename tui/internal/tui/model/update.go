package model

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	filteringpkg "crona/tui/internal/tui/filtering"
	helperpkg "crona/tui/internal/tui/helpers"
	inputpkg "crona/tui/internal/tui/input"
	overlaypkg "crona/tui/internal/tui/overlays"
	selectionpkg "crona/tui/internal/tui/selection"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	if model, ok := next.(Model); ok {
		return model.withTerminalTitle(cmd)
	}
	return next, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		touch := m.timerActivityTouchCmd(time.Now())
		if m.dialog != "" {
			next, cmd := m.updateDialog(key)
			return next, batchCmds(touch, cmd)
		}
		if m.sessionDetailOpen {
			next, cmd := m.updateSessionDetailOverlay(key)
			return next, batchCmds(touch, cmd)
		}
		if m.sessionContextOpen {
			next, cmd := m.updateSessionContextOverlay(key)
			return next, batchCmds(touch, cmd)
		}
		if m.helpOpen {
			state, cmd := overlaypkg.HandleHelp(m.overlayState(), key)
			return m.applyOverlayState(state), batchCmds(touch, cmd)
		}
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			next, cmd := m.updateScratchpadPane(key)
			return next, batchCmds(touch, cmd)
		}
		if m.filterEditing {
			state, cmd := overlaypkg.HandleFilter(m.overlayState(), key, m.overlayDeps())
			if cmd != nil || state.FilterEditing != m.filterEditing {
				return m.applyOverlayState(state), batchCmds(touch, cmd)
			}
			next, cmd := filteringpkg.Update(m.filterState(), key, m.filterDeps())
			return m.applyFilterState(next), batchCmds(touch, cmd)
		}
		state, cmd := inputpkg.Handle(m.inputState(), key, m.inputDeps())
		return m.applyInputState(state), batchCmds(touch, cmd)
	}
	state, cmd, handled := dispatchpkg.HandleMessage(m.dispatchMessageState(), msg, m.dispatchMessageDeps())
	if handled {
		return m.applyDispatchMessageState(state), cmd
	}
	return m, nil
}

func (m *Model) timerActivityTouchCmd(now time.Time) tea.Cmd {
	if m.client == nil || m.timer == nil || m.timer.State == "idle" {
		return nil
	}
	if !m.lastTimerActivityTouch.IsZero() && now.Sub(m.lastTimerActivityTouch) < time.Minute {
		return nil
	}
	m.lastTimerActivityTouch = now
	return commands.TouchTimerActivity(m.client)
}

func batchCmds(cmds ...tea.Cmd) tea.Cmd {
	filtered := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return tea.Batch(filtered...)
}

func (m Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state, action, status := m.dialogSnapshot().Update(msg)
	if status != "" {
		state.ErrorMessage = status
		return m.withDialogState(state), nil
	}
	state.ErrorMessage = ""
	next := m.withDialogState(state)
	if action == nil {
		return next, nil
	}
	return next.handleDialogAction(next, *action)
}

func (m Model) filterState() filteringpkg.State {
	return filteringpkg.State{
		Filters:       m.filters,
		Cursor:        m.cursor,
		FilterEditing: m.filterEditing,
		FilterPane:    m.filterPane,
		FilterInput:   m.filterInput,
	}
}

func (m Model) applyFilterState(state filteringpkg.State) Model {
	m.filters = state.Filters
	m.cursor = state.Cursor
	m.filterEditing = state.FilterEditing
	m.filterPane = state.FilterPane
	m.filterInput = state.FilterInput
	return m
}

func (m Model) filterDeps() filteringpkg.Deps {
	return filteringpkg.Deps{
		ItemCount: func(state filteringpkg.State, pane Pane) int {
			next := m.applyFilterState(state)
			snapshot := next.selectionSnapshot()
			return len(selectionpkg.FilteredIndices(snapshot, pane))
		},
		Clamp: func(cursor map[Pane]int, pane Pane, max int) {
			if max == 0 {
				cursor[pane] = 0
				return
			}
			if cursor[pane] >= max {
				cursor[pane] = max - 1
			}
		},
	}
}


func (m Model) dispatchMessageState() dispatchpkg.MessageState {
	var selectedHabitHistoryID *int64
	if entry, ok := selectionpkg.SelectedHabitHistoryEntry(m.selectionSnapshot()); ok {
		selectedHabitHistoryID = &entry.ID
	}
	return dispatchpkg.MessageState{
		Width:                   m.width,
		Height:                  m.height,
		View:                    m.view,
		Pane:                    m.pane,
		DailyTaskSection:        m.dailyTaskSection,
		DashboardDate:           m.dashboardDate,
		Cursor:                  m.cursor,
		Filters:                 m.filters,
		Repos:                   m.repos,
		Streams:                 m.streams,
		Issues:                  m.issues,
		Habits:                  m.habits,
		AllHabits:               m.allHabits,
		AllIssues:               m.allIssues,
		DueHabits:               m.dueHabits,
		DailySummary:            m.dailySummary,
		DailyPlan:               m.dailyPlan,
		RollupStartDate:         m.currentRollupStartDate(),
		RollupEndDate:           m.currentRollupEndDate(),
		WellbeingDate:           m.wellbeingDate,
		DailyCheckIn:            m.dailyCheckIn,
		MetricsRange:            m.metricsRange,
		MetricsRollup:           m.metricsRollup,
		Streaks:                 m.streaks,
		DashboardWindow:         m.dashboardWindow,
		DailyFocusScore:         m.dailyFocusScore,
		WeeklyFocusScore:        m.weeklyFocusScore,
		RepoDistribution:        m.repoDistribution,
		StreamDistribution:      m.streamDistribution,
		IssueDistribution:       m.issueDistribution,
		SegmentDistribution:     m.segmentDistribution,
		GoalProgress:            m.goalProgress,
		ExportAssets:            m.exportAssets,
		ExportReports:           m.exportReports,
		IssueSessions:           m.issueSessions,
		SessionHistory:          m.sessionHistory,
		HabitHistory:            m.habitHistory,
		SelectedHabitHistoryID:  selectedHabitHistoryID,
		HabitHistoryTitle:       m.habitHistoryTitle,
		HabitHistoryMeta:        m.habitHistoryMeta,
		SessionDetail:           m.sessionDetail,
		SessionDetailOpen:       m.sessionDetailOpen,
		SessionDetailY:          m.sessionDetailY,
		SessionContextOpen:      m.sessionContextOpen,
		SessionContextY:         m.sessionContextY,
		Scratchpads:             m.scratchpads,
		Stashes:                 m.stashes,
		DialogStashCursor:       m.dialogStashCursor,
		Ops:                     m.ops,
		Context:                 m.context,
		Timer:                   m.timer,
		Health:                  m.health,
		AlertStatus:             m.alertStatus,
		AlertReminders:          m.alertReminders,
		UpdateStatus:            m.updateStatus,
		UpdateChecking:          m.updateChecking,
		UpdateInstalling:        m.updateInstalling,
		UpdateInstallError:      m.updateInstallError,
		Settings:                m.settings,
		KernelInfo:              m.kernelInfo,
		Elapsed:                 m.elapsed,
		TimerTickSeq:            m.timerTickSeq,
		ScratchpadOpen:          m.scratchpadOpen,
		ScratchpadMeta:          m.scratchpadMeta,
		ScratchpadFilePath:      m.scratchpadFilePath,
		ScratchpadRendered:      m.scratchpadRendered,
		StatusMsg:               m.statusMsg,
		StatusSeq:               m.statusSeq,
		StatusErr:               m.statusErr,
		Dialog:                  m.dialog,
		DialogErrorMessage:      m.dialogErrorMessage,
		DialogDeleteID:          m.dialogDeleteID,
		DialogRepoID:            m.dialogRepoID,
		DialogStreamID:          m.dialogStreamID,
		DialogIssueID:           m.dialogIssueID,
		DialogChoiceItems:       m.dialogChoiceItems,
		DialogChoiceValues:      m.dialogChoiceValues,
		DialogChoiceDetails:     m.dialogChoiceDetails,
		DialogChoiceCursor:      m.dialogChoiceCursor,
		DialogProcessing:        m.dialogProcessing,
		DialogProcessingLabel:   m.dialogProcessingLabel,
		DialogViewTitle:         m.dialogViewTitle,
		DialogViewName:          m.dialogViewName,
		DialogViewMeta:          m.dialogViewMeta,
		DialogViewBody:          m.dialogViewBody,
		DialogViewPath:          m.dialogViewPath,
		DialogSupportBundlePath: m.dialogSupportBundlePath,
		OpsLimit:                m.opsLimit,
		OpsLimitPinned:          m.opsLimitPinned,
	}
}

func (m Model) applyDispatchMessageState(state dispatchpkg.MessageState) Model {
	m.width = state.Width
	m.height = state.Height
	m.view = state.View
	m.pane = state.Pane
	m.cursor = state.Cursor
	m.filters = state.Filters
	m.repos = state.Repos
	m.streams = state.Streams
	m.issues = state.Issues
	m.habits = state.Habits
	m.allHabits = state.AllHabits
	m.allIssues = state.AllIssues
	m.dueHabits = state.DueHabits
	m.dailySummary = state.DailySummary
	m.dailyPlan = state.DailyPlan
	m.dashboardDate = state.DashboardDate
	m.rollupStartDate = state.RollupStartDate
	m.rollupEndDate = state.RollupEndDate
	m.wellbeingDate = state.WellbeingDate
	m.dailyCheckIn = state.DailyCheckIn
	m.metricsRange = state.MetricsRange
	m.metricsRollup = state.MetricsRollup
	m.streaks = state.Streaks
	m.dashboardWindow = state.DashboardWindow
	m.dailyFocusScore = state.DailyFocusScore
	m.weeklyFocusScore = state.WeeklyFocusScore
	m.repoDistribution = state.RepoDistribution
	m.streamDistribution = state.StreamDistribution
	m.issueDistribution = state.IssueDistribution
	m.segmentDistribution = state.SegmentDistribution
	m.goalProgress = state.GoalProgress
	m.exportAssets = state.ExportAssets
	m.exportReports = state.ExportReports
	m.issueSessions = state.IssueSessions
	m.sessionHistory = state.SessionHistory
	m.habitHistory = state.HabitHistory
	m.habitHistoryTitle = state.HabitHistoryTitle
	m.habitHistoryMeta = state.HabitHistoryMeta
	m.sessionDetail = state.SessionDetail
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionContextOpen = state.SessionContextOpen
	m.sessionContextY = state.SessionContextY
	m.scratchpads = state.Scratchpads
	m.stashes = state.Stashes
	m.dialogStashCursor = state.DialogStashCursor
	m.ops = state.Ops
	m.context = state.Context
	m.timer = state.Timer
	m.health = state.Health
	m.alertStatus = state.AlertStatus
	m.alertReminders = state.AlertReminders
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.updateInstalling = state.UpdateInstalling
	m.updateInstallError = state.UpdateInstallError
	m.settings = state.Settings
	m.kernelInfo = state.KernelInfo
	m.elapsed = state.Elapsed
	m.timerTickSeq = state.TimerTickSeq
	m.scratchpadOpen = state.ScratchpadOpen
	m.scratchpadMeta = state.ScratchpadMeta
	m.scratchpadFilePath = state.ScratchpadFilePath
	m.scratchpadRendered = state.ScratchpadRendered
	m.statusMsg = state.StatusMsg
	m.statusSeq = state.StatusSeq
	m.statusErr = state.StatusErr
	m.dialog = state.Dialog
	m.dialogErrorMessage = state.DialogErrorMessage
	m.dialogDeleteID = state.DialogDeleteID
	m.dialogRepoID = state.DialogRepoID
	m.dialogStreamID = state.DialogStreamID
	m.dialogIssueID = state.DialogIssueID
	m.dialogChoiceItems = state.DialogChoiceItems
	m.dialogChoiceValues = state.DialogChoiceValues
	m.dialogChoiceDetails = state.DialogChoiceDetails
	m.dialogChoiceCursor = state.DialogChoiceCursor
	m.dialogProcessing = state.DialogProcessing
	m.dialogProcessingLabel = state.DialogProcessingLabel
	m.dialogViewTitle = state.DialogViewTitle
	m.dialogViewName = state.DialogViewName
	m.dialogViewMeta = state.DialogViewMeta
	m.dialogViewBody = state.DialogViewBody
	m.dialogViewPath = state.DialogViewPath
	m.dialogSupportBundlePath = state.DialogSupportBundlePath
	m.opsLimit = state.OpsLimit
	m.opsLimitPinned = state.OpsLimitPinned
	return m
}

func (m Model) dispatchMessageDeps() dispatchpkg.MessageDeps {
	return dispatchpkg.MessageDeps{
		DefaultOpsLimit: func(state dispatchpkg.MessageState) int {
			next := m.applyDispatchMessageState(state)
			return next.defaultOpsLimit()
		},
		CurrentOpsLimit: func(state dispatchpkg.MessageState) int {
			next := m.applyDispatchMessageState(state)
			return next.currentOpsLimit()
		},
		ClampFiltered: func(state *dispatchpkg.MessageState, pane Pane) {
			next := m.applyDispatchMessageState(*state)
			filterState := next.filterState()
			deps := next.filterDeps()
			if deps.Clamp != nil && deps.ItemCount != nil {
				deps.Clamp(filterState.Cursor, pane, deps.ItemCount(filterState, pane))
			}
			next = next.applyFilterState(filterState)
			*state = next.dispatchMessageState()
		},
		SyncScratchpadViewport: func(state *dispatchpkg.MessageState) {
			next := m.applyDispatchMessageState(*state)
			next.scratchpadViewport = helperpkg.SyncScratchpadViewport(next.scratchpadViewport, next.mainContentWidth(), next.contentHeight(), next.scratchpadRendered)
			*state = next.dispatchMessageState()
		},
		ScratchpadTabIndexByID: func(state *dispatchpkg.MessageState, id string) int {
			next := m.applyDispatchMessageState(*state)
			return helperpkg.ScratchpadTabIndexByID(next.scratchpads, id)
		},
		FilteredCursorForRawIndex: func(state *dispatchpkg.MessageState, pane Pane, rawIdx int) int {
			next := m.applyDispatchMessageState(*state)
			snapshot := next.selectionSnapshot()
			return selectionpkg.FilteredCursorForRawIndex(snapshot, pane, rawIdx)
		},
		SetActiveScratchpadByIndex: func(state *dispatchpkg.MessageState, idx int) {
			next := m.applyDispatchMessageState(*state)
			next.scratchpadMeta = helperpkg.ScratchpadMetaAt(next.scratchpads, idx)
			*state = next.dispatchMessageState()
		},
		SetStatus: func(state *dispatchpkg.MessageState, message string, isErr bool) tea.Cmd {
			next := m.applyDispatchMessageState(*state)
			cmd := next.setStatus(message, isErr)
			*state = next.dispatchMessageState()
			return cmd
		},
		OpenViewEntityDialog: func(state *dispatchpkg.MessageState, title, name, meta, body string) {
			next := m.applyDispatchMessageState(*state)
			next = next.openViewEntityDialog(title, name, meta, body)
			*state = next.dispatchMessageState()
		},
		OpenStashConflictDialog: func(state *dispatchpkg.MessageState, conflict api.StashConflict, repoID, streamID, issueID int64) {
			next := m.applyDispatchMessageState(*state)
			next = next.openStashConflictDialog(conflict, repoID, streamID, issueID)
			*state = next.dispatchMessageState()
		},
		OpenSupportBundleDialog: func(state *dispatchpkg.MessageState, path string, sizeBytes int64, windowLabel string) {
			next := m.applyDispatchMessageState(*state)
			next = next.openSupportBundleDialog(path, sizeBytes, windowLabel)
			*state = next.dispatchMessageState()
		},
		EnterScratchpadPane: func(state *dispatchpkg.MessageState, msg commands.OpenScratchpadMsg) {
			next := m.applyDispatchMessageState(*state)
			next.scratchpadOpen = true
			next.scratchpadMeta = helperpkg.ScratchpadMetaAt([]api.ScratchPad{{
				ID:           msg.Meta.ID,
				Name:         msg.Meta.Name,
				Path:         msg.Meta.Path,
				Pinned:       msg.Meta.Pinned,
				LastOpenedAt: msg.Meta.LastOpenedAt,
			}}, 0)
			next.scratchpadFilePath = msg.FilePath
			next.scratchpadRendered = msg.Content
			next.scratchpadViewport = helperpkg.SyncScratchpadViewport(next.scratchpadViewport, next.mainContentWidth(), next.contentHeight(), next.scratchpadRendered)
			next.scratchpadViewport.GotoTop()
			*state = next.dispatchMessageState()
		},
		SetScratchpadContent: func(state *dispatchpkg.MessageState, rendered, filePath string) {
			next := m.applyDispatchMessageState(*state)
			next.scratchpadFilePath = filePath
			next.scratchpadRendered = rendered
			next.scratchpadViewport.SetContent(rendered)
			*state = next.dispatchMessageState()
		},
		CurrentDashboardDate: func(state dispatchpkg.MessageState) string {
			return m.applyDispatchMessageState(state).currentDashboardDate()
		},
		CurrentWellbeingDate: func(state dispatchpkg.MessageState) string {
			return m.applyDispatchMessageState(state).currentWellbeingDate()
		},
		LoadRollupSummaries: func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		LoadRepos:           func() tea.Cmd { return commands.LoadRepos(m.client) },
		LoadAllIssues:       func() tea.Cmd { return commands.LoadAllIssues(m.client) },
		LoadStreams:         func(id int64) tea.Cmd { return commands.LoadStreams(m.client, id) },
		LoadIssues:          func(id int64) tea.Cmd { return commands.LoadIssues(m.client, id) },
		LoadHabits:          func(id int64) tea.Cmd { return commands.LoadHabits(m.client, id) },
		LoadDueHabits:       func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadDailySummary:    func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadWellbeing:       func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		LoadDailyPlan:       func(date string) tea.Cmd { return commands.LoadDailyPlan(m.client, date) },
		LoadExportAssets:    func() tea.Cmd { return commands.LoadExportAssets(m.client) },
		LoadExportReports:   func() tea.Cmd { return commands.LoadExportReports(m.client) },
		LoadIssueSessions:   func(id int64) tea.Cmd { return commands.LoadIssueSessions(m.client, id) },
		LoadHabitHistory: func(ctx *api.ActiveContext, selectedID *int64) tea.Cmd {
			return commands.LoadHabitHistory(m.client, ctx, selectedID)
		},
		LoadSessionHistoryFor200: func(state dispatchpkg.MessageState) tea.Cmd {
			next := m.applyDispatchMessageState(state)
			return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(next.timer), 200)
		},
		LoadSessionDetail:  func(id string) tea.Cmd { return commands.LoadSessionDetail(m.client, id) },
		LoadScratchpads:    func() tea.Cmd { return commands.LoadScratchpads(m.client) },
		LoadStashes:        func() tea.Cmd { return commands.LoadStashes(m.client) },
		LoadOps:            func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		LoadContext:        func() tea.Cmd { return commands.LoadContext(m.client) },
		LoadTimer:          func() tea.Cmd { return commands.LoadTimer(m.client) },
		LoadHealth:         func() tea.Cmd { return commands.LoadHealth(m.client) },
		LoadAlertStatus:    func() tea.Cmd { return commands.LoadAlertStatus(m.client) },
		LoadAlertReminders: func() tea.Cmd { return commands.LoadAlertReminders(m.client) },
		LoadUpdateStatus:   func() tea.Cmd { return commands.LoadUpdateStatus(m.client) },
		LoadSettings:       func() tea.Cmd { return commands.LoadSettings(m.client) },
		LoadKernelInfo:     func() tea.Cmd { return commands.LoadKernelInfo(m.client) },
		NotifyAlert:        func(input sharedtypes.AlertRequest) tea.Cmd { return commands.NotifyAlert(m.client, input) },
		HealthTickAfter:    commands.HealthTickAfter,
		TickAfter:          commands.TickAfter,
		WaitForEvent:       func() tea.Cmd { return commands.WaitForEvent(eventChannel) },
		HandleKernelEvent: func(state dispatchpkg.MessageState, event api.KernelEvent) (dispatchpkg.MessageState, tea.Cmd) {
			next := m.applyDispatchMessageState(state)
			nextModel, cmd := next.handleKernelEvent(event)
			return nextModel.dispatchMessageState(), cmd
		},
		CloseEventStop: func() { m.stopEventStream() },
	}
}
