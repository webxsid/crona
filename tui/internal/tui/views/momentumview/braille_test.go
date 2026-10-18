package momentumview

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestBrailleColumnFillsFromBottom(t *testing.T) {
	for dots, want := range []string{"⠀", "⣀", "⣤", "⣶", "⣿"} {
		if got := brailleColumn(dots); got != want {
			t.Fatalf("%d dots: expected %q, got %q", dots, want, got)
		}
	}
}

func barDef() sharedtypes.HabitStreakDefinition {
	return sharedtypes.HabitStreakDefinition{Name: "m", Enabled: true, Period: sharedtypes.HabitStreakPeriodDay, RequiredCount: 4}
}

func TestBarsFillTheGivenHeight(t *testing.T) {
	for _, height := range []int{6, 12, 20} {
		lines := renderMomentumBars(types.Theme{}, barDef(), points(4, 1, 2, 3, 4, 5), 50, height, true)
		if len(lines) != height {
			t.Fatalf("expected %d lines, got %d", height, len(lines))
		}
		for _, line := range lines {
			if w := ansi.StringWidth(line); w > 50 {
				t.Fatalf("expected lines within 50 cells, got %d: %q", w, ansi.Strip(line))
			}
		}
	}
}

func TestBarsUseQuarterCellSteps(t *testing.T) {
	// 8 chart rows give 32 dots; a count of 1 against an axis of 8 is 4 dots, one full cell.
	lines := renderMomentumBars(types.Theme{}, barDef(), points(8, 1, 8), 40, 10, true)
	chart := ansi.Strip(strings.Join(lines[:8], "\n"))
	if strings.Count(chart, "⣿") < 9 {
		t.Fatalf("expected a full-height bar and a one-cell bar in braille, got\n%s", chart)
	}
	if strings.Contains(chart, "█") {
		t.Fatalf("expected braille bars, not blocks, got\n%s", chart)
	}
}

func TestBarsGroupCrowdedBucketsByWorstStatus(t *testing.T) {
	series := points(4, 4, 0, 4, 4)
	groups := groupBuckets(series, 2, true)
	if len(groups) != 2 || groups[0].status != "missed" || groups[1].status != "met" || groups[0].count != 2 {
		t.Fatalf("expected [missed avg 2, met], got %+v", groups)
	}
	lines := renderMomentumBars(types.Theme{}, barDef(), points(4, make([]int, 90)...), 40, 10, true)
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 40 {
			t.Fatalf("expected 90 buckets squeezed within 40 cells, got %d", w)
		}
	}
}

func TestBarsDrawTargetAsDottedLine(t *testing.T) {
	chart := ansi.Strip(strings.Join(renderMomentumBars(types.Theme{}, barDef(), points(4, 0, 0), 40, 10, true), "\n"))
	if !strings.ContainsAny(chart, "⠉⠒⠤⣀") {
		t.Fatalf("expected a braille target line, got\n%s", chart)
	}
}

func TestBarsWidenWhenBucketsAreFew(t *testing.T) {
	chart := ansi.Strip(strings.Join(renderMomentumBars(types.Theme{}, barDef(), points(4, 4, 4, 4), 40, 10, true), "\n"))
	if !strings.Contains(chart, "⣿⣿⣿") {
		t.Fatalf("expected few buckets to draw wide bars, got\n%s", chart)
	}
}

func TestBarsKeepTargetLineAcrossSkippedBuckets(t *testing.T) {
	series := points(4, 0, 0, 0)
	series[1].Skipped = true
	series[1].Target = 0
	chart := ansi.Strip(strings.Join(renderMomentumBars(types.Theme{}, barDef(), series, 40, 10, true), "\n"))
	if !strings.ContainsAny(chart, "⠉⠒⠤⣀") {
		t.Fatalf("expected the target line despite a skipped bucket, got\n%s", chart)
	}
}
