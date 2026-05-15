package model

import (
	"testing"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func TestDispatchMessageStatePreservesStashConflictDialogPayload(t *testing.T) {
	model := Model{width: 100}
	model = model.openStashConflictDialog(sharedtypes.StashConflict{
		IssueID: 42,
		Stashes: []sharedtypes.Stash{{
			ID:        "stash-1",
			CreatedAt: "2026-04-10T09:00:00Z",
		}},
	}, 7, 8, 42)

	roundTripped := model.applyDispatchMessageState(model.dispatchMessageState())
	if roundTripped.dialog != "stash_conflict" {
		t.Fatalf("expected stash conflict dialog, got %q", roundTripped.dialog)
	}
	if roundTripped.dialogDeleteID != "stash-1" {
		t.Fatalf("expected stash id to survive dispatch bridge, got %q", roundTripped.dialogDeleteID)
	}
	if roundTripped.dialogRepoID != 7 || roundTripped.dialogStreamID != 8 || roundTripped.dialogIssueID != 42 {
		t.Fatalf("expected issue path to survive dispatch bridge, got repo=%d stream=%d issue=%d", roundTripped.dialogRepoID, roundTripped.dialogStreamID, roundTripped.dialogIssueID)
	}
	if len(roundTripped.dialogChoiceValues) != 2 || roundTripped.dialogChoiceValues[0] != "resume" || roundTripped.dialogChoiceValues[1] != "continue" {
		t.Fatalf("expected choice values to survive dispatch bridge, got %#v", roundTripped.dialogChoiceValues)
	}
}

func TestTimerActivityTouchCmdOnlyForActiveTimerAndThrottles(t *testing.T) {
	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	model := Model{client: api.NewClient("unix", "/tmp/missing.sock", ""), timer: &api.TimerState{State: "idle"}}
	if cmd := model.timerActivityTouchCmd(now); cmd != nil {
		t.Fatalf("expected no touch command while idle")
	}

	model.timer = &api.TimerState{State: "running"}
	if cmd := model.timerActivityTouchCmd(now); cmd == nil {
		t.Fatalf("expected touch command for active timer")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(30 * time.Second)); cmd != nil {
		t.Fatalf("expected touch command to be throttled")
	}
	if cmd := model.timerActivityTouchCmd(now.Add(61 * time.Second)); cmd == nil {
		t.Fatalf("expected touch command after throttle window")
	}
}
