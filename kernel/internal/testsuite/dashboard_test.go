package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestDashboardSummariesExposeExecutionFocusDistributionAndProgress(t *testing.T) {
	ctx := context.Background()
	coreCtx, _ := newTestCoreContext(t, func() string { return "2026-04-01T12:00:00Z" })

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
	estimate60 := 60
	estimate30 := 30
	issueDone, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Ship dashboard cards", EstimateMinutes: &estimate60, TodoForDate: stringPtr("2026-04-01")})
	if err != nil {
		t.Fatalf("create issueDone: %v", err)
	}
	issueCarry, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Refine streak strip", EstimateMinutes: &estimate30, TodoForDate: stringPtr("2026-03-31")})
	if err != nil {
		t.Fatalf("create issueCarry: %v", err)
	}
	if _, err := corecommands.MarkIssueTodoForDate(ctx, coreCtx, issueCarry.ID, "2026-04-01"); err != nil {
		t.Fatalf("move issueCarry: %v", err)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             issueDone.ID,
		Date:                "2026-04-01",
		WorkDurationSeconds: 2700,
	}); err != nil {
		t.Fatalf("log manual session: %v", err)
	}
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, issueDone.ID, sharedtypes.IssueStatusDone, nil); err != nil {
		t.Fatalf("complete issueDone: %v", err)
	}

	window, err := corecommands.ComputeDashboardWindowSummary(ctx, coreCtx, shareddto.DashboardWindowQuery{
		Start:    "2026-03-26",
		End:      "2026-04-01",
		StreamID: &stream.ID,
	})
	if err != nil {
		t.Fatalf("window summary: %v", err)
	}
	if window.CompletedCount < 1 {
		t.Fatalf("expected completed count, got %+v", window)
	}
	if window.CarryOverCount < 1 {
		t.Fatalf("expected carry-over count, got %+v", window)
	}
	if len(window.Days) == 0 || window.Days[len(window.Days)-1].AccountabilityScore <= 0 {
		t.Fatalf("expected per-day accountability score, got %+v", window.Days)
	}

	focus, err := corecommands.ComputeFocusScoreSummary(ctx, coreCtx, shareddto.DashboardSummaryQuery{
		Start: "2026-04-01",
		End:   "2026-04-01",
	})
	if err != nil {
		t.Fatalf("focus summary: %v", err)
	}
	if focus.Score <= 0 {
		t.Fatalf("expected positive focus score, got %+v", focus)
	}

	distribution, err := corecommands.ComputeTimeDistributionSummary(ctx, coreCtx, shareddto.DashboardSummaryQuery{
		Start:    "2026-04-01",
		End:      "2026-04-01",
		GroupBy:  string(sharedtypes.DistributionGroupIssue),
		StreamID: &stream.ID,
	})
	if err != nil {
		t.Fatalf("distribution summary: %v", err)
	}
	if len(distribution.Rows) == 0 || distribution.Rows[0].Label != issueDone.Title {
		t.Fatalf("expected issue distribution row for worked issue, got %+v", distribution.Rows)
	}

	progress, err := corecommands.ComputeGoalProgressSummary(ctx, coreCtx, shareddto.DashboardSummaryQuery{
		Start:    "2026-04-01",
		End:      "2026-04-01",
		GroupBy:  string(sharedtypes.GoalProgressGroupIssue),
		StreamID: &stream.ID,
	})
	if err != nil {
		t.Fatalf("goal progress summary: %v", err)
	}
	if len(progress.Rows) == 0 {
		t.Fatalf("expected progress rows, got %+v", progress)
	}
	if progress.TotalEstimateMinutes != 90 {
		t.Fatalf("expected total estimate 90, got %+v", progress)
	}
	if progress.EstimatedItems != 1 {
		t.Fatalf("expected one estimated worked item, got %+v", progress)
	}
	if progress.EstimateBias != "over" {
		t.Fatalf("expected over estimate bias, got %+v", progress)
	}
}
