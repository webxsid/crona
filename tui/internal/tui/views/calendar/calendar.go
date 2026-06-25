package calendar

import (
	"fmt"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
	sharedtypes "crona/shared/types"
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
	WeekStart    sharedtypes.WeekStart
	Mode         Mode
}

type Mode string

const (
	ModeAuto  Mode = ""
	ModeMonth Mode = "month"
	ModeWeek  Mode = "week"
)

func Render(theme types.Theme, selection Selection) []string {
	anchor, ok := parseISODate(
		firstNonEmpty(
			selection.AnchorDate,
			selection.SelectedDate,
			selection.RangeEnd,
			selection.RangeStart,
		),
	)
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
	weekStart := sharedtypes.NormalizeWeekStart(selection.WeekStart)
	mode := selection.Mode
	if mode == ModeAuto {
		mode = ModeMonth
	}
	if mode == ModeWeek {
		return renderWeek(
			theme,
			selection,
			anchor,
			today,
			selected,
			rangeStart,
			rangeEnd,
			hasSelected,
			hasRangeStart,
			hasRangeEnd,
			weekStart,
		)
	}
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")),
		calendarMetaLine(
			theme,
			selected,
			hasSelected,
			rangeStart,
			rangeEnd,
			hasRangeStart && hasRangeEnd,
			today,
			weekStart,
		),
		theme.StyleDim.Render(weekHeader(weekStart)),
	}
	gridStart := shareddatefmt.StartOfWeek(monthStart, weekStart)
	selectedWeek := time.Time{}
	if hasSelected {
		selectedWeek = shareddatefmt.StartOfWeek(selected, weekStart)
	} else if hasRangeEnd {
		selectedWeek = shareddatefmt.StartOfWeek(rangeEnd, weekStart)
	}
	currentWeek := shareddatefmt.StartOfWeek(today, weekStart)
	selectedDateStyle := lipgloss.NewStyle().
		Background(theme.ColorGreen).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	todayStyle := lipgloss.NewStyle().
		Background(theme.ColorYellow).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	rangeStyle := lipgloss.NewStyle().Background(theme.ColorBlue).Foreground(theme.ColorWhite)
	currentWeekStyle := lipgloss.NewStyle().
		Background(theme.ColorYellow).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	for week := range 6 {
		rowStart := gridStart.AddDate(0, 0, week*7)
		rowWeek := WeekNumber(rowStart, weekStart)
		weekLabel := fmt.Sprintf("%2d", rowWeek)
		switch {
		case sameDay(rowStart, currentWeek):
			weekLabel = currentWeekStyle.Render(weekLabel)
		case sameDay(rowStart, selectedWeek):
			weekLabel = theme.StyleHeader.Render(weekLabel)
		default:
			weekLabel = theme.StyleDim.Render(weekLabel)
		}
		cells := make([]string, 0, 7)
		for day := range 7 {
			current := rowStart.AddDate(0, 0, day)
			inSelected := hasSelected && sameDay(current, selected)
			inRange := hasRangeStart && hasRangeEnd && !current.Before(rangeStart) &&
				!current.After(rangeEnd)
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
			case sameDay(rowStart, selectedWeek) && !selectedWeek.IsZero():
				cell = theme.StyleHeader.Render(cell)
			default:
				cell = theme.StyleNormal.Render(cell)
			}
			cells = append(cells, cell)
		}
		lines = append(lines, weekLabel+"  "+strings.Join(cells, " "))
	}
	return Window(lines, anchor, selection.MaxLines, weekStart)
}

func renderWeek(
	theme types.Theme,
	selection Selection,
	anchor, today, selected, rangeStart, rangeEnd time.Time,
	hasSelected, hasRangeStart, hasRangeEnd bool,
	weekStart sharedtypes.WeekStart,
) []string {
	rowStart := shareddatefmt.StartOfWeek(anchor, weekStart)
	rowWeek := WeekNumber(rowStart, weekStart)
	selectedDateStyle := lipgloss.NewStyle().
		Background(theme.ColorGreen).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	todayStyle := lipgloss.NewStyle().
		Background(theme.ColorYellow).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	rangeStyle := lipgloss.NewStyle().Background(theme.ColorBlue).Foreground(theme.ColorWhite)
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
		case current.Month() != anchor.Month():
			cell = theme.StyleDim.Render(cell)
		default:
			cell = theme.StyleNormal.Render(cell)
		}
		cells = append(cells, cell)
	}
	lines := []string{
		theme.StyleHeader.Render(anchor.Format("Jan 2006")),
		calendarMetaLine(
			theme,
			selected,
			hasSelected,
			rangeStart,
			rangeEnd,
			hasRangeStart && hasRangeEnd,
			today,
			weekStart,
		),
		theme.StyleDim.Render(compactWeekHeader(weekStart)),
		theme.StyleHeader.Render(fmt.Sprintf("W%02d", rowWeek)) + " " + strings.Join(cells, " "),
	}
	return Window(lines, anchor, selection.MaxLines, weekStart)
}

