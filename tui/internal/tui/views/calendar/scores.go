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

// ScoreGrid is the Summary Calendar as the C dialog draws it: a box per day holding its
// number and focus score (green 75+, yellow 45-74, red under 45), away days marked, today's
// box yellow and the shown day's cyan. Over the grid sit the month and the shown day's score,
// then the weekday names. week keeps only the shown day's week. It returns nil when even the
// narrowest boxes do not fit width.
func ScoreGrid(theme types.Theme, selection ScoreSelection, width int, week bool) []string {
	anchor, ok := parseISODate(selection.AnchorDate)
	if !ok {
		return nil
	}
	today, ok := parseISODate(selection.Today)
	if !ok {
		today, _ = parseISODate(time.Now().Format("2006-01-02"))
	}
	cellWidth := 0
	for _, candidate := range []int{7, 5, 4} {
		if weekColumn+7*(candidate+2)+6 <= width {
			cellWidth = candidate
			break
		}
	}
	if cellWidth == 0 {
		return nil
	}
	monthStart := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
	weekStart := sharedtypes.NormalizeWeekStart(selection.WeekStart)
	gridStart := shareddatefmt.StartOfWeek(monthStart, weekStart)
	header := []string{}
	for _, name := range strings.Fields(compactWeekHeader(weekStart)) {
		header = append(header, lipgloss.NewStyle().Width(cellWidth+2).PaddingLeft(2).Render(name))
	}
	lines := []string{
		theme.StyleHeader.Render(monthStart.Format("January 2006")) + "  " + scoreSummary(theme, selection, anchor),
		strings.Repeat(" ", weekColumn) + theme.StyleDim.Render(strings.Join(header, " ")),
	}
	first, last := 0, 5
	if week {
		first = min(5, max(0, weekIndexForDate(monthStart, anchor, weekStart)))
		last = first
	}
	for row := first; row <= last; row++ {
		rowStart := gridStart.AddDate(0, 0, row*7)
		cells := make([]string, 0, 13)
		for day := range 7 {
			if day > 0 {
				cells = append(cells, " ")
			}
			cells = append(cells, scoreBox(theme, selection, rowStart.AddDate(0, 0, day), anchor, today, monthStart, cellWidth))
		}
		lines = append(lines, WeekRowLines(theme.StyleDim.Render(WeekLabel(rowStart, weekStart)),
			strings.Split(lipgloss.JoinHorizontal(lipgloss.Top, cells...), "\n"))...)
	}
	return lines
}

// scoreBox is one day's box: its number over its score, "Away", "…" while a past day's
// score loads, or "—" for days without one.
func scoreBox(theme types.Theme, selection ScoreSelection, date, anchor, today, month time.Time, cellWidth int) string {
	iso := date.Format("2006-01-02")
	away := slices.Contains(selection.AwayDates, iso)
	future := date.After(today)
	dayStyle := theme.StyleNormal
	if date.Month() != month.Month() || future {
		dayStyle = theme.StyleDim
	}
	if away {
		dayStyle = dayStyle.Foreground(theme.ColorGreen)
	}
	if sameDay(date, anchor) {
		dayStyle = theme.StyleCursor
	}
	value := theme.StyleDim.Render("—")
	switch score, ok := selection.Scores[iso]; {
	case away:
		value = lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(AwayLabel(cellWidth - 2))
	case ok && score.HasData && !future:
		value = lipgloss.NewStyle().Foreground(scoreColor(theme, score.Score)).Bold(true).Render(fmt.Sprintf("%3d", score.Score))
	case !future && date.Month() == month.Month():
		value = theme.StyleDim.Render("…")
	}
	border := theme.ColorDim
	if sameDay(date, today) {
		border = theme.ColorYellow
	}
	if sameDay(date, anchor) {
		border = theme.ColorCyan
	}
	return lipgloss.NewStyle().
		Width(cellWidth).
		Height(2).
		PaddingLeft(1).
		PaddingRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Render(dayStyle.Render(fmt.Sprintf("%-2d", date.Day())) + "\n" + value)
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

// weekColumn is the week-number column before a boxed week: W45 and a gap.
const weekColumn = 4

// WeekRowLines puts a week's label beside its boxed row, on the day-number line, and pads
// the box's other lines to match.
func WeekRowLines(label string, row []string) []string {
	out := make([]string, len(row))
	for i, line := range row {
		cell := strings.Repeat(" ", weekColumn)
		if i == 1 {
			cell = label + " "
		}
		out[i] = cell + line
	}
	return out
}

// AwayLabel is "Away" cut to fit a box's inner width (its width less the side padding), so
// it never wraps: Away, Awy, then A.
func AwayLabel(inner int) string {
	switch {
	case inner >= 4:
		return "Away"
	case inner == 3:
		return "Awy"
	}
	return "A"
}
