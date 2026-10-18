package dialogs

import (
	"fmt"
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
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
		rows := []string{
			theme.StylePaneTitle.Render("Custom Date Format"),
			"",
			theme.StyleDim.Render("Moment-style pattern"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Examples: YYYY-MM-DD   Do MMM YYYY   MM/DD/YYYY   [Week] W"),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
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
	case "edit_frozen_streaks":
		rows := []string{
			theme.StylePaneTitle.Render("Frozen Streaks"),
			"",
			theme.StyleDim.Render("Comma-separated streak kinds"),
			dialogInputView(state, 0),
			"",
			theme.StyleDim.Render("Use: focus_days,checkin_days"),
			theme.StyleDim.Render("Leave blank to freeze all available streaks"),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_rest_weekdays":
		rows := []string{
			theme.StylePaneTitle.Render("Rest Weekdays"),
			"",
			theme.StyleDim.Render("Comma-separated weekdays"),
			dialogInputView(state, 0),
			"",
			theme.StyleDim.Render("Use: sun,mon,tue,wed,thu,fri,sat"),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_rest_dates":
		rows := []string{
			theme.StylePaneTitle.Render("Rest Dates"),
			"",
			theme.StyleDim.Render("Comma-separated one-off dates"),
			dialogInputView(state, 0),
			"",
			theme.StyleDim.Render("Use YYYY-MM-DD"),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_recurring_rest_dates":
		rows := []string{
			theme.StylePaneTitle.Render("Recurring Rest Dates"),
			"",
			theme.StyleDim.Render("Comma-separated recurring dates"),
			dialogInputView(state, 0),
			"",
			theme.StyleDim.Render("Use MM-DD"),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
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
func renderTelemetrySettingsDialog(theme Theme, state controllerpkg.State) string {
	steps := []string{"Usage", "Diagnostics", "Review"}
	progress := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf("%d.%s", i+1, step)
		if i == state.TelemetryStep {
			progress = append(progress, theme.StyleCursor.Render(label))
			continue
		}
		progress = append(progress, theme.StyleDim.Render(label))
	}
	rows := []string{
		theme.StylePaneTitle.Render("Privacy & Diagnostics"),
		"",
		strings.Join(progress, "   "),
		"",
	}
	switch state.TelemetryStep {
	case 0:
		rows = append(
			rows,
			theme.StyleDim.Render(
				"Share anonymous usage signals so Crona can spot problems early and understand which parts of the app people use.",
			),
			theme.StyleDim.Render(
				"This does not include your work content, paths, notes, or anything you type into Crona.",
			),
			"",
			theme.StyleCursor.Render(
				viewchrome.SelectionCursor+" "+toggleLabel(
					state.TelemetryUsage,
					"Share usage signals",
				),
			),
		)
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[space/enter] toggle   [tab] next   [esc] cancel",
		)
	case 1:
		rows = append(
			rows,
			theme.StyleDim.Render(
				"Share sanitized diagnostic reports when Crona runs into an error it can explain or recover from.",
			),
			theme.StyleDim.Render(
				"The report is stripped down to help us debug the problem without collecting your work content.",
			),
			"",
			theme.StyleCursor.Render(
				viewchrome.SelectionCursor+" "+toggleLabel(
					state.TelemetryErrors,
					"Share diagnostics",
				),
			),
		)
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[space/enter] toggle   [tab] next   [shift+tab] back   [esc] cancel",
		)
	default:
		rows = append(
			rows,
			theme.StyleDim.Render("Usage signals"),
			theme.StyleHeader.Render(telemetryStateLabel(state.TelemetryUsage)),
			"",
			theme.StyleDim.Render("Diagnostics"),
			theme.StyleHeader.Render(telemetryStateLabel(state.TelemetryErrors)),
			"",
			theme.StyleDim.Render(
				"Usage signals help us understand the app and catch regressions sooner.",
			),
			theme.StyleDim.Render(
				"Diagnostics help us investigate failures without your work content.",
			),
			"",
			theme.StyleError.Render("Changes take effect after restart."),
			theme.StyleDim.Render("Save now, or restart immediately to apply them right away."),
			"",
		)
		items := []string{"Save", "Save and Restart Now", "Cancel"}
		for idx, item := range items {
			line := "  " + item
			if idx == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[↑/↓] move   [enter] choose   [shift+tab] back   [esc] cancel",
		)
	}
	return modal(theme, state.Width, 82, theme.ColorCyan, rows)
}

func toggleLabel(enabled bool, label string) string {
	prefix := "[ ] "
	if enabled {
		prefix = "[x] "
	}
	return prefix + label
}

func telemetryStateLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
