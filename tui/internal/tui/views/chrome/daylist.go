package viewchrome

import (
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// DayRow is one selectable list row and the ISO date it is grouped under.
type DayRow struct {
	Date, Line string
}

// dayLine is a display line: a row, or a divider when row is -1.
type dayLine struct {
	text string
	row  int
}

// DayList draws the cursor's window of rows, at most height lines, with a dim divider
// before each date's first row. Rows keep their order; dividers are never selectable, and
// the divider of the window's first row is kept in view while the cursor stays visible.
func DayList(
	theme Theme,
	rows []DayRow,
	cursor, width, height int,
	summary func(date string) string,
) []string {
	lines := make([]dayLine, 0, len(rows)+8)
	for i, row := range rows {
		if i == 0 || row.Date != rows[i-1].Date {
			lines = append(
				lines,
				dayLine{text: dayDivider(theme, row.Date, summary(row.Date), width), row: -1},
			)
		}
		lines = append(lines, dayLine{text: row.Line, row: i})
	}
	return dividedList(theme, lines, cursor, width, height)
}

// dividedList windows rows and dividers around the cursor's row, keeping the divider
// above the window's first row in view while the cursor stays visible.
func dividedList(theme Theme, lines []dayLine, cursor, width, height int) []string {
	if len(lines) == 0 || height < 1 {
		return nil
	}
	cursorLine := 0
	for i, line := range lines {
		if line.row == cursor {
			cursorLine = i
		}
	}
	start, end := ListWindow(cursorLine, len(lines), height)
	if start > 0 && lines[start-1].row < 0 && cursorLine < start-1+height {
		start--
		end = min(len(lines), start+height)
	}
	out := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		switch {
		case line.row < 0:
			out = append(out, line.text)
		case line.row == cursor:
			out = append(
				out,
				theme.StyleCursor.Render(
					SelectionCursor+" "+ansi.Truncate(line.text, max(1, width-2), "…"),
				),
			)
		default:
			out = append(out, "  "+ansi.Truncate(line.text, max(1, width-2), "…"))
		}
	}
	return out
}

func dayDivider(theme Theme, date, summary string, width int) string {
	label := date
	if day, err := time.Parse("2006-01-02", date); err == nil {
		label = day.Format("Mon 2 Jan")
	}
	text := "── " + label
	if summary != "" {
		text += " · " + summary
	}
	text = ansi.Truncate(text+" ", width, "")
	return theme.StyleDim.Render(text + strings.Repeat("─", max(0, width-ansi.StringWidth(text))))
}

// StripRows packs [top, bottom] column groups into two rows of exactly width cells,
// dropping groups from the right that do not fit.
func StripRows(groups [][2]string, width int) []string {
	rows := []string{"", ""}
	if width <= 0 {
		return rows
	}
	for _, group := range groups {
		columnWidth := max(ansi.StringWidth(group[0]), ansi.StringWidth(group[1]))
		gap := ""
		if rows[0] != "" {
			gap = "   "
		}
		if rows[0] != "" && ansi.StringWidth(rows[0])+len(gap)+columnWidth > width {
			break
		}
		for i := range rows {
			rows[i] += gap + group[i] + strings.Repeat(" ", columnWidth-ansi.StringWidth(group[i]))
		}
	}
	for i := range rows {
		rows[i] = ansi.Truncate(rows[i], width, "")
		rows[i] += strings.Repeat(" ", width-ansi.StringWidth(rows[i]))
	}
	return rows
}

// LogPaneWidths splits a Log sub-tab into list and detail widths; detail is 0 when narrow.
func LogPaneWidths(width int) (list, detail int) {
	switch {
	case width >= 160:
		detail = max(44, (width-6)*2/5)
	case width >= 100:
		detail = max(40, (width-3)*2/5)
	default:
		return width, 0
	}
	return width - detail - 3, detail
}
