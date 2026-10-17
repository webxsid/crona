package viewhelpers

import "testing"

func TestClassForBoundaries(t *testing.T) {
	cases := map[int]WidthClass{99: WidthNarrow, 100: WidthMedium, 159: WidthMedium, 160: WidthWide}
	for width, want := range cases {
		if got := ClassFor(width); got != want {
			t.Fatalf("ClassFor(%d) = %v, want %v", width, got, want)
		}
	}
}

func TestColumnsSumToTotal(t *testing.T) {
	cols := Columns(200, 3, 2, 2)
	if sum := cols[0] + cols[1] + cols[2]; sum != 194 || cols[0] < cols[1] {
		t.Fatalf("Columns(200,3,2,2) = %v", cols)
	}
	if got := Columns(100, 3, 2); got[0] != 59 || got[1] != 38 {
		t.Fatalf("Columns(100,3,2) = %v, want [59 38]", got)
	}
	for _, w := range Columns(5, 1, 1) {
		if w < 0 {
			t.Fatalf("Columns(5,1,1) has negative width: %v", Columns(5, 1, 1))
		}
	}
}
