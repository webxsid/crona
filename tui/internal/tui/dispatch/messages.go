package dispatch

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui/commands"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type MessageState struct {
	Width                 int
	Height                int
	View                  uistate.View
	Pane                  uistate.Pane
	Cursor                map[uistate.Pane]int
	Repos                 []api.Repo
	Streams               []api.Stream
	Issues                []api.Issue
	Habits                []api.Habit
	AllIssues             []api.IssueWithMeta
	DueHabits             []api.HabitDailyItem
	DailySummary          *api.DailyIssueSummary
	DailyPlan             *api.DailyPlan
	DashboardDate         string
	RollupStartDate       string
	RollupEndDate         string
	WellbeingDate         string
	DailyCheckIn          *api.DailyCheckIn
	MetricsRange          []api.DailyMetricsDay
	MetricsRollup         *api.MetricsRollup
	Streaks               *api.StreakSummary
	DashboardWindow       *api.DashboardWindowSummary
	DailyFocusScore       *api.FocusScoreSummary
	WeeklyFocusScore      *api.FocusScoreSummary
	RepoDistribution      *api.TimeDistributionSummary
	StreamDistribution    *api.TimeDistributionSummary
	IssueDistribution     *api.TimeDistributionSummary
	SegmentDistribution   *api.TimeDistributionSummary
	GoalProgress          *api.GoalProgressSummary
	ExportAssets          *api.ExportAssetStatus
	ExportReports         []api.ExportReportFile
	IssueSessions         []api.Session
	SessionHistory        []api.SessionHistoryEntry
	SessionDetail         *api.SessionDetail
	SessionDetailOpen     bool
	SessionDetailY        int
	SessionContextOpen    bool
	SessionContextY       int
	Scratchpads           []api.ScratchPad
	Stashes               []api.Stash
	DialogStashCursor     int
	Ops                   []api.Op
	Context               *api.ActiveContext
	Timer                 *api.TimerState
	Health                *api.Health
	UpdateStatus          *api.UpdateStatus
	UpdateChecking        bool
	UpdateInstalling      bool
	UpdateInstallPhase    string
	UpdateInstallDetail   string
	UpdateInstallOutput   string
	UpdateInstallError    string
	UpdateInstallProgress <-chan tea.Msg
	Settings              *api.CoreSettings
	KernelInfo            *api.KernelInfo
	Elapsed               int
	TimerTickSeq          int
	ScratchpadOpen        bool
	ScratchpadMeta        *api.ScratchPad
	ScratchpadFilePath    string
	ScratchpadRendered    string
	StatusMsg             string
	StatusSeq             int
	StatusErr             bool
	Dialog                string
	DialogErrorMessage    string
	DialogChoiceItems     []string
	DialogChoiceCursor    int
	DialogProcessing      bool
	DialogProcessingLabel string
	OpsLimit              int
	OpsLimitPinned        bool
}

type MessageDeps struct {
	DefaultOpsLimit            func(MessageState) int
	CurrentOpsLimit            func(MessageState) int
	ClampFiltered              func(*MessageState, uistate.Pane)
	SyncScratchpadViewport     func(*MessageState)
	ScratchpadTabIndexByID     func(*MessageState, string) int
	FilteredCursorForRawIndex  func(*MessageState, uistate.Pane, int) int
	SetActiveScratchpadByIndex func(*MessageState, int)
	SetStatus                  func(*MessageState, string, bool) tea.Cmd
	OpenViewEntityDialog       func(*MessageState, string, string, string, string)
	EnterScratchpadPane        func(*MessageState, commands.OpenScratchpadMsg)
	SetScratchpadContent       func(*MessageState, string, string)
	CurrentDashboardDate       func(MessageState) string
	CurrentWellbeingDate       func(MessageState) string
	LoadRepos                  func() tea.Cmd
	LoadAllIssues              func() tea.Cmd
	LoadStreams                func(int64) tea.Cmd
	LoadIssues                 func(int64) tea.Cmd
	LoadHabits                 func(int64) tea.Cmd
	LoadDueHabits              func(string) tea.Cmd
	LoadDailySummary           func(string) tea.Cmd
	LoadWellbeing              func(string) tea.Cmd
	LoadRollupSummaries        func(string, string) tea.Cmd
	LoadDailyPlan              func(string) tea.Cmd
	LoadExportAssets           func() tea.Cmd
	LoadExportReports          func() tea.Cmd
	LoadIssueSessions          func(int64) tea.Cmd
	LoadSessionHistoryFor200   func(MessageState) tea.Cmd
	LoadSessionDetail          func(string) tea.Cmd
	LoadScratchpads            func() tea.Cmd
	LoadStashes                func() tea.Cmd
	LoadOps                    func(int) tea.Cmd
	LoadContext                func() tea.Cmd
	LoadTimer                  func() tea.Cmd
	LoadHealth                 func() tea.Cmd
	LoadUpdateStatus           func() tea.Cmd
	LoadSettings               func() tea.Cmd
	LoadKernelInfo             func() tea.Cmd
	HealthTickAfter            func() tea.Cmd
	TickAfter                  func(int) tea.Cmd
	WaitForEvent               func() tea.Cmd
	HandleKernelEvent          func(MessageState, api.KernelEvent) (MessageState, tea.Cmd)
	CloseEventStop             func()
}

