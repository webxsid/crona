//go:build e2e

package e2e

import (
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestFocusStartRequiresEligibleIssueStatusAndSessionEndDoesNotCompleteIssue(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Repo")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Lifecycle issue", nil)

	var ctx sharedtypes.ActiveContext
	kernel.call(t, protocol.MethodContextSet, map[string]any{
		"repoId":   repo.ID,
		"streamId": stream.ID,
		"issueId":  issue.ID,
	}, &ctx)

	errMessage := kernel.callError(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{})
	if !strings.Contains(errMessage, "focus sessions cannot be started") {
		t.Fatalf("expected focus eligibility error, got %q", errMessage)
	}

	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusPlanned)

	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{}, &timer)
	if timer.State != "running" {
		t.Fatalf("expected running timer, got %+v", timer)
	}

	errMessage = kernel.callError(t, protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issue.ID,
		Status: sharedtypes.IssueStatusDone,
	})
	if !strings.Contains(errMessage, "cannot change issue status while a focus session is active") {
		t.Fatalf("expected active-session status change denial, got %q", errMessage)
	}

	issues := listIssues(t, kernel, stream.ID)
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Status != sharedtypes.IssueStatusInProgress {
		t.Fatalf("expected issue to auto-move to in_progress, got %s", issues[0].Status)
	}

	kernel.call(t, protocol.MethodTimerEnd, shareddto.EndSessionRequest{
		CommitMessage: stringPtr("session ended"),
		Outcome:       stringPtr("Kernel lifecycle migrated"),
		NextStep:      stringPtr("Polish footer and CRUD"),
	}, &timer)
	if timer.State != "idle" {
		t.Fatalf("expected idle timer after end, got %+v", timer)
	}

	issues = listIssues(t, kernel, stream.ID)
	if issues[0].Status != sharedtypes.IssueStatusInProgress {
		t.Fatalf("expected issue to remain in_progress after session end, got %s", issues[0].Status)
	}

	var history []sharedtypes.SessionHistoryEntry
	kernel.call(t, protocol.MethodSessionHistory, shareddto.SessionHistoryQuery{}, &history)
	if len(history) == 0 {
		t.Fatalf("expected session history entry after timer end")
	}
	notes := history[0].ParsedNotes[sharedtypes.SessionNoteSectionNotes]
	if !strings.Contains(notes, "Outcome: Kernel lifecycle migrated") {
		t.Fatalf("expected structured outcome note, got %q", notes)
	}
	if !strings.Contains(notes, "Next step: Polish footer and CRUD") {
		t.Fatalf("expected structured next-step note, got %q", notes)
	}

	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusDone)
	errMessage = kernel.callError(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{})
	if !strings.Contains(errMessage, "focus sessions cannot be started") {
		t.Fatalf("expected done issue focus denial, got %q", errMessage)
	}

	issue2 := createIssue(t, kernel, stream.ID, "Blocked lifecycle issue", nil)
	changeIssueStatus(t, kernel, issue2.ID, sharedtypes.IssueStatusPlanned)
	errMessage = kernel.callError(t, protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issue2.ID,
		Status: sharedtypes.IssueStatusAbandoned,
	})
	if !strings.Contains(errMessage, "requires a reason") {
		t.Fatalf("expected abandoned reason validation, got %q", errMessage)
	}
}
