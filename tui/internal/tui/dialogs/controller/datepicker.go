package controller

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
)

type dueDatePolicyKind string

const (
	dueDateAllowed dueDatePolicyKind = ""
	dueDatePast    dueDatePolicyKind = "past"
	dueDateRest    dueDatePolicyKind = "rest"
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
	state.DateGrid = renderCalendarGrid(theme, monthStart, selected, state)
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

func renderCalendarGrid(theme Theme, monthStart, selected time.Time, state State) string {
	headers := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	lines := []string{strings.Join(headers, "  ")}
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	for week := range 6 {
		cells := make([]string, 0, 7)
		for day := range 7 {
			current := gridStart.AddDate(0, 0, week*7+day)
			label := fmt.Sprintf("%2d", current.Day())
			cell := " " + label + " "
			style := theme.StyleNormal
			invalidKind := dueDateAllowed
			if isIssueDueDatePicker(state) {
				invalidKind = dueDateInvalidKind(state, current.Format("2006-01-02"))
			}
			if current.Month() != monthStart.Month() {
				style = theme.StyleDim
			}
			if sameDay(current, selected) {
				style = theme.StyleCursor
			} else {
				if invalidKind == dueDatePast {
					style = theme.StyleDim
				}
			}
			switch invalidKind {
			case dueDatePast:
				style = style.Foreground(theme.ColorDim)
			case dueDateRest:
				style = style.Foreground(theme.ColorRed)
			}
			cell = style.Render(cell)
			cells = append(cells, cell)
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return strings.Join(lines, "\n")
}

func isIssueDueDatePicker(state State) bool {
	return state.Parent == "create_issue_meta" ||
		state.Parent == "create_issue_default" ||
		state.Parent == "edit_issue" ||
		(state.Parent == "" && state.IssueID > 0)
}

func dueDateInvalidKind(state State, date string) dueDatePolicyKind {
	selected, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	if err != nil {
		return dueDateAllowed
	}
	return dueDateInvalidKindForTime(state, selected)
}

func dueDateInvalidKindForTime(state State, selected time.Time) dueDatePolicyKind {
	today, err := time.Parse("2006-01-02", strings.TrimSpace(state.DueDateToday))
	if err != nil {
		return dueDateAllowed
	}
	if selected.Before(today) {
		return dueDatePast
	}
	date := selected.Format("2006-01-02")
	if slices.Contains(state.DueDateRestDates, date) || slices.Contains(state.DueDateAwayDates, date) {
		return dueDateRest
	}
	if slices.Contains(state.DueDateRestWeekdays, int(selected.Weekday())) {
		return dueDateRest
	}
	return dueDateAllowed
}

func ValidateDueDate(state State, date string) error {
	date = strings.TrimSpace(date)
	if date == "" {
		return nil
	}
	selected, err := time.Parse("2006-01-02", date)
	if err != nil {
		return errors.New("due date must be YYYY-MM-DD")
	}
	switch dueDateInvalidKindForTime(state, selected) {
	case dueDatePast:
		return errors.New("due dates cannot be in the past")
	case dueDateRest:
		return errors.New("due dates cannot fall on configured rest days")
	default:
		return nil
	}
}

func ValidateDueDateAt(state State, date, currentDate string) error {
	if strings.TrimSpace(state.DueDateToday) == "" {
		state.DueDateToday = strings.TrimSpace(currentDate)
	}
	return ValidateDueDate(state, date)
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
