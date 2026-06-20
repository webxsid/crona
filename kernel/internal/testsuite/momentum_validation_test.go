package testsuite

import (
	"context"
	"strings"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestCreateHabitMomentumRejectsInvalidDailyAllSelection(t *testing.T) {
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
	habitA, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Journal", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit A: %v", err)
	}
	habitB, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Walk", ScheduleType: string(sharedtypes.HabitScheduleWeekdays)})
	if err != nil {
		t.Fatalf("create habit B: %v", err)
	}

	_, err = corecommands.CreateHabitStreakDefinition(ctx, coreCtx, sharedtypes.HabitStreakDefinition{
		Name:          "Daily all",
		Enabled:       true,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
		MatchMode:     sharedtypes.MomentumMatchModeAll,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
		HabitIDs:      []int64{habitA.ID, habitB.ID},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "repeat daily") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHabitMomentumRejectsRequiredCountAboveCapacity(t *testing.T) {
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
	}{StreamID: stream.ID, Name: "Train", ScheduleType: string(sharedtypes.HabitScheduleWeekly), Weekdays: []int{1, 3, 5}})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	_, err = corecommands.CreateHabitStreakDefinition(ctx, coreCtx, sharedtypes.HabitStreakDefinition{
		Name:          "Weekly cap",
		Enabled:       true,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
		MatchMode:     sharedtypes.MomentumMatchModeAny,
		Period:        sharedtypes.HabitStreakPeriodWeek,
		RequiredCount: 4,
		HabitIDs:      []int64{habit.ID},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "maximum possible") {
		t.Fatalf("unexpected error: %v", err)
	}
}
