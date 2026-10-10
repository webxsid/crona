//go:build e2e

package e2e

import (
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestSessionEndRequiresAdhocAttribution(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)
	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{Adhoc: true}, &timer)
	summary := "Ad hoc work through session endpoint"
	message := kernel.callError(t, protocol.MethodSessionEnd, shareddto.EndSessionRequest{CommitMessage: &summary})
	if !strings.Contains(message, "keep the session ad hoc") {
		t.Fatalf("expected attribution rejection, got %q", message)
	}
	var active sharedtypes.Session
	kernel.call(t, protocol.MethodSessionGetActive, shareddto.Empty{}, &active)
	if timer.SessionID == nil || active.ID != *timer.SessionID || active.EndTime != nil {
		t.Fatalf("rejected request ended session: %+v", active)
	}
	var stopped sharedtypes.Session
	kernel.call(t, protocol.MethodSessionEnd, shareddto.EndSessionRequest{CommitMessage: &summary, AdhocResolution: sharedtypes.AdhocSessionResolutionKeep}, &stopped)
	if stopped.EndTime == nil {
		t.Fatal("attributed session did not end")
	}
	var issues []sharedtypes.IssueWithMeta
	kernel.call(t, protocol.MethodIssueListAll, shareddto.Empty{}, &issues)
	for _, issue := range issues {
		if issue.ID == stopped.IssueID && issue.Adhoc && issue.Title == summary {
			return
		}
	}
	t.Fatalf("completion did not create visible ad hoc issue: %+v", issues)
}
