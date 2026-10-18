package daily

import (
	"fmt"
	"strings"

	viewcalendar "crona/tui/internal/tui/views/calendar"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

const (
	dayStripBarWidth = 10
	// dayStripColumnGap separates the meter, signal and streak columns.
	dayStripColumnGap = "     "
)

// dayStripMaxRows is the strip's full height: the Tasks and Habits meter rows.
const dayStripMaxRows = 2

// renderDayStrip is the htop-style header above the Daily lists at its full height.
func renderDayStrip(theme types.Theme, state types.ContentState, width int) []string {
	return dayStripRows(theme, state, width, dayStripMaxRows)
}

// dayStripRows renders at most maxRows of the strip: Tasks over Habits meters, then columns
// pairing a check-in signal over a streak (Mood over Focus, Energy over Check-ins, then
// Sleep), with the shown week pinned to the right end under the sub-tab line's date. When
// space runs out the rightmost columns go first, then the week, then the rest of the columns.
func dayStripRows(theme types.Theme, state types.ContentState, width, maxRows int) []string {
	if width < 1 || maxRows < 1 {
		return nil
	}
	meters := dayMeterColumn(theme, state, true)
	if columnWidth(meters)+1 > width {
		meters = dayMeterColumn(theme, state, false)
	}
	grid := dayGridColumns(theme, state)
	// fill adds grid columns left to right while they fit beside reserve cells.
	fill := func(reserve int) [][]string {
		columns := [][]string{meters}
		for _, column := range grid {
			if stripWidth(append(columns, column))+reserve+1 > width {
				break
			}
			columns = append(columns, column)
		}
		return columns
	}
	week := []string(nil)
	if maxRows >= 2 {
		week = dayWeek(theme, state)
	}
	columns := fill(0)
	if len(week) > 0 {
		// The week sits right-aligned, so it costs its width, a gap and the edge cell.
		reserve := len(dayStripColumnGap) + columnWidth(week) + 1
		withWeek := fill(reserve)
		switch {
		case stripWidth(withWeek)+reserve+1 > width, len(withWeek) == 1 && len(grid) > 0:
			week = nil
		default:
			columns = withWeek
		}
	}
	height := len(week)
	for _, column := range columns {
		height = max(height, len(column))
	}
	rows := make([]string, min(height, maxRows))
	for i := range rows {
		cells := make([]string, 0, len(columns))
		for _, column := range columns {
			cell := ""
			if i < len(column) {
				cell = column[i]
			}
			cells = append(cells, padCells(cell, columnWidth(column)))
		}
		left := " " + strings.TrimRight(strings.Join(cells, dayStripColumnGap), " ")
		if i < len(week) {
			right := padCells(week[i], columnWidth(week)) + " "
			left += strings.Repeat(" ", max(0, width-ansi.StringWidth(left)-ansi.StringWidth(right))) + right
		}
		rows[i] = padCells(left, width)
	}
	return rows
}

// dayWeek is the shown date's week as weekday names over day numbers.
func dayWeek(theme types.Theme, state types.ContentState) []string {
	lines, ok := viewcalendar.WeekRow(theme, viewcalendar.Selection{
		SelectedDate: dailyDate(state),
		WeekStart:    state.WeekStart,
		AwayDates:    state.AwayDates(),
	})
	if !ok {
		return nil
	}
	return lines
}

// SubTabMeta is the right side of Daily's sub-tab line: the shown date and its week,
// flanked by the date-navigation keys written out as in the status line. Within room it
// drops the key words first, then the keys, then the week, so the date always shows.
func SubTabMeta(theme types.Theme, state types.ContentState, room int) string {
	return viewcalendar.DateNav(theme, dailyDate(state), "", state.Settings, state.WeekStart, room)
}

func dayMeterColumn(theme types.Theme, state types.ContentState, withBars bool) []string {
	resolved, estimate := 0, 0
	statusCounts := map[string]int{}
	for _, issue := range state.DailyPlannedIssues {
		statusCounts[string(issue.Status)]++
		if issue.EstimateMinutes != nil {
			estimate += *issue.EstimateMinutes
		}
		if issue.Status == "done" || issue.Status == "abandoned" {
			resolved++
		}
	}
	habitsDone, habitsFailed := 0, 0
	for _, habit := range state.DueHabits {
		switch habit.Status {
		case "completed":
			habitsDone++
		case "failed":
			habitsFailed++
		}
	}
	worked := 0
	if state.DailySummary != nil {
		worked = state.DailySummary.WorkedSeconds
	}
	values := []string{
		fmt.Sprintf("%d/%d", resolved, len(state.DailyPlannedIssues)),
		fmt.Sprintf("%d/%d", habitsDone, len(state.DueHabits)),
	}
	valueWidth := max(len(values[0]), len(values[1]))
	row := func(label, value, bar, tail string) string {
		line := theme.StyleHeader.Render(padCells(label, 6)) + "  " + theme.StyleNormal.Render(padCells(value, valueWidth))
		if bar != "" {
			line += "  " + bar
		}
		if tail != "" {
			line += "  " + tail
		}
		return line
	}
	tasksBar, habitsBar := "", ""
	if withBars {
		tasksBar = renderIssueStatusBar(theme, statusCounts, dayStripBarWidth)
		habitsBar = renderHabitBar(theme, habitsDone, habitsFailed, len(state.DueHabits), dayStripBarWidth)
	}
	return []string{
		row("Tasks", values[0], tasksBar, theme.StyleDim.Render(issueProgressText(worked, estimate, true))),
		row("Habits", values[1], habitsBar, ""),
	}
}

// dayGridColumns pairs each check-in signal (top row) with a streak (bottom row) in one
// aligned column, then the day's focus score over its plan accountability; a column with
// neither value is left out.
func dayGridColumns(theme types.Theme, state types.ContentState) [][]string {
	type cell struct{ label, value string }
	top := []cell{
		{"Mood", signalMoodValue(state)},
		{"Energy", signalEnergyValue(state)},
		{"Sleep", signalSleepValue(state)},
	}
	bottom := []cell{{"Focus", ""}, {"Check-ins", ""}, {"", ""}, {"Plan", ""}}
	if streaks := state.DailyStreaks; streaks != nil {
		bottom[0].value = fmt.Sprintf("%dd  ", streaks.CurrentFocusDays) + theme.StyleDim.Render(fmt.Sprintf("best %dd", streaks.LongestFocusDays))
		bottom[1].value = fmt.Sprintf("%dd  ", streaks.CurrentCheckInDays) + theme.StyleDim.Render(fmt.Sprintf("best %dd", streaks.LongestCheckInDays))
	}
	// The day's headline scores come from the summary snapshot, once it is the shown date's.
	top = append(top, cell{"Score", ""})
	if snapshot := state.SummarySnapshot; snapshot != nil && snapshot.Date == state.SummaryDate {
		if snapshot.Focus != nil {
			top[3].value = fmt.Sprintf("%d/100", snapshot.Focus.Score)
		}
		if snapshot.Plan != nil && strings.TrimSpace(snapshot.Plan.Date) != "" {
			bottom[3].value = fmt.Sprintf("%.0f%%", snapshot.Plan.Summary.AccountabilityScore)
		}
	}
	columns := [][]string{}
	for i := range top {
		pair := []cell{top[i]}
		if i < len(bottom) {
			pair = append(pair, bottom[i])
		}
		labelWidth := 0
		for _, c := range pair {
			if c.value != "" {
				labelWidth = max(labelWidth, len(c.label))
			}
		}
		if labelWidth == 0 {
			continue
		}
		column := make([]string, len(pair))
		for row, c := range pair {
			if c.value != "" {
				column[row] = theme.StyleHeader.Render(padCells(c.label, labelWidth)) + "  " +
					theme.StyleNormal.Render(strings.TrimPrefix(c.value, "today "))
			}
		}
		columns = append(columns, column)
	}
	return columns
}

func columnWidth(column []string) int {
	width := 0
	for _, cell := range column {
		width = max(width, ansi.StringWidth(cell))
	}
	return width
}

func stripWidth(columns [][]string) int {
	width := 0
	for i, column := range columns {
		if i > 0 {
			width += len(dayStripColumnGap)
		}
		width += columnWidth(column)
	}
	return width
}

// dailyDate is the day Daily shows: the loaded summary's date, else the dashboard date.
func dailyDate(state types.ContentState) string {
	if state.DailySummary != nil && state.DailySummary.Date != "" {
		return state.DailySummary.Date
	}
	return viewruntime.CurrentDashboardDate(state)
}

// padCells truncates or pads s to exactly width cells.
func padCells(s string, width int) string {
	s = viewhelpers.TruncateANSI(s, width)
	if w := ansi.StringWidth(s); w < width {
		s += strings.Repeat(" ", width-w)
	}
	return ansi.Truncate(s, width, "")
}
