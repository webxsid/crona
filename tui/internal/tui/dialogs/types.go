package dialogs

import (
	"crona/tui/internal/logger"
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
	matched := ""
	var rendered string
	switch state.Kind {
	case "create_repo", "edit_repo", "create_stream", "edit_stream", "create_habit", "edit_habit", "checkout_context":
		matched = "repo_stream"
		rendered = renderRepoStreamDialog(theme, state)
	case "create_issue_meta", "create_issue_default", "edit_issue", "issue_status", "issue_status_note":
		matched = "issues"
		rendered = renderIssueDialog(theme, state)
	case "end_session", "stash_session", "issue_session_transition", "stash_list", "amend_session", "manual_session":
		matched = "session"
		rendered = renderSessionDialog(theme, state)
	case "edit_habit_streaks":
		matched = "habit_streaks"
		rendered = renderHabitStreakDialog(theme, state)
	case "confirm_delete", "confirm_wipe", "confirm_uninstall", "pick_date", "create_scratchpad", "create_checkin", "edit_checkin", "export_report_category", "export_report", "export_preset", "export_calendar_repo", "edit_export_reports_dir", "edit_export_ics_dir", "edit_date_display_format", "edit_rest_protection", "create_alert_reminder", "edit_alert_reminder", "view_entity", "support_bundle_result", "complete_habit", "view_jump", "beta_support", "stash_conflict_pick", "stash_conflict":
		matched = "utility"
		rendered = renderUtilityDialog(theme, state)
	case "edit_telemetry_settings", "onboarding":
		matched = "utility"
		rendered = renderUtilityDialog(theme, state)
	default:
		matched = "unhandled"
	}
	logger.Infof("dialogs.Render kind=%q matched=%q width=%d rendered_len=%d", state.Kind, matched, state.Width, len(rendered))
	return rendered
}
