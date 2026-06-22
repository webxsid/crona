package model

import (
	sharedposthog "crona/shared/posthog"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	commands "crona/tui/internal/tui/commands"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	helperpkg "crona/tui/internal/tui/helpers"
	inputpkg "crona/tui/internal/tui/input"
	selectionpkg "crona/tui/internal/tui/selection"
	wellbeingview "crona/tui/internal/tui/views/wellbeing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type inputScope int

const (
	scopeShell inputScope = iota
	scopeDialog
	scopeSessionDetailOverlay
	scopeSessionContextOverlay
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
		logger.Infof(
			"tui key msg: key=%q view=%s pane=%s dialog=%q help=%t session_detail=%t session_context=%t",
			key.String(),
			m.view,
			m.pane,
			m.dialog,
			m.helpOpen,
			m.sessionDetailOpen,
			m.sessionContextOpen,
		)
		touch := m.timerActivityTouchCmd(time.Now())
		if key.String() == "ctrl+c" {
			logger.Infof("tui key route: quit")
			m.stopEventStream()
			return m, tea.Quit
		}
		switch m.activeInputScope() {
		case scopeDialog:
			logger.Infof("tui key route: dialog")
			next, cmd := m.updateDialog(key)
			if model, ok := next.(Model); ok {
				logger.Infof(
					"tui key route result: dialog next_view=%s next_pane=%s next_dialog=%q",
					model.view,
					model.pane,
					model.dialog,
				)
			}
			return next, batchCmds(touch, cmd)
		case scopeSessionDetailOverlay:
			logger.Infof("tui key route: session_detail_overlay")
			next, cmd := m.updateSessionDetailOverlay(key)
			return next, batchCmds(touch, cmd)
		case scopeSessionContextOverlay:
			logger.Infof("tui key route: session_context_overlay")
			next, cmd := m.updateSessionContextOverlay(key)
			return next, batchCmds(touch, cmd)
		case scopeShell:
			logger.Infof("tui key route: input")
			state, cmd := inputpkg.Handle(m.inputState(), key, m.inputDeps())
			logger.Infof(
				"tui key route result: input next_view=%s next_pane=%s dialog=%q",
				state.ActiveView,
				state.ActivePane,
				state.Dialog,
			)
			return m.applyInputState(state), batchCmds(touch, cmd)
		}
	}
	switch msg := msg.(type) {
	case commands.IssueActionPreflightClearMsg:
		switch msg.Mode {
		case commands.IssueActionModeManual:
			return m.openManualSessionFromIssue(
				msg.Target.RepoID,
				msg.Target.StreamID,
				msg.Target.IssueID,
			), nil
		default:
			return m.openFocusSessionFromIssue(
				msg.Target.RepoID,
				msg.Target.StreamID,
				msg.Target.IssueID,
			), nil
		}
	case commands.MomentumDetailLoadedMsg:
		if msg.Detail == nil {
			return m, nil
		}
		return m.openMomentumDetailDialog(*msg.Detail), nil
	case commands.MomentumDetailFailedMsg:
		return m, m.setStatus(msg.Err.Error(), true)
	}
	state, cmd, handled := dispatchpkg.HandleMessage(
		m.dispatchMessageState(),
		msg,
		m.dispatchMessageDeps(),
	)
	if handled {
		return m.applyDispatchMessageState(state), cmd
	}
	return m, nil
}

func (m Model) activeInputScope() inputScope {
	switch {
	case m.dialog != "":
		return scopeDialog
	case m.sessionDetailOpen:
		return scopeSessionDetailOverlay
	case m.sessionContextOpen:
		return scopeSessionContextOverlay
	default:
		return scopeShell
	}
}

