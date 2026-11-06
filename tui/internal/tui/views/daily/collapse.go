package daily

import (
	uistate "crona/tui/internal/tui/state"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
)

// The calendar shrinks to its week, then folds, before these panes drop below their rows.
const (
	dailyMinTaskRows   = 8
	dailyMinHabitRows  = 5
	dailyMinDetailRows = 8
)

func dailyCollapsed(state types.ContentState, pane string) bool {
	if state.Settings == nil {
		return false
	}
	return uistate.PaneCollapsed(state.Settings.CollapsedPanes, pane)
}

// foldHint is a pane header's `[⌥1] fold` (or `expand` when folded).
func foldHint(theme types.Theme, pane, action string) string {
	return viewchrome.KeyHint(theme, uistate.KeyLabel(uistate.FoldKey(pane)), action)
}

// collapsedPane is a folded pane: its title row alone, with the key that opens it.
func collapsedPane(theme types.Theme, title, pane string, width int) string {
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:  title + " ▸",
		Meta:   foldHint(theme, pane, "expand"),
		Dimmed: true,
		Width:  width,
		Height: 1,
	})
}

// scoreCalendarLines is the Summary Calendar for the shown date, boxed as the C dialog draws
// it; week keeps only the shown week.
func scoreCalendarLines(theme types.Theme, state types.ContentState, width int, week bool) []string {
	selection := viewcalendar.ScoreSelection{AnchorDate: dailyDate(state)}
	if state.Health != nil {
		selection.Today = state.Health.CurrentDate
	}
	if state.Settings != nil {
		selection.WeekStart = state.Settings.WeekStart
		selection.AwayDates = state.Settings.AwayDates
	}
	if state.SummarySnapshot != nil {
		selection.Scores = state.SummarySnapshot.CalendarScores
	}
	return viewcalendar.ScoreGrid(theme, selection, width, week)
}

// calendarPane is the score calendar at the largest size fits allows: the month, then the
// shown week, else folded to its title row.
func calendarPane(
	theme types.Theme,
	state types.ContentState,
	width int,
	fits func(calendarH int) bool,
) (string, int) {
	mode := calendarMode(state)
	if mode == uistate.CalendarHidden {
		return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
	}
	shapes := []bool{false, true}
	if mode == uistate.CalendarWeek {
		shapes = []bool{true}
	}
	for _, week := range shapes {
		lines := scoreCalendarLines(theme, state, width-2, week)
		height := len(lines) + 1
		if len(lines) > 0 && fits(height) {
			// The fold key steps a month to its week, and a week to hidden.
			next := "week"
			if week {
				next = "hide"
			}
			// The calendar never takes focus, so it is not faded like an unfocused list.
			return viewchrome.RenderSection(theme, viewchrome.Section{
				Title:  "Calendar",
				Meta:   foldHint(theme, uistate.CollapseDailyCalendar, next),
				Width:  width,
				Height: height,
				Lines:  lines,
			}), height
		}
	}
	return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
}

func calendarMode(state types.ContentState) uistate.CalendarMode {
	if state.Settings == nil {
		return uistate.CalendarMonth
	}
	return uistate.DailyCalendarMode(state.Settings.CollapsedPanes)
}

// CalendarMonthFits reports whether Today has room for the calendar's full month where it
// sits now, so the fold key can skip a month that would only show as its week.
func CalendarMonthFits(theme types.Theme, state types.ContentState) bool {
	if state.HistoricalAwayDate != "" {
		return false
	}
	width, fits := calendarSlot(state)
	lines := scoreCalendarLines(theme, state, width-2, false)
	return len(lines) > 0 && fits(len(lines)+1)
}
