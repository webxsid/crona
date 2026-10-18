package dispatch

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	"crona/tui/internal/tui/commands"
	"crona/tui/internal/tui/scope"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type MessageState struct {
	Width               int
	Height              int
	View                uistate.View
	Pane                uistate.Pane
	DailyTaskSection    uistate.DailyTaskSection
	DefaultIssueSection uistate.DefaultIssueSection
	BrowseScope         scope.Scope
	// Date is the shared date every dated view shows; "" follows the logical today.
	Date                    string
	RollupWindowDays        int
	Cursor                  map[uistate.Pane]int
	Filters                 map[uistate.Pane]string
	Repos                   []api.Repo
	Streams                 []api.Stream
	AllStreams              []api.Stream
	Issues                  []api.Issue
	Habits                  []api.Habit
	AllHabits               []api.HabitWithMeta
	HabitStreakDefs         []api.HabitStreakDefinition
	MomentumCards           []api.MomentumCard
	AllIssues               []api.IssueWithMeta
	DueHabits               []api.HabitDailyItem
	DailySummary            *api.DailyIssueSummary
	DailyPlan               *api.DailyPlan
	MomentumWindowDays      int
	MomentumTab             string
	MomentumHistoryY        int
	WellbeingWindowDays     int
	DailyCheckIn            *api.DailyCheckIn
	MetricsRange            []api.DailyMetricsDay
	MetricsRollup           *api.MetricsRollup
	RollupMetricsRange      []api.DailyMetricsDay
	RollupMetricsRollup     *api.MetricsRollup
	MomentumMetricsRange    []api.DailyMetricsDay
	MomentumMetricsRollup   *api.MetricsRollup
	Streaks                 *api.StreakSummary
	DailyStreaks            *api.StreakSummary
	DashboardWindow         *api.DashboardWindowSummary
	DailyFocusScore         *api.FocusScoreSummary
	WeeklyFocusScore        *api.FocusScoreSummary
	RepoDistribution        *api.TimeDistributionSummary
	StreamDistribution      *api.TimeDistributionSummary
	IssueDistribution       *api.TimeDistributionSummary
	SegmentDistribution     *api.TimeDistributionSummary
	GoalProgress            *api.GoalProgressSummary
	ExportAssets            *api.ExportAssetStatus
	ExportReports           []api.ExportReportFile
	IssueSessions           []api.Session
	SessionHistory          []api.SessionHistoryEntry
	HabitHistory            []api.HabitCompletion
	SelectedHabitHistoryID  *int64
	HabitHistoryTitle       string
	HabitHistoryMeta        string
	SessionDetail           *api.SessionDetail
	SessionDetailOpen       bool
	SessionDetailY          int
	SessionContextOpen      bool
	SessionContextY         int
	Ops                     []api.Op
	Context                 *api.ActiveContext
	Timer                   *api.TimerState
	Health                  *api.Health
	AlertStatus             *api.AlertStatus
	AlertReminders          []api.AlertReminder
	UpdateStatus            *api.UpdateStatus
	UpdateChecking          bool
	Settings                *api.CoreSettings
	KernelInfo              *api.KernelInfo
	Elapsed                 int
	TimerTickSeq            int
	StatusMsg               string
	StatusSeq               int
	StatusErr               bool
	Dialog                  string
	DialogErrorMessage      string
	DialogDeleteID          string
	DialogRepoID            int64
	DialogStreamID          int64
	DialogIssueID           int64
	DialogChoiceItems       []string
	DialogChoiceValues      []string
	DialogChoiceDetails     []string
	DialogChoiceCursor      int
	DialogProcessing        bool
	DialogProcessingLabel   string
	DialogViewTitle         string
	DialogViewName          string
	DialogViewMeta          string
	DialogViewBody          string
	DialogViewPath          string
	DialogSupportBundlePath string
	OpsLimit                int
	OpsLimitPinned          bool
}

