package app

import (
	"strings"
	"sync"
	"time"

	"crona/shared/config"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	commands "crona/tui/internal/tui/commands"
	configitems "crona/tui/internal/tui/configitems"
	dialogruntime "crona/tui/internal/tui/dialog_runtime"
	dialogpkg "crona/tui/internal/tui/dialogs"
	dialogstate "crona/tui/internal/tui/dialogstate"
	filteringpkg "crona/tui/internal/tui/filtering"
	helperpkg "crona/tui/internal/tui/helpers"
	inputpkg "crona/tui/internal/tui/input"
	appruntime "crona/tui/internal/tui/runtime"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/terminaltitle"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	viewruntime "crona/tui/internal/tui/views/runtime"
	wellbeingview "crona/tui/internal/tui/views/wellbeing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------- View / Pane types ----------

type View = uistate.View

const (
	ViewAway           = uistate.ViewAway
	ViewDefault        = uistate.ViewDefault
	ViewDaily          = uistate.ViewDaily
	ViewRollup         = uistate.ViewRollup
	ViewMeta           = uistate.ViewMeta
	ViewSessionHistory = uistate.ViewSessionHistory
	ViewSessionActive  = uistate.ViewSessionActive
	ViewScratch        = uistate.ViewScratch
	ViewOps            = uistate.ViewOps
	ViewWellbeing      = uistate.ViewWellbeing
	ViewReports        = uistate.ViewReports
	ViewConfig         = uistate.ViewConfig
	ViewSettings       = uistate.ViewSettings
	ViewAlerts         = uistate.ViewAlerts
	ViewUpdates        = uistate.ViewUpdates
	ViewSupport        = uistate.ViewSupport
)

type Pane = uistate.Pane

const (
	PaneRepos            = uistate.PaneRepos
	PaneStreams          = uistate.PaneStreams
	PaneIssues           = uistate.PaneIssues
	PaneHabits           = uistate.PaneHabits
	PaneRollupDays       = uistate.PaneRollupDays
	PaneSessions         = uistate.PaneSessions
	PaneScratchpads      = uistate.PaneScratchpads
	PaneOps              = uistate.PaneOps
	PaneExportReports    = uistate.PaneExportReports
	PaneConfig           = uistate.PaneConfig
	PaneSettings         = uistate.PaneSettings
	PaneAlerts           = uistate.PaneAlerts
	PaneWellbeingSummary = uistate.PaneWellbeingSummary
	PaneWellbeingTrends  = uistate.PaneWellbeingTrends
)

type DefaultIssueSection = uistate.DefaultIssueSection

const (
	DefaultIssueSectionOpen      = uistate.DefaultIssueSectionOpen
	DefaultIssueSectionCompleted = uistate.DefaultIssueSectionCompleted
)

// ---------- Model ----------

type Model struct {
	// kernel client
	client *api.Client

	// kernel event stream
	eventStop *eventStreamStop

	// view / navigation
	view                View
	pane                Pane
	cursor              map[Pane]int
	filters             map[Pane]string
	defaultIssueSection DefaultIssueSection

	// pane-local search/filter input
	filterEditing  bool
	filterPane     Pane
	filterInput    textinput.Model
	opsLimit       int
	opsLimitPinned bool

	// data
	repos                  []api.Repo
	streams                []api.Stream
	issues                 []api.Issue // context-filtered (by active streamId)
	habits                 []api.Habit
	allIssues              []api.IssueWithMeta
	dueHabits              []api.HabitDailyItem
	dailySummary           *api.DailyIssueSummary
	dailyPlan              *api.DailyPlan
	dashboardDate          string
	rollupStartDate        string
	rollupEndDate          string
	wellbeingDate          string
	dailyCheckIn           *api.DailyCheckIn
	metricsRange           []api.DailyMetricsDay
	metricsRollup          *api.MetricsRollup
	streaks                *api.StreakSummary
	dashboardWindow        *api.DashboardWindowSummary
	dailyFocusScore        *api.FocusScoreSummary
	weeklyFocusScore       *api.FocusScoreSummary
	repoDistribution       *api.TimeDistributionSummary
	streamDistribution     *api.TimeDistributionSummary
	issueDistribution      *api.TimeDistributionSummary
	segmentDistribution    *api.TimeDistributionSummary
	goalProgress           *api.GoalProgressSummary
	exportAssets           *api.ExportAssetStatus
	exportReports          []api.ExportReportFile
	issueSessions          []api.Session
	sessionHistory         []api.SessionHistoryEntry
	sessionDetail          *api.SessionDetail
	scratchpads            []api.ScratchPad
	stashes                []api.Stash
	ops                    []api.Op
	context                *api.ActiveContext
	timer                  *api.TimerState
	health                 *api.Health
	alertStatus            *api.AlertStatus
	alertReminders         []api.AlertReminder
	updateStatus           *api.UpdateStatus
	updateChecking         bool
	updateInstalling       bool
	updateInstallError     string
	currentExecutablePath  string
	settings               *api.CoreSettings
	kernelInfo             *api.KernelInfo
	elapsed                int // local seconds since last timer.state event
	timerTickSeq           int
	lastTimerActivityTouch time.Time
	terminalTitleEnabled   bool
	lastTerminalTitle      string

	// terminal dimensions
	width  int
	height int

	// scratchpad reader state within the scratchpads pane
	scratchpadOpen     bool
	scratchpadMeta     *api.ScratchPad
	scratchpadFilePath string // resolved absolute path for $EDITOR
	scratchpadRendered string // plain text scratchpad content
	scratchpadViewport viewport.Model

	// dialog state
	dialog                   string // "" | "create_scratchpad" | "confirm_delete" | "stash_list"
	dialogInputs             []textinput.Model
	dialogDescription        textarea.Model
	dialogDescriptionOn      bool
	dialogDescriptionIdx     int
	dialogFocusIdx           int
	dialogErrorMessage       string
	dialogDeleteID           string // scratchpad id pending deletion
	dialogDeleteKind         string
	dialogDeleteLabel        string
	dialogSessionID          string
	dialogIssueID            int64
	dialogHabitID            int64
	dialogIssueStatus        string
	dialogCheckInDate        string
	dialogRepoID             int64
	dialogRepoName           string
	dialogRepoItems          []string
	dialogRepoItemIDs        []int64
	dialogStreamID           int64
	dialogStreamName         string
	dialogRepoIndex          int
	dialogStreamIndex        int
	dialogParent             string
	dialogDateMonth          string
	dialogDateCursor         string
	dialogStashCursor        int
	dialogStatusItems        []sharedtypes.IssueStatus
	dialogStatusCursor       int
	dialogChoiceItems        []string
	dialogChoiceValues       []string
	dialogChoiceDetails      []string
	dialogTemplateAssets     []sharedtypes.ExportTemplateAsset
	dialogChoiceCursor       int
	dialogProcessing         bool
	dialogProcessingLabel    string
	dialogStatusLabel        string
	dialogStatusRequired     bool
	dialogViewTitle          string
	dialogViewName           string
	dialogIssueEstimateMins  *int
	dialogReminderID         string
	dialogReminderKind       sharedtypes.AlertReminderKind
	dialogViewMeta           string
	dialogViewBody           string
	dialogSupportBundlePath  string
	dialogProtectionStep     int
	dialogProtectionCursor   int
	dialogProtectionStreaks  []sharedtypes.StreakKind
	dialogProtectionWeekdays []int
	dialogProtectionDates    []string
	dialogExportPresetKind   sharedtypes.ExportReportKind
	dialogExportPresetFormat sharedtypes.ExportFormat
	dialogExportPresetOutput sharedtypes.ExportOutputMode
	dialogExportIncludePDF   bool

	// status / error flash
	statusMsg string
	statusSeq int
	statusErr bool

	// overlay help
	helpOpen           bool
	sessionDetailOpen  bool
	sessionDetailY     int
	sessionContextOpen bool
	sessionContextY    int
}

func (m Model) selfUpdateInstallAvailable() bool {
	return m.updateStatus != nil && m.updateStatus.InstallAvailable && m.selfUpdateUnsupportedReason() == ""
}

func (m Model) selfUpdateUnsupportedReason() string {
	if m.updateStatus != nil && commands.IsLocalLoopbackUpdateURL(m.updateStatus.InstallScriptURL) && m.isDevMode() {
		return ""
	}
	if reason := appruntime.NonStandardRuntimeReason(m.currentExecutablePath, config.TUIBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	if m.kernelInfo == nil {
		return "Kernel info is unavailable. Please update manually."
	}
	if reason := appruntime.NonStandardRuntimeReason(m.kernelInfo.ExecutablePath, config.KernelBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	return ""
}

func (m *Model) stopEventStream() {
	if m.eventStop == nil {
		return
	}
	m.eventStop.Stop()
}

type eventStreamStop struct {
	ch   chan struct{}
	once sync.Once
}

func newEventStreamStop(ch chan struct{}) *eventStreamStop {
	if ch == nil {
		return nil
	}
	return &eventStreamStop{ch: ch}
}

func (s *eventStreamStop) Stop() {
	if s == nil || s.ch == nil {
		return
	}
	s.once.Do(func() {
		close(s.ch)
	})
}

func kernelEnvMode(info *api.KernelInfo) string {
	if info == nil {
		return ""
	}
	return info.Env
}

func kernelExecutablePath(info *api.KernelInfo) string {
	if info == nil {
		return ""
	}
	return info.ExecutablePath
}

// SetEventChannel provides the kernel event channel from main before the program starts.
func SetEventChannel(ch <-chan api.KernelEvent) {
	eventChannel = ch
}

func New(transport, endpoint, scratchDir string, env string, executablePath string, done chan struct{}) Model {
	model := Model{
		client:              api.NewClient(transport, endpoint, scratchDir),
		eventStop:           newEventStreamStop(done),
		view:                ViewDaily,
		pane:                PaneIssues,
		defaultIssueSection: DefaultIssueSectionOpen,
		cursor: map[Pane]int{
			PaneRepos:            0,
			PaneStreams:          0,
			PaneIssues:           0,
			PaneHabits:           0,
			PaneRollupDays:       0,
			PaneSessions:         0,
			PaneScratchpads:      0,
			PaneOps:              0,
			PaneExportReports:    0,
			PaneConfig:           0,
			PaneSettings:         0,
			PaneAlerts:           0,
			PaneWellbeingSummary: 0,
			PaneWellbeingTrends:  0,
		},
		filters: map[Pane]string{
			PaneRepos:            "",
			PaneStreams:          "",
			PaneIssues:           "",
			PaneHabits:           "",
			PaneRollupDays:       "",
			PaneSessions:         "",
			PaneScratchpads:      "",
			PaneOps:              "",
			PaneExportReports:    "",
			PaneConfig:           "",
			PaneSettings:         "",
			PaneAlerts:           "",
			PaneWellbeingSummary: "",
			PaneWellbeingTrends:  "",
		},
		currentExecutablePath: executablePath,
		kernelInfo:            &api.KernelInfo{Env: env},
		terminalTitleEnabled:  true,
	}
	model.lastTerminalTitle = terminaltitle.Sanitize(model.terminalTitle())
	return model
}

// eventChannel receives kernel events forwarded from main.go.
var eventChannel <-chan api.KernelEvent

// ---------- Init ----------

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		commands.LoadRepos(m.client),
		commands.LoadAllIssues(m.client),
		commands.LoadDueHabits(m.client, time.Now().Format("2006-01-02")),
		commands.LoadDailySummary(m.client, ""),
		commands.LoadWellbeing(m.client, time.Now().Format("2006-01-02")),
		commands.LoadRollupSummaries(m.client, shiftISODate(time.Now().Format("2006-01-02"), -6), time.Now().Format("2006-01-02")),
		loadSessionHistoryForModel(m, 200),
		commands.LoadScratchpads(m.client),
		commands.LoadOps(m.client, m.currentOpsLimit()),
		commands.LoadContext(m.client),
		commands.LoadTimer(m.client),
		commands.LoadHealth(m.client),
		commands.LoadAlertStatus(m.client),
		commands.LoadAlertReminders(m.client),
		commands.LoadUpdateStatus(m.client),
		commands.LoadSettings(m.client),
		commands.LoadKernelInfo(m.client),
		commands.LoadExportAssets(m.client),
		commands.LoadExportReports(m.client),
		commands.HealthTickAfter(),
		commands.WaitForEvent(eventChannel),
	}
	if m.terminalTitleEnabled {
		cmds = append(cmds, terminaltitle.Command(m.terminalTitle()))
	}
	return tea.Batch(cmds...)
}

// ---------- Helpers: clamp cursor ----------

func (m *Model) clamp(p Pane, max int) {
	if max == 0 {
		m.cursor[p] = 0
		return
	}
	if m.cursor[p] >= max {
		m.cursor[p] = max - 1
	}
}

func (m *Model) listLen(p Pane) int {
	if p == PaneRollupDays {
		if m.dashboardWindow == nil {
			return 0
		}
		return len(m.dashboardWindow.Days)
	}
	if p == PaneWellbeingSummary || p == PaneWellbeingTrends {
		snapshot := m.selectionSnapshot()
		activeIssue := selectionpkg.ActiveIssue(snapshot)
		state := m.viewContentState(m.mainContentWidth(), m.contentHeight(), snapshot, activeIssue)
		return wellbeingview.PaneLineCount(state, string(p))
	}
	if p == PaneAlerts {
		return alertsmeta.FilteredSelectableCount(m.filters[PaneAlerts], m.settings, m.alertStatus, m.alertReminders)
	}
	snapshot := m.selectionSnapshot()
	return len(selectionpkg.FilteredIndices(snapshot, p))
}

func (m *Model) defaultOpsLimit() int {
	availableHeight := m.contentHeight()
	if availableHeight < 4 {
		availableHeight = 4
	}
	visibleRows := availableHeight - 6
	if visibleRows < 10 {
		visibleRows = 10
	}
	return visibleRows
}

func (m *Model) currentOpsLimit() int {
	if m.opsLimit > 0 {
		return m.opsLimit
	}
	return m.defaultOpsLimit()
}

func (m Model) selectionSnapshot() selectionpkg.Snapshot {
	return selectionpkg.PrepareSnapshot(selectionpkg.Snapshot{
		View:                m.view,
		Pane:                m.pane,
		DefaultIssueSection: m.defaultIssueSection,
		PreferActiveIssue:   m.timer != nil && m.timer.State != "idle" && (m.view == ViewSessionActive || m.view == ViewScratch),
		Cursors:             m.cursor,
		Filters:             m.filters,
		Context:             m.context,
		Timer:               m.timer,
		Repos:               m.repos,
		Streams:             m.streams,
		Issues:              m.dailyIssuesForSelection(),
		Habits:              m.habits,
		AllIssues:           m.allIssues,
		DueHabits:           m.dueHabits,
		ExportReports:       m.exportReports,
		Scratchpads:         m.scratchpads,
		SessionHistory:      m.sessionHistory,
		Ops:                 m.ops,
		Settings:            m.settings,
		AlertStatus:         m.alertStatus,
		AlertReminders:      m.alertReminders,
		ConfigItems:         m.configItemsForSnapshot(),
	})
}

func (m Model) configItemsForSnapshot() []configitems.Item {
	if m.view != ViewConfig && m.pane != PaneConfig && (!m.filterEditing || m.filterPane != PaneConfig) {
		return nil
	}
	return configitems.Build(m.exportAssets)
}

func (m Model) dailyIssuesForSelection() []api.Issue {
	if m.dailySummary == nil {
		return m.issues
	}
	if m.view == ViewDaily {
		return m.dailySummary.Issues
	}
	return m.issues
}

func (m Model) inputState() inputpkg.State {
	protected := false
	activeView := m.view
	activePane := m.pane
	if nextProtected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); nextProtected {
		protected = true
		if activeView != ViewReports && activeView != ViewSessionHistory {
			activeView = ViewAway
			activePane = uistate.DefaultPane(activeView)
		}
	}
	return inputpkg.State{
		ActiveView:          activeView,
		ActivePane:          activePane,
		ProtectedModeActive: protected,
		Cursor:              m.cursor,
		Filters:             m.filters,
		DefaultIssueSection: m.defaultIssueSection,
		DashboardDate:       m.dashboardDate,
		RollupStartDate:     m.rollupStartDate,
		RollupEndDate:       m.rollupEndDate,
		WellbeingDate:       m.wellbeingDate,
		Dialog:              m.dialog,
		DialogState:         m.dialogState(),
		HelpOpen:            m.helpOpen,
		SessionDetailOpen:   m.sessionDetailOpen,
		SessionDetailY:      m.sessionDetailY,
		SessionContextOpen:  m.sessionContextOpen,
		SessionContextY:     m.sessionContextY,
		ScratchpadOpen:      m.scratchpadOpen,
		OpsLimit:            m.opsLimit,
		OpsLimitPinned:      m.opsLimitPinned,
		Context:             m.context,
		Timer:               m.timer,
		UpdateStatus:        m.updateStatus,
		UpdateChecking:      m.updateChecking,
		UpdateInstalling:    m.updateInstalling,
		UpdateInstallError:  m.updateInstallError,
		CurrentExecutable:   m.currentExecutablePath,
		RunningIsBeta:       m.isBetaBuild(),
		Settings:            m.settings,
		AlertStatus:         m.alertStatus,
		AlertReminders:      m.alertReminders,
		ExportAssets:        m.exportAssets,
		DailyCheckIn:        m.dailyCheckIn,
	}
}

