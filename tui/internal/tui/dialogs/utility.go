package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"

	"github.com/charmbracelet/lipgloss"
)

func renderUtilityDialog(theme Theme, state controllerpkg.State) string {
	displayDate := helperpkg.FormatDisplayDate(state.CheckInDate, nil)
	switch state.Kind {
	case "confirm_delete":
		rows := []string{theme.StylePaneTitle.Render("Confirm Delete"), "", theme.StyleError.Render(fallback(state.DeleteLabel, "this item"))}
		rows = appendDialogFooter(theme, state, rows, "[enter] delete   [esc] cancel")
		return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorRed).Padding(1, 3).Width(min(state.Width-8, 44)).Render(strings.Join(rows, "\n"))
	case "confirm_wipe":
		rows := []string{
			theme.StylePaneTitle.Render("Wipe All Data"),
			"",
			theme.StyleError.Render("This deletes all Crona runtime data."),
			theme.StyleDim.Render("Issues, sessions, habits, reports, scratchpads, and settings will be reset."),
			theme.StyleDim.Render("Installed binaries will remain untouched."),
		}
		rows = appendDialogFooter(theme, state, rows, "[enter] wipe data   [esc] cancel")
		return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorRed).Padding(1, 3).Width(min(state.Width-8, 68)).Render(strings.Join(rows, "\n"))
	case "confirm_uninstall":
		rows := []string{
			theme.StylePaneTitle.Render("Uninstall Crona"),
			"",
			theme.StyleError.Render("This removes all Crona runtime data and installed binaries."),
			theme.StyleDim.Render("The CLI, TUI, and local engine binaries in the install directory will be deleted."),
			theme.StyleDim.Render("This action closes the app immediately after uninstall starts."),
		}
		rows = appendDialogFooter(theme, state, rows, "[enter] uninstall   [esc] cancel")
		return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorRed).Padding(1, 3).Width(min(state.Width-8, 74)).Render(strings.Join(rows, "\n"))
	case "pick_date":
		rows := []string{theme.StylePaneTitle.Render(state.DateTitle), "", theme.StyleHeader.Render(state.DateHeader), theme.StyleDim.Render(state.DateMonth), "", state.DateGrid}
		rows = appendDialogFooter(theme, state, rows, "[h/j/k/l] move   [,/.] month   [enter] choose   [c] clear   [esc] back")
		return modal(theme, state.Width, 46, theme.ColorCyan, rows)
	case "create_scratchpad":
		rows := []string{theme.StylePaneTitle.Render("New Scratchpad"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Path  (supports [[date]], [[timestamp]])"), state.Inputs[1].View()}
		rows = appendDialogFooter(theme, state, rows, "[tab] next field   "+dialogSubmitHint(state, "create")+"   [esc] cancel")
		return modal(theme, state.Width, 54, theme.ColorCyan, rows)
	case "create_checkin", "edit_checkin":
		title := "New Check-In"
		border := theme.ColorCyan
		hint := "[tab] next field   " + dialogSubmitHint(state, "save") + "   [esc] cancel"
		if state.Kind == "edit_checkin" {
			title = "Edit Check-In"
			border = theme.ColorYellow
		}
		rows := []string{
			theme.StylePaneTitle.Render(title),
			"",
			theme.StyleDim.Render("Date"),
			theme.StyleHeader.Render(displayDate),
			"",
			theme.StyleDim.Render("Mood"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Energy"),
			state.Inputs[1].View(),
			"",
			theme.StyleDim.Render("Sleep Duration"),
			state.Inputs[2].View(),
			"",
			theme.StyleDim.Render("Sleep Score"),
			state.Inputs[3].View(),
			"",
			theme.StyleDim.Render("Screen Time"),
			state.Inputs[4].View(),
			"",
			theme.StyleDim.Render("Notes"),
			state.Inputs[5].View(),
			"",
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return modal(theme, state.Width, 68, border, rows)
	case "export_report":
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
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = append(rows, "")
		if state.Processing {
			rows = appendDialogFooter(theme, state, append(rows, theme.StyleHeader.Render(state.ProcessingLabel), "", theme.StyleDim.Render("Please wait...")), "")
		} else {
			rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] choose   [esc] back")
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
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render("  "+item))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] choose category   [esc] cancel")
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
	case "export_preset":
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
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(rows, "", theme.StyleDim.Render("Preview"), renderViewEntityBody(theme, state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] use style   [esc] back")
		return modal(theme, state.Width, 74, theme.ColorGreen, rows)
	case "export_calendar_repo":
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
				line = "▶ " + item
				rows = append(rows, theme.StyleCursor.Render(line))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] export   [esc] back")
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
	case "stash_conflict_pick":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(state.ViewName),
		}
		if strings.TrimSpace(state.ViewMeta) != "" {
			rows = append(rows, theme.StyleDim.Render(state.ViewMeta))
		}
		rows = append(rows, "")
		for i, item := range state.ChoiceItems {
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+item))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] inspect stash   [esc] cancel")
		return modal(theme, state.Width, 72, theme.ColorYellow, rows)
	case "stash_conflict":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(state.ViewName),
		}
		if strings.TrimSpace(state.ViewMeta) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewMeta))
		}
		if strings.TrimSpace(state.ViewBody) != "" {
			rows = append(rows, theme.StyleDim.Render(state.ViewBody))
		}
		rows = append(rows, "")
		for i, item := range state.ChoiceItems {
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+item))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [r] resume   [c] continue fresh   [esc] cancel")
		return modal(theme, state.Width, 74, theme.ColorYellow, rows)
	case "edit_export_reports_dir":
		rows := []string{
			theme.StylePaneTitle.Render("Export Reports Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_export_ics_dir":
		rows := []string{
			theme.StylePaneTitle.Render("ICS Export Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
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
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_rest_protection":
		return renderRestProtectionDialog(theme, state)
	case "edit_habit_streaks":
		return renderHabitStreakDialog(theme, state)
	case "create_alert_reminder", "edit_alert_reminder":
		title := "Add Check-In Reminder"
		border := theme.ColorCyan
		if state.Kind == "edit_alert_reminder" {
			title = "Edit Check-In Reminder"
			border = theme.ColorYellow
		}
		rows := []string{
			theme.StylePaneTitle.Render(title),
			"",
			theme.StyleDim.Render("Schedule"),
			dialogInputView(state, 0),
			"",
			theme.StyleDim.Render("Time"),
			dialogInputView(state, 1),
			"",
			theme.StyleDim.Render("Use daily, weekdays, or weekday lists like mon,wed,fri."),
			theme.StyleDim.Render("Time uses 24-hour HH:MM, for example 20:00."),
		}
		rows = appendDialogFooter(theme, state, rows, "[tab] next field   "+dialogSubmitHint(state, "save")+"   [esc] cancel")
		return modal(theme, state.Width, 72, border, rows)
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
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
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
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
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
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
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
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [esc] cancel")
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "view_entity":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
		}
		if state.ViewMeta != "" {
			rows = append(rows, renderViewMeta(theme, state.ViewMeta)...)
		}
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody))
		footer := "[enter/esc] close"
		if strings.TrimSpace(state.ViewPath) != "" {
			footer = "[e] edit   " + footer
		}
		rows = appendDialogFooter(theme, state, rows, footer)
		return modal(theme, state.Width, 76, theme.ColorCyan, rows)
	case "support_bundle_result":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
		}
		if state.ViewMeta != "" {
			rows = append(rows, renderViewMeta(theme, state.ViewMeta)...)
		}
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody))
		rows = appendDialogFooter(theme, state, rows, "[o] open folder   [c] copy path   [g] report issue   [enter/esc] close")
		return modal(theme, state.Width, 76, theme.ColorGreen, rows)
	case "view_jump", "beta_support":
		rows := []string{
			theme.StylePaneTitle.Render(state.ViewTitle),
			"",
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
			"",
		}
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+item))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) && strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		var footer string
		if state.Kind == "view_jump" {
			footer = "[key] jump   [j/k] move   [enter] jump   [esc] cancel"
		} else {
			footer = "[key] run   [j/k] move   [enter] run   [esc] cancel"
		}
		rows = appendDialogFooter(theme, state, rows, footer)
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
	case "complete_habit":
		title := "Habit Log"
		if strings.TrimSpace(state.ViewTitle) != "" {
			title = state.ViewTitle
		}
		rows := []string{
			theme.StylePaneTitle.Render(title),
			"",
			theme.StyleDim.Render("Date"),
			theme.StyleHeader.Render(state.CheckInDate),
			"",
			theme.StyleDim.Render("Duration (Optional)"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Notes (Optional)"),
			state.Description.View(),
			"",
		}
		hint := "[tab] next   " + dialogSubmitHint(state, "save") + "   [esc] cancel"
		if state.FocusIdx == state.DescriptionIndex {
			hint = "[enter] newline in notes   [tab] next   " + dialogSubmitHint(state, "save") + "   [esc] cancel"
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	default:
		return ""
	}
}
