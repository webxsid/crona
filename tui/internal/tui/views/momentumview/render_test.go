package momentumview

import (
	"regexp"
	"strings"
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func renderState(width, height int, cards ...sharedtypes.MomentumCard) types.ContentState {
	return types.ContentState{
		View: "momentum", Pane: "momentum_cards", Width: width, Height: height,
		MomentumDate: "2026-10-18", MomentumWindowDays: 28,
		MomentumCards: cards,
		Cursors:       map[string]int{"momentum_cards": 0},
		Filters:       map[string]string{},
	}
}

func plain(s string) string { return ansi.Strip(s) }

func assertFits(t *testing.T, rendered string, width, height int) {
	t.Helper()
	lines := strings.Split(rendered, "\n")
	if len(lines) > height {
		t.Fatalf("expected at most %d lines, got %d", height, len(lines))
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > width {
			t.Fatalf("expected lines within %d cells, got %d: %q", width, w, plain(line))
		}
	}
}

func TestRenderNarrowListOnly(t *testing.T) {
	rendered := Render(types.Theme{}, renderState(60, 20, momentumTestCard("Deep work", "W1", "W2")))
	assertFits(t, rendered, 60, 20)
	text := plain(rendered)
	if strings.Contains(text, "Cadence") || !strings.Contains(text, "[enter]") {
		t.Fatalf("expected the list alone with an enter hint, got\n%s", text)
	}
}

func TestRenderMediumAndWideShowDetail(t *testing.T) {
	for _, size := range [][2]int{{120, 36}, {200, 50}} {
		rendered := Render(types.Theme{}, renderState(size[0], size[1], momentumTestCard("Deep work", "W1", "W2"), momentumTestCard("Reading", "W1")))
		assertFits(t, rendered, size[0], size[1])
		text := plain(rendered)
		for _, want := range []string{"Cadence", "Target", "Streak", "Deep work"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%v: expected %q in\n%s", size, want, text)
			}
		}
	}
}

func TestRenderInactiveDimmedAndPaused(t *testing.T) {
	off := momentumTestCard("Piano", "W1")
	off.Definition.Enabled = false
	state := renderState(120, 36, off)
	text := plain(Render(types.Theme{}, state))
	if !strings.Contains(text, "paused") || !strings.Contains(text, "[x] enable") {
		t.Fatalf("expected a paused row and an enable key, got\n%s", text)
	}
	if text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1")))); !strings.Contains(text, "[x] disable") {
		t.Fatalf("expected a disable key, got\n%s", text)
	}
}

func TestRenderListCreateLineAndEmpty(t *testing.T) {
	if text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1")))); !strings.Contains(text, "[a] new momentum") {
		t.Fatalf("expected a create line, got\n%s", text)
	}
	if text := squash(plain(Render(types.Theme{}, renderState(120, 36)))); !strings.Contains(text, "No momentums yet [a] new momentum") {
		t.Fatalf("expected the empty state, got\n%s", text)
	}
}

func TestRenderNoBordersOrHeader(t *testing.T) {
	text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1"))))
	if strings.Contains(text, "╭") || regexp.MustCompile(`Momentum\s+2026-`).MatchString(text) {
		t.Fatalf("expected no box or dated header, got\n%s", text)
	}
}

func TestRenderTinyHeight(t *testing.T) {
	for _, height := range []int{1, 4} {
		assertFits(t, Render(types.Theme{}, renderState(120, height, momentumTestCard("Deep work", "W1"))), 120, height)
	}
}

func TestRenderUsesFilteredRows(t *testing.T) {
	state := renderState(120, 36, momentumTestCard("Deep work", "W1"), momentumTestCard("Reading", "W1", "W2"))
	state.Filters["momentum_cards"] = "read"
	text := plain(Render(types.Theme{}, state))
	if strings.Contains(text, "Deep work") || !strings.Contains(text, "Reading") {
		t.Fatalf("expected only the filtered momentum, got\n%s", text)
	}
}

func TestDetailLinesWrapDescription(t *testing.T) {
	card := momentumTestCard("Deep work", "W1")
	description := strings.Repeat("steady focused work ", 4)
	card.Definition.Description = &description
	for _, line := range DetailLines(types.Theme{}, card, 30, 40, false) {
		if w := ansi.StringWidth(line); w > 30 {
			t.Fatalf("expected lines within 30 cells, got %q", plain(line))
		}
	}
	if !strings.Contains(squash(plain(strings.Join(DetailLines(types.Theme{}, card, 30, 40, false), " "))), strings.TrimSpace(description)) {
		t.Fatalf("expected the whole description")
	}
}

func TestRenderWideListFitsLongerNames(t *testing.T) {
	text := plain(Render(types.Theme{}, renderState(200, 50, momentumTestCard("Daily Context Any", "W1"), momentumTestCard("Daily Context All", "W1"))))
	for _, want := range []string{"▶ Daily Context Any ", "  Daily Context All "} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected list row %q whole, got\n%s", want, text)
		}
	}
}

func labelledCard(name string, labels ...string) sharedtypes.MomentumCard {
	card := momentumTestCard(name, labels...)
	card.Definition.Period = sharedtypes.HabitStreakPeriodDay
	return card
}