type MessageDeps struct {
	DefaultOpsLimit            func(MessageState) int
	CurrentOpsLimit            func(MessageState) int
	ClampFiltered              func(*MessageState, uistate.Pane)
	FilteredCursorForRawIndex  func(*MessageState, uistate.Pane, int) int
	SetStatus                  func(*MessageState, string, bool) tea.Cmd
	OpenViewEntityDialog       func(*MessageState, string, string, string, string)
	OpenSupportBundleDialog    func(*MessageState, string, int64, string)
	OpenOnboardingDialog       func(*MessageState)
	AnchorWellbeingScroll      func(*MessageState, uistate.Pane)
	CurrentDate                func(MessageState) string
	LoadRepos                  func() tea.Cmd
	LoadAllIssues              func() tea.Cmd
	LoadStreams                func(int64) tea.Cmd
	LoadIssues                 func(int64) tea.Cmd
	LoadHabits                 func(int64) tea.Cmd
	LoadAllHabits              func() tea.Cmd
	LoadAllStreams             func() tea.Cmd
	LoadHabitStreakDefinitions func() tea.Cmd
	LoadMomentumRange          func(string, int) tea.Cmd
	LoadMomentumFocus          func(string, int) tea.Cmd
	LoadDueHabits              func(string) tea.Cmd
	LoadTodayHabits            func(string) tea.Cmd
	LoadDailySummary           func(string) tea.Cmd
	LoadSummarySnapshot        func(string) tea.Cmd
	LoadDailyStreaks           func(string) tea.Cmd
	LoadWellbeing              func(string, int) tea.Cmd
	LoadRollupSummaries        func(string, string) tea.Cmd
	LoadDailyPlan              func(string) tea.Cmd
	LoadExportAssets           func() tea.Cmd
	LoadExportReports          func() tea.Cmd
	LoadIssueSessions          func(int64) tea.Cmd
	LoadHabitHistory           func(*api.ActiveContext, *int64) tea.Cmd
	LoadSessionHistoryFor200   func(MessageState) tea.Cmd
	LoadSessionDetail          func(string) tea.Cmd
	LoadOps                    func(int) tea.Cmd
	LoadContext                func() tea.Cmd
	LoadTimer                  func() tea.Cmd
	LoadHealth                 func() tea.Cmd
	LoadAlertStatus            func() tea.Cmd
	LoadAlertReminders         func() tea.Cmd
	LoadUpdateStatus           func() tea.Cmd
	LoadSettings               func() tea.Cmd
	LoadKernelInfo             func() tea.Cmd
	NotifyAlert                func(sharedtypes.AlertRequest) tea.Cmd
	ReportHandledError         func(error, string) tea.Cmd
	HealthTickAfter            func() tea.Cmd
	TickAfter                  func(int) tea.Cmd
	WaitForEvent               func() tea.Cmd
	HandleKernelEvent          func(MessageState, api.KernelEvent) (MessageState, tea.Cmd)
	HardLimitRemaining         func(*api.TimerState, int) int
	CloseEventStop             func()
}

