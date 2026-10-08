package dialogs

import (
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 3).
		Width(min(width-8, maxWidth)).
		Render(strings.Join(rows, "\n"))
}

func renderSelector(theme Theme, state controllerpkg.State, label string, active bool) string {
	style := theme.StyleNormal
	if active {
		style = theme.StyleCursor
	}
	return style.Render("[ " + promptGlyphSet(state).Value + label + " ]")
}

func renderInputColumns(width, maxWidth int, left string, right string) string {
	contentWidth := min(width-8, maxWidth) - 8
	contentWidth = max(28, contentWidth)
	colWidth := max((contentWidth-2)/2, 12)
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

func appendDialogFooter(
	theme Theme,
	state controllerpkg.State,
	rows []string,
	hint string,
) []string {
	if strings.TrimSpace(state.ErrorMessage) != "" {
		rows = append(rows, "", theme.StyleError.Render(state.ErrorMessage))
	}
	if strings.TrimSpace(hint) != "" {
		rows = append(rows, "", theme.StyleDim.Render(hint))
	}
	return rows
}

func dialogSubmitChord(state controllerpkg.State) string {
	return "ctrl+s"
}

func dialogSubmitHint(state controllerpkg.State, label string) string {
	return "[" + dialogSubmitChord(state) + "] " + label
}

type dialogPromptSet struct {
	Search string
	Value  string
	Time   string
	Date   string
}

func promptGlyphSet(state controllerpkg.State) dialogPromptSet {
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

// formModal keeps the title, active field and footer visible in short terminals.
func formModal(theme Theme, state controllerpkg.State, maxWidth int, rows []string) string {
	width := max(12, min(state.Width-8, maxWidth)-6)
	if state.Height <= 0 {
		return modal(theme, state.Width, maxWidth, theme.ColorCyan, rows)
	}
	wrapped := make([]string, len(rows))
	for i, row := range rows {
		wrapped[i] = lipgloss.NewStyle().Width(width).Render(row)
	}
	if lipgloss.Height(strings.Join(wrapped, "\n"))+4 <= state.Height-2 {
		return modal(theme, state.Width, maxWidth, theme.ColorCyan, rows)
	}
	footer := wrapped[len(wrapped)-1]
	title := wrapped[0]
	errorRow := ""
	if state.ErrorMessage != "" {
		errorRow = lipgloss.NewStyle().Width(width).Render(theme.StyleError.Render(state.ErrorMessage))
	}
	budget := max(1, state.Height-8-lipgloss.Height(title)-lipgloss.Height(footer)-lipgloss.Height(errorRow))
	bodyRows := wrapped[1 : len(wrapped)-1]
	if errorRow != "" && len(bodyRows) >= 2 {
		bodyRows = bodyRows[:len(bodyRows)-2]
	}
	body := strings.Split(strings.Join(bodyRows, "\n"), "\n")
	target := 0
	active := ""
	for _, input := range state.Inputs {
		if input.Focused() {
			active = input.View()
			break
		}
	}
	if state.Description.Focused() {
		active = state.Description.View()
	}
	if active != "" {
		activeLines := strings.Split(lipgloss.NewStyle().Width(width).Render(active), "\n")
		needle := []rune(strings.TrimSpace(ansi.Strip(activeLines[0])))
		if len(needle) > 12 {
			needle = needle[:12]
		}
		for i, line := range body {
			if strings.Contains(ansi.Strip(line), string(needle)) {
				target = i
				break
			}
		}
	}
	first := max(0, target-budget/2)
	last := min(len(body), first+budget)
	visible := []string{title, theme.StyleDim.Render("[tab/shift+tab] move between fields")}
	visible = append(visible, body[first:last]...)
	if errorRow != "" {
		visible = append(visible, errorRow)
	}
	visible = append(visible, footer)
	return modal(theme, state.Width, maxWidth, theme.ColorCyan, visible)
}
