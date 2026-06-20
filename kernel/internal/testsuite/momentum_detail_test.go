package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestGetMomentumDetailReturnsCurrentBucketHabitContributors(t *testing.T) {
	ctx := context.Background()
	coreCtx, _ := newTestCoreContext(t, func() string { return "2026-06-19T12:00:00Z" })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Personal"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "wellbeing"})
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	habit, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Journal", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-06-19", sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
		t.Fatalf("complete habit: %v", err)
	}
	created, err := corecommands.CreateHabitStreakDefinition(ctx, coreCtx, sharedtypes.HabitStreakDefinition{
		Name:          "Daily journal",
		Enabled:       true,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
		MatchMode:     sharedtypes.MomentumMatchModeAny,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
		HabitIDs:      []int64{habit.ID},
	})
	if err != nil {
		t.Fatalf("create momentum: %v", err)
	}

	detail, err := corecommands.GetMomentumDetail(ctx, coreCtx, created.ID, "2026-06-19", 30)
	if err != nil {
		t.Fatalf("get momentum detail: %v", err)
	}
	if detail == nil {
		t.Fatal("expected momentum detail")
	}
	if detail.CurrentBucket.Count != 1 {
		t.Fatalf("expected current bucket count 1, got %+v", detail.CurrentBucket)
	}
	if len(detail.Contributors) != 1 {
		t.Fatalf("expected one contributor, got %+v", detail.Contributors)
	}
	if detail.Contributors[0].Kind != sharedtypes.MomentumContributorKindHabitCompletion {
		t.Fatalf("expected habit completion contributor, got %+v", detail.Contributors[0])
	}
}