func HandleMessage(
	state MessageState,
	raw tea.Msg,
	deps MessageDeps,
) (MessageState, tea.Cmd, bool) {
	switch msg := raw.(type) {
	case tea.WindowSizeMsg:
		state.Width = msg.Width
		state.Height = msg.Height
		if !state.OpsLimitPinned {
			nextLimit := deps.DefaultOpsLimit(state)
			if nextLimit != state.OpsLimit {
				state.OpsLimit = nextLimit
				return state, deps.LoadOps(deps.CurrentOpsLimit(state)), true
			}
		}
		return state, nil, true
	case commands.ReposLoadedMsg:
		state.Repos = msg.Repos
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneRepos)
		}
		// The scope tree needs every repo's streams, so they follow the repos.
		if deps.LoadAllStreams != nil {
			return state, deps.LoadAllStreams(), true
		}
		return state, nil, true
	case commands.AllStreamsLoadedMsg:
		state.AllStreams = msg.Streams
		return state, nil, true
	case commands.StreamsLoadedMsg:
		state.Streams = msg.Streams
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneStreams)
		}
		return state, nil, true
	case commands.AllHabitsLoadedMsg:
		state.AllHabits = msg.Habits
		if state.View == uistate.ViewHabitList && deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneHabits)
		}
		return state, nil, true
	case commands.HabitStreakDefinitionsLoadedMsg:
		state.HabitStreakDefs = msg.Definitions
		return state, nil, true
	case commands.MomentumRangeLoadedMsg:
		state.MomentumCards = msg.Cards
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneMomentumCards)
		}
		return state, nil, true
	case commands.IssuesLoadedMsg:
		state.Issues = msg.Issues
		if state.View == uistate.ViewDaily {
			if !restoreIssueCursorByID(&state, uistate.PaneIssues, msg.SelectedIssueID) {
				if deps.ClampFiltered != nil {
					deps.ClampFiltered(&state, uistate.PaneIssues)
				}
			}
		}
		return state, nil, true
	case commands.HabitsLoadedMsg:
		state.Habits = msg.Habits
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneHabits)
		}
		return state, nil, true
	case commands.HabitHistoryLoadedMsg:
		state.HabitHistory = msg.Completions
		state.HabitHistoryTitle = "Habit History"
		state.HabitHistoryMeta = habitHistoryScopeLabel(msg.Scope)
		if state.View == uistate.ViewHabitHistory {
			if !restoreHabitHistoryCursorByID(&state, uistate.PaneHabitHistory, msg.SelectedHabitHistoryID) {
				if deps.ClampFiltered != nil {
					deps.ClampFiltered(&state, uistate.PaneHabitHistory)
				}
			}
		}
		return state, nil, true
	case commands.AllIssuesLoadedMsg:
		state.AllIssues = msg.Issues
		if state.View == uistate.ViewDefault || state.View == uistate.ViewDaily {
			if !restoreIssueCursorByID(&state, uistate.PaneIssues, msg.SelectedIssueID) {
				if deps.ClampFiltered != nil {
					deps.ClampFiltered(&state, uistate.PaneIssues)
				}
			}
		}
		return state, nil, true
	case commands.DueHabitsLoadedMsg:
		state.DueHabits = msg.Habits
		if state.View == uistate.ViewDaily {
			if deps.ClampFiltered != nil {
				deps.ClampFiltered(&state, uistate.PaneHabits)
			}
		}
		return state, nil, true
	case commands.DailySummaryLoadedMsg:
		state.DailySummary = msg.Summary
		return state, nil, true
	case commands.DailyPlanLoadedMsg:
		state.DailyPlan = msg.Plan
		return state, nil, true
	case commands.DailyCheckInLoadedMsg:
		state.DailyCheckIn = msg.CheckIn
		if deps.AnchorWellbeingScroll != nil {
			deps.AnchorWellbeingScroll(&state, uistate.PaneWellbeingSummary)
		}
		return state, nil, true
	case commands.MetricsRangeLoadedMsg:
		state.MetricsRange = msg.Days
		if deps.AnchorWellbeingScroll != nil {
			deps.AnchorWellbeingScroll(&state, uistate.PaneWellbeingTrends)
		}
		return state, nil, true
	case commands.RollupMetricsRangeLoadedMsg:
		state.RollupMetricsRange = msg.Days
		return state, nil, true
	case commands.MomentumMetricsRangeLoadedMsg:
		state.MomentumMetricsRange = msg.Days
		if state.MomentumTab == "focus" {
			if state.MomentumHistoryY < 0 {
				state.MomentumHistoryY = 0
			}
			if len(msg.Days) == 0 {
				state.MomentumHistoryY = 0
			} else if state.MomentumHistoryY >= len(msg.Days) {
				state.MomentumHistoryY = len(msg.Days) - 1
			}
		}
		return state, nil, true
	case commands.MetricsRollupLoadedMsg:
		state.MetricsRollup = msg.Rollup
		if deps.AnchorWellbeingScroll != nil {
			deps.AnchorWellbeingScroll(&state, uistate.PaneWellbeingTrends)
		}
		return state, nil, true
	case commands.RollupMetricsRollupLoadedMsg:
		state.RollupMetricsRollup = msg.Rollup
		return state, nil, true
	case commands.MomentumMetricsRollupLoadedMsg:
		state.MomentumMetricsRollup = msg.Rollup
		return state, nil, true
	case commands.StreaksLoadedMsg:
		state.Streaks = msg.Streaks
		return state, nil, true
	case commands.DailyStreaksLoadedMsg:
		state.DailyStreaks = msg.Streaks
		return state, nil, true
	case commands.RollupRefreshRequestedMsg:
		return state, deps.LoadRollupSummaries(rollupRange(deps.CurrentDate(state), state.RollupWindowDays)), true
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
	case commands.ExportAssetsLoadedMsg:
		state.ExportAssets = msg.Assets
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneConfig)
		}
		return state, deps.LoadExportReports(), true
	case commands.ExportReportsLoadedMsg:
		state.ExportReports = msg.Reports
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneExportReports)
		}
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
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneSessions)
		}
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
	case commands.OpsLoadedMsg:
		state.Ops = msg.Ops
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneOps)
		}
		return state, nil, true
	case commands.ContextLoadedMsg:
		state.Context = msg.Ctx
		if state.View == uistate.ViewDefault || state.View == uistate.ViewDaily {
			if deps.ClampFiltered != nil {
				deps.ClampFiltered(&state, uistate.PaneIssues)
				deps.ClampFiltered(&state, uistate.PaneHabits)
			}
		}
		cmds := []tea.Cmd{deps.LoadRollupSummaries(rollupRange(deps.CurrentDate(state), state.RollupWindowDays))}
		if state.View == uistate.ViewHabitHistory {
			cmds = append(cmds, deps.LoadHabitHistory(state.Context, state.SelectedHabitHistoryID))
		}
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
			if state.View != uistate.ViewSessionHistory {
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
				if state.Timer.State == "ready" {
					cmds := []tea.Cmd{deps.LoadIssueSessions(*state.Timer.IssueID), historyCmd}
					if state.Timer.HardLimitActive && !state.Timer.HardLimitExpired {
						cmds = append(cmds, deps.TickAfter(state.TimerTickSeq))
					}
					return state, tea.Batch(cmds...), true
				}
				return state, tea.Batch(deps.LoadIssueSessions(*state.Timer.IssueID), historyCmd, deps.TickAfter(state.TimerTickSeq)), true
			}
		}
		if state.Timer != nil && state.Timer.State != "idle" {
			if state.Timer.State == "ready" {
				if state.Timer.HardLimitActive && !state.Timer.HardLimitExpired {
					return state, tea.Batch(historyCmd, deps.TickAfter(state.TimerTickSeq)), true
				}
				return state, historyCmd, true
			}
			return state, tea.Batch(historyCmd, deps.TickAfter(state.TimerTickSeq)), true
		}
		return state, historyCmd, true
	case commands.HealthLoadedMsg:
		previousDate := ""
		if state.Health != nil {
			previousDate = strings.TrimSpace(state.Health.CurrentDate)
		}
		state.Health = msg.Health
		if msg.Health == nil || strings.TrimSpace(msg.Health.CurrentDate) == "" {
			return state, nil, true
		}
		date := strings.TrimSpace(msg.Health.CurrentDate)
		if previousDate == date && state.Date != "" {
			return state, nil, true
		}
		cmd := applyLogicalDate(&state, date, previousDate, deps)
		return state, cmd, true
	case commands.AlertStatusLoadedMsg:
		state.AlertStatus = msg.Status
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneAlerts)
		}
		return state, nil, true
	case commands.AlertRemindersLoadedMsg:
		state.AlertReminders = msg.Reminders
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneAlerts)
		}
		return state, nil, true
	case commands.UpdateStatusLoadedMsg:
		state.UpdateChecking = false
		state.UpdateStatus = msg.Status
		return state, nil, true
	case commands.AlertTestedMsg:
		return state, deps.SetStatus(&state, msg.Label, false), true
	case commands.AlertReminderChangedMsg:
		return state, tea.Batch(deps.SetStatus(&state, msg.Label, false), deps.LoadAlertReminders()), true
	case commands.SettingsLoadedMsg:
		state.Settings = msg.Settings
		if deps.ClampFiltered != nil {
			deps.ClampFiltered(&state, uistate.PaneSettings)
			deps.ClampFiltered(&state, uistate.PaneAlerts)
		}
		if msg.Settings != nil && !msg.Settings.OnboardingCompleted && state.Dialog == "" && deps.OpenOnboardingDialog != nil {
			logger.Infof("tui settings loaded: opening onboarding dialog onboarding_completed=%t usage=%t errors=%t dialog=%q", msg.Settings.OnboardingCompleted, msg.Settings.UsageTelemetryEnabled, msg.Settings.ErrorReportingEnabled, state.Dialog)
			deps.OpenOnboardingDialog(&state)
		}
		return state, nil, true
	case commands.KernelInfoLoadedMsg:
		state.KernelInfo = msg.Info
		return state, nil, true
	case commands.HealthTickMsg:
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
	case commands.KernelRestartedMsg:
		deps.CloseEventStop()
		return state, tea.Quit, true
	case commands.TelemetrySettingsSavedMsg:
		return state, deps.SetStatus(&state, "Telemetry settings saved. Restart required.", false), true
	case commands.DevSeededMsg:
		cmd := deps.SetStatus(&state, "Dev seed loaded", false)
		state.View = uistate.ViewDaily
		state.Pane = uistate.DefaultPane(state.View)
		return state, fullReloadCmd(state, deps, cmd), true
	case commands.DevClearedMsg:
		cmd := deps.SetStatus(&state, "Dev data cleared", false)
		state.View = uistate.ViewDaily
		state.Pane = uistate.DefaultPane(state.View)
		return state, fullReloadCmd(state, deps, cmd), true
	case commands.RuntimeDataWipedMsg:
		cmd := deps.SetStatus(&state, "All runtime data wiped", false)
		state.View = uistate.ViewDaily
		state.Pane = uistate.DefaultPane(state.View)
		return state, fullReloadCmd(state, deps, cmd), true
	case commands.UninstallStartedMsg:
		deps.CloseEventStop()
		return state, tea.Quit, true
	case commands.SessionAmendedMsg:
		cmd := deps.SetStatus(&state, "Session amended", false)
		return state, historyRefreshCmd(state, deps, cmd, deps.LoadSessionDetail(msg.ID)), true
	case commands.SessionHistoryDeletedMsg:
		state.SessionDetailOpen = false
		state.SessionDetail = nil
		cmd := deps.SetStatus(&state, "Session deleted", false)
		return state, historyRefreshCmd(state, deps, cmd), true
	case commands.ManualSessionLoggedMsg:
		cmds := []tea.Cmd{deps.SetStatus(&state, "Manual session logged", false), deps.LoadSessionHistoryFor200(state), deps.LoadAllIssues()}
		if state.Context != nil && state.Context.IssueID != nil && *state.Context.IssueID == msg.IssueID {
			cmds = append(cmds, deps.LoadIssueSessions(msg.IssueID))
		}
		if deps.CurrentDate(state) == strings.TrimSpace(msg.Date) {
			cmds = append(cmds, deps.LoadDailySummary(msg.Date))
			cmds = append(cmds, deps.LoadDailyStreaks(msg.Date))
			cmds = append(cmds, deps.LoadWellbeing(msg.Date, state.WellbeingWindowDays))
		}
		return state, tea.Batch(cmds...), true
	case commands.IssuePinnedDailyChangedMsg:
		label := "Issue pinned"
		if !msg.Pinned {
			label = "Issue unpinned"
		}
		return state, deps.SetStatus(&state, label, false), true
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
		if state.Timer != nil && state.Timer.State != "idle" && state.Timer.State != "expired" {
			state.Elapsed++
			if state.Timer.HardLimitActive &&
				deps.HardLimitRemaining != nil &&
				deps.HardLimitRemaining(state.Timer, state.Elapsed) <= 0 {
				return state, deps.LoadTimer(), true
			}
			return state, deps.TickAfter(state.TimerTickSeq), true
		}
		return state, nil, true
	case commands.KernelEventMsg:
		updated, cmd := deps.HandleKernelEvent(state, msg.Event)
		return updated, tea.Batch(cmd, deps.WaitForEvent()), true
	case commands.ErrMsg:
		if state.Dialog != "" {
			if isExportDialog(state.Dialog) && state.DialogProcessing {
				state.DialogProcessing = false
				state.DialogProcessingLabel = ""
			}
			state.DialogErrorMessage = "Error: " + msg.Err.Error()
			logger.Errorf("update error: %v", msg.Err)
			return state, deps.ReportHandledError(msg.Err, msg.Operation), true
		}
		logger.Errorf("update error: %v", msg.Err)
		return state, tea.Batch(
			deps.ReportHandledError(msg.Err, msg.Operation),
			deps.SetStatus(&state, "Error: "+msg.Err.Error(), true),
		), true
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
			return state, tea.Batch(
				deps.SetStatus(&state, label+" written to "+*msg.Result.FilePath, false),
				deps.LoadExportReports(),
				deps.NotifyAlert(sharedtypes.AlertRequest{
					Kind:      sharedtypes.AlertEventExportCompleted,
					Title:     label + " ready",
					Subtitle:  "Export completed",
					Body:      *msg.Result.FilePath,
					Urgency:   sharedtypes.AlertUrgencyLow,
					PlaySound: false,
				}),
			), true
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
		return state, tea.Batch(
			deps.SetStatus(&state, "Calendar export written", false),
			deps.NotifyAlert(sharedtypes.AlertRequest{
				Kind:      sharedtypes.AlertEventExportCompleted,
				Title:     "Calendar export ready",
				Subtitle:  name,
				Body:      msg.Result.IssuesFilePath,
				Urgency:   sharedtypes.AlertUrgencyLow,
				PlaySound: false,
			}),
		), true
	case commands.ClipboardCopiedMsg:
		if isExportDialog(state.Dialog) && state.DialogProcessing {
			state.Dialog = ""
			state.DialogChoiceItems = nil
			state.DialogChoiceCursor = 0
			state.DialogProcessing = false
			state.DialogProcessingLabel = ""
		}
		return state, deps.SetStatus(&state, msg.Message, false), true
	case commands.SupportBundleGeneratedMsg:
		deps.OpenSupportBundleDialog(&state, msg.Path, msg.SizeBytes, msg.WindowLabel)
		return state, deps.NotifyAlert(sharedtypes.AlertRequest{
			Kind:      sharedtypes.AlertEventSupportBundleReady,
			Title:     "Support bundle ready",
			Subtitle:  msg.WindowLabel,
			Body:      msg.Path,
			Urgency:   sharedtypes.AlertUrgencyLow,
			PlaySound: false,
		}), true
	case commands.SupportDiagnosticsWrittenMsg:
		return state, deps.SetStatus(&state, "Recent diagnostics written to "+msg.Path, false), true
	default:
		return state, nil, false
	}
}

