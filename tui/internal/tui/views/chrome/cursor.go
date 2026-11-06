package viewchrome

import (
	rootchrome "crona.work/tui/internal/tui/chrome"

	"github.com/charmbracelet/lipgloss"
)

const SelectionCursor = rootchrome.SelectionCursor

// SelectedRow is a focused list's selected row: the caret in the cursor accent, the row
// bold in its own colours. Green stays for done and success, not for selection.
func SelectedRow(theme Theme, line string) string {
	return theme.StyleCursor.Render(SelectionCursor+" ") + lipgloss.NewStyle().Bold(true).Render(line)
}
