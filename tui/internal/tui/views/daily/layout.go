package daily

import (
	uistate "crona.work/tui/internal/tui/state"
	viewaway "crona.work/tui/internal/tui/views/away"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

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
	width := state.Width
	dayLine, strip, rest := dailyChrome(theme, state)

	var body string
	if viewhelpers.ClassFor(width) == viewhelpers.WidthWide {
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

// dailyChrome is the day line and strip over Today's lists, and the rows left beneath them.
func dailyChrome(theme types.Theme, state types.ContentState) (dayLine, strip []string, rest int) {
	width, height := state.Width, state.Height
	// Without the detail pane, the day summary it holds shows as one line over the strip.
	if viewhelpers.ClassFor(width) != viewhelpers.WidthWide && height-1-dayStripMaxRows >= dailyMinBodyRows {
		if line := viewsummary.DayLine(theme, state, width); line != "" {
			dayLine, height = []string{line}, height-1
		}
	}
	stripRows := dayStripMaxRows
	for stripRows > 1 && height-stripRows < dailyMinBodyRows {
		stripRows--
	}
	// The detail pane's Day summary holds the strip's meters and signals, so with that pane
	// on screen the strip is left out.
	if !dailySummaryPaneShown(state) {
		strip = dayStripRows(theme, state, width, stripRows)
	}
	return dayLine, strip, height - len(strip)
}

// calendarSlot is the calendar's width and the test of whether a calendar of a given
// height fits where Today puts it: under the details on wide screens, under Habits otherwise.
func calendarSlot(state types.ContentState) (int, func(int) bool) {
	_, _, rest := dailyChrome(types.Theme{}, state)
	if viewhelpers.ClassFor(state.Width) == viewhelpers.WidthWide {
		return viewhelpers.Columns(state.Width, 3, 2)[1], detailCalendarFits(rest)
	}
	return state.Width, listsCalendarFits(state, rest, 0)
}

// detailCalendarFits keeps the details at least half the column.
func detailCalendarFits(height int) func(int) bool {
	return func(h int) bool { return height-h >= max(dailyMinDetailRows, height/2) }
}

// listsCalendarFits keeps Habits its rows in the lower half, or, with Habits folded, Tasks
// its rows; belowH is what already sits under Habits.
func listsCalendarFits(state types.ContentState, height, belowH int) func(int) bool {
	if dailyCollapsed(state, uistate.CollapseDailyHabits) {
		return func(h int) bool { return height-1-belowH-h >= dailyMinTaskRows }
	}
	return func(h int) bool { return height/2-belowH-h >= dailyMinHabitRows }
}

// dailyRightColumn is the score calendar over the detail pane. The calendar gives way first,
// to its week and then folded, so details keep half the column.
func dailyRightColumn(theme types.Theme, state types.ContentState, width, height int) string {
	// The boxed month is tall, so it shows only while details keep half the column;
	// otherwise the shown week does.
	calendar, calendarH := calendarPane(theme, state, width, detailCalendarFits(height))
	return lipgloss.JoinVertical(lipgloss.Left, calendar, renderDetail(theme, state, width, height-calendarH))
}

// dailyLists is Tasks over Habits. Tasks keeps the top half and Habits the bottom half;
// folded, Habits keeps its title row and Tasks takes the rest. withCalendar puts the
// calendar under Habits, where it shares Habits' half while Habits keeps dailyMinHabitRows.
func dailyLists(theme types.Theme, state types.ContentState, width, height int, withCalendar bool) string {
	habitsFolded := dailyCollapsed(state, uistate.CollapseDailyHabits)
	var below []string
	belowH := 0
	lower := height / 2
	if withCalendar {
		calendar, calendarH := calendarPane(theme, state, width, listsCalendarFits(state, height, belowH))
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

// dailySummaryPaneShown is true when the wide layout shows the detail pane.
func dailySummaryPaneShown(state types.ContentState) bool {
	return viewhelpers.ClassFor(state.Width) == viewhelpers.WidthWide
}
