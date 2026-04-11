package dialogs

import (
	sharedtypes "crona/shared/types"

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
	Kind                string
	Width               int
	Inputs              []textinput.Model
	Description         textarea.Model
	DescriptionEnabled  bool
	DescriptionIndex    int
	FocusIdx            int
	ErrorMessage        string
	DeleteID            string
	DeleteKind          string
	DeleteLabel         string
	SessionID           string
	IssueID             int64
	HabitID             int64
	TargetView          string
	StashCursor         int
	Stashes             []StashItem
	RepoID              int64
	StreamID            int64
	StatusItems         []sharedtypes.IssueStatus
	StatusCursor        int
	ChoiceItems         []string
	ChoiceValues        []string
	ChoiceDetails       []string
	TemplateAssets      []sharedtypes.ExportTemplateAsset
	ChoiceCursor        int
	Processing          bool
	ProcessingLabel     string
	StatusLabel         string
	StatusRequired      bool
	IssueStatus         string
	CheckInDate         string
	RepoName            string
	RepoItems           []string
	RepoItemIDs         []int64
	StreamName          string
	RepoIndex           int
	StreamIndex         int
	Parent              string
	DateMonthValue      string
	DateCursorValue     string
	RepoSelectorLabel   string
	StreamSelectorLabel string
	ViewTitle           string
	ViewName            string
	IssueEstimateMins   *int
	ReminderID          string
	ReminderKind        sharedtypes.AlertReminderKind
	ViewMeta            string
	ViewBody            string
	SupportBundlePath   string
	DateTitle           string
	DateHeader          string
	DateMonth           string
	DateGrid            string
	ProtectionStep      int
	ProtectionCursor    int
	ProtectionStreaks   []sharedtypes.StreakKind
	ProtectionWeekdays  []int
	ProtectionDates     []string
	ExportPresetKind    sharedtypes.ExportReportKind
	ExportPresetFormat  sharedtypes.ExportFormat
	ExportPresetOutput  sharedtypes.ExportOutputMode
	ExportIncludePDF    bool
	ExportCategory      string
}

type StashItem struct {
	Label string
	Meta  string
}

func Render(theme Theme, state State) string {
	switch state.Kind {
	case "create_repo", "edit_repo", "create_stream", "edit_stream", "create_habit", "edit_habit", "checkout_context":
		return renderRepoStreamDialog(theme, state)
	case "create_issue_meta", "create_issue_default", "edit_issue", "issue_status", "issue_status_note":
		return renderIssueDialog(theme, state)
	case "end_session", "stash_session", "issue_session_transition", "stash_list", "amend_session", "manual_session":
		return renderSessionDialog(theme, state)
	case "confirm_delete", "confirm_wipe", "confirm_uninstall", "pick_date", "create_scratchpad", "create_checkin", "edit_checkin", "export_report_category", "export_report", "export_preset", "export_calendar_repo", "edit_export_reports_dir", "edit_export_ics_dir", "edit_rest_protection", "create_alert_reminder", "edit_alert_reminder", "view_entity", "support_bundle_result", "complete_habit", "view_jump", "beta_support", "stash_conflict_pick", "stash_conflict":
		return renderUtilityDialog(theme, state)
	default:
		return ""
	}
}
