package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestCustomMomentumSnapshotSeedsYesterdayWhenEmpty(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-10T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

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
	}{StreamID: stream.ID, Name: "Training", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	for _, date := range []string{"2026-04-01", "2026-04-03", "2026-04-08", "2026-04-09"} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit %s: %v", date, err)
		}
	}
	defs := []sharedtypes.HabitStreakDefinition{{
		ID:            "weekly",
		Name:          "Weekly training",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodWeek,
		RequiredCount: 2,
		HabitIDs:      []int64{habit.ID},
	}}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set streak defs: %v", err)
	}
	if err := corecommands.SeedCustomHabitMomentumSnapshot(ctx, coreCtx); err != nil {
		t.Fatalf("seed custom momentum snapshot: %v", err)
	}
	if got, err := coreCtx.CustomHabitMomentumSnapshots.GetByDate(ctx, coreCtx.UserID, "2026-04-09"); err != nil {
		t.Fatalf("load seeded snapshot: %v", err)
	} else if got == nil {
		t.Fatalf("expected a snapshot row for yesterday")
	}
	if got, err := coreCtx.CustomHabitMomentumSnapshots.GetByDate(ctx, coreCtx.UserID, "2026-04-10"); err != nil {
		t.Fatalf("load today snapshot: %v", err)
	} else if got != nil {
		t.Fatalf("did not expect a snapshot row for today on seed-only startup")
	}
}

func TestCustomMomentumSnapshotCarriesForwardAcrossDefinitionChange(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-10T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

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
	}{StreamID: stream.ID, Name: "Training", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	for _, date := range []string{
		"2026-02-03", "2026-02-05", "2026-02-10",
		"2026-03-03", "2026-03-05", "2026-03-10",
		"2026-04-01", "2026-04-03", "2026-04-08", "2026-04-09",
	} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit %s: %v", date, err)
		}
	}
	defs := []sharedtypes.HabitStreakDefinition{
		{
			ID:            "weekly",
			Name:          "Weekly training",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			RequiredCount: 2,
			HabitIDs:      []int64{habit.ID},
		},
		{
			ID:            "monthly",
			Name:          "Monthly training",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodMonth,
			RequiredCount: 3,
			HabitIDs:      []int64{habit.ID},
		},
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set streak defs: %v", err)
	}
	if err := corecommands.SeedCustomHabitMomentumSnapshot(ctx, coreCtx); err != nil {
		t.Fatalf("seed custom momentum snapshot: %v", err)
	}

	before, err := corecommands.ComputeMetricsLifetimeStreaks(ctx, coreCtx, "2026-04-09")
	if err != nil {
		t.Fatalf("compute lifetime streaks before change: %v", err)
	}
	if len(before.CustomHabitStreaks) != 2 {
		t.Fatalf("expected two custom streaks before change, got %+v", before.CustomHabitStreaks)
	}
	if before.CustomHabitStreaks[0].Current != 2 {
		t.Fatalf("expected weekly current streak of 2 before change, got %+v", before.CustomHabitStreaks[0])
	}

	defs[0].RequiredCount = 3
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("update streak defs: %v", err)
	}
	if err := corecommands.InvalidateCustomHabitMomentumSnapshotsFrom(ctx, coreCtx, "2026-04-10"); err != nil {
		t.Fatalf("invalidate snapshots from today: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-04-10", sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
		t.Fatalf("complete habit for new bucket: %v", err)
	}

	after, err := corecommands.ComputeMetricsLifetimeStreaks(ctx, coreCtx, "2026-04-10")
	if err != nil {
		t.Fatalf("compute lifetime streaks after change: %v", err)
	}
	if len(after.CustomHabitStreaks) != 2 {
		t.Fatalf("expected two custom streaks after change, got %+v", after.CustomHabitStreaks)
	}
	if after.CustomHabitStreaks[0].Current != 3 || after.CustomHabitStreaks[0].Longest != 3 {
		t.Fatalf("expected weekly streak to carry forward to 3 after change, got %+v", after.CustomHabitStreaks[0])
	}
	if after.CustomHabitStreaks[1].Current != 3 || after.CustomHabitStreaks[1].Longest != 3 {
		t.Fatalf("expected monthly streak to stay at 3 after weekly change, got %+v", after.CustomHabitStreaks[1])
	}
}
