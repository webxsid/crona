package dialogs

import (
	"fmt"
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

func renderDayBoundaryDialog(theme Theme, state controllerpkg.State) string {
	title := "Start of Day Schedule"
	if state.SettingKey == "endOfDay" {
		title = "End of Day Schedule"
	}
	progress := []string{
		stepLabel(theme, state, controllerpkg.DayBoundaryStepOverview, "Default"),
		stepLabel(theme, state, controllerpkg.DayBoundaryStepDays, "Add Override"),
		stepLabel(theme, state, controllerpkg.DayBoundaryStepTime, "Time"),
	}
	rows := []string{
		theme.StylePaneTitle.Render(title),
		"",
		strings.Join(progress, "   "),
		"",
		theme.StyleDim.Render("Timezone: " + fallbackTimezone(state.DayBoundaryTimezone)),
	}

	switch state.DayBoundaryStep {
	case controllerpkg.DayBoundaryStepDays:
		rows = append(rows, renderDayBoundaryDays(theme, state)...)
	case controllerpkg.DayBoundaryStepTime:
		rows = append(rows,
			theme.StyleDim.Render("Override days"),
			theme.StyleHeader.Render(dayBoundaryDaysLabel(state.DayBoundarySelectedDays)),
			"",
			theme.StyleDim.Render("Time (HH:mm)"),
			state.Inputs[1].View(),
		)
		rows = appendDialogFooter(theme, state, rows, "[ctrl+s] apply   [esc] back")
	default:
		rows = append(rows, renderDayBoundaryOverview(theme, state)...)
	}
	return modal(theme, state.Width, 78, theme.ColorCyan, rows)
}

func renderDayBoundaryOverview(theme Theme, state controllerpkg.State) []string {
	allOverridden := controllerpkg.AllDayBoundaryDaysOverridden(state.DayBoundarySchedule)
	defaultStyle := theme.StyleNormal
	if allOverridden {
		defaultStyle = theme.StyleDim
	}
	defaultRow := "Default time"
	if state.DayBoundaryOverrideCursor == 0 {
		defaultRow = viewchrome.SelectionCursor + " " + defaultRow
	}
	rows := []string{
		theme.StyleDim.Render("Enabled: " + enabledLabel(state.DayBoundaryEnabled) + "  [space] toggle"),
		defaultStyle.Render(defaultRow),
	}
	if allOverridden {
		rows = append(rows, theme.StyleDim.Render("  "+state.Inputs[0].View()+"  (all days have overrides)"))
	} else {
		rows = append(rows, defaultStyle.Render("  "+state.Inputs[0].View()))
	}
	rows = append(rows, "", theme.StyleDim.Render("Configured overrides"))
	if len(state.DayBoundaryOverrides) == 0 {
		rows = append(rows, theme.StyleDim.Render("  None"))
	} else {
		for i, override := range state.DayBoundaryOverrides {
			line := fmt.Sprintf("%-22s %s", dayBoundaryDaysLabel(override.Days), override.Time)
			if i+1 == state.DayBoundaryOverrideCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
	}
	addHint := "[a] add override"
	if allOverridden {
		addHint = "[a] add override (unavailable)"
	}
	rows = appendDialogFooter(theme, state, rows,
		addHint+"   [enter] edit   [D] remove   [ctrl+s] save   [esc] cancel")
	return rows
}

func renderDayBoundaryDays(theme Theme, state controllerpkg.State) []string {
	rows := []string{theme.StyleDim.Render("Select available days (space to toggle)")}
	for day, label := range []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"} {
		day++
		available := controllerpkg.DayBoundaryDayAvailable(state, day)
		selected := containsInt(state.DayBoundarySelectedDays, day)
		line := "[ ] " + label
		if selected {
			line = "[x] " + label
		}
		style := theme.StyleNormal
		if !available {
			style = theme.StyleDim
			line += " (configured)"
		}
		if day == state.FocusIdx && available {
			rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
		} else {
			rows = append(rows, style.Render("  "+line))
		}
	}
	rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [space] toggle   [enter] next   [esc] back")
	return rows
}

func stepLabel(theme Theme, state controllerpkg.State, step int, label string) string {
	value := fmt.Sprintf("%d.%s", step+1, label)
	if state.DayBoundaryStep == step {
		return theme.StyleCursor.Render(value)
	}
	return theme.StyleDim.Render(value)
}

func dayBoundaryDaysLabel(days []int) string {
	labels := make([]string, 0, len(days))
	for _, day := range days {
		if day >= 1 && day <= 7 {
			labels = append(labels, []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[day-1])
		}
	}
	if len(labels) == 0 {
		return "None"
	}
	return strings.Join(labels, ", ")
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func fallbackTimezone(value string) string {
	if value == "" {
		return "daemon local time"
	}
	return value
}