func (m Model) applyInputState(state inputpkg.State) Model {
	m.view = state.ActiveView
	m.pane = state.ActivePane
	m.cursor = state.Cursor
	m.filters = state.Filters
	m.defaultIssueSection = state.DefaultIssueSection
	m.dashboardDate = state.DashboardDate
	m.rollupStartDate = state.RollupStartDate
	m.rollupEndDate = state.RollupEndDate
	m.wellbeingDate = state.WellbeingDate
	m = m.withDialogState(state.DialogState)
	if state.DialogState.Kind == "" {
		m.dialog = state.Dialog
	}
	m.helpOpen = state.HelpOpen
	m.sessionDetailOpen = state.SessionDetailOpen
	m.sessionDetailY = state.SessionDetailY
	m.sessionContextOpen = state.SessionContextOpen
	m.sessionContextY = state.SessionContextY
	m.scratchpadOpen = state.ScratchpadOpen
	m.opsLimit = state.OpsLimit
	m.opsLimitPinned = state.OpsLimitPinned
	m.context = state.Context
	m.timer = state.Timer
	m.updateStatus = state.UpdateStatus
	m.updateChecking = state.UpdateChecking
	m.updateInstalling = state.UpdateInstalling
	m.updateInstallError = state.UpdateInstallError
	m.currentExecutablePath = state.CurrentExecutable
	m.settings = state.Settings
	m.alertStatus = state.AlertStatus
	m.alertReminders = state.AlertReminders
	m.exportAssets = state.ExportAssets
	m.dailyCheckIn = state.DailyCheckIn
	return m
}

