package commands

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"crona/kernel/internal/core"
	"crona/kernel/internal/events"
	runtimepkg "crona/kernel/internal/runtime"
	"crona/kernel/internal/store"
	"crona/shared/config"
	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

func TestTimerStructuredBoundaryPreparesBreakWhenAutoStartBreaksDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentWork,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil ||
		*state.ReadySegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected prepared short break, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed for prepared segment, got %d", state.ElapsedSeconds)
	}
	activeSegment, err := coreCtx.SessionSegments.GetActive(
		ctx,
		coreCtx.UserID,
		coreCtx.DeviceID,
		session.ID,
	)
	if err != nil {
		t.Fatalf("get active segment: %v", err)
	}
	if activeSegment != nil {
		t.Fatalf("expected no active segment for prepared break, got %+v", activeSegment)
	}
}

func TestTimerStructuredBoundaryAutoStartsBreakWhenEnabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentWork,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "paused" {
		t.Fatalf("expected paused state for active break, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected active short break, got %+v", state.SegmentType)
	}
	if state.ReadySegmentType != nil {
		t.Fatalf("expected no prepared segment, got %+v", state.ReadySegmentType)
	}
}

func TestTimerStructuredBoundaryPreparesWorkWhenAutoStartWorkDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentShortBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil || *state.ReadySegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected prepared work segment, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed for prepared work, got %d", state.ElapsedSeconds)
	}
}

func TestTimerStructuredBoundaryAutoStartsWorkWhenEnabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, true)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentShortBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "running" {
		t.Fatalf("expected running work state, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected active work segment, got %+v", state.SegmentType)
	}
}

func TestTimerStructuredBoundaryPreparesWorkAfterLongBreakWhenAutoStartWorkDisabled(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentLongBreak); err != nil {
		t.Fatalf("start long break: %v", err)
	}
	state, err := service.applyBoundaryTransition(
		ctx,
		session.ID,
		sharedtypes.SessionSegmentLongBreak,
		&boundaryResult{NextSegment: sharedtypes.SessionSegmentWork},
	)
	if err != nil {
		t.Fatalf("apply boundary: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected ready state, got %q", state.State)
	}
	if state.ReadySegmentType == nil || *state.ReadySegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected prepared work segment after long break, got %+v", state.ReadySegmentType)
	}
}

func TestTimerAdvanceStartsPreparedSegment(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := service.applyBoundaryTransition(ctx, session.ID, sharedtypes.SessionSegmentWork, &boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak}); err != nil {
		t.Fatalf("prepare short break: %v", err)
	}
	state, err := service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance timer: %v", err)
	}
	if state.State != "paused" {
		t.Fatalf("expected paused active break after advance, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected active short break after advance, got %+v", state.SegmentType)
	}
	if state.ReadySegmentType != nil {
		t.Fatalf(
			"expected prepared segment to clear after advance, got %+v",
			state.ReadySegmentType,
		)
	}
}

func TestTimerAdvanceFromActiveStructuredWorkStartsNextBreak(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	if _, err := StartSession(ctx, coreCtx, issue.ID); err != nil {
		t.Fatalf("start session: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType == nil ||
		*state.NextSegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf(
			"expected active work to advertise short break next, got %+v",
			state.NextSegmentType,
		)
	}
	state, err = service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance active work: %v", err)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected short break after manual advance, got %+v", state.SegmentType)
	}
}

func TestTimerAdvanceFromActiveStructuredBreakStartsWork(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, true, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := PauseSession(ctx, coreCtx, sharedtypes.SessionSegmentShortBreak); err != nil {
		t.Fatalf("start short break: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.NextSegmentType == nil || *state.NextSegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected active break to advertise work next, got %+v", state.NextSegmentType)
	}
	state, err = service.Advance(ctx)
	if err != nil {
		t.Fatalf("advance active break: %v", err)
	}
	if state.State != "running" {
		t.Fatalf("expected running work after advancing break, got %q", state.State)
	}
	if state.SegmentType == nil || *state.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected work after manual advance, got %+v", state.SegmentType)
	}
	activeSegment, err := coreCtx.SessionSegments.GetActive(
		ctx,
		coreCtx.UserID,
		coreCtx.DeviceID,
		session.ID,
	)
	if err != nil {
		t.Fatalf("get active segment: %v", err)
	}
	if activeSegment == nil || activeSegment.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected persisted active work segment, got %+v", activeSegment)
	}
}

