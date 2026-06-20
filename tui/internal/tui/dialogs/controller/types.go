package controller

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ColorCyan   lipgloss.Color
	ColorYellow lipgloss.Color
	ColorRed    lipgloss.Color
	ColorGreen  lipgloss.Color

	StylePaneTitle lipgloss.Style
	StyleDim       lipgloss.Style
	StyleCursor    lipgloss.Style
	StyleHeader    lipgloss.Style
	StyleError     lipgloss.Style
	StyleSelected  lipgloss.Style
	StyleNormal    lipgloss.Style
}

type State struct {
	Kind                           string
	Width                          int
	Inputs                         []textinput.Model
	Description                    textarea.Model
	DescriptionEnabled             bool
	DescriptionIndex               int
	FocusIdx                       int
	ErrorMessage                   string
	DeleteID                       string
	DeleteKind                     string
	DeleteLabel                    string
	SessionID                      string
	IssueID                        int64
	HabitID                        int64
	TargetView                     string
	StashCursor                    int
	Stashes                        []StashItem
	RepoID                         int64
	StreamID                       int64
	StatusItems                    []sharedtypes.IssueStatus
	StatusCursor                   int
	ChoiceItems                    []string
	ChoiceValues                   []string
	ChoiceDetails                  []string
	TemplateAssets                 []sharedtypes.ExportTemplateAsset
	ChoiceCursor                   int
	MomentumRepoInput              textinput.Model
	MomentumStreamInput            textinput.Model
	Processing                     bool
	ProcessingLabel                string
	StatusLabel                    string
	StatusRequired                 bool
	IssueStatus                    string
	CheckInDate                    string
	RepoName                       string
	RepoItems                      []string
	RepoItemIDs                    []int64
	StreamName                     string
	RepoIndex                      int
	StreamIndex                    int
	Parent                         string
	DateMonthValue                 string
	DateCursorValue                string
	RepoSelectorLabel              string
	StreamSelectorLabel            string
	ViewTitle                      string
	ViewName                       string
	IssueEstimateMins              *int
	ReminderID                     string
	ReminderKind                   sharedtypes.AlertReminderKind
	ViewMeta                       string
	ViewBody                       string
	ViewPath                       string
	SupportBundlePath              string
	DateTitle                      string
	DateHeader                     string
	DateMonth                      string
	DateGrid                       string
	ProtectionStep                 int
	ProtectionCursor               int
	ProtectionStreaks              []sharedtypes.StreakKind
	ProtectionWeekdays             []int
	ProtectionDates                []string
	TelemetryStep                  int
	TelemetryUsage                 bool
	TelemetryErrors                bool
	TelemetryPrivacyCursor         int
	TelemetryReviewCursor          int
	HabitItems                     []sharedtypes.HabitWithMeta
	MomentumRepos                  []api.Repo
	MomentumStreams                []api.Stream
	MomentumAllIssues              []api.IssueWithMeta
	HabitStreakStep                int
	HabitStreakCursor              int
	HabitStreakOriginalDefs        []sharedtypes.HabitStreakDefinition
	HabitStreakDefs                []sharedtypes.HabitStreakDefinition
	HabitStreakDraft               sharedtypes.HabitStreakDefinition
	HabitStreakEditIdx             int
	ExportPresetKind               sharedtypes.ExportReportKind
	ExportPresetFormat             sharedtypes.ExportFormat
	ExportPresetOutput             sharedtypes.ExportOutputMode
	ExportIncludePDF               bool
	ExportCategory                 string
	PromptGlyphMode                sharedtypes.PromptGlyphMode
	PomodoroFocusSeconds           int
	PomodoroFocusChoice            int
	PomodoroBreakSeconds           int
	PomodoroBreakChoice            int
	PomodoroLongBreakSeconds       int
	PomodoroLongBreakChoice        int
	PomodoroCyclesBeforeLongBreak  int
	PomodoroCycles                 int
	HardLimitTotalSeconds          int
	HardLimitFocusSeconds          int
	HardLimitBreakSeconds          int
	HardLimitLongBreakSeconds      int
	HardLimitCyclesBeforeLongBreak int
}

type StashItem struct {
	Label string
	Meta  string
}
