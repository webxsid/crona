package calendar

import (
	"fmt"
	"strings"
	"time"

	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Selection struct {
	AnchorDate   string
	SelectedDate string
	RangeStart   string
	RangeEnd     string
	MaxLines     int
	Today        string
}

func Render(theme types.Theme, selection Selection) []string {
	anchor, ok := parseISODate(firstNonEmpty(selection.AnchorDate, selection.SelectedDate, selection.RangeEnd, selection.RangeStart))
	if !ok {
		return nil
	}
	today, ok := parseISODate(selection.Today)
	if !ok {
		today, _ = parseISODate(time.Now().Format("2006-01-02"))
	}
	selected, hasSelected := parseISODate(selection.SelectedDate)
	rangeStart, hasRangeStart := parseISODate(selection.RangeStart)
	rangeEnd, hasRangeEnd := parseISODate(selection.RangeEnd)
	if hasRangeStart && hasRangeEnd && rangeEnd.Before(rangeStart) {
		rangeStart, rangeEnd = rangeEnd, rangeStart
	}

	monthStart := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")),
		calendarMetaLine(theme, selected, hasSelected, rangeStart, rangeEnd, hasRangeStart && hasRangeEnd, today),
		theme.StyleDim.Render("Wk  Mo Tu We Th Fr Sa Su"),
	}
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	selectedWeek := 0
	if hasSelected {
		selectedWeek = ISOWeek(selected)
	} else if hasRangeEnd {
		selectedWeek = ISOWeek(rangeEnd)
	}
	currentWeek := ISOWeek(today)
	selectedDateStyle := lipgloss.NewStyle().Background(theme.ColorGreen).Foreground(lipgloss.Color("0")).Bold(true)
	todayStyle := lipgloss.NewStyle().Background(theme.ColorYellow).Foreground(lipgloss.Color("0")).Bold(true)
	rangeStyle := lipgloss.NewStyle().Background(theme.ColorBlue).Foreground(theme.ColorWhite)
	currentWeekStyle := lipgloss.NewStyle().Background(theme.ColorYellow).Foreground(lipgloss.Color("0")).Bold(true)
	for week := range 6 {
		rowStart := gridStart.AddDate(0, 0, week*7)
		rowWeek := ISOWeek(rowStart)
		weekLabel := fmt.Sprintf("%2d", rowWeek)
		switch rowWeek {
		case currentWeek:
			weekLabel = currentWeekStyle.Render(weekLabel)
		case selectedWeek:
			weekLabel = theme.StyleHeader.Render(weekLabel)
		default:
			weekLabel = theme.StyleDim.Render(weekLabel)
		}
		cells := make([]string, 0, 7)
		for day := range 7 {
			current := rowStart.AddDate(0, 0, day)
			inSelected := hasSelected && sameDay(current, selected)
			inRange := hasRangeStart && hasRangeEnd && !current.Before(rangeStart) && !current.After(rangeEnd)
			isToday := sameDay(current, today)
			cell := fmt.Sprintf("%2d", current.Day())
			switch {
			case inSelected:
				cell = selectedDateStyle.Render(cell)
			case isToday:
				cell = todayStyle.Render(cell)
			case inRange:
				cell = rangeStyle.Render(cell)
			case current.Month() != monthStart.Month():
				cell = theme.StyleDim.Render(cell)
			case rowWeek == selectedWeek && selectedWeek != 0:
				cell = theme.StyleHeader.Render(cell)
			default:
				cell = theme.StyleNormal.Render(cell)
			}
			cells = append(cells, cell)
		}
		lines = append(lines, weekLabel+"  "+strings.Join(cells, " "))
	}
	return Window(lines, anchor, selection.MaxLines)
}

func Window(lines []string, anchor time.Time, maxLines int) []string {
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	if maxLines <= 3 {
		return lines[:maxLines]
	}
	headers := lines[:3]
	weeks := lines[3:]
	visibleWeeks := maxLines - len(headers)
	if visibleWeeks >= len(weeks) {
		return lines
	}
	monthStart := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
	selectedWeek := weekIndexForDate(monthStart, anchor)
	start := selectedWeek - (visibleWeeks / 2)
	start = max(0, start)
	if start+visibleWeeks > len(weeks) {
		start = len(weeks) - visibleWeeks
	}
	window := append([]string{}, headers...)
	window = append(window, weeks[start:start+visibleWeeks]...)
	return window
}

func ShouldRender(innerWidth int) bool {
	return innerWidth >= 84
}

func MergeBeside(leftLines, calendarLines []string, innerWidth, gutterWidth int) []string {
	leftWidth, rightWidth := ColumnWidths(innerWidth, MaxLineWidth(calendarLines), gutterWidth)
	if rightWidth < 1 {
		return leftLines
	}
	totalLines := max(len(leftLines), len(calendarLines))
	start := 0
	if totalLines > len(calendarLines) {
		start = (totalLines - len(calendarLines)) / 2
	}
	merged := make([]string, 0, totalLines)
	gutter := strings.Repeat(" ", max(0, gutterWidth))
	for i := range totalLines {
		left := ""
		if i < len(leftLines) {
			left = ansi.Truncate(leftLines[i], leftWidth, "")
		}
		left = lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(left)
		if i >= start && i < start+len(calendarLines) {
			right := lipgloss.NewStyle().Width(rightWidth).MaxWidth(rightWidth).Render(calendarLines[i-start])
			merged = append(merged, left+gutter+right)
			continue
		}
		merged = append(merged, left)
	}
	return merged
}

func ColumnWidths(innerWidth, calendarWidth, gutterWidth int) (int, int) {
	if calendarWidth < 1 {
		return innerWidth, 0
	}
	if gutterWidth < 0 {
		gutterWidth = 0
	}
	rightWidth := calendarWidth
	maxRightWidth := max(34, innerWidth/2)
	if rightWidth > maxRightWidth {
		rightWidth = maxRightWidth
	}
	leftWidth := innerWidth - gutterWidth - rightWidth
	if leftWidth < 32 {
		leftWidth = 32
		rightWidth = max(0, innerWidth-gutterWidth-leftWidth)
	}
	return leftWidth, rightWidth
}

func MaxLineWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	return width
}

func ParseDate(raw string) (time.Time, bool) {
	return parseISODate(raw)
}

func ISOWeek(value time.Time) int {
	_, week := value.ISOWeek()
	return week
}

func ShiftDate(raw string, days int) string {
	parsed, ok := parseISODate(raw)
	if !ok {
		return raw
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func calendarMetaLine(theme types.Theme, selected time.Time, hasSelected bool, rangeStart, rangeEnd time.Time, hasRange bool, today time.Time) string {
	parts := []string{}
	if hasRange {
		parts = append(parts, fmt.Sprintf("Range W%02d-W%02d", ISOWeek(rangeStart), ISOWeek(rangeEnd)))
	} else if hasSelected {
		parts = append(parts, fmt.Sprintf("Week %02d", ISOWeek(selected)))
	}
	parts = append(parts, fmt.Sprintf("Today %02d", today.Day()), fmt.Sprintf("Wk %02d", ISOWeek(today)))
	return theme.StyleDim.Render(viewhelpers.Truncate(strings.Join(parts, "   "), 44))
}

func parseISODate(raw string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return time.Now().Format("2006-01-02")
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func weekIndexForDate(monthStart, selected time.Time) int {
	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)
	days := int(selected.Sub(gridStart).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days / 7
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
