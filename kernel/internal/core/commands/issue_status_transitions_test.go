package commands

import (
	"context"
	"testing"

	sharedtypes "crona/shared/types"
)

func TestIssueStatusTransitionsReturnsAllowedStatuses(t *testing.T) {
	ctx := context.Background()
	coreCtx, _, issue := newTimerTestContext(t, func() string {
		return "2026-07-26T10:00:00Z"
	})

	response, err := IssueStatusTransitions(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("get issue status transitions: %v", err)
	}
	if response.CurrentStatus != sharedtypes.IssueStatusBacklog {
		t.Fatalf("expected backlog, got %q", response.CurrentStatus)
	}
	want := sharedtypes.AllowedIssueStatusTransitions(sharedtypes.IssueStatusBacklog)
	if len(response.AllowedStatuses) != len(want) {
		t.Fatalf("expected %d statuses, got %d", len(want), len(response.AllowedStatuses))
	}
	for index := range want {
		if response.AllowedStatuses[index] != want[index] {
			t.Fatalf(
				"status %d: expected %q, got %q",
				index,
				want[index],
				response.AllowedStatuses[index],
			)
		}
	}
	if response.BlockedReason != nil {
		t.Fatalf("expected no blocked reason, got %q", *response.BlockedReason)
	}
}

func TestIssueStatusTransitionsReportsActiveSessionRestriction(t *testing.T) {
	ctx := context.Background()
	coreCtx, timerService, issue := newTimerTestContext(t, func() string {
		return "2026-07-26T10:00:00Z"
	})
	mustMakeIssuePlanned(t, ctx, coreCtx, issue.ID)
	if _, err := timerService.Start(ctx, nil, new(issue.StreamID), new(issue.ID), nil); err != nil {
		t.Fatalf("start session: %v", err)
	}

	response, err := IssueStatusTransitions(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("get issue status transitions: %v", err)
	}
	if len(response.AllowedStatuses) != 0 {
		t.Fatalf("expected no allowed statuses, got %v", response.AllowedStatuses)
	}
	if response.BlockedReason == nil || *response.BlockedReason == "" {
		t.Fatal("expected an active-session blocked reason")
	}
}

func TestIssueStatusTransitionsRejectsMissingIssue(t *testing.T) {
	ctx := context.Background()
	coreCtx, _, _ := newTimerTestContext(t, func() string {
		return "2026-07-26T10:00:00Z"
	})

	if _, err := IssueStatusTransitions(ctx, coreCtx, 999999); err == nil {
		t.Fatal("expected missing issue to fail")
	}
}
