package daily

import (
	"fmt"
	"strings"
	"testing"

	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"

	"github.com/charmbracelet/x/ansi"
)

func stripRows(width int) []string {
	lines := renderDayStrip(testTheme(), dailySummaryTestState(), width)
	for i, line := range lines {
		lines[i] = ansi.Strip(line)
	}
	return lines
}

// fieldsOf collapses runs of spaces so assertions ignore column padding.
func fieldsOf(row string) string {
	return strings.Join(strings.Fields(row), " ")
}

// cellIndex is the display column of needle in row, or -1.
func cellIndex(row, needle string) int {
	idx := strings.Index(row, needle)
	if idx < 0 {
		return -1
	}
	return ansi.StringWidth(row[:idx])
}

func TestDayStripStacksTasksOverHabits(t *testing.T) {
	rows := stripRows(160)
	if got := fieldsOf(rows[0]); !strings.HasPrefix(got, "Tasks 1/1") || !strings.Contains(got, "2h/55m") {
		t.Fatalf("expected the Tasks meter on row 0, got %q", rows[0])
	}
	if got := fieldsOf(rows[1]); !strings.HasPrefix(got, "Habits 1/1") {
		t.Fatalf("expected the Habits meter on row 1, got %q", rows[1])
	}
	if cellIndex(rows[0], "1/1") != cellIndex(rows[1], "1/1") {
		t.Fatalf("expected meter values in one column, got %q / %q", rows[0], rows[1])
	}
}

func TestDayStripAlignsSignalsAndStreaksInColumns(t *testing.T) {
	rows := stripRows(160)
	if len(rows) != 3 {
		t.Fatalf("expected 3 strip rows, got %d: %q", len(rows), rows)
	}
	for _, col := range [][]string{{"Mood", "Energy", "Sleep"}, {"Focus", "Check-ins"}} {
		at := cellIndex(rows[0], col[0])
		for i, label := range col {
			if got := cellIndex(rows[i], label); got != at || at < 0 {
				t.Fatalf("expected %q at column %d on row %d, got %d: %q", label, at, i, got, rows)
			}
		}
	}
	for i, want := range []string{"Mood 4/5 Focus 1d best 3d", "Energy 3/5 Check-ins 2d best 5d", "Sleep 7h30m"} {
		if !strings.Contains(fieldsOf(rows[i]), want) {
			t.Fatalf("expected row %d to contain %q, got %q", i, want, rows[i])
		}
	}
}

func TestDayStripSignalsAverageWithoutCheckIn(t *testing.T) {
	state := dailySummaryTestState()
	state.DailyCheckIn = nil
	rows := renderDayStrip(testTheme(), state, 160)
	for i, want := range []string{"Mood avg 3.1/5", "Energy avg 3.8/5", "Sleep avg 6h15m"} {
		if !strings.Contains(fieldsOf(ansi.Strip(rows[i])), want) {
			t.Fatalf("expected row %d to contain %q, got %q", i, want, ansi.Strip(rows[i]))
		}
	}
}

func TestDayStripIsTwoRowsWithoutSignals(t *testing.T) {
	state := dailySummaryTestState()
	state.DailyCheckIn = nil
	state.MetricsRollup = nil
	state.DailyStreaks = nil
	if lines := renderDayStrip(testTheme(), state, 160); len(lines) != 2 {
		t.Fatalf("expected only the two meter rows, got %d", len(lines))
	}
}

func TestDayStripLeavesDateToSubTabLine(t *testing.T) {
	for _, row := range stripRows(285) {
		if strings.Contains(row, helperpkg.FormatDisplayDate("2026-05-27", nil)) {
			t.Fatalf("expected no date in the strip, got %q", row)
		}
	}
}

func TestDayStripDropsStreaksThenSignalsWhenNarrow(t *testing.T) {
	medium := strings.Join(stripRows(72), "\n")
	if strings.Contains(medium, "Check-ins") || !strings.Contains(medium, "Mood") {
		t.Fatalf("expected streaks dropped before signals at 72, got %q", medium)
	}
	if narrow := stripRows(48); len(narrow) != 2 || strings.Contains(strings.Join(narrow, ""), "Mood") {
		t.Fatalf("expected meters only at 48, got %q", narrow)
	}
}

func TestDayStripFitsWidth(t *testing.T) {
	for _, width := range []int{40, 60, 80, 120, 285} {
		for i, line := range renderDayStrip(testTheme(), dailySummaryTestState(), width) {
			if got := ansi.StringWidth(line); got != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, got, ansi.Strip(line))
			}
		}
	}
}

func TestSubTabMetaShowsDateWeekAndKeys(t *testing.T) {
	state := dailySummaryTestState()
	date, _ := viewcalendar.ParseDate("2026-05-27")
	got := ansi.Strip(SubTabMeta(testTheme(), state))
	for _, want := range []string{
		helperpkg.FormatDisplayDate("2026-05-27", nil),
		fmt.Sprintf("W%02d", viewcalendar.WeekNumber(date, state.WeekStart)),
		",/. g",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected sub-tab meta to contain %q, got %q", want, got)
		}
	}
}

func TestDayStripShowsCurrentWeekAtRightEnd(t *testing.T) {
	rows := stripRows(160)
	for i, want := range []string{"Mo Tu We Th Fr Sa Su", "25 26 27 28 29 30 31"} {
		row := strings.TrimRight(rows[i], " ")
		if !strings.HasSuffix(row, want) || ansi.StringWidth(row) != 159 {
			t.Fatalf("expected row %d to end with %q one cell from the edge, got %q", i, want, rows[i])
		}
	}
}

func TestDayStripKeepsWeekOverStreaks(t *testing.T) {
	joined := strings.Join(stripRows(100), "\n")
	if !strings.Contains(joined, "Mo Tu We") || strings.Contains(joined, "Check-ins") {
		t.Fatalf("expected the week kept and streaks dropped at 100, got %q", joined)
	}
}

func TestDayStripDropsWeekWhenNarrowOrOneRow(t *testing.T) {
	if joined := strings.Join(stripRows(60), "\n"); strings.Contains(joined, "Mo Tu") {
		t.Fatalf("expected no week at 60 columns, got %q", joined)
	}
	if rows := dayStripRows(testTheme(), dailySummaryTestState(), 160, 1); len(rows) != 1 || strings.Contains(ansi.Strip(rows[0]), "Mo Tu") {
		t.Fatalf("expected a one-row strip without the week, got %q", rows)
	}
}
