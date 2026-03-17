package app

import (
	"time"

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
	ViewSettings       View = "settings"
)

// viewOrder only includes the tab-switchable views.
var viewOrder = []View{ViewSessionHistory, ViewDaily, ViewWellbeing, ViewDefault, ViewMeta, ViewScratch, ViewOps, ViewSettings}

type Pane string

const (
	PaneRepos       Pane = "repos"
	PaneStreams     Pane = "streams"
	PaneIssues      Pane = "issues"
	PaneHabits      Pane = "habits"
	PaneSessions    Pane = "sessions"
	PaneScratchpads Pane = "scratchpads"
	PaneOps         Pane = "ops"
	PaneSettings    Pane = "settings"
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
	ViewSettings:       {PaneSettings},
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
	ViewSettings:       PaneSettings,
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
	repos          []api.Repo
	streams        []api.Stream
	issues         []api.Issue // context-filtered (by active streamId)
	habits         []api.Habit
	allIssues      []api.IssueWithMeta
	dueHabits      []api.HabitDailyItem
	dailySummary   *api.DailyIssueSummary
	dashboardDate  string
	wellbeingDate  string
	dailyCheckIn   *api.DailyCheckIn
	metricsRange   []api.DailyMetricsDay
	metricsRollup  *api.MetricsRollup
	streaks        *api.StreakSummary
	issueSessions  []api.Session
	sessionHistory []api.SessionHistoryEntry
	sessionDetail  *api.SessionDetail
	scratchpads    []api.ScratchPad
	stashes        []api.Stash
	ops            []api.Op
	context        *api.ActiveContext
	timer          *api.TimerState
	health         *api.Health
	settings       *api.CoreSettings
	kernelInfo     *api.KernelInfo
	elapsed        int // local seconds since last timer.state event
	timerTickSeq   int

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
	dialog               string // "" | "create_scratchpad" | "confirm_delete" | "stash_list"
	dialogInputs         []textinput.Model
	dialogDescription    textarea.Model
	dialogDescriptionOn  bool
	dialogDescriptionIdx int
	dialogFocusIdx       int
	dialogDeleteID       string // scratchpad id pending deletion
	dialogDeleteKind     string
	dialogDeleteLabel    string
	dialogSessionID      string
	dialogIssueID        int64
	dialogHabitID        int64
	dialogIssueStatus    string
	dialogCheckInDate    string
	dialogRepoID         int64
	dialogRepoName       string
	dialogStreamID       int64
	dialogStreamName     string
	dialogRepoIndex      int
	dialogStreamIndex    int
	dialogParent         string
	dialogDateMonth      string
	dialogDateCursor     string
	dialogStashCursor    int
	dialogStatusItems    []sharedtypes.IssueStatus
	dialogStatusCursor   int
	dialogStatusLabel    string
	dialogStatusRequired bool
	dialogViewTitle      string
	dialogViewName       string
	dialogViewMeta       string
	dialogViewBody       string

	// status / error flash
	statusMsg string
	statusSeq int
	statusErr bool

	// overlay help
	helpOpen          bool
	sessionDetailOpen bool
	sessionDetailY    int
}

// SetEventChannel provides the kernel event channel from main before the program starts.
func SetEventChannel(ch <-chan api.KernelEvent) {
	eventChannel = ch
}

func New(socketPath, scratchDir string, env string, done chan struct{}) Model {
	return Model{
		client:              api.NewClient(socketPath, scratchDir),
		eventStop:           done,
		view:                ViewDaily,
		pane:                PaneIssues,
		defaultIssueSection: DefaultIssueSectionOpen,
		cursor: map[Pane]int{
			PaneRepos:       0,
			PaneStreams:     0,
			PaneIssues:      0,
			PaneHabits:      0,
			PaneSessions:    0,
			PaneScratchpads: 0,
			PaneOps:         0,
			PaneSettings:    0,
		},
		filters: map[Pane]string{
			PaneRepos:       "",
			PaneStreams:     "",
			PaneIssues:      "",
			PaneHabits:      "",
			PaneSessions:    "",
			PaneScratchpads: "",
			PaneOps:         "",
			PaneSettings:    "",
		},
		kernelInfo: &api.KernelInfo{Env: env},
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
		loadSessionHistory(m.client, 200),
		loadScratchpads(m.client),
		loadOps(m.client, m.currentOpsLimit()),
		loadContext(m.client),
		loadTimer(m.client),
		loadHealth(m.client),
		loadSettings(m.client),
		loadKernelInfo(m.client),
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
