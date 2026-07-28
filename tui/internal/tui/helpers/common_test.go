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

func TestDerivedSegmentElapsedPrefersTimestampAndHardLimitUsesSnapshot(t *testing.T) {
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
	if got := DerivedHardLimitRemainingSeconds(timer, 7, now); got != 992 {
		t.Fatalf("expected snapshot-derived hard-limit remaining 992, got %d", got)
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

func TestExtendedCountdownUsesOverallRemainingInsteadOfOriginalSegment(t *testing.T) {
	segment := sharedtypes.SessionSegmentWork
	timer := &api.TimerState{
		State:                     "running",
		SegmentType:               &segment,
		HardLimitActive:           true,
		HardLimitKind:             sharedtypes.TimerHardLimitKindCountdown,
		HardLimitTotalSeconds:     20 * 60,
		HardLimitRemainingSeconds: 5 * 60,
		HardLimitWorkSeconds:      15 * 60,
	}

	remaining, gotSegment, ok := DerivedHardLimitSegmentRemainingSeconds(
		timer,
		0,
		time.Now(),
	)
	if !ok || gotSegment == nil || *gotSegment != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected an active countdown work segment, got %v %v", gotSegment, ok)
	}
	if remaining != 5*60 {
		t.Fatalf("expected extended countdown to display five minutes, got %d", remaining)
	}
}
