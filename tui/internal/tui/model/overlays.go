package model

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	helperpkg "crona/tui/internal/tui/helpers"
	overlaypkg "crona/tui/internal/tui/overlays"
	selectionpkg "crona/tui/internal/tui/selection"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) overlayState() overlaypkg.State {
	return overlaypkg.State{
		HelpOpen:          m.helpOpen,
		DialogState:       m.dialogState(),
		SessionDetailOpen: m.sessionDetailOpen,
		SessionDetailY:    m.sessionDetailY,
		SessionDetail:     m.sessionDetail,
		ManageHistory:     m.settings != nil && m.settings.ManageHistory,
		Cursor:            map[string]int{},
		Timer:             m.timer,
	}
}

func (m Model) applyOverlayState(state overlaypkg.State) Model {
	m.helpOpen = state.HelpOpen
	m = m.withDialogState(state.DialogState)
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionDetail = state.SessionDetail
	m.timer = state.Timer
	return m
}

func (m Model) overlayDeps() overlaypkg.Deps {
	return overlaypkg.Deps{
		SessionDetailMaxOffset: func(state overlaypkg.State) int {
			next := m.applyOverlayState(state)
			return helperpkg.SessionDetailMaxOffset(
				next.width,
				next.height,
				helperpkg.SessionDetailContentLines(next.sessionDetail),
			)
		},
		OpenAmendSessionDialog: func(state *overlaypkg.State, sessionID string, commit string) {
			next := m.applyOverlayState(*state)
			next = next.openAmendSessionDialog(sessionID, commit)
			next.sessionDetailOpen = false
			next.sessionDetail = nil
			next.sessionDetailY = 0
			*state = next.overlayState()
		},
		OpenEditSessionHistory: func(state *overlaypkg.State, detail *api.SessionDetail) {
			next := m.applyOverlayState(*state)
			next.sessionDetailOpen = false
			next.sessionDetail = nil
			next = next.withDialogState(dialogstate.OpenEditSessionHistory(next.dialogSnapshot().Dialog, detail))
			*state = next.overlayState()
		},
		OpenDeleteSessionHistory: func(state *overlaypkg.State, id string) {
			next := m.applyOverlayState(*state)
			next.sessionDetailOpen = false
			next.sessionDetail = nil
			next = next.withDialogState(next.dialogSnapshot().OpenConfirmDeleteEntity("session_history", id, "this session history entry"))
			*state = next.overlayState()
		},
		SessionCommit: helperpkg.SessionCommit,
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
				next = next.withDialogState(
					next.dialogSnapshot().OpenIssueSessionTransition(issue.ID, "abandoned"),
				)
				*state = next.overlayState()
				return nil
			}
			next = next.withDialogState(
				next.dialogSnapshot().OpenIssueStatusNote("abandoned", "Abandon reason", true),
			)
			next.dialogIssueID = issue.ID
			next.dialogStreamID = issue.StreamID
			*state = next.overlayState()
			return nil
		},
	}
}

func (m Model) updateSessionDetailOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	state, cmd := overlaypkg.HandleSessionDetail(m.overlayState(), msg, m.overlayDeps())
	return m.applyOverlayState(state), cmd
}

func (m Model) updateSessionContextOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.sessionContextOpen = false
		m.sessionContextY = 0
		return m, nil
	case "down":
		snapshot := m.selectionSnapshot()
		maxOffset := helperpkg.SessionContextMaxOffset(
			m.width,
			m.height,
			helperpkg.SessionContextContentLines(selectionpkg.ActiveIssue(snapshot)),
		)
		if m.sessionContextY < maxOffset {
			m.sessionContextY++
		}
		return m, nil
	case "up":
		if m.sessionContextY > 0 {
			m.sessionContextY--
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) dispatchEventState() dispatchpkg.EventState {
	var selectedIssueID *int64
	if issue, ok := selectionpkg.SelectedIssueDetail(m.selectionSnapshot()); ok {
		selectedIssueID = &issue.ID
	}
	var selectedHabitHistoryID *int64
	if entry, ok := selectionpkg.SelectedHabitHistoryEntry(m.selectionSnapshot()); ok {
		selectedHabitHistoryID = &entry.ID
	}
	return dispatchpkg.EventState{
		View:                   m.view,
		Pane:                   m.pane,
		Cursor:                 m.cursor,
		Dialog:                 m.dialog,
		DialogParent:           m.dialogParent,
		SelectedIssueID:        selectedIssueID,
		Streams:                m.streams,
		Issues:                 m.issues,
		Habits:                 m.habits,
		SelectedHabitHistoryID: selectedHabitHistoryID,
		Context:                m.context,
		Timer:                  m.timer,
		Elapsed:                m.elapsed,
		TimerTickSeq:           m.timerTickSeq,
		Date:                   m.date,
		CurrentDate:            m.currentDate(),
		Today:                  m.currentLogicalDate(),
		ReviewWindowDays:       m.reviewWindowDays,
		ReviewWindowEnd:        m.reviewWindowEnd,
		MomentumWindowDays:     m.currentMomentumWindowDays(),
		CurrentOpsLim:          m.currentOpsLimit(),
	}
}