func TestTimerRecoverBoundaryPreservesPreparedSegment(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	configureStructuredTimer(t, ctx, coreCtx, false, false)

	session, err := StartSession(ctx, coreCtx, issue.ID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := service.applyBoundaryTransition(ctx, session.ID, sharedtypes.SessionSegmentWork, &boundaryResult{NextSegment: sharedtypes.SessionSegmentShortBreak}); err != nil {
		t.Fatalf("prepare short break: %v", err)
	}
	if err := service.RecoverBoundary(ctx); err != nil {
		t.Fatalf("recover boundary: %v", err)
	}
	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get timer state: %v", err)
	}
	if state.State != "ready" {
		t.Fatalf("expected prepared state after recovery, got %q", state.State)
	}
	if state.ReadySegmentType == nil ||
		*state.ReadySegmentType != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected prepared short break after recovery, got %+v", state.ReadySegmentType)
	}
	if state.ElapsedSeconds != 0 {
		t.Fatalf("expected zero elapsed after recovery, got %d", state.ElapsedSeconds)
	}
}

func TestTimerHardLimitStartRecordsCapMetadata(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	state, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitTotalSeconds: new(3600),
			HardLimitWorkSeconds:  new(1500),
			HardLimitBreakSeconds: new(300),
		},
	)
	if err != nil {
		t.Fatalf("start hard-limit session: %v", err)
	}
	if !state.HardLimitActive {
		t.Fatal("expected hard-limit session to be active")
	}
	if state.HardLimitExpired {
		t.Fatal("expected fresh hard-limit session to not be expired")
	}
	if state.HardLimitTotalSeconds != 3600 {
		t.Fatalf("expected total seconds 3600, got %d", state.HardLimitTotalSeconds)
	}
	if state.HardLimitRemainingSeconds != 3600 {
		t.Fatalf("expected full remaining seconds, got %d", state.HardLimitRemainingSeconds)
	}
	if state.HardLimitWorkSeconds != 1500 {
		t.Fatalf("expected work seconds 1500, got %d", state.HardLimitWorkSeconds)
	}
	if state.HardLimitBreakSeconds != 300 {
		t.Fatalf("expected break seconds 300, got %d", state.HardLimitBreakSeconds)
	}
	if state.HardLimitLongBreakSeconds != 0 {
		t.Fatalf(
			"expected long break seconds to default to 0, got %d",
			state.HardLimitLongBreakSeconds,
		)
	}
	if state.HardLimitCyclesBeforeLongBreak != 0 {
		t.Fatalf(
			"expected cycles before long break to default to 0, got %d",
			state.HardLimitCyclesBeforeLongBreak,
		)
	}

	runtimeState, err := service.activeRuntimeState(ctx)
	if err != nil {
		t.Fatalf("read runtime state: %v", err)
	}
	if runtimeState == nil || !runtimeState.HasHardLimit() {
		t.Fatalf("expected runtime hard-limit state, got %+v", runtimeState)
	}
}

func TestTimerHardLimitGetStateIncludesFullPomodoroMetadata(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	startState, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitTotalSeconds:          new(7200),
			HardLimitWorkSeconds:           new(3600),
			HardLimitBreakSeconds:          new(300),
			HardLimitLongBreakSeconds:      new(900),
			HardLimitCyclesBeforeLongBreak: new(4),
		},
	)
	if err != nil {
		t.Fatalf("start hard-limit session: %v", err)
	}
	if !startState.HardLimitActive {
		t.Fatal("expected hard-limit timer to be active")
	}
	if startState.HardLimitLongBreakSeconds != 900 {
		t.Fatalf(
			"expected long break seconds 900 on start state, got %d",
			startState.HardLimitLongBreakSeconds,
		)
	}
	if startState.HardLimitCyclesBeforeLongBreak != 4 {
		t.Fatalf(
			"expected cycles before long break 4 on start state, got %d",
			startState.HardLimitCyclesBeforeLongBreak,
		)
	}

	state, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get timer state: %v", err)
	}
	if !state.HardLimitActive {
		t.Fatal("expected hard-limit timer to remain active")
	}
	if state.HardLimitLongBreakSeconds != 900 {
		t.Fatalf(
			"expected long break seconds 900 on get state, got %d",
			state.HardLimitLongBreakSeconds,
		)
	}
	if state.HardLimitCyclesBeforeLongBreak != 4 {
		t.Fatalf(
			"expected cycles before long break 4 on get state, got %d",
			state.HardLimitCyclesBeforeLongBreak,
		)
	}
}

