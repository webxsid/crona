package testsuite

import (
	"context"
	"testing"

	corecommands "crona/kernel/internal/core/commands"
	sharedtypes "crona/shared/types"
)

func TestListAllIssuesIncludesWorkedSeconds(t *testing.T) {
	ctx := context.Background()
	currentNow := "2026-05-25T09:00:00Z"
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
	}{RepoID: repo.ID, Name: "cli"})
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
	}{StreamID: stream.ID, Title: "Track spent time"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if _, err := corecommands.ChangeIssueStatus(ctx, coreCtx, issue.ID, sharedtypes.IssueStatusPlanned, nil); err != nil {
		t.Fatalf("plan issue: %v", err)
	}
	if _, err := corecommands.LogManualSession(ctx, coreCtx, corecommands.ManualSessionInput{
		IssueID:             issue.ID,
		Date:                "2026-05-25",
		WorkDurationSeconds: 4500,
	}); err != nil {
		t.Fatalf("log manual session: %v", err)
	}

	allIssues, err := corecommands.ListAllIssues(ctx, coreCtx)
	if err != nil {
		t.Fatalf("list all issues: %v", err)
	}
	if len(allIssues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(allIssues))
	}
	if allIssues[0].WorkedSeconds != 4500 {
		t.Fatalf("expected worked seconds to survive list-all mapping, got %d", allIssues[0].WorkedSeconds)
	}
}
