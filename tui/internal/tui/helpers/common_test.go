package helpers

import (
	"testing"

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