func TestTimerHardLimitSchedulesWorkBoundaryFromFocusDuration(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	state, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitTotalSeconds:          new(180),
			HardLimitWorkSeconds:           new(60),
			HardLimitBreakSeconds:          new(30),
			HardLimitLongBreakSeconds:      new(120),
			HardLimitCyclesBeforeLongBreak: new(2),
		},
	)
	if err != nil {
		t.Fatalf("start hard-limit session: %v", err)
	}
	if state.State != "running" {
		t.Fatalf("expected running hard-limit session, got %q", state.State)
	}
	runtimeState, err := service.activeRuntimeState(ctx)
	if err != nil {
		t.Fatalf("active runtime state: %v", err)
	}
	if state.SessionID == nil || *state.SessionID == "" {
		t.Fatalf("expected session id, got %+v", state.SessionID)
	}
	boundary, err := service.nextBoundary(
		ctx,
		*state.SessionID,
		sharedtypes.SessionSegmentWork,
		runtimeState,
	)
	if err != nil {
		t.Fatalf("next boundary: %v", err)
	}
	if boundary == nil {
		t.Fatal("expected a work boundary for hard-limit session")
	}
	if boundary.AfterSeconds != 60 {
		t.Fatalf("expected work boundary after 60s, got %d", boundary.AfterSeconds)
	}
	if boundary.NextSegment != sharedtypes.SessionSegmentShortBreak {
		t.Fatalf("expected next segment short break, got %q", boundary.NextSegment)
	}
}

func TestTimerHardLimitSkipsZeroLengthBreaks(t *testing.T) {
	runtimeState := runtimepkg.NewHardLimitTimerRuntimeState(
		"session-123",
		42,
		sharedtypes.TimerHardLimitKindPomodoro,
		1800,
		600,
		0,
		0,
		4,
	)
	boundary := computeNextBoundary(
		sharedtypes.SessionSegmentShortBreak,
		nil,
		0,
		&runtimeState,
	)
	if boundary == nil {
		t.Fatal("expected a work boundary for zero-length break")
	}
	if boundary.NextSegment != sharedtypes.SessionSegmentWork {
		t.Fatalf("expected next segment work, got %q", boundary.NextSegment)
	}
	if boundary.AfterSeconds != 0 {
		t.Fatalf("expected zero-second skip, got %d", boundary.AfterSeconds)
	}
}

func TestTimerHardLimitRejectsPauseAndResume(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	if _, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{HardLimitTotalSeconds: new(1800)},
	); err != nil {
		t.Fatalf("start hard-limit session: %v", err)
	}
	if _, err := service.Pause(ctx); err == nil {
		t.Fatal("expected pause to be rejected for hard-limit sessions")
	}
	if _, err := service.Resume(ctx); err == nil {
		t.Fatal("expected resume to be rejected for hard-limit sessions")
	}
}

