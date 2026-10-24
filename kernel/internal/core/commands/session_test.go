package commands

import (
	"context"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
	"sync"
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

func TestStopSessionInvalidIssueLeavesWorkRecording(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	c, _, issue := newTimerTestContext(t, func() string { return now })
	session, err := StartSession(ctx, c, issue.ID)
	if err != nil {
		t.Fatal(err)
	}
	missing := int64(999999)
	commit := "Work"
	if _, err := StopSession(ctx, c, SessionEndInput{CommitMessage: &commit, IssueID: &missing}); err == nil {
		t.Fatal("expected invalid issue rejection")
	}
	active, err := c.SessionSegments.GetActive(ctx, c.UserID, c.DeviceID, session.ID)
	if err != nil || active == nil || active.EndTime != nil {
		t.Fatalf("work recording ended after invalid request: %+v %v", active, err)
	}
	if err := c.SessionSegments.ApplyElapsedOffset(ctx, session.ID, 300); err != nil {
		t.Fatal(err)
	}
	now = "2026-05-24T10:05:00Z"
	stopped, err := StopSession(ctx, c, SessionEndInput{CommitMessage: &commit})
	if err != nil {
		t.Fatal(err)
	}
	if stopped.DurationSeconds == nil ||
		(*stopped.DurationSeconds < 300 || *stopped.DurationSeconds > 305) {
		t.Fatalf("lost work duration: %+v", stopped)
	}
}
func TestStopSessionEnforcesAdhocAttribution(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	c, service, _ := newTimerTestContext(t, func() string { return now })
	started, err := service.Start(ctx, nil, nil, nil, &shareddto.TimerStartRequest{Adhoc: true})
	if err != nil {
		t.Fatal(err)
	}
	commit := "Work"
	if _, err := StopSession(ctx, c, SessionEndInput{CommitMessage: &commit}); err == nil {
		t.Fatal("direct completion bypassed attribution")
	}
	active, err := c.SessionSegments.GetActive(ctx, c.UserID, c.DeviceID, *started.SessionID)
	if err != nil || active == nil {
		t.Fatalf("rejected completion ended segment: %v", err)
	}
	stopped, err := StopSession(
		ctx,
		c,
		SessionEndInput{
			CommitMessage:   &commit,
			AdhocResolution: sharedtypes.AdhocSessionResolutionKeep,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	issue, err := c.Issues.GetByID(ctx, stopped.IssueID, c.UserID)
	if err != nil || issue == nil || !issue.Adhoc || issue.Status != sharedtypes.IssueStatusDone {
		t.Fatalf("missing visible retained issue: %+v %v", issue, err)
	}
}

func TestConcurrentAdhocCompletionCreatesOneAttribution(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		name := "session ends"
		if mixed {
			name = "session and timer ends"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			c, timer, _ := newTimerTestContext(t, func() string { return "2026-05-24T10:00:00Z" })
			started, err := timer.Start(
				ctx,
				nil,
				nil,
				nil,
				&shareddto.TimerStartRequest{Adhoc: true},
			)
			if err != nil {
				t.Fatal(err)
			}
			summary := "Concurrent ad hoc work"
			input := SessionEndInput{
				CommitMessage:   &summary,
				AdhocResolution: sharedtypes.AdhocSessionResolutionKeep,
			}
			start := make(chan struct{})
			var wg sync.WaitGroup
			var results [2]error
			for i := range results {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					select {
					case <-start:
					case <-ctx.Done():
						results[index] = ctx.Err()
						return
					}
					if mixed && index == 1 {
						_, results[index] = timer.End(ctx, input)
					} else {
						_, results[index] = StopSession(ctx, c, input)
					}
				}(i)
			}
			close(start)
			wg.Wait()
			for _, err := range results {
				if err != nil {
					t.Fatalf("completion failed: %v", err)
				}
			}
			issues, err := c.Issues.ListAll(ctx, c.UserID)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, issue := range issues {
				if issue.Adhoc {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("concurrent completion created %d ad hoc issues", count)
			}
			session, err := c.Sessions.GetByID(ctx, *started.SessionID, c.UserID)
			if err != nil || session == nil || session.EndTime == nil {
				t.Fatalf("session not completed: %+v %v", session, err)
			}
		})
	}
}
