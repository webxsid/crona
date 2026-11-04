package daily

import (
	uistate "crona/tui/internal/tui/state"
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

	var body string
	if viewhelpers.ClassFor(width) == viewhelpers.WidthWide &&
		(!dailyCollapsed(state, uistate.CollapseDailyDetail) || !dailyCollapsed(state, uistate.CollapseDailyCalendar)) {
		cols := viewhelpers.Columns(width, 3, 2)
		body = viewchrome.JoinColumns(theme, rest,
			dailyLists(theme, state, cols[0], rest, false),
			dailyRightColumn(theme, state, cols[1], rest),
		)
	} else {
		body = dailyLists(theme, state, width, rest, true)
	}
	return lipgloss.JoinVertical(lipgloss.Left, append(append(dayLine, strip...), body)...)
}

// dailyRightColumn is the detail pane over the score calendar. Folded, each keeps its title
// row; the calendar gives way first, to its week and then folded, so details keep
// dailyMinDetailRows.
func dailyRightColumn(theme types.Theme, state types.ContentState, width, height int) string {
	detailFolded := dailyCollapsed(state, uistate.CollapseDailyDetail)
	detailMin := dailyMinDetailRows
	if detailFolded {
		detailMin = 1
	}
	calendar, calendarH := collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
	if !dailyCollapsed(state, uistate.CollapseDailyCalendar) {
		calendar, calendarH = calendarPane(theme, state, width, func(h int) bool { return height-h >= detailMin })
	}
	if detailFolded {
		top := collapsedPane(theme, "Details", uistate.CollapseDailyDetail, width)
		filler := lipgloss.NewStyle().Width(width).Height(max(0, height-1-calendarH)).Render("")
		if height-1-calendarH <= 0 {
			return lipgloss.JoinVertical(lipgloss.Left, top, calendar)
		}
		return lipgloss.JoinVertical(lipgloss.Left, top, calendar, filler)
	}
	return lipgloss.JoinVertical(lipgloss.Left, renderDetail(theme, state, width, height-calendarH), calendar)
}

// dailyLists is Tasks over Habits. Tasks keeps the top half and Habits the bottom half;
// folded, Habits keeps its title row and Tasks takes the rest. withCalendar puts the
// calendar (and, on wide screens with both side panes folded, the detail's folded row) under
// Habits, where the calendar shares Habits' half while Habits keeps dailyMinHabitRows.
func dailyLists(theme types.Theme, state types.ContentState, width, height int, withCalendar bool) string {
	habitsFolded := dailyCollapsed(state, uistate.CollapseDailyHabits)
	var below []string
	belowH := 0
	if withCalendar && viewhelpers.ClassFor(state.Width) == viewhelpers.WidthWide {
		below = append(below, collapsedPane(theme, "Details", uistate.CollapseDailyDetail, width))
		belowH++
	}
	lower := height / 2
	if withCalendar {
		fits := func(h int) bool { return lower-belowH-h >= dailyMinHabitRows }
		if habitsFolded {
			fits = func(h int) bool { return height-1-belowH-h >= dailyMinTaskRows }
		}
		calendar, calendarH := collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
		if !dailyCollapsed(state, uistate.CollapseDailyCalendar) {
			calendar, calendarH = calendarPane(theme, state, width, fits)
		}
		below = append([]string{calendar}, below...)
		belowH += calendarH
	}
	habits, habitsH := collapsedPane(theme, "Habits", uistate.CollapseDailyHabits, width), 1
	if !habitsFolded {
		habitsH = max(1, lower-belowH)
		habits = renderHabits(theme, state, width, habitsH)
	}
	parts := append([]string{renderIssues(theme, state, width, height-habitsH-belowH), habits}, below...)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
