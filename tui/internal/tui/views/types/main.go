package viewtypes

import (
	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ColorBlue    lipgloss.Color
	ColorCyan    lipgloss.Color
	ColorGreen   lipgloss.Color
	ColorMagenta lipgloss.Color
	ColorSubtle  lipgloss.Color
	ColorYellow  lipgloss.Color
	ColorRed     lipgloss.Color
	ColorDim     lipgloss.Color
	ColorWhite   lipgloss.Color

	StyleActive    lipgloss.Style
	StyleInactive  lipgloss.Style
	StylePaneTitle lipgloss.Style
	StyleDim       lipgloss.Style
	StyleCursor    lipgloss.Style
	StyleHeader    lipgloss.Style
	StyleError     lipgloss.Style
	StyleSelected  lipgloss.Style
	StyleNormal    lipgloss.Style
}

type ContentState struct {
	View                string
	Pane                string
	Width               int
	Height              int
	Cursors             map[string]int
	Filters             map[string]string
	ScratchpadOpen      bool
	ScratchpadName      string
	ScratchpadPath      string
	ScratchpadRendered  string
	Elapsed             int
	DashboardDate       string
	RollupStartDate     string
	RollupEndDate       string
	WellbeingDate       string
	DefaultIssueSection string
	SessionHistoryTitle string
	SessionHistoryMeta  string
	RestModeActive      bool
	AwayModeActive      bool
	RestModeMessage     string
	RestModeDetail      string

	Repos                  []api.Repo
	Streams                []api.Stream
	Issues                 []api.Issue
	DailyIssues            []api.Issue
	Habits                 []api.Habit
	AllIssues              []api.IssueWithMeta
	DefaultIssues          []api.IssueWithMeta
	DueHabits              []api.HabitDailyItem
	DailySummary           *api.DailyIssueSummary
	DailyPlan              *api.DailyPlan
	DailyCheckIn           *api.DailyCheckIn
	MetricsRange           []api.DailyMetricsDay
	MetricsRollup          *api.MetricsRollup
	Streaks                *api.StreakSummary
	DashboardWindow        *api.DashboardWindowSummary
	DailyFocusScore        *api.FocusScoreSummary
	WeeklyFocusScore       *api.FocusScoreSummary
	RepoDistribution       *api.TimeDistributionSummary
	StreamDistribution     *api.TimeDistributionSummary
	IssueDistribution      *api.TimeDistributionSummary
	SegmentDistribution    *api.TimeDistributionSummary
	GoalProgress           *api.GoalProgressSummary
	ExportAssets           *api.ExportAssetStatus
	ExportReports          []api.ExportReportFile
	IssueSessions          []api.Session
	SessionHistory         []api.SessionHistoryEntry
	Scratchpads            []api.ScratchPad
	Ops                    []api.Op
	Context                *api.ActiveContext
	Timer                  *api.TimerState
	Health                 *api.Health
	AlertStatus            *api.AlertStatus
	AlertReminders         []api.AlertReminder
	UpdateStatus           *api.UpdateStatus
	UpdateChecking         bool
	UpdateInstalling       bool
	UpdateInstallError     string
	UpdateInstallAvailable bool
	UpdateManualReason     string
	TUIExecutablePath      string
	KernelExecutablePath   string
	KernelInfo             *api.KernelInfo
	Settings               *api.CoreSettings
}