func (m Model) applyDispatchEventState(state dispatchpkg.EventState) Model {
	m.view = state.View
	m.pane = state.Pane
	m.cursor = state.Cursor
	m.dialog = state.Dialog
	m.dialogParent = state.DialogParent
	m.streams = state.Streams
	m.issues = state.Issues
	m.habits = state.Habits
	m.context = state.Context
	m.timer = state.Timer
	m.elapsed = state.Elapsed
	m.timerTickSeq = state.TimerTickSeq
	m.date = state.Date
	if state.Today != "" && state.Today != m.currentLogicalDate() {
		health := api.Health{}
		if m.health != nil {
			health = *m.health
		}
		health.CurrentDate = state.Today
		m.health = &health
	}
	return m
}

func (m Model) handleKernelEvent(event api.KernelEvent) (Model, tea.Cmd) {
	state, cmd := dispatchpkg.HandleEvent(m.dispatchEventState(), dispatchpkg.EventDeps{
		LoadSummarySnapshot: func(date string) tea.Cmd { return commands.LoadSummarySnapshot(m.client, date) },
		LoadIssueSessions:   func(id int64) tea.Cmd { return commands.LoadIssueSessions(m.client, id) },
		LoadRepos:           func() tea.Cmd { return commands.LoadRepos(m.client) },
		LoadStreams:         func(repoID int64) tea.Cmd { return commands.LoadStreams(m.client, repoID) },
		LoadIssues:          func(streamID int64) tea.Cmd { return commands.LoadIssues(m.client, streamID) },
		LoadIssuesSelecting: func(streamID, selectedIssueID int64) tea.Cmd {
			return commands.LoadIssuesSelecting(m.client, streamID, selectedIssueID)
		},
		LoadHabits:     func(streamID int64) tea.Cmd { return commands.LoadHabits(m.client, streamID) },
		LoadAllHabits:  func() tea.Cmd { return commands.LoadAllHabits(m.client) },
		LoadAllStreams: func() tea.Cmd { return commands.LoadAllStreams(m.client) },
		LoadAllIssues:  func() tea.Cmd { return commands.LoadAllIssues(m.client) },
		LoadAllIssuesSelecting: func(selectedIssueID int64) tea.Cmd {
			return commands.LoadAllIssuesSelecting(m.client, selectedIssueID)
		},
		LoadDailySummary: func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDailyPlan:    func(date string) tea.Cmd { return commands.LoadDailyPlan(m.client, date) },
		LoadDueHabits:    func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadTodayHabits:  func(date string) tea.Cmd { return commands.LoadTodayHabits(m.client, date) },
		LoadDailyStreaks: func(date string) tea.Cmd { return commands.LoadDailyStreaks(m.client, date) },
		LoadHabitHistory: func(ctx *api.ActiveContext, selectedID *int64) tea.Cmd {
			return commands.LoadHabitHistory(m.client, ctx, selectedID)
		},
		LoadWellbeing: func(date string) tea.Cmd { return commands.LoadWellbeingDay(m.client, date) },
		LoadMomentumRange: func(date string, windowDays int) tea.Cmd {
			return commands.LoadMomentumRange(m.client, date, windowDays)
		},
		LoadReviewWindow: func(start, end string) tea.Cmd { return commands.LoadReviewWindow(m.client, start, end) },
		LoadSessionHistoryFor200: func(state dispatchpkg.EventState) tea.Cmd {
			return commands.LoadSessionHistory(
				m.client,
				helperpkg.SessionHistoryScopeIssueID(m.timer),
				200,
			)
		},
		LoadContext:      func() tea.Cmd { return commands.LoadContext(m.client) },
		LoadTimer:        func() tea.Cmd { return commands.LoadTimer(m.client) },
		LoadAlertStatus:  func() tea.Cmd { return commands.LoadAlertStatus(m.client) },
		LoadSettings:     func() tea.Cmd { return commands.LoadSettings(m.client) },
		LoadUpdateStatus: func() tea.Cmd { return commands.LoadUpdateStatus(m.client) },
		LoadOps:          func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		TickAfter:        commands.TickAfter,
	}, event)
	next := m.applyDispatchEventState(state)
	if event.Type == sharedtypes.EventTypeTimerHardLimitReached && next.dialog == "" {
		label := next.terminalSessionTitle()
		next = next.openHardLimitExpiredDialog(label)
	}
	return next, cmd
}