func (m Model) inputDeps() inputpkg.Deps {
	return inputpkg.Deps{
		CloseEventStop:     func() { m.stopEventStream() },
		ShutdownKernel:     func() tea.Cmd { return commands.ShutdownKernel(m.client) },
		SeedDevData:        func() tea.Cmd { return commands.SeedDevData(m.client) },
		ClearDevData:       func() tea.Cmd { return commands.ClearDevData(m.client) },
		PrepareLocalUpdate: func() tea.Cmd { return commands.PrepareLocalUpdate(m.client) },
		IsDevMode:          func(state inputpkg.State) bool { return m.applyInputState(state).isDevMode() },
		NextActiveSessionView: func(state inputpkg.State, dir int) uistate.View {
			return m.applyInputState(state).nextActiveSessionView(dir)
		},
		NextWorkspaceView: func(state inputpkg.State, dir int) uistate.View {
			return m.applyInputState(state).nextWorkspaceView(dir)
		},
		DefaultPane: uistate.DefaultPane,
		NextPane:    nextPane,
		SetDefaultIssueSection: func(state *inputpkg.State, section uistate.DefaultIssueSection) {
			next := m.applyInputState(*state)
			next.setDefaultIssueSection(section)
			*state = next.inputState()
		},
		ListLen: func(state inputpkg.State, pane uistate.Pane) int {
			next := m.applyInputState(state)
			return (&next).listLen(pane)
		},
		LoadExportAssets: func() tea.Cmd { return commands.LoadExportAssets(m.client) },
		SetStatus: func(state *inputpkg.State, message string, isErr bool) tea.Cmd {
			next := m.applyInputState(*state)
			cmd := next.setStatus(message, isErr)
			*state = next.inputState()
			return cmd
		},
		LoadDailySummary:     func(date string) tea.Cmd { return commands.LoadDailySummary(m.client, date) },
		LoadDueHabits:        func(date string) tea.Cmd { return commands.LoadDueHabits(m.client, date) },
		CurrentDashboardDate: func(state inputpkg.State) string { return m.applyInputState(state).currentDashboardDate() },
		LoadRollupSummaries:  func(start, end string) tea.Cmd { return commands.LoadRollupSummaries(m.client, start, end) },
		CurrentRollupStartDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupStartDate()
		},
		CurrentRollupEndDate: func(state inputpkg.State) string {
			return m.applyInputState(state).currentRollupEndDate()
		},
		LoadWellbeing:        func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		CurrentWellbeingDate: func(state inputpkg.State) string { return m.applyInputState(state).currentWellbeingDate() },
		ConfigChangeSelected: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if item, ok := selectionpkg.SelectedConfigItem(snapshot); ok && next.exportAssets != nil {
				switch {
				case item.PresetStyle:
					for _, asset := range next.exportAssets.TemplateAssets {
						if asset.ReportKind != item.ReportKind || asset.AssetKind != item.AssetKind || len(asset.Presets) == 0 {
							continue
						}
						currentID := ""
						if asset.SelectedPreset != nil {
							currentID = asset.SelectedPreset.ID
						}
						nextID := currentID
						for idx, preset := range asset.Presets {
							if preset.ID == currentID {
								nextID = asset.Presets[(idx+1)%len(asset.Presets)].ID
								break
							}
						}
						if nextID == "" && len(asset.Presets) > 0 {
							nextID = asset.Presets[0].ID
						}
						*state = next.inputState()
						return commands.ApplyExportTemplatePreset(m.client, item.ReportKind, item.AssetKind, nextID)
					}
				case item.Label == "Reports directory":
					next = next.openExportReportsDirDialog(next.exportAssets.ReportsDir)
				case item.Label == "ICS export directory":
					next = next.openExportICSDirDialog(next.exportAssets.ICSDir)
				default:
					return nil
				}
				*state = next.inputState()
				return nil
			}
			return nil
		},
		OpenCheckoutContextDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openCheckoutContextDialog()
			*state = next.inputState()
			return true
		},
		Checkout: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			model, cmd := next.checkout()
			*state = model.inputState()
			return cmd
		},
		CheckUpdateNow:             func() tea.Cmd { return commands.CheckUpdateNow(m.client) },
		SelfUpdateInstallAvailable: func(state inputpkg.State) bool { return m.applyInputState(state).selfUpdateInstallAvailable() },
		SelfUpdateUnsupportedReason: func(state inputpkg.State) string {
			return m.applyInputState(state).selfUpdateUnsupportedReason()
		},
		InstallUpdate: func(state inputpkg.State) tea.Cmd {
			next := m.applyInputState(state)
			return commands.InstallUpdate(next.updateStatus, next.selfUpdateInstallAvailable(), next.selfUpdateUnsupportedReason())
		},
		DismissUpdate: func() tea.Cmd { return commands.DismissUpdate(m.client) },
		ResumeSession: func() tea.Cmd { return commands.ResumeFocusSession(m.client) },
		PauseSession:  func() tea.Cmd { return commands.PauseFocusSession(m.client) },
		OpenEndSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openSessionMessageDialog("end_session")
			*state = next.inputState()
			return true
		},
		OpenStashSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openSessionMessageDialog("stash_session")
			*state = next.inputState()
			return true
		},
		CanOpenStashList: func(state inputpkg.State) bool {
			next := m.applyInputState(state)
			return (next.timer == nil || next.timer.State == "idle") && next.dialog == ""
		},
		OpenStashListDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openStashListDialog()
			*state = next.inputState()
			return true
		},
		LoadStashes: func() tea.Cmd { return commands.LoadStashes(m.client) },
		ClampFiltered: func(state *inputpkg.State, pane uistate.Pane) {
			next := m.applyInputState(*state)
			filterState := next.filterState()
			if deps := next.filterDeps(); deps.Clamp != nil && deps.ItemCount != nil {
				deps.Clamp(filterState.Cursor, pane, deps.ItemCount(filterState, pane))
			}
			next = next.applyFilterState(filterState)
			*state = next.inputState()
		},
		CurrentOpsLimit: func(state inputpkg.State) int {
			next := m.applyInputState(state)
			return (&next).currentOpsLimit()
		},
		LoadOps: func(limit int) tea.Cmd { return commands.LoadOps(m.client, limit) },
		StartFilterEdit: func(state *inputpkg.State, pane uistate.Pane) {
			next := m.applyInputState(*state)
			next = next.applyFilterState(filteringpkg.Start(next.filterState(), pane))
			*state = next.inputState()
		},
		OpenIssueStatusFromSelection: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				next = next.withDialogState(dialogstate.OpenIssueStatus(next.dialogSnapshot(), issue.Status))
				*state = next.inputState()
				return true
			}
			return false
		},
		AbandonSelectedIssue: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			issue, ok := selectionpkg.SelectedIssueDetail(snapshot)
			if !ok {
				return nil
			}
			if issue.Status == "done" {
				return next.setStatus("Done issues cannot be abandoned", true)
			}
			if issue.Status == "abandoned" {
				return nil
			}
			if next.timer != nil && next.timer.State != "idle" {
				next = next.withDialogState(dialogstate.OpenIssueSessionTransition(next.dialogSnapshot(), issue.ID, "abandoned"))
				*state = next.inputState()
				return nil
			}
			next = next.withDialogState(dialogstate.OpenIssueStatusNote(next.dialogSnapshot(), "abandoned", "Abandon reason", true))
			next.dialogIssueID = issue.ID
			next.dialogStreamID = issue.StreamID
			*state = next.inputState()
			return nil
		},
		ToggleSelectedIssueToday: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				date := ""
				if issue.TodoForDate != nil && *issue.TodoForDate == next.currentDashboardDate() {
					date = ""
				} else {
					date = next.currentDashboardDate()
				}
				return commands.SetIssueTodoDate(m.client, issue.ID, date, issue.StreamID, next.currentDashboardDate())
			}
			return nil
		},
		OpenSelectedIssueTodoDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				next = next.withDialogState(dialogstate.OpenDatePicker(next.dialogSnapshot(), "", issue.ID, 0, issue.TodoForDate))
			}
			*state = next.inputState()
			return true
		},
		HandleCreateAction: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.handleInputCreateAction()
			*state = next.inputState()
			return true
		},
		OpenExportDailyDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openExportDailyDialog()
			*state = next.inputState()
			return true
		},
		OpenEditorAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputOpenEditor()
			*state = next.inputState()
			return cmd, handled
		},
		DeleteSelectionAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputDeleteSelection()
			*state = next.inputState()
			return cmd, handled
		},
		OpenSelectionAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputOpenSelection()
			*state = next.inputState()
			return cmd, handled
		},
		EnterAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			next, cmd, handled := next.handleInputEnter()
			*state = next.inputState()
			return cmd, handled
		},
		ToggleHabitCompletedAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputToggleHabitCompleted()
			*state = next.inputState()
			return cmd, handled
		},
		SetHabitFailedAction: func(state *inputpkg.State) (tea.Cmd, bool) {
			next := m.applyInputState(*state)
			cmd, handled := next.handleInputSetHabitFailed()
			*state = next.inputState()
			return cmd, handled
		},
		StartFocusFromSelection: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			model, cmd := next.handleInputStartFocusFromSelection()
			*state = model.(Model).inputState()
			return cmd
		},
		OpenManualSessionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			if next.dialog != "" {
				return false
			}
			if next.view == ViewDaily && next.pane == PaneHabits {
				if habit, ok := next.selectedDailyHabitRecord(); ok {
					next = next.openHabitCompletionDialog(habit.ID, next.currentDashboardDate(), habit.DurationMinutes, habit.Notes)
					*state = next.inputState()
					return true
				}
				return false
			}
			snapshot := next.selectionSnapshot()
			if issue, ok := selectionpkg.SelectedIssueDetail(snapshot); ok {
				issueLabel := ""
				var estimateMinutes *int
				if meta := selectionpkg.IssueMetaByID(snapshot, issue.ID); meta != nil {
					issueLabel = meta.Title
					estimateMinutes = meta.EstimateMinutes
				}
				next = next.openManualSessionDialog(issue.ID, issueLabel, estimateMinutes, next.currentDashboardDate())
				*state = next.inputState()
				return true
			}
			return false
		},
		OpenSessionContextOverlay: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			if next.view != ViewSessionActive || next.timer == nil || next.timer.State == "idle" {
				return false
			}
			snapshot := next.selectionSnapshot()
			if selectionpkg.ActiveIssue(snapshot) == nil {
				return false
			}
			next.sessionContextOpen = true
			next.sessionContextY = 0
			*state = next.inputState()
			return true
		},
		ConfigReset: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			cmd, _ := next.handleInputConfigReset()
			*state = next.inputState()
			return cmd
		},
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.PatchSetting(m.client, key, value, repoID, streamID, dashboardDate)
		},
		TestAlertNotification: func() tea.Cmd { return commands.TestAlertNotification(m.client) },
		TestAlertSound:        func() tea.Cmd { return commands.TestAlertSound(m.client) },
		CreateAlertReminder: func(input shareddto.AlertReminderCreateRequest) tea.Cmd {
			return commands.CreateAlertReminder(m.client, input)
		},
		UpdateAlertReminder: func(input shareddto.AlertReminderUpdateRequest) tea.Cmd {
			return commands.UpdateAlertReminder(m.client, input)
		},
		ToggleAlertReminder: func(id string, enabled bool) tea.Cmd {
			return commands.ToggleAlertReminder(m.client, id, enabled)
		},
		DeleteAlertReminder: func(id string) tea.Cmd {
			return commands.DeleteAlertReminder(m.client, id)
		},
		OpenCreateAlertReminderDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openCreateAlertReminderDialog()
			*state = next.inputState()
			return true
		},
		OpenEditAlertReminderDialog: func(state *inputpkg.State, id string) bool {
			next := m.applyInputState(*state)
			next = next.openEditAlertReminderDialog(id)
			*state = next.inputState()
			return true
		},
		OpenEditRestProtectionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditRestProtectionDialog()
			*state = next.inputState()
			return true
		},
		OpenConfirmWipeDataDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openConfirmWipeDataDialog()
			*state = next.inputState()
			return true
		},
		OpenConfirmUninstallDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openConfirmUninstallDialog()
			*state = next.inputState()
			return true
		},
		WipeRuntimeData:           func() tea.Cmd { return commands.WipeRuntimeData(m.client) },
		OpenSupportIssueURL:       func() tea.Cmd { return m.openSupportIssueURL() },
		OpenSupportDiscussionsURL: func() tea.Cmd { return m.openSupportDiscussionsURL() },
		OpenSupportReleasesURL:    func() tea.Cmd { return m.openSupportReleasesURL() },
		OpenSupportRoadmapURL:     func() tea.Cmd { return m.openSupportRoadmapURL() },
		CopySupportDiagnostics:    func(state inputpkg.State) tea.Cmd { return m.copySupportDiagnosticsCmd(state) },
		GenerateSupportBundle:     func(state inputpkg.State) tea.Cmd { return m.generateSupportBundleCmd(state) },
		OpenViewJumpDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openViewJumpDialog()
			*state = next.inputState()
			return true
		},
		OpenBetaSupportDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openBetaSupportDialog()
			*state = next.inputState()
			return true
		},
		OpenRollupStartDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openDatePickerDialog("rollup_start", 0, 0, dialogpkg.ValueToPointer(next.currentRollupStartDate()))
			*state = next.inputState()
			return true
		},
		OpenRollupEndDateDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openDatePickerDialog("rollup_end", 0, 0, dialogpkg.ValueToPointer(next.currentRollupEndDate()))
			*state = next.inputState()
			return true
		},
	}
}

