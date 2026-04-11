package app

import (
	"testing"

	sharedtypes "crona/shared/types"
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