func restoreIssueCursorByID(state *MessageState, pane uistate.Pane, selectedIssueID *int64) bool {
	if selectedIssueID == nil || *selectedIssueID == 0 {
		return false
	}
	snapshot := selectionSnapshotFromMessageState(*state)
	if filteredCur := selectionpkg.FilteredCursorForIssueID(snapshot, pane, *selectedIssueID); filteredCur >= 0 {
		if state.Cursor == nil {
			state.Cursor = map[uistate.Pane]int{}
		}
		state.Cursor[pane] = filteredCur
		return true
	}
	return false
}

func restoreHabitHistoryCursorByID(
	state *MessageState,
	pane uistate.Pane,
	selectedHabitHistoryID *int64,
) bool {
	if selectedHabitHistoryID == nil || *selectedHabitHistoryID == 0 {
		return false
	}
	snapshot := selectionSnapshotFromMessageState(*state)
	if filteredCur := selectionpkg.FilteredCursorForHabitHistoryID(snapshot, pane, *selectedHabitHistoryID); filteredCur >= 0 {
		if state.Cursor == nil {
			state.Cursor = map[uistate.Pane]int{}
		}
		state.Cursor[pane] = filteredCur
		return true
	}
	return false
}

func habitHistoryScopeLabel(ctx *api.ActiveContext) string {
	if ctx == nil {
		return "Recent habit activity across the workspace"
	}
	repoName := ""
	if ctx.RepoName != nil {
		repoName = strings.TrimSpace(*ctx.RepoName)
	}
	streamName := ""
	if ctx.StreamName != nil {
		streamName = strings.TrimSpace(*ctx.StreamName)
	}
	switch {
	case repoName != "" && streamName != "":
		return "Recent habit activity in " + repoName + " > " + streamName
	case repoName != "":
		return "Recent habit activity in " + repoName
	case streamName != "":
		return "Recent habit activity in " + streamName
	default:
		return "Recent habit activity across the workspace"
	}
}

