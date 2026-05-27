package helpers

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func TestSessionHistorySummaryPrefixesManualEntries(t *testing.T) {
	entry := api.SessionHistoryEntry{
		Session: sharedtypes.Session{
			ID:      "s1",
			IssueID: 42,
			Source:  sharedtypes.SessionSourceManual,
		},
		ParsedNotes: sharedtypes.ParsedSessionNotes{
			sharedtypes.SessionNoteSectionCommit: "Manual catch-up",
		},
	}

	got := SessionHistorySummary(entry)
	if got != "[Manual] Manual catch-up" {
		t.Fatalf("unexpected summary %q", got)
	}
}

func TestFormatCompactDurationSeconds(t *testing.T) {
	if got := FormatCompactDurationSeconds(4500); got != "1h15m" {
		t.Fatalf("expected 1h15m, got %q", got)
	}
	if got := FormatCompactDurationSeconds(1500); got != "25m" {
		t.Fatalf("expected 25m, got %q", got)
	}
}

func TestDerivedTimerValuesPreferTimestamps(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	sessionStart := now.Add(-55 * time.Second).Format(time.RFC3339)
	segmentStart := now.Add(-25 * time.Second).Format(time.RFC3339)
	offset := 5

	timer := &api.TimerState{
		State:                       "running",
		SessionStartTime:            &sessionStart,
		SegmentStartTime:            &segmentStart,
		SegmentElapsedOffsetSeconds: &offset,
		ElapsedSeconds:              3,
		HardLimitActive:             true,
		HardLimitTotalSeconds:       300,
		HardLimitRemainingSeconds:   999,
	}

	if got := DerivedSegmentElapsedSeconds(timer, 7, now); got != 30 {
		t.Fatalf("expected timestamp-derived segment elapsed 30, got %d", got)
	}
	if got := DerivedHardLimitRemainingSeconds(timer, 7, now); got != 245 {
		t.Fatalf("expected timestamp-derived hard-limit remaining 245, got %d", got)
	}
}

func TestDerivedTimerValuesFallBackToKernelCounters(t *testing.T) {
	timer := &api.TimerState{
		State:                     "running",
		ElapsedSeconds:            12,
		HardLimitActive:           true,
		HardLimitTotalSeconds:     300,
		HardLimitRemainingSeconds: 120,
	}

	if got := DerivedSegmentElapsedSeconds(timer, 5, time.Now()); got != 17 {
		t.Fatalf("expected fallback segment elapsed 17, got %d", got)
	}
	if got := DerivedHardLimitRemainingSeconds(timer, 5, time.Now()); got != 115 {
		t.Fatalf("expected fallback hard-limit remaining 115, got %d", got)
	}
}
