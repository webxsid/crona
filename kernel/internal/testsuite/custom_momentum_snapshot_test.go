package testsuite

import (
	"context"
	"slices"
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
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)
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
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)
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
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)
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

func TestMomentumAllModeUsesFinalBucketCountsForHabitsAndContexts(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-03T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

	habitRepo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Habits"})
	if err != nil {
		t.Fatalf("create habit repo: %v", err)
	}
	habitStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: habitRepo.ID, Name: "weekly"})
	if err != nil {
		t.Fatalf("create habit stream: %v", err)
	}
	habitA, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: habitStream.ID, Name: "One", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit one: %v", err)
	}
	habitB, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: habitStream.ID, Name: "Two", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit two: %v", err)
	}
	for _, date := range []string{"2026-04-01", "2026-04-02"} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, habitA.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit one %s: %v", date, err)
		}
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, habitB.ID, date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit two %s: %v", date, err)
		}
	}

	contextRepoA, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create context repo a: %v", err)
	}
	contextRepoB, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Personal"})
	if err != nil {
		t.Fatalf("create context repo b: %v", err)
	}
	contextStreamA, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: contextRepoA.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create context stream a: %v", err)
	}
	contextStreamB, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: contextRepoB.ID, Name: "ops"})
	if err != nil {
		t.Fatalf("create context stream b: %v", err)
	}
	contextIssueA, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: contextStreamA.ID, Title: "Delivery"})
	if err != nil {
		t.Fatalf("create context issue a: %v", err)
	}
	contextIssueB, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: contextStreamB.ID, Title: "Support"})
	if err != nil {
		t.Fatalf("create context issue b: %v", err)
	}
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, contextIssueA.ID, sharedtypes.IssueStatusPlanned, nil); err != nil {
		t.Fatalf("plan context issue a: %v", err)
	}
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, contextIssueB.ID, sharedtypes.IssueStatusPlanned, nil); err != nil {
		t.Fatalf("plan context issue b: %v", err)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             contextIssueA.ID,
		Date:                "2026-04-01",
		WorkDurationSeconds: 3600,
	}); err != nil {
		t.Fatalf("log context session a1: %v", err)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             contextIssueA.ID,
		Date:                "2026-04-02",
		WorkDurationSeconds: 3600,
	}); err != nil {
		t.Fatalf("log context session a2: %v", err)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             contextIssueB.ID,
		Date:                "2026-04-02",
		WorkDurationSeconds: 3600,
	}); err != nil {
		t.Fatalf("log context session b1: %v", err)
	}

	defs := []sharedtypes.HabitStreakDefinition{
		{
			ID:            "habit-all",
			Name:          "Habit all",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 2,
			HabitIDs:      []int64{habitA.ID, habitB.ID},
		},
		{
			ID:            "context-all",
			Name:          "Context all",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 7200,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: contextRepoA.ID, StreamID: int64Ptr(contextStreamA.ID)},
				{RepoID: contextRepoB.ID, StreamID: int64Ptr(contextStreamB.ID)},
			},
		},
	}
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)

	cards, err := corecommands.ListMomentumCards(ctx, coreCtx, "2026-04-03", 7)
	if err != nil {
		t.Fatalf("list momentum cards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected two momentum cards, got %+v", cards)
	}

	habitCard := momentumCardByID(cards, "habit-all")
	if habitCard == nil {
		t.Fatalf("expected habit-all card in results, got %+v", cards)
	}
	if len(habitCard.Series) == 0 {
		t.Fatalf("expected habit card series, got %+v", habitCard)
	}
	habitLatest := habitCard.Series[len(habitCard.Series)-1]
	if habitLatest.Target != 2 {
		t.Fatalf("expected grouped habit target to stay at 2, got %+v", habitLatest)
	}
	if habitLatest.Count != 2 {
		t.Fatalf("expected habit bucket to use the final grouped count, got %+v", habitLatest)
	}

	contextCard := momentumCardByID(cards, "context-all")
	if contextCard == nil {
		t.Fatalf("expected context-all card in results, got %+v", cards)
	}
	if len(contextCard.Series) == 0 {
		t.Fatalf("expected context card series, got %+v", contextCard)
	}
	contextLatest := contextCard.Series[len(contextCard.Series)-1]
	if contextLatest.Target != 7200 {
		t.Fatalf("expected grouped context target to stay at 2h, got %+v", contextLatest)
	}
	if contextLatest.Count != 3600 {
		t.Fatalf("expected context bucket to use the final shared time, got %+v", contextLatest)
	}
}

