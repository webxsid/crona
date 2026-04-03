package app

import (
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	dialogstate "crona/tui/internal/tui/dialogstate"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	filteringpkg "crona/tui/internal/tui/filtering"
	helperpkg "crona/tui/internal/tui/helpers"
	inputpkg "crona/tui/internal/tui/input"
	overlaypkg "crona/tui/internal/tui/overlays"
	selectionpkg "crona/tui/internal/tui/selection"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.updateInstallPhase != "" {
			return m.updateInstallScreen(key)
		}
		if m.dialog != "" {
			return m.updateDialog(key)
		}
		if m.sessionDetailOpen {
			return m.updateSessionDetailOverlay(key)
		}
		if m.sessionContextOpen {
			return m.updateSessionContextOverlay(key)
		}
		if m.helpOpen {
			state, cmd := overlaypkg.HandleHelp(m.overlayState(), key)
			return m.applyOverlayState(state), cmd
		}
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			return m.updateScratchpadPane(key)
		}
		if m.filterEditing {
			state, cmd := overlaypkg.HandleFilter(m.overlayState(), key, m.overlayDeps())
			if cmd != nil || state.FilterEditing != m.filterEditing {
				return m.applyOverlayState(state), cmd
			}
			next, cmd := filteringpkg.Update(m.filterState(), key, m.filterDeps())
			return m.applyFilterState(next), cmd
		}
		state, cmd := inputpkg.Handle(m.inputState(), key, m.inputDeps())
		return m.applyInputState(state), cmd
	}
	state, cmd, handled := dispatchpkg.HandleMessage(m.dispatchMessageState(), msg, m.dispatchMessageDeps())
	if handled {
		return m.applyDispatchMessageState(state), cmd
	}
	return m, nil
}

