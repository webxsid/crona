package tui

import (
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------- View / Pane types ----------

type View string

const (
	ViewDefault View = "default"
	ViewDaily   View = "daily"
	ViewMeta    View = "meta"
	ViewSession View = "session"
	ViewScratch View = "scratchpads"
	ViewOps     View = "ops"
)

// viewOrder only includes the tab-switchable views.
var viewOrder = []View{ViewDefault, ViewMeta, ViewScratch, ViewOps, ViewDaily}

type Pane string

const (
	PaneRepos       Pane = "repos"
	PaneStreams     Pane = "streams"
	PaneIssues      Pane = "issues"
	PaneScratchpads Pane = "scratchpads"
	PaneOps         Pane = "ops"
)

// viewPanes lists the focusable panes for each view.
var viewPanes = map[View][]Pane{
	ViewDefault: {PaneIssues},
	ViewDaily:   {PaneIssues},
	ViewMeta:    {PaneRepos, PaneStreams, PaneIssues},
	ViewSession: {},
	ViewScratch: {PaneScratchpads},
	ViewOps:     {PaneOps},
}

// viewDefaultPane is the initial focused pane when entering a view.
var viewDefaultPane = map[View]Pane{
	ViewDefault: PaneIssues,
	ViewDaily:   PaneIssues,
	ViewMeta:    PaneRepos,
	ViewSession: PaneIssues,
	ViewScratch: PaneScratchpads,
	ViewOps:     PaneOps,
}

// ---------- Messages ----------

type reposLoadedMsg struct{ repos []api.Repo }
type streamsLoadedMsg struct{ streams []api.Stream }
type issuesLoadedMsg struct {
	streamID string
	issues   []api.Issue
}
type allIssuesLoadedMsg struct{ issues []api.IssueWithMeta }
type dailySummaryLoadedMsg struct{ summary *api.DailyIssueSummary }
type issueSessionsLoadedMsg struct {
	issueID  string
	sessions []api.Session
}
type scratchpadsLoadedMsg struct{ pads []api.ScratchPad }
type opsLoadedMsg struct{ ops []api.Op }
type contextLoadedMsg struct{ ctx *api.ActiveContext }
type timerLoadedMsg struct{ timer *api.TimerState }
type healthLoadedMsg struct{ health *api.Health }
type kernelEventMsg struct{ event api.KernelEvent }
type kernelShutdownMsg struct{}
type timerTickMsg struct{ seq int }
type healthTickMsg struct{}
type errMsg struct{ err error }

// ---------- Model ----------

type Model struct {
	// kernel client
	client *api.Client

	// SSE
	sseStop chan struct{}

	// view / navigation
	view    View
	pane    Pane
	cursor  map[Pane]int
	filters map[Pane]string

	// pane-local search/filter input
	filterEditing  bool
	filterPane     Pane
	filterInput    textinput.Model
	opsLimit       int
	opsLimitPinned bool

	// data
	repos         []api.Repo
	streams       []api.Stream
	issues        []api.Issue // context-filtered (by active streamId)
	allIssues     []api.IssueWithMeta
	dailySummary  *api.DailyIssueSummary
	dashboardDate string
	issueSessions []api.Session
	scratchpads   []api.ScratchPad
	ops           []api.Op
	context       *api.ActiveContext
	timer         *api.TimerState
	health        *api.Health
	elapsed       int // local seconds since last timer.state event
	timerTickSeq  int

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
	dialog            string // "" | "create_scratchpad" | "confirm_delete"
	dialogInputs      []textinput.Model
	dialogFocusIdx    int
	dialogDeleteID    string // scratchpad id pending deletion
	dialogIssueID     string
	dialogIssueStatus string
	dialogRepoID      string
	dialogRepoName    string
	dialogStreamID    string
	dialogStreamName  string
	dialogRepoIndex   int
	dialogStreamIndex int
	dialogParent      string
	dialogDateMonth   string
	dialogDateCursor  string

	// status / error flash
	statusMsg string
}

// SetSSEChannel provides the SSE event channel from main before the program starts.
func SetSSEChannel(ch <-chan api.KernelEvent) {
	sseChannel = ch
}

func New(baseURL, token, scratchDir string, done chan struct{}) Model {
	return Model{
		client:  api.NewClient(baseURL, token, scratchDir),
		sseStop: done,
		view:    ViewDefault,
		pane:    PaneIssues,
		cursor: map[Pane]int{
			PaneRepos:       0,
			PaneStreams:     0,
			PaneIssues:      0,
			PaneScratchpads: 0,
			PaneOps:         0,
		},
		filters: map[Pane]string{
			PaneRepos:       "",
			PaneStreams:     "",
			PaneIssues:      "",
			PaneScratchpads: "",
			PaneOps:         "",
		},
	}
}

// sseChannel receives kernel events forwarded from main.go.
var sseChannel <-chan api.KernelEvent

// ---------- Init ----------

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadRepos(m.client),
		loadAllIssues(m.client),
		loadDailySummary(m.client, ""),
		loadScratchpads(m.client),
		loadOps(m.client, m.currentOpsLimit()),
		loadContext(m.client),
		loadTimer(m.client),
		loadHealth(m.client),
		healthTickAfter(),
		waitForSSE(sseChannel),
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

func (m Model) contentHeight() int {
	headerH := 4
	if m.width > 0 {
		headerH = lipgloss.Height(m.renderHeader())
	}
	helpH := 1
	if m.width > 0 {
		helpH = lipgloss.Height(m.renderHelpBar())
	}
	availableHeight := m.height - headerH - helpH
	if m.statusMsg != "" {
		availableHeight--
	}
	if availableHeight < 4 {
		availableHeight = 4
	}
	return availableHeight
}

func (m Model) sidebarWidth() int {
	if m.width < 90 {
		return 18
	}
	return 22
}

func (m Model) mainContentWidth() int {
	width := m.width - m.sidebarWidth() - 4
	if width < 40 {
		return 40
	}
	return width
}
