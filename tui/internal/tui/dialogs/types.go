package dialogs

import (
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

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

func Render(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "create_repo", "edit_repo", "create_stream", "edit_stream", "create_habit", "edit_habit", "checkout_context":
		return renderRepoStreamDialog(theme, state)
	case "create_issue_meta", "create_issue_default", "edit_issue", "issue_status", "issue_status_note":
		return renderIssueDialog(theme, state)
	case "end_session", "stash_session", "issue_session_transition", "stash_list", "amend_session", "manual_session":
		return renderSessionDialog(theme, state)
	case "edit_habit_streaks":
		return renderHabitStreakDialog(theme, state)
	case "confirm_delete", "confirm_wipe", "confirm_uninstall", "pick_date", "create_scratchpad", "create_checkin", "edit_checkin", "export_report_category", "export_report", "export_preset", "export_calendar_repo", "edit_export_reports_dir", "edit_export_ics_dir", "edit_date_display_format", "edit_rest_protection", "create_alert_reminder", "edit_alert_reminder", "view_entity", "support_bundle_result", "complete_habit", "view_jump", "beta_support", "stash_conflict_pick", "stash_conflict":
		return renderUtilityDialog(theme, state)
	default:
		return ""
	}
}
