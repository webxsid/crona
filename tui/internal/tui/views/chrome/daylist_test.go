package viewchrome

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func dayRows(dates []string, perDay int) []DayRow {
	rows := []DayRow{}
	for _, date := range dates {
		for i := range perDay {
			rows = append(rows, DayRow{Date: date, Line: fmt.Sprintf("%s row %d", date, i)})
		}
	}
	return rows
}

func countSummary(string) string { return "summary" }

func TestDayListDividesDays(t *testing.T) {
	rows := dayRows([]string{"2026-10-17", "2026-10-16", "2026-10-15"}, 2)
	lines := DayList(Theme{}, rows, 0, 50, 20, countSummary)
	text := ansi.Strip(strings.Join(lines, "\n"))
	for _, day := range []string{"Sat 17 Oct", "Fri 16 Oct", "Thu 15 Oct"} {
		divider := strings.Index(text, "── "+day+" · summary ")
		if divider < 0 {
			t.Fatalf("expected a divider for %s, got\n%s", day, text)
		}
	}
	if strings.Index(text, "Fri 16 Oct") > strings.Index(text, "2026-10-16 row 0") {
		t.Fatalf("expected each divider before its day's rows, got\n%s", text)
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 50 {
			t.Fatalf("expected lines within 50 cells, got %d", w)
		}
	}
}

func TestDayListKeepsCursorInView(t *testing.T) {
	rows := dayRows([]string{"2026-10-17", "2026-10-16", "2026-10-15", "2026-10-14"}, 10)
	lines := DayList(Theme{}, rows, 39, 50, 10, countSummary)
	text := ansi.Strip(strings.Join(lines, "\n"))
	if len(lines) > 10 || !strings.Contains(text, SelectionCursor+" 2026-10-14 row 9") {
		t.Fatalf("expected the cursor row in 10 lines, got\n%s", text)
	}
	rows = dayRows([]string{"2026-10-17", "2026-10-16"}, 10)
	text = ansi.Strip(strings.Join(DayList(Theme{}, rows, 10, 50, 10, countSummary), "\n"))
	if !strings.Contains(text, "Fri 16 Oct") {
		t.Fatalf("expected the cursor's day divider above its first row, got\n%s", text)
	}
}

func TestDayListEmpty(t *testing.T) {
	if lines := DayList(Theme{}, nil, 0, 50, 10, countSummary); len(lines) != 0 {
		t.Fatalf("expected no lines, got %q", lines)
	}
}

func TestStripRowsPackAndDrop(t *testing.T) {
	rows := StripRows([][2]string{{"Sessions 3", "This week 2h"}, {"total 1h40m  avg 33m", "longest 1h"}}, 30)
	for _, row := range rows {
		if ansi.StringWidth(row) != 30 || strings.Contains(row, "total") {
			t.Fatalf("expected 30-cell rows keeping the first group only, got %q", rows)
		}
	}
}
