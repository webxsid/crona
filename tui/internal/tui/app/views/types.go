package views

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
	WellbeingDate       string
	DefaultIssueSection string
	SessionHistoryTitle string
	SessionHistoryMeta  string

	Repos                  []api.Repo
	Streams                []api.Stream
	Issues                 []api.Issue
	DailyIssues            []api.Issue
	Habits                 []api.Habit
	AllIssues              []api.IssueWithMeta
	DefaultIssues          []api.IssueWithMeta
	DueHabits              []api.HabitDailyItem
	DailySummary           *api.DailyIssueSummary
	DailyCheckIn           *api.DailyCheckIn
	MetricsRange           []api.DailyMetricsDay
	MetricsRollup          *api.MetricsRollup
	Streaks                *api.StreakSummary
	ExportAssets           *api.ExportAssetStatus
	ExportReports          []api.ExportReportFile
	IssueSessions          []api.Session
	SessionHistory         []api.SessionHistoryEntry
	Scratchpads            []api.ScratchPad
	Ops                    []api.Op
	Context                *api.ActiveContext
	Timer                  *api.TimerState
	Health                 *api.Health
	UpdateStatus           *api.UpdateStatus
	UpdateChecking         bool
	UpdateInstalling       bool
	UpdateInstallOutput    string
	UpdateInstallError     string
	UpdateInstallAvailable bool
	UpdateManualReason     string
	TUIExecutablePath      string
	KernelExecutablePath   string
	Settings               *api.CoreSettings
}

func RenderContent(theme Theme, state ContentState) string {
	switch state.View {
	case "default":
		return renderDefaultView(theme, state)
	case "daily":
		return renderDailyView(theme, state)
	case "meta":
		return renderMetaView(theme, state)
	case "session_history", "session_active":
		return renderSessionView(theme, state)
	case "scratchpads":
		return renderScratchpadView(theme, state)
	case "ops":
		return renderOpsView(theme, state)
	case "wellbeing":
		return renderWellbeingView(theme, state)
	case "config":
		return renderConfigView(theme, state)
	case "reports":
		return renderReportsView(theme, state)
	case "settings":
		return renderSettingsView(theme, state)
	case "updates":
		return renderUpdatesView(theme, state)
	default:
		return ""
	}
}
