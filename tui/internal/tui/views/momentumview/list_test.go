package momentumview

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func points(target int, counts ...int) []sharedtypes.MomentumSeriesPoint {
	series := make([]sharedtypes.MomentumSeriesPoint, len(counts))
	for i, count := range counts {
		series[i] = sharedtypes.MomentumSeriesPoint{Count: count, Target: target}
	}
	return series
}

func cardWith(name string, enabled bool, current, longest int, series []sharedtypes.MomentumSeriesPoint) sharedtypes.MomentumCard {
	return sharedtypes.MomentumCard{
		Definition: sharedtypes.HabitStreakDefinition{Name: name, Enabled: enabled, Period: sharedtypes.HabitStreakPeriodDay, RequiredCount: 1},
		Current:    current, Longest: longest, Series: series,
	}
}

func TestSparklineScalesByTarget(t *testing.T) {
	if got := ansi.Strip(sparkline(types.Theme{}, points(4, 0, 1, 2, 4), true)); got != "    ▁▃▅█" {
		t.Fatalf("expected %q, got %q", "    ▁▃▅█", got)
	}
}

func TestSparklineSkippedAndZeroTarget(t *testing.T) {
	series := points(0, 0, 0)
	series[0].Skipped = true
	if got := ansi.Strip(sparkline(types.Theme{}, series, true)); got != "      ·█" {
		t.Fatalf("expected a skipped dot and a full zero-target cell, got %q", got)
	}
}

func TestSparklineEmptySeries(t *testing.T) {
	if got := sparkline(types.Theme{}, nil, true); got != strings.Repeat(" ", 8) {
		t.Fatalf("expected 8 blanks, got %q", got)
	}
}

func TestSparklineKeepsLastEight(t *testing.T) {
	if got := ansi.Strip(sparkline(types.Theme{}, points(1, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1), true)); got != "████████" {
		t.Fatalf("expected the last 8 buckets, got %q", got)
	}
}

func TestStatusMarks(t *testing.T) {
	skipped := points(4, 4)
	skipped[0].Skipped = true
	cases := []struct {
		card sharedtypes.MomentumCard
		want string
	}{
		{cardWith("m", true, 0, 0, points(4, 4)), "✓"},
		{cardWith("n", true, 0, 0, points(4, 3)), "◐"},
		{cardWith("x", true, 0, 0, points(4, 1)), "✗"},
		{cardWith("s", true, 0, 0, skipped), "·"},
		{cardWith("p", false, 0, 0, points(4, 4)), "paused"},
		{cardWith("e", true, 0, 0, nil), ""},
	}
	for _, c := range cases {
		if got := ansi.Strip(statusMark(types.Theme{}, c.card)); got != c.want {
			t.Fatalf("%s: expected %q, got %q", c.card.Definition.Name, c.want, got)
		}
	}
}

func TestListRowFitsWidth(t *testing.T) {
	row := listRow(types.Theme{}, cardWith("A remarkably long momentum name that will not fit", true, 27, 30, points(1, 1, 1)), 38)
	plain := ansi.Strip(row)
	if w := ansi.StringWidth(row); w != 38 {
		t.Fatalf("expected 38 cells, got %d: %q", w, plain)
	}
	for _, want := range []string{"…", "██", "27d", "✓"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected %q in %q", want, plain)
		}
	}
}

func stripState(cards ...sharedtypes.MomentumCard) types.ContentState {
	return types.ContentState{MomentumCards: cards}
}

func TestStripRowsCounts(t *testing.T) {
	state := stripState(
		cardWith("Deep work", true, 27, 30, points(4, 4)),
		cardWith("Reading", true, 3, 40, points(4, 3)),
		cardWith("Gym", true, 1, 5, points(4, 0)),
		cardWith("Piano", false, 0, 12, points(4, 4)),
	)
	rows := StripRows(types.Theme{}, state, 120)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for _, row := range rows {
		if w := ansi.StringWidth(row); w != 120 {
			t.Fatalf("expected 120 cells, got %d", w)
		}
	}
	first, second := squash(ansi.Strip(rows[0])), squash(ansi.Strip(rows[1]))
	for _, want := range []string{"Active 3 inactive 1", "Latest met 1 near 1 missed 1"} {
		if !strings.Contains(first, want) {
			t.Fatalf("expected %q in row 1 %q", want, first)
		}
	}
	for _, want := range []string{"Best streak Deep work 27d", "Longest ever Reading 40d"} {
		if !strings.Contains(second, want) {
			t.Fatalf("expected %q in row 2 %q", want, second)
		}
	}
}

func TestStripRowsDropRightFirst(t *testing.T) {
	rows := StripRows(types.Theme{}, stripState(cardWith("Deep work", true, 27, 30, points(4, 4))), 30)
	for _, row := range rows {
		if w := ansi.StringWidth(row); w != 30 || strings.Contains(row, "Latest") {
			t.Fatalf("expected a 30-cell row without Latest, got %d %q", w, ansi.Strip(row))
		}
	}
}

func TestStripRowsEmpty(t *testing.T) {
	rows := StripRows(types.Theme{}, stripState(), 120)
	if !strings.Contains(rows[0], "Active 0") || !strings.Contains(rows[1], "–") {
		t.Fatalf("expected zero counts and dashes, got %q", rows)
	}
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }
