package commands

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestMomentumBucketLabelPrefixesWeeklyRangesWithWeekNumber(t *testing.T) {
	got := momentumBucketLabel(
		"2026-W23",
		sharedtypes.HabitStreakPeriodWeek,
		"2026-06-01",
		"2026-06-07",
	)
	if got != "[23] Jun 1-7" {
		t.Fatalf("expected same-month weekly label with week prefix, got %q", got)
	}
}

func TestMomentumBucketLabelKeepsWeekPrefixAcrossMonthBoundary(t *testing.T) {
	got := momentumBucketLabel(
		"2026-W18",
		sharedtypes.HabitStreakPeriodWeek,
		"2026-04-27",
		"2026-05-03",
	)
	if got != "[18] Apr 27-May 3" {
		t.Fatalf("expected cross-month weekly label with week prefix, got %q", got)
	}
}
