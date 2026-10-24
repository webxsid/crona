package model

import (
	"slices"
	"testing"

	dialogstate "crona/tui/internal/tui/dialogs/controller"
)

func TestTagDialogStateSurvivesTheModel(t *testing.T) {
	opened := dialogstate.OpenTagPicker(dialogstate.State{}, dialogstate.TagTargetHabit, 4, "Run", []string{"health"}, []string{"health", "quick"})
	opened.TagCursor = 1
	got := Model{}.withDialogState(opened).dialogState()
	if got.TagTarget != dialogstate.TagTargetHabit || got.TagCursor != 1 ||
		!slices.Equal(got.TagSelected, []string{"health"}) || !slices.Equal(got.TagVocab, []string{"health", "quick"}) {
		t.Fatalf("tag picker state lost: %+v", got)
	}
}
