package viewtypes

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	"crona/tui/internal/tui/scope"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ColorBlue      lipgloss.Color
	ColorCyan      lipgloss.Color
	ColorGreen     lipgloss.Color
	ColorDullGreen lipgloss.Color
	ColorMagenta   lipgloss.Color
	ColorSubtle    lipgloss.Color
	ColorYellow    lipgloss.Color
	ColorRed       lipgloss.Color
	ColorDullRed   lipgloss.Color
	ColorOrange    lipgloss.Color
	ColorDim       lipgloss.Color
	ColorWhite     lipgloss.Color

	StyleActive          lipgloss.Style
	StyleInactive        lipgloss.Style
	StylePaneTitle       lipgloss.Style
	StyleDim             lipgloss.Style
	StyleCursor          lipgloss.Style
	StyleHeader          lipgloss.Style
	StyleError           lipgloss.Style
	StyleSelected        lipgloss.Style
	StyleSelectedInverse lipgloss.Style
	StyleNormal          lipgloss.Style
}

type ContentState struct {
	View               string
	Pane               string
	SelectedIssueAdhoc bool
	// BrowseScope is the Issues tab's scope tree selection; AllStreams and AllHabits feed it.
	BrowseScope         scope.Scope
	AllStreams          []api.Stream
	AllHabits           []api.HabitWithMeta
	Width               int
	Height              int
	Cursors             map[string]int
	Filters             map[string]string
	Elapsed             int
	DashboardDate       string
	SummaryDate         string
	RollupStartDate     string
	RollupEndDate       string
	MomentumDate        string
	MomentumWindowDays  int
	MomentumTab         string
	MomentumHistoryY    int
	WellbeingDate       string
	WellbeingWindowDays int
	WeekStart           sharedtypes.WeekStart
	DefaultIssueSection string
	DailyTaskSection    string
	SessionHistoryTitle string
	SessionHistoryMeta  string
	HabitHistoryTitle   string
	HabitHistoryMeta    string
	RestModeActive      bool
	AwayModeActive      bool
	RestModeMessage     string
	RestModeDetail      string
	HistoricalAwayDate  string

	Repos       []api.Repo
	Streams     []api.Stream
	Issues      []api.Issue
	DailyIssues []api.Issue
	// DailyPlannedIssues is the Planned section whatever tab is shown; the strip counts it.
	DailyPlannedIssues []api.Issue
	// DailyDetailDay shows the Day summary in the detail pane instead of the selection.
	DailyDetailDay bool
	// DailyIssueIndices and DailyHabitIndices are the model's text-filtered raw indices into
	// DailyIssues and DueHabits; the pane cursors are positions in these lists.
	DailyIssueIndices         []int
	DailyHabitIndices         []int
	Habits                    []api.Habit
	AllIssues                 []api.IssueWithMeta
	DefaultIssues             []api.IssueWithMeta
	DueHabits                 []api.HabitDailyItem
	MomentumCards             []api.MomentumCard
	DailySummary              *api.DailyIssueSummary
	SummarySnapshot           *api.SummarySnapshot
	DailyPlan                 *api.DailyPlan
	DailyCheckIn              *api.DailyCheckIn
	MetricsRange              []api.DailyMetricsDay
	MetricsRollup             *api.MetricsRollup
	RollupMetricsRange        []api.DailyMetricsDay
	RollupMetricsRollup       *api.MetricsRollup
	MomentumMetricsRange      []api.DailyMetricsDay
	MomentumMetricsRollup     *api.MetricsRollup
	Streaks                   *api.StreakSummary
	DailyStreaks              *api.StreakSummary
	DashboardWindow           *api.DashboardWindowSummary
	DailyFocusScore           *api.FocusScoreSummary
	WeeklyFocusScore          *api.FocusScoreSummary
	RepoDistribution          *api.TimeDistributionSummary
	StreamDistribution        *api.TimeDistributionSummary
	IssueDistribution         *api.TimeDistributionSummary
	SegmentDistribution       *api.TimeDistributionSummary
	GoalProgress              *api.GoalProgressSummary
	ExportAssets              *api.ExportAssetStatus
	ExportReports             []api.ExportReportFile
	IssueSessions             []api.Session
	SessionHistory            []api.SessionHistoryEntry
	HabitHistory              []api.HabitCompletion
	Ops                       []api.Op
	Context                   *api.ActiveContext
	Timer                     *api.TimerState
	Health                    *api.Health
	AlertStatus               *api.AlertStatus
	AlertReminders            []api.AlertReminder
	UpdateStatus              *api.UpdateStatus
	UpdateChecking            bool
	UpdateInstallAvailable    bool
	UpdateDiagnosticsExpanded bool
	UpdateCommand             string
	UpdateManualReason        string
	TUIExecutablePath         string
	KernelExecutablePath      string
	KernelInfo                *api.KernelInfo
	Settings                  *api.CoreSettings
}

func (s ContentState) AwayDates() []string {
	if s.Settings == nil {
		return nil
	}
	return s.Settings.AwayDates
}
