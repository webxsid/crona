package daily

import (
	"fmt"
	"strings"
	"testing"

	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"

	"github.com/charmbracelet/x/ansi"
)

func TestDayStripShowsTotalsAndDate(t *testing.T) {
	state := dailySummaryTestState()
	lines := renderDayStrip(testTheme(), state, 160)
	row := ansi.Strip(lines[0])
	date, _ := viewcalendar.ParseDate("2026-05-27")
	for _, want := range []string{
		"Tasks 1/1",
		"2h/55m",
		"Habits 1/1",
		helperpkg.FormatDisplayDate("2026-05-27", nil),
		fmt.Sprintf("W%02d", viewcalendar.WeekNumber(date, state.WeekStart)),
	} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected strip row to contain %q, got %q", want, row)
		}
	}
}

func TestDayStripUsesDailySummaryWork(t *testing.T) {
	state := dailySummaryTestState()
	state.DailySummary.WorkedSeconds = 5400
	state.MetricsRollup.WorkedSeconds = 7200
	row := ansi.Strip(renderDayStrip(testTheme(), state, 160)[0])
	if !strings.Contains(row, "1h30m/55m") {
		t.Fatalf("expected the selected day's worked time, got %q", row)
	}
}

func TestDayStripSignalsTodayAndAverage(t *testing.T) {
	state := dailySummaryTestState()
	lines := renderDayStrip(testTheme(), state, 160)
	if len(lines) != 2 {
		t.Fatalf("expected a signals row, got %d rows", len(lines))
	}
	row := ansi.Strip(lines[1])
	for _, want := range []string{"Mood 4/5", "Energy 3/5", "Sleep 7h30m", "Focus 1d (best 3d)", "Check-ins 2d (best 5d)"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected signals row to contain %q, got %q", want, row)
		}
	}
	state.DailyCheckIn = nil
	row = ansi.Strip(renderDayStrip(testTheme(), state, 160)[1])
	for _, want := range []string{"Mood avg 3.1/5", "Energy avg 3.8/5", "Sleep avg 6h15m"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected averaged signals row to contain %q, got %q", want, row)
		}
	}
}

func TestDayStripOmitsSecondRowWithoutData(t *testing.T) {
	state := dailySummaryTestState()
	state.DailyCheckIn = nil
	state.MetricsRollup = nil
	state.DailyStreaks = nil
	if lines := renderDayStrip(testTheme(), state, 160); len(lines) != 1 {
		t.Fatalf("expected a single strip row without signals, got %d", len(lines))
	}
}

func TestDayStripFitsWidth(t *testing.T) {
	for _, width := range []int{60, 120, 285} {
		for i, line := range renderDayStrip(testTheme(), dailySummaryTestState(), width) {
			if got := ansi.StringWidth(line); got != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, got, ansi.Strip(line))
			}
		}
	}
}
