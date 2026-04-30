package input

import (
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogpkg "crona/tui/internal/tui/dialogs"
	keyregistry "crona/tui/internal/tui/key_registry"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	ActiveView          uistate.View
	ActivePane          uistate.Pane
	ProtectedModeActive bool
	Cursor              map[uistate.Pane]int
	Filters             map[uistate.Pane]string
	DefaultIssueSection uistate.DefaultIssueSection
	DashboardDate       string
	RollupStartDate     string
	RollupEndDate       string
	WellbeingDate       string
	Dialog              string
	DialogState         dialogpkg.State
	HelpOpen            bool
	SessionDetailOpen   bool
	SessionDetailY      int
	SessionContextOpen  bool
	SessionContextY     int
	ScratchpadOpen      bool
	OpsLimit            int
	OpsLimitPinned      bool
	Context             *api.ActiveContext
	Timer               *api.TimerState
	UpdateStatus        *api.UpdateStatus
	UpdateChecking      bool
	UpdateInstalling    bool
	UpdateInstallError  string
	CurrentExecutable   string
	RunningIsBeta       bool
	Settings            *api.CoreSettings
	AlertStatus         *api.AlertStatus
	AlertReminders      []api.AlertReminder
	ExportAssets        *api.ExportAssetStatus
	DailyCheckIn        *api.DailyCheckIn
}

func (s State) Init() tea.Cmd                           { return nil }
func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s State) View() string                            { return "" }

type Deps struct {
	CloseEventStop                  func()
	ShutdownKernel                  func() tea.Cmd
	SeedDevData                     func() tea.Cmd
	ClearDevData                    func() tea.Cmd
	PrepareLocalUpdate              func() tea.Cmd
	IsDevMode                       func(State) bool
	NextActiveSessionView           func(State, int) uistate.View
	NextWorkspaceView               func(State, int) uistate.View
	DefaultPane                     func(uistate.View) uistate.Pane
	NextPane                        func(uistate.View, uistate.Pane, int) uistate.Pane
	SetDefaultIssueSection          func(*State, uistate.DefaultIssueSection)
	ListLen                         func(State, uistate.Pane) int
	LoadExportAssets                func() tea.Cmd
	SetStatus                       func(*State, string, bool) tea.Cmd
	LoadDailySummary                func(string) tea.Cmd
	LoadDueHabits                   func(string) tea.Cmd
	CurrentDashboardDate            func(State) string
	LoadRollupSummaries             func(string, string) tea.Cmd
	CurrentRollupStartDate          func(State) string
	CurrentRollupEndDate            func(State) string
	LoadWellbeing                   func(string) tea.Cmd
	CurrentWellbeingDate            func(State) string
	ConfigChangeSelected            func(*State) tea.Cmd
	OpenCheckoutContextDialog       func(*State) bool
	Checkout                        func(*State) tea.Cmd
	CheckUpdateNow                  func() tea.Cmd
	SelfUpdateInstallAvailable      func(State) bool
	SelfUpdateUnsupportedReason     func(State) string
	InstallUpdate                   func(State) tea.Cmd
	DismissUpdate                   func() tea.Cmd
	ResumeSession                   func() tea.Cmd
	PauseSession                    func() tea.Cmd
	OpenEndSessionDialog            func(*State) bool
	OpenStashSessionDialog          func(*State) bool
	CanOpenStashList                func(State) bool
	OpenStashListDialog             func(*State) bool
	LoadStashes                     func() tea.Cmd
	ClampFiltered                   func(*State, uistate.Pane)
	CurrentOpsLimit                 func(State) int
	LoadOps                         func(int) tea.Cmd
	StartFilterEdit                 func(*State, uistate.Pane)
	OpenIssueStatusFromSelection    func(*State) bool
	AbandonSelectedIssue            func(*State) tea.Cmd
	ToggleSelectedIssueToday        func(*State) tea.Cmd
	OpenSelectedIssueTodoDateDialog func(*State) bool
	HandleCreateAction              func(*State) bool
	OpenExportDailyDialog           func(*State) bool
	OpenEditorAction                func(*State) (tea.Cmd, bool)
	DeleteSelectionAction           func(*State) (tea.Cmd, bool)
	OpenSelectionAction             func(*State) (tea.Cmd, bool)
	EnterAction                     func(*State) (tea.Cmd, bool)
	ToggleHabitCompletedAction      func(*State) (tea.Cmd, bool)
	SetHabitFailedAction            func(*State) (tea.Cmd, bool)
	StartFocusFromSelection         func(*State) tea.Cmd
	OpenManualSessionDialog         func(*State) bool
	OpenSessionContextOverlay       func(*State) bool
	ConfigReset                     func(*State) tea.Cmd
	PatchSetting                    func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd
	TestAlertNotification           func() tea.Cmd
	TestAlertSound                  func() tea.Cmd
	CreateAlertReminder             func(shareddto.AlertReminderCreateRequest) tea.Cmd
	UpdateAlertReminder             func(shareddto.AlertReminderUpdateRequest) tea.Cmd
	ToggleAlertReminder             func(string, bool) tea.Cmd
	DeleteAlertReminder             func(string) tea.Cmd
	OpenCreateAlertReminderDialog   func(*State) bool
	OpenEditAlertReminderDialog     func(*State, string) bool
	OpenEditDateDisplayFormatDialog func(*State) bool
	OpenEditRestProtectionDialog    func(*State) bool
	OpenConfirmWipeDataDialog       func(*State) bool
	OpenConfirmUninstallDialog      func(*State) bool
	WipeRuntimeData                 func() tea.Cmd
	OpenRollupStartDateDialog       func(*State) bool
	OpenRollupEndDateDialog         func(*State) bool
	OpenSupportIssueURL             func() tea.Cmd
	OpenSupportDiscussionsURL       func() tea.Cmd
	OpenSupportReleasesURL          func() tea.Cmd
	OpenSupportRoadmapURL           func() tea.Cmd
	CopySupportDiagnostics          func(State) tea.Cmd
	GenerateSupportBundle           func(State) tea.Cmd
	OpenViewJumpDialog              func(*State) bool
	OpenBetaSupportDialog           func(*State) bool
}

