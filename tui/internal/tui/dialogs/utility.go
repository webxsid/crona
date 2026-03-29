package dialogs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderUtilityDialog(theme Theme, state State) string {
	switch state.Kind {
	case "confirm_delete":
		content := fmt.Sprintf("%s\n\nDelete %s?\n\n%s", theme.StylePaneTitle.Render("Confirm Delete"), theme.StyleError.Render(fallback(state.DeleteLabel, "this item")), theme.StyleDim.Render("[enter] delete   [esc] cancel"))
		return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.ColorRed).Padding(1, 3).Width(min(state.Width-8, 44)).Render(content)
	case "pick_date":
		rows := []string{theme.StylePaneTitle.Render(state.DateTitle), "", theme.StyleHeader.Render(state.DateHeader), theme.StyleDim.Render(state.DateMonth), "", state.DateGrid, "", theme.StyleDim.Render("[h/j/k/l] move   [,/.] month   [enter] choose   [c] clear   [esc] back")}
		return modal(theme, state.Width, 46, theme.ColorCyan, rows)
	case "create_scratchpad":
		rows := []string{theme.StylePaneTitle.Render("New Scratchpad"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Path  (supports [[date]], [[timestamp]])"), state.Inputs[1].View(), "", theme.StyleDim.Render("[tab] next field   [ctrl+s] create   [esc] cancel")}
		return modal(theme, state.Width, 54, theme.ColorCyan, rows)
	case "create_checkin", "edit_checkin":
		title := "New Check-In"
		border := theme.ColorCyan
		hint := "[tab] next field   [ctrl+s] save   [esc] cancel"
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
			theme.StyleDim.Render("Sleep Hours"),
			state.Inputs[2].View(),
			"",
			theme.StyleDim.Render("Sleep Score"),
			state.Inputs[3].View(),
			"",
			theme.StyleDim.Render("Screen Time Minutes"),
			state.Inputs[4].View(),
			"",
			theme.StyleDim.Render("Notes"),
			state.Inputs[5].View(),
			"",
			theme.StyleDim.Render(hint),
		}
		return modal(theme, state.Width, 68, border, rows)
	case "export_report":
		rows := []string{
			theme.StylePaneTitle.Render("Export Report"),
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
				line = "▶ " + item
				rows = append(rows, theme.StyleCursor.Render(line))
				continue
			}
			rows = append(rows, theme.StyleNormal.Render(line))
		}
		rows = append(rows, "")
		if state.Processing {
			rows = append(rows, theme.StyleHeader.Render(state.ProcessingLabel))
			rows = append(rows, "")
			rows = append(rows, theme.StyleDim.Render("Please wait..."))
		} else {
			rows = append(rows, theme.StyleDim.Render("[j/k] move   [enter] choose   [esc] cancel"))
		}
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
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
		rows = append(rows, "", theme.StyleDim.Render("[j/k] move   [enter] export   [esc] back"))
		return modal(theme, state.Width, 54, theme.ColorGreen, rows)
	case "edit_export_reports_dir":
		rows := []string{
			theme.StylePaneTitle.Render("Export Reports Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_export_ics_dir":
		rows := []string{
			theme.StylePaneTitle.Render("ICS Export Directory"),
			"",
			theme.StyleDim.Render("Path"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use an absolute path or ~/..."),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_frozen_streaks":
		rows := []string{
			theme.StylePaneTitle.Render("Frozen Streaks"),
			"",
			theme.StyleDim.Render("Comma-separated streak kinds"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use: focus_days,checkin_days"),
			theme.StyleDim.Render("Leave blank to freeze all available streaks"),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_rest_weekdays":
		rows := []string{
			theme.StylePaneTitle.Render("Rest Weekdays"),
			"",
			theme.StyleDim.Render("Comma-separated weekdays"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use: sun,mon,tue,wed,thu,fri,sat"),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_rest_dates":
		rows := []string{
			theme.StylePaneTitle.Render("Rest Dates"),
			"",
			theme.StyleDim.Render("Comma-separated one-off dates"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use YYYY-MM-DD"),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "edit_recurring_rest_dates":
		rows := []string{
			theme.StylePaneTitle.Render("Recurring Rest Dates"),
			"",
			theme.StyleDim.Render("Comma-separated recurring dates"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Use MM-DD"),
			theme.StyleDim.Render("[enter] save   [esc] cancel"),
		}
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
		rows = append(rows, "", renderViewEntityBody(theme, state.ViewBody), "", theme.StyleDim.Render("[enter/esc] close"))
		return modal(theme, state.Width, 76, theme.ColorCyan, rows)
	case "complete_habit":
		rows := []string{
			theme.StylePaneTitle.Render("Habit Log"),
			"",
			theme.StyleDim.Render("Date"),
			theme.StyleHeader.Render(state.CheckInDate),
			"",
			theme.StyleDim.Render("Duration Minutes (Optional)"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Notes (Optional)"),
			state.Description.View(),
			"",
			theme.StyleDim.Render("[enter] newline in notes   [ctrl+s] save   [tab] next   [esc] cancel"),
		}
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	default:
		return ""
	}
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
