package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestHabitMetricsCountDueCompletedAndFailedDays(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-01T09:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
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
	}{StreamID: stream.ID, Name: "Focus", ScheduleType: string(sharedtypes.HabitScheduleDaily), TargetMinutes: ptrTo(30)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-04-01", sharedtypes.HabitCompletionStatusCompleted, ptrTo(30), nil); err != nil {
		t.Fatalf("complete habit day 1: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-04-02", sharedtypes.HabitCompletionStatusCompleted, ptrTo(25), nil); err != nil {
		t.Fatalf("complete habit day 2: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-04-03", sharedtypes.HabitCompletionStatusFailed, nil, nil); err != nil {
		t.Fatalf("fail habit day 3: %v", err)
	}

	days, err := corecommands.ComputeMetricsRange(ctx, coreCtx, "2026-04-01", "2026-04-03")
	if err != nil {
		t.Fatalf("compute metrics range: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("expected 3 metric days, got %d", len(days))
	}
	for _, day := range days {
		if day.HabitDueCount != 1 {
			t.Fatalf("expected one due habit per day, got %+v", day)
		}
	}
	if days[0].HabitCompletedCount != 1 || days[1].HabitCompletedCount != 1 || days[2].HabitFailedCount != 1 {
		t.Fatalf("unexpected habit counts: %+v %+v %+v", days[0], days[1], days[2])
	}

	rollup := corecommands.ComputeMetricsRollupFromDays("2026-04-01", "2026-04-03", days)
	if rollup.HabitDueCount != 3 || rollup.HabitCompletedCount != 2 || rollup.HabitFailedCount != 1 {
		t.Fatalf("unexpected habit rollup counts: %+v", rollup)
	}

	streaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-04-01", "2026-04-03")
	if err != nil {
		t.Fatalf("compute streaks: %v", err)
	}
	if streaks.CurrentHabitDays != 0 {
		t.Fatalf("expected current habit streak to break on failed day, got %+v", streaks)
	}
	if streaks.LongestHabitDays != 2 {
		t.Fatalf("expected longest habit streak of 2, got %+v", streaks)
	}
}

func TestCustomHabitStreaksUseCalendarBucketsAndThresholds(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-15T09:00:00Z"
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
	walk, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Walk", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create walk: %v", err)
	}
	journal, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Journal", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create journal: %v", err)
	}

	for _, date := range []string{"2026-04-01", "2026-04-03", "2026-04-08", "2026-04-10"} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, walk.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete walk %s: %v", date, err)
		}
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, journal.ID, "2026-04-09", sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
		t.Fatalf("complete journal: %v", err)
	}

	defs := []sharedtypes.HabitStreakDefinition{{
		ID:            "health",
		Name:          "Health streak",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodWeek,
		RequiredCount: 2,
		HabitIDs:      []int64{walk.ID, journal.ID},
	}}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set custom streak defs: %v", err)
	}

	streaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-04-01", "2026-04-12")
	if err != nil {
		t.Fatalf("compute streaks: %v", err)
	}
	if len(streaks.CustomHabitStreaks) != 1 {
		t.Fatalf("expected one custom streak, got %+v", streaks.CustomHabitStreaks)
	}
	if streaks.CustomHabitStreaks[0].Current != 2 || streaks.CustomHabitStreaks[0].Longest != 2 {
		t.Fatalf("unexpected custom streak summary: %+v", streaks.CustomHabitStreaks[0])
	}
}
