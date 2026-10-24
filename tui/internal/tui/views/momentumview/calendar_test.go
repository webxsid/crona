package momentumview

import (
	"strings"
	"testing"
	"time"

	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// twoWeeks runs Mon 5 to Sun 18 Oct 2026 (W41, W42); Mondays are missed, all else done.
func twoWeeks() []string {
	series := dailyMomentumSeries(
		time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		[]int{0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1},
		1,
	)
	return calendarLines(types.Theme{}, series, 70, 20, true, "")
}

func TestCalendarFillsHeightWithinWidth(t *testing.T) {
	lines := twoWeeks()
	if len(lines) != 20 {
		t.Fatalf("expected 20 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 70 {
			t.Fatalf("expected lines within 70 cells, got %d: %q", w, ansi.Strip(line))
		}
	}
}

func TestCalendarLaysWeeksAsRows(t *testing.T) {
	text := ansi.Strip(strings.Join(twoWeeks(), "\n"))
	if !strings.Contains(squash(text), "Mon Tue Wed Thu Fri Sat Sun") {
		t.Fatalf("expected weekday columns, got\n%s", text)
	}
	if w41, w42 := strings.Index(text, "W41"), strings.Index(text, "W42"); w41 < 0 || w42 < w41 {
		t.Fatalf("expected W41 above W42, got\n%s", text)
	}
}

func TestCalendarCellsGrowWithRoom(t *testing.T) {
	if text := ansi.Strip(strings.Join(twoWeeks(), "\n")); !strings.Contains(text, "██████") {
		t.Fatalf("expected wide cells with room to spare, got\n%s", text)
	}
}

func TestCalendarCountsHitsPerWeekday(t *testing.T) {
	hit := squash(lineContaining(ansi.Strip(strings.Join(twoWeeks(), "\n")), "hit"))
	if !strings.HasPrefix(hit, "hit 0/2 2/2") {
		t.Fatalf("expected Mondays 0/2 and Tuesdays 2/2, got %q", hit)
	}
}

func TestCalendarDropsOldestWeeksWhenShort(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), make([]int, 70), 1)
	text := ansi.Strip(strings.Join(calendarLines(types.Theme{}, series, 70, 6, true, ""), "\n"))
	if strings.Contains(text, "W33") || !strings.Contains(text, "W42") {
		t.Fatalf("expected the latest weeks only, got\n%s", text)
	}
}

func TestCalendarHitRowSitsUnderTheGrid(t *testing.T) {
	lines := twoWeeks()
	for i, line := range lines {
		if strings.HasPrefix(ansi.Strip(line), "hit") {
			if i == len(lines)-1 || ansi.Strip(lines[i-1]) == "" {
				t.Fatalf(
					"expected the hit row right under the grid with room below, got line %d of %d",
					i,
					len(lines),
				)
			}
			return
		}
	}
	t.Fatalf("expected a hit row")
}

func TestCalendarCellsUseTheWholePane(t *testing.T) {
	series := dailyMomentumSeries(
		time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		[]int{
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
			1,
		},
		1,
	)
	text := ansi.Strip(strings.Join(calendarLines(types.Theme{}, series, 83, 40, true, ""), "\n"))
	if !strings.Contains(text, strings.Repeat("█", 10)) {
		t.Fatalf("expected 10-cell squares in an 83-cell pane, got\n%s", text)
	}
	if rows := strings.Count(text, "W38  "+strings.Repeat("█", 10)); rows != 1 {
		t.Fatalf("expected one labelled row per week, got %d", rows)
	}
}

func TestCalendarMarksToday(t *testing.T) {
	series := dailyMomentumSeries(
		time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		[]int{0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1},
		1,
	)
	lines := calendarLines(types.Theme{}, series, 70, 20, true, "2026-10-17")
	header := ansi.Strip(lines[0])
	if !strings.Contains(header, "Sat•") || strings.Contains(header, "Fri•") {
		t.Fatalf("expected Saturday marked in the header, got %q", header)
	}
	marked := 0
	for _, line := range lines[1:] {
		marked += strings.Count(ansi.Strip(line), "•")
	}
	if marked != 1 {
		t.Fatalf(
			"expected exactly one marked cell, got %d in\n%s",
			marked,
			ansi.Strip(strings.Join(lines, "\n")),
		)
	}
}
