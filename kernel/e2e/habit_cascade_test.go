package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestStreamAndRepoDeleteCascadeHabits(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Work")
	stream := createStream(t, kernel, repo.ID, "app")
	createHabit(t, kernel, stream.ID, "Inbox Zero")

	var ok shareddto.OKResponse
	kernel.call(t, protocol.MethodStreamDelete, shareddto.NumericIDRequest{ID: stream.ID}, &ok)
	if !ok.OK {
		t.Fatalf("expected stream delete ok response")
	}

	var habits []sharedtypes.Habit
	kernel.call(t, protocol.MethodHabitList, shareddto.ListHabitsQuery{StreamID: stream.ID}, &habits)
	if len(habits) != 0 {
		t.Fatalf("expected no habits after stream delete, got %+v", habits)
	}

	stream2 := createStream(t, kernel, repo.ID, "ops")
	createHabit(t, kernel, stream2.ID, "Review Notes")
	kernel.call(t, protocol.MethodRepoDelete, shareddto.NumericIDRequest{ID: repo.ID}, &ok)
	if !ok.OK {
		t.Fatalf("expected repo delete ok response")
	}

	kernel.call(t, protocol.MethodHabitList, shareddto.ListHabitsQuery{StreamID: stream2.ID}, &habits)
	if len(habits) != 0 {
		t.Fatalf("expected no habits after repo delete, got %+v", habits)
	}
}
