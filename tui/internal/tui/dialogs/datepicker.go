package dialogs

import (
	"fmt"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
)

func PopulateDatePresentation(theme Theme, state State, currentDate string) State {
	selected := DialogDate(state, currentDate)
	monthStart := DialogMonth(state, currentDate)
	title := "Pick Due Date"
	if state.Parent == "create_issue_meta" || state.Parent == "create_issue_default" {
		title = "Pick Due Date For New Issue (Optional)"
	}
	state.DateTitle = title
	state.DateHeader = selected.Format("Mon") + ", " + shareddatefmt.FormatDate(selected, nil)
	state.DateMonth = monthStart.Format("January 2006")
	state.DateGrid = renderCalendarGrid(theme, monthStart, selected)
	return state
}

func ResolveDialogDate(initial *string, currentDate string) time.Time {
	if initial != nil {
		if parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*initial)); err == nil {
			return parsed
		}
	}
	if parsed, err := time.Parse("2006-01-02", currentDate); err == nil {
		return parsed
	}
	return time.Now()
}

func DialogDate(state State, currentDate string) time.Time {
	if parsed, err := time.Parse("2006-01-02", state.DateCursorValue); err == nil {
		return parsed
	}
	return ResolveDialogDate(nil, currentDate)
}

func DialogMonth(state State, currentDate string) time.Time {
	if parsed, err := time.Parse("2006-01-02", state.DateMonthValue); err == nil {
		return parsed
	}
	date := DialogDate(state, currentDate)
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func renderCalendarGrid(theme Theme, monthStart, selected time.Time) string {
	headers := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	lines := []string{strings.Join(headers, "  ")}
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	for week := 0; week < 6; week++ {
		cells := make([]string, 0, 7)
		for day := 0; day < 7; day++ {
			current := gridStart.AddDate(0, 0, week*7+day)
			label := fmt.Sprintf("%2d", current.Day())
			cell := " " + label + " "
			style := theme.StyleNormal
			if current.Month() != monthStart.Month() {
				style = theme.StyleDim
			}
			if sameDay(current, selected) {
				cell = theme.StyleCursor.Render(cell)
			} else {
				cell = style.Render(cell)
			}
			cells = append(cells, cell)
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return strings.Join(lines, "\n")
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
