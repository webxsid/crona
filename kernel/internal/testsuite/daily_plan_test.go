package testsuite

import (
	"context"
	"testing"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/export"
	sharedtypes "crona/shared/types"
)

func TestDailyPlanMarksMovedTodayIssueFailedAfterRollbackWindow(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-25T09:00:00Z"
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
	today := "2026-03-25"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Ship accountability", TodoForDate: &today})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	currentNow = "2026-03-25T09:10:00Z"
	tomorrow := "2026-03-26"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, tomorrow); err != nil {
		t.Fatalf("move issue: %v", err)
	}

	plan, err := corecommands.GetDailyPlan(ctx, coreCtx, today)
	if err != nil {
		t.Fatalf("get plan before finalization: %v", err)
	}
	if plan == nil || len(plan.Entries) != 1 || plan.Entries[0].PendingFailureAt == nil {
		t.Fatalf("expected pending failure after move, got %+v", plan)
	}
	if plan.Entries[0].Status != sharedtypes.DailyPlanEntryStatusPlanned {
		t.Fatalf("expected planned status before grace expiry, got %s", plan.Entries[0].Status)
	}

	currentNow = "2026-03-25T09:16:00Z"
	plan, err = corecommands.GetDailyPlan(ctx, coreCtx, today)
	if err != nil {
		t.Fatalf("get plan after finalization: %v", err)
	}
	if plan == nil || len(plan.Entries) != 1 {
		t.Fatalf("expected one plan entry, got %+v", plan)
	}
	entry := plan.Entries[0]
	if entry.Status != sharedtypes.DailyPlanEntryStatusFailed {
		t.Fatalf("expected failed entry after rollback window, got %s", entry.Status)
	}
	if entry.FailureReason == nil || *entry.FailureReason != sharedtypes.DailyPlanFailureReasonMoved {
		t.Fatalf("expected moved failure reason, got %+v", entry.FailureReason)
	}
}

func TestDailyPlanRollbackRestoresEntryWithoutFailure(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-25T09:00:00Z"
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
	today := "2026-03-25"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Keep commitment", TodoForDate: &today})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	currentNow = "2026-03-25T09:10:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-26"); err != nil {
		t.Fatalf("move issue away: %v", err)
	}
	currentNow = "2026-03-25T09:13:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, today); err != nil {
		t.Fatalf("move issue back: %v", err)
	}
	currentNow = "2026-03-25T09:20:00Z"

	plan, err := corecommands.GetDailyPlan(ctx, coreCtx, today)
	if err != nil {
		t.Fatalf("get restored plan: %v", err)
	}
	if plan == nil || len(plan.Entries) != 1 {
		t.Fatalf("expected restored entry, got %+v", plan)
	}
	entry := plan.Entries[0]
	if entry.Status != sharedtypes.DailyPlanEntryStatusPlanned {
		t.Fatalf("expected restored entry to remain planned, got %s", entry.Status)
	}
	if entry.PendingFailureAt != nil || entry.FailureReason != nil {
		t.Fatalf("expected no failure after rollback, got %+v", entry)
	}
}

func TestDailyReportKeepsFailedPlannedIssueAfterReschedule(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-25T09:00:00Z"
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
	today := "2026-03-25"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Export should retain this", TodoForDate: &today})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	currentNow = "2026-03-25T09:10:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-26"); err != nil {
		t.Fatalf("reschedule issue: %v", err)
	}
	currentNow = "2026-03-25T09:16:00Z"

	data, err := export.BuildDailyReportData(ctx, coreCtx, today)
	if err != nil {
		t.Fatalf("build daily report data: %v", err)
	}
	if data.Plan == nil || len(data.Plan.Entries) != 1 {
		t.Fatalf("expected daily plan in report data, got %+v", data.Plan)
	}
	if len(data.Issues) != 1 {
		t.Fatalf("expected failed planned issue in report issues, got %d", len(data.Issues))
	}
	if data.Issues[0].PlanStatus != sharedtypes.DailyPlanEntryStatusFailed {
		t.Fatalf("expected report issue plan status failed, got %s", data.Issues[0].PlanStatus)
	}
}

