package input

import (
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogpkg "crona/tui/internal/tui/dialogs"
	keyregistry "crona/tui/internal/tui/key_registry"
	navigationutil "crona/tui/internal/tui/navigationutil"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

type State struct {
	ActiveView          uistate.View
	ActivePane          uistate.Pane
	ProtectedModeActive bool
	Cursor              map[uistate.Pane]int
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
	UpdateInstallPhase  string
	UpdateInstallDetail string
	UpdateInstallOutput string
	UpdateInstallError  string
	CurrentExecutable   string
	Settings            *api.CoreSettings
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
	OpenEditRestProtectionDialog    func(*State) bool
	OpenRollupStartDateDialog       func(*State) bool
	OpenRollupEndDateDialog         func(*State) bool
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
		"]":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, 1) },
		"[":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, -1) },
		"tab":       func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, 1) },
		"shift+tab": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, -1) },
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
			cmd, handled := deps.DeleteSelectionAction(&s)
			return s, cmd, handled
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

func handleDevCmd(s State, isDev func(State) bool, cmd func() tea.Cmd) (tea.Model, tea.Cmd, bool) {
	if !isDev(s) {
		return s, nil, false
	}
	return s, cmd(), true
}

func handleCycleView(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.Timer != nil && s.Timer.State != "idle" {
		s.ActiveView = deps.NextActiveSessionView(s, dir)
	} else {
		s.ActiveView = deps.NextWorkspaceView(s, dir)
	}
	s.ActivePane = deps.DefaultPane(s.ActiveView)
	return s, nil, true
}

func handleCyclePane(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView == uistate.ViewDefault && s.ActivePane == uistate.PaneIssues {
		if dir >= 0 {
			if s.DefaultIssueSection == uistate.DefaultIssueSectionOpen {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionCompleted)
			} else {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionOpen)
			}
		} else {
			if s.DefaultIssueSection == uistate.DefaultIssueSectionCompleted {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionOpen)
			} else {
				deps.SetDefaultIssueSection(&s, uistate.DefaultIssueSectionCompleted)
			}
		}
		return s, nil, true
	}
	s.ActivePane = deps.NextPane(s.ActiveView, s.ActivePane, dir)
	return s, nil, true
}

func handleOpenUpdates(s State) (tea.Model, tea.Cmd, bool) {
	s.ActiveView = uistate.ViewUpdates
	s.ActivePane = uistate.DefaultPane(s.ActiveView)
	return s, nil, true
}