func (m Model) dialogSnapshot() dialogstate.Snapshot {
	selectionSnapshot := m.selectionSnapshot()
	dialogSnapshot := dialogstate.Snapshot{
		Dialog:               m.dialogState(),
		Repos:                m.repos,
		Streams:              m.streams,
		AllIssues:            m.allIssues,
		Context:              m.context,
		Stashes:              m.stashes,
		DailyCheckIn:         m.dailyCheckIn,
		UpdateStatus:         m.updateStatus,
		ExportAssets:         m.exportAssets,
		Settings:             m.settings,
		AlertReminders:       m.alertReminders,
		CurrentDashboardDate: m.currentDashboardDate(),
		CurrentWellbeingDate: m.currentWellbeingDate(),
		HasActiveTimer:       m.timer != nil && m.timer.State != "idle",
		AvailableViews:       m.jumpAvailableViews(),
	}
	dialogSnapshot.ProtectedModeActive, _, _ = viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02"))
	if issue, ok := selectionpkg.SelectedIssueDetail(selectionSnapshot); ok {
		dialogSnapshot.SelectedIssueID = issue.ID
		dialogSnapshot.SelectedStreamID = issue.StreamID
		dialogSnapshot.HasSelectedIssue = true
	}
	if issue := selectionpkg.ActiveIssue(selectionSnapshot); issue != nil {
		dialogSnapshot.ActiveIssueStream = issue.StreamID
		dialogSnapshot.HasActiveIssue = true
	}
	return dialogSnapshot
}

