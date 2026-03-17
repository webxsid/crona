package app

import (
	"crona/tui/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.opsLimitPinned {
			nextLimit := m.defaultOpsLimit()
			if nextLimit != m.opsLimit {
				m.opsLimit = nextLimit
				if m.client != nil {
					if m.scratchpadOpen {
						m.syncScratchpadViewport()
					}
					return m, loadOps(m.client, m.currentOpsLimit())
				}
			}
		}
		if m.scratchpadOpen {
			m.syncScratchpadViewport()
		}
		return m, nil
	case reposLoadedMsg:
		m.repos = msg.repos
		m.clampFiltered(PaneRepos)
		return m, nil
	case streamsLoadedMsg:
		m.streams = msg.streams
		m.clampFiltered(PaneStreams)
		return m, nil
	case issuesLoadedMsg:
		m.issues = msg.issues
		m.clampFiltered(PaneIssues)
		return m, nil
	case habitsLoadedMsg:
		m.habits = msg.habits
		m.clampFiltered(PaneHabits)
		return m, nil
	case allIssuesLoadedMsg:
		m.allIssues = msg.issues
		if m.view == ViewDefault || m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
		}
		return m, nil
	case dueHabitsLoadedMsg:
		m.dueHabits = msg.habits
		if m.view == ViewDaily {
			m.clampFiltered(PaneHabits)
		}
		return m, nil
	case dailySummaryLoadedMsg:
		m.dailySummary = msg.summary
		if m.dashboardDate == "" && msg.summary != nil {
			m.dashboardDate = msg.summary.Date
		}
		if m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
		}
		return m, nil
	case dailyCheckInLoadedMsg:
		m.dailyCheckIn = msg.checkIn
		if m.dailyCheckIn != nil && m.wellbeingDate == "" {
			m.wellbeingDate = m.dailyCheckIn.Date
		}
		return m, nil
	case metricsRangeLoadedMsg:
		m.metricsRange = msg.days
		return m, nil
	case metricsRollupLoadedMsg:
		m.metricsRollup = msg.rollup
		return m, nil
	case streaksLoadedMsg:
		m.streaks = msg.streaks
		return m, nil
	case issueSessionsLoadedMsg:
		var activeIssueID int64
		if m.context != nil && m.context.IssueID != nil {
			activeIssueID = *m.context.IssueID
		} else if m.timer != nil && m.timer.IssueID != nil {
			activeIssueID = *m.timer.IssueID
		}
		if msg.issueID == activeIssueID {
			m.issueSessions = msg.sessions
		}
		return m, nil
	case sessionHistoryLoadedMsg:
		m.sessionHistory = msg.sessions
		m.clampFiltered(PaneSessions)
		return m, nil
	case sessionDetailLoadedMsg:
		m.sessionDetail = msg.detail
		m.sessionDetailY = 0
		if msg.detail == nil {
			m.sessionDetailOpen = false
			return m, m.setStatus("Session detail is unavailable", true)
		}
		return m, nil
	case sessionDetailFailedMsg:
		m.sessionDetailOpen = false
		m.sessionDetail = nil
		m.sessionDetailY = 0
		logger.Errorf("session detail failed: %v", msg.err)
		return m, m.setStatus("Error: "+msg.err.Error(), true)
	case scratchpadsLoadedMsg:
		m.scratchpads = msg.pads
		m.clampFiltered(PaneScratchpads)
		if m.scratchpadMeta != nil {
			if idx := m.scratchpadTabIndexByID(m.scratchpadMeta.ID); idx >= 0 {
				if filteredCur := m.filteredCursorForRawIndex(PaneScratchpads, idx); filteredCur >= 0 {
					m.cursor[PaneScratchpads] = filteredCur
				}
				m.setActiveScratchpadByIndex(idx)
			} else {
				m.scratchpadOpen = false
				m.scratchpadMeta = nil
				m.scratchpadFilePath = ""
				m.scratchpadRendered = ""
			}
		}
		return m, nil
	case stashesLoadedMsg:
		m.stashes = msg.stashes
		if m.dialogStashCursor >= len(m.stashes) {
			if len(m.stashes) == 0 {
				m.dialogStashCursor = 0
			} else {
				m.dialogStashCursor = len(m.stashes) - 1
			}
		}
		return m, nil
	case opsLoadedMsg:
		m.ops = msg.ops
		m.clampFiltered(PaneOps)
		return m, nil
	case contextLoadedMsg:
		m.context = msg.ctx
		if m.view == ViewDefault || m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
			m.clampFiltered(PaneHabits)
		}
		if m.context != nil && m.context.IssueID != nil {
			return m, loadIssueSessions(m.client, *m.context.IssueID)
		}
		m.issueSessions = nil
		return m, nil
	case timerLoadedMsg:
		m.timer = msg.timer
		m.elapsed = 0
		m.timerTickSeq++
		if m.timer != nil && m.timer.State != "idle" {
			if m.view != ViewScratch {
				m.view = ViewSessionActive
			}
			m.pane = viewDefaultPane[m.view]
		} else if m.view == ViewSessionActive {
			m.view = ViewDaily
			m.pane = viewDefaultPane[m.view]
		}
		if m.timer != nil && m.timer.IssueID != nil {
			if m.context == nil || m.context.IssueID == nil || *m.context.IssueID != *m.timer.IssueID {
				return m, tea.Batch(loadIssueSessions(m.client, *m.timer.IssueID), tickAfter(m.timerTickSeq))
			}
		}
		if m.timer != nil && m.timer.State != "idle" {
			return m, tickAfter(m.timerTickSeq)
		}
		return m, nil
	case healthLoadedMsg:
		m.health = msg.health
		return m, nil
	case settingsLoadedMsg:
		m.settings = msg.settings
		m.clampFiltered(PaneSettings)
		return m, nil
	case kernelInfoLoadedMsg:
		m.kernelInfo = msg.info
		return m, nil
	case healthTickMsg:
		return m, tea.Batch(loadHealth(m.client), healthTickAfter())
	case clearStatusMsg:
		if msg.seq != m.statusSeq {
			return m, nil
		}
		m.statusMsg = ""
		m.statusErr = false
		return m, nil
	case kernelShutdownMsg:
		close(m.eventStop)
		return m, tea.Quit
	case devSeededMsg:
		cmd := m.setStatus("Dev seed loaded", false)
		m.view = ViewDaily
		m.pane = viewDefaultPane[m.view]
		return m, tea.Batch(cmd, loadKernelInfo(m.client), loadRepos(m.client), loadAllIssues(m.client), loadDueHabits(m.client, m.currentDashboardDate()), loadDailySummary(m.client, m.dashboardDate), loadWellbeing(m.client, m.currentWellbeingDate()), loadSessionHistory(m.client, 200), loadScratchpads(m.client), loadStashes(m.client), loadOps(m.client, m.currentOpsLimit()), loadContext(m.client), loadTimer(m.client), loadSettings(m.client))
	case devClearedMsg:
		cmd := m.setStatus("Dev data cleared", false)
		m.view = ViewDaily
		m.pane = viewDefaultPane[m.view]
		return m, tea.Batch(cmd, loadKernelInfo(m.client), loadRepos(m.client), loadAllIssues(m.client), loadDueHabits(m.client, m.currentDashboardDate()), loadDailySummary(m.client, m.dashboardDate), loadWellbeing(m.client, m.currentWellbeingDate()), loadSessionHistory(m.client, 200), loadScratchpads(m.client), loadStashes(m.client), loadOps(m.client, m.currentOpsLimit()), loadContext(m.client), loadTimer(m.client), loadSettings(m.client))
	case sessionAmendedMsg:
		cmd := m.setStatus("Session amended", false)
		return m, tea.Batch(cmd, loadSessionHistory(m.client, 200), loadSessionDetail(m.client, msg.id))
	case focusSessionChangedMsg:
		cmds := []tea.Cmd{}
		if msg.reloadContext {
			cmds = append(cmds, loadContext(m.client))
		}
		if msg.reloadTimer {
			cmds = append(cmds, loadTimer(m.client))
		}
		if len(cmds) == 0 {
			return m, nil
		}
		return m, tea.Batch(cmds...)
	case timerTickMsg:
		if msg.seq != m.timerTickSeq {
			return m, nil
		}
		if m.timer != nil && m.timer.State != "idle" {
			m.elapsed++
			return m, tickAfter(m.timerTickSeq)
		}
		return m, nil
	case kernelEventMsg:
		updated, cmd := handleKernelEvent(m, msg.event)
		return updated, tea.Batch(cmd, waitForEvent(eventChannel))
	case errMsg:
		logger.Errorf("update error: %v", msg.err)
		return m, m.setStatus("Error: "+msg.err.Error(), true)
	case openScratchpadMsg:
		return m.enterScratchpadPane(msg), nil
	case scratchpadReloadedMsg:
		m.scratchpadFilePath = msg.filePath
		m.scratchpadRendered = msg.rendered
		m.scratchpadViewport.SetContent(msg.rendered)
		return m, nil
	case editorDoneMsg:
		cmds := []tea.Cmd{loadScratchpads(m.client)}
		if m.scratchpadOpen && m.scratchpadMeta != nil {
			cmds = append(cmds, cmdReloadScratchpad(m.client, m.scratchpadMeta, m.scratchpadRenderWidth()))
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		if m.dialog != "" {
			return m.updateDialog(msg)
		}
		if m.sessionDetailOpen {
			return m.updateSessionDetailOverlay(msg)
		}
		if m.helpOpen {
			switch msg.String() {
			case "?", "esc", "q":
				m.helpOpen = false
			}
			return m, nil
		}
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			return m.updateScratchpadPane(msg)
		}
		if m.filterEditing {
			return m.updateFilter(msg)
		}
		return m.handleKey(msg)
	default:
		return m, nil
	}
}
