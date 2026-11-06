package controller

import (
	"fmt"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
	sharedtypes "crona/shared/types"
	viewcalendar "crona/tui/internal/tui/views/calendar"

	"github.com/charmbracelet/lipgloss"
)

func PopulateSummaryCalendarPresentation(theme Theme, state State, today string) State {
	month := DialogMonth(state, today)
	selected := DialogDate(state, today)
	state.DateTitle = "Summary Calendar"
	state.DateHeader = selected.Format("Mon, Jan 2, 2006")
	state.DateMonth = month.Format("January 2006")
	state.DateGrid = renderSummaryCalendarGrid(theme, state, month, selected, today)
	return state
}

func renderSummaryCalendarGrid(
	theme Theme,
	state State,
	month, selected time.Time,
	today string,
) string {
	cellWidth := 7
	if state.Width > 0 && state.Width < 110 {
		cellWidth = 5
	}
	weekStart := sharedtypes.NormalizeWeekStart(state.DueDateWeekStart)
	weekdays := viewcalendar.WeekdayNames(weekStart)
	header := make([]string, len(weekdays))
	for i, weekday := range weekdays {
		// A box is cellWidth plus its two border cells; the name sits over the day number.
		header[i] = lipgloss.NewStyle().
			Width(cellWidth + 2).
			PaddingLeft(2).
			Render(weekday)
	}
	lines := []string{"    " + strings.Join(header, " ")}
	start := shareddatefmt.StartOfWeek(month, weekStart)
	todayDate, _ := time.Parse("2006-01-02", today)
	for week := 0; week < 6; week++ {
		cells := make([]string, 0, 7)
		for day := 0; day < 7; day++ {
			date := start.AddDate(0, 0, week*7+day)
			cells = append(
				cells,
				renderSummaryDayCell(theme, state, date, selected, todayDate, month, cellWidth),
			)
		}
		// Each cell is multiline. Joining the raw strings would concatenate their
		// newlines vertically, which collapses the grid into one column. Lip Gloss
		// joins corresponding lines horizontally and preserves the boxed layout.
		row := make([]string, 0, len(cells)*2-1)
		for i, cell := range cells {
			if i > 0 {
				row = append(row, " ")
			}
			row = append(row, cell)
		}
		rowStart := start.AddDate(0, 0, week*7)
		lines = append(lines, viewcalendar.WeekRowLines(
			theme.StyleDim.Render(viewcalendar.WeekLabel(rowStart, weekStart)),
			strings.Split(lipgloss.JoinHorizontal(lipgloss.Top, row...), "\n"),
		)...)
	}
	return strings.Join(lines, "\n")
}

func renderSummaryDayCell(
	theme Theme,
	state State,
	date, selected, today, month time.Time,
	cellWidth int,
) string {
	iso := date.Format("2006-01-02")
	isAway := containsDate(state.DueDateAwayDates, iso)
	isFuture := !today.IsZero() && date.After(today)
	isSelected := date.Equal(selected)
	dayStyle := theme.StyleNormal
	if date.Month() != month.Month() {
		dayStyle = theme.StyleDim
	}
	if isFuture {
		dayStyle = theme.StyleDim
	}
	if isAway {
		dayStyle = dayStyle.Foreground(theme.ColorGreen)
	}
	if isSelected {
		dayStyle = theme.StyleCursor
	}
	label := dayStyle.Render(fmt.Sprintf("%-2d", date.Day()))
	value := theme.StyleDim.Render("—")
	if isAway {
		awayLabel := "Away"
		if cellWidth < len(awayLabel) {
			awayLabel = "Awy"
		}
		value = theme.StyleNormal.Foreground(theme.ColorGreen).Render(awayLabel)
	} else if score, ok := state.CalendarScores[iso]; ok && score.HasData && !isFuture {
		color := theme.ColorYellow
		if score.Score >= 75 {
			color = theme.ColorGreen
		} else if score.Score < 45 {
			color = theme.ColorRed
		}
		value = lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%3d", score.Score))
	} else if !isFuture && date.Month() == month.Month() {
		value = theme.StyleDim.Render("…")
	}
	border := theme.ColorDim
	if isToday(date, today) {
		border = theme.ColorYellow
	}
	if isSelected {
		border = theme.ColorCyan
	}
	return lipgloss.NewStyle().
		Width(cellWidth).
		Height(2).
		PaddingLeft(1).
		PaddingRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Render(label + "\n" + value)
}

func isToday(date, today time.Time) bool { return !today.IsZero() && date.Equal(today) }

func containsDate(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