func TestDailyReportIncludesWorkedPinnedFutureIssue(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-05-06T09:00:00Z"
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
	reportDate := "2026-05-06"
	futureDate := "2026-05-10"

	planned, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Today planned", TodoForDate: &reportDate})
	if err != nil {
		t.Fatalf("create planned issue: %v", err)
	}
	workedPinned, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Worked pinned future", TodoForDate: &futureDate})
	if err != nil {
		t.Fatalf("create worked pinned issue: %v", err)
	}
	quietPinned, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Quiet pinned future", TodoForDate: &futureDate})
	if err != nil {
		t.Fatalf("create quiet pinned issue: %v", err)
	}
	pinned := true
	for _, issueID := range []int64{workedPinned.ID, quietPinned.ID} {
		if _, err := corecommands.UpdateIssue(ctx, coreCtx, issueID, struct {
			Title           sharedtypes.Patch[string]
			Description     sharedtypes.Patch[string]
			EstimateMinutes sharedtypes.Patch[int]
			Notes           sharedtypes.Patch[string]
			PinnedDaily     sharedtypes.Patch[bool]
		}{PinnedDaily: sharedtypes.Patch[bool]{Set: true, Value: &pinned}}); err != nil {
			t.Fatalf("pin issue %d: %v", issueID, err)
		}
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             workedPinned.ID,
		Date:                reportDate,
		WorkDurationSeconds: 1800,
		StartTime:           strPtr("10:00"),
	}); err != nil {
		t.Fatalf("log manual session: %v", err)
	}

	data, err := export.BuildDailyReportData(ctx, coreCtx, reportDate)
	if err != nil {
		t.Fatalf("build daily report data: %v", err)
	}
	titles := map[string]sharedtypes.DailyReportIssue{}
	for _, issue := range data.Issues {
		titles[issue.Title] = issue
	}
	if _, ok := titles[planned.Title]; !ok {
		t.Fatalf("expected planned issue in report, got %+v", data.Issues)
	}
	if got, ok := titles[workedPinned.Title]; !ok || got.WorkedSeconds != 1800 {
		t.Fatalf("expected worked pinned future issue with 1800s, got %+v", got)
	}
	if _, ok := titles[quietPinned.Title]; ok {
		t.Fatalf("expected quiet pinned future issue to stay out of report, got %+v", data.Issues)
	}
}

func TestDailyReportBurnoutUsesSevenDayWindow(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-05-07T09:00:00Z"
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
	reportDate := "2026-05-07"
	startDate := "2026-05-01"
	for day := startDate; day < reportDate; day = mustShiftISODate(t, day, 1) {
		issue := mustCreateCommandIssue(t, ctx, coreCtx, stream.ID, "Burnout window work "+day, &day)
		if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
			IssueID:             issue.ID,
			Date:                day,
			WorkDurationSeconds: 8 * 60 * 60,
			StartTime:           strPtr("09:00"),
		}); err != nil {
			t.Fatalf("log heavy session for %s: %v", day, err)
		}
	}
	issue := mustCreateCommandIssue(t, ctx, coreCtx, stream.ID, "Burnout window work "+reportDate, &reportDate)
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             issue.ID,
		Date:                reportDate,
		WorkDurationSeconds: 30 * 60,
		StartTime:           strPtr("09:00"),
	}); err != nil {
		t.Fatalf("log report date session: %v", err)
	}

	data, err := export.BuildDailyReportData(ctx, coreCtx, reportDate)
	if err != nil {
		t.Fatalf("build daily report data: %v", err)
	}
	if data.Metrics == nil || data.Metrics.Burnout == nil {
		t.Fatalf("expected report metrics burnout, got %+v", data.Metrics)
	}
	days, err := corecommands.ComputeMetricsRange(ctx, coreCtx, startDate, reportDate)
	if err != nil {
		t.Fatalf("compute metrics range: %v", err)
	}
	if len(days) != 7 || days[len(days)-1].Burnout == nil {
		t.Fatalf("expected seven-day burnout range, got %+v", days)
	}
	if data.Metrics.Burnout.Score != days[len(days)-1].Burnout.Score {
		t.Fatalf("expected report burnout score %d from seven-day window, got %d", days[len(days)-1].Burnout.Score, data.Metrics.Burnout.Score)
	}
	oneDay, err := corecommands.ComputeMetricsRange(ctx, coreCtx, reportDate, reportDate)
	if err != nil {
		t.Fatalf("compute one-day metrics range: %v", err)
	}
	if len(oneDay) != 1 || oneDay[0].Burnout == nil {
		t.Fatalf("expected one-day burnout, got %+v", oneDay)
	}
	if oneDay[0].Burnout.Score == data.Metrics.Burnout.Score {
		t.Fatalf("test setup should distinguish one-day and seven-day burnout scores; both were %d", data.Metrics.Burnout.Score)
	}
}

