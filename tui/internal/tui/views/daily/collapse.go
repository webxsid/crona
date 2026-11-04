package daily

import (
	"strings"

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

// scoreCalendarLines is the Summary Calendar's month for the shown date with each day's focus
// score, centred in width; week keeps the month and weekday rows and only the shown week.
func scoreCalendarLines(theme types.Theme, state types.ContentState, width int, week bool) []string {
	selection := viewcalendar.ScoreSelection{AnchorDate: dailyDate(state)}
	if state.Settings != nil {
		selection.WeekStart = state.Settings.WeekStart
		selection.AwayDates = state.Settings.AwayDates
	}
	if state.SummarySnapshot != nil {
		selection.Scores = state.SummarySnapshot.CalendarScores
	}
	lines := viewcalendar.ScoreMonth(theme, selection)
	if len(lines) < 3 || viewcalendar.MaxLineWidth(lines) > width {
		return nil
	}
	if week {
		lines = append(lines[:2:2], lines[2+viewcalendar.ScoreWeekIndex(selection)])
	}
	pad := strings.Repeat(" ", max(0, (width-viewcalendar.MaxLineWidth(lines))/2))
	for i, line := range lines {
		lines[i] = pad + line
	}
	return lines
}

// calendarPane is the score calendar at the largest size fits allows: the month, then the
// shown week, else folded to its title row.
func calendarPane(
	theme types.Theme,
	state types.ContentState,
	width int,
	fits func(calendarH int) bool,
) (string, int) {
	for _, week := range []bool{false, true} {
		lines := scoreCalendarLines(theme, state, width-2, week)
		height := len(lines) + 1
		if len(lines) > 0 && fits(height) {
			return dailySection(theme, "Calendar", foldHint(theme, uistate.CollapseDailyCalendar, "fold"),
				false, width, height, lines), height
		}
	}
	return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
}
