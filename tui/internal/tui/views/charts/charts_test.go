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

func TestBlockStepsInEighths(t *testing.T) {
	for eighths, want := range []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"} {
		if got := Block(eighths); got != want {
			t.Fatalf("%d eighths: expected %q, got %q", eighths, want, got)
		}
	}
}

func TestBarGridStacksBlocksAndMarksMissing(t *testing.T) {
	// Two rows give 16 eighths: 10 of 16 is a full bottom cell and a ▂ above it.
	grid := BarGrid([]float64{10, math.NaN()}, []int{0, 3}, 2, 5, 2, 16)
	if strings.Join(grid[0], "") != "▂▂   " || strings.Join(grid[1], "") != "██ · " {
		t.Fatalf("expected a stacked bar and a dot for the missing day, got %q", grid)
	}
}
