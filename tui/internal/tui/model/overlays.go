package model

import (
	"strings"

	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	dispatchpkg "crona/tui/internal/tui/dispatch"
	filteringpkg "crona/tui/internal/tui/filtering"
	helperpkg "crona/tui/internal/tui/helpers"
	overlaypkg "crona/tui/internal/tui/overlays"
	selectionpkg "crona/tui/internal/tui/selection"

	tea "github.com/charmbracelet/bubbletea"
)

func viewsShouldShowUpdate(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if !status.Enabled || !status.PromptEnabled || !status.UpdateAvailable {
		return false
	}
	return strings.TrimSpace(status.LatestVersion) != "" && strings.TrimSpace(status.LatestVersion) != strings.TrimSpace(status.DismissedVersion)
}

func (m Model) overlayState() overlaypkg.State {
	cursor := map[string]int{"scratchpads": m.cursor[PaneScratchpads]}
	return overlaypkg.State{
		HelpOpen:           m.helpOpen,
		FilterEditing:      m.filterEditing,
		DialogState:        m.dialogState(),
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
	m = m.withDialogState(state.DialogState)
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
			next.sessionDetailOpen = false
			next.sessionDetail = nil
			next.sessionDetailY = 0
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
				next = next.withDialogState(next.dialogSnapshot().OpenIssueSessionTransition(issue.ID, "abandoned"))
				*state = next.overlayState()
				return nil
			}
			next = next.withDialogState(next.dialogSnapshot().OpenIssueStatusNote("abandoned", "Abandon reason", true))
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
		SelectedIssueID:        selectedIssueID,
		Streams:                m.streams,
		Issues:                 m.issues,
		Habits:                 m.habits,
		SelectedHabitHistoryID: selectedHabitHistoryID,
		Context:                m.context,
		Timer:                  m.timer,
		Elapsed:                m.elapsed,
		TimerTickSeq:           m.timerTickSeq,
		CurrentDash:            m.currentDashboardDate(),
		CurrentRollupStart:     m.currentRollupStartDate(),
		CurrentRollupEnd:       m.currentRollupEndDate(),
		CurrentWell:            m.currentWellbeingDate(),
		CurrentOpsLim:          m.currentOpsLimit(),
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
		LoadRepos:   func() tea.Cmd { return commands.LoadRepos(m.client) },
		LoadStreams: func(repoID int64) tea.Cmd { return commands.LoadStreams(m.client, repoID) },
		LoadIssues:  func(streamID int64) tea.Cmd { return commands.LoadIssues(m.client, streamID) },
		LoadIssuesSelecting: func(streamID, selectedIssueID int64) tea.Cmd {
			return commands.LoadIssuesSelecting(m.client, streamID, selectedIssueID)
		},
		LoadHabits:    func(streamID int64) tea.Cmd { return commands.LoadHabits(m.client, streamID) },
		LoadAllHabits: func() tea.Cmd { return commands.LoadAllHabits(m.client) },
		LoadAllIssues: func() tea.Cmd { return commands.LoadAllIssues(m.client) },
		LoadAllIssuesSelecting: func(selectedIssueID int64) tea.Cmd {
			return commands.LoadAllIssuesSelecting(m.client, selectedIssueID)
		},
		LoadDailySummary: func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDueHabits:    func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		LoadHabitHistory: func(ctx *api.ActiveContext, selectedID *int64) tea.Cmd {
			return commands.LoadHabitHistory(m.client, ctx, selectedID)
		},
		LoadWellbeing:       func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		LoadRollupSummaries: func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		LoadScratchpads:     func() tea.Cmd { return commands.LoadScratchpads(m.client) },
		LoadSessionHistoryFor200: func(state dispatchpkg.EventState) tea.Cmd {
			return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(m.timer), 200)
		},
		LoadStashes:      func() tea.Cmd { return commands.LoadStashes(m.client) },
		LoadContext:      func() tea.Cmd { return commands.LoadContext(m.client) },
		LoadTimer:        func() tea.Cmd { return commands.LoadTimer(m.client) },
		LoadAlertStatus:  func() tea.Cmd { return commands.LoadAlertStatus(m.client) },
		LoadUpdateStatus: func() tea.Cmd { return commands.LoadUpdateStatus(m.client) },
		LoadOps:          func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		TickAfter:        commands.TickAfter,
	}, event)
	return m.applyDispatchEventState(state), cmd
}
