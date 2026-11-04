package daily

import (
	viewaway "crona/tui/internal/tui/views/away"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// dailyMinBodyRows is the room Tasks and Habits keep before the strip gives up rows.
const dailyMinBodyRows = 12

// renderView lays Daily out as a day strip over Tasks and Habits, each holding half the
// height so neither moves as the other changes, and on wide screens a detail pane.
func renderView(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	width, height := state.Width, state.Height
	// Without the detail pane, the day summary it holds shows as one line over the strip.
	var dayLine []string
	if viewhelpers.ClassFor(width) != viewhelpers.WidthWide && height-1-dayStripMaxRows >= dailyMinBodyRows {
		if line := viewsummary.DayLine(theme, state, width); line != "" {
			dayLine, height = []string{line}, height-1
		}
	}
	stripRows := dayStripMaxRows
	for stripRows > 1 && height-stripRows < dailyMinBodyRows {
		stripRows--
	}
	strip := dayStripRows(theme, state, width, stripRows)
	rest := height - len(strip)

	listW := width
	detail := ""
	if viewhelpers.ClassFor(width) == viewhelpers.WidthWide {
		cols := viewhelpers.Columns(width, 3, 2)
		listW = cols[0]
		detail = renderDetail(theme, state, cols[1], rest)
	}
	habitsH := rest / 2
	body := lipgloss.JoinVertical(lipgloss.Left,
		renderIssues(theme, state, listW, rest-habitsH),
		renderHabits(theme, state, listW, habitsH),
	)
	if detail != "" {
		body = viewchrome.JoinColumns(theme, rest, body, detail)
	}
	return lipgloss.JoinVertical(lipgloss.Left, append(append(dayLine, strip...), body)...)
}
