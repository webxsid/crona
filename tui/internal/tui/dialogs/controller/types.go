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
	ColorDim    lipgloss.Color

	StylePaneTitle lipgloss.Style
	StyleDim       lipgloss.Style
	StyleCursor    lipgloss.Style
	StyleHeader    lipgloss.Style
	StyleError     lipgloss.Style
	StyleSelected  lipgloss.Style
	StyleNormal    lipgloss.Style
}

type SessionHistorySegmentRow struct {
	Type      sharedtypes.SessionSegmentType
	StartTime textinput.Model
	EndTime   textinput.Model
}

type State struct {
	Kind                     string
	Width                    int
	Height                   int
	Inputs                   []textinput.Model
	Description              textarea.Model
	DescriptionEnabled       bool
	DescriptionIndex         int
	FocusIdx                 int
	ErrorMessage             string
	DeleteID                 string
	DeleteKind               string
	DeleteLabel              string
	SessionID                string
	SessionHistoryContext    string
	SessionHistoryNoteInputs []textarea.Model
	SessionHistorySegments   []SessionHistorySegmentRow
	IssueID                  int64
	AdhocTimer               bool
	PromotingAdhoc           bool
	ViewEntityAdhoc          bool
	AdhocResolution          sharedtypes.AdhocSessionResolution
	AdhocIssueIDs            []int64
	AdhocIssueTab            int
	AdhocIssueTabCounts      [4]int
	AdhocElapsedSeconds      int
	AdhocDate                string
	NewIssueTitle            string
	HabitID                  int64
	TargetView               string
	RepoID                   int64
	StreamID                 int64
	StatusItems              []sharedtypes.IssueStatus
	StatusCursor             int
	ChoiceItems              []string
	ChoiceValues             []string
	ChoiceDetails            []string
	TemplateAssets           []sharedtypes.ExportTemplateAsset
	ChoiceCursor             int
	MomentumRepoInput        textinput.Model
	MomentumStreamInput      textinput.Model
	Processing               bool
	ProcessingLabel          string
	StatusLabel              string
	StatusRequired           bool
	IssueStatus              string
	CheckInDate              string
	RepoName                 string
	RepoItems                []string
	RepoItemIDs              []int64
	StreamName               string
	RepoIndex                int
	StreamIndex              int
	Parent                   string
	DateMonthValue           string
	// PenaltyHint shows what rescheduling the issue would do to its delay penalty.
	PenaltyHint         string
	DateCursorValue     string
	DueDateToday        string
	DueDateRestWeekdays []int
	DueDateRestDates    []string
	DueDateAwayDates    []string
	RepoSelectorLabel   string
	RepoOptionLabels    []string
	StreamOptionLabels  []string
	StreamSelectorLabel string
	ViewTitle           string
	ViewName            string
	IssueEstimateMins   *int
	IssueWorkedSeconds  int
	ReminderID          string
	ReminderKind        sharedtypes.AlertReminderKind
	ViewMeta            string
	ViewBody            string
	ViewPath            string
	SupportBundlePath   string
	DateTitle           string
	DateHeader          string
	DateMonth           string
	DateGrid            string
	CalendarScores      map[string]api.FocusScoreRangeDay
	ProtectionStep      int
	ProtectionCursor    int
	ProtectionStreaks   []sharedtypes.StreakKind
	ProtectionWeekdays  []int
	// CadenceIdx is the cadence picker choice (see CadenceLabels); CadenceDays the custom
	// weekdays (0 = Sunday) and CadenceCursor the day under the cursor, Monday first.
	CadenceIdx                     int
	CadenceDays                    []int
	CadenceCursor                  int
	ProtectionDates                []string
	TelemetryStep                  int
	TelemetryUsage                 bool
	TelemetryErrors                bool
	TelemetryPrivacyCursor         int
	TelemetryReviewCursor          int
	DayBoundaryEnabled             bool
	DayBoundaryTimezone            string
	SettingKey                     sharedtypes.CoreSettingsKey
	DayBoundaryStep                int
	DayBoundarySchedule            sharedtypes.DayBoundarySchedule
	DayBoundaryOverrides           []DayBoundaryOverride
	DayBoundaryOverrideCursor      int
	DayBoundarySelectedDays        []int
	DayBoundaryEditingTime         string
	DayBoundaryEditingOverride     bool
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
	ExportStart                    string
	ExportEnd                      string
	PromptGlyphMode                sharedtypes.PromptGlyphMode
	PomodoroFocusSeconds           int
	PomodoroFocusChoice            int
	PomodoroBreakSeconds           int
	PomodoroBreakChoice            int
	PomodoroLongBreakSeconds       int
	PomodoroLongBreakChoice        int
	PomodoroCyclesBeforeLongBreak  int
	PomodoroCycles                 int
	TimerCountdownSeconds          int
	HardLimitKind                  sharedtypes.TimerHardLimitKind
	HardLimitTotalSeconds          int
	HardLimitFocusSeconds          int
	HardLimitBreakSeconds          int
	HardLimitLongBreakSeconds      int
	HardLimitCyclesBeforeLongBreak int
}

const (
	IssueCreateFollowUpDefault = iota
	IssueCreateFollowUpMore
	IssueCreateFollowUpFocus
	IssueCreateFollowUpAdhoc
)
