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

// scoreCellWidth is a day number and its gap.
const scoreCellWidth = 4

// ScoreMonthWidth is the width of the compact score calendar's widest line.
const ScoreMonthWidth = 7*scoreCellWidth - 2

// ScoreMonth is the Summary Calendar's month as a heat map: one line per week, each day's
// number coloured by its focus score (green 75+, yellow 45-74, red under 45), away days
// italic green, past days without a score dim and future days plain. The shown day is
// inverted and today underlined. A line under the month spells out the shown day's score,
// and a legend closes it.
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
		header = append(header, fmt.Sprintf("%2s", name))
	}
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")),
		scoreSummary(theme, selection, anchor),
		theme.StyleDim.Render(strings.Join(header, "  ")),
	}
	for week := range 6 {
		rowStart := gridStart.AddDate(0, 0, week*7)
		cells := make([]string, 0, 7)
		for day := range 7 {
			cells = append(cells, scoreCell(theme, selection, rowStart.AddDate(0, 0, day), anchor, today, monthStart))
		}
		lines = append(lines, strings.Join(cells, "  "))
	}
	return append(lines, scoreLegend(theme))
}

// scoreColor is the band colour for a focus score.
func scoreColor(theme types.Theme, score int) lipgloss.Color {
	switch {
	case score >= 75:
		return theme.ColorGreen
	case score < 45:
		return theme.ColorRed
	default:
		return theme.ColorYellow
	}
}

func scoreCell(theme types.Theme, selection ScoreSelection, date, anchor, today, month time.Time) string {
	iso := date.Format("2006-01-02")
	style := theme.StyleNormal
	score, scored := selection.Scores[iso]
	switch {
	case date.Month() != month.Month():
		style = theme.StyleDim
	case slices.Contains(selection.AwayDates, iso):
		style = lipgloss.NewStyle().Foreground(theme.ColorGreen).Italic(true)
	case date.After(today):
	case scored && score.HasData:
		style = lipgloss.NewStyle().Foreground(scoreColor(theme, score.Score)).Bold(true)
	default:
		style = theme.StyleDim
	}
	if sameDay(date, today) {
		style = style.Underline(true)
	}
	if sameDay(date, anchor) {
		style = style.Reverse(true)
	}
	return style.Render(fmt.Sprintf("%2d", date.Day()))
}

// scoreSummary names the shown day and its score, so the colours have a number to go by.
func scoreSummary(theme types.Theme, selection ScoreSelection, anchor time.Time) string {
	label := theme.StyleNormal.Render(anchor.Format("Mon 2"))
	iso := anchor.Format("2006-01-02")
	switch score, ok := selection.Scores[iso]; {
	case slices.Contains(selection.AwayDates, iso):
		return label + theme.StyleDim.Render(" · away")
	case ok && score.HasData:
		level := strings.ReplaceAll(string(score.Level), "_", " ")
		return label + theme.StyleDim.Render(" · focus ") +
			lipgloss.NewStyle().Foreground(scoreColor(theme, score.Score)).Bold(true).Render(fmt.Sprint(score.Score)) +
			theme.StyleDim.Render(" "+level)
	default:
		return label + theme.StyleDim.Render(" · no score yet")
	}
}

func scoreLegend(theme types.Theme) string {
	dot := func(color lipgloss.Color, text string) string {
		return lipgloss.NewStyle().Foreground(color).Render("■") + theme.StyleDim.Render(" "+text)
	}
	return strings.Join([]string{
		dot(theme.ColorGreen, "75+"), dot(theme.ColorYellow, "45+"), dot(theme.ColorRed, "<45"),
		lipgloss.NewStyle().Foreground(theme.ColorGreen).Italic(true).Render("away"),
	}, "  ")
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
