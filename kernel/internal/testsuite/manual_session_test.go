package testsuite

import (
	"context"
	"strings"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestManualSessionLoggingAppliesFocusStatusRules(t *testing.T) {
	ctx := context.Background()
	nowValue := "2026-03-30T09:00:00Z"
	coreCtx, _ := newTestCoreContext(t, func() string { return nowValue })

	repo, err := corecommands.CreateRepo(ctx, coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{
		Name: "Work",
	})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := corecommands.CreateStream(ctx, coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{
		RepoID: repo.ID,
		Name:   "app",
	})
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	issue, err := corecommands.CreateIssue(ctx, coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{
		StreamID: stream.ID,
		Title:    "Log manual work",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if issue.Status != sharedtypes.IssueStatusBacklog {
		t.Fatalf("expected backlog issue, got %s", issue.Status)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             issue.ID,
		Date:                "2026-03-30",
		WorkDurationSeconds: 1800,
	}); err == nil || !strings.Contains(err.Error(), "focus sessions cannot be started") {
		t.Fatalf("expected focus status denial for backlog issue, got %v", err)
	}

	planned, err := corecommands.ChangeIssueStatus(ctx, coreCtx, issue.ID, sharedtypes.IssueStatusPlanned, nil)
	if err != nil {
		t.Fatalf("plan issue: %v", err)
	}
	if planned.Status != sharedtypes.IssueStatusPlanned {
		t.Fatalf("expected planned issue, got %s", planned.Status)
	}

	nowValue = "2026-03-30T10:00:00Z"
	logged, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             issue.ID,
		Date:                "2026-03-30",
		WorkDurationSeconds: 1800,
		StartTime:           ptrTo("09:00"),
		EndTime:             ptrTo("09:30"),
	})
	if err != nil {
		t.Fatalf("log manual session: %v", err)
	}
	if logged.Source != sharedtypes.SessionSourceManual {
		t.Fatalf("expected manual source, got %+v", logged)
	}

	updated, err := coreCtx.Issues.GetByID(ctx, issue.ID, coreCtx.UserID)
	if err != nil {
		t.Fatalf("get updated issue: %v", err)
	}
	if updated == nil || updated.Status != sharedtypes.IssueStatusInProgress {
		t.Fatalf("expected issue moved to in_progress, got %+v", updated)
	}
}
