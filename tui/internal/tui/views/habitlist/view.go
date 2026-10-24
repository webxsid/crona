// Package habitlist renders scoped habit definitions.
package habitlist

import (
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	"crona/tui/internal/tui/views/daily"
	"crona/tui/internal/tui/views/scopepane"
	types "crona/tui/internal/tui/views/types"
)

// Render shows the habit summary, scope, cards and wide-screen detail.
func Render(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	rows := StripRows(theme, state, state.Width)
	rows = rows[:min(len(rows), state.Height)]
	height := state.Height - len(rows)
	if height == 0 {
		return strings.Join(rows, "\n")
	}
	if state.Width < 100 {
		rows = append(rows, scopepane.RenderRow(theme, state, state.Width))
		if height > 1 {
			rows = append(rows, renderHabits(theme, state, state.Width, height-1))
		}
		return strings.Join(rows, "\n")
	}
	listWidth := state.Width - 27
	columns := []string{
		scopepane.Render(theme, state, scopepane.HabitCounts(state.AllHabits), 24, height),
	}
	detailWidth := 0
	if state.Width >= 160 {
		detailWidth = (listWidth - 3) / 2
		listWidth -= detailWidth + 3
	}
	columns = append(columns, renderHabits(theme, state, listWidth, height))
	if detailWidth > 0 {
		title := "Selected"
		lines := []string{theme.StyleDim.Render("Nothing selected")}
		indices := habitIndices(state)
		cur := state.Cursors["habits"]
		if cur >= 0 && cur < len(indices) {
			habit := state.ScopedHabits[indices[cur]]
			title = habit.Name
			lines = daily.HabitDetailLines(theme, state, habit, detailWidth-2)
		}
		columns = append(
			columns,
			viewchrome.RenderSection(
				theme,
				viewchrome.Section{Title: title, Width: detailWidth, Height: height, Lines: lines},
			),
		)
	}
	rows = append(rows, viewchrome.JoinColumns(theme, height, columns...))
	return strings.Join(rows, "\n")
}
