package dialogs

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func modal(theme Theme, width, maxWidth int, border lipgloss.Color, rows []string) string {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(border).Padding(1, 3).Width(min(width-8, maxWidth)).Render(strings.Join(rows, "\n"))
}

func renderSingleInput(theme Theme, width int, title, label string, inputs []textinput.Model, border lipgloss.Color, hint string) string {
	rows := []string{theme.StylePaneTitle.Render(title), "", theme.StyleDim.Render(label), inputs[0].View()}
	rows = appendDialogFooter(theme, State{}, rows, hint)
	return modal(theme, width, 52, border, rows)
}

func renderSelector(theme Theme, label string, active bool) string {
	style := theme.StyleNormal
	if active {
		style = theme.StyleCursor
	}
	return style.Render("[ " + label + " ]")
}

func renderInputColumns(width, maxWidth int, left string, right string) string {
	contentWidth := min(width-8, maxWidth) - 8
	contentWidth = max(28, contentWidth)
	colWidth := (contentWidth - 2) / 2
	if colWidth < 12 {
		colWidth = 12
	}
	leftCol := lipgloss.NewStyle().Width(colWidth).Render(left)
	rightCol := lipgloss.NewStyle().Width(colWidth).Render(right)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol)
}

func plainIssueStatus(status string) string {
	switch status {
	case "in_progress":
		return "in progress"
	case "in_review":
		return "in review"
	default:
		return status
	}
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func appendDialogFooter(theme Theme, state State, rows []string, hint string) []string {
	if strings.TrimSpace(state.ErrorMessage) != "" {
		rows = append(rows, "", theme.StyleError.Render(state.ErrorMessage))
	}
	if strings.TrimSpace(hint) != "" {
		rows = append(rows, "", theme.StyleDim.Render(hint))
	}
	return rows
}

func dialogSubmitChord(state State) string {
	return "ctrl+s"
}

func dialogSubmitHint(state State, label string) string {
	return "[" + dialogSubmitChord(state) + "] " + label
}

func isDialogSubmitKey(state State, key string) bool {
	switch key {
	case "ctrl+s":
		return true
	default:
		return false
	}
}
