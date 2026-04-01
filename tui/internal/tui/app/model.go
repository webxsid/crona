package app

import (
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
	"crona/tui/internal/tui/views"

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
	ViewMeta           = uistate.ViewMeta
	ViewSessionHistory = uistate.ViewSessionHistory
	ViewSessionActive  = uistate.ViewSessionActive
	ViewScratch        = uistate.ViewScratch
	ViewOps            = uistate.ViewOps
	ViewWellbeing      = uistate.ViewWellbeing
	ViewReports        = uistate.ViewReports
	ViewConfig         = uistate.ViewConfig
	ViewSettings       = uistate.ViewSettings
	ViewUpdates        = uistate.ViewUpdates
)

type Pane = uistate.Pane

const (
	PaneRepos         = uistate.PaneRepos
	PaneStreams       = uistate.PaneStreams
	PaneIssues        = uistate.PaneIssues
	PaneHabits        = uistate.PaneHabits
	PaneSessions      = uistate.PaneSessions
	PaneScratchpads   = uistate.PaneScratchpads
	PaneOps           = uistate.PaneOps
	PaneExportReports = uistate.PaneExportReports
	PaneConfig        = uistate.PaneConfig
	PaneSettings      = uistate.PaneSettings
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
	eventStop chan struct{}

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
	repos                 []api.Repo
	streams               []api.Stream
	issues                []api.Issue // context-filtered (by active streamId)
	habits                []api.Habit
	allIssues             []api.IssueWithMeta
	dueHabits             []api.HabitDailyItem
	dailySummary          *api.DailyIssueSummary
	dailyPlan             *api.DailyPlan
	dashboardDate         string
	wellbeingDate         string
	dailyCheckIn          *api.DailyCheckIn
	metricsRange          []api.DailyMetricsDay
	metricsRollup         *api.MetricsRollup
	streaks               *api.StreakSummary
	exportAssets          *api.ExportAssetStatus
	exportReports         []api.ExportReportFile
	issueSessions         []api.Session
	sessionHistory        []api.SessionHistoryEntry
	sessionDetail         *api.SessionDetail
	scratchpads           []api.ScratchPad
	stashes               []api.Stash
	ops                   []api.Op
	context               *api.ActiveContext
	timer                 *api.TimerState
	health                *api.Health
	updateStatus          *api.UpdateStatus
	updateChecking        bool
	updateInstalling      bool
	updateInstallPhase    string
	updateInstallDetail   string
	updateInstallOutput   string
	updateInstallError    string
	currentExecutablePath string
	settings              *api.CoreSettings
	kernelInfo            *api.KernelInfo
	elapsed               int // local seconds since last timer.state event
	timerTickSeq          int

	// terminal dimensions
	width  int
	height int

	// scratchpad reader state within the scratchpads pane
	scratchpadOpen     bool
	scratchpadMeta     *api.ScratchPad
	scratchpadFilePath string // resolved absolute path for $EDITOR
	scratchpadRendered string // glamour-rendered content
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
	dialogViewMeta           string
	dialogViewBody           string
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
	close(m.eventStop)
	m.eventStop = nil
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
		eventStop:           done,
		view:                ViewDaily,
		pane:                PaneIssues,
		defaultIssueSection: DefaultIssueSectionOpen,
		cursor: map[Pane]int{
			PaneRepos:         0,
			PaneStreams:       0,
			PaneIssues:        0,
			PaneHabits:        0,
			PaneSessions:      0,
			PaneScratchpads:   0,
			PaneOps:           0,
			PaneExportReports: 0,
			PaneConfig:        0,
			PaneSettings:      0,
		},
		filters: map[Pane]string{
			PaneRepos:         "",
			PaneStreams:       "",
			PaneIssues:        "",
			PaneHabits:        "",
			PaneSessions:      "",
			PaneScratchpads:   "",
			PaneOps:           "",
			PaneExportReports: "",
			PaneConfig:        "",
			PaneSettings:      "",
		},
		currentExecutablePath: executablePath,
		kernelInfo:            &api.KernelInfo{Env: env},
	}
	return model
}

// eventChannel receives kernel events forwarded from main.go.
var eventChannel <-chan api.KernelEvent

// ---------- Init ----------

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		commands.LoadRepos(m.client),
		commands.LoadAllIssues(m.client),
		commands.LoadDueHabits(m.client, time.Now().Format("2006-01-02")),
		commands.LoadDailySummary(m.client, ""),
		commands.LoadWellbeing(m.client, time.Now().Format("2006-01-02")),
		loadSessionHistoryForModel(m, 200),
		commands.LoadScratchpads(m.client),
		commands.LoadOps(m.client, m.currentOpsLimit()),
		commands.LoadContext(m.client),
		commands.LoadTimer(m.client),
		commands.LoadHealth(m.client),
		commands.LoadUpdateStatus(m.client),
		commands.LoadSettings(m.client),
		commands.LoadKernelInfo(m.client),
		commands.LoadExportAssets(m.client),
		commands.LoadExportReports(m.client),
		commands.HealthTickAfter(),
		commands.WaitForEvent(eventChannel),
	)
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
	return len(selectionpkg.FilteredIndices(m.selectionSnapshot(), p))
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
	return selectionpkg.Snapshot{
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
		ConfigItems:         configitems.Build(m.exportAssets),
	}
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
	if nextProtected, _, _ := views.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); nextProtected {
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
		DefaultIssueSection: m.defaultIssueSection,
		DashboardDate:       m.dashboardDate,
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
		UpdateInstallPhase:  m.updateInstallPhase,
		UpdateInstallDetail: m.updateInstallDetail,
		UpdateInstallOutput: m.updateInstallOutput,
		UpdateInstallError:  m.updateInstallError,
		CurrentExecutable:   m.currentExecutablePath,
		Settings:            m.settings,
		ExportAssets:        m.exportAssets,
		DailyCheckIn:        m.dailyCheckIn,
	}
}

