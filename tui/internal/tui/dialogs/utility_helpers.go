package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/bubbles/textinput"
)

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

func renderRestProtectionDialog(theme Theme, state controllerpkg.State) string {
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

func dialogInputView(state controllerpkg.State, idx int) string {
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
