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

func TestLifetimeStreaksUseHistoryBeforeWellbeingWindow(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-10T12:00:00Z"
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
	}{StreamID: stream.ID, Name: "Daily review", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	for _, date := range []string{"2026-04-01", "2026-04-02", "2026-04-03", "2026-04-04", "2026-04-05", "2026-04-06", "2026-04-07", "2026-04-08", "2026-04-09", "2026-04-10"} {
		seedFocusDay(t, ctx, coreCtx, stream.ID, 5000+int64(len(date))+int64(date[len(date)-1]), date)
		seedCheckInDay(t, ctx, coreCtx, date)
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit %s: %v", date, err)
		}
	}

	windowStreaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-04-04", "2026-04-10")
	if err != nil {
		t.Fatalf("compute window streaks: %v", err)
	}
	if windowStreaks.CurrentFocusDays != 7 || windowStreaks.CurrentCheckInDays != 7 || windowStreaks.CurrentHabitDays != 7 {
		t.Fatalf("expected 7-day window streaks, got %+v", windowStreaks)
	}

	lifetimeStreaks, err := corecommands.ComputeMetricsLifetimeStreaks(ctx, coreCtx, "2026-04-10")
	if err != nil {
		t.Fatalf("compute lifetime streaks: %v", err)
	}
	if lifetimeStreaks.CurrentFocusDays != 10 || lifetimeStreaks.CurrentCheckInDays != 10 || lifetimeStreaks.CurrentHabitDays != 10 {
		t.Fatalf("expected lifetime streaks to include all history, got %+v", lifetimeStreaks)
	}
}

func TestLifetimeCustomHabitStreaksUseHistoryBeforeWellbeingWindow(t *testing.T) {
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
		{ID: "weekly", Name: "Weekly training", Enabled: true, Period: sharedtypes.HabitStreakPeriodWeek, RequiredCount: 2, HabitIDs: []int64{habit.ID}},
		{ID: "monthly", Name: "Monthly training", Enabled: true, Period: sharedtypes.HabitStreakPeriodMonth, RequiredCount: 3, HabitIDs: []int64{habit.ID}},
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set custom streak defs: %v", err)
	}

	windowStreaks, err := corecommands.ComputeMetricsStreaks(ctx, coreCtx, "2026-04-04", "2026-04-10")
	if err != nil {
		t.Fatalf("compute window streaks: %v", err)
	}
	lifetimeStreaks, err := corecommands.ComputeMetricsLifetimeStreaks(ctx, coreCtx, "2026-04-10")
	if err != nil {
		t.Fatalf("compute lifetime streaks: %v", err)
	}
	if len(windowStreaks.CustomHabitStreaks) != 2 || len(lifetimeStreaks.CustomHabitStreaks) != 2 {
		t.Fatalf("expected two custom streaks, got window=%+v lifetime=%+v", windowStreaks.CustomHabitStreaks, lifetimeStreaks.CustomHabitStreaks)
	}
	if windowStreaks.CustomHabitStreaks[0].Current != 1 || windowStreaks.CustomHabitStreaks[1].Current != 0 {
		t.Fatalf("expected 7-day custom streaks to miss older buckets, got %+v", windowStreaks.CustomHabitStreaks)
	}
	if lifetimeStreaks.CustomHabitStreaks[0].Current != 2 || lifetimeStreaks.CustomHabitStreaks[0].Longest != 2 {
		t.Fatalf("unexpected lifetime weekly custom streak: %+v", lifetimeStreaks.CustomHabitStreaks[0])
	}
	if lifetimeStreaks.CustomHabitStreaks[1].Current != 3 || lifetimeStreaks.CustomHabitStreaks[1].Longest != 3 {
		t.Fatalf("unexpected lifetime monthly custom streak: %+v", lifetimeStreaks.CustomHabitStreaks[1])
	}
}

func TestCustomWeeklyAndMonthlyStreaksSurviveIncompleteOpenBucket(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-10T12:00:00Z"
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
	weeklyHabit, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Practice", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create weekly habit: %v", err)
	}
	monthlyHabit, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "Review", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create monthly habit: %v", err)
	}
	for _, date := range []string{
		"2026-02-24", "2026-02-26", "2026-03-03", "2026-03-05",
		"2026-03-10",
	} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, weeklyHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete weekly habit %s: %v", date, err)
		}
	}
	for _, date := range []string{
		"2026-01-07", "2026-01-14", "2026-01-21",
		"2026-02-04", "2026-02-11", "2026-02-18",
		"2026-03-10",
	} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, monthlyHabit.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete monthly habit %s: %v", date, err)
		}
	}
	defs := []sharedtypes.HabitStreakDefinition{
		{ID: "weekly", Name: "Weekly practice", Enabled: true, Period: sharedtypes.HabitStreakPeriodWeek, RequiredCount: 2, HabitIDs: []int64{weeklyHabit.ID}},
		{ID: "monthly", Name: "Monthly practice", Enabled: true, Period: sharedtypes.HabitStreakPeriodMonth, RequiredCount: 3, HabitIDs: []int64{monthlyHabit.ID}},
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyHabitStreakDefs, defs); err != nil {
		t.Fatalf("set custom streak defs: %v", err)
	}

	streaks, err := corecommands.ComputeMetricsLifetimeStreaks(ctx, coreCtx, "2026-03-10")
	if err != nil {
		t.Fatalf("compute lifetime streaks: %v", err)
	}
	if len(streaks.CustomHabitStreaks) != 2 {
		t.Fatalf("expected two custom streaks, got %+v", streaks.CustomHabitStreaks)
	}
	if streaks.CustomHabitStreaks[0].Current != 2 || streaks.CustomHabitStreaks[0].Longest != 2 {
		t.Fatalf("expected incomplete open week not to break weekly streak, got %+v", streaks.CustomHabitStreaks[0])
	}
	if streaks.CustomHabitStreaks[1].Current != 2 || streaks.CustomHabitStreaks[1].Longest != 2 {
		t.Fatalf("expected incomplete open month not to break monthly streak, got %+v", streaks.CustomHabitStreaks[1])
	}
}