func HandleMessage(state MessageState, raw tea.Msg, deps MessageDeps) (MessageState, tea.Cmd, bool) {
	switch msg := raw.(type) {
	case tea.WindowSizeMsg:
		state.Width = msg.Width
		state.Height = msg.Height
		if !state.OpsLimitPinned {
			nextLimit := deps.DefaultOpsLimit(state)
			if nextLimit != state.OpsLimit {
				state.OpsLimit = nextLimit
				if state.ScratchpadOpen {
					deps.SyncScratchpadViewport(&state)
				}
				return state, deps.LoadOps(deps.CurrentOpsLimit(state)), true
			}
		}
		if state.ScratchpadOpen {
			deps.SyncScratchpadViewport(&state)
		}
		return state, nil, true
	case commands.ReposLoadedMsg:
		state.Repos = msg.Repos
		deps.ClampFiltered(&state, uistate.PaneRepos)
		return state, nil, true
	case commands.StreamsLoadedMsg:
		state.Streams = msg.Streams
		deps.ClampFiltered(&state, uistate.PaneStreams)
		return state, nil, true
	case commands.IssuesLoadedMsg:
		state.Issues = msg.Issues
		deps.ClampFiltered(&state, uistate.PaneIssues)
		return state, nil, true
	case commands.HabitsLoadedMsg:
		state.Habits = msg.Habits
		deps.ClampFiltered(&state, uistate.PaneHabits)
		return state, nil, true
	case commands.AllIssuesLoadedMsg:
		state.AllIssues = msg.Issues
		if state.View == uistate.ViewDefault || state.View == uistate.ViewDaily {
			deps.ClampFiltered(&state, uistate.PaneIssues)
		}
		return state, nil, true
	case commands.DueHabitsLoadedMsg:
		state.DueHabits = msg.Habits
		if state.View == uistate.ViewDaily {
			deps.ClampFiltered(&state, uistate.PaneHabits)
		}
		return state, nil, true
	case commands.DailySummaryLoadedMsg:
		state.DailySummary = msg.Summary
		if state.DashboardDate == "" && msg.Summary != nil {
			state.DashboardDate = msg.Summary.Date
		}
		if state.View == uistate.ViewDaily {
			deps.ClampFiltered(&state, uistate.PaneIssues)
		}
		return state, nil, true
	case commands.DailyPlanLoadedMsg:
		state.DailyPlan = msg.Plan
		return state, nil, true
	case commands.DailyCheckInLoadedMsg:
		state.DailyCheckIn = msg.CheckIn
		if state.DailyCheckIn != nil && state.WellbeingDate == "" {
			state.WellbeingDate = state.DailyCheckIn.Date
		}
		return state, nil, true
	case commands.MetricsRangeLoadedMsg:
		state.MetricsRange = msg.Days
		return state, nil, true
	case commands.MetricsRollupLoadedMsg:
		state.MetricsRollup = msg.Rollup
		return state, nil, true
	case commands.StreaksLoadedMsg:
		state.Streaks = msg.Streaks
		return state, nil, true
	case commands.DashboardWindowLoadedMsg:
		state.DashboardWindow = msg.Summary
		return state, nil, true
	case commands.FocusScoreLoadedMsg:
		switch msg.WindowDays {
		case 1:
			state.DailyFocusScore = msg.Summary
		default:
			state.WeeklyFocusScore = msg.Summary
		}
		return state, nil, true
	case commands.DistributionLoadedMsg:
		switch msg.GroupBy {
		case "repo":
			state.RepoDistribution = msg.Summary
		case "stream":
			state.StreamDistribution = msg.Summary
		case "issue":
			state.IssueDistribution = msg.Summary
		case "segment_type":
			state.SegmentDistribution = msg.Summary
		}
		return state, nil, true
	case commands.GoalProgressLoadedMsg:
		state.GoalProgress = msg.Summary
		return state, nil, true
	case commands.RollupRangeChangedMsg:
		state.RollupStartDate = msg.Start
		state.RollupEndDate = msg.End
		state.Cursor[uistate.PaneRollupDays] = 0
		return state, deps.LoadRollupSummaries(msg.Start, msg.End), true
	case commands.ExportAssetsLoadedMsg:
		state.ExportAssets = msg.Assets
		deps.ClampFiltered(&state, uistate.PaneConfig)
		return state, deps.LoadExportReports(), true
	case commands.ExportReportsLoadedMsg:
		state.ExportReports = msg.Reports
		deps.ClampFiltered(&state, uistate.PaneExportReports)
		return state, nil, true
	case commands.ExportReportDeletedMsg:
		return state, tea.Batch(deps.SetStatus(&state, "Deleted report "+msg.Name, false), deps.LoadExportReports()), true
	case commands.IssueSessionsLoadedMsg:
		var activeIssueID int64
		if state.Context != nil && state.Context.IssueID != nil {
			activeIssueID = *state.Context.IssueID
		} else if state.Timer != nil && state.Timer.IssueID != nil {
			activeIssueID = *state.Timer.IssueID
		}
		if msg.IssueID == activeIssueID {
			state.IssueSessions = msg.Sessions
		}
		return state, nil, true
	case commands.SessionHistoryLoadedMsg:
		state.SessionHistory = msg.Sessions
		deps.ClampFiltered(&state, uistate.PaneSessions)
		return state, nil, true
	case commands.SessionDetailLoadedMsg:
		state.SessionDetail = msg.Detail
		state.SessionDetailY = 0
		if msg.Detail == nil {
			state.SessionDetailOpen = false
			return state, deps.SetStatus(&state, "Session detail is unavailable", true), true
		}
		return state, nil, true
	case commands.SessionDetailFailedMsg:
		state.SessionDetailOpen = false
		state.SessionDetail = nil
		state.SessionDetailY = 0
		logger.Errorf("session detail failed: %v", msg.Err)
		return state, deps.SetStatus(&state, "Error: "+msg.Err.Error(), true), true
	case commands.ScratchpadsLoadedMsg:
		state.Scratchpads = msg.Pads
		deps.ClampFiltered(&state, uistate.PaneScratchpads)
		if state.ScratchpadMeta != nil {
			if idx := deps.ScratchpadTabIndexByID(&state, state.ScratchpadMeta.ID); idx >= 0 {
				if filteredCur := deps.FilteredCursorForRawIndex(&state, uistate.PaneScratchpads, idx); filteredCur >= 0 {
					state.Cursor[uistate.PaneScratchpads] = filteredCur
				}
				deps.SetActiveScratchpadByIndex(&state, idx)
			} else {
				state.ScratchpadOpen = false
				state.ScratchpadMeta = nil
				state.ScratchpadFilePath = ""
				state.ScratchpadRendered = ""
			}
		}
		return state, nil, true
	case commands.StashesLoadedMsg:
		state.Stashes = msg.Stashes
		if state.DialogStashCursor >= len(state.Stashes) {
			if len(state.Stashes) == 0 {
				state.DialogStashCursor = 0
			} else {
				state.DialogStashCursor = len(state.Stashes) - 1
			}
		}
		return state, nil, true
	case commands.OpsLoadedMsg:
		state.Ops = msg.Ops
		deps.ClampFiltered(&state, uistate.PaneOps)
		return state, nil, true
	case commands.ContextLoadedMsg:
		state.Context = msg.Ctx
		if state.View == uistate.ViewDefault || state.View == uistate.ViewDaily {
			deps.ClampFiltered(&state, uistate.PaneIssues)
			deps.ClampFiltered(&state, uistate.PaneHabits)
		}
		cmds := []tea.Cmd{deps.LoadRollupSummaries(state.RollupStartDate, state.RollupEndDate)}
		if state.Context != nil && state.Context.IssueID != nil {
			cmds = append(cmds, deps.LoadIssueSessions(*state.Context.IssueID))
			return state, tea.Batch(cmds...), true
		}
		state.IssueSessions = nil
		return state, tea.Batch(cmds...), true
	case commands.TimerLoadedMsg:
		state.Timer = msg.Timer
		state.Elapsed = 0
		state.TimerTickSeq++
		if state.Timer != nil && state.Timer.State != "idle" {
			if state.View != uistate.ViewScratch && state.View != uistate.ViewSessionHistory {
				state.View = uistate.ViewSessionActive
			}
			state.Pane = uistate.DefaultPane(state.View)
		} else if state.View == uistate.ViewSessionActive {
			state.View = uistate.ViewDaily
			state.Pane = uistate.DefaultPane(state.View)
		}
		historyCmd := deps.LoadSessionHistoryFor200(state)
		if state.Timer != nil && state.Timer.IssueID != nil {
			if state.Context == nil || state.Context.IssueID == nil || *state.Context.IssueID != *state.Timer.IssueID {
				return state, tea.Batch(deps.LoadIssueSessions(*state.Timer.IssueID), historyCmd, deps.TickAfter(state.TimerTickSeq)), true
			}
		}
		if state.Timer != nil && state.Timer.State != "idle" {
			return state, tea.Batch(historyCmd, deps.TickAfter(state.TimerTickSeq)), true
		}
		return state, historyCmd, true
	case commands.HealthLoadedMsg:
		state.Health = msg.Health
		return state, nil, true
	case commands.UpdateStatusLoadedMsg:
		state.UpdateChecking = false
		state.UpdateStatus = msg.Status
		return state, nil, true
	case commands.UpdateDismissedMsg:
		state.UpdateStatus = msg.Status
		state.UpdateChecking = false
		if strings.TrimSpace(msg.Status.DismissedVersion) == "" {
			return state, deps.SetStatus(&state, "No update prompt dismissed", false), true
		}
		return state, deps.SetStatus(&state, "Update prompt dismissed for v"+msg.Status.DismissedVersion, false), true
	case commands.UpdateInstallStartedMsg:
		state.View = uistate.ViewUpdates
		state.Pane = uistate.DefaultPane(state.View)
		state.UpdateInstalling = true
		state.UpdateInstallPhase = "starting"
		state.UpdateInstallDetail = "Preparing update..."
		state.UpdateInstallOutput = ""
		state.UpdateInstallError = ""
		state.UpdateInstallProgress = msg.Progress
		deps.CloseEventStop()
		return state, commands.WaitForUpdateInstall(msg.Progress), true
	case commands.UpdateInstallPhaseMsg:
		state.View = uistate.ViewUpdates
		state.Pane = uistate.DefaultPane(state.View)
		state.UpdateInstalling = true
		if strings.TrimSpace(msg.Phase) != "" {
			state.UpdateInstallPhase = strings.TrimSpace(msg.Phase)
		}
		state.UpdateInstallDetail = strings.TrimSpace(msg.Detail)
		if chunk := strings.TrimSpace(msg.Output); chunk != "" {
			if strings.TrimSpace(state.UpdateInstallOutput) == "" {
				state.UpdateInstallOutput = chunk
			} else {
				state.UpdateInstallOutput = strings.TrimSpace(state.UpdateInstallOutput) + "\n" + chunk
			}
		}
		if state.UpdateInstallProgress != nil {
			return state, commands.WaitForUpdateInstall(state.UpdateInstallProgress), true
		}
		return state, nil, true
	case commands.UpdateInstallFinishedMsg:
		state.UpdateInstalling = false
		state.UpdateInstallOutput = strings.TrimSpace(msg.Output)
		state.UpdateInstallProgress = nil
		if msg.Err != nil {
			state.View = uistate.ViewUpdates
			state.Pane = uistate.DefaultPane(state.View)
			state.UpdateInstallPhase = "failed"
			state.UpdateInstallDetail = "Update failed"
			state.UpdateInstallError = msg.Err.Error()
			return state, nil, true
		}
		state.UpdateInstallPhase = "relaunching"
		state.UpdateInstallDetail = "Relaunching Crona..."
		deps.CloseEventStop()
		return state, tea.Quit, true
	case commands.SettingsLoadedMsg:
		state.Settings = msg.Settings
		deps.ClampFiltered(&state, uistate.PaneSettings)
		return state, nil, true
	case commands.KernelInfoLoadedMsg:
		state.KernelInfo = msg.Info
		return state, nil, true
	case commands.HealthTickMsg:
		if state.UpdateInstalling {
			return state, nil, true
		}
		return state, tea.Batch(deps.LoadHealth(), deps.HealthTickAfter()), true
	case commands.ClearStatusMsg:
		if msg.Seq != state.StatusSeq {
			return state, nil, true
		}
		state.StatusMsg = ""
		state.StatusErr = false
		return state, nil, true
	case commands.KernelShutdownMsg:
		deps.CloseEventStop()
		return state, tea.Quit, true
	case commands.DevSeededMsg:
		cmd := deps.SetStatus(&state, "Dev seed loaded", false)
		state.View = uistate.ViewDaily
		state.Pane = uistate.DefaultPane(state.View)
		return state, tea.Batch(cmd, deps.LoadKernelInfo(), deps.LoadRepos(), deps.LoadAllIssues(), deps.LoadDueHabits(deps.CurrentDashboardDate(state)), deps.LoadDailySummary(state.DashboardDate), deps.LoadWellbeing(deps.CurrentWellbeingDate(state)), deps.LoadRollupSummaries(state.RollupStartDate, state.RollupEndDate), deps.LoadSessionHistoryFor200(state), deps.LoadScratchpads(), deps.LoadStashes(), deps.LoadOps(deps.CurrentOpsLimit(state)), deps.LoadContext(), deps.LoadTimer(), deps.LoadUpdateStatus(), deps.LoadSettings()), true
	case commands.DevClearedMsg:
		cmd := deps.SetStatus(&state, "Dev data cleared", false)
		state.View = uistate.ViewDaily
		state.Pane = uistate.DefaultPane(state.View)
		return state, tea.Batch(cmd, deps.LoadKernelInfo(), deps.LoadRepos(), deps.LoadAllIssues(), deps.LoadDueHabits(deps.CurrentDashboardDate(state)), deps.LoadDailySummary(state.DashboardDate), deps.LoadWellbeing(deps.CurrentWellbeingDate(state)), deps.LoadRollupSummaries(state.RollupStartDate, state.RollupEndDate), deps.LoadSessionHistoryFor200(state), deps.LoadScratchpads(), deps.LoadStashes(), deps.LoadOps(deps.CurrentOpsLimit(state)), deps.LoadContext(), deps.LoadTimer(), deps.LoadUpdateStatus(), deps.LoadSettings()), true
	case commands.SessionAmendedMsg:
		cmd := deps.SetStatus(&state, "Session amended", false)
		return state, tea.Batch(cmd, deps.LoadSessionHistoryFor200(state), deps.LoadSessionDetail(msg.ID)), true
	case commands.ManualSessionLoggedMsg:
		cmds := []tea.Cmd{deps.SetStatus(&state, "Manual session logged", false), deps.LoadSessionHistoryFor200(state), deps.LoadAllIssues()}
		if state.Context != nil && state.Context.IssueID != nil && *state.Context.IssueID == msg.IssueID {
			cmds = append(cmds, deps.LoadIssueSessions(msg.IssueID))
		}
		if deps.CurrentDashboardDate(state) == strings.TrimSpace(msg.Date) {
			cmds = append(cmds, deps.LoadDailySummary(msg.Date))
		}
		if deps.CurrentWellbeingDate(state) == strings.TrimSpace(msg.Date) {
			cmds = append(cmds, deps.LoadWellbeing(msg.Date))
		}
		return state, tea.Batch(cmds...), true
	case commands.FocusSessionChangedMsg:
		cmds := []tea.Cmd{}
		if msg.ReloadContext {
			cmds = append(cmds, deps.LoadContext())
		}
		if msg.ReloadTimer {
			cmds = append(cmds, deps.LoadTimer())
		}
		if len(cmds) == 0 {
			return state, nil, true
		}
		return state, tea.Batch(cmds...), true
	case commands.TimerTickMsg:
		if msg.Seq != state.TimerTickSeq {
			return state, nil, true
		}
		if state.Timer != nil && state.Timer.State != "idle" {
			state.Elapsed++
			return state, deps.TickAfter(state.TimerTickSeq), true
		}
		return state, nil, true
	case commands.KernelEventMsg:
		if state.UpdateInstalling {
			return state, nil, true
		}
		updated, cmd := deps.HandleKernelEvent(state, msg.Event)
		return updated, tea.Batch(cmd, deps.WaitForEvent()), true
	case commands.ErrMsg:
		if shouldSuppressUpdateInstallError(state, msg.Err) {
			logger.Infof("suppressing expected update handoff error: %v", msg.Err)
			return state, nil, true
		}
		if state.Dialog != "" {
			if isExportDialog(state.Dialog) && state.DialogProcessing {
				state.DialogProcessing = false
				state.DialogProcessingLabel = ""
			}
			state.DialogErrorMessage = "Error: " + msg.Err.Error()
			logger.Errorf("update error: %v", msg.Err)
			return state, nil, true
		}
		logger.Errorf("update error: %v", msg.Err)
		return state, deps.SetStatus(&state, "Error: "+msg.Err.Error(), true), true
	case commands.OpenScratchpadMsg:
		deps.EnterScratchpadPane(&state, msg)
		return state, nil, true
	case commands.ScratchpadReloadedMsg:
		deps.SetScratchpadContent(&state, msg.Rendered, msg.FilePath)
		return state, nil, true
	case commands.EditorDoneMsg:
		cmds := []tea.Cmd{deps.LoadScratchpads(), deps.LoadExportAssets()}
		return state, tea.Batch(cmds...), true
	case commands.DailyReportGeneratedMsg:
		state.ExportAssets = &msg.Result.Assets
		if isExportDialog(state.Dialog) && state.DialogProcessing {
			state.Dialog = ""
			state.DialogChoiceItems = nil
			state.DialogChoiceCursor = 0
			state.DialogProcessing = false
			state.DialogProcessingLabel = ""
		}
		if msg.Result.OutputMode == "file" && msg.Result.FilePath != nil {
			label := msg.Result.Label
			if strings.TrimSpace(label) == "" {
				label = "Report"
			}
			return state, tea.Batch(deps.SetStatus(&state, label+" written to "+*msg.Result.FilePath, false), deps.LoadExportReports()), true
		}
		return state, nil, true
	case commands.CalendarExportGeneratedMsg:
		state.ExportAssets = &msg.Result.Assets
		if isExportDialog(state.Dialog) && state.DialogProcessing {
			state.Dialog = ""
			state.DialogChoiceItems = nil
			state.DialogChoiceCursor = 0
			state.DialogProcessing = false
			state.DialogProcessingLabel = ""
		}
		title := "Calendar Export"
		name := strings.TrimSpace(msg.Result.RepoName)
		if name == "" {
			name = "Repo"
		}
		meta := fmt.Sprintf("Issues %s   Sessions %s", msg.Result.IssuesFilePath, msg.Result.SessionsFilePath)
		body := strings.Join([]string{"Issues ICS", msg.Result.IssuesFilePath, "", "Sessions ICS", msg.Result.SessionsFilePath}, "\n")
		deps.OpenViewEntityDialog(&state, title, name, meta, body)
		return state, deps.SetStatus(&state, "Calendar export written", false), true
	case commands.ClipboardCopiedMsg:
		if isExportDialog(state.Dialog) && state.DialogProcessing {
			state.Dialog = ""
			state.DialogChoiceItems = nil
			state.DialogChoiceCursor = 0
			state.DialogProcessing = false
			state.DialogProcessingLabel = ""
		}
		return state, deps.SetStatus(&state, msg.Message, false), true
	default:
		return state, nil, false
	}
}

func shouldSuppressUpdateInstallError(state MessageState, err error) bool {
	if !state.UpdateInstalling || err == nil {
		return false
	}
	return isExpectedUpdateTransportError(err.Error())
}

func isExpectedUpdateTransportError(raw string) bool {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return false
	}
	for _, fragment := range []string{
		"dial unix",
		"connect: connection refused",
		"no such file or directory",
		"broken pipe",
		"socket",
		"deadline exceeded",
		"eof",
	} {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func isExportDialog(kind string) bool {
	switch kind {
	case "export_report", "export_preset", "export_calendar_repo":
		return true
	default:
		return false
	}
}