func (m *Model) timerActivityTouchCmd(now time.Time) tea.Cmd {
	if m.client == nil ||
		m.timer == nil ||
		m.timer.State == "idle" ||
		m.timer.State == "ready" ||
		m.timer.State == "expired" {
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
		HabitStreakDefs:         m.habitStreakDefs,
		MomentumCards:           m.momentumCards,
		AllIssues:               m.allIssues,
		DueHabits:               m.dueHabits,
		DailySummary:            m.dailySummary,
		DailyPlan:               m.dailyPlan,
		RollupStartDate:         m.currentRollupStartDate(),
		RollupEndDate:           m.currentRollupEndDate(),
		MomentumDate:            m.currentMomentumDate(),
		MomentumWindowDays:      m.currentMomentumWindowDays(),
		MomentumTab:             string(m.currentMomentumTab()),
		MomentumHistoryY:        m.momentumHistoryCursor,
		WellbeingDate:           m.wellbeingDate,
		WellbeingWindowDays:     m.currentWellbeingWindowDays(),
		DailyCheckIn:            m.dailyCheckIn,
		MetricsRange:            m.metricsRange,
		MetricsRollup:           m.metricsRollup,
		RollupMetricsRange:      m.rollupMetricsRange,
		RollupMetricsRollup:     m.rollupMetricsRollup,
		MomentumMetricsRange:    m.momentumMetricsRange,
		MomentumMetricsRollup:   m.momentumMetricsRollup,
		Streaks:                 m.streaks,
		DailyStreaks:            m.dailyStreaks,
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
		Ops:                     m.ops,
		Context:                 m.context,
		Timer:                   m.timer,
		Health:                  m.health,
		AlertStatus:             m.alertStatus,
		AlertReminders:          m.alertReminders,
		UpdateStatus:            m.updateStatus,
		UpdateChecking:          m.updateChecking,
		Settings:                m.settings,
		KernelInfo:              m.kernelInfo,
		Elapsed:                 m.elapsed,
		TimerTickSeq:            m.timerTickSeq,
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
	prevTimerExpired := m.timer != nil && m.timer.State == "expired"
	m.width = state.Width
	m.height = state.Height
	m.view = state.View
	m.pane = state.Pane
	m.cursor = state.Cursor
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	m.filters = state.Filters
	m.repos = state.Repos
	m.streams = state.Streams
	m.issues = state.Issues
	m.habits = state.Habits
	m.allHabits = state.AllHabits
	m.habitStreakDefs = state.HabitStreakDefs
	m.momentumCards = state.MomentumCards
	m.allIssues = state.AllIssues
	m.dueHabits = state.DueHabits
	m.dailySummary = state.DailySummary
	m.dailyPlan = state.DailyPlan
	m.dashboardDate = state.DashboardDate
	m.rollupStartDate = state.RollupStartDate
	m.rollupEndDate = state.RollupEndDate
	m.momentumDate = state.MomentumDate
	m.momentumWindowDays = state.MomentumWindowDays
	m.momentumTab = MomentumTab(state.MomentumTab)
	m.momentumHistoryCursor = state.MomentumHistoryY
	m.wellbeingDate = state.WellbeingDate
	m.wellbeingWindowDays = state.WellbeingWindowDays
	m.dailyCheckIn = state.DailyCheckIn
	m.metricsRange = state.MetricsRange
	m.metricsRollup = state.MetricsRollup
	m.rollupMetricsRange = state.RollupMetricsRange
	m.rollupMetricsRollup = state.RollupMetricsRollup
	m.momentumMetricsRange = state.MomentumMetricsRange
	m.momentumMetricsRollup = state.MomentumMetricsRollup
	m.streaks = state.Streaks
	m.dailyStreaks = state.DailyStreaks
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
	m.ops = state.Ops
	m.context = state.Context
	m.timer = state.Timer
	m.health = state.Health
	m.alertStatus = state.AlertStatus
	m.alertReminders = state.AlertReminders
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.settings = state.Settings
	m.kernelInfo = state.KernelInfo
	m.elapsed = state.Elapsed
	m.timerTickSeq = state.TimerTickSeq
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
	if m.dialog == "" && m.timer != nil && m.timer.State == "expired" && !prevTimerExpired {
		m = m.openHardLimitExpiredDialog(m.terminalSessionTitle())
	}
	if m.dialog == "hard_limit_expired" || m.dialog == "hard_limit_extend" {
		m = m.withDialogState(m.hydrateHardLimitDialogStateFromTimer(m.dialogState()))
	}
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
			snapshot := next.selectionSnapshot()
			max := len(selectionpkg.FilteredIndices(snapshot, pane))
			if max == 0 {
				next.cursor[pane] = 0
			} else if next.cursor[pane] >= max {
				next.cursor[pane] = max - 1
			}
			*state = next.dispatchMessageState()
		},
		FilteredCursorForRawIndex: func(state *dispatchpkg.MessageState, pane Pane, rawIdx int) int {
			next := m.applyDispatchMessageState(*state)
			snapshot := next.selectionSnapshot()
			return selectionpkg.FilteredCursorForRawIndex(snapshot, pane, rawIdx)
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
		OpenSupportBundleDialog: func(state *dispatchpkg.MessageState, path string, sizeBytes int64, windowLabel string) {
			next := m.applyDispatchMessageState(*state)
			next = next.openSupportBundleDialog(path, sizeBytes, windowLabel)
			*state = next.dispatchMessageState()
		},
		OpenOnboardingDialog: func(state *dispatchpkg.MessageState) {
			next := m.applyDispatchMessageState(*state)
			next = next.withDialogState(next.dialogSnapshot().OpenOnboarding())
			*state = next.dispatchMessageState()
		},
		AnchorWellbeingScroll: func(state *dispatchpkg.MessageState, pane Pane) {
			next := m.applyDispatchMessageState(*state)
			snapshot := next.selectionSnapshot()
			count := wellbeingview.PaneLineCount(
				next.viewContentState(
					next.mainContentWidth(),
					next.contentHeight(),
					snapshot,
					selectionpkg.ActiveIssue(snapshot),
				),
				string(pane),
			)
			if count > 0 {
				if current := next.cursor[pane]; current <= 0 {
					next.cursor[pane] = count - 1
				}
			}
			*state = next.dispatchMessageState()
		},
		CurrentDashboardDate: func(state dispatchpkg.MessageState) string {
			return m.applyDispatchMessageState(state).currentDashboardDate()
		},
		CurrentMomentumDate: func(state dispatchpkg.MessageState) string {
			return m.applyDispatchMessageState(state).currentMomentumDate()
		},
		CurrentWellbeingDate: func(state dispatchpkg.MessageState) string {
			return m.applyDispatchMessageState(state).currentWellbeingDate()
		},
		LoadRollupSummaries: func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		LoadMomentumRange: func(date string, windowDays int) tea.Cmd {
			return commands.LoadMomentumRange(m.client, date, windowDays)
		},
		LoadMomentumFocus: func(date string, windowDays int) tea.Cmd {
			return commands.LoadMomentumFocusWindow(m.client, date, windowDays)
		},
		LoadRepos:     func() tea.Cmd { return commands.LoadRepos(m.client) },
		LoadAllIssues: func() tea.Cmd { return commands.LoadAllIssues(m.client) },
		LoadStreams:   func(id int64) tea.Cmd { return commands.LoadStreams(m.client, id) },
		LoadIssues:    func(id int64) tea.Cmd { return commands.LoadIssues(m.client, id) },
		LoadHabits:    func(id int64) tea.Cmd { return commands.LoadHabits(m.client, id) },
		LoadHabitStreakDefinitions: func() tea.Cmd {
			return commands.LoadHabitStreakDefinitions(m.client)
		},
		LoadDueHabits:    func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadDailySummary: func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDailyStreaks: func(date string) tea.Cmd { return commands.LoadDailyStreaks(m.client, date) },
		LoadWellbeing: func(date string, windowDays int) tea.Cmd {
			return commands.LoadWellbeingWindow(m.client, date, windowDays)
		},
		LoadDailyPlan:     func(date string) tea.Cmd { return commands.LoadDailyPlan(m.client, date) },
		LoadExportAssets:  func() tea.Cmd { return commands.LoadExportAssets(m.client) },
		LoadExportReports: func() tea.Cmd { return commands.LoadExportReports(m.client) },
		LoadIssueSessions: func(id int64) tea.Cmd { return commands.LoadIssueSessions(m.client, id) },
		LoadHabitHistory: func(ctx *api.ActiveContext, selectedID *int64) tea.Cmd {
			return commands.LoadHabitHistory(m.client, ctx, selectedID)
		},
		LoadSessionHistoryFor200: func(state dispatchpkg.MessageState) tea.Cmd {
			next := m.applyDispatchMessageState(state)
			return commands.LoadSessionHistory(
				m.client,
				helperpkg.SessionHistoryScopeIssueID(next.timer),
				200,
			)
		},
		LoadSessionDetail:  func(id string) tea.Cmd { return commands.LoadSessionDetail(m.client, id) },
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
		ReportHandledError: func(err error, operation string) tea.Cmd {
			if m.telemetry == nil {
				return nil
			}
			return func() tea.Msg {
				_ = m.telemetry.ReportError("handled", err, sharedposthog.Properties{
					"entrypoint": "tui",
					"operation":  operation,
				})
				return nil
			}
		},
		HealthTickAfter: commands.HealthTickAfter,
		TickAfter:       commands.TickAfter,
		WaitForEvent:    func() tea.Cmd { return commands.WaitForEvent(eventChannel) },
		HandleKernelEvent: func(state dispatchpkg.MessageState, event api.KernelEvent) (dispatchpkg.MessageState, tea.Cmd) {
			next := m.applyDispatchMessageState(state)
			nextModel, cmd := next.handleKernelEvent(event)
			return nextModel.dispatchMessageState(), cmd
		},
		CloseEventStop: func() { m.stopEventStream() },
	}
}