func Window(lines []string, anchor time.Time, maxLines int, weekStart sharedtypes.WeekStart) []string {
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
	selectedWeek := weekIndexForDate(monthStart, anchor, weekStart)
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

func ModeForWidth(innerWidth int) Mode {
	switch {
	case innerWidth >= 84:
		return ModeMonth
	case innerWidth >= 58:
		return ModeWeek
	default:
		return ModeAuto
	}
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
			right := lipgloss.NewStyle().
				Width(rightWidth).
				MaxWidth(rightWidth).
				Render(calendarLines[i-start])
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

func ColumnWidthsForMode(innerWidth, calendarWidth, gutterWidth int, mode Mode) (int, int) {
	if mode == ModeAuto || calendarWidth < 1 {
		return innerWidth, 0
	}
	switch mode {
	case ModeWeek:
		rightWidth := min(calendarWidth, max(22, innerWidth/3))
		leftWidth := innerWidth - gutterWidth - rightWidth
		if leftWidth < 44 {
			return innerWidth, 0
		}
		return leftWidth, rightWidth
	default:
		leftWidth, rightWidth := ColumnWidths(innerWidth, calendarWidth, gutterWidth)
		if leftWidth < 36 || rightWidth < 24 {
			return innerWidth, 0
		}
		return leftWidth, rightWidth
	}
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

func WeekNumber(value time.Time, weekStart sharedtypes.WeekStart) int {
	switch sharedtypes.NormalizeWeekStart(weekStart) {
	case sharedtypes.WeekStartSunday:
		return sundayWeekNumber(value)
	default:
		return ISOWeek(value)
	}
}

func ShiftDate(raw string, days int) string {
	parsed, ok := parseISODate(raw)
	if !ok {
		return raw
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func calendarMetaLine(
	theme types.Theme,
	selected time.Time,
	hasSelected bool,
	rangeStart, rangeEnd time.Time,
	hasRange bool,
	today time.Time,
	weekStart sharedtypes.WeekStart,
) string {
	parts := []string{}
	if hasRange {
		parts = append(
			parts,
			fmt.Sprintf(
				"Range W%02d-W%02d",
				WeekNumber(rangeStart, weekStart),
				WeekNumber(rangeEnd, weekStart),
			),
		)
	} else if hasSelected {
		parts = append(parts, fmt.Sprintf("Week %02d", WeekNumber(selected, weekStart)))
	}
	parts = append(
		parts,
		fmt.Sprintf("Today %02d", today.Day()),
		fmt.Sprintf("Wk %02d", WeekNumber(today, weekStart)),
	)
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

func weekIndexForDate(monthStart, selected time.Time, weekStart sharedtypes.WeekStart) int {
	gridStart := shareddatefmt.StartOfWeek(monthStart, weekStart)
	days := int(selected.Sub(gridStart).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days / 7
}

func sundayWeekNumber(value time.Time) int {
	yearStart := time.Date(value.Year(), 1, 1, 0, 0, 0, 0, value.Location())
	weekStart := shareddatefmt.StartOfWeek(yearStart, sharedtypes.WeekStartSunday)
	currentStart := shareddatefmt.StartOfWeek(value, sharedtypes.WeekStartSunday)
	if currentStart.Before(weekStart) {
		return 1
	}
	days := int(currentStart.Sub(weekStart).Hours() / 24)
	return days/7 + 1
}

func weekHeader(weekStart sharedtypes.WeekStart) string {
	switch sharedtypes.NormalizeWeekStart(weekStart) {
	case sharedtypes.WeekStartSunday:
		return "Wk  Su Mo Tu We Th Fr Sa"
	default:
		return "Wk  Mo Tu We Th Fr Sa Su"
	}
}

func compactWeekHeader(weekStart sharedtypes.WeekStart) string {
	switch sharedtypes.NormalizeWeekStart(weekStart) {
	case sharedtypes.WeekStartSunday:
		return "Su Mo Tu We Th Fr Sa"
	default:
		return "Mo Tu We Th Fr Sa Su"
	}
}
