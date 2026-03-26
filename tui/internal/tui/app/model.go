package app

import (
	"time"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------- View / Pane types ----------

type View string

const (
	ViewDefault        View = "default"
	ViewDaily          View = "daily"
	ViewMeta           View = "meta"
	ViewSessionHistory View = "session_history"
	ViewSessionActive  View = "session_active"
	ViewScratch        View = "scratchpads"
	ViewOps            View = "ops"
	ViewWellbeing      View = "wellbeing"
	ViewReports        View = "reports"
	ViewConfig         View = "config"
	ViewSettings       View = "settings"
	ViewUpdates        View = "updates"
)

// viewOrder only includes the tab-switchable views.
var viewOrder = []View{ViewSessionHistory, ViewDaily, ViewWellbeing, ViewReports, ViewDefault, ViewMeta, ViewScratch, ViewOps, ViewConfig, ViewSettings, ViewUpdates}

type Pane string

const (
	PaneRepos         Pane = "repos"
	PaneStreams       Pane = "streams"
	PaneIssues        Pane = "issues"
	PaneHabits        Pane = "habits"
	PaneSessions      Pane = "sessions"
	PaneScratchpads   Pane = "scratchpads"
	PaneOps           Pane = "ops"
	PaneExportReports Pane = "export_reports"
	PaneConfig        Pane = "config"
	PaneSettings      Pane = "settings"
)

type DefaultIssueSection string

const (
	DefaultIssueSectionOpen      DefaultIssueSection = "open"
	DefaultIssueSectionCompleted DefaultIssueSection = "completed"
)

// viewPanes lists the focusable panes for each view.
var viewPanes = map[View][]Pane{
	ViewDefault:        {PaneIssues},
	ViewDaily:          {PaneIssues, PaneHabits},
	ViewMeta:           {PaneRepos, PaneStreams, PaneIssues, PaneHabits},
	ViewSessionHistory: {PaneSessions},
	ViewSessionActive:  {},
	ViewScratch:        {PaneScratchpads},
	ViewOps:            {PaneOps},
	ViewWellbeing:      {},
	ViewReports:        {PaneExportReports},
	ViewConfig:         {PaneConfig},
	ViewSettings:       {PaneSettings},
	ViewUpdates:        {},
}

// viewDefaultPane is the initial focused pane when entering a view.
var viewDefaultPane = map[View]Pane{
	ViewDefault:        PaneIssues,
	ViewDaily:          PaneIssues,
	ViewMeta:           PaneRepos,
	ViewSessionHistory: PaneSessions,
	ViewSessionActive:  PaneIssues,
	ViewScratch:        PaneScratchpads,
	ViewOps:            PaneOps,
	ViewWellbeing:      PaneIssues,
	ViewReports:        PaneExportReports,
	ViewConfig:         PaneConfig,
	ViewSettings:       PaneSettings,
	ViewUpdates:        PaneIssues,
}

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
	dialog                string // "" | "create_scratchpad" | "confirm_delete" | "stash_list"
	dialogInputs          []textinput.Model
	dialogDescription     textarea.Model
	dialogDescriptionOn   bool
	dialogDescriptionIdx  int
	dialogFocusIdx        int
	dialogDeleteID        string // scratchpad id pending deletion
	dialogDeleteKind      string
	dialogDeleteLabel     string
	dialogSessionID       string
	dialogIssueID         int64
	dialogHabitID         int64
	dialogIssueStatus     string
	dialogCheckInDate     string
	dialogRepoID          int64
	dialogRepoName        string
	dialogRepoItems       []string
	dialogRepoItemIDs     []int64
	dialogStreamID        int64
	dialogStreamName      string
	dialogRepoIndex       int
	dialogStreamIndex     int
	dialogParent          string
	dialogDateMonth       string
	dialogDateCursor      string
	dialogStashCursor     int
	dialogStatusItems     []sharedtypes.IssueStatus
	dialogStatusCursor    int
	dialogChoiceItems     []string
	dialogChoiceCursor    int
	dialogProcessing      bool
	dialogProcessingLabel string
	dialogStatusLabel     string
	dialogStatusRequired  bool
	dialogViewTitle       string
	dialogViewName        string
	dialogViewMeta        string
	dialogViewBody        string

	// status / error flash
	statusMsg string
	statusSeq int
	statusErr bool

	// overlay help
	helpOpen          bool
	sessionDetailOpen bool
	sessionDetailY    int
}

func (m Model) selfUpdateInstallAvailable() bool {
	return m.updateStatus != nil && m.updateStatus.InstallAvailable && m.selfUpdateUnsupportedReason() == ""
}

func (m Model) selfUpdateUnsupportedReason() string {
	if reason := nonStandardRuntimeReason(m.currentExecutablePath, config.TUIBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	if m.kernelInfo == nil {
		return "Kernel info is unavailable. Please update manually."
	}
	if reason := nonStandardRuntimeReason(m.kernelInfo.ExecutablePath, config.KernelBinaryNameForMode(kernelEnvMode(m.kernelInfo))); reason != "" {
		return reason
	}
	return ""
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
	return Model{
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
}

// eventChannel receives kernel events forwarded from main.go.
var eventChannel <-chan api.KernelEvent

// ---------- Init ----------

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadRepos(m.client),
		loadAllIssues(m.client),
		loadDueHabits(m.client, time.Now().Format("2006-01-02")),
		loadDailySummary(m.client, ""),
		loadWellbeing(m.client, time.Now().Format("2006-01-02")),
		loadSessionHistoryForModel(m, 200),
		loadScratchpads(m.client),
		loadOps(m.client, m.currentOpsLimit()),
		loadContext(m.client),
		loadTimer(m.client),
		loadHealth(m.client),
		loadUpdateStatus(m.client),
		loadSettings(m.client),
		loadKernelInfo(m.client),
		loadExportAssets(m.client),
		loadExportReports(m.client),
		healthTickAfter(),
		waitForEvent(eventChannel),
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
	return len(m.filteredIndices(p))
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
