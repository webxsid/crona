package input

import (
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
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
	DailyTaskSection    uistate.DailyTaskSection
	// DailyDetailDay shows the Day summary in Daily's detail pane instead of the selection.
	DailyDetailDay bool
	// Date is the one date every dated view shows; "" means the logical today.
	Date             string
	ReviewWindowDays int
	// ReviewWindowEnd anchors Review's window; ReviewShowTime puts Time in Trends' slot.
	ReviewWindowEnd           string
	ReviewShowTime            bool
	MomentumWindowDays        int
	Dialog                    string
	DialogState               dialogstate.State
	HelpOpen                  bool
	SessionDetailOpen         bool
	SessionDetailY            int
	SessionContextOpen        bool
	SessionContextY           int
	OpsLimit                  int
	OpsLimitPinned            bool
	Context                   *api.ActiveContext
	Timer                     *api.TimerState
	UpdateStatus              *api.UpdateStatus
	UpdateChecking            bool
	UpdateDiagnosticsExpanded bool
	SupportCursor             int
	CurrentExecutable         string
	RunningIsBeta             bool
	Settings                  *api.CoreSettings
	AlertStatus               *api.AlertStatus
	AlertReminders            []api.AlertReminder
	ExportAssets              *api.ExportAssetStatus
	DailyCheckIn              *api.DailyCheckIn
}

func (s State) Init() tea.Cmd                           { return nil }
func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s State) View() string                            { return "" }

type Deps struct {
	CloseEventStop           func()
	ShutdownKernel           func() tea.Cmd
	SeedDevData              func() tea.Cmd
	ClearDevData             func() tea.Cmd
	IsDevMode                func(State) bool
	TabTargetView            func(State, uistate.Tab) (uistate.View, bool)
	NextSubTabView           func(State, int) uistate.View
	DefaultPane              func(uistate.View) uistate.Pane
	NextPane                 func(uistate.View, uistate.Pane, int) uistate.Pane
	SetDefaultIssueSection   func(*State, uistate.DefaultIssueSection)
	SetDailyTaskSection      func(*State, uistate.DailyTaskSection)
	ListLen                  func(State, uistate.Pane) int
	LoadExportAssets         func() tea.Cmd
	SetStatus                func(*State, string, bool) tea.Cmd
	CurrentDate              func(State) string
	CurrentLogicalDate       func(State) string
	ReloadDateSurfaces       func(State) tea.Cmd
	LoadDailySummary         func(string) tea.Cmd
	LoadSummarySnapshot      func(string) tea.Cmd
	LoadDailyStreaks         func(string) tea.Cmd
	LoadDueHabits            func(string) tea.Cmd
	LoadReviewWindow         func(string, string) tea.Cmd
	LoadMomentumRange        func(string, int) tea.Cmd
	CreateMomentumDefinition func(api.HabitStreakDefinition, string, string, int) tea.Cmd
	UpdateMomentumDefinition func(api.HabitStreakDefinition, string, string, int) tea.Cmd
	DeleteMomentumDefinition func(string, string, string, int) tea.Cmd
	OpenCheckInDialog        func(*State) bool
	OpenSummaryCalendar      func(*State) bool
	// OpenDaySummary opens the Day summary as a dialog when Daily has no detail pane, and
	// reports whether it did.
	OpenDaySummary func(*State) bool
	// OpenReviewDialog opens Review's "day" or "time" as a dialog on narrow screens, where
	// neither has a pane, and reports whether it did.
	OpenReviewDialog                func(*State, string) bool
	OpenHelpDialog                  func(*State) bool
	ConfigChangeSelected            func(*State) tea.Cmd
	OpenCheckoutContextDialog       func(*State) bool
	Checkout                        func(*State) tea.Cmd
	CheckUpdateNow                  func() tea.Cmd
	ResumeSession                   func(State) tea.Cmd
	AdvanceSession                  func() tea.Cmd
	PauseSession                    func() tea.Cmd
	OpenEndSessionDialog            func(*State) bool
	ClampFiltered                   func(*State, uistate.Pane)
	CurrentOpsLimit                 func(State) int
	LoadOps                         func(int) tea.Cmd
	OpenIssueStatusFromSelection    func(*State) bool
	AbandonSelectedIssue            func(*State) tea.Cmd
	ToggleSelectedIssueToday        func(*State) tea.Cmd
	ToggleSelectedIssuePinnedDaily  func(*State) tea.Cmd
	OpenSelectedIssueTodoDateDialog func(*State) bool
	HandleCreateAction              func(*State) bool
	OpenExportDailyDialog           func(*State) bool
	OpenEditorAction                func(*State) (tea.Cmd, bool)
	DeleteSelectionAction           func(*State) (tea.Cmd, bool)
	OpenSelectionAction             func(*State) (tea.Cmd, bool)
	EnterAction                     func(*State) (tea.Cmd, bool)
	ToggleMomentumAction            func(*State) (tea.Cmd, bool)
	OpenHabitLogAction              func(*State) (tea.Cmd, bool)
	EnterHabitHistoryView           func(*State) (tea.Cmd, bool)
	ToggleHabitCompletedAction      func(*State) (tea.Cmd, bool)
	SetHabitFailedAction            func(*State) (tea.Cmd, bool)
	StartFocusFromSelection         func(*State) tea.Cmd
	OpenAdhocTimerDialog            func(*State) bool
	OpenManualSessionDialog         func(*State) (tea.Cmd, bool)
	OpenSessionContextOverlay       func(*State) bool
	ConfigReset                     func(*State) tea.Cmd
	PatchSetting                    func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd
	SetAwayMode                     func(enabled bool, repoID, streamID int64, dashboardDate string) tea.Cmd
	TestAlertNotification           func() tea.Cmd
	TestAlertSound                  func() tea.Cmd
	CreateAlertReminder             func(shareddto.AlertReminderCreateRequest) tea.Cmd
	UpdateAlertReminder             func(shareddto.AlertReminderUpdateRequest) tea.Cmd
	ToggleAlertReminder             func(string, bool) tea.Cmd
	DeleteAlertReminder             func(string) tea.Cmd
	OpenCreateAlertReminderDialog   func(*State, sharedtypes.AlertReminderKind) bool
	OpenEditAlertReminderDialog     func(*State, string) bool
	OpenEditDateDisplayFormatDialog func(*State) bool
	OpenEditRestProtectionDialog    func(*State) bool
	OpenEditTelemetrySettingsDialog func(*State) bool
	OpenEditDayBoundaryDialog       func(*State, sharedtypes.CoreSettingsKey) bool
	OpenConfirmWipeDataDialog       func(*State) bool
	WipeRuntimeData                 func() tea.Cmd
	OpenDateDialog                  func(*State) bool
	OpenSupportFeedbackRoadmapURL   func() tea.Cmd
	OpenSupportReleasesURL          func() tea.Cmd
	OpenSupportDocumentationURL     func() tea.Cmd
	CopySupportDiagnostics          func(State) tea.Cmd
	GenerateSupportBundle           func(State) tea.Cmd
	OpenViewJumpDialog              func(*State) bool
	OpenBetaSupportDialog           func(*State) bool
}

