package tui

import (
	"crona/tui/internal/logger"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ---------- Terminal ----------

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

	// ---------- Data loaded ----------

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

	case allIssuesLoadedMsg:
		m.allIssues = msg.issues
		if m.view == ViewDefault || m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
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

	case issueSessionsLoadedMsg:
		activeIssueID := ""
		if m.context != nil && m.context.IssueID != nil {
			activeIssueID = *m.context.IssueID
		} else if m.timer != nil && m.timer.IssueID != nil {
			activeIssueID = *m.timer.IssueID
		}
		if msg.issueID == activeIssueID {
			m.issueSessions = msg.sessions
		}
		return m, nil

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

	case opsLoadedMsg:
		m.ops = msg.ops
		m.clampFiltered(PaneOps)
		return m, nil

	case contextLoadedMsg:
		m.context = msg.ctx
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
				m.view = ViewSession
			}
			m.pane = viewDefaultPane[m.view]
		} else if m.view == ViewSession {
			m.view = ViewDefault
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

	case healthTickMsg:
		return m, tea.Batch(loadHealth(m.client), healthTickAfter())

	case kernelShutdownMsg:
		close(m.sseStop)
		return m, tea.Quit

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

	// ---------- Timer tick ----------

	case timerTickMsg:
		if msg.seq != m.timerTickSeq {
			return m, nil
		}
		if m.timer != nil && m.timer.State != "idle" {
			m.elapsed++
			return m, tickAfter(m.timerTickSeq)
		}
		return m, nil

	// ---------- SSE ----------

	case kernelEventMsg:
		updated, cmd := handleKernelEvent(m, msg.event)
		// immediately wait for next event
		return updated, tea.Batch(cmd, waitForSSE(sseChannel))

	// ---------- Errors ----------

	case errMsg:
		logger.Errorf("update error: %v", msg.err)
		m.statusMsg = "Error: " + msg.err.Error()
		return m, nil

	// ---------- Scratchpad view ----------

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

	// ---------- Keyboard ----------

	case tea.KeyMsg:
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			return m.updateScratchpadPane(msg)
		}
		if m.filterEditing {
			return m.updateFilter(msg)
		}
		// if a dialog is open, route keys there
		if m.dialog != "" {
			return m.updateDialog(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// ---------- Global ----------
	switch key {
	case "q", "ctrl+c":
		close(m.sseStop)
		return m, tea.Quit

	case "K":
		return m, cmdShutdownKernel(m.client)

	case "]":
		if m.timer != nil && m.timer.State != "idle" {
			if m.view == ViewSession {
				m.view = ViewScratch
			} else {
				m.view = ViewSession
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
				m.view = ViewSession
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
		m.pane = nextPane(m.view, m.pane, 1)
		return m, nil

	case "shift+tab":
		if m.filterEditing {
			return m, nil
		}
		m.pane = nextPane(m.view, m.pane, -1)
		return m, nil

	case "R":
		// TODO: kernel restart (POST /kernel/restart, re-init)
		logger.Info("Kernel restart requested")
		return m, nil
	}

	// ---------- Cursor movement ----------
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

	// ---------- Pane-number shortcuts (Meta view) ----------
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
		}
	}
	if m.view == ViewDefault {
		switch key {
		case "1":
			m.pane = PaneIssues
			return m, nil
		}
	}
	if m.view == ViewScratch {
		switch key {
		case "1":
			m.pane = PaneScratchpads
			return m, nil
		}
	}

	// ---------- Actions ----------
	switch key {
	case "s":
		if m.view == ViewSession {
			if m.timer == nil || m.timer.State == "idle" {
				if m.context == nil || m.context.RepoID == nil || m.context.StreamID == nil || m.context.IssueID == nil {
					m.statusMsg = "No active issue in context"
					return m, nil
				}
				return m, cmdStartFocusSession(m.client, *m.context.RepoID, *m.context.StreamID, *m.context.IssueID)
			}
			return m, nil
		}
		if m.pane == PaneIssues {
			if m.view == ViewDefault {
				rawIdx := m.filteredIndexAtCursor(PaneIssues)
				if rawIdx < 0 || rawIdx >= len(m.allIssues) {
					return m, nil
				}
				issue := m.allIssues[rawIdx]
				return m, cmdStartFocusSession(m.client, issue.RepoID, issue.StreamID, issue.ID)
			}
			if m.view == ViewDaily {
				rawIdx := m.filteredIndexAtCursor(PaneIssues)
				if m.dailySummary == nil || rawIdx < 0 || rawIdx >= len(m.dailySummary.Issues) {
					return m, nil
				}
				issue := m.dailySummary.Issues[rawIdx]
				meta := m.issueMetaByID(issue.ID)
				if meta == nil {
					m.statusMsg = "Issue metadata unavailable"
					return m, nil
				}
				return m, cmdStartFocusSession(m.client, meta.RepoID, issue.StreamID, issue.ID)
			}
			rawIdx := m.filteredIndexAtCursor(PaneIssues)
			if rawIdx < 0 || rawIdx >= len(m.issues) {
				return m, nil
			}
			if m.context == nil || m.context.RepoID == nil {
				m.statusMsg = "No repo in context for selected issue"
				return m, nil
			}
			issue := m.issues[rawIdx]
			return m, cmdStartFocusSession(m.client, *m.context.RepoID, issue.StreamID, issue.ID)
		}

	case "t":
		return m.cycleSelectedIssueStatus()

	case "A":
		return m.abandonSelectedIssue()

	case "y":
		return m.toggleSelectedIssueToday()

	case "D":
		return m.openSelectedIssueTodoDateDialog(), nil

	case ",":
		if m.view == ViewDaily {
			m.dashboardDate = shiftISODate(m.currentDashboardDate(), -1)
			return m, loadDailySummary(m.client, m.dashboardDate)
		}

	case ".":
		if m.view == ViewDaily {
			m.dashboardDate = shiftISODate(m.currentDashboardDate(), 1)
			return m, loadDailySummary(m.client, m.dashboardDate)
		}

	case "g":
		if m.view == ViewDaily {
			m.dashboardDate = ""
			return m, loadDailySummary(m.client, "")
		}

	case "c":
		return m.checkout()

	case "p":
		if m.view == ViewSession && m.timer != nil && m.timer.State == "running" {
			return m, cmdPauseFocusSession(m.client)
		}

	case "r":
		if m.view == ViewSession && m.timer != nil && m.timer.State == "paused" {
			return m, cmdResumeFocusSession(m.client)
		}

	case "x":
		if m.view == ViewSession && m.timer != nil && m.timer.State != "idle" {
			return m.openSessionMessageDialog("end_session"), nil
		}

	case "z":
		if m.view == ViewSession && m.timer != nil && m.timer.State != "idle" {
			return m.openSessionMessageDialog("stash_session"), nil
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
		if m.pane != PaneOps && m.pane != PaneIssues && m.pane != PaneRepos && m.pane != PaneStreams && m.pane != PaneScratchpads {
			return m, nil
		}
		if m.pane == PaneScratchpads && m.scratchpadOpen {
			return m, nil
		}
		m.startFilterEdit(m.pane)
		return m, nil

	case "a":
		if m.view == ViewDefault && m.pane == PaneIssues {
			return m.openCreateIssueDefaultDialog(), nil
		}
		if m.view == ViewMeta {
			switch m.pane {
			case PaneRepos:
				return m.openCreateRepoDialog(), nil
			case PaneStreams:
				repoID, repoName, ok := m.selectedMetaRepo()
				if !ok {
					m.statusMsg = "Select or checkout a repo first"
					return m, nil
				}
				return m.openCreateStreamDialog(repoID, repoName), nil
			case PaneIssues:
				streamID, streamName, repoName, ok := m.selectedMetaStream()
				if !ok {
					m.statusMsg = "Select or checkout a stream first"
					return m, nil
				}
				return m.openCreateIssueMetaDialog(streamID, streamName, repoName), nil
			}
		}
		if m.pane == PaneScratchpads {
			return m.openCreateScratchpad(), nil
		}

	case "d":
		if m.pane == PaneScratchpads {
			rawIdx := m.filteredIndexAtCursor(PaneScratchpads)
			if rawIdx >= 0 && rawIdx < len(m.scratchpads) {
				return m.openConfirmDelete(m.scratchpads[rawIdx].ID), nil
			}
		}

	case "enter", "o":
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

	return m, nil
}

// ---------- View / pane cycling ----------

func nextView(current View, dir int) View {
	for i, v := range viewOrder {
		if v == current {
			next := (i + dir + len(viewOrder)) % len(viewOrder)
			return viewOrder[next]
		}
	}
	return current
}

func nextPane(view View, current Pane, dir int) Pane {
	panes := viewPanes[view]
	if len(panes) == 0 {
		return current
	}
	for i, p := range panes {
		if p == current {
			return panes[(i+dir+len(panes))%len(panes)]
		}
	}
	return panes[0]
}

// ---------- Checkout ----------

func (m Model) selectedIssue() (string, string, string, *string, bool) {
	if m.timer != nil && m.timer.State != "idle" && (m.view == ViewSession || m.view == ViewScratch) {
		if issue := m.activeIssueWithMeta(); issue != nil {
			return issue.ID, issue.StreamID, issue.Status, issue.TodoForDate, true
		}
	}
	if m.pane != PaneIssues {
		return "", "", "", nil, false
	}
	if m.view == ViewDefault {
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if rawIdx < 0 || rawIdx >= len(m.allIssues) {
			return "", "", "", nil, false
		}
		issue := m.allIssues[rawIdx]
		return issue.ID, issue.StreamID, issue.Status, issue.TodoForDate, true
	}
	if m.view == ViewDaily {
		rawIdx := m.filteredIndexAtCursor(PaneIssues)
		if m.dailySummary == nil || rawIdx < 0 || rawIdx >= len(m.dailySummary.Issues) {
			return "", "", "", nil, false
		}
		issue := m.dailySummary.Issues[rawIdx]
		return issue.ID, issue.StreamID, issue.Status, issue.TodoForDate, true
	}
	if m.view != ViewMeta {
		return "", "", "", nil, false
	}
	rawIdx := m.filteredIndexAtCursor(PaneIssues)
	if rawIdx < 0 || rawIdx >= len(m.issues) {
		return "", "", "", nil, false
	}
	issue := m.issues[rawIdx]
	return issue.ID, issue.StreamID, issue.Status, issue.TodoForDate, true
}

func (m Model) selectedMetaRepo() (string, string, bool) {
	if m.view != ViewMeta {
		return "", "", false
	}
	rawIdx := m.filteredIndexAtCursor(PaneRepos)
	if rawIdx >= 0 && rawIdx < len(m.repos) {
		return m.repos[rawIdx].ID, m.repos[rawIdx].Name, true
	}
	if m.context != nil && m.context.RepoID != nil {
		return *m.context.RepoID, firstNonEmpty(m.context.RepoName, m.context.RepoID), true
	}
	return "", "", false
}

func (m Model) selectedMetaStream() (string, string, string, bool) {
	if m.view != ViewMeta {
		return "", "", "", false
	}
	rawIdx := m.filteredIndexAtCursor(PaneStreams)
	if rawIdx >= 0 && rawIdx < len(m.streams) {
		stream := m.streams[rawIdx]
		repoName := ""
		for _, repo := range m.repos {
			if repo.ID == stream.RepoID {
				repoName = repo.Name
				break
			}
		}
		return stream.ID, stream.Name, repoName, true
	}
	if m.context != nil && m.context.StreamID != nil {
		repoName := "-"
		if m.context.RepoName != nil {
			repoName = *m.context.RepoName
		}
		return *m.context.StreamID, firstNonEmpty(m.context.StreamName, m.context.StreamID), repoName, true
	}
	return "", "", "", false
}

func (m Model) cycleSelectedIssueStatus() (tea.Model, tea.Cmd) {
	issueID, streamID, status, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	nextStatus := nextIssueStatus(status)
	if nextStatus == "" {
		return m, nil
	}
	if m.timer != nil && m.timer.State != "idle" && nextStatus == "done" {
		return m.openIssueSessionTransitionDialog(issueID, nextStatus), nil
	}
	return m, cmdChangeIssueStatus(m.client, issueID, nextStatus, streamID, m.currentDashboardDate())
}

func (m Model) abandonSelectedIssue() (tea.Model, tea.Cmd) {
	issueID, streamID, status, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	if status == "done" {
		m.statusMsg = "Done issues cannot be abandoned"
		return m, nil
	}
	if status == "abandoned" {
		return m, nil
	}
	if m.timer != nil && m.timer.State != "idle" {
		return m.openIssueSessionTransitionDialog(issueID, "abandoned"), nil
	}
	return m, cmdChangeIssueStatus(m.client, issueID, "abandoned", streamID, m.currentDashboardDate())
}

func (m Model) toggleSelectedIssueToday() (tea.Model, tea.Cmd) {
	issueID, streamID, _, todoForDate, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	return m, cmdToggleIssueToday(m.client, issueID, todoForDate != nil && *todoForDate != "", streamID, m.currentDashboardDate())
}

func (m Model) openSelectedIssueTodoDateDialog() Model {
	issueID, _, _, todoForDate, ok := m.selectedIssue()
	if !ok {
		return m
	}
	return m.openDatePickerDialog("", issueID, 0, todoForDate)
}

func nextIssueStatus(status string) string {
	switch status {
	case "todo":
		return "active"
	case "active":
		return "done"
	case "done":
		return "todo"
	case "abandoned":
		return "todo"
	default:
		return ""
	}
}

func (m Model) currentDashboardDate() string {
	if m.dashboardDate != "" {
		return m.dashboardDate
	}
	return time.Now().Format("2006-01-02")
}

func shiftISODate(date string, days int) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now().AddDate(0, 0, days).Format("2006-01-02")
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func (m Model) checkout() (tea.Model, tea.Cmd) {
	switch m.pane {

	case PaneRepos:
		rawIdx := m.filteredIndexAtCursor(PaneRepos)
		if rawIdx < 0 || rawIdx >= len(m.repos) {
			return m, nil
		}
		repo := m.repos[rawIdx]
		logger.Infof("checkout repo: %s (%s)", repo.Name, repo.ID)
		return m, cmdCheckoutRepo(m.client, repo.ID)

	case PaneStreams:
		rawIdx := m.filteredIndexAtCursor(PaneStreams)
		if rawIdx < 0 || rawIdx >= len(m.streams) {
			return m, nil
		}
		stream := m.streams[rawIdx]
		logger.Infof("checkout stream: %s (%s)", stream.Name, stream.ID)
		return m, cmdCheckoutStream(m.client, stream.ID)

	case PaneIssues:
		return m, nil
	}

	return m, nil
}
