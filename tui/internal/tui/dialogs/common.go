package dialogs

import (
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	dialogSearchPromptEmoji   = "🔎 "
	dialogTimePromptEmoji     = "🕒 "
	dialogDatePromptEmoji     = "📅 "
	dialogSearchPromptUnicode = "⌕ "
	dialogTimePromptUnicode   = "◷ "
	dialogDatePromptUnicode   = "◫ "
	dialogSearchPromptASCII   = "? "
	dialogTimePromptASCII     = "t "
	dialogDatePromptASCII     = "d "
	dialogValuePrompt         = "> "
)

func modal(theme Theme, width, maxWidth int, border lipgloss.Color, rows []string) string {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(border).Padding(1, 3).Width(min(width-8, maxWidth)).Render(strings.Join(rows, "\n"))
}

func renderSelector(theme Theme, state State, label string, active bool) string {
	style := theme.StyleNormal
	if active {
		style = theme.StyleCursor
	}
	return style.Render("[ " + promptGlyphSet(state).Value + label + " ]")
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

func withDialogPrompt(input textinput.Model, prompt string) textinput.Model {
	input.Prompt = prompt
	return input
}

type dialogPromptSet struct {
	Search string
	Value  string
	Time   string
	Date   string
}

func promptGlyphSet(state State) dialogPromptSet {
	switch sharedtypes.NormalizePromptGlyphMode(state.PromptGlyphMode) {
	case sharedtypes.PromptGlyphModeUnicode:
		return dialogPromptSet{
			Search: dialogSearchPromptUnicode,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptUnicode,
			Date:   dialogDatePromptUnicode,
		}
	case sharedtypes.PromptGlyphModeASCII:
		return dialogPromptSet{
			Search: dialogSearchPromptASCII,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptASCII,
			Date:   dialogDatePromptASCII,
		}
	default:
		return dialogPromptSet{
			Search: dialogSearchPromptEmoji,
			Value:  dialogValuePrompt,
			Time:   dialogTimePromptEmoji,
			Date:   dialogDatePromptEmoji,
		}
	}
}

func withSearchPrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Search)
}

func withTimePrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Time)
}

func withDatePrompt(state State, input textinput.Model) textinput.Model {
	return withDialogPrompt(input, promptGlyphSet(state).Date)
}