type (
	handler = keyregistry.Handler[State]
	router  = keyregistry.Router[State, uistate.View, uistate.Pane]
)

func Handle(state State, key tea.KeyMsg, deps Deps) (State, tea.Cmd) {
	logger.Infof(
		"input.Handle: key=%q view=%s pane=%s dialog=%q help=%t session_detail=%t session_context=%t",
		key.String(),
		state.ActiveView,
		state.ActivePane,
		state.Dialog,
		state.HelpOpen,
		state.SessionDetailOpen,
		state.SessionContextOpen,
	)
	next, cmd := newRouter(deps).Handle(state, state.ActiveView, state.ActivePane, key)
	if routed, ok := next.(State); ok {
		logger.Infof(
			"input.Handle result: key=%q next_view=%s next_pane=%s dialog=%q",
			key.String(),
			routed.ActiveView,
			routed.ActivePane,
			routed.Dialog,
		)
	}
	return next.(State), cmd
}

func newRouter(deps Deps) *router {
	r := keyregistry.New[State, uistate.View, uistate.Pane]()
	keyregistry.RegisterGlobal(r, map[string]handler{
		"A": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleStartAdhocTimer(s, deps)
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
		"f9": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if !s.RunningIsBeta {
				return s, nil, false
			}
			deps.OpenBetaSupportDialog(&s)
			return s, nil, true
		},
		"1": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleJumpToTab(s, deps, uistate.TabToday)
		},
		"2": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleJumpToTab(s, deps, uistate.TabIssues)
		},
		"3": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleJumpToTab(s, deps, uistate.TabInsights)
		},
		"4": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleJumpToTab(s, deps, uistate.TabLog)
		},
		"5": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleJumpToTab(s, deps, uistate.TabSystem)
		},
		"]":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, 1) },
		"[":         func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCycleView(s, deps, -1) },
		"tab":       func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, 1) },
		"shift+tab": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCyclePane(s, deps, -1) },
		"left": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ActivePane == uistate.PaneScope {
				return handleCursor(s, deps, -1)
			}
			if (s.ActiveView == uistate.ViewDaily || s.ActiveView == uistate.ViewDefault) && s.ActivePane == uistate.PaneIssues {
				return handleCycleIssueSection(s, deps, -1)
			}
			return s, nil, false
		},
		"right": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ActivePane == uistate.PaneScope {
				return handleCursor(s, deps, 1)
			}
			if (s.ActiveView == uistate.ViewDaily || s.ActiveView == uistate.ViewDefault) && s.ActivePane == uistate.PaneIssues {
				return handleCycleIssueSection(s, deps, 1)
			}
			return s, nil, false
		},
		"v":    func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleOpenViewJump(s, deps) },
		"R":    func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleRescanExportAssets(s, deps) },
		"down": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, 1) },
		"up":   func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleCursor(s, deps, -1) },
		"f": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if !focusStartContext(s) {
				return s, nil, false
			}
			cmd := deps.StartFocusFromSelection(&s)
			return s, cmd, true
		},
		"m": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive || timerIsActive(s) || !issueEditorContext(s) {
				return s, nil, false
			}
			if cmd, handled := deps.OpenManualSessionDialog(&s); handled {
				return s, cmd, true
			}
			return s, nil, false
		},
		"s": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return handleIssueStatus(s, deps)
		},
		"y": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			return s, deps.ToggleSelectedIssueToday(&s), true
		},
		"P": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive || s.ActivePane != uistate.PaneIssues {
				return s, nil, false
			}
			return s, deps.ToggleSelectedIssuePinnedDaily(&s), true
		},
		"d": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if !issueEditorContext(s) || s.ProtectedModeActive {
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
		"+": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		"=": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		"-": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, -10) },
		"?": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if deps.OpenHelpDialog == nil {
				return s, nil, false
			}
			if deps.OpenHelpDialog(&s) {
				return s, nil, true
			}
			return s, nil, false
		},
		"a": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive {
				return s, nil, false
			}
			deps.HandleCreateAction(&s)
			return s, nil, true
		},
		"e": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive || timerIsActive(s) || !issueEditorContext(s) {
				return s, nil, false
			}
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		"D": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ActiveView == uistate.ViewAlerts {
				return handleDeleteSelectedAlertReminder(s, deps)
			}
			cmd, handled := deps.DeleteSelectionAction(&s)
			return s, cmd, handled
		},
		"x": func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.ToggleMomentumAction(&s)
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
	keyregistry.RegisterDaily(
		r,
		uistate.ViewDaily,
		uistate.PaneIssues,
		uistate.PaneHabits,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetDate(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleOpenExportDaily(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			deps.OpenCheckoutContextDialog(&s)
			return s, nil, true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return s, nil, false },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.ProtectedModeActive || timerIsActive(s) {
				return s, nil, false
			}
			if deps.OpenHabitLogAction != nil {
				cmd, handled := deps.OpenHabitLogAction(&s)
				if handled {
					return s, cmd, true
				}
			}
			if cmd, handled := deps.OpenManualSessionDialog(&s); handled {
				return s, cmd, true
			}
			return s, nil, false
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return s, nil, false },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.SetHabitFailedAction(&s)
			return s, cmd, handled
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.ToggleHabitCompletedAction(&s)
			return s, cmd, handled
		},
	)
	r.RegisterView(uistate.ViewDaily, "C", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenSummaryCalendar == nil {
			return s, nil, false
		}
		return s, nil, deps.OpenSummaryCalendar(&s)
	})
	r.RegisterView(uistate.ViewDaily, "S", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenDaySummary != nil && deps.OpenDaySummary(&s) {
			return s, nil, true
		}
		s.DailyDetailDay = !s.DailyDetailDay
		return s, nil, true
	})
	r.RegisterView(uistate.ViewDaily, "w", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleOpenCheckIn(s, deps)
	})
	for _, view := range []uistate.View{uistate.ViewDaily, uistate.ViewAway} {
		r.RegisterView(view, "W", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleToggleAwayMode(s, deps)
		})
	}
	for _, view := range []uistate.View{
		uistate.ViewDefault,
		uistate.ViewHabitList,
		uistate.ViewReports,
	} {
		r.RegisterView(view, "E", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleOpenExportDaily(s, deps)
		})
	}
	// The Issues tab's scope tree edits its repo or stream with e, whatever the timer.
	for _, view := range []uistate.View{uistate.ViewDefault, uistate.ViewHabitList} {
		r.RegisterPane(
			view,
			uistate.PaneScope,
			"e",
			func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
				if s.ProtectedModeActive {
					return s, nil, false
				}
				cmd, handled := deps.OpenEditorAction(&s)
				return s, cmd, handled
			},
		)
	}
	r.RegisterPane(
		uistate.ViewHabitList,
		uistate.PaneHabits,
		"e",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenEditorAction(&s)
			return s, cmd, handled
		},
	)
	r.RegisterPane(
		uistate.ViewHabitList,
		uistate.PaneHabits,
		"f",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return s, nil, false
		},
	)
	r.RegisterPane(
		uistate.ViewHabitList,
		uistate.PaneHabits,
		"m",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			if s.Timer != nil && s.Timer.State != "idle" {
				return s, nil, false
			}
			if deps.OpenHabitLogAction != nil {
				cmd, handled := deps.OpenHabitLogAction(&s)
				return s, cmd, handled
			}
			return s, nil, false
		},
	)
	keyregistry.RegisterSettings(
		r,
		uistate.ViewSettings,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleAdjustSelectedSetting(s, deps, -1)
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustSelectedSetting(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedSetting(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedSetting(s, deps) },
	)
	keyregistry.RegisterSettings(
		r,
		uistate.ViewAlerts,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleAdjustSelectedAlert(s, deps, -1)
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustSelectedAlert(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedAlert(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleActivateSelectedAlert(s, deps) },
	)
	keyregistry.RegisterConfig(
		r,
		uistate.ViewConfig,
		uistate.PaneConfig,
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
	keyregistry.RegisterSession(
		r,
		uistate.ViewSessionActive,
		uistate.ViewSessionHistory,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handlePauseSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResumeSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdvanceSession(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleEndSession(s, deps) },
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
	r.RegisterView(
		uistate.ViewSessionActive,
		"m",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleStructuredManualPause(s, deps)
		},
	)
	keyregistry.RegisterUpdates(
		r,
		uistate.ViewUpdates,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.UpdateChecking = true
			return s, deps.CheckUpdateNow(), true
		},
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			cmd, handled := deps.OpenSelectionAction(&s)
			return s, cmd, handled
		},
	)
	r.RegisterView(
		uistate.ViewUpdates,
		"d",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			s.UpdateDiagnosticsExpanded = !s.UpdateDiagnosticsExpanded
			return s, nil, true
		},
	)
	for _, action := range supportActions {
		run := action.run
		r.RegisterView(uistate.ViewSupport, action.key, func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return s, run(s, deps), true
		})
	}
	r.RegisterView(uistate.ViewSupport, "up", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		s.SupportCursor = max(0, s.SupportCursor-1)
		return s, nil, true
	})
	r.RegisterView(uistate.ViewSupport, "down", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		s.SupportCursor = min(len(supportActions)-1, s.SupportCursor+1)
		return s, nil, true
	})
	r.RegisterView(uistate.ViewSupport, "enter", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		action := supportActions[min(max(0, s.SupportCursor), len(supportActions)-1)]
		return s, action.run(s, deps), true
	})
	r.RegisterView(
		uistate.ViewMomentum,
		"+",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleShiftMomentumWindow(s, deps, 1)
		},
	)
	r.RegisterView(
		uistate.ViewMomentum,
		"-",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleShiftMomentumWindow(s, deps, -1)
		},
	)
	r.RegisterView(
		uistate.ViewMomentum,
		"0",
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
			return handleResetMomentumWindow(s, deps)
		},
	)
	keyregistry.RegisterWellbeing(
		r,
		uistate.ViewMomentum,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetDate(s, deps) },
	)
	keyregistry.RegisterRollup(
		r,
		uistate.ViewReview,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftDate(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetDate(s, deps) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftReviewWindow(s, deps, 1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleShiftReviewWindow(s, deps, -1) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleResetReviewWindow(s, deps) },
	)
	r.RegisterView(uistate.ViewReview, "E", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenDateDialog(&s) {
			return s, nil, true
		}
		return s, nil, false
	})
	r.RegisterView(uistate.ViewReview, "up", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleReviewStepDay(s, deps, -1)
	})
	r.RegisterView(uistate.ViewReview, "down", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleReviewStepDay(s, deps, 1)
	})
	r.RegisterView(uistate.ViewReview, "S", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenReviewDialog != nil && deps.OpenReviewDialog(&s, "time") {
			return s, nil, true
		}
		return handleToggleReviewTime(s, deps)
	})
	r.RegisterView(uistate.ViewReview, "enter", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		if deps.OpenReviewDialog != nil {
			deps.OpenReviewDialog(&s, "day")
		}
		return s, nil, true
	})
	r.RegisterView(uistate.ViewReview, "w", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleOpenCheckIn(s, deps)
	})
	r.RegisterView(uistate.ViewReview, "W", func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
		return handleToggleAwayMode(s, deps)
	})
	keyregistry.RegisterOps(
		r,
		uistate.ViewOps,
		uistate.PaneOps,
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, 10) },
		func(s State, _ tea.KeyMsg) (tea.Model, tea.Cmd, bool) { return handleAdjustOpsLimit(s, deps, -10) },
	)
	return r
}
