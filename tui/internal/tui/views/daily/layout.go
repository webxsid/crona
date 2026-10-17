package daily

import (
	"strings"

	viewaway "crona/tui/internal/tui/views/away"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// dailyMinTaskRows keeps the primary list usable when habits or the strip compete for rows.
const dailyMinTaskRows = 5

// renderView lays Daily out as a day strip over one growing task list, habits sized to
// their content, and on wide screens a detail pane for the selection.
func renderView(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	width, height := state.Width, state.Height
	strip := renderDayStrip(theme, state, width)
	if len(strip) > 1 && height-len(strip) < dailyMinTaskRows+3 {
		strip = strip[:1]
	}
	rest := height - len(strip)

	listW := width
	detail := ""
	if viewhelpers.ClassFor(width) == viewhelpers.WidthWide {
		cols := viewhelpers.Columns(width, 3, 2)
		listW = cols[0]
		detail = renderDetail(theme, state, cols[1], rest)
	}
	habitsH := dailyHabitsHeight(state, rest)
	// A short task list ends at its last row so Habits follows it; spare rows go below.
	tasksH := min(rest-habitsH, max(dailyMinTaskRows, dailyTaskRows(state, listW)))
	sections := []string{
		renderIssues(theme, state, listW, tasksH),
		renderHabits(theme, state, listW, habitsH),
	}
	if spare := rest - tasksH - habitsH; spare > 0 {
		sections = append(sections, strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", listW)+"\n", spare), "\n"))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if detail != "" {
		body = viewchrome.JoinColumns(theme, rest, body, detail)
	}
	return lipgloss.JoinVertical(lipgloss.Left, append(strip, body)...)
}

// dailyHabitsHeight sizes Habits to its rows, capped so the task list keeps most of the space.
func dailyHabitsHeight(state types.ContentState, rest int) int {
	rows := len(viewhelpers.FilteredStrings(
		viewruntime.HabitDailyItems(state.DueHabits),
		state.Filters["habits"],
	))
	limit := max(3, rest/3)
	if state.Pane == "habits" {
		limit = max(3, rest/2)
	}
	want := 1 + max(1, rows)
	return max(2, min(want, limit, rest-dailyMinTaskRows))
}