type handler = keyregistry.Handler[State]
type router = keyregistry.Router[State, uistate.View, uistate.Pane]

func Handle(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	next, cmd := newRouter(deps).Handle(state, state.ActiveView, state.ActivePane, key)
	return next.(State), cmd
}

func newRouter(deps Deps) *router {
	r := keyregistry.New[State, uistate.View, uistate.Pane]()
	keyregistry.RegisterGlobal(r, map[string]handler{
		"q": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			deps.CloseEventStop()
			return s, tea.Quit, true
		},
		"ctrl+c": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			deps.CloseEventStop()
			return s, tea.Quit, true
		},
		"K": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return s, deps.ShutdownKernel(), true },
		"f6": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleDevCmd(s, deps.IsDevMode, deps.SeedDevData)
		},
		"f7": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleDevCmd(s, deps.IsDevMode, deps.ClearDevData)
		},
		"f8": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if !deps.IsDevMode(s) {
				return s, nil, false
			}
			s.ActiveView = uistate.ViewUpdates
			s.ActivePane = uistate.DefaultPane(s.ActiveView)
			s.UpdateChecking = true
			s.UpdateInstallError = ""
			return s, deps.PrepareLocalUpdate(), true
		},
		"f9": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if !s.RunningIsBeta {
				return s, nil, false
			}
			deps.OpenBetaSupportDialog(&s)
			return s, nil, true
		},
		"]":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, 1) },
		"[":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, -1) },
		"tab":       func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, 1) },
		"shift+tab": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, -1) },
		"v":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleOpenViewJump(s, deps) },
		"u":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleOpenUpdates(s) },
		"R":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleRescanExportAssets(s, deps) },
		"j":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, 1) },
		"down":      func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, 1) },
		"k":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, -1) },
		"up":        func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, -1) },
		"f": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return s, deps.StartFocusFromSelection(&s), true
		},
		"m": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			if deps.OpenManualSessionDialog(&s) {
				return s, nil, true
			}
			return s, nil, false
		},
		"s": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return handleIssueStatus(s, deps)
		},
		"A": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return s, deps.AbandonSelectedIssue(&s), true
		},
		"y": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return s, deps.ToggleSelectedIssueToday(&s), true
		},
		"D": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			deps.OpenSelectedIssueTodoDateDialog(&s)
			return s, nil, true
		},
		"c": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return handleContextCheckout(s, deps)
		},
		"Z": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return handleOpenStashList(s, deps)
		},
		"+": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		"=": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		"-": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, -10) },
		"/": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleStartFilter(s, deps) },
		"?": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { s.HelpOpen = true; return s, nil, true },
		"a": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			deps.HandleCreateAction(&s)
			return s, nil, true
		},
		"e": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		"d": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ActiveView == uistate.ViewAlerts {
				return handleDeleteSelectedAlertReminder(s, deps)
			}
			cmd, handled := deps.DeleteSelectionAction(&s)
			return s, cmd, handled
		},
		"x": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ActiveView == uistate.ViewAlerts {
				return handleDeleteSelectedAlertReminder(s, deps)
			}
			return s, nil, false
		},
		"o": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenSelectionAction(&s)
			return s, cmd, handled
		},
		"enter": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.EnterAction(&s)
			return s, cmd, handled
		},
		" ": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleSpace(s, deps) },
	})
	keyregistry.RegisterDefault(r, uistate.ViewDefault, uistate.PaneIssues,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneIssues
			deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionOpen)
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneIssues
			deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionCompleted)
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			deps.OpenCheckoutContextDialog(&s)
			return s, nil, true
		},
	)
	keyregistry.RegisterDaily(r, uistate.ViewDaily, uistate.PaneIssues, uistate.PaneHabits,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneIssues
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneHabits
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDailyDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDailyDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetDailyDate(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleOpenExportDaily(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			deps.OpenCheckoutContextDialog(&s)
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			if deps.OpenManualSessionDialog(&s) {
				return s, nil, true
			}
			return s, nil, false
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.SetHabitFailedAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.ToggleHabitCompletedAction(&s)
			return s, cmd, handled
		},
	)
	r.RegisterView(uistate.ViewDaily, "w", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleToggleAwayMode(s, deps)
	})
	r.RegisterView(uistate.ViewAway, "w", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleToggleAwayMode(s, deps)
	})
	keyregistry.RegisterMeta(r, uistate.ViewMeta, uistate.PaneRepos, uistate.PaneStreams, uistate.PaneIssues, uistate.PaneHabits,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneRepos
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneStreams
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneIssues
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneHabits
			return s, nil, true
		},
	)
	keyregistry.RegisterSettings(r, uistate.ViewSettings,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneSettings
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleAdjustSelectedSetting(s, deps, -1)
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustSelectedSetting(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedSetting(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedSetting(s, deps) },
	)
	keyregistry.RegisterSettings(r, uistate.ViewAlerts,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneAlerts
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleAdjustSelectedAlert(s, deps, -1)
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustSelectedAlert(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedAlert(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedAlert(s, deps) },
	)
	keyregistry.RegisterConfig(r, uistate.ViewConfig, uistate.PaneConfig,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneConfig
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleRescanExportAssets(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return s, deps.ConfigChangeSelected(&s), true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return s, deps.ConfigReset(&s), true },
	)
	keyregistry.RegisterReports(r, uistate.ViewReports, uistate.PaneExportReports,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneExportReports
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenSelectionAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.EnterAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.DeleteSelectionAction(&s)
			return s, cmd, handled
		},
	)
	keyregistry.RegisterSession(r, uistate.ViewSessionActive, uistate.ViewSessionHistory,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handlePauseSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResumeSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleEndSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleStashSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if deps.OpenSessionContextOverlay(&s) {
				return s, nil, true
			}
			return s, nil, false
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.EnterAction(&s)
			return s, cmd, handled
		},
	)
	keyregistry.RegisterScratch(r, uistate.ViewScratch,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.ActivePane = uistate.PaneScratchpads
			return s, nil, true
		},
	)
	keyregistry.RegisterUpdates(r, uistate.ViewUpdates,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.UpdateChecking = true
			s.UpdateInstallError = ""
			return s, deps.CheckUpdateNow(), true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleInstallUpdate(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleDismissUpdate(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenSelectionAction(&s)
			return s, cmd, handled
		},
	)
	r.RegisterView(uistate.ViewSupport, "o", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.OpenSupportIssueURL(), true
	})
	r.RegisterView(uistate.ViewSupport, "d", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.OpenSupportDiscussionsURL(), true
	})
	r.RegisterView(uistate.ViewSupport, "r", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.OpenSupportReleasesURL(), true
	})
	r.RegisterView(uistate.ViewSupport, "g", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.OpenSupportRoadmapURL(), true
	})
	r.RegisterView(uistate.ViewSupport, "c", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.CopySupportDiagnostics(s), true
	})
	r.RegisterView(uistate.ViewSupport, "b", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return s, deps.GenerateSupportBundle(s), true
	})
	keyregistry.RegisterWellbeing(r, uistate.ViewWellbeing,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftWellbeingDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftWellbeingDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetWellbeingDate(s, deps) },
	)
	keyregistry.RegisterRollup(r, uistate.ViewRollup,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftRollupStartDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftRollupStartDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftRollupEndDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftRollupEndDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetRollupRange(s, deps) },
	)
	r.RegisterView(uistate.ViewRollup, "S", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenRollupStartDateDialog(&s) {
			return s, nil, true
		}
		return s, nil, false
	})
	r.RegisterView(uistate.ViewRollup, "E", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenRollupEndDateDialog(&s) {
			return s, nil, true
		}
		return s, nil, false
	})
	r.RegisterView(uistate.ViewWellbeing, "w", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleToggleAwayMode(s, deps)
	})
	keyregistry.RegisterOps(r, uistate.ViewOps, uistate.PaneOps,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, -10) },
	)
	return r
}