func handleShiftRollupStartDate(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	start := shiftInputISODate(deps.CurrentRollupStartDate(s), delta)
	end := deps.CurrentRollupEndDate(s)
	if start > end {
		end = start
	}
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func handleShiftRollupEndDate(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	end := shiftInputISODate(deps.CurrentRollupEndDate(s), delta)
	start := deps.CurrentRollupStartDate(s)
	if start > end {
		start = end
	}
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func handleResetRollupRange(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	end := time.Now().Format("2006-01-02")
	start := shiftInputISODate(end, -6)
	s.RollupStartDate = start
	s.RollupEndDate = end
	return s, deps.LoadRollupSummaries(start, end), true
}

func shiftInputISODate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func handleRescanExportAssets(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewConfig {
		return s, nil, true
	}
	return s, tea.Batch(deps.SetStatus(&s, "Rescanning export tools...", false), deps.LoadExportAssets()), true
}

func handleCursor(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	max := deps.ListLen(s, s.ActivePane)
	if max == 0 {
		return s, nil, true
	}
	next := s.Cursor[s.ActivePane] + delta
	if next < 0 {
		next = 0
	}
	if next >= max {
		next = max - 1
	}
	s.Cursor[s.ActivePane] = next
	return s, nil, true
}

func handleIssueStatus(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.Timer != nil && s.Timer.State != "idle" && (s.ActiveView == uistate.ViewSessionActive || s.ActiveView == uistate.ViewScratch) {
		return s, deps.SetStatus(&s, "End or stash the active session before changing issue status", true), true
	}
	deps.OpenIssueStatusFromSelection(&s)
	return s, nil, true
}

func handleContextCheckout(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if (s.ActiveView == uistate.ViewDefault && s.ActivePane == uistate.PaneIssues) || s.ActiveView == uistate.ViewDaily {
		deps.OpenCheckoutContextDialog(&s)
		return s, nil, true
	}
	return s, deps.Checkout(&s), true
}

func handleOpenStashList(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if !deps.CanOpenStashList(s) {
		return s, nil, false
	}
	deps.OpenStashListDialog(&s)
	return s, deps.LoadStashes(), true
}

func handleAdjustOpsLimit(s State, deps Deps, delta int) (tea.Model, tea.Cmd, bool) {
	if s.ActivePane != uistate.PaneOps {
		return s, nil, false
	}
	s.OpsLimitPinned = true
	s.OpsLimit += delta
	if s.OpsLimit < 10 {
		s.OpsLimit = 10
	}
	if delta < 0 {
		deps.ClampFiltered(&s, uistate.PaneOps)
	}
	return s, deps.LoadOps(deps.CurrentOpsLimit(s)), true
}

func handleStartFilter(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	switch s.ActivePane {
	case uistate.PaneOps, uistate.PaneIssues, uistate.PaneHabits, uistate.PaneRepos, uistate.PaneStreams, uistate.PaneScratchpads, uistate.PaneSessions, uistate.PaneConfig, uistate.PaneExportReports:
	default:
		return s, nil, false
	}
	if s.ActivePane == uistate.PaneScratchpads && s.ScratchpadOpen {
		return s, nil, true
	}
	deps.StartFilterEdit(&s, s.ActivePane)
	return s, nil, true
}

func handleSpace(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView == uistate.ViewSettings {
		return handleActivateSelectedSetting(s, deps)
	}
	cmd, handled := deps.ToggleHabitCompletedAction(&s)
	return s, cmd, handled
}

func handleShiftDailyDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.DashboardDate = navigationutil.ShiftISODate(deps.CurrentDashboardDate(s), dir)
	return s, tea.Batch(deps.LoadDailySummary(s.DashboardDate), deps.LoadDueHabits(s.DashboardDate)), true
}

func handleResetDailyDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.DashboardDate = ""
	return s, tea.Batch(deps.LoadDailySummary(""), deps.LoadDueHabits(deps.CurrentDashboardDate(s))), true
}

func handleOpenExportDaily(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.Dialog != "" {
		return s, nil, false
	}
	deps.OpenExportDailyDialog(&s)
	return s, nil, true
}

func handleShiftWellbeingDate(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = navigationutil.ShiftISODate(deps.CurrentWellbeingDate(s), dir)
	return s, deps.LoadWellbeing(s.WellbeingDate), true
}

func handleResetWellbeingDate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	s.WellbeingDate = ""
	return s, deps.LoadWellbeing(deps.CurrentWellbeingDate(s)), true
}

func handleInstallUpdate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewUpdates || !viewsShouldShowUpdate(s.UpdateStatus) || s.UpdateInstalling {
		return s, nil, false
	}
	if !deps.SelfUpdateInstallAvailable(s) {
		reason := strings.TrimSpace(deps.SelfUpdateUnsupportedReason(s))
		if reason == "" && s.UpdateStatus != nil {
			reason = strings.TrimSpace(s.UpdateStatus.InstallUnavailableReason)
		}
		if reason == "" {
			reason = "Please update manually."
		}
		return s, deps.SetStatus(&s, reason, true), true
	}
	s.UpdateInstalling = true
	s.UpdateInstallError = ""
	s.UpdateInstallOutput = ""
	return s, deps.InstallUpdate(s), true
}

func handleDismissUpdate(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if !viewsShouldShowUpdate(s.UpdateStatus) {
		return s, nil, false
	}
	return s, deps.DismissUpdate(), true
}

func handlePauseSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State != "running" {
		return s, nil, false
	}
	return s, deps.PauseSession(), true
}

func handleResumeSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State != "paused" {
		return s, nil, false
	}
	return s, deps.ResumeSession(), true
}

func handleEndSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State == "idle" {
		return s, nil, false
	}
	deps.OpenEndSessionDialog(&s)
	return s, nil, true
}

func handleStashSession(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSessionActive || s.Timer == nil || s.Timer.State == "idle" {
		return s, nil, false
	}
	deps.OpenStashSessionDialog(&s)
	return s, nil, true
}

func handleAdjustSelectedSetting(s State, deps Deps, dir int) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSettings || s.Settings == nil {
		return s, nil, true
	}
	repoID, streamID := int64(0), int64(0)
	if s.Context != nil && s.Context.RepoID != nil {
		repoID = *s.Context.RepoID
	}
	if s.Context != nil && s.Context.StreamID != nil {
		streamID = *s.Context.StreamID
	}
	rawIdx := s.Cursor[uistate.PaneSettings]
	switch rawIdx {
	case 0:
		next := sharedtypes.TimerModeStructured
		if s.Settings.TimerMode == sharedtypes.TimerModeStructured {
			next = sharedtypes.TimerModeStopwatch
		}
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyTimerMode, next, repoID, streamID, s.DashboardDate), true
	case 1:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyWorkDurationMinutes, clampMin(s.Settings.WorkDurationMinutes+dir*5, 5), repoID, streamID, s.DashboardDate), true
	case 2:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBreaksEnabled, !s.Settings.BreaksEnabled, repoID, streamID, s.DashboardDate), true
	case 3:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyShortBreakMinutes, clampMin(s.Settings.ShortBreakMinutes+dir, 1), repoID, streamID, s.DashboardDate), true
	case 4:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyLongBreakMinutes, clampMin(s.Settings.LongBreakMinutes+dir*5, 5), repoID, streamID, s.DashboardDate), true
	case 5:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyLongBreakEnabled, !s.Settings.LongBreakEnabled, repoID, streamID, s.DashboardDate), true
	case 6:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak, clampMin(s.Settings.CyclesBeforeLongBreak+dir, 1), repoID, streamID, s.DashboardDate), true
	case 7:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAutoStartBreaks, !s.Settings.AutoStartBreaks, repoID, streamID, s.DashboardDate), true
	case 8:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAutoStartWork, !s.Settings.AutoStartWork, repoID, streamID, s.DashboardDate), true
	case 9:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundaryNotifications, !s.Settings.BoundaryNotifications, repoID, streamID, s.DashboardDate), true
	case 10:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyBoundarySound, !s.Settings.BoundarySound, repoID, streamID, s.DashboardDate), true
	case 11:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdateChecksEnabled, !s.Settings.UpdateChecksEnabled, repoID, streamID, s.DashboardDate), true
	case 12:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyUpdatePromptEnabled, !s.Settings.UpdatePromptEnabled, repoID, streamID, s.DashboardDate), true
	case 13:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyRepoSort, nextRepoSort(s.Settings.RepoSort, dir), repoID, streamID, s.DashboardDate), true
	case 14:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyStreamSort, nextStreamSort(s.Settings.StreamSort, dir), repoID, streamID, s.DashboardDate), true
	case 15:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyIssueSort, nextIssueSort(s.Settings.IssueSort, dir), repoID, streamID, s.DashboardDate), true
	case 16:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyHabitSort, nextHabitSort(s.Settings.HabitSort, dir), repoID, streamID, s.DashboardDate), true
	case 17:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyAwayModeEnabled, !s.Settings.AwayModeEnabled, repoID, streamID, s.DashboardDate), true
	case 18:
		return s, deps.PatchSetting(sharedtypes.CoreSettingsKeyDailyPlanRollbackMins, clampMin(currentRollbackMinutes(s.Settings.DailyPlanRollbackMins)+dir, 1), repoID, streamID, s.DashboardDate), true
	default:
		return s, nil, true
	}
}

