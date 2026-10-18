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
	case "pick_date":
		rows := []string{
			theme.StylePaneTitle.Render(state.DateTitle),
			"",
			theme.StyleHeader.Render(state.DateHeader),
			theme.StyleDim.Render(state.DateMonth),
			"",
			state.DateGrid,
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[h/j/k/l] move   [,/.] month   [enter] choose   [c] clear   [esc] back",
		)
		return modal(theme, state.Width, 46, theme.ColorCyan, rows)
	case "summary_calendar":
		rows := []string{
			theme.StylePaneTitle.Render("Summary Calendar"),
			"",
			theme.StyleHeader.Render(helperpkg.FormatDisplayDate(state.DateCursorValue, nil)),
			theme.StyleDim.Render(state.DateMonth),
			"",
			state.DateGrid,
		}
		rows = appendDialogFooter(theme, state, rows, "[←/→/↑/↓] move   [,/.] month   [g] today   [enter] choose   [esc] back")
		return modal(theme, state.Width, 100, theme.ColorCyan, rows)
	case "create_checkin", "edit_checkin":
		return renderCheckInForm(theme, state, displayDate)
	case "export_report", "export_summary_period":
		rows := []string{
			theme.StylePaneTitle.Render("Export Report"),
			"",
			theme.StyleDim.Render("Category"),
			theme.StyleHeader.Render(exportCategoryTitle(state.ExportCategory)),
			"",
			theme.StyleDim.Render("Anchor Date"),
			theme.StyleHeader.Render(displayDate),
			"",
		}
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if state.Processing {
				rows = append(rows, theme.StyleDim.Render(line))
				continue
			}
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) &&
			strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = append(rows, "")
		if state.Processing {
			rows = appendDialogFooter(
				theme,
				state,
				append(
					rows,
					theme.StyleHeader.Render(state.ProcessingLabel),
					"",
					theme.StyleDim.Render("Please wait..."),
				),
				"",
			)
		} else {
			rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose   [esc] back")
		}
		return modal(theme, state.Width, 64, theme.ColorGreen, rows)
	case "export_report_category":
		rows := []string{
			theme.StylePaneTitle.Render("Export Report"),
			"",
			theme.StyleDim.Render("Anchor Date"),
			theme.StyleHeader.Render(displayDate),
			"",
		}
		for i, item := range state.ChoiceItems {
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render("  "+item))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) &&
			strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[↑/↓] move   [enter] choose category   [esc] cancel",
		)
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
	case "export_preset", "export_summary_output":
		rows := []string{
			theme.StylePaneTitle.Render("Choose Report Style"),
			"",
			theme.StyleDim.Render("Anchor Date"),
			theme.StyleHeader.Render(displayDate),
			"",
		}
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render("Preview"),
				renderViewEntityBody(theme, state.ChoiceDetails[state.ChoiceCursor]),
			)
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] use style   [esc] back")
		return modal(theme, state.Width, 74, theme.ColorGreen, rows)
	case "export_calendar_repo", "export_scope":
		rows := []string{
			theme.StylePaneTitle.Render("Calendar Export Repo"),
			"",
			theme.StyleDim.Render("Anchor Date"),
			theme.StyleHeader.Render(displayDate),
			"",
			theme.StyleDim.Render("Select Repo"),
		}
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if state.Processing {
				rows = append(rows, theme.StyleDim.Render(line))
				continue
			}
			if i == state.ChoiceCursor {
				line = viewchrome.SelectionCursor + " " + item
				rows = append(rows, theme.StyleCursor.Render(line))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] export   [esc] back")
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
	case "edit_export_reports_dir":
		rows := []string{
			theme.StylePaneTitle.Render("Export Reports Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "export_summary_range":
		rows := []string{
			theme.StylePaneTitle.Render("Custom Summary Range"),
			"",
			theme.StyleDim.Render("Start date"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("End date"),
			state.Inputs[1].View(),
		}
		rows = appendDialogFooter(theme, state, rows, "[tab] switch   [enter] continue   [esc] back")
		return modal(theme, state.Width, 58, theme.ColorGreen, rows)
	case "edit_export_ics_dir":
		rows := []string{
			theme.StylePaneTitle.Render("ICS Export Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
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
	case "view_entity":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
		}
		if state.ViewMeta != "" {
			rows = append(rows, renderViewMeta(theme, state.ViewMeta, modalTextWidth(state.Width, 76))...)
		}
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody))
		footer := "[enter/esc] close"
		if state.ViewEntityAdhoc {
			footer = "[e] convert to regular issue   " + footer
		}
		if strings.TrimSpace(state.ViewPath) != "" {
			footer = "[e] edit   " + footer
		}
		rows = appendDialogFooter(theme, state, rows, footer)
		return modal(theme, state.Width, 76, theme.ColorCyan, rows)
	case "view_momentum_detail":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
		}
		if state.ViewMeta != "" {
			rows = append(rows, renderViewMeta(theme, state.ViewMeta, modalTextWidth(state.Width, 88))...)
		}
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody))
		rows = append(rows, "", theme.StyleDim.Render("Contributors"))
		if len(state.ChoiceItems) == 0 {
			rows = append(
				rows,
				theme.StyleDim.Render("No contributing entries in the current bucket"),
			)
		} else {
			inner := 8
			start, end := viewchrome.ListWindow(state.ChoiceCursor, len(state.ChoiceItems), inner)
			if start > 0 {
				rows = append(rows, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
			}
			for i := start; i < end; i++ {
				line := state.ChoiceItems[i]
				if i == state.ChoiceCursor {
					rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
				} else {
					rows = append(rows, theme.StyleNormal.Render("  "+line))
				}
				if i < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[i]) != "" {
					rows = append(rows, theme.StyleDim.Render("    "+state.ChoiceDetails[i]))
				}
			}
			if remaining := len(state.ChoiceItems) - end; remaining > 0 {
				rows = append(rows, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
			}
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] scroll   [e] edit   [enter/esc] close")
		return modal(theme, state.Width, 88, theme.ColorCyan, rows)
	case "support_bundle_result":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
		}
		if state.ViewMeta != "" {
			rows = append(rows, renderViewMeta(theme, state.ViewMeta, modalTextWidth(state.Width, 76))...)
		}
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody))
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[o] open folder   [c] copy path   [g] report issue   [enter/esc] close",
		)
		return modal(theme, state.Width, 76, theme.ColorGreen, rows)
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
