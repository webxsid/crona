package commands

import (
	"context"
	"testing"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestTimerStateReportsPomodoroCycles(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)
	if _, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), &shareddto.TimerStartRequest{
		HardLimitTotalSeconds: new(3600),
		HardLimitWorkSeconds:  new(1500),
		HardLimitBreakSeconds: new(300),
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.HardLimitCycles != 2 || state.CompletedWorkCycles != 0 {
		t.Fatalf(
			"expected cycle 0 of 2, got %d of %d",
			state.CompletedWorkCycles,
			state.HardLimitCycles,
		)
	}
	now = "2026-05-24T10:25:00Z"
	if _, err := service.Advance(ctx); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if state, err = service.GetState(ctx); err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.CompletedWorkCycles != 1 {
		t.Fatalf("expected one finished focus cycle, got %d", state.CompletedWorkCycles)
	}
}

func TestTimerStateReportsPriorWorkOnTheIssue(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)
	end, duration := "2026-05-24T10:20:00Z", 1200
	if _, err := service.ctx.Sessions.CreateCompleted(ctx, sharedtypes.Session{
		IssueID: issue.ID, StartTime: now, EndTime: &end, DurationSeconds: &duration,
	}, service.ctx.UserID, service.ctx.DeviceID, now); err != nil {
		t.Fatalf("create completed session: %v", err)
	}
	now = "2026-05-24T11:00:00Z"
	if _, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), nil); err != nil {
		t.Fatalf("restart: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.IssueWorkedSeconds != 1200 || state.IssueCompletedSessions != 1 {
		t.Fatalf(
			"expected 1200s over 1 session, got %ds over %d",
			state.IssueWorkedSeconds,
			state.IssueCompletedSessions,
		)
	}
}

func TestCountdownReportsNoNextPhase(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)
	if _, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), &shareddto.TimerStartRequest{
		HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
		HardLimitTotalSeconds: new(600),
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType != nil {
		t.Fatalf("expected no next phase for a countdown, got %q", *state.NextSegmentType)
	}
}

func TestPomodoroLastPhaseReportsNoNextPhase(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)
	if _, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), &shareddto.TimerStartRequest{
		HardLimitTotalSeconds: new(1800),
		HardLimitWorkSeconds:  new(1500),
		HardLimitBreakSeconds: new(300),
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType == nil ||
		*state.NextSegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected a short break after the focus, got %v", state.NextSegmentType)
	}
	now = "2026-05-24T10:25:00Z"
	if _, err := service.Advance(ctx); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if state, err = service.GetState(ctx); err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType != nil {
		t.Fatalf(
			"expected the session to end after its last break, got next %q",
			*state.NextSegmentType,
		)
	}
}