func handleActivateSelectedSetting(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if s.ActiveView != uistate.ViewSettings || s.Settings == nil {
		return s, nil, true
	}
	switch s.Cursor[uistate.PaneSettings] {
	case 19:
		deps.OpenEditRestProtectionDialog(&s)
		return s, nil, true
	default:
		return handleAdjustSelectedSetting(s, deps, 1)
	}
}

func handleToggleAwayMode(s State, deps Deps) (tea.Model, tea.Cmd, bool) {
	if (s.ActiveView != uistate.ViewDaily && s.ActiveView != uistate.ViewWellbeing && s.ActiveView != uistate.ViewAway) || s.Settings == nil {
		return s, nil, false
	}
	repoID := int64(0)
	if s.Context != nil && s.Context.RepoID != nil {
		repoID = *s.Context.RepoID
	}
	streamID := int64(0)
	if s.Context != nil && s.Context.StreamID != nil {
		streamID = *s.Context.StreamID
	}
	next := !s.Settings.AwayModeEnabled
	status := "Away mode disabled"
	if next {
		status = "Away mode enabled"
	}
	settingsCopy := *s.Settings
	settingsCopy.AwayModeEnabled = next
	s.Settings = &settingsCopy
	if s.ActiveView == uistate.ViewAway && !next {
		s.ActiveView = uistate.ViewDaily
		s.ActivePane = uistate.DefaultPane(s.ActiveView)
	}
	date := s.DashboardDate
	if s.ActiveView == uistate.ViewWellbeing && strings.TrimSpace(s.WellbeingDate) != "" {
		date = s.WellbeingDate
	}
	return s, tea.Batch(
		deps.PatchSetting(sharedtypes.CoreSettingsKeyAwayModeEnabled, next, repoID, streamID, date),
		deps.SetStatus(&s, status, false),
	), true
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

func clampMin(value, min int) int {
	if value < min {
		return min
	}
	return value
}

func currentRollbackMinutes(value int) int {
	if value <= 0 {
		return 5
	}
	return value
}

func nextRepoSort(current sharedtypes.RepoSort, dir int) sharedtypes.RepoSort {
	options := []sharedtypes.RepoSort{
		sharedtypes.RepoSortAlphabeticalAsc,
		sharedtypes.RepoSortAlphabeticalDesc,
		sharedtypes.RepoSortChronologicalAsc,
		sharedtypes.RepoSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextStreamSort(current sharedtypes.StreamSort, dir int) sharedtypes.StreamSort {
	options := []sharedtypes.StreamSort{
		sharedtypes.StreamSortAlphabeticalAsc,
		sharedtypes.StreamSortAlphabeticalDesc,
		sharedtypes.StreamSortChronologicalAsc,
		sharedtypes.StreamSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextIssueSort(current sharedtypes.IssueSort, dir int) sharedtypes.IssueSort {
	options := []sharedtypes.IssueSort{
		sharedtypes.IssueSortPriority,
		sharedtypes.IssueSortDueDateAsc,
		sharedtypes.IssueSortDueDateDesc,
		sharedtypes.IssueSortAlphabeticalAsc,
		sharedtypes.IssueSortAlphabeticalDesc,
		sharedtypes.IssueSortChronologicalAsc,
		sharedtypes.IssueSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextHabitSort(current sharedtypes.HabitSort, dir int) sharedtypes.HabitSort {
	options := []sharedtypes.HabitSort{
		sharedtypes.HabitSortSchedule,
		sharedtypes.HabitSortTargetMinutesAsc,
		sharedtypes.HabitSortTargetMinutesDesc,
		sharedtypes.HabitSortAlphabeticalAsc,
		sharedtypes.HabitSortAlphabeticalDesc,
		sharedtypes.HabitSortChronologicalAsc,
		sharedtypes.HabitSortChronologicalDesc,
	}
	return options[nextIndex(current, options, dir)]
}

func nextIndex[T comparable](current T, options []T, dir int) int {
	if len(options) == 0 {
		return 0
	}
	index := 0
	for i, option := range options {
		if option == current {
			index = i
			break
		}
	}
	index += dir
	if index < 0 {
		index = len(options) - 1
	}
	if index >= len(options) {
		index = 0
	}
	return index
}
