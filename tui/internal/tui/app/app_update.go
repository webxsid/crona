package app

import (
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	"fmt"
	"strings"

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
		m.repos = msg.Repos
		m.clampFiltered(PaneRepos)
		return m, nil
	case streamsLoadedMsg:
		m.streams = msg.Streams
		m.clampFiltered(PaneStreams)
		return m, nil
	case issuesLoadedMsg:
		m.issues = msg.Issues
		m.clampFiltered(PaneIssues)
		return m, nil
	case habitsLoadedMsg:
		m.habits = msg.Habits
		m.clampFiltered(PaneHabits)
		return m, nil
	case allIssuesLoadedMsg:
		m.allIssues = msg.Issues
		if m.view == ViewDefault || m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
		}
		return m, nil
	case dueHabitsLoadedMsg:
		m.dueHabits = msg.Habits
		if m.view == ViewDaily {
			m.clampFiltered(PaneHabits)
		}
		return m, nil
	case dailySummaryLoadedMsg:
		m.dailySummary = msg.Summary
		if m.dashboardDate == "" && msg.Summary != nil {
			m.dashboardDate = msg.Summary.Date
		}
		if m.view == ViewDaily {
			m.clampFiltered(PaneIssues)
		}
		return m, nil
	case dailyCheckInLoadedMsg:
		m.dailyCheckIn = msg.CheckIn
		if m.dailyCheckIn != nil && m.wellbeingDate == "" {
			m.wellbeingDate = m.dailyCheckIn.Date
		}
		return m, nil
	case metricsRangeLoadedMsg:
		m.metricsRange = msg.Days
		return m, nil
	case metricsRollupLoadedMsg:
		m.metricsRollup = msg.Rollup
		return m, nil
	case streaksLoadedMsg:
		m.streaks = msg.Streaks
		return m, nil
	case exportAssetsLoadedMsg:
		m.exportAssets = msg.Assets
		m.clampFiltered(PaneConfig)
		return m, loadExportReports(m.client)
	case exportReportsLoadedMsg:
		m.exportReports = msg.Reports
		m.clampFiltered(PaneExportReports)
		return m, nil
	case exportReportDeletedMsg:
		return m, tea.Batch(m.setStatus("Deleted report "+msg.Name, false), loadExportReports(m.client))
	case issueSessionsLoadedMsg:
		var activeIssueID int64
		if m.context != nil && m.context.IssueID != nil {
			activeIssueID = *m.context.IssueID
		} else if m.timer != nil && m.timer.IssueID != nil {
			activeIssueID = *m.timer.IssueID
		}
		if msg.IssueID == activeIssueID {
			m.issueSessions = msg.Sessions
		}
		return m, nil
	case sessionHistoryLoadedMsg:
		m.sessionHistory = msg.Sessions
		m.clampFiltered(PaneSessions)
		return m, nil
	case sessionDetailLoadedMsg:
		m.sessionDetail = msg.Detail
		m.sessionDetailY = 0
		if msg.Detail == nil {
			m.sessionDetailOpen = false
			return m, m.setStatus("Session detail is unavailable", true)
		}
		return m, nil
	case sessionDetailFailedMsg:
		m.sessionDetailOpen = false
		m.sessionDetail = nil
		m.sessionDetailY = 0
		logger.Errorf("session detail failed: %v", msg.Err)
		return m, m.setStatus("Error: "+msg.Err.Error(), true)
	case scratchpadsLoadedMsg:
		m.scratchpads = msg.Pads
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
		m.stashes = msg.Stashes
		if m.dialogStashCursor >= len(m.stashes) {
			if len(m.stashes) == 0 {
				m.dialogStashCursor = 0
			} else {
				m.dialogStashCursor = len(m.stashes) - 1
			}
		}
		return m, nil
	case opsLoadedMsg:
		m.ops = msg.Ops
		m.clampFiltered(PaneOps)
		return m, nil
	case contextLoadedMsg:
		m.context = msg.Ctx
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
		m.timer = msg.Timer
		m.elapsed = 0
		m.timerTickSeq++
		if m.timer != nil && m.timer.State != "idle" {
			if m.view != ViewScratch && m.view != ViewSessionHistory {
				m.view = ViewSessionActive
			}
			m.pane = viewDefaultPane[m.view]
		} else if m.view == ViewSessionActive {
			m.view = ViewDaily
			m.pane = viewDefaultPane[m.view]
		}
		historyCmd := loadSessionHistoryForModel(m, 200)
		if m.timer != nil && m.timer.IssueID != nil {
			if m.context == nil || m.context.IssueID == nil || *m.context.IssueID != *m.timer.IssueID {
				return m, tea.Batch(loadIssueSessions(m.client, *m.timer.IssueID), historyCmd, tickAfter(m.timerTickSeq))
			}
		}
		if m.timer != nil && m.timer.State != "idle" {
			return m, tea.Batch(historyCmd, tickAfter(m.timerTickSeq))
		}
		return m, historyCmd
	case healthLoadedMsg:
		m.health = msg.Health
		return m, nil
	case updateStatusLoadedMsg:
		m.updateChecking = false
		m.updateStatus = msg.Status
		return m, nil
	case updateDismissedMsg:
		m.updateStatus = msg.Status
		m.updateChecking = false
		if strings.TrimSpace(msg.Status.DismissedVersion) == "" {
			return m, m.setStatus("No update prompt dismissed", false)
		}
		return m, m.setStatus("Update prompt dismissed for v"+msg.Status.DismissedVersion, false)
	case updateInstallFinishedMsg:
		m.updateInstalling = false
		m.updateInstallOutput = strings.TrimSpace(msg.Output)
		if msg.Err != nil {
			m.updateInstallError = msg.Err.Error()
			return m, nil
		}
		close(m.eventStop)
		return m, tea.Quit
	case settingsLoadedMsg:
		m.settings = msg.Settings
		m.clampFiltered(PaneSettings)
		return m, nil
	case kernelInfoLoadedMsg:
		m.kernelInfo = msg.Info
		return m, nil
	case healthTickMsg:
		return m, tea.Batch(loadHealth(m.client), healthTickAfter())
	case clearStatusMsg:
		if msg.Seq != m.statusSeq {
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
		return m, tea.Batch(cmd, loadKernelInfo(m.client), loadRepos(m.client), loadAllIssues(m.client), loadDueHabits(m.client, m.currentDashboardDate()), loadDailySummary(m.client, m.dashboardDate), loadWellbeing(m.client, m.currentWellbeingDate()), loadSessionHistoryForModel(m, 200), loadScratchpads(m.client), loadStashes(m.client), loadOps(m.client, m.currentOpsLimit()), loadContext(m.client), loadTimer(m.client), loadUpdateStatus(m.client), loadSettings(m.client))
	case devClearedMsg:
		cmd := m.setStatus("Dev data cleared", false)
		m.view = ViewDaily
		m.pane = viewDefaultPane[m.view]
		return m, tea.Batch(cmd, loadKernelInfo(m.client), loadRepos(m.client), loadAllIssues(m.client), loadDueHabits(m.client, m.currentDashboardDate()), loadDailySummary(m.client, m.dashboardDate), loadWellbeing(m.client, m.currentWellbeingDate()), loadSessionHistoryForModel(m, 200), loadScratchpads(m.client), loadStashes(m.client), loadOps(m.client, m.currentOpsLimit()), loadContext(m.client), loadTimer(m.client), loadUpdateStatus(m.client), loadSettings(m.client))
	case sessionAmendedMsg:
		cmd := m.setStatus("Session amended", false)
		return m, tea.Batch(cmd, loadSessionHistoryForModel(m, 200), loadSessionDetail(m.client, msg.ID))
	case focusSessionChangedMsg:
		cmds := []tea.Cmd{}
		if msg.ReloadContext {
			cmds = append(cmds, loadContext(m.client))
		}
		if msg.ReloadTimer {
			cmds = append(cmds, loadTimer(m.client))
		}
		if len(cmds) == 0 {
			return m, nil
		}
		return m, tea.Batch(cmds...)
	case timerTickMsg:
		if msg.Seq != m.timerTickSeq {
			return m, nil
		}
		if m.timer != nil && m.timer.State != "idle" {
			m.elapsed++
			return m, tickAfter(m.timerTickSeq)
		}
		return m, nil
	case kernelEventMsg:
		updated, cmd := handleKernelEvent(m, msg.Event)
		return updated, tea.Batch(cmd, waitForEvent(eventChannel))
	case errMsg:
		if (m.dialog == "export_report" || m.dialog == "export_calendar_repo") && m.dialogProcessing {
			m.dialog = ""
			m.dialogChoiceItems = nil
			m.dialogChoiceCursor = 0
			m.dialogProcessing = false
			m.dialogProcessingLabel = ""
		}
		logger.Errorf("update error: %v", msg.Err)
		return m, m.setStatus("Error: "+msg.Err.Error(), true)
	case openScratchpadMsg:
		return m.enterScratchpadPane(msg), nil
	case scratchpadReloadedMsg:
		m.scratchpadFilePath = msg.FilePath
		m.scratchpadRendered = msg.Rendered
		m.scratchpadViewport.SetContent(msg.Rendered)
		return m, nil
	case editorDoneMsg:
		cmds := []tea.Cmd{loadScratchpads(m.client), loadExportAssets(m.client)}
		if m.scratchpadOpen && m.scratchpadMeta != nil {
			cmds = append(cmds, cmdReloadScratchpad(m.client, m.scratchpadMeta, m.scratchpadRenderWidth()))
		}
		return m, tea.Batch(cmds...)
	case dailyReportGeneratedMsg:
		m.exportAssets = &msg.Result.Assets
		if (m.dialog == "export_report" || m.dialog == "export_calendar_repo") && m.dialogProcessing {
			m.dialog = ""
			m.dialogChoiceItems = nil
			m.dialogChoiceCursor = 0
			m.dialogProcessing = false
			m.dialogProcessingLabel = ""
		}
		if msg.Result.OutputMode == "file" && msg.Result.FilePath != nil {
			label := msg.Result.Label
			if strings.TrimSpace(label) == "" {
				label = "Report"
			}
			return m, tea.Batch(m.setStatus(label+" written to "+*msg.Result.FilePath, false), loadExportReports(m.client))
		}
		return m, nil
	case calendarExportGeneratedMsg:
		m.exportAssets = &msg.Result.Assets
		if (m.dialog == "export_report" || m.dialog == "export_calendar_repo") && m.dialogProcessing {
			m.dialog = ""
			m.dialogChoiceItems = nil
			m.dialogChoiceCursor = 0
			m.dialogProcessing = false
			m.dialogProcessingLabel = ""
		}
		title := "Calendar Export"
		name := strings.TrimSpace(msg.Result.RepoName)
		if name == "" {
			name = "Repo"
		}
		meta := fmt.Sprintf("Issues %s   Sessions %s", msg.Result.IssuesFilePath, msg.Result.SessionsFilePath)
		body := strings.Join([]string{
			"Issues ICS",
			msg.Result.IssuesFilePath,
			"",
			"Sessions ICS",
			msg.Result.SessionsFilePath,
		}, "\n")
		m = m.openViewEntityDialog(title, name, meta, body)
		return m, m.setStatus("Calendar export written", false)
	case clipboardCopiedMsg:
		if (m.dialog == "export_report" || m.dialog == "export_calendar_repo") && m.dialogProcessing {
			m.dialog = ""
			m.dialogChoiceItems = nil
			m.dialogChoiceCursor = 0
			m.dialogProcessing = false
			m.dialogProcessingLabel = ""
		}
		return m, m.setStatus(msg.Message, false)
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
