package model

import (
	"crona/tui/internal/tui/scope"
	"time"

	sharedposthog "crona/shared/posthog"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	commands "crona/tui/internal/tui/commands"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	selectionpkg "crona/tui/internal/tui/selection"
	uistate "crona/tui/internal/tui/state"
	"crona/tui/internal/tui/terminaltitle"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------- View / Pane types ----------

type View = uistate.View

const (
	ViewAway           = uistate.ViewAway
	ViewDefault        = uistate.ViewDefault
	ViewDaily          = uistate.ViewDaily
	ViewReview         = uistate.ViewReview
	ViewMomentum       = uistate.ViewMomentum
	ViewHabitList      = uistate.ViewHabitList
	ViewSessionHistory = uistate.ViewSessionHistory
	ViewHabitHistory   = uistate.ViewHabitHistory
	ViewSessionActive  = uistate.ViewSessionActive
	ViewOps            = uistate.ViewOps
	ViewReports        = uistate.ViewReports
	ViewConfig         = uistate.ViewConfig
	ViewSettings       = uistate.ViewSettings
	ViewAlerts         = uistate.ViewAlerts
	ViewUpdates        = uistate.ViewUpdates
	ViewSupport        = uistate.ViewSupport
)

type Pane = uistate.Pane

const (
	PaneRepos         = uistate.PaneRepos
	PaneStreams       = uistate.PaneStreams
	PaneScope         = uistate.PaneScope
	PaneIssues        = uistate.PaneIssues
	PaneHabits        = uistate.PaneHabits
	PaneReviewDays    = uistate.PaneReviewDays
	PaneSessions      = uistate.PaneSessions
	PaneHabitHistory  = uistate.PaneHabitHistory
	PaneMomentumCards = uistate.PaneMomentumCards
	PaneOps           = uistate.PaneOps
	PaneExportReports = uistate.PaneExportReports
	PaneConfig        = uistate.PaneConfig
	PaneSettings      = uistate.PaneSettings
	PaneAlerts        = uistate.PaneAlerts
)

type DefaultIssueSection = uistate.DefaultIssueSection
type DailyTaskSection = uistate.DailyTaskSection

const (
	DefaultIssueSectionOpen      = uistate.DefaultIssueSectionOpen
	DefaultIssueSectionCompleted = uistate.DefaultIssueSectionCompleted
	DailyTaskSectionPlanned      = uistate.DailyTaskSectionPlanned
	DailyTaskSectionOverdue      = uistate.DailyTaskSectionOverdue
	DailyTaskSectionUpcoming     = uistate.DailyTaskSectionUpcoming
)

// ---------- Model ----------

type Model struct {
	// kernel client
	client    *api.Client
	telemetry sharedposthog.Client

	// kernel event stream
	eventStop *eventStreamStop

	// view / navigation
	view                View
	pane                Pane
	cursor              map[Pane]int
	filters             map[Pane]string
	defaultIssueSection DefaultIssueSection
	dailyTaskSection    DailyTaskSection

	opsLimit       int
	opsLimitPinned bool

	// data
	repos      []api.Repo
	streams    []api.Stream
	issues     []api.Issue // context-filtered (by active streamId)
	habits     []api.Habit
	allHabits  []api.HabitWithMeta
	allStreams []api.Stream
	// browseScope is the Issues tab's scope tree selection, separate from the checked-out context.
	browseScope  scope.Scope
	allIssues    []api.IssueWithMeta
	dueHabits    []api.HabitDailyItem
	todayHabits  []api.HabitDailyItem
	dailySummary *api.DailyIssueSummary
	dailyPlan    *api.DailyPlan
	tags         []api.Tag
	// date is the shared date every dated view shows; "" follows the logical today.
	date string
	// reviewWindowDays end on reviewWindowEnd; reviewShowTime puts Time in Trends' slot.
	reviewWindowDays int
	reviewWindowEnd  string
	reviewShowTime   bool
	summarySnapshot  *api.SummarySnapshot
	// dailyDetailDay shows the Day summary in Daily's detail pane until the selection moves.
	dailyDetailDay        bool
	momentumWindowDays    int
	dailyCheckIn          *api.DailyCheckIn
	metricsRange          []api.DailyMetricsDay
	metricsRollup         *api.MetricsRollup
	rollupMetricsRange    []api.DailyMetricsDay
	rollupMetricsRollup   *api.MetricsRollup
	momentumMetricsRange  []api.DailyMetricsDay
	momentumMetricsRollup *api.MetricsRollup
	streaks               *api.StreakSummary
	dailyStreaks          *api.StreakSummary
	dashboardWindow       *api.DashboardWindowSummary
	dailyFocusScore       *api.FocusScoreSummary
	weeklyFocusScore      *api.FocusScoreSummary
	repoDistribution      *api.TimeDistributionSummary
	streamDistribution    *api.TimeDistributionSummary
	issueDistribution     *api.TimeDistributionSummary
	segmentDistribution   *api.TimeDistributionSummary
	goalProgress          *api.GoalProgressSummary
	exportAssets          *api.ExportAssetStatus
	exportReports         []api.ExportReportFile
	issueSessions         []api.Session
	sessionHistory        []api.SessionHistoryEntry
	habitHistory          []api.HabitCompletion
	habitHistoryHabitID   int64
	habitHistoryTitle     string
	habitHistoryMeta      string
	sessionDetail         *api.SessionDetail
	// sessionPreview feeds Log's Sessions detail pane; sessionPreviewFailed holds the ID whose
	// fetch failed.
	sessionPreview            *api.SessionDetail
	sessionPreviewFailed      string
	ops                       []api.Op
	context                   *api.ActiveContext
	timer                     *api.TimerState
	health                    *api.Health
	alertStatus               *api.AlertStatus
	alertReminders            []api.AlertReminder
	updateStatus              *api.UpdateStatus
	updateChecking            bool
	updateDiagnosticsExpanded bool
	supportCursor             int
	currentExecutablePath     string
	settings                  *api.CoreSettings
	habitStreakDefs           []api.HabitStreakDefinition
	momentumCards             []api.MomentumCard
	kernelInfo                *api.KernelInfo
	elapsed                   int // local seconds since last timer.state event
	timerTickSeq              int
	pendingEvents             []api.KernelEvent
	eventFlushScheduled       bool
	// reviewStale is set when data changed while Review was off screen.
	reviewStale            bool
	lastTimerActivityTouch time.Time
	terminalTitleEnabled   bool
	lastTerminalTitle      string

	// terminal dimensions
	width  int
	height int

	// dialog state
	dialog                               string // "" | "confirm_delete"
	dialogInputs                         []textinput.Model
	dialogDescription                    textarea.Model
	dialogDescriptionOn                  bool
	dialogDescriptionIdx                 int
	dialogFocusIdx                       int
	dialogErrorMessage                   string
	dialogDeleteID                       string
	dialogDeleteKind                     string
	dialogDeleteLabel                    string
	dialogSessionID                      string
	dialogSessionHistoryContext          string
	dialogSessionHistoryNoteInputs       []textarea.Model
	dialogSessionHistorySegments         []dialogstate.SessionHistorySegmentRow
	dialogIssueID                        int64
	dialogAdhocTimer                     bool
	dialogPromotingAdhoc                 bool
	dialogViewEntityAdhoc                bool
	dialogAdhocResolution                sharedtypes.AdhocSessionResolution
	dialogAdhocIssueIDs                  []int64
	dialogAdhocIssueTab                  int
	dialogAdhocIssueTabCounts            [4]int
	dialogAdhocElapsedSeconds            int
	dialogAdhocDate                      string
	dialogNewIssueTitle                  string
	dialogHabitID                        int64
	dialogIssueStatus                    string
	dialogCheckInDate                    string
	dialogRepoID                         int64
	dialogRepoName                       string
	dialogRepoItems                      []string
	dialogRepoItemIDs                    []int64
	dialogStreamID                       int64
	dialogStreamName                     string
	dialogRepoIndex                      int
	dialogStreamIndex                    int
	dialogParent                         string
	dialogDateMonth                      string
	dialogDateCursor                     string
	dialogStatusItems                    []sharedtypes.IssueStatus
	dialogStatusCursor                   int
	dialogChoiceItems                    []string
	dialogChoiceValues                   []string
	dialogChoiceDetails                  []string
	dialogTemplateAssets                 []sharedtypes.ExportTemplateAsset
	dialogChoiceCursor                   int
	dialogMomentumRepoInput              textinput.Model
	dialogMomentumStreamInput            textinput.Model
	dialogProcessing                     bool
	dialogProcessingLabel                string
	dialogStatusLabel                    string
	dialogStatusRequired                 bool
	dialogViewTitle                      string
	dialogViewName                       string
	dialogIssueEstimateMins              *int
	dialogIssueWorkedSeconds             int
	dialogReminderID                     string
	dialogReminderKind                   sharedtypes.AlertReminderKind
	dialogViewMeta                       string
	dialogViewBody                       string
	dialogViewPath                       string
	dialogSupportBundlePath              string
	dialogProtectionStep                 int
	dialogProtectionCursor               int
	dialogProtectionStreaks              []sharedtypes.StreakKind
	dialogProtectionWeekdays             []int
	dialogCadenceIdx                     int
	dialogCadenceDays                    []int
	dialogCadenceCursor                  int
	dialogProtectionDates                []string
	dialogHabitStreakStep                int
	dialogHabitStreakCursor              int
	dialogHabitStreakOriginalDefs        []sharedtypes.HabitStreakDefinition
	dialogHabitStreakDefs                []sharedtypes.HabitStreakDefinition
	dialogHabitStreakDraft               sharedtypes.HabitStreakDefinition
	dialogHabitStreakEditIdx             int
	dialogExportPresetKind               sharedtypes.ExportReportKind
	dialogExportPresetFormat             sharedtypes.ExportFormat
	dialogExportPresetOutput             sharedtypes.ExportOutputMode
	dialogExportIncludePDF               bool
	dialogExportCategory                 string
	dialogExportStart                    string
	dialogExportEnd                      string
	dialogPromptGlyphMode                sharedtypes.PromptGlyphMode
	dialogTelemetryStep                  int
	dialogTelemetryUsage                 bool
	dialogTelemetryErrors                bool
	dialogTelemetryPrivacyCursor         int
	dialogTelemetryReviewCursor          int
	dialogDayBoundaryEnabled             bool
	dialogDayBoundaryTimezone            string
	dialogSettingKey                     sharedtypes.CoreSettingsKey
	dialogDayBoundaryStep                int
	dialogDayBoundarySchedule            sharedtypes.DayBoundarySchedule
	dialogDayBoundaryOverrides           []dialogstate.DayBoundaryOverride
	dialogDayBoundaryOverrideCursor      int
	dialogDayBoundarySelectedDays        []int
	dialogDayBoundaryEditingTime         string
	dialogDayBoundaryEditingOverride     bool
	dialogPomodoroFocusSeconds           int
	dialogPomodoroFocusChoice            int
	dialogPomodoroBreakSeconds           int
	dialogPomodoroBreakChoice            int
	dialogPomodoroLongBreakSeconds       int
	dialogPomodoroLongBreakChoice        int
	dialogPomodoroCyclesBeforeLongBreak  int
	dialogPomodoroCycles                 int
	dialogTimerCountdownSeconds          int
	dialogHardLimitKind                  sharedtypes.TimerHardLimitKind
	dialogHardLimitTotalSeconds          int
	dialogHardLimitFocusSeconds          int
	dialogHardLimitBreakSeconds          int
	dialogHardLimitLongBreakSeconds      int
	dialogHardLimitCyclesBeforeLongBreak int

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

// SetEventChannel provides the kernel event channel from main before the program starts.
func SetEventChannel(ch <-chan api.KernelEvent) {
	eventChannel = ch
}

func New(
	transport, endpoint, scratchDir string,
	env string,
	executablePath string,
	done chan struct{},
	telemetry sharedposthog.Client,
) Model {
	model := Model{
		client:              api.NewClient(transport, endpoint),
		telemetry:           telemetry,
		eventStop:           newEventStreamStop(done),
		view:                ViewDaily,
		pane:                PaneIssues,
		defaultIssueSection: DefaultIssueSectionOpen,
		dailyTaskSection:    DailyTaskSectionPlanned,
		cursor: map[Pane]int{
			PaneRepos:         0,
			PaneStreams:       0,
			PaneIssues:        0,
			PaneHabits:        0,
			PaneSessions:      0,
			PaneHabitHistory:  0,
			PaneMomentumCards: 0,
			PaneOps:           0,
			PaneExportReports: 0,
			PaneConfig:        0,
			PaneSettings:      0,
			PaneAlerts:        0,
		},
		filters: map[Pane]string{
			PaneRepos:         "",
			PaneStreams:       "",
			PaneIssues:        "",
			PaneHabits:        "",
			PaneSessions:      "",
			PaneHabitHistory:  "",
			PaneMomentumCards: "",
			PaneOps:           "",
			PaneExportReports: "",
			PaneConfig:        "",
			PaneSettings:      "",
			PaneAlerts:        "",
		},
		currentExecutablePath: executablePath,
		kernelInfo:            &api.KernelInfo{Env: env},
		terminalTitleEnabled:  true,
		momentumWindowDays:    30,
	}
	model.lastTerminalTitle = terminaltitle.Sanitize(model.terminalTitle())
	logger.Infof(
		"tui model new: view=%s pane=%s env=%s telemetry=%t executable=%q",
		model.view,
		model.pane,
		env,
		telemetry != nil,
		executablePath,
	)
	return model
}

// eventChannel receives kernel events forwarded from main.go.
var eventChannel <-chan api.KernelEvent

// ---------- Init ----------

func (m Model) Init() tea.Cmd {
	logger.Infof(
		"tui model init: view=%s pane=%s width=%d height=%d dialog=%q",
		m.view,
		m.pane,
		m.width,
		m.height,
		m.dialog,
	)
	cmds := []tea.Cmd{
		commands.LoadRepos(m.client),
		commands.LoadAllHabits(m.client),
		commands.LoadTags(m.client),
		commands.LoadHabitStreakDefinitions(m.client),
		commands.LoadMomentumRange(
			m.client,
			time.Now().Format("2006-01-02"),
			m.currentMomentumWindowDays(),
		),
		commands.LoadAllIssues(m.client),
		commands.LoadDueHabits(m.client, time.Now().Format("2006-01-02")),
		commands.LoadTodayHabits(m.client, m.currentLogicalDate()),
		commands.LoadDailySummary(m.client, ""),
		commands.LoadSummarySnapshot(m.client, time.Now().Format("2006-01-02")),
		commands.LoadDailyStreaks(m.client, time.Now().Format("2006-01-02")),
		commands.LoadWellbeingDay(m.client, time.Now().Format("2006-01-02")),
		m.reloadReview(),
		loadSessionHistoryForModel(m, 200),
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
	if p == PaneMomentumCards {
		snapshot := m.selectionSnapshot()
		return len(selectionpkg.FilteredIndices(snapshot, p))
	}
	if p == PaneHabitHistory {
		return len(m.habitHistory)
	}
	if p == PaneAlerts {
		return alertsmeta.FilteredSelectableCount(
			m.filters[PaneAlerts],
			m.settings,
			m.alertStatus,
			m.alertReminders,
		)
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