// messageStateDate resolves the shared date without deps: an unset date follows the
// daemon's logical today, and selection falls back to the clock before health loads.
func messageStateDate(state MessageState) string {
	if state.Date != "" {
		return state.Date
	}
	if state.Health != nil {
		return strings.TrimSpace(state.Health.CurrentDate)
	}
	return ""
}

func selectionSnapshotFromMessageState(state MessageState) selectionpkg.Snapshot {
	dailySection := state.DailyTaskSection
	if state.View == uistate.ViewDaily && dailySection == "" {
		dailySection = uistate.DailyTaskSectionPlanned
	}
	return selectionpkg.PrepareSnapshot(selectionpkg.Snapshot{
		View:                state.View,
		Pane:                state.Pane,
		DailyTaskSection:    dailySection,
		DefaultIssueSection: state.DefaultIssueSection,
		BrowseScope:         state.BrowseScope,
		DashboardDate:       messageStateDate(state),
		Cursors:             state.Cursor,
		Filters:             state.Filters,
		Context:             state.Context,
		Timer:               state.Timer,
		Settings:            state.Settings,
		AllIssues:           state.AllIssues,
		Issues:              state.Issues,
		HabitHistory:        state.HabitHistory,
	})
}

func fullReloadCmd(state MessageState, deps MessageDeps, extra ...tea.Cmd) tea.Cmd {
	date := deps.CurrentDate(state)
	cmds := make([]tea.Cmd, 0, len(extra)+20)
	cmds = append(cmds, extra...)
	cmds = append(cmds,
		deps.LoadKernelInfo(),
		deps.LoadRepos(),
		deps.LoadHabitStreakDefinitions(),
		deps.LoadMomentumRange(date, state.MomentumWindowDays),
		deps.LoadAllIssues(),
		deps.LoadDueHabits(date),
		deps.LoadDailySummary(date),
		deps.LoadDailyStreaks(date),
		deps.LoadWellbeing(date, state.WellbeingWindowDays),
		deps.LoadRollupSummaries(rollupRange(date, state.RollupWindowDays)),
		deps.LoadSessionHistoryFor200(state),
		deps.LoadOps(deps.CurrentOpsLimit(state)),
		deps.LoadContext(),
		deps.LoadTimer(),
		deps.LoadAlertStatus(),
		deps.LoadAlertReminders(),
		deps.LoadUpdateStatus(),
		deps.LoadSettings(),
		deps.LoadExportAssets(),
		deps.LoadExportReports(),
	)
	if deps.LoadTodayHabits != nil && state.Health != nil {
		cmds = append(cmds, deps.LoadTodayHabits(state.Health.CurrentDate))
	}
	if deps.LoadSummarySnapshot != nil {
		cmds = append(cmds, deps.LoadSummarySnapshot(date))
	}
	return tea.Batch(cmds...)
}

