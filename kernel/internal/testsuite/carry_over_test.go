package testsuite

import (
	"context"
	"testing"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func newCarryFixture(t *testing.T) (context.Context, *core.Context, *string, sharedtypes.Issue) {
	t.Helper()
	ctx := context.Background()
	currentNow := "2026-03-24T12:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return currentNow })
	repo, _ := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	stream, _ := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
	today := "2026-03-24"
	issue := mustCreateCommandIssue(t, ctx, coreCtx, stream.ID, "Long task", &today)
	return ctx, coreCtx, &currentNow, issue
}

func logWork(t *testing.T, ctx context.Context, c *core.Context, issueID int64, date string, minutes int) {
	t.Helper()
	if _, err := corecommands.LogManualSession(ctx, c, corecommands.ManualSessionInput{
		IssueID: issueID, Date: date, WorkDurationSeconds: minutes * 60,
	}); err != nil {
		t.Fatalf("log work: %v", err)
	}
}

func mustPlanEntry(t *testing.T, ctx context.Context, c *core.Context, date string, issueID int64) sharedtypes.DailyPlanEntry {
	t.Helper()
	plan, err := corecommands.GetDailyPlan(ctx, c, date)
	if err != nil {
		t.Fatalf("get plan %s: %v", date, err)
	}
	for _, entry := range plan.Entries {
		if entry.IssueID == issueID {
			return entry
		}
	}
	t.Fatalf("no entry for issue %d on %s: %+v", issueID, date, plan.Entries)
	return sharedtypes.DailyPlanEntry{}
}

func TestCarryOverMovesWorkedIssueToNextDay(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	logWork(t, ctx, coreCtx, issue.ID, "2026-03-24", 20)

	*now = "2026-03-25T12:00:00Z"
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("carry over: %v", err)
	}
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("second carry over run: %v", err)
	}

	carried := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID)
	if carried.Status != sharedtypes.DailyPlanEntryStatusCarried || carried.WorkedSeconds != 1200 ||
		carried.CarriedToDate == nil || *carried.CarriedToDate != "2026-03-25" {
		t.Fatalf("expected carried entry with worked time and target, got %+v", carried)
	}
	next := mustPlanEntry(t, ctx, coreCtx, "2026-03-25", issue.ID)
	if next.Status != sharedtypes.DailyPlanEntryStatusPlanned || next.CarryCount != 1 ||
		next.CarriedFromDate == nil || *next.CarriedFromDate != "2026-03-24" ||
		next.PostponeCount != 0 || next.CurrentDelayedDays != 0 || next.FailScore != 0 {
		t.Fatalf("expected a fresh planned entry carried from yesterday, got %+v", next)
	}
	plan, _ := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-24")
	if plan.Summary.CarriedCount != 1 || plan.Summary.FailedCount != 0 {
		t.Fatalf("expected one carried and no failures, got %+v", plan.Summary)
	}
	updated, _ := coreCtx.Issues.GetByID(ctx, issue.ID, coreCtx.UserID)
	if updated.TodoForDate == nil || *updated.TodoForDate != "2026-03-25" {
		t.Fatalf("expected issue planned for the next day, got %+v", updated.TodoForDate)
	}
}

func TestCarryOverRespectsThreshold(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	logWork(t, ctx, coreCtx, issue.ID, "2026-03-24", 5)

	*now = "2026-03-25T12:00:00Z"
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("carry over: %v", err)
	}
	if entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID); entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		t.Fatalf("expected short work to stay planned (overdue), got %+v", entry)
	}

	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyCarryOverMinMinutes, 0); err != nil {
		t.Fatalf("set threshold: %v", err)
	}
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("carry over: %v", err)
	}
	if entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID); entry.Status != sharedtypes.DailyPlanEntryStatusCarried {
		t.Fatalf("expected any session to count with a zero threshold, got %+v", entry)
	}
}

func TestCarryOverSkipsUntouchedAndDisabled(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyCarryOverEnabled, false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	logWork(t, ctx, coreCtx, issue.ID, "2026-03-24", 30)
	*now = "2026-03-25T12:00:00Z"
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("carry over: %v", err)
	}
	if entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID); entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		t.Fatalf("expected disabled carry-over to leave the entry, got %+v", entry)
	}
}

func TestCarryOverCatchUpLandsOnToday(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	logWork(t, ctx, coreCtx, issue.ID, "2026-03-24", 20)
	*now = "2026-03-27T12:00:00Z"
	if err := corecommands.CarryOverPastDays(ctx, coreCtx); err != nil {
		t.Fatalf("carry over: %v", err)
	}
	if entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-27", issue.ID); entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		t.Fatalf("expected catch-up to land on today, got %+v", entry)
	}
}

func TestManualCarryAndUndo(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	if _, err := corecommands.CarryIssue(ctx, coreCtx, issue.ID, ""); err != nil {
		t.Fatalf("carry: %v", err)
	}
	if entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID); entry.Status != sharedtypes.DailyPlanEntryStatusCarried {
		t.Fatalf("expected manual carry without work, got %+v", entry)
	}

	*now = "2026-03-24T12:02:00Z"
	updated, err := corecommands.UndoCarry(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if updated.TodoForDate == nil || *updated.TodoForDate != "2026-03-24" {
		t.Fatalf("expected issue back on its day, got %+v", updated.TodoForDate)
	}
	entry := mustPlanEntry(t, ctx, coreCtx, "2026-03-24", issue.ID)
	if entry.Status != sharedtypes.DailyPlanEntryStatusPlanned || entry.CarriedToEntryID != nil {
		t.Fatalf("expected entry reopened without a carry link, got %+v", entry)
	}
	plan, _ := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-25")
	if len(plan.Entries) != 0 {
		t.Fatalf("expected the carried-to entry removed, got %+v", plan.Entries)
	}

	if _, err := corecommands.CarryIssue(ctx, coreCtx, issue.ID, ""); err != nil {
		t.Fatalf("carry again: %v", err)
	}
	*now = "2026-03-24T13:00:00Z"
	if _, err := corecommands.UndoCarry(ctx, coreCtx, issue.ID); err == nil {
		t.Fatal("expected undo to fail after the rollback window")
	}
}

func TestCarryOverStallFeedsAccountability(t *testing.T) {
	ctx, coreCtx, now, issue := newCarryFixture(t)
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyCarryOverStallAfter, 1); err != nil {
		t.Fatalf("set stall limit: %v", err)
	}
	if _, err := corecommands.CarryIssue(ctx, coreCtx, issue.ID, ""); err != nil {
		t.Fatalf("carry 1: %v", err)
	}
	first := mustPlanEntry(t, ctx, coreCtx, "2026-03-25", issue.ID)
	if !first.Stalled || first.FailScore != 0 {
		t.Fatalf("expected stalled at the limit with no stall score yet, got %+v", first)
	}
	*now = "2026-03-25T12:00:00Z"
	if _, err := corecommands.CarryIssue(ctx, coreCtx, issue.ID, ""); err != nil {
		t.Fatalf("carry 2: %v", err)
	}
	second := mustPlanEntry(t, ctx, coreCtx, "2026-03-26", issue.ID)
	if second.CarryCount != 2 || second.Penalty == nil || second.Penalty.StallComponent != 0.5 ||
		second.FailScore != 0.5 {
		t.Fatalf("expected a stall component past the limit, got %+v / %+v", second, second.Penalty)
	}
	plan, _ := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-26")
	if plan.Summary.AccountabilityScore != 0.5 || plan.Summary.BacklogPressure != 0 {
		t.Fatalf("expected stall to feed accountability only, got %+v", plan.Summary)
	}
}
