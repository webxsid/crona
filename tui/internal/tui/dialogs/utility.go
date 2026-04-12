package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func renderUtilityDialog(theme Theme, state State) string {
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
			theme.StyleHeader.Render(state.CheckInDate),
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
			theme.StyleHeader.Render(state.CheckInDate),
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
			theme.StyleHeader.Render(state.CheckInDate),
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
			theme.StyleHeader.Render(state.CheckInDate),
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
			theme.StyleHeader.Render(state.CheckInDate),
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
	case "edit_rest_protection":
		return renderRestProtectionDialog(theme, state)
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
		rows = appendDialogFooter(theme, state, rows, "[enter/esc] close")
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
		rows := []string{
			theme.StylePaneTitle.Render("Habit Log"),
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

func exportCategoryTitle(category string) string {
	switch category {
	case "project":
		return "Project Reports"
	case "data":
		return "Data Exports"
	default:
		return "Narrative Reports"
	}
}

func renderRestProtectionDialog(theme Theme, state State) string {
	steps := []string{"Streaks", "Weekdays", "Dates", "Review"}
	progress := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf("%d.%s", i+1, step)
		if i == state.ProtectionStep {
			progress = append(progress, theme.StyleCursor.Render(label))
			continue
		}
		progress = append(progress, theme.StyleDim.Render(label))
	}
	rows := []string{
		theme.StylePaneTitle.Render("Rest & Streak Protection"),
		"",
		strings.Join(progress, "   "),
		"",
	}
	switch state.ProtectionStep {
	case 0:
		rows = append(rows, theme.StyleDim.Render("Select which streaks are protected"))
		for i, kind := range sharedtypes.AvailableStreakKinds() {
			label := "Focus Days"
			if kind == sharedtypes.StreakKindCheckInDays {
				label = "Check-In Days"
			}
			prefix := "[ ] "
			if hasStreakKind(state.ProtectionStreaks, kind) {
				prefix = "[x] "
			}
			line := prefix + label
			if i == state.ProtectionCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [space] toggle   [a] all   [c] none   [tab] next")
	case 1:
		rows = append(rows, theme.StyleDim.Render("Select default rest weekdays"))
		labels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		for i, label := range labels {
			prefix := "[ ] "
			if containsInt(state.ProtectionWeekdays, i) {
				prefix = "[x] "
			}
			line := prefix + label
			if i == state.ProtectionCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [space] toggle   [c] clear   [tab] next")
	case 2:
		rows = append(rows, theme.StyleDim.Render("Manage one-off rest dates"))
		if len(state.ProtectionDates) == 0 {
			rows = append(rows, theme.StyleDim.Render("No rest dates added"))
		} else {
			for i, value := range state.ProtectionDates {
				line := value
				if i == state.ProtectionCursor {
					rows = append(rows, theme.StyleCursor.Render("▶ "+line))
				} else {
					rows = append(rows, theme.StyleNormal.Render("  "+line))
				}
			}
		}
		rows = appendDialogFooter(theme, state, rows, "[a] add date   [d] remove selected   [tab] next")
	case 3:
		rows = append(rows,
			theme.StyleDim.Render("Protected Streaks"),
			theme.StyleHeader.Render(fallback(streakKindsSummary(state.ProtectionStreaks), "None")),
			"",
			theme.StyleDim.Render("Rest Weekdays"),
			theme.StyleHeader.Render(weekdaysSummary(state.ProtectionWeekdays)),
			"",
			theme.StyleDim.Render("Rest Dates"),
			theme.StyleHeader.Render(restDatesSummary(state.ProtectionDates)),
		)
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [shift+tab] back   [esc] cancel")
	}
	return modal(theme, state.Width, 76, theme.ColorCyan, rows)
}

func hasStreakKind(values []sharedtypes.StreakKind, target sharedtypes.StreakKind) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func streakKindsSummary(values []sharedtypes.StreakKind) string {
	if len(values) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(values))
	for _, value := range values {
		if value == sharedtypes.StreakKindCheckInDays {
			labels = append(labels, "Check-ins")
		} else {
			labels = append(labels, "Focus")
		}
	}
	return strings.Join(labels, " • ")
}

func weekdaysSummary(values []int) string {
	if len(values) == 0 {
		return "None"
	}
	labels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value >= 0 && value < len(labels) {
			out = append(out, labels[value])
		}
	}
	if len(out) == 0 {
		return "None"
	}
	return strings.Join(out, ", ")
}

func restDatesSummary(values []string) string {
	if len(values) == 0 {
		return "None"
	}
	if len(values) == 1 {
		return values[0]
	}
	return fmt.Sprintf("%s +%d", values[0], len(values)-1)
}

func dialogInputView(state State, idx int) string {
	if idx >= 0 && idx < len(state.Inputs) {
		return state.Inputs[idx].View()
	}
	input := textinput.New()
	input.Width = 56
	input.Placeholder = "<missing input>"
	return input.View()
}

func renderViewEntityBody(theme Theme, body string) string {
	lines := strings.Split(body, "\n")
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "Description", "Notes":
			rendered = append(rendered, theme.StyleDim.Render(trimmed))
		default:
			rendered = append(rendered, line)
		}
	}
	return strings.Join(rendered, "\n")
}

func renderViewMeta(theme Theme, meta string) []string {
	parts := strings.Split(meta, "   ")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, " ")
		if !ok {
			lines = append(lines, theme.StyleDim.Render(part))
			continue
		}
		lines = append(lines, theme.StyleDim.Render(key)+": "+theme.StyleHeader.Render(strings.TrimSpace(value)))
	}
	return lines
}