func TestCountdownTimerCarriesExplicitKindAndUsesDurationOnlyExtensions(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	total := 30 * 60
	state, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
			HardLimitTotalSeconds: &total,
		},
	)
	if err != nil {
		t.Fatalf("start countdown timer: %v", err)
	}
	if state.HardLimitKind != sharedtypes.TimerHardLimitKindCountdown {
		t.Fatalf("expected countdown state, got %+v", state)
	}
	if state.HardLimitWorkSeconds != total ||
		state.HardLimitBreakSeconds != 0 ||
		state.HardLimitLongBreakSeconds != 0 ||
		state.HardLimitCyclesBeforeLongBreak != 0 {
		t.Fatalf("unexpected countdown cadence: %+v", state)
	}

	if _, err := service.ExtendConfigured(ctx, shareddto.TimerExtendRequest{
		AdditionalSessions: 1,
	}); err == nil {
		t.Fatal("expected configured countdown extension to be rejected")
	}
	if _, err := service.ExtendBySessions(ctx, 1); err == nil {
		t.Fatal("expected session-count countdown extension to be rejected")
	}
	extended, err := service.Extend(ctx, 5*60)
	if err != nil {
		t.Fatalf("extend countdown by duration: %v", err)
	}
	if extended.HardLimitKind != sharedtypes.TimerHardLimitKindCountdown {
		t.Fatalf("expected extension to preserve countdown kind, got %+v", extended)
	}
	if extended.HardLimitTotalSeconds != 35*60 {
		t.Fatalf("expected 35-minute extended total, got %d", extended.HardLimitTotalSeconds)
	}
	if extended.HardLimitWorkSeconds != 35*60 {
		t.Fatalf(
			"expected an active countdown extension to extend its segment to 35 minutes, got %d",
			extended.HardLimitWorkSeconds,
		)
	}

	runtimeState, err := service.runtimeStateForSession(*state.SessionID)
	if err != nil {
		t.Fatalf("read countdown runtime state: %v", err)
	}
	if runtimeState == nil ||
		runtimeState.HardLimitKind != sharedtypes.TimerHardLimitKindCountdown {
		t.Fatalf("expected persisted countdown kind, got %+v", runtimeState)
	}
}

func TestExpiredCountdownExtensionUsesAddedDurationAndExpiresAgain(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	hardLimitEvents := 0
	unsubscribe := coreCtx.Events.Subscribe(func(event sharedtypes.KernelEvent) {
		if event.Type == sharedtypes.EventTypeTimerHardLimitReached {
			hardLimitEvents++
		}
	})
	defer unsubscribe()

	total := 15 * 60
	started, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
			HardLimitTotalSeconds: &total,
		},
	)
	if err != nil {
		t.Fatalf("start countdown: %v", err)
	}

	now = "2026-05-24T10:15:01Z"
	if _, err := service.applyHardLimitExpiry(
		ctx,
		*started.SessionID,
		sharedtypes.SessionSegmentWork,
	); err != nil {
		t.Fatalf("expire initial countdown: %v", err)
	}

	now = "2026-05-24T10:16:00Z"
	extended, err := service.Extend(ctx, 5*60)
	if err != nil {
		t.Fatalf("extend expired countdown: %v", err)
	}
	if extended.HardLimitTotalSeconds != 20*60 {
		t.Fatalf("expected cumulative total of 20 minutes, got %d", extended.HardLimitTotalSeconds)
	}
	if extended.HardLimitRemainingSeconds != 5*60 {
		t.Fatalf("expected five minutes remaining, got %d", extended.HardLimitRemainingSeconds)
	}
	if extended.HardLimitWorkSeconds != 5*60 {
		t.Fatalf(
			"expected a five-minute active countdown segment, got %d",
			extended.HardLimitWorkSeconds,
		)
	}

	now = "2026-05-24T10:21:01Z"
	expiredAgain, err := service.applyHardLimitExpiry(
		ctx,
		*started.SessionID,
		sharedtypes.SessionSegmentWork,
	)
	if err != nil {
		t.Fatalf("expire extended countdown: %v", err)
	}
	if expiredAgain.State != "expired" || !expiredAgain.HardLimitExpired {
		t.Fatalf("expected extended countdown to expire again, got %+v", expiredAgain)
	}
	if hardLimitEvents != 2 {
		t.Fatalf("expected two hard-limit events across both expiries, got %d", hardLimitEvents)
	}
}

