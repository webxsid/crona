package commands

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

func setIssuePinned(ctx context.Context, coreCtx *core.Context, issueID int64, pinned bool) error {
	_, err := UpdateIssue(ctx, coreCtx, issueID, struct {
		Title           sharedtypes.Patch[string]
		Description     sharedtypes.Patch[string]
		EstimateMinutes sharedtypes.Patch[int]
		Notes           sharedtypes.Patch[string]
		PinnedDaily     sharedtypes.Patch[bool]
	}{PinnedDaily: sharedtypes.Patch[bool]{Set: true, Value: &pinned}})
	return err
}

func pinLimitIssues(t *testing.T, n int) (context.Context, *core.Context, []int64) {
	t.Helper()
	ctx := context.Background()
	coreCtx, _, first := newTimerTestContext(t, func() string {
		return "2026-07-26T10:00:00Z"
	})
	ids := []int64{first.ID}
	for i := 1; i < n; i++ {
		issue, err := CreateIssue(ctx, coreCtx, struct {
			StreamID        int64
			Title           string
			Description     *string
			EstimateMinutes *int
			Notes           *string
			TodoForDate     *string
		}{StreamID: first.StreamID, Title: fmt.Sprintf("Pin candidate %d", i)})
		if err != nil {
			t.Fatalf("create issue %d: %v", i, err)
		}
		ids = append(ids, issue.ID)
	}
	return ctx, coreCtx, ids
}

func TestUpdateIssueRefusesSixthPin(t *testing.T) {
	ctx, coreCtx, ids := pinLimitIssues(t, 6)
	for _, id := range ids[:sharedtypes.MaxPinnedDailyIssues] {
		if err := setIssuePinned(ctx, coreCtx, id, true); err != nil {
			t.Fatalf("pin issue %d: %v", id, err)
		}
	}
	err := setIssuePinned(ctx, coreCtx, ids[5], true)
	if err == nil || !strings.Contains(err.Error(), "5") {
		t.Fatalf("expected the sixth pin refused with the limit in the message, got %v", err)
	}
	if err := setIssuePinned(ctx, coreCtx, ids[0], true); err != nil {
		t.Fatalf("expected re-pinning a pinned issue to pass, got %v", err)
	}
	if err := setIssuePinned(ctx, coreCtx, ids[0], false); err != nil {
		t.Fatalf("unpin: %v", err)
	}
	if err := setIssuePinned(ctx, coreCtx, ids[5], true); err != nil {
		t.Fatalf("expected a pin to fit after unpinning, got %v", err)
	}
}

func TestUpdateIssuePinLimitIgnoresClosedIssues(t *testing.T) {
	ctx, coreCtx, ids := pinLimitIssues(t, 6)
	for _, id := range ids[:sharedtypes.MaxPinnedDailyIssues] {
		if err := setIssuePinned(ctx, coreCtx, id, true); err != nil {
			t.Fatalf("pin issue %d: %v", id, err)
		}
	}
	reason := "no longer needed"
	if _, err := changeIssueStatus(ctx, coreCtx, ids[0], sharedtypes.IssueStatusAbandoned, &reason, true); err != nil {
		t.Fatalf("abandon issue: %v", err)
	}
	if err := setIssuePinned(ctx, coreCtx, ids[5], true); err != nil {
		t.Fatalf("expected a closed pinned issue not to count, got %v", err)
	}
}