func TestDetailHeaderPairsSettings(t *testing.T) {
	lines := DetailLines(types.Theme{}, momentumTestCard("Deep work", "W1"), 80, 30, false)
	if !strings.Contains(plain(lines[0]), "Cadence") || !strings.Contains(plain(lines[0]), "Habits") {
		t.Fatalf("expected settings side by side, got %q", plain(lines[0]))
	}
}

func TestChartFillsTheDetailHeight(t *testing.T) {
	lines := DetailLines(types.Theme{}, labelledCard("Deep work", "Oct 1", "Oct 2", "Oct 3"), 80, 30, false)
	if len(lines) != 30 {
		t.Fatalf("expected the chart to fill 30 lines, got %d", len(lines))
	}
	if axis := plain(lines[len(lines)-3]); !strings.Contains(axis, "└") {
		t.Fatalf("expected the chart axis just above the dates and footnote, got %q", axis)
	}
}

func TestMediumStacksBucketsUnderChart(t *testing.T) {
	text := plain(Render(types.Theme{}, renderState(120, 36, labelledCard("Deep work", "Oct 1", "Oct 2", "Oct 3"))))
	chart, buckets := strings.Index(text, "└"), strings.Index(text, "Buckets")
	if chart < 0 || buckets < chart {
		t.Fatalf("expected Buckets below the chart, got\n%s", text)
	}
}

func TestWideGivesBucketsTheirOwnColumn(t *testing.T) {
	lines := strings.Split(plain(Render(types.Theme{}, renderState(200, 50, labelledCard("Deep work", "Oct 1", "Oct 2")))), "\n")
	title := lineContaining(strings.Join(lines, "\n"), "Momentums")
	if !strings.Contains(title, "Deep work") || !strings.Contains(title, "Buckets") {
		t.Fatalf("expected list, detail and buckets side by side, got %q", title)
	}
}

func TestPairedHeaderKeepsLongTargets(t *testing.T) {
	card := momentumTestCard("Deep work", "W1")
	card.Definition.TargetKind = sharedtypes.MomentumTargetKindContext
	text := plain(strings.Join(DetailLines(types.Theme{}, card, 72, 30, false), "\n"))
	if strings.Contains(text, "…") || strings.Contains(text, "...") {
		t.Fatalf("expected no truncated settings at 72 cells, got\n%s", text)
	}
}

// dailyCard has one daily bucket per count, from Sun 11 Oct 2026, against a target of 4.
func dailyCard(counts ...int) sharedtypes.MomentumCard {
	card := labelledCard("Deep work")
	card.Series = dailyMomentumSeries(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC), counts, 4)
	return card
}

func TestBucketLinesOpenWithSummary(t *testing.T) {
	head := plain(BucketLines(types.Theme{}, dailyCard(4, 2, 6), 60)[0])
	if head != "target 4 · met 2 of 3 · avg 4" {
		t.Fatalf("expected the summary line, got %q", head)
	}
}

func TestBucketRowsUseMarksWeekdaysAndBareCounts(t *testing.T) {
	text := plain(strings.Join(BucketLines(types.Theme{}, dailyCard(4, 2, 6), 60), "\n"))
	for _, want := range []string{"✓ Tue 13  6 ", "✗ Mon 12  2 ", "✓ Sun 11  4 "} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "/4") || strings.Contains(text, "missed") {
		t.Fatalf("expected no repeated target or status words, got\n%s", text)
	}
	if strings.Index(text, "Tue 13") > strings.Index(text, "Sun 11") {
		t.Fatalf("expected the newest bucket first, got\n%s", text)
	}
}

func TestBucketBarsFillTheColumn(t *testing.T) {
	for _, line := range BucketLines(types.Theme{}, dailyCard(4, 2, 6), 60)[1:] {
		if w := ansi.StringWidth(line); w != 60 && !strings.HasPrefix(plain(line), "──") {
			t.Fatalf("expected rows to fill 60 cells, got %d: %q", w, plain(line))
		}
	}
}

func TestBucketLinesBreakWeeks(t *testing.T) {
	text := plain(strings.Join(BucketLines(types.Theme{}, dailyCard(4, 2, 6), 60), "\n"))
	w41, w42 := strings.Index(text, "W41"), strings.Index(text, "W42")
	mon, sun := strings.Index(text, "Mon 12"), strings.Index(text, "Sun 11")
	if w42 < 0 || w41 < 0 || !(w42 < mon && mon < w41 && w41 < sun) {
		t.Fatalf("expected W42 over Mon 12 and W41 over Sun 11, got\n%s", text)
	}
}

func TestBucketLinesShowExceptions(t *testing.T) {
	card := dailyCard(4, 2, 6)
	card.Series[0].Skipped = true
	card.Series[0].ProtectedDayCount = 1
	card.Series[1].Target, card.Series[1].OriginalTarget = 3, 4
	text := plain(strings.Join(BucketLines(types.Theme{}, card, 60), "\n"))
	for _, want := range []string{"· Sun 11  skip  protected 1d", "Mon 12  2/3"} {
		if !strings.Contains(squash(text), squash(want)) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
}

func TestBucketLinesWithoutWeeksForWeeklyBuckets(t *testing.T) {
	text := plain(strings.Join(BucketLines(types.Theme{}, momentumTestCard("Deep work", "W1", "W2"), 60), "\n"))
	if strings.Contains(text, "──") {
		t.Fatalf("expected no week rules for weekly buckets, got\n%s", text)
	}
}
