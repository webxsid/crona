package commands

import (
	"context"
	"testing"
)

func TestStopSessionUsesSegmentDerivedDuration(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	if err := coreCtx.SessionSegments.ApplyElapsedOffset(ctx, session.ID, 85); err != nil {
		t.Fatalf("apply elapsed offset: %v", err)
	}

	now = "2026-05-24T10:10:00Z"
	stopped, err := StopSession(ctx, coreCtx, SessionEndInput{})
	if err != nil {
		t.Fatalf("stop session: %v", err)
	}
	if stopped == nil || stopped.DurationSeconds == nil {
		t.Fatalf("expected stopped session with duration, got %+v", stopped)
	}
	if *stopped.DurationSeconds < 85 || *stopped.DurationSeconds > 90 {
		t.Fatalf("expected segment-derived duration near 85s, got %d", *stopped.DurationSeconds)
	}
}