func TestDailyPlanScoreTracksPostponeAndRecovery(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-25T09:00:00Z"
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
	today := "2026-03-25"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Score this", TodoForDate: &today})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	currentNow = "2026-03-25T09:02:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-30"); err != nil {
		t.Fatalf("delay issue: %v", err)
	}

	plan, err := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-30")
	if err != nil {
		t.Fatalf("get delayed plan: %v", err)
	}
	if plan == nil || len(plan.Entries) != 1 {
		t.Fatalf("expected active delayed entry, got %+v", plan)
	}
	entry := plan.Entries[0]
	if entry.PostponeCount != 1 {
		t.Fatalf("expected postpone count 1, got %d", entry.PostponeCount)
	}
	if entry.CurrentDelayedDays != 5 {
		t.Fatalf("expected delayed days 5, got %d", entry.CurrentDelayedDays)
	}
	if entry.FailScore <= 0 {
		t.Fatalf("expected positive fail score, got %.2f", entry.FailScore)
	}
	if plan.Summary.DelayedIssueCount != 1 {
		t.Fatalf("expected delayed issue count 1, got %+v", plan.Summary)
	}

	currentNow = "2026-03-25T09:04:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-28"); err != nil {
		t.Fatalf("pull issue back: %v", err)
	}
	plan, err = corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-28")
	if err != nil {
		t.Fatalf("get recovered plan: %v", err)
	}
	entry = plan.Entries[0]
	if entry.PostponeCount != 1 {
		t.Fatalf("expected postpone count to remain 1, got %d", entry.PostponeCount)
	}
	if entry.CurrentDelayedDays != 3 {
		t.Fatalf("expected delayed days to reduce to 3, got %d", entry.CurrentDelayedDays)
	}
	if entry.FailScore >= 2.6 {
		t.Fatalf("expected reduced fail score after recovery, got %.2f", entry.FailScore)
	}
}

func mustShiftISODate(t *testing.T, date string, days int) string {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatalf("parse date %q: %v", date, err)
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func mustCreateCommandIssue(t *testing.T, ctx context.Context, coreCtx *core.Context, streamID int64, title string, todoForDate *string) sharedtypes.Issue {
	t.Helper()
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: streamID, Title: title, TodoForDate: todoForDate})
	if err != nil {
		t.Fatalf("create issue %q: %v", title, err)
	}
	return issue
}

func TestDailyPlanScoreExcludesRestDaysAndAwayMode(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-27T09:00:00Z"
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
	friday := "2026-03-27"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Respect rest days", TodoForDate: &friday})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyRestWeekdays, []int{0, 6}); err != nil {
		t.Fatalf("set rest weekdays: %v", err)
	}

	currentNow = "2026-03-27T09:05:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-30"); err != nil {
		t.Fatalf("delay issue across weekend: %v", err)
	}

	plan, err := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-30")
	if err != nil {
		t.Fatalf("get plan with weekend rest: %v", err)
	}
	entry := plan.Entries[0]
	if entry.CurrentDelayedDays != 1 {
		t.Fatalf("expected only Monday to count after excluding weekend rest days, got %d", entry.CurrentDelayedDays)
	}
	if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, sharedtypes.CoreSettingsKeyAwayModeEnabled, true); err != nil {
		t.Fatalf("enable away mode: %v", err)
	}
	plan, err = corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-30")
	if err != nil {
		t.Fatalf("get plan with away mode: %v", err)
	}
	entry = plan.Entries[0]
	if entry.CurrentDelayedDays != 0 || entry.FailScore != 0 {
		t.Fatalf("expected away mode to zero out delay burden, got delayed=%d score=%.2f", entry.CurrentDelayedDays, entry.FailScore)
	}
	if plan.Summary.DelayedIssueCount != 0 || plan.Summary.AccountabilityScore != 0 {
		t.Fatalf("expected away mode to clear aggregate burden, got %+v", plan.Summary)
	}
}

func TestDailyPlanActiveSummaryExcludesResolvedDelayedIssues(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-03-25T09:00:00Z"
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
	today := "2026-03-25"
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Finish delayed item", TodoForDate: &today})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	currentNow = "2026-03-25T09:02:00Z"
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issue.ID, "2026-03-30"); err != nil {
		t.Fatalf("delay issue: %v", err)
	}

	plan, err := corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-30")
	if err != nil {
		t.Fatalf("get delayed plan: %v", err)
	}
	if plan.Summary.DelayedIssueCount != 1 || plan.Summary.HighRiskIssueCount != 1 {
		t.Fatalf("expected delayed issue to count before resolution, got %+v", plan.Summary)
	}

	currentNow = "2026-03-25T09:05:00Z"
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, issue.ID, sharedtypes.IssueStatusInProgress, nil); err != nil {
		t.Fatalf("start issue: %v", err)
	}
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, issue.ID, sharedtypes.IssueStatusDone, nil); err != nil {
		t.Fatalf("complete issue: %v", err)
	}

	active, err := coreCtx.DailyPlans.ListActiveEntries(ctx, coreCtx.UserID)
	if err != nil {
		t.Fatalf("list active entries: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("expected resolved issue to drop out of active entries, got %+v", active)
	}

	plan, err = corecommands.GetDailyPlan(ctx, coreCtx, "2026-03-30")
	if err != nil {
		t.Fatalf("get resolved plan: %v", err)
	}
	if plan.Summary.DelayedIssueCount != 0 || plan.Summary.HighRiskIssueCount != 0 || plan.Summary.AccountabilityScore != 0 {
		t.Fatalf("expected resolved issue to stop contributing to active accountability, got %+v", plan.Summary)
	}
}