func TestMomentumContextTargetNamesIncludeStreamsWithoutHabits(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-06-19T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	appStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create app stream: %v", err)
	}
	infraStream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "infra"})
	if err != nil {
		t.Fatalf("create infra stream: %v", err)
	}
	habit, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: appStream.ID, Name: "Tracked", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	if _, err := corecommands.CompleteHabit(ctx, coreCtx, habit.ID, "2026-06-19", sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
		t.Fatalf("complete habit: %v", err)
	}

	defs := []sharedtypes.HabitStreakDefinition{
		{
			ID:            "context-all",
			Name:          "Context all",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 3600,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: repo.ID, StreamID: int64Ptr(appStream.ID)},
				{RepoID: repo.ID, StreamID: int64Ptr(infraStream.ID)},
			},
		},
	}
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)

	cards, err := corecommands.ListMomentumCards(ctx, coreCtx, "2026-06-19", 7)
	if err != nil {
		t.Fatalf("list momentum cards: %v", err)
	}
	card := momentumCardByID(cards, "context-all")
	if card == nil {
		t.Fatalf("expected context-all card, got %+v", cards)
	}
	if got, want := card.TargetNames, []string{"Work/app", "Work/infra"}; !slices.Equal(got, want) {
		t.Fatalf("expected full context names including streams, got %+v want %+v", got, want)
	}
}

func TestMomentumAllModeCountsOnlyGroupedDays(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-04-03T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Grouped"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "daily"})
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
	}{StreamID: stream.ID, Name: "A", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
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
	}{StreamID: stream.ID, Name: "B", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit b: %v", err)
	}
	habitC, err := corecommands.CreateHabit(ctx, coreCtx, struct {
		StreamID      int64
		Name          string
		Description   *string
		ScheduleType  string
		Weekdays      []int
		TargetMinutes *int
	}{StreamID: stream.ID, Name: "C", ScheduleType: string(sharedtypes.HabitScheduleDaily)})
	if err != nil {
		t.Fatalf("create habit c: %v", err)
	}
	for _, step := range []struct {
		date string
		hid  int64
	}{
		{date: "2026-04-01", hid: habitA.ID},
		{date: "2026-04-01", hid: habitB.ID},
		{date: "2026-04-02", hid: habitA.ID},
		{date: "2026-04-02", hid: habitC.ID},
		{date: "2026-04-03", hid: habitB.ID},
		{date: "2026-04-03", hid: habitC.ID},
	} {
		if _, err := corecommands.CompleteHabit(ctx, coreCtx, step.hid, step.date, sharedtypes.HabitCompletionStatusCompleted, nil, nil); err != nil {
			t.Fatalf("complete habit %d on %s: %v", step.hid, step.date, err)
		}
	}

	defs := []sharedtypes.HabitStreakDefinition{
		{
			ID:            "grouped-all",
			Name:          "Grouped all",
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodMonth,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			RequiredCount: 1,
			HabitIDs:      []int64{habitA.ID, habitB.ID, habitC.ID},
		},
	}
	mustReplaceHabitStreakDefinitions(t, ctx, coreCtx, defs)

	cards, err := corecommands.ListMomentumCards(ctx, coreCtx, "2026-04-03", 7)
	if err != nil {
		t.Fatalf("list momentum cards: %v", err)
	}
	card := momentumCardByID(cards, "grouped-all")
	if card == nil {
		t.Fatalf("expected grouped-all card in results, got %+v", cards)
	}
	if len(card.Series) == 0 {
		t.Fatalf("expected grouped-all card series, got %+v", card)
	}
	last := card.Series[len(card.Series)-1]
	if last.Count != 0 {
		t.Fatalf("expected no grouped days to count, got %+v", last)
	}
}

func momentumCardByID(cards []sharedtypes.MomentumCard, id string) *sharedtypes.MomentumCard {
	for i := range cards {
		if cards[i].Definition.ID == id {
			return &cards[i]
		}
	}
	return nil
}

func int64Ptr(v int64) *int64 {
	return &v
}