func TestCountdownTimerRejectsPomodoroConfigurationBeforeStartingSession(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	total := 30 * 60
	work := 25 * 60
	if _, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitKind:         sharedtypes.TimerHardLimitKindCountdown,
			HardLimitTotalSeconds: &total,
			HardLimitWorkSeconds:  &work,
		},
	); err == nil {
		t.Fatal("expected countdown cadence configuration to be rejected")
	}
	active, err := coreCtx.Sessions.GetActiveSession(ctx, coreCtx.UserID)
	if err != nil {
		t.Fatalf("load active session: %v", err)
	}
	if active != nil {
		t.Fatalf("expected rejected countdown not to create a session, got %+v", active)
	}
	if _, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitKind: sharedtypes.TimerHardLimitKindCountdown,
		},
	); err == nil {
		t.Fatal("expected countdown without a duration to be rejected")
	}
}

func TestTimerHardLimitExpiryAndExtend(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	coreCtx, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	var hardLimitEvents []sharedtypes.KernelEvent
	var extendedEvents []sharedtypes.KernelEvent
	var endedEvents []sharedtypes.KernelEvent
	unsubscribe := coreCtx.Events.Subscribe(func(event sharedtypes.KernelEvent) {
		if event.Type == sharedtypes.EventTypeTimerHardLimitReached {
			hardLimitEvents = append(hardLimitEvents, event)
		}
		if event.Type == sharedtypes.EventTypeTimerExtended {
			extendedEvents = append(extendedEvents, event)
		}
		if event.Type == sharedtypes.EventTypeSessionEnded {
			endedEvents = append(endedEvents, event)
		}
	})
	defer unsubscribe()

	startState, err := service.Start(
		ctx,
		nil,
		new(issue.StreamID),
		new(issue.ID),
		&shareddto.TimerStartRequest{
			HardLimitTotalSeconds: new(60),
			HardLimitWorkSeconds:  new(25),
			HardLimitBreakSeconds: new(5),
		},
	)
	if err != nil {
		t.Fatalf("start hard-limit session: %v", err)
	}
	if startState.State != "running" {
		t.Fatalf("expected running hard-limit session, got %q", startState.State)
	}
	if startState.SessionID == nil || *startState.SessionID == "" {
		t.Fatalf("expected timer start to return a session id, got %+v", startState.SessionID)
	}

	now = "2026-05-24T10:01:01Z"
	expiredState, err := service.applyHardLimitExpiry(
		ctx,
		*startState.SessionID,
		sharedtypes.SessionSegmentWork,
	)
	if err != nil {
		t.Fatalf("apply hard-limit expiry: %v", err)
	}
	if expiredState.State != "expired" {
		t.Fatalf("expected expired state after cap expiry, got %q", expiredState.State)
	}
	if !expiredState.HardLimitActive || !expiredState.HardLimitExpired {
		t.Fatalf("expected hard-limit expired flags, got %+v", expiredState)
	}
	if expiredState.HardLimitRemainingSeconds != 0 {
		t.Fatalf(
			"expected no remaining hard-limit time, got %d",
			expiredState.HardLimitRemainingSeconds,
		)
	}
	activeSegmentAfterExpiry, err := service.ctx.SessionSegments.GetActive(
		ctx,
		service.ctx.UserID,
		service.ctx.DeviceID,
		*startState.SessionID,
	)
	if err != nil {
		t.Fatalf("load active segment after expiry: %v", err)
	}
	if activeSegmentAfterExpiry != nil {
		t.Fatalf("expected expiry to end the active segment, got %+v", activeSegmentAfterExpiry)
	}
	now = "2026-05-24T10:05:00Z"
	frozenState, err := service.GetState(ctx)
	if err != nil {
		t.Fatalf("get frozen expired state: %v", err)
	}
	if frozenState.State != "expired" {
		t.Fatalf("expected expired state to stay expired, got %q", frozenState.State)
	}
	if frozenState.ElapsedSeconds != 0 {
		t.Fatalf(
			"expected expired state to stop accumulating elapsed time, got %d",
			frozenState.ElapsedSeconds,
		)
	}
	if frozenState.HardLimitRemainingSeconds != 0 {
		t.Fatalf(
			"expected hard-limit remaining time to stay frozen at 0, got %d",
			frozenState.HardLimitRemainingSeconds,
		)
	}
	if len(hardLimitEvents) != 1 {
		t.Fatalf("expected exactly one hard-limit event, got %d", len(hardLimitEvents))
	}

	extendedState, err := service.Extend(ctx, 600)
	if err != nil {
		t.Fatalf("extend hard-limit session: %v", err)
	}
	if extendedState.State != "running" {
		t.Fatalf("expected running state after extend, got %q", extendedState.State)
	}
	if !extendedState.HardLimitActive || extendedState.HardLimitExpired {
		t.Fatalf(
			"expected active non-expired hard-limit session after extend, got %+v",
			extendedState,
		)
	}
	if extendedState.HardLimitTotalSeconds != 660 {
		t.Fatalf(
			"expected total seconds 660 after extend, got %d",
			extendedState.HardLimitTotalSeconds,
		)
	}
	if extendedState.HardLimitRemainingSeconds <= 0 {
		t.Fatalf(
			"expected positive remaining hard-limit time after extend, got %d",
			extendedState.HardLimitRemainingSeconds,
		)
	}
	if extendedState.HardLimitRemainingSeconds != 600 {
		t.Fatalf(
			"expected extend to rebase remaining hard-limit time to 600s, got %d",
			extendedState.HardLimitRemainingSeconds,
		)
	}
	afterExtendSegment, err := service.ctx.SessionSegments.GetActive(
		ctx,
		service.ctx.UserID,
		service.ctx.DeviceID,
		*startState.SessionID,
	)
	if err != nil {
		t.Fatalf("load active segment after extend: %v", err)
	}
	if afterExtendSegment == nil {
		t.Fatal("expected active segment after extend")
	}
	if afterExtendSegment.SegmentType != sharedtypes.SessionSegmentWork {
		t.Fatalf(
			"expected extend to restart the expired work segment, got %q",
			afterExtendSegment.SegmentType,
		)
	}
	boundaryAfterExtend, err := service.nextBoundary(
		ctx,
		*startState.SessionID,
		sharedtypes.SessionSegmentWork,
		func() *runtimepkg.TimerRuntimeState {
			state, stateErr := service.activeRuntimeState(ctx)
			if stateErr != nil {
				t.Fatalf("active runtime state after extend: %v", stateErr)
			}
			return state
		}(),
	)
	if err != nil {
		t.Fatalf("next boundary after extend: %v", err)
	}
	if boundaryAfterExtend == nil || boundaryAfterExtend.AfterSeconds != 25 {
		t.Fatalf("expected restarted work segment boundary after 25s, got %+v", boundaryAfterExtend)
	}
	if len(extendedEvents) != 1 {
		t.Fatalf("expected exactly one timer.extended event, got %d", len(extendedEvents))
	}
	var extendedPayload sharedtypes.SessionEventPayload
	if err := json.Unmarshal(extendedEvents[0].Payload, &extendedPayload); err != nil {
		t.Fatalf("decode timer.extended payload: %v", err)
	}
	if extendedPayload.SessionID != *startState.SessionID {
		t.Fatalf(
			"expected timer.extended session %q, got %q",
			*startState.SessionID,
			extendedPayload.SessionID,
		)
	}

	now = "2026-05-24T10:10:00Z"
	commitMessage := "feat: finish hard-limit session"
	endedState, err := service.End(ctx, SessionEndInput{CommitMessage: &commitMessage})
	if err != nil {
		t.Fatalf("end extended hard-limit session: %v", err)
	}
	if endedState.State != "idle" {
		t.Fatalf("expected idle state after end, got %q", endedState.State)
	}
	lastSession, err := service.ctx.Sessions.GetLastSessionForUser(ctx, service.ctx.UserID)
	if err != nil {
		t.Fatalf("load last session: %v", err)
	}
	if lastSession == nil || lastSession.DurationSeconds == nil {
		t.Fatalf("expected ended session with duration, got %+v", lastSession)
	}
	if *lastSession.DurationSeconds < 0 {
		t.Fatalf("expected non-negative duration, got %d", *lastSession.DurationSeconds)
	}
	if len(endedEvents) != 1 {
		t.Fatalf("expected exactly one session.ended event, got %d", len(endedEvents))
	}
	var endedPayload sharedtypes.SessionEventPayload
	if err := json.Unmarshal(endedEvents[0].Payload, &endedPayload); err != nil {
		t.Fatalf("decode session.ended payload: %v", err)
	}
	if endedPayload.SessionID != *startState.SessionID {
		t.Fatalf(
			"expected session.ended session %q, got %q",
			*startState.SessionID,
			endedPayload.SessionID,
		)
	}
}

