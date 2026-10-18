package dialogs

import (
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
)

func renderUtilityDialog(theme Theme, state controllerpkg.State) string {
	displayDate := helperpkg.FormatDisplayDate(state.CheckInDate, nil)
	switch state.Kind {
	case "confirm_delete":
		return renderConfirmDelete(theme, state)
	case "confirm_wipe":
		return renderConfirmWipe(theme, state)
	case "pick_date", "summary_calendar":
		return renderDateGrid(theme, state)
	case "create_checkin", "edit_checkin":
		return renderCheckInForm(theme, state, displayDate)
	case "export_report", "export_summary_period", "export_report_category", "export_preset",
		"export_summary_output", "export_calendar_repo", "export_scope":
		return renderExportChoice(theme, state)
	case "edit_export_reports_dir":
		return renderExportPathForm(theme, state, "Export Reports Directory")
	case "export_summary_range":
		return renderSummaryRangeForm(theme, state)
	case "edit_export_ics_dir":
		return renderExportPathForm(theme, state, "ICS Export Directory")
	case "edit_date_display_format":
		return renderDateFormatForm(theme, state)
	case "edit_day_boundary":
		return renderDayBoundaryDialog(theme, state)
	case "edit_rest_protection":
		return renderRestProtectionDialog(theme, state)
	case "edit_telemetry_settings":
		return renderTelemetrySettingsDialog(theme, state)
	case "onboarding":
		return renderOnboardingScreen(theme, state)
	case "create_alert_reminder", "edit_alert_reminder":
		return renderReminderForm(theme, state)
	case "view_entity", "view_momentum_detail", "support_bundle_result":
		return renderDetailView(theme, state)
	case "view_jump", "beta_support":
		return renderMenuList(theme, state)
	case "complete_habit":
		return renderHabitLogForm(theme, state)
	default:
		return ""
	}
}
func telemetryStateLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
