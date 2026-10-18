package charts

import (
	"math"
	"strings"
	"testing"
)

func TestColumnFillsFromBottom(t *testing.T) {
	for dots, want := range []string{"⠀", "⣀", "⣤", "⣶", "⣿"} {
		if got := Column(dots); got != want {
			t.Fatalf("%d dots: expected %q, got %q", dots, want, got)
		}
	}
}

func TestDotRowMarksOneHeight(t *testing.T) {
	for within, want := range []string{"⣀", "⠤", "⠒", "⠉"} {
		if got := DotRow(within); got != want {
			t.Fatalf("dot row %d: expected %q, got %q", within, want, got)
		}
	}
}

func TestSlotsCentreBarsAndWidenWhenFew(t *testing.T) {
	starts, width := Slots(3, 30)
	if width != 6 || starts[0] != 2 || starts[2] != 22 {
		t.Fatalf("expected 6-wide bars at 2 and 22, got %d %v", width, starts)
	}
	if _, width := Slots(40, 30); width != 1 {
		t.Fatalf("expected 1-wide bars when crowded, got %d", width)
	}
}

func TestLineJoinsPointsAndGapsMissingDays(t *testing.T) {
	// Three points at cells 1, 4 and 7 of 9; the middle one is missing.
	lines := Line([]float64{3, math.NaN(), 3}, []int{1, 4, 7}, 1, 5, 9, 2)
	if len(lines) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(lines))
	}
	row := []rune(strings.Join(lines, ""))
	if row[4] != ' ' && row[4] != '⠀' && []rune(lines[1])[4] != ' ' {
		t.Fatalf("expected a gap at the missing day, got %q", lines)
	}
	joined := Line([]float64{1, 5}, []int{1, 7}, 1, 5, 9, 2)
	top, bottom := []rune(joined[0]), []rune(joined[1])
	for col := 1; col <= 7; col++ {
		if top[col] == ' ' && bottom[col] == ' ' {
			t.Fatalf("expected the line unbroken from day to day, gap at %d in %q", col, joined)
		}
	}
}
