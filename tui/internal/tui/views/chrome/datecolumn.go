package viewchrome

import (
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// dateColumnWidth fits "Wed 30 Sep" and a gap.
const dateColumnWidth = 12

// DateColumnList draws the cursor's window of rows, one line each, with the date in a left
// column on each day's first row (and on the window's first row) and blank below it. The
// day's summary sits dim at the right end of its first row when there is room.
func DateColumnList(
	theme Theme,
	rows []DayRow,
	cursor, width, height int,
	summary func(date string) string,
) []string {
	if len(rows) == 0 || height < 1 {
		return nil
	}
	start, end := ListWindow(cursor, len(rows), height)
	inner := max(1, width-2)
	out := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		row := rows[i]
		first := i == 0 || row.Date != rows[i-1].Date
		date := ""
		if first || i == start {
			date = dateColumnLabel(row.Date)
		}
		line := theme.StyleDim.Render(padCellsRight(date, dateColumnWidth)) + row.Line
		if first && summary != nil {
			if tail := summary(row.Date); tail != "" {
				room := inner - ansi.StringWidth(tail) - 2
				if room >= dateColumnWidth+20 {
					line = ansi.Truncate(line, room, "…")
					line += strings.Repeat(" ", inner-ansi.StringWidth(line)-ansi.StringWidth(tail)) +
						theme.StyleDim.Render(tail)
				}
			}
		}
		line = ansi.Truncate(line, inner, "…")
		if i == cursor {
			out = append(out, SelectedRow(theme, line))
		} else {
			out = append(out, "  "+line)
		}
	}
	return out
}

func dateColumnLabel(date string) string {
	if day, err := time.Parse("2006-01-02", date); err == nil {
		return day.Format("Mon 2 Jan")
	}
	return date
}

func padCellsRight(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-ansi.StringWidth(text)))
}