func (m Model) dialogActionCmd(action dialogpkg.Action) tea.Cmd {
	return dialogruntime.Resolve(action, m.dialogRuntimeState(), m.dialogRuntimeDeps())
}

func (m Model) handleDialogAction(next Model, action dialogpkg.Action) (Model, tea.Cmd) {
	switch action.Kind {
	case "jump_view":
		target := View(strings.TrimSpace(action.TargetView))
		if target == "" {
			return next, nil
		}
		if !next.canJumpToView(target) {
			return next, next.setStatus("That view is not available right now", true)
		}
		if target == ViewSessionActive && (next.timer == nil || next.timer.State == "idle") {
			return next, next.setStatus("Active session view is only available while a timer is running", true)
		}
		if target == ViewAway {
			protected, _, _ := viewruntime.ProtectedRestMode(next.settings, time.Now().Format("2006-01-02"))
			if !protected {
				return next, next.setStatus("Away view is only available when away mode or rest protection is active", true)
			}
		}
		next.view = target
		next.pane = uistate.DefaultPane(target)
		return next, nil
	case "continue_focus_fresh":
		return next, commands.ContinueFocusSessionFresh(next.client, action.RepoID, action.StreamID, action.IssueID)
	case "copy_support_diagnostics":
		return next, next.copySupportDiagnosticsCmd(next.inputState())
	case "generate_support_bundle":
		return next, next.generateSupportBundleCmd(next.inputState())
	default:
		return next, next.dialogActionCmd(action)
	}
}