func (m Model) applyInputState(state inputpkg.State) Model {
	m.view = state.ActiveView
	m.pane = state.ActivePane
	m.cursor = state.Cursor
	m.defaultIssueSection = state.DefaultIssueSection
	m.dashboardDate = state.DashboardDate
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
	m.updateInstallPhase = state.UpdateInstallPhase
	m.updateInstallDetail = state.UpdateInstallDetail
	m.updateInstallOutput = state.UpdateInstallOutput
	m.updateInstallError = state.UpdateInstallError
	m.currentExecutablePath = state.CurrentExecutable
	m.settings = state.Settings
	m.exportAssets = state.ExportAssets
	m.dailyCheckIn = state.DailyCheckIn
	return m
}

func (m Model) inputDeps() inputpkg.Deps {
	return inputpkg.Deps{
		CloseEventStop: func() { m.stopEventStream() },
		ShutdownKernel: func() tea.Cmd { return commands.ShutdownKernel(m.client) },
		SeedDevData:    func() tea.Cmd { return commands.SeedDevData(m.client) },
		ClearDevData:   func() tea.Cmd { return commands.ClearDevData(m.client) },
		IsDevMode:      func(state inputpkg.State) bool { return m.applyInputState(state).isDevMode() },
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
		LoadWellbeing:        func(date string) tea.Cmd { return commands.LoadWellbeing(m.client, date) },
		CurrentWellbeingDate: func(state inputpkg.State) string { return m.applyInputState(state).currentWellbeingDate() },
		ConfigChangeSelected: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			if item, ok := selectionpkg.SelectedConfigItem(next.selectionSnapshot()); ok && next.exportAssets != nil {
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
			return commands.InstallUpdate(next.updateStatus, next.currentExecutablePath, next.selfUpdateInstallAvailable(), next.selfUpdateUnsupportedReason())
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
			if issue, ok := selectionpkg.SelectedIssueDetail(next.selectionSnapshot()); ok {
				next = next.withDialogState(dialogstate.OpenIssueStatus(next.dialogSnapshot(), issue.Status))
				*state = next.inputState()
				return true
			}
			return false
		},
		AbandonSelectedIssue: func(state *inputpkg.State) tea.Cmd {
			next := m.applyInputState(*state)
			issue, ok := selectionpkg.SelectedIssueDetail(next.selectionSnapshot())
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
			if issue, ok := selectionpkg.SelectedIssueDetail(next.selectionSnapshot()); ok {
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
			if issue, ok := selectionpkg.SelectedIssueDetail(next.selectionSnapshot()); ok {
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
			if issue, ok := selectionpkg.SelectedIssueDetail(next.selectionSnapshot()); ok {
				issueLabel := ""
				var estimateMinutes *int
				if meta := selectionpkg.IssueMetaByID(next.selectionSnapshot(), issue.ID); meta != nil {
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
			if selectionpkg.ActiveIssue(next.selectionSnapshot()) == nil {
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
		OpenEditRestProtectionDialog: func(state *inputpkg.State) bool {
			next := m.applyInputState(*state)
			next = next.openEditRestProtectionDialog()
			*state = next.inputState()
			return true
		},
	}
}

func (m Model) dialogSnapshot() dialogstate.Snapshot {
	snapshot := dialogstate.Snapshot{
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
		CurrentDashboardDate: m.currentDashboardDate(),
		CurrentWellbeingDate: m.currentWellbeingDate(),
	}
	if issue, ok := selectionpkg.SelectedIssueDetail(m.selectionSnapshot()); ok {
		snapshot.SelectedIssueID = issue.ID
		snapshot.SelectedStreamID = issue.StreamID
		snapshot.HasSelectedIssue = true
	}
	if issue := selectionpkg.ActiveIssue(m.selectionSnapshot()); issue != nil {
		snapshot.ActiveIssueStream = issue.StreamID
		snapshot.HasActiveIssue = true
	}
	return snapshot
}

func (m Model) dialogActionCmd(action dialogpkg.Action) tea.Cmd {
	return dialogruntime.Resolve(action, m.dialogRuntimeState(), m.dialogRuntimeDeps())
}

func (m Model) openCreateScratchpad() Model {
	return m.withDialogState(dialogstate.OpenCreateScratchpad(m.dialogSnapshot()))
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
func (m Model) openEditRestProtectionDialog() Model {
	return m.withDialogState(dialogstate.OpenEditRestProtection(m.dialogSnapshot()))
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
		ViewMeta:           m.dialogViewMeta,
		ViewBody:           m.dialogViewBody,
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
	m.dialogViewMeta = state.ViewMeta
	m.dialogViewBody = state.ViewBody
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
		Context:       m.context,
		Repos:         m.repos,
		DashboardDate: m.currentDashboardDate(),
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