func TestTimerEndRequiresCommitMessage(t *testing.T) {
	ctx := context.Background()
	now := "2026-05-24T10:00:00Z"
	_, service, issue := newTimerTestContext(t, func() string { return now })
	mustMakeIssuePlanned(t, ctx, service.ctx, issue.ID)

	if _, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), nil); err != nil {
		t.Fatalf("start session: %v", err)
	}
	if _, err := service.End(ctx, SessionEndInput{}); err == nil {
		t.Fatal("expected timer end without commit message to fail")
	}
	blank := "   "
	if _, err := service.End(ctx, SessionEndInput{CommitMessage: &blank}); err == nil {
		t.Fatal("expected timer end with blank commit message to fail")
	}
}

func newTimerTestContext(
	t *testing.T,
	now func() string,
) (*core.Context, *TimerService, sharedtypes.Issue) {
	t.Helper()

	base := t.TempDir()
	t.Setenv(config.EnvVarRuntimeDir, base)

	db, err := store.Open(filepath.Join(base, "crona.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := store.InitSchema(context.Background(), db.DB()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	registry := store.NewRegistry(db.DB())
	coreCtx := core.NewContext(
		db,
		registry,
		"local",
		"test-device",
		filepath.Join(base, "scratch"),
		now,
		events.NewBus(),
	)
	if err := coreCtx.InitDefaults(context.Background()); err != nil {
		t.Fatalf("init defaults: %v", err)
	}

	repo, err := CreateRepo(context.Background(), coreCtx, struct {
		Name        string
		Description *string
		Color       *string
	}{Name: "Work"})
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	stream, err := CreateStream(context.Background(), coreCtx, struct {
		RepoID      int64
		Name        string
		Description *string
		Visibility  *sharedtypes.StreamVisibility
	}{RepoID: repo.ID, Name: "app"})
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	issue, err := CreateIssue(context.Background(), coreCtx, struct {
		StreamID        int64
		Title           string
		Description     *string
		EstimateMinutes *int
		Notes           *string
		TodoForDate     *string
	}{StreamID: stream.ID, Title: "Structured timer test"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	return coreCtx, GetTimerService(coreCtx), issue
}

func configureStructuredTimer(
	t *testing.T,
	ctx context.Context,
	coreCtx *core.Context,
	autoStartBreaks bool,
	autoStartWork bool,
) {
	t.Helper()

	settings := map[sharedtypes.CoreSettingsKey]any{
		sharedtypes.CoreSettingsKeyTimerMode:             string(sharedtypes.TimerModeStructured),
		sharedtypes.CoreSettingsKeyBreaksEnabled:         true,
		sharedtypes.CoreSettingsKeyWorkDurationMinutes:   25,
		sharedtypes.CoreSettingsKeyShortBreakMinutes:     5,
		sharedtypes.CoreSettingsKeyLongBreakMinutes:      15,
		sharedtypes.CoreSettingsKeyLongBreakEnabled:      true,
		sharedtypes.CoreSettingsKeyCyclesBeforeLongBreak: 4,
		sharedtypes.CoreSettingsKeyAutoStartBreaks:       autoStartBreaks,
		sharedtypes.CoreSettingsKeyAutoStartWork:         autoStartWork,
	}
	for key, value := range settings {
		if err := coreCtx.CoreSettings.SetSetting(ctx, coreCtx.UserID, key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}
}

func mustMakeIssuePlanned(
	t *testing.T,
	ctx context.Context,
	coreCtx *core.Context,
	issueID int64,
) {
	t.Helper()
	if _, err := changeIssueStatus(
		ctx,
		coreCtx,
		issueID,
		sharedtypes.IssueStatusPlanned,
		nil,
		true,
	); err != nil {
		t.Fatalf("mark issue planned: %v", err)
	}
}
