package e2e

import (
	"encoding/json"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestTimerStashRestoreFlowOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repoA := createRepo(t, kernel, "Repo A")
	streamA := createStream(t, kernel, repoA.ID, "main")
	issueA := createIssue(t, kernel, streamA.ID, "Issue A", nil)

	repoB := createRepo(t, kernel, "Repo B")
	streamB := createStream(t, kernel, repoB.ID, "dev")
	issueB := createIssue(t, kernel, streamB.ID, "Issue B", nil)
	changeIssueStatus(t, kernel, issueA.ID, sharedtypes.IssueStatusPlanned)
	changeIssueStatus(t, kernel, issueB.ID, sharedtypes.IssueStatusPlanned)

	var ctx sharedtypes.ActiveContext
	kernel.call(t, protocol.MethodContextSet, map[string]any{
		"repoId":   repoA.ID,
		"streamId": streamA.ID,
		"issueId":  issueA.ID,
	}, &ctx)
	if ctx.RepoName == nil || *ctx.RepoName != "Repo A" {
		t.Fatalf("expected resolved repo name in context, got %+v", ctx)
	}
	if ctx.StreamName == nil || *ctx.StreamName != "main" {
		t.Fatalf("expected resolved stream name in context, got %+v", ctx)
	}
	if ctx.IssueTitle == nil || *ctx.IssueTitle != "Issue A" {
		t.Fatalf("expected resolved issue title in context, got %+v", ctx)
	}

	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{}, &timer)
	if timer.State != "running" || timer.IssueID == nil || *timer.IssueID != issueA.ID {
		t.Fatalf("unexpected timer state after start: %+v", timer)
	}

	kernel.call(t, protocol.MethodTimerPause, nil, &timer)
	if timer.State != "paused" {
		t.Fatalf("expected paused timer, got %+v", timer)
	}

	var stash sharedtypes.Stash
	kernel.call(t, protocol.MethodStashPush, shareddto.CreateStashRequest{
		StashNote: stringPtr("Switching tasks"),
	}, &stash)
	if stash.IssueID == nil || *stash.IssueID != issueA.ID {
		t.Fatalf("unexpected stash payload: %+v", stash)
	}

	kernel.call(t, protocol.MethodTimerGetState, nil, &timer)
	if timer.State != "idle" {
		t.Fatalf("expected idle timer after stash, got %+v", timer)
	}

	kernel.call(t, protocol.MethodContextSet, map[string]any{
		"repoId":   repoB.ID,
		"streamId": streamB.ID,
		"issueId":  issueB.ID,
	}, &ctx)
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{}, &timer)
	if timer.IssueID == nil || *timer.IssueID != issueB.ID {
		t.Fatalf("expected second issue to run, got %+v", timer)
	}

	errMessage := kernel.callError(t, protocol.MethodStashApply, shareddto.StashIDRequest{ID: stash.ID})
	if errMessage != "cannot apply stash while a focus session is active" {
		t.Fatalf("expected active-session stash apply denial, got %q", errMessage)
	}

	kernel.call(t, protocol.MethodTimerEnd, shareddto.EndSessionRequest{
		CommitMessage: stringPtr("Finished issue B"),
	}, &timer)
	if timer.State != "idle" {
		t.Fatalf("expected idle timer after end, got %+v", timer)
	}

	var ok shareddto.OKResponse
	kernel.call(t, protocol.MethodStashApply, shareddto.StashIDRequest{ID: stash.ID}, &ok)
	if !ok.OK {
		t.Fatalf("expected stash apply ok")
	}

	kernel.call(t, protocol.MethodContextGet, nil, &ctx)
	if ctx.IssueID == nil || *ctx.IssueID != issueA.ID {
		t.Fatalf("expected original issue context restored, got %+v", ctx)
	}

	kernel.call(t, protocol.MethodTimerGetState, nil, &timer)
	if timer.State != "paused" && timer.State != "running" {
		t.Fatalf("expected restored timer to be active, got %+v", timer)
	}
	if timer.IssueID == nil || *timer.IssueID != issueA.ID {
		t.Fatalf("expected restored issue to resume immediately, got %+v", timer)
	}
}

func TestTimerStartReturnsStashConflictForMatchingIssue(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Repo A")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Issue A", nil)
	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusPlanned)

	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: &issue.ID}, &timer)
	kernel.call(t, protocol.MethodTimerPause, nil, &timer)

	var stash sharedtypes.Stash
	kernel.call(t, protocol.MethodStashPush, shareddto.CreateStashRequest{StashNote: stringPtr("Paused work")}, &stash)

	resp := kernel.callResponse(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: &issue.ID})
	if resp.Error == nil {
		t.Fatalf("expected stash conflict response")
	}
	if resp.Error.Code != protocol.ErrorCodeStashConflict {
		t.Fatalf("expected stash conflict code, got %+v", resp.Error)
	}
	var conflict sharedtypes.StashConflict
	if err := json.Unmarshal(resp.Error.Data, &conflict); err != nil {
		t.Fatalf("decode stash conflict data: %v", err)
	}
	if conflict.IssueID != issue.ID || len(conflict.Stashes) != 1 || conflict.Stashes[0].ID != stash.ID {
		t.Fatalf("unexpected stash conflict payload: %+v", conflict)
	}

	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: &issue.ID, IgnoreExistingStashes: true}, &timer)
	if timer.State != "running" || timer.IssueID == nil || *timer.IssueID != issue.ID {
		t.Fatalf("expected timer to start with stash override, got %+v", timer)
	}
}

func stringPtr(value string) *string {
	return &value
}