func (m Model) openCreateScratchpad() Model {
	return m.withDialogState(dialogstate.OpenCreateScratchpad(m.dialogSnapshot()))
}
func (m Model) openViewJumpDialog() Model {
	return m.withDialogState(dialogstate.OpenViewJump(m.dialogSnapshot()))
}
func (m Model) openBetaSupportDialog() Model {
	return m.withDialogState(dialogstate.OpenBetaSupport(m.dialogSnapshot()))
}
func (m Model) openStashConflictDialog(conflict api.StashConflict, repoID, streamID, issueID int64) Model {
	state := dialogstate.OpenStashConflict(m.dialogSnapshot(), conflict)
	state.RepoID = repoID
	state.StreamID = streamID
	state.IssueID = issueID
	return m.withDialogState(state)
}
func (m Model) openCreateRepoDialog() Model {
	return m.withDialogState(dialogstate.OpenCreateRepo(m.dialogSnapshot()))
}
func (m Model) openEditRepoDialog(repoID int64, name string) Model {
	return m.withDialogState(dialogstate.OpenEditRepo(m.dialogSnapshot(), repoID, name, m.repoDescriptionByID(repoID)))
}
func (m Model) openCreateStreamDialog(repoID int64, repoName string) Model {
	return m.withDialogState(dialogstate.OpenCreateStream(m.dialogSnapshot(), repoID, repoName))
}
func (m Model) openEditStreamDialog(streamID, repoID int64, streamName, repoName string) Model {
	return m.withDialogState(dialogstate.OpenEditStream(m.dialogSnapshot(), streamID, repoID, streamName, repoName, m.streamDescriptionByID(streamID)))
}
func (m Model) openCreateIssueMetaDialog(streamID int64, streamName, repoName string) Model {
	return m.withDialogState(dialogstate.OpenCreateIssueMeta(m.dialogSnapshot(), streamID, streamName, repoName))
}
func (m Model) openCreateHabitDialog(streamID int64, streamName, repoName string) Model {
	return m.withDialogState(dialogstate.OpenCreateHabit(m.dialogSnapshot(), streamID, streamName, repoName))
}
func (m Model) openEditIssueDialog(issueID, streamID int64, title string, description *string, estimateMinutes *int, todoForDate *string) Model {
	return m.withDialogState(dialogstate.OpenEditIssue(m.dialogSnapshot(), issueID, streamID, title, description, estimateMinutes, todoForDate))
}
func (m Model) openEditHabitDialog(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, targetMinutes *int, active bool) Model {
	return m.withDialogState(dialogstate.OpenEditHabit(m.dialogSnapshot(), habitID, streamID, name, description, schedule, weekdays, targetMinutes, active))
}
func (m Model) openHabitCompletionDialog(habitID int64, date string, durationMinutes *int, notes *string) Model {
	return m.withDialogState(dialogstate.OpenHabitCompletion(m.dialogSnapshot(), habitID, date, durationMinutes, notes))
}
func (m Model) openCreateIssueDefaultDialog() Model {
	return m.withDialogState(dialogstate.OpenCreateIssueDefault(m.dialogSnapshot()))
}
func (m Model) openCheckoutContextDialog() Model {
	return m.withDialogState(dialogstate.OpenCheckoutContext(m.dialogSnapshot()))
}
func (m Model) openCreateCheckInDialog() Model {
	return m.withDialogState(dialogstate.OpenCreateCheckIn(m.dialogSnapshot()))
}
func (m Model) openEditCheckInDialog() Model {
	return m.withDialogState(dialogstate.OpenEditCheckIn(m.dialogSnapshot()))
}
func (m Model) openConfirmDelete(id string) Model {
	return m.withDialogState(dialogstate.OpenConfirmDelete(m.dialogSnapshot(), id))
}
func (m Model) openConfirmDeleteEntity(kind, id, label string) Model {
	return m.withDialogState(dialogstate.OpenConfirmDeleteEntity(m.dialogSnapshot(), kind, id, label))
}
func (m Model) openStashListDialog() Model {
	return m.withDialogState(dialogstate.OpenStashList(m.dialogSnapshot()))
}
func (m Model) openIssueStatusDialog(status string) Model {
	return m.withDialogState(dialogstate.OpenIssueStatus(m.dialogSnapshot(), status))
}
func (m Model) openIssueStatusNoteDialog(status, label string, required bool) Model {
	return m.withDialogState(dialogstate.OpenIssueStatusNote(m.dialogSnapshot(), status, label, required))
}
func (m Model) openSessionMessageDialog(kind string) Model {
	return m.withDialogState(dialogstate.OpenSessionMessage(m.dialogSnapshot(), kind))
}
func (m Model) openIssueSessionTransitionDialog(issueID int64, status string) Model {
	return m.withDialogState(dialogstate.OpenIssueSessionTransition(m.dialogSnapshot(), issueID, status))
}
func (m Model) openAmendSessionDialog(sessionID string, commit string) Model {
	return m.withDialogState(dialogstate.OpenAmendSession(m.dialogSnapshot(), sessionID, commit))
}
func (m Model) openManualSessionDialog(issueID int64, issueLabel string, estimateMinutes *int, date string) Model {
	return m.withDialogState(dialogstate.OpenManualSession(m.dialogSnapshot(), issueID, issueLabel, estimateMinutes, date))
}
func (m Model) openDatePickerDialog(parentDialog string, issueID int64, inputIndex int, initial *string) Model {
	return m.withDialogState(dialogstate.OpenDatePicker(m.dialogSnapshot(), parentDialog, issueID, inputIndex, initial))
}
func (m Model) openViewEntityDialog(title string, name string, meta string, body string) Model {
	return m.withDialogState(dialogstate.OpenViewEntity(m.dialogSnapshot(), title, name, meta, body))
}
func (m Model) openSupportBundleDialog(path string, sizeBytes int64, windowLabel string) Model {
	meta := strings.Join([]string{
		"Size " + helperpkg.HumanizeSupportBytes(sizeBytes),
		"Window " + strings.TrimSpace(windowLabel),
		"Redaction safe",
	}, "   ")
	body := strings.Join([]string{
		"Bundle",
		path,
		"",
		"Attach this zip to a bug report if you need deeper diagnostics.",
		"Use o to open the folder, c to copy the path, or g to open the issue tracker.",
	}, "\n")
	return m.withDialogState(dialogstate.OpenSupportBundleResult(m.dialogSnapshot(), helperpkg.SupportBundleDisplayName(path), meta, body, path))
}
func (m Model) openUpdateNotesDialog() Model {
	return m.withDialogState(dialogstate.OpenUpdateNotes(m.dialogSnapshot()))
}
func (m Model) openExportDailyDialog() Model {
	return m.withDialogState(dialogstate.OpenExportDaily(m.dialogSnapshot()))
}
func (m Model) openExportReportsDirDialog(current string) Model {
	return m.withDialogState(dialogstate.OpenExportReportsDir(m.dialogSnapshot(), current))
}
func (m Model) openExportICSDirDialog(current string) Model {
	return m.withDialogState(dialogstate.OpenExportICSDir(m.dialogSnapshot(), current))
}
func (m Model) openCreateAlertReminderDialog() Model {
	return m.withDialogState(dialogstate.OpenCreateAlertReminder(m.dialogSnapshot()))
}
func (m Model) openEditAlertReminderDialog(id string) Model {
	return m.withDialogState(dialogstate.OpenEditAlertReminder(m.dialogSnapshot(), id))
}
func (m Model) openEditRestProtectionDialog() Model {
	return m.withDialogState(dialogstate.OpenEditRestProtection(m.dialogSnapshot()))
}
func (m Model) openConfirmWipeDataDialog() Model {
	return m.withDialogState(dialogstate.OpenConfirmWipeData(m.dialogSnapshot()))
}
func (m Model) openConfirmUninstallDialog() Model {
	return m.withDialogState(dialogstate.OpenConfirmUninstall(m.dialogSnapshot()))
}