// defaultRollupWindowDays is Rollup's window when none has been chosen.
const defaultRollupWindowDays = 7

// rollupRange is the Rollup window that ends on date.
func rollupRange(date string, windowDays int) (string, string) {
	if windowDays < 1 {
		windowDays = defaultRollupWindowDays
	}
	return shiftISODate(date, -(windowDays - 1)), date
}

// applyLogicalDate moves the shared date to the new logical today when it was on the
// previous one (or unset), so a user looking at another day stays there.
func applyLogicalDate(state *MessageState, date, previousDate string, deps MessageDeps) tea.Cmd {
	if previousDate == "" || state.Date == "" || state.Date == previousDate {
		state.Date = date
	}
	shown := state.Date
	cmds := []tea.Cmd{
		deps.LoadAllIssues(),
		deps.LoadAllHabits(),
		deps.LoadDailySummary(shown),
		deps.LoadDailyPlan(shown),
		deps.LoadDueHabits(shown),
		deps.LoadDailyStreaks(shown),
		deps.LoadWellbeing(shown, state.WellbeingWindowDays),
		deps.LoadMomentumRange(shown, state.MomentumWindowDays),
		deps.LoadRollupSummaries(rollupRange(shown, state.RollupWindowDays)),
	}
	if deps.LoadTodayHabits != nil {
		cmds = append(cmds, deps.LoadTodayHabits(date))
	}
	if deps.LoadSummarySnapshot != nil {
		cmds = append(cmds, deps.LoadSummarySnapshot(shown))
	}
	return tea.Batch(cmds...)
}

