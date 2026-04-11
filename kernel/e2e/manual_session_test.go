//go:build e2e

package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestManualSessionLoggingOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Repo")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Missed timer issue", nil)

	var created sharedtypes.Session
	kernel.call(t, protocol.MethodSessionLogManual, shareddto.ManualSessionLogRequest{
		IssueID:              issue.ID,
		Date:                 "2026-03-28",
		WorkDurationSeconds:  1800,
		BreakDurationSeconds: 600,
		StartTime:            stringPtr("09:00"),
		EndTime:              stringPtr("09:40"),
		CommitMessage:        stringPtr("Manual catch-up"),
		Notes:                stringPtr("Forgot to start timer"),
	}, &created)

	if created.Source != sharedtypes.SessionSourceManual {
		t.Fatalf("expected manual source, got %+v", created)
	}

	var history []sharedtypes.SessionHistoryEntry
	kernel.call(t, protocol.MethodSessionHistory, shareddto.SessionHistoryQuery{}, &history)
	if len(history) != 1 {
		t.Fatalf("expected 1 session in history, got %d", len(history))
	}
	if history[0].Source != sharedtypes.SessionSourceManual {
		t.Fatalf("expected history source manual, got %+v", history[0])
	}

	var detail sharedtypes.SessionDetail
	kernel.call(t, protocol.MethodSessionDetail, shareddto.SessionIDRequest{ID: history[0].ID}, &detail)
	if detail.WorkSummary.WorkSeconds != 1800 || detail.WorkSummary.RestSeconds != 600 {
		t.Fatalf("unexpected work summary %+v", detail.WorkSummary)
	}

	var metrics []sharedtypes.DailyMetricsDay
	kernel.call(t, protocol.MethodMetricsRange, shareddto.DateRangeQuery{Start: "2026-03-28", End: "2026-03-28"}, &metrics)
	if len(metrics) != 1 {
		t.Fatalf("expected metrics day for manual session, got %+v", metrics)
	}
	if metrics[0].WorkedSeconds != 1800 || metrics[0].RestSeconds != 600 || metrics[0].SessionCount != 1 {
		t.Fatalf("unexpected metrics %+v", metrics[0])
	}
}
