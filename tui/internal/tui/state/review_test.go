package state

import "testing"

func TestReviewRangeKeepsAnchorWhileDateInside(t *testing.T) {
	start, end := ReviewRange("2026-10-10", "2026-10-14", 7)
	if start != "2026-10-08" || end != "2026-10-14" {
		t.Fatalf("expected the anchored window, got %s..%s", start, end)
	}
}

func TestReviewRangeReanchorsWhenDateLeaves(t *testing.T) {
	for _, date := range []string{"2026-10-15", "2026-10-07"} {
		start, end := ReviewRange(date, "2026-10-14", 7)
		if end != date || start != shiftDate(date, -6) {
			t.Fatalf("%s: expected the window to end on the date, got %s..%s", date, start, end)
		}
	}
}

func TestReviewRangeDefaults(t *testing.T) {
	start, end := ReviewRange("2026-10-14", "", 0)
	if start != "2026-10-08" || end != "2026-10-14" {
		t.Fatalf("expected a 7-day window on the date, got %s..%s", start, end)
	}
}
