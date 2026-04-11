//go:build e2e

package e2e

import (
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestSessionDetailAndAmendFlowOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Repo")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Session detail issue", nil)
	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusPlanned)

	var ctx sharedtypes.ActiveContext
	kernel.call(t, protocol.MethodContextSet, map[string]any{
		"repoId":   repo.ID,
		"streamId": stream.ID,
		"issueId":  issue.ID,
	}, &ctx)

	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{}, &timer)
	if timer.SessionID == nil || *timer.SessionID == "" {
		t.Fatalf("expected active session id, got %+v", timer)
	}

	errMessage := kernel.callError(t, protocol.MethodSessionAmendNote, shareddto.AmendSessionNoteRequest{
		ID:   timer.SessionID,
		Note: "should fail",
	})
	if !strings.Contains(errMessage, "cannot amend an active session") {
		t.Fatalf("expected active-session amend denial, got %q", errMessage)
	}

	kernel.call(t, protocol.MethodTimerEnd, shareddto.EndSessionRequest{
		CommitMessage: stringPtr("Initial commit"),
		Outcome:       stringPtr("Detailed outcome"),
		NextStep:      stringPtr("Detailed next step"),
	}, &timer)

	var history []sharedtypes.SessionHistoryEntry
	kernel.call(t, protocol.MethodSessionHistory, shareddto.SessionHistoryQuery{}, &history)
	if len(history) == 0 {
		t.Fatalf("expected session history")
	}

	var detail sharedtypes.SessionDetail
	kernel.call(t, protocol.MethodSessionDetail, shareddto.SessionIDRequest{ID: history[0].ID}, &detail)
	if detail.RepoName != "Repo" || detail.StreamName != "main" || detail.IssueTitle != "Session detail issue" {
		t.Fatalf("unexpected session detail context: %+v", detail)
	}
	if strings.TrimSpace(detail.ParsedNotes[sharedtypes.SessionNoteSectionCommit]) != "Initial commit" {
		t.Fatalf("expected initial commit in detail, got %+v", detail.ParsedNotes)
	}
	if !strings.Contains(detail.ParsedNotes[sharedtypes.SessionNoteSectionNotes], "Outcome: Detailed outcome") {
		t.Fatalf("expected outcome note to survive in detail, got %+v", detail.ParsedNotes)
	}

	var amended sharedtypes.Session
	kernel.call(t, protocol.MethodSessionAmendNote, shareddto.AmendSessionNoteRequest{
		ID:   stringPtr(history[0].ID),
		Note: "Updated commit",
	}, &amended)

	kernel.call(t, protocol.MethodSessionDetail, shareddto.SessionIDRequest{ID: history[0].ID}, &detail)
	if strings.TrimSpace(detail.ParsedNotes[sharedtypes.SessionNoteSectionCommit]) != "Updated commit" {
		t.Fatalf("expected amended commit in detail, got %+v", detail.ParsedNotes)
	}
	if !strings.Contains(detail.ParsedNotes[sharedtypes.SessionNoteSectionNotes], "Outcome: Detailed outcome") {
		t.Fatalf("expected non-commit notes preserved after amend, got %+v", detail.ParsedNotes)
	}
}