func (m Model) dialogState() dialogpkg.State {
	return dialogpkg.State{
		Kind:               m.dialog,
		Width:              m.width,
		Inputs:             m.dialogInputs,
		Description:        m.dialogDescription,
		DescriptionEnabled: m.dialogDescriptionOn,
		DescriptionIndex:   m.dialogDescriptionIdx,
		FocusIdx:           m.dialogFocusIdx,
		ErrorMessage:       m.dialogErrorMessage,
		DeleteID:           m.dialogDeleteID,
		DeleteKind:         m.dialogDeleteKind,
		DeleteLabel:        m.dialogDeleteLabel,
		SessionID:          m.dialogSessionID,
		IssueID:            m.dialogIssueID,
		HabitID:            m.dialogHabitID,
		IssueStatus:        m.dialogIssueStatus,
		CheckInDate:        m.dialogCheckInDate,
		RepoID:             m.dialogRepoID,
		RepoName:           m.dialogRepoName,
		RepoItems:          m.dialogRepoItems,
		RepoItemIDs:        m.dialogRepoItemIDs,
		StreamID:           m.dialogStreamID,
		StreamName:         m.dialogStreamName,
		RepoIndex:          m.dialogRepoIndex,
		StreamIndex:        m.dialogStreamIndex,
		Parent:             m.dialogParent,
		DateMonthValue:     m.dialogDateMonth,
		DateCursorValue:    m.dialogDateCursor,
		StashCursor:        m.dialogStashCursor,
		StatusItems:        m.dialogStatusItems,
		StatusCursor:       m.dialogStatusCursor,
		ChoiceItems:        m.dialogChoiceItems,
		ChoiceValues:       m.dialogChoiceValues,
		ChoiceDetails:      m.dialogChoiceDetails,
		TemplateAssets:     m.dialogTemplateAssets,
		ChoiceCursor:       m.dialogChoiceCursor,
		Processing:         m.dialogProcessing,
		ProcessingLabel:    m.dialogProcessingLabel,
		StatusLabel:        m.dialogStatusLabel,
		StatusRequired:     m.dialogStatusRequired,
		ViewTitle:          m.dialogViewTitle,
		ViewName:           m.dialogViewName,
		IssueEstimateMins:  m.dialogIssueEstimateMins,
		ReminderID:         m.dialogReminderID,
		ReminderKind:       m.dialogReminderKind,
		ViewMeta:           m.dialogViewMeta,
		ViewBody:           m.dialogViewBody,
		SupportBundlePath:  m.dialogSupportBundlePath,
		ProtectionStep:     m.dialogProtectionStep,
		ProtectionCursor:   m.dialogProtectionCursor,
		ProtectionStreaks:  m.dialogProtectionStreaks,
		ProtectionWeekdays: m.dialogProtectionWeekdays,
		ProtectionDates:    m.dialogProtectionDates,
		ExportPresetKind:   m.dialogExportPresetKind,
		ExportPresetFormat: m.dialogExportPresetFormat,
		ExportPresetOutput: m.dialogExportPresetOutput,
		ExportIncludePDF:   m.dialogExportIncludePDF,
	}
}

func (m Model) withDialogState(state dialogpkg.State) Model {
	m.dialog = state.Kind
	m.dialogInputs = state.Inputs
	m.dialogDescription = state.Description
	m.dialogDescriptionOn = state.DescriptionEnabled
	m.dialogDescriptionIdx = state.DescriptionIndex
	m.dialogFocusIdx = state.FocusIdx
	m.dialogErrorMessage = state.ErrorMessage
	m.dialogDeleteID = state.DeleteID
	m.dialogDeleteKind = state.DeleteKind
	m.dialogDeleteLabel = state.DeleteLabel
	m.dialogSessionID = state.SessionID
	m.dialogIssueID = state.IssueID
	m.dialogHabitID = state.HabitID
	m.dialogIssueStatus = state.IssueStatus
	m.dialogCheckInDate = state.CheckInDate
	m.dialogRepoID = state.RepoID
	m.dialogRepoName = state.RepoName
	m.dialogRepoItems = state.RepoItems
	m.dialogRepoItemIDs = state.RepoItemIDs
	m.dialogStreamID = state.StreamID
	m.dialogStreamName = state.StreamName
	m.dialogRepoIndex = state.RepoIndex
	m.dialogStreamIndex = state.StreamIndex
	m.dialogParent = state.Parent
	m.dialogDateMonth = state.DateMonthValue
	m.dialogDateCursor = state.DateCursorValue
	m.dialogStashCursor = state.StashCursor
	m.dialogStatusItems = state.StatusItems
	m.dialogStatusCursor = state.StatusCursor
	m.dialogChoiceItems = state.ChoiceItems
	m.dialogChoiceValues = state.ChoiceValues
	m.dialogChoiceDetails = state.ChoiceDetails
	m.dialogTemplateAssets = state.TemplateAssets
	m.dialogChoiceCursor = state.ChoiceCursor
	m.dialogProcessing = state.Processing
	m.dialogProcessingLabel = state.ProcessingLabel
	m.dialogStatusLabel = state.StatusLabel
	m.dialogStatusRequired = state.StatusRequired
	m.dialogViewTitle = state.ViewTitle
	m.dialogViewName = state.ViewName
	m.dialogIssueEstimateMins = state.IssueEstimateMins
	m.dialogReminderID = state.ReminderID
	m.dialogReminderKind = state.ReminderKind
	m.dialogViewMeta = state.ViewMeta
	m.dialogViewBody = state.ViewBody
	m.dialogSupportBundlePath = state.SupportBundlePath
	m.dialogProtectionStep = state.ProtectionStep
	m.dialogProtectionCursor = state.ProtectionCursor
	m.dialogProtectionStreaks = state.ProtectionStreaks
	m.dialogProtectionWeekdays = state.ProtectionWeekdays
	m.dialogProtectionDates = state.ProtectionDates
	m.dialogExportPresetKind = state.ExportPresetKind
	m.dialogExportPresetFormat = state.ExportPresetFormat
	m.dialogExportPresetOutput = state.ExportPresetOutput
	m.dialogExportIncludePDF = state.ExportIncludePDF
	return m
}

func (m Model) dialogRuntimeState() dialogruntime.State {
	return dialogruntime.State{
		Context:               m.context,
		Repos:                 m.repos,
		DashboardDate:         m.currentDashboardDate(),
		RollupStartDate:       m.currentRollupStartDate(),
		RollupEndDate:         m.currentRollupEndDate(),
		CurrentExecutablePath: m.currentExecutablePath,
		KernelExecutablePath:  kernelExecutablePath(m.kernelInfo),
		KernelInfo:            m.kernelInfo,
	}
}

