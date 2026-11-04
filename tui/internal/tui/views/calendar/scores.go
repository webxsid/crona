package calendar

import (
	"fmt"
	"slices"
	"strings"
	"time"

	shareddatefmt "crona/shared/datefmt"
	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// ScoreSelection is what the compact score calendar shows: the anchor's month, with each
// past day's focus score.
type ScoreSelection struct {
	AnchorDate string
	Today      string
	WeekStart  sharedtypes.WeekStart
	AwayDates  []string
	Scores     map[string]sharedtypes.FocusScoreRangeDay
}

// scoreCellWidth fits "dd sss": the day, a space and a three-cell score.
const scoreCellWidth = 6

// ScoreMonthWidth is the width of the compact score calendar's widest line.
const ScoreMonthWidth = 7*scoreCellWidth + 6

// ScoreMonth is the Summary Calendar's month in one line per week: each day's number, then
// its focus score coloured as the dialog colours it (green 75+, red under 45, yellow
// between), Away on away days, and nothing for the future. The anchor day is bold in the
// accent and today is underlined.
func ScoreMonth(theme types.Theme, selection ScoreSelection) []string {
	anchor, ok := parseISODate(selection.AnchorDate)
	if !ok {
		return nil
	}
	today, ok := parseISODate(selection.Today)
	if !ok {
		today, _ = parseISODate(time.Now().Format("2006-01-02"))
	}
	monthStart := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
	weekStart := sharedtypes.NormalizeWeekStart(selection.WeekStart)
	gridStart := shareddatefmt.StartOfWeek(monthStart, weekStart)
	header := []string{}
	for _, name := range strings.Fields(compactWeekHeader(weekStart)) {
		header = append(header, fmt.Sprintf("%-*s", scoreCellWidth, name))
	}
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")),
		theme.StyleDim.Render(strings.TrimRight(strings.Join(header, " "), " ")),
	}
	for week := range 6 {
		rowStart := gridStart.AddDate(0, 0, week*7)
		cells := make([]string, 0, 7)
		for day := range 7 {
			cells = append(cells, scoreCell(theme, selection, rowStart.AddDate(0, 0, day), anchor, today, monthStart))
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return lines
}

func scoreCell(theme types.Theme, selection ScoreSelection, date, anchor, today, month time.Time) string {
	iso := date.Format("2006-01-02")
	future := date.After(today)
	dayStyle := theme.StyleNormal
	if date.Month() != month.Month() || future {
		dayStyle = theme.StyleDim
	}
	if sameDay(date, today) {
		dayStyle = dayStyle.Underline(true)
	}
	if sameDay(date, anchor) {
		dayStyle = theme.StyleCursor.Underline(sameDay(date, today))
	}
	day := dayStyle.Render(fmt.Sprintf("%2d", date.Day()))
	value := "   "
	switch score, ok := selection.Scores[iso]; {
	case slices.Contains(selection.AwayDates, iso):
		value = lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("awy")
	case future:
	case ok && score.HasData:
		color := theme.ColorYellow
		if score.Score >= 75 {
			color = theme.ColorGreen
		} else if score.Score < 45 {
			color = theme.ColorRed
		}
		value = lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%3d", score.Score))
	case date.Month() == month.Month():
		value = theme.StyleDim.Render("  ·")
	}
	return day + " " + value
}

// ScoreWeekIndex is which of ScoreMonth's week rows holds the anchor date.
func ScoreWeekIndex(selection ScoreSelection) int {
	anchor, ok := parseISODate(selection.AnchorDate)
	if !ok {
		return 0
	}
	monthStart := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
	return min(5, max(0, weekIndexForDate(monthStart, anchor, sharedtypes.NormalizeWeekStart(selection.WeekStart))))
}