func isExportDialog(kind string) bool {
	switch kind {
	case "export_report", "export_preset", "export_calendar_repo":
		return true
	default:
		return false
	}
}

func historyRefreshCmd(state MessageState, deps MessageDeps, extra ...tea.Cmd) tea.Cmd {
	date := state.Date
	if deps.CurrentDate != nil {
		date = deps.CurrentDate(state)
	}
	cmds := append([]tea.Cmd{}, extra...)
	if deps.LoadSessionHistoryFor200 != nil {
		cmds = append(cmds, deps.LoadSessionHistoryFor200(state))
	}
	if deps.LoadAllIssues != nil {
		cmds = append(cmds, deps.LoadAllIssues())
	}
	if state.Context != nil {
		if state.Context.StreamID != nil && deps.LoadIssues != nil {
			cmds = append(cmds, deps.LoadIssues(*state.Context.StreamID))
		}
		if state.Context.IssueID != nil && deps.LoadIssueSessions != nil {
			cmds = append(cmds, deps.LoadIssueSessions(*state.Context.IssueID))
		}
	}
	if deps.LoadDailySummary != nil {
		cmds = append(cmds, deps.LoadDailySummary(date))
	}
	if deps.LoadSummarySnapshot != nil {
		cmds = append(cmds, deps.LoadSummarySnapshot(date))
	}
	if deps.LoadDailyStreaks != nil {
		cmds = append(cmds, deps.LoadDailyStreaks(date))
	}
	if deps.LoadRollupSummaries != nil {
		cmds = append(cmds, deps.LoadRollupSummaries(rollupRange(date, state.RollupWindowDays)))
	}
	if deps.LoadWellbeing != nil {
		cmds = append(cmds, deps.LoadWellbeing(date, state.WellbeingWindowDays))
	}
	if deps.LoadMomentumRange != nil {
		cmds = append(cmds, deps.LoadMomentumRange(date, state.MomentumWindowDays))
	}
	return tea.Batch(cmds...)
}
