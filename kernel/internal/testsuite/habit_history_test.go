package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestHabitHistoryScopesByRepoAndStream(t *testing.T) {
	ctx := context.Background()
	now := "2026-04-04T09:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return now })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	streamA, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create stream a: %v", err)
	}
	streamB, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "ops"})
	if err != nil {
		t.Fatalf("create stream b: %v", err)
	}
	habitA, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: streamA.ID, Name: "Inbox Zero", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit a: %v", err)
	}
	habitB, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: streamB.ID, Name: "Review Queue", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit b: %v", err)
	}

	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habitA.ID, "2026-04-04", sharedtypes.HabitCompletionStatusCompleted, ptrTo(25), nil); err != nil {
		t.Fatalf("complete habit a: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habitB.ID, "2026-04-03", sharedtypes.HabitCompletionStatusFailed, nil, nil); err != nil {
		t.Fatalf("complete habit b: %v", err)
	}

	allRows, err := corecommands.ListHabitHistory(ctx, coreCtx, &repo.ID, nil)
	if err != nil {
		t.Fatalf("list repo-scoped history: %v", err)
	}
	if len(allRows) != 2 {
		t.Fatalf("expected 2 repo-scoped rows, got %d", len(allRows))
	}
	if allRows[0].HabitName != "Inbox Zero" || allRows[0].StreamName != "app" || allRows[0].RepoName != "Work" {
		t.Fatalf("unexpected first repo-scoped row: %+v", allRows[0])
	}
	if allRows[1].HabitName != "Review Queue" || allRows[1].StreamName != "ops" || allRows[1].RepoName != "Work" {
		t.Fatalf("unexpected second repo-scoped row: %+v", allRows[1])
	}

	streamRows, err := corecommands.ListHabitHistory(ctx, coreCtx, nil, &streamA.ID)
	if err != nil {
		t.Fatalf("list stream-scoped history: %v", err)
	}
	if len(streamRows) != 1 {
		t.Fatalf("expected 1 stream-scoped row, got %d", len(streamRows))
	}
	if streamRows[0].HabitName != "Inbox Zero" || streamRows[0].StreamName != "app" {
		t.Fatalf("unexpected stream-scoped row: %+v", streamRows[0])
	}
}