func (m Model) updateInstallScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.updateInstalling {
		return m, nil
	}
	switch msg.String() {
	case "esc", "enter", "q":
		m.updateInstallPhase = ""
		m.updateInstallDetail = ""
		m.updateInstallError = ""
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state, action, status := dialogstate.Update(m.dialogSnapshot(), msg)
	if status != "" {
		state.ErrorMessage = status
		return m.withDialogState(state), nil
	}
	state.ErrorMessage = ""
	next := m.withDialogState(state)
	if action == nil {
		return next, nil
	}
	return next, next.dialogActionCmd(*action)
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

func viewsShouldShowUpdate(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if !status.Enabled || !status.PromptEnabled || !status.UpdateAvailable {
		return false
	}
	return strings.TrimSpace(status.LatestVersion) != "" && strings.TrimSpace(status.LatestVersion) != strings.TrimSpace(status.DismissedVersion)
}

func viewsFirstUpdateSummary(status *api.UpdateStatus) string {
	if status == nil {
		return ""
	}
	if title := strings.TrimSpace(status.ReleaseName); title != "" {
		return title
	}
	for _, line := range strings.Split(status.ReleaseNotes, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}

func (m Model) overlayState() overlaypkg.State {
	cursor := map[string]int{"scratchpads": m.cursor[PaneScratchpads]}
	return overlaypkg.State{
		HelpOpen:           m.helpOpen,
		FilterEditing:      m.filterEditing,
		SessionDetailOpen:  m.sessionDetailOpen,
		SessionDetailY:     m.sessionDetailY,
		SessionDetail:      m.sessionDetail,
		ScratchpadOpen:     m.scratchpadOpen,
		ScratchpadMeta:     m.scratchpadMeta,
		ScratchpadFilePath: m.scratchpadFilePath,
		ScratchpadRendered: m.scratchpadRendered,
		ScratchpadViewport: m.scratchpadViewport,
		Scratchpads:        m.scratchpads,
		Cursor:             cursor,
		Timer:              m.timer,
	}
}

func (m Model) applyOverlayState(state overlaypkg.State) Model {
	m.helpOpen = state.HelpOpen
	m.filterEditing = state.FilterEditing
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionDetail = state.SessionDetail
	m.scratchpadOpen = state.ScratchpadOpen
	m.scratchpadMeta = state.ScratchpadMeta
	m.scratchpadFilePath = state.ScratchpadFilePath
	m.scratchpadRendered = state.ScratchpadRendered
	m.scratchpadViewport = state.ScratchpadViewport
	m.scratchpads = state.Scratchpads
	if state.Cursor != nil {
		m.cursor[PaneScratchpads] = state.Cursor["scratchpads"]
	}
	m.timer = state.Timer
	return m
}

func (m Model) overlayDeps() overlaypkg.Deps {
	return overlaypkg.Deps{
		StopFilterEdit: func(state *overlaypkg.State) {
			next := m.applyOverlayState(*state)
			next = next.applyFilterState(filteringpkg.Stop(next.filterState()))
			*state = next.overlayState()
		},
		SessionDetailMaxOffset: func(state overlaypkg.State) int {
			next := m.applyOverlayState(state)
			return helperpkg.SessionDetailMaxOffset(next.width, next.height, helperpkg.SessionDetailContentLines(next.sessionDetail))
		},
		OpenAmendSessionDialog: func(state *overlaypkg.State, sessionID string, commit string) {
			next := m.applyOverlayState(*state)
			next = next.openAmendSessionDialog(sessionID, commit)
			*state = next.overlayState()
		},
		SessionCommit: helperpkg.SessionCommit,
		OpenEditor: func(filePath string) tea.Cmd {
			return dialogruntime.OpenEditor(filePath, func(err error) tea.Msg { return commands.ErrMsg{Err: err} })
		},
		OpenDefaultViewer: func(filePath string) tea.Cmd {
			return dialogruntime.OpenDefaultViewer(filePath, func(err error) tea.Msg { return commands.ErrMsg{Err: err} })
		},
		SetStatus: func(state *overlaypkg.State, message string, isErr bool) tea.Cmd {
			next := m.applyOverlayState(*state)
			cmd := next.setStatus(message, isErr)
			*state = next.overlayState()
			return cmd
		},
		AbandonSelectedIssue: func(state *overlaypkg.State) tea.Cmd {
			next := m.applyOverlayState(*state)
			snapshot := next.selectionSnapshot()
			issue, ok := selectionpkg.SelectedIssueDetail(snapshot)
			if !ok {
				*state = next.overlayState()
				return nil
			}
			if issue.Status == "done" {
				cmd := next.setStatus("Done issues cannot be abandoned", true)
				*state = next.overlayState()
				return cmd
			}
			if issue.Status == "abandoned" {
				*state = next.overlayState()
				return nil
			}
			if next.timer != nil && next.timer.State != "idle" {
				next = next.withDialogState(dialogstate.OpenIssueSessionTransition(next.dialogSnapshot(), issue.ID, "abandoned"))
				*state = next.overlayState()
				return nil
			}
			next = next.withDialogState(dialogstate.OpenIssueStatusNote(next.dialogSnapshot(), "abandoned", "Abandon reason", true))
			next.dialogIssueID = issue.ID
			next.dialogStreamID = issue.StreamID
			*state = next.overlayState()
			return nil
		},
		FilteredIndexAtCursor: func(state overlaypkg.State, pane string) int {
			next := m.applyOverlayState(state)
			if pane == "scratchpads" {
				snapshot := next.selectionSnapshot()
				return selectionpkg.FilteredIndexAtCursor(snapshot, PaneScratchpads)
			}
			return -1
		},
		SetActiveScratchpadByIndex: func(state *overlaypkg.State, idx int) {
			next := m.applyOverlayState(*state)
			next.scratchpadMeta = helperpkg.ScratchpadMetaAt(next.scratchpads, idx)
			*state = next.overlayState()
		},
		ListLen: func(state overlaypkg.State, pane string) int {
			next := m.applyOverlayState(state)
			if pane == "scratchpads" {
				return (&next).listLen(PaneScratchpads)
			}
			return 0
		},
		OpenScratchpad: func(idx int) tea.Cmd {
			return commands.OpenScratchpad(m.client, m.scratchpads, idx)
		},
	}
}

func (m Model) updateSessionDetailOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state, cmd := overlaypkg.HandleSessionDetail(m.overlayState(), msg, m.overlayDeps())
	return m.applyOverlayState(state), cmd
}

func (m Model) updateSessionContextOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		m.sessionContextOpen = false
		m.sessionContextY = 0
		return m, nil
	case "j", "down":
		snapshot := m.selectionSnapshot()
		maxOffset := helperpkg.SessionContextMaxOffset(m.width, m.height, helperpkg.SessionContextContentLines(selectionpkg.ActiveIssue(snapshot)))
		if m.sessionContextY < maxOffset {
			m.sessionContextY++
		}
		return m, nil
	case "k", "up":
		if m.sessionContextY > 0 {
			m.sessionContextY--
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateScratchpadPane(msg tea.KeyMsg) (Model, tea.Cmd) {
	state, cmd := overlaypkg.HandleScratchpad(m.overlayState(), msg, m.overlayDeps())
	m = m.applyOverlayState(state)
	if cmd != nil {
		return m, cmd
	}
	var nextCmd tea.Cmd
	m.scratchpadViewport, nextCmd = m.scratchpadViewport.Update(msg)
	return m, nextCmd
}

func (m Model) dispatchEventState() dispatchpkg.EventState {
	return dispatchpkg.EventState{
		View:               m.view,
		Pane:               m.pane,
		Cursor:             m.cursor,
		Streams:            m.streams,
		Issues:             m.issues,
		Habits:             m.habits,
		Context:            m.context,
		Timer:              m.timer,
		Elapsed:            m.elapsed,
		TimerTickSeq:       m.timerTickSeq,
		CurrentDash:        m.currentDashboardDate(),
		CurrentRollupStart: m.currentRollupStartDate(),
		CurrentRollupEnd:   m.currentRollupEndDate(),
		CurrentWell:        m.currentWellbeingDate(),
		CurrentOpsLim:      m.currentOpsLimit(),
	}
}

func (m Model) applyDispatchEventState(state dispatchpkg.EventState) Model {
	m.view = state.View
	m.pane = state.Pane
	m.cursor = state.Cursor
	m.streams = state.Streams
	m.issues = state.Issues
	m.habits = state.Habits
	m.context = state.Context
	m.timer = state.Timer
	m.elapsed = state.Elapsed
	m.timerTickSeq = state.TimerTickSeq
	return m
}

func (m Model) handleKernelEvent(event api.KernelEvent) (Model, tea.Cmd) {
	state, cmd := dispatchpkg.HandleEvent(m.dispatchEventState(), dispatchpkg.EventDeps{
		LoadRepos:           func() tea.Cmd { return commands.LoadRepos(m.client) },
		LoadStreams:         func(repoID int64) tea.Cmd { return commands.LoadStreams(m.client, repoID) },
		LoadIssues:          func(streamID int64) tea.Cmd { return commands.LoadIssues(m.client, streamID) },
		LoadHabits:          func(streamID int64) tea.Cmd { return commands.LoadHabits(m.client, streamID) },
		LoadAllIssues:       func() tea.Cmd { return commands.LoadAllIssues(m.client) },
		LoadDailySummary:    func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDueHabits:       func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadWellbeing:       func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		LoadRollupSummaries: func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		LoadScratchpads:     func() tea.Cmd { return commands.LoadScratchpads(m.client) },
		LoadSessionHistoryFor200: func(state dispatchpkg.EventState) tea.Cmd {
			return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(m.timer), 200)
		},
		LoadStashes:      func() tea.Cmd { return commands.LoadStashes(m.client) },
		LoadContext:      func() tea.Cmd { return commands.LoadContext(m.client) },
		LoadTimer:        func() tea.Cmd { return commands.LoadTimer(m.client) },
		LoadUpdateStatus: func() tea.Cmd { return commands.LoadUpdateStatus(m.client) },
		LoadOps:          func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		TickAfter:        commands.TickAfter,
	}, event)
	return m.applyDispatchEventState(state), cmd
}

func (m Model) dispatchMessageState() dispatchpkg.MessageState {
	return dispatchpkg.MessageState{
		Width:                   m.width,
		Height:                  m.height,
		View:                    m.view,
		Pane:                    m.pane,
		Cursor:                  m.cursor,
		Repos:                   m.repos,
		Streams:                 m.streams,
		Issues:                  m.issues,
		Habits:                  m.habits,
		AllIssues:               m.allIssues,
		DueHabits:               m.dueHabits,
		DailySummary:            m.dailySummary,
		DailyPlan:               m.dailyPlan,
		DashboardDate:           m.dashboardDate,
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
		UpdateStatus:            m.updateStatus,
		UpdateChecking:          m.updateChecking,
		UpdateInstalling:        m.updateInstalling,
		UpdateInstallPhase:      m.updateInstallPhase,
		UpdateInstallDetail:     m.updateInstallDetail,
		UpdateInstallOutput:     m.updateInstallOutput,
		UpdateInstallError:      m.updateInstallError,
		UpdateInstallProgress:   nil,
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
		DialogChoiceItems:       m.dialogChoiceItems,
		DialogChoiceCursor:      m.dialogChoiceCursor,
		DialogProcessing:        m.dialogProcessing,
		DialogProcessingLabel:   m.dialogProcessingLabel,
		DialogViewTitle:         m.dialogViewTitle,
		DialogViewName:          m.dialogViewName,
		DialogViewMeta:          m.dialogViewMeta,
		DialogViewBody:          m.dialogViewBody,
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
	m.repos = state.Repos
	m.streams = state.Streams
	m.issues = state.Issues
	m.habits = state.Habits
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
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.updateInstalling = state.UpdateInstalling
	m.updateInstallPhase = state.UpdateInstallPhase
	m.updateInstallDetail = state.UpdateInstallDetail
	m.updateInstallOutput = state.UpdateInstallOutput
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
	m.dialogChoiceItems = state.DialogChoiceItems
	m.dialogChoiceCursor = state.DialogChoiceCursor
	m.dialogProcessing = state.DialogProcessing
	m.dialogProcessingLabel = state.DialogProcessingLabel
	m.dialogViewTitle = state.DialogViewTitle
	m.dialogViewName = state.DialogViewName
	m.dialogViewMeta = state.DialogViewMeta
	m.dialogViewBody = state.DialogViewBody
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
		LoadSessionHistoryFor200: func(state dispatchpkg.MessageState) tea.Cmd {
			next := m.applyDispatchMessageState(state)
			return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(next.timer), 200)
		},
		LoadSessionDetail: func(id string) tea.Cmd { return commands.LoadSessionDetail(m.client, id) },
		LoadScratchpads:   func() tea.Cmd { return commands.LoadScratchpads(m.client) },
		LoadStashes:       func() tea.Cmd { return commands.LoadStashes(m.client) },
		LoadOps:           func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		LoadContext:       func() tea.Cmd { return commands.LoadContext(m.client) },
		LoadTimer:         func() tea.Cmd { return commands.LoadTimer(m.client) },
		LoadHealth:        func() tea.Cmd { return commands.LoadHealth(m.client) },
		LoadUpdateStatus:  func() tea.Cmd { return commands.LoadUpdateStatus(m.client) },
		LoadSettings:      func() tea.Cmd { return commands.LoadSettings(m.client) },
		LoadKernelInfo:    func() tea.Cmd { return commands.LoadKernelInfo(m.client) },
		HealthTickAfter:   commands.HealthTickAfter,
		TickAfter:         commands.TickAfter,
		WaitForEvent:      func() tea.Cmd { return commands.WaitForEvent(eventChannel) },
		HandleKernelEvent: func(state dispatchpkg.MessageState, event api.KernelEvent) (dispatchpkg.MessageState, tea.Cmd) {
			next := m.applyDispatchMessageState(state)
			nextModel, cmd := next.handleKernelEvent(event)
			return nextModel.dispatchMessageState(), cmd
		},
		CloseEventStop: func() { m.stopEventStream() },
	}
}