func (m Model) dialogRuntimeDeps() dialogruntime.Deps {
	return dialogruntime.Deps{
		CreateScratchpad: func(name, path string) tea.Cmd { return commands.CreateScratchpad(m.client, name, path) },
		CreateRepo: func(name string, description *string) tea.Cmd {
			return commands.CreateRepoOnly(m.client, name, description)
		},
		UpdateRepo: func(repoID int64, name string, description *string) tea.Cmd {
			return commands.UpdateRepo(m.client, repoID, name, description)
		},
		CreateStream: func(repoID int64, name string, description *string) tea.Cmd {
			return commands.CreateStreamOnly(m.client, repoID, name, description)
		},
		UpdateStream: func(repoID, streamID int64, name string, description *string) tea.Cmd {
			return commands.UpdateStream(m.client, repoID, streamID, name, description)
		},
		CreateIssueOnly: func(streamID int64, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd {
			return commands.CreateIssueOnly(m.client, streamID, title, description, estimateMinutes, dueDate)
		},
		CreateHabitWithPath: func(repoName, streamName, name string, description *string, schedule string, weekdays []int, estimateMinutes *int) tea.Cmd {
			return commands.CreateHabitWithPath(m.client, repoName, "", streamName, "", name, description, schedule, weekdays, estimateMinutes)
		},
		UpdateHabit: func(habitID, streamID int64, name string, description *string, schedule string, weekdays []int, estimateMinutes *int, active bool, dashboardDate string) tea.Cmd {
			return commands.UpdateHabit(m.client, habitID, streamID, name, description, schedule, weekdays, estimateMinutes, active, dashboardDate)
		},
		CreateIssueWithPath: func(repoName, streamName, title string, description *string, estimateMinutes *int, dueDate *string) tea.Cmd {
			return commands.CreateIssueWithPath(m.client, repoName, "", streamName, "", title, description, estimateMinutes, dueDate)
		},
		CheckoutContext: func(repoID int64, repoName string, streamID int64, streamName string) tea.Cmd {
			return commands.CheckoutContext(m.client, repoID, repoName, streamID, streamName)
		},
		UpsertDailyCheckIn: func(req shareddto.DailyCheckInUpsertRequest, refreshDate string) tea.Cmd {
			return commands.UpsertDailyCheckIn(m.client, req, refreshDate)
		},
		UpdateIssue: func(issueID, streamID int64, title string, description *string, estimateMinutes *int, dueDate *string, dashboardDate string) tea.Cmd {
			return commands.UpdateIssue(m.client, issueID, streamID, title, description, estimateMinutes, dueDate, dashboardDate)
		},
		SetHabitStatus: func(habitID int64, date string, status sharedtypes.HabitCompletionStatus, estimateMinutes *int, note *string) tea.Cmd {
			return commands.SetHabitStatus(m.client, habitID, date, status, estimateMinutes, note)
		},
		CopyDailyReport: func(date string) tea.Cmd { return commands.CopyDailyReport(m.client, date) },
		GenerateCalendarExport: func(req shareddto.ExportCalendarRequest) tea.Cmd {
			return commands.GenerateCalendarExport(m.client, req)
		},
		GenerateReport:      func(req shareddto.ExportReportRequest) tea.Cmd { return commands.GenerateReport(m.client, req) },
		SetExportReportsDir: func(path string) tea.Cmd { return commands.SetExportReportsDir(m.client, path) },
		SetExportICSDir:     func(path string) tea.Cmd { return commands.SetExportICSDir(m.client, path) },
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.PatchSetting(m.client, key, value, repoID, streamID, dashboardDate)
		},
		CreateAlertReminder: func(input shareddto.AlertReminderCreateRequest) tea.Cmd {
			return commands.CreateAlertReminder(m.client, input)
		},
		UpdateAlertReminder: func(input shareddto.AlertReminderUpdateRequest) tea.Cmd {
			return commands.UpdateAlertReminder(m.client, input)
		},
		DeleteRepo:   func(id int64) tea.Cmd { return commands.DeleteRepo(m.client, id) },
		DeleteStream: func(repoID, streamID int64) tea.Cmd { return commands.DeleteStream(m.client, repoID, streamID) },
		DeleteIssue: func(issueID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.DeleteIssue(m.client, issueID, streamID, dashboardDate)
		},
		DeleteHabit: func(habitID, streamID int64, dashboardDate string) tea.Cmd {
			return commands.DeleteHabit(m.client, habitID, streamID, dashboardDate)
		},
		DeleteDailyCheckIn: func(id string) tea.Cmd { return commands.DeleteDailyCheckIn(m.client, id) },
		DeleteExportReport: func(report api.ExportReportFile) tea.Cmd { return commands.DeleteExportReport(m.client, report) },
		DeleteScratchpad:   func(id string) tea.Cmd { return commands.DeleteScratchpad(m.client, id) },
		ApplyStash:         func(id string) tea.Cmd { return commands.ApplyStash(m.client, id) },
		DropStash:          func(id string) tea.Cmd { return commands.DropStash(m.client, id) },
		ChangeIssueStatus: func(issueID int64, status string, note *string, streamID int64, dashboardDate string) tea.Cmd {
			return commands.ChangeIssueStatus(m.client, issueID, status, note, streamID, dashboardDate)
		},
		AmendSessionNote: func(id string, note string) tea.Cmd { return commands.AmendSessionNote(m.client, id, note) },
		LogManualSession: func(input shareddto.ManualSessionLogRequest) tea.Cmd {
			return commands.LogManualSession(m.client, input)
		},
		EndFocusSession: func(streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd {
			return commands.EndFocusSession(m.client, streamID, dashboardDate, payload)
		},
		StashFocusSession: func(note string) tea.Cmd { return commands.StashFocusSession(m.client, note) },
		ChangeIssueStatusAndEndSession: func(issueID int64, status string, note *string, streamID int64, dashboardDate string, payload shareddto.EndSessionRequest) tea.Cmd {
			return commands.ChangeIssueStatusAndEndSession(m.client, issueID, status, note, streamID, dashboardDate, payload)
		},
		SetIssueTodoDate: func(issueID int64, date string, streamID int64, dashboardDate string) tea.Cmd {
			return commands.SetIssueTodoDate(m.client, issueID, date, streamID, dashboardDate)
		},
		SetRollupStartDate: func(date, currentEnd string) tea.Cmd {
			if date > currentEnd {
				currentEnd = date
			}
			return commands.SetRollupRange(date, currentEnd)
		},
		SetRollupEndDate: func(currentStart, date string) tea.Cmd {
			if currentStart > date {
				currentStart = date
			}
			return commands.SetRollupRange(currentStart, date)
		},
		WipeRuntimeData: func() tea.Cmd { return commands.WipeRuntimeData(m.client) },
		UninstallCrona: func(currentExecutablePath, kernelExecutablePath string, kernelInfo *api.KernelInfo) tea.Cmd {
			return commands.UninstallCrona(m.client, currentExecutablePath, kernelExecutablePath, kernelInfo)
		},
		OpenSupportIssueURL:       func() tea.Cmd { return m.openSupportIssueURL() },
		OpenSupportDiscussionsURL: func() tea.Cmd { return m.openSupportDiscussionsURL() },
		OpenSupportReleasesURL:    func() tea.Cmd { return m.openSupportReleasesURL() },
		OpenSupportRoadmapURL:     func() tea.Cmd { return m.openSupportRoadmapURL() },
		OpenExternalPath: func(path string) tea.Cmd {
			return commands.OpenExternalPath(path)
		},
		CopyText: func(text, message string) tea.Cmd { return commands.CopyTextToClipboard(text, message) },
		ErrorCmd: func(err error) tea.Cmd { return func() tea.Msg { return commands.ErrMsg{Err: err} } },
		ResolvePatchSettingValue: func(action dialogpkg.Action) any {
			switch action.SettingKey {
			case sharedtypes.CoreSettingsKeyRestWeekdays:
				return action.IntList
			default:
				return action.StringList
			}
		},
	}
}

func loadSessionHistoryForModel(m Model, limit int) tea.Cmd {
	return commands.LoadSessionHistory(m.client, helperpkg.SessionHistoryScopeIssueID(m.timer), limit)
}
