package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"crona/kernel/internal/core"
	runtimepkg "crona/kernel/internal/runtime"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type TimerService struct {
	ctx           *core.Context
	boundaryTimer *time.Timer
	mu            sync.Mutex
}

var (
	timerMu       sync.Mutex
	timerServices = map[*core.Context]*TimerService{}
)

func GetTimerService(c *core.Context) *TimerService {
	timerMu.Lock()
	defer timerMu.Unlock()
	if service, ok := timerServices[c]; ok {
		return service
	}
	service := &TimerService{ctx: c}
	timerServices[c] = service
	return service
}

func (t *TimerService) GetState(ctx context.Context) (sharedtypes.TimerState, error) {
	now := t.ctx.Now()
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		activeSession.ID,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	base := hardLimitTimerState(runtimeState, activeSession.StartTime, now)
	if activeSegment == nil {
		if runtimeState != nil && runtimeState.HasPreparedSegment() {
			sessionID := activeSession.ID
			issueID := activeSession.IssueID
			segmentType := *runtimeState.PreparedSegmentType
			state := "ready"
			if runtimeState.HardLimitExpired {
				state = "expired"
			}
			return sharedtypes.TimerState{
				State:                          state,
				SessionID:                      &sessionID,
				SessionStartTime:               &activeSession.StartTime,
				IssueID:                        &issueID,
				ReadySegmentType:               &segmentType,
				NextSegmentType:                &segmentType,
				ElapsedSeconds:                 0,
				HardLimitActive:                base.HardLimitActive,
				HardLimitExpired:               base.HardLimitExpired,
				HardLimitTotalSeconds:          base.HardLimitTotalSeconds,
				HardLimitRemainingSeconds:      base.HardLimitRemainingSeconds,
				HardLimitWorkSeconds:           base.HardLimitWorkSeconds,
				HardLimitBreakSeconds:          base.HardLimitBreakSeconds,
				HardLimitLongBreakSeconds:      base.HardLimitLongBreakSeconds,
				HardLimitCyclesBeforeLongBreak: base.HardLimitCyclesBeforeLongBreak,
			}, nil
		}
		segmentType := sharedtypes.SessionSegmentWork
		return sharedtypes.TimerState{
			State:                          "running",
			SessionID:                      &activeSession.ID,
			SessionStartTime:               &activeSession.StartTime,
			IssueID:                        &activeSession.IssueID,
			SegmentType:                    &segmentType,
			ElapsedSeconds:                 elapsedSeconds(activeSession.StartTime, now),
			HardLimitActive:                base.HardLimitActive,
			HardLimitExpired:               base.HardLimitExpired,
			HardLimitTotalSeconds:          base.HardLimitTotalSeconds,
			HardLimitRemainingSeconds:      base.HardLimitRemainingSeconds,
			HardLimitWorkSeconds:           base.HardLimitWorkSeconds,
			HardLimitBreakSeconds:          base.HardLimitBreakSeconds,
			HardLimitLongBreakSeconds:      base.HardLimitLongBreakSeconds,
			HardLimitCyclesBeforeLongBreak: base.HardLimitCyclesBeforeLongBreak,
		}, nil
	}
	state := "running"
	if activeSegment.SegmentType != sharedtypes.SessionSegmentWork &&
		(runtimeState == nil || !runtimeState.HasHardLimit()) {
		state = "paused"
	}
	if runtimeState != nil && runtimeState.HardLimitExpired {
		state = "expired"
	}
	nextSegment, err := t.nextSegmentForActiveState(
		ctx,
		activeSession.ID,
		activeSegment.SegmentType,
		settings,
		runtimeState,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	elapsed := elapsedSeconds(activeSegment.StartTime, now)
	if activeSegment.ElapsedOffsetSeconds != nil {
		elapsed += *activeSegment.ElapsedOffsetSeconds
	}
	return sharedtypes.TimerState{
		State:                          state,
		SessionID:                      &activeSession.ID,
		SessionStartTime:               &activeSession.StartTime,
		IssueID:                        &activeSession.IssueID,
		SegmentType:                    &activeSegment.SegmentType,
		SegmentStartTime:               &activeSegment.StartTime,
		SegmentElapsedOffsetSeconds:    activeSegment.ElapsedOffsetSeconds,
		NextSegmentType:                nextSegment,
		ElapsedSeconds:                 elapsed,
		HardLimitActive:                base.HardLimitActive,
		HardLimitExpired:               base.HardLimitExpired,
		HardLimitTotalSeconds:          base.HardLimitTotalSeconds,
		HardLimitRemainingSeconds:      base.HardLimitRemainingSeconds,
		HardLimitWorkSeconds:           base.HardLimitWorkSeconds,
		HardLimitBreakSeconds:          base.HardLimitBreakSeconds,
		HardLimitLongBreakSeconds:      base.HardLimitLongBreakSeconds,
		HardLimitCyclesBeforeLongBreak: base.HardLimitCyclesBeforeLongBreak,
	}, nil
}

func (t *TimerService) Start(
	ctx context.Context,
	repoID, streamID, issueID *int64,
	hardLimit *shareddto.TimerStartRequest,
) (sharedtypes.TimerState, error) {
	return t.start(ctx, repoID, streamID, issueID, false, hardLimit)
}

func (t *TimerService) StartIgnoringExistingStashes(
	ctx context.Context,
	repoID, streamID, issueID *int64,
	hardLimit *shareddto.TimerStartRequest,
) (sharedtypes.TimerState, error) {
	return t.start(ctx, repoID, streamID, issueID, true, hardLimit)
}

func (t *TimerService) start(
	ctx context.Context,
	repoID,
	streamID,
	issueID *int64,
	ignoreExistingStashes bool,
	input *shareddto.TimerStartRequest,
) (sharedtypes.TimerState, error) {
	if err := runtimepkg.ClearTimerRuntimeState(); err != nil {
		return sharedtypes.TimerState{}, err
	}
	active, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if active != nil {
		return sharedtypes.TimerState{}, errors.New(
			"cannot start a new session while another session is active",
		)
	}
	var resolvedIssueID int64
	if issueID != nil {
		resolvedIssueID = *issueID
	}
	if resolvedIssueID == 0 {
		activeContext, err := t.ctx.ActiveContext.Get(ctx, t.ctx.UserID, t.ctx.DeviceID)
		if err != nil {
			return sharedtypes.TimerState{}, err
		}
		if activeContext != nil && activeContext.IssueID != nil {
			resolvedIssueID = *activeContext.IssueID
		}
	}
	if resolvedIssueID == 0 {
		return sharedtypes.TimerState{}, errors.New(
			"no issue specified and no active issue in context",
		)
	}
	issue, err := t.ctx.Issues.GetByID(ctx, resolvedIssueID, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if issue == nil {
		return sharedtypes.TimerState{}, errors.New("issue not found")
	}
	if !sharedtypes.CanStartFocus(issue.Status) {
		return sharedtypes.TimerState{}, errors.New(
			"focus sessions cannot be started for the current issue status",
		)
	}
	if !ignoreExistingStashes {
		stashes, err := t.ctx.Stash.ListByIssue(ctx, resolvedIssueID, t.ctx.UserID)
		if err != nil {
			return sharedtypes.TimerState{}, err
		}
		if len(stashes) > 0 {
			return sharedtypes.TimerState{}, StashConflictError{
				Conflict: sharedtypes.StashConflict{
					IssueID: resolvedIssueID,
					Stashes: stashes,
				},
			}
		}
	}
	session, err := StartSession(ctx, t.ctx, resolvedIssueID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if input != nil && input.HardLimitTotalSeconds != nil && *input.HardLimitTotalSeconds > 0 {
		workSeconds := derefInt(input.HardLimitWorkSeconds)
		if workSeconds <= 0 {
			workSeconds = *input.HardLimitTotalSeconds
		}
		breakSeconds := derefInt(input.HardLimitBreakSeconds)
		longBreakSeconds := derefInt(input.HardLimitLongBreakSeconds)
		cyclesBeforeLongBreak := derefInt(input.HardLimitCyclesBeforeLongBreak)
		if err := runtimepkg.WriteTimerRuntimeState(
			func() runtimepkg.TimerRuntimeState {
				state := runtimepkg.NewHardLimitTimerRuntimeState(
					session.ID,
					resolvedIssueID,
					*input.HardLimitTotalSeconds,
					workSeconds,
					breakSeconds,
					longBreakSeconds,
					cyclesBeforeLongBreak,
				)
				state.HardLimitElapsedStartedAt = session.StartTime
				return state
			}(),
		); err != nil {
			return sharedtypes.TimerState{}, err
		}
	}
	if nextStatus := sharedtypes.AutoStatusOnFocusStart(issue.Status); nextStatus != sharedtypes.NormalizeIssueStatus(
		issue.Status,
	) {
		if _, err := changeIssueStatus(ctx, t.ctx, resolvedIssueID, nextStatus, nil, true); err != nil {
			return sharedtypes.TimerState{}, err
		}
	}
	if issueID != nil {
		contextRepoID := repoID
		contextStreamID := streamID
		if _, err := t.ctx.ActiveContext.Set(ctx, t.ctx.UserID, t.ctx.DeviceID, struct {
			RepoID   *int64
			StreamID *int64
			IssueID  *int64
		}{RepoID: contextRepoID, StreamID: contextStreamID, IssueID: &resolvedIssueID}); err == nil {
			payload, _ := json.Marshal(sharedtypes.ContextChangedPayload{
				DeviceID: t.ctx.DeviceID,
				IssueID:  &resolvedIssueID,
			})
			t.ctx.Events.Emit(sharedtypes.KernelEvent{
				Type:    sharedtypes.EventTypeContextIssueChanged,
				Payload: payload,
			})
		}
	}
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

type StashConflictError struct {
	Conflict sharedtypes.StashConflict
}

func (e StashConflictError) Error() string {
	count := len(e.Conflict.Stashes)
	if count == 1 {
		return fmt.Sprintf("cannot start focus: 1 stash exists for issue #%d", e.Conflict.IssueID)
	}
	return fmt.Sprintf(
		"cannot start focus: %d stashes exist for issue #%d",
		count,
		e.Conflict.IssueID,
	)
}

func (e StashConflictError) ProtocolErrorCode() string {
	return protocol.ErrorCodeStashConflict
}

func (e StashConflictError) ProtocolErrorData() any {
	return e.Conflict
}

func (t *TimerService) Pause(ctx context.Context) (sharedtypes.TimerState, error) {
	runtimeState, err := t.activeRuntimeState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState != nil && runtimeState.HasHardLimit() {
		return sharedtypes.TimerState{}, errors.New("hard-limit sessions cannot be paused")
	}
	if err := runtimepkg.ClearTimerRuntimeState(); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if err := PauseSession(ctx, t.ctx, sharedtypes.SessionSegmentRest); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

func (t *TimerService) Resume(ctx context.Context) (sharedtypes.TimerState, error) {
	runtimeState, err := t.activeRuntimeState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState != nil && runtimeState.HasHardLimit() {
		return sharedtypes.TimerState{}, errors.New("hard-limit sessions cannot be resumed")
	}
	if err := runtimepkg.ClearTimerRuntimeState(); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if err := ResumeSession(ctx, t.ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

func (t *TimerService) Advance(ctx context.Context) (sharedtypes.TimerState, error) {
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState != nil && runtimeState.HardLimitExpired {
		return sharedtypes.TimerState{}, errors.New(
			"hard-limit session has expired and must be committed, stashed, or extended",
		)
	}
	if runtimeState != nil && runtimeState.HasPreparedSegment() {
		if _, err := t.ctx.SessionSegments.StartSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID, *runtimeState.PreparedSegmentType); err != nil {
			return sharedtypes.TimerState{}, err
		}
		runtimeState.PreparedSegmentType = nil
		if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
			return sharedtypes.TimerState{}, err
		}
		if err := t.ScheduleNextBoundary(ctx); err != nil {
			return sharedtypes.TimerState{}, err
		}
		state, err := t.GetState(ctx)
		if err != nil {
			return sharedtypes.TimerState{}, err
		}
		emit(t.ctx, sharedtypes.EventTypeTimerState, state)
		return state, nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		activeSession.ID,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSegment == nil {
		return sharedtypes.TimerState{}, errors.New("no prepared timer segment")
	}
	nextSegment, err := t.nextSegmentForActiveState(
		ctx,
		activeSession.ID,
		activeSegment.SegmentType,
		settings,
		runtimeState,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if nextSegment == nil {
		return sharedtypes.TimerState{}, errors.New("timer cannot advance from the current state")
	}
	if err := t.ctx.SessionSegments.EndActiveSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if _, err := t.ctx.SessionSegments.StartSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID, *nextSegment); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState != nil {
		runtimeState.PreparedSegmentType = nil
	}
	if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
		return sharedtypes.TimerState{}, err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(
		t.ctx,
		sharedtypes.EventTypeTimerBoundary,
		t.boundaryPayload(activeSegment.SegmentType, *nextSegment, true),
	)
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	return state, nil
}

func (t *TimerService) Extend(
	ctx context.Context,
	additionalSeconds int,
) (sharedtypes.TimerState, error) {
	if additionalSeconds <= 0 {
		return sharedtypes.TimerState{}, errors.New("extension duration must be positive")
	}
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil {
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	return t.extendHardLimit(ctx, activeSession, runtimeState, additionalSeconds)
}

func (t *TimerService) ExtendBySessions(
	ctx context.Context,
	additionalSessions int,
) (sharedtypes.TimerState, error) {
	if additionalSessions <= 0 {
		return sharedtypes.TimerState{}, errors.New("sessions to add must be positive")
	}
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState == nil || !runtimeState.HasHardLimit() {
		return sharedtypes.TimerState{}, errors.New("no hard-limit session is active")
	}
	additionalSeconds, err := t.hardLimitExtensionSecondsForSessions(ctx, activeSession.ID, additionalSessions, runtimeState)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	return t.extendHardLimit(ctx, activeSession, runtimeState, additionalSeconds)
}

func (t *TimerService) ExtendConfigured(
	ctx context.Context,
	input shareddto.TimerExtendRequest,
) (sharedtypes.TimerState, error) {
	if input.AdditionalSessions <= 0 {
		return sharedtypes.TimerState{}, errors.New("sessions to add must be positive")
	}
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState == nil || !runtimeState.HasHardLimit() {
		return sharedtypes.TimerState{}, errors.New("no hard-limit session is active")
	}
	workSeconds := runtimeState.HardLimitWorkSeconds
	if input.HardLimitWorkSeconds != nil && *input.HardLimitWorkSeconds > 0 {
		workSeconds = *input.HardLimitWorkSeconds
	}
	breakSeconds := runtimeState.HardLimitBreakSeconds
	if input.HardLimitBreakSeconds != nil {
		breakSeconds = max(0, *input.HardLimitBreakSeconds)
	}
	longBreakSeconds := runtimeState.HardLimitLongBreakSeconds
	if input.HardLimitLongBreakSeconds != nil {
		longBreakSeconds = max(0, *input.HardLimitLongBreakSeconds)
	}
	cyclesBeforeLongBreak := runtimeState.HardLimitCyclesBeforeLongBreak
	if input.HardLimitCyclesBeforeLongBreak != nil {
		cyclesBeforeLongBreak = max(0, *input.HardLimitCyclesBeforeLongBreak)
	}
	totalPerSession := 0
	if input.HardLimitTotalSeconds != nil {
		totalPerSession = max(0, *input.HardLimitTotalSeconds)
	}
	if totalPerSession <= 0 {
		totalPerSession = workSeconds
	}
	additionalSeconds := input.AdditionalSeconds
	if additionalSeconds <= 0 {
		additionalSeconds = totalPerSession * input.AdditionalSessions
	}
	runtimeState.HardLimitWorkSeconds = workSeconds
	runtimeState.HardLimitBreakSeconds = breakSeconds
	runtimeState.HardLimitLongBreakSeconds = longBreakSeconds
	runtimeState.HardLimitCyclesBeforeLongBreak = cyclesBeforeLongBreak
	return t.extendHardLimit(ctx, activeSession, runtimeState, additionalSeconds)
}

func (t *TimerService) extendHardLimit(
	ctx context.Context,
	activeSession *sharedtypes.Session,
	runtimeState *runtimepkg.TimerRuntimeState,
	additionalSeconds int,
) (sharedtypes.TimerState, error) {
	if runtimeState == nil || !runtimeState.HasHardLimit() {
		return sharedtypes.TimerState{}, errors.New("no hard-limit session is active")
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		activeSession.ID,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	segmentToStart := sharedtypes.SessionSegmentWork
	if activeSegment != nil {
		segmentToStart = activeSegment.SegmentType
	} else if runtimeState.HasPreparedSegment() {
		segmentToStart = *runtimeState.PreparedSegmentType
	}
	if _, err := t.ctx.SessionSegments.StartSegment(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		activeSession.ID,
		segmentToStart,
	); err != nil {
		return sharedtypes.TimerState{}, err
	}
	runtimeState.PreparedSegmentType = nil
	runtimeState.HardLimitTotalSeconds += additionalSeconds
	runtimeState.HardLimitExpired = false
	runtimeState.HardLimitExpiredAt = ""
	runtimeState.HardLimitElapsedStartedAt = t.ctx.Now()
	if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return sharedtypes.TimerState{}, err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

func (t *TimerService) hardLimitExtensionSecondsForSessions(
	ctx context.Context,
	sessionID string,
	additionalSessions int,
	runtimeState *runtimepkg.TimerRuntimeState,
) (int, error) {
	if additionalSessions <= 0 {
		return 0, errors.New("sessions to add must be positive")
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		sessionID,
	)
	if err != nil {
		return 0, err
	}
	currentSegment := sharedtypes.SessionSegmentWork
	if activeSegment != nil {
		currentSegment = activeSegment.SegmentType
	} else if runtimeState != nil && runtimeState.HasPreparedSegment() {
		currentSegment = *runtimeState.PreparedSegmentType
	}
	completedWorkCycles, err := t.ctx.SessionSegments.CountWorkSegments(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	if currentSegment == sharedtypes.SessionSegmentWork &&
		(activeSegment != nil || runtimeState == nil || !runtimeState.HardLimitExpired) {
		completedWorkCycles++
	}
	total := 0
	seg := currentSegment
	cycles := completedWorkCycles
	for added := 0; added < additionalSessions; {
		if seg != sharedtypes.SessionSegmentWork {
			boundary := computeNextBoundary(seg, nil, cycles, runtimeState)
			if boundary == nil {
				return total, nil
			}
			total += boundary.AfterSeconds
			seg = boundary.NextSegment
			continue
		}
		boundary := computeNextBoundary(seg, nil, cycles, runtimeState)
		if boundary == nil {
			return total, nil
		}
		total += boundary.AfterSeconds
		seg = boundary.NextSegment
		cycles++
		if seg != sharedtypes.SessionSegmentWork {
			boundary = computeNextBoundary(seg, nil, cycles, runtimeState)
			if boundary == nil {
				return total, nil
			}
			total += boundary.AfterSeconds
			seg = boundary.NextSegment
		}
		added++
	}
	return total, nil
}

func (t *TimerService) End(
	ctx context.Context,
	input SessionEndInput,
) (sharedtypes.TimerState, error) {
	if _, err := StopSession(ctx, t.ctx, input); err != nil {
		return sharedtypes.TimerState{}, err
	}
	_ = runtimepkg.ClearTimerRuntimeState()
	t.clearBoundary()
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

func (t *TimerService) RecoverBoundary(ctx context.Context) error {
	t.clearBoundary()
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return err
	}
	if activeSession == nil {
		return runtimepkg.ClearTimerRuntimeState()
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return err
	}
	if runtimeState != nil && runtimeState.HasPreparedSegment() {
		return nil
	}
	return t.ScheduleNextBoundary(ctx)
}

func (t *TimerService) RestoreFromStash(ctx context.Context, input struct {
	IssueID        int64
	SegmentType    sharedtypes.SessionSegmentType
	ElapsedSeconds int
},
) error {
	if err := runtimepkg.ClearTimerRuntimeState(); err != nil {
		return err
	}
	session, err := StartSession(ctx, t.ctx, input.IssueID)
	if err != nil {
		return err
	}
	if _, err := t.ctx.SessionSegments.StartSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, session.ID, input.SegmentType); err != nil {
		return err
	}
	if err := t.ctx.SessionSegments.ApplyElapsedOffset(ctx, session.ID, input.ElapsedSeconds); err != nil {
		return err
	}
	if err := t.ScheduleNextBoundary(ctx); err != nil {
		return err
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return nil
}

func (t *TimerService) ScheduleNextBoundary(ctx context.Context) error {
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil || activeSession == nil {
		return err
	}
	runtimeState, err := t.runtimeStateForSession(activeSession.ID)
	if err != nil {
		return err
	}
	if runtimeState != nil && (runtimeState.HasPreparedSegment() || runtimeState.HardLimitExpired) {
		return nil
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		activeSession.ID,
	)
	if err != nil || activeSegment == nil {
		return err
	}
	if activeSegment.SegmentType == sharedtypes.SessionSegmentRest {
		return nil
	}
	boundary, err := t.nextBoundary(ctx, activeSession.ID, activeSegment.SegmentType, runtimeState)
	if err != nil {
		return err
	}
	delay, hardLimitFirst, ok := t.nextTimerDelay(activeSession.StartTime, boundary, runtimeState)
	if !ok {
		return nil
	}
	t.scheduleBoundary(delay, func() {
		if hardLimitFirst {
			_, _ = t.applyHardLimitExpiry(
				context.Background(),
				activeSession.ID,
				activeSegment.SegmentType,
			)
			return
		}
		_, _ = t.applyBoundaryTransition(
			context.Background(),
			activeSession.ID,
			activeSegment.SegmentType,
			boundary,
		)
	})
	return nil
}

func (t *TimerService) applyBoundaryTransition(
	ctx context.Context,
	sessionID string,
	currentType sharedtypes.SessionSegmentType,
	boundary *boundaryResult,
) (sharedtypes.TimerState, error) {
	current, err := t.ctx.SessionSegments.GetActive(ctx, t.ctx.UserID, t.ctx.DeviceID, sessionID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if current == nil || current.SegmentType != currentType {
		return t.GetState(ctx)
	}
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil || activeSession.ID != sessionID {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	runtimeState, err := t.runtimeStateForSession(sessionID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	autoStart := shouldAutoStart(boundary.NextSegment, settings)
	if runtimeState != nil && runtimeState.HasHardLimit() {
		autoStart = true
		runtimeState.PreparedSegmentType = nil
	}
	if autoStart {
		if boundary.NextSegment == sharedtypes.SessionSegmentWork {
			if runtimeState != nil {
				runtimeState.PreparedSegmentType = nil
			}
			if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
				return sharedtypes.TimerState{}, err
			}
			if err := ResumeSession(ctx, t.ctx); err != nil {
				return sharedtypes.TimerState{}, err
			}
		} else {
			if runtimeState != nil {
				runtimeState.PreparedSegmentType = nil
			}
			if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
				return sharedtypes.TimerState{}, err
			}
			if err := PauseSession(ctx, t.ctx, boundary.NextSegment); err != nil {
				return sharedtypes.TimerState{}, err
			}
		}
	} else {
		if err := t.ctx.SessionSegments.EndActiveSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, sessionID); err != nil {
			return sharedtypes.TimerState{}, err
		}
		if runtimeState == nil {
			state := runtimepkg.NewPreparedTimerRuntimeState(sessionID, activeSession.IssueID, boundary.NextSegment)
			runtimeState = &state
		} else {
			runtimeState.PreparedSegmentType = segmentPtr(boundary.NextSegment)
		}
		if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
			return sharedtypes.TimerState{}, err
		}
	}
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	emit(
		t.ctx,
		sharedtypes.EventTypeTimerBoundary,
		t.boundaryPayload(currentType, boundary.NextSegment, autoStart),
	)
	if autoStart {
		if err := t.ScheduleNextBoundary(ctx); err != nil {
			return sharedtypes.TimerState{}, err
		}
	}
	return state, nil
}

func (t *TimerService) applyHardLimitExpiry(
	ctx context.Context,
	sessionID string,
	currentType sharedtypes.SessionSegmentType,
) (sharedtypes.TimerState, error) {
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSession == nil || activeSession.ID != sessionID {
		_ = runtimepkg.ClearTimerRuntimeState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	runtimeState, err := t.runtimeStateForSession(sessionID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if runtimeState == nil || !runtimeState.HasHardLimit() || runtimeState.HardLimitExpired {
		return t.GetState(ctx)
	}
	runtimeState.HardLimitElapsedOffsetSeconds = hardLimitConsumedSeconds(
		runtimeState,
		activeSession.StartTime,
		t.ctx.Now(),
	)
	runtimeState.HardLimitElapsedStartedAt = ""
	if err := t.ctx.SessionSegments.EndActiveSegment(
		ctx,
		t.ctx.UserID,
		t.ctx.DeviceID,
		sessionID,
	); err != nil {
		return sharedtypes.TimerState{}, err
	}
	runtimeState.PreparedSegmentType = segmentPtr(currentType)
	runtimeState.HardLimitExpired = true
	runtimeState.HardLimitExpiredAt = t.ctx.Now()
	if err := t.writeOrClearRuntimeState(runtimeState); err != nil {
		return sharedtypes.TimerState{}, err
	}
	t.clearBoundary()
	state, err := t.GetState(ctx)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	emit(
		t.ctx,
		sharedtypes.EventTypeTimerHardLimitReached,
		sharedtypes.TimerHardLimitReachedPayload{
			SessionID:                      sessionID,
			IssueID:                        activeSession.IssueID,
			SegmentType:                    segmentPtr(currentType),
			HardLimitTotalSeconds:          runtimeState.HardLimitTotalSeconds,
			HardLimitWorkSeconds:           runtimeState.HardLimitWorkSeconds,
			HardLimitBreakSeconds:          runtimeState.HardLimitBreakSeconds,
			HardLimitLongBreakSeconds:      runtimeState.HardLimitLongBreakSeconds,
			HardLimitCyclesBeforeLongBreak: runtimeState.HardLimitCyclesBeforeLongBreak,
			ElapsedSeconds: max(
				0,
				runtimeState.HardLimitTotalSeconds-state.HardLimitRemainingSeconds,
			),
		},
	)
	emit(t.ctx, sharedtypes.EventTypeTimerState, state)
	return state, nil
}

func (t *TimerService) boundaryPayload(
	from, to sharedtypes.SessionSegmentType,
	started bool,
) sharedtypes.TimerBoundaryPayload {
	payload := sharedtypes.TimerBoundaryPayload{
		From:    from,
		To:      to,
		Started: started,
		Title:   boundaryTitle(to, started),
		Message: boundaryMessage(from, to, started),
	}
	activeContext, err := t.ctx.ActiveContext.Get(
		context.Background(),
		t.ctx.UserID,
		t.ctx.DeviceID,
	)
	if err != nil || activeContext == nil {
		return payload
	}
	if activeContext.RepoName != nil && strings.TrimSpace(*activeContext.RepoName) != "" {
		payload.RepoName = activeContext.RepoName
	}
	if activeContext.StreamName != nil && strings.TrimSpace(*activeContext.StreamName) != "" {
		payload.StreamName = activeContext.StreamName
	}
	if activeContext.IssueID != nil {
		payload.IssueID = activeContext.IssueID
	}
	if activeContext.IssueTitle != nil && strings.TrimSpace(*activeContext.IssueTitle) != "" {
		payload.IssueTitle = activeContext.IssueTitle
		payload.Message = payload.Message + ": " + strings.TrimSpace(*activeContext.IssueTitle)
	}
	return payload
}

func boundaryTitle(segment sharedtypes.SessionSegmentType, started bool) string {
	switch segment {
	case sharedtypes.SessionSegmentShortBreak:
		if !started {
			return "Short break ready"
		}
		return "Short break started"
	case sharedtypes.SessionSegmentLongBreak:
		if !started {
			return "Long break ready"
		}
		return "Long break started"
	case sharedtypes.SessionSegmentWork:
		if !started {
			return "Focus block ready"
		}
		return "Focus block started"
	default:
		return "Timer boundary reached"
	}
}

func boundaryMessage(from, to sharedtypes.SessionSegmentType, started bool) string {
	switch {
	case from == sharedtypes.SessionSegmentWork && to == sharedtypes.SessionSegmentShortBreak:
		if !started {
			return "Work block complete. Short break is ready to start"
		}
		return "Work block complete. Time for a short break"
	case from == sharedtypes.SessionSegmentWork && to == sharedtypes.SessionSegmentLongBreak:
		if !started {
			return "Work cycle complete. Long break is ready to start"
		}
		return "Work cycle complete. Time for a long break"
	case to == sharedtypes.SessionSegmentWork:
		if !started {
			return "Break complete. Focus block is ready to start"
		}
		return "Break complete. Back to focused work"
	default:
		return "Structured timer boundary reached"
	}
}

func (t *TimerService) scheduleBoundary(delay time.Duration, callback func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.boundaryTimer != nil {
		t.boundaryTimer.Stop()
	}
	t.boundaryTimer = time.AfterFunc(delay, callback)
}

func (t *TimerService) clearBoundary() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.boundaryTimer != nil {
		t.boundaryTimer.Stop()
		t.boundaryTimer = nil
	}
}

func (t *TimerService) ClearBoundary() {
	t.clearBoundary()
}

type boundaryResult struct {
	NextSegment  sharedtypes.SessionSegmentType
	AfterSeconds int
}

func computeNextBoundary(
	current sharedtypes.SessionSegmentType,
	settings *sharedtypes.CoreSettings,
	completedWorkCycles int,
	runtimeState *runtimepkg.TimerRuntimeState,
) *boundaryResult {
	if runtimeState != nil && runtimeState.HasHardLimit() {
		if current == sharedtypes.SessionSegmentWork {
			duration := runtimeState.HardLimitWorkSeconds
			if duration <= 0 {
				return nil
			}
			next := sharedtypes.SessionSegmentShortBreak
			if hardLimitShouldUseLongBreak(runtimeState, completedWorkCycles) {
				next = sharedtypes.SessionSegmentLongBreak
			}
			return &boundaryResult{
				NextSegment:  next,
				AfterSeconds: duration,
			}
		}
		if current == sharedtypes.SessionSegmentShortBreak ||
			current == sharedtypes.SessionSegmentLongBreak {
			duration := runtimeState.HardLimitBreakSeconds
			if current == sharedtypes.SessionSegmentLongBreak &&
				runtimeState.HardLimitLongBreakSeconds > 0 {
				duration = runtimeState.HardLimitLongBreakSeconds
			}
			if duration <= 0 {
				return &boundaryResult{
					NextSegment:  sharedtypes.SessionSegmentWork,
					AfterSeconds: 0,
				}
			}
			return &boundaryResult{
				NextSegment:  sharedtypes.SessionSegmentWork,
				AfterSeconds: duration,
			}
		}
		return nil
	}
	if settings.TimerMode != "structured" || !settings.BreaksEnabled {
		return nil
	}
	if current == sharedtypes.SessionSegmentWork {
		isLongBreak := settings.LongBreakEnabled && completedWorkCycles > 0 &&
			completedWorkCycles%settings.CyclesBeforeLongBreak == 0
		next := sharedtypes.SessionSegmentShortBreak
		if isLongBreak {
			next = sharedtypes.SessionSegmentLongBreak
		}
		return &boundaryResult{NextSegment: next, AfterSeconds: settings.WorkDurationMinutes * 60}
	}
	if current == sharedtypes.SessionSegmentShortBreak {
		return &boundaryResult{
			NextSegment:  sharedtypes.SessionSegmentWork,
			AfterSeconds: settings.ShortBreakMinutes * 60,
		}
	}
	if current == sharedtypes.SessionSegmentLongBreak {
		return &boundaryResult{
			NextSegment:  sharedtypes.SessionSegmentWork,
			AfterSeconds: settings.LongBreakMinutes * 60,
		}
	}
	return nil
}

func hardLimitShouldUseLongBreak(
	state *runtimepkg.TimerRuntimeState,
	completedWorkCycles int,
) bool {
	if state == nil || !state.HasHardLimit() {
		return false
	}
	if state.HardLimitLongBreakSeconds <= 0 || state.HardLimitCyclesBeforeLongBreak <= 0 {
		return false
	}
	if completedWorkCycles <= 0 {
		return false
	}
	return completedWorkCycles%state.HardLimitCyclesBeforeLongBreak == 0
}

func (t *TimerService) nextBoundary(
	ctx context.Context,
	sessionID string,
	current sharedtypes.SessionSegmentType,
	runtimeState *runtimepkg.TimerRuntimeState,
) (*boundaryResult, error) {
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	completedWorkCycles, err := t.ctx.SessionSegments.CountWorkSegments(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if current == sharedtypes.SessionSegmentWork {
		completedWorkCycles++
	}
	return computeNextBoundary(current, settings, completedWorkCycles, runtimeState), nil
}

func (t *TimerService) nextSegmentForActiveState(
	ctx context.Context,
	sessionID string,
	current sharedtypes.SessionSegmentType,
	settings *sharedtypes.CoreSettings,
	runtimeState *runtimepkg.TimerRuntimeState,
) (*sharedtypes.SessionSegmentType, error) {
	if settings == nil {
		var err error
		settings, err = t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
		if err != nil {
			return nil, err
		}
	}
	if settings == nil {
		return nil, nil
	}
	if current == sharedtypes.SessionSegmentRest {
		if runtimeState != nil && runtimeState.HasHardLimit() {
			next := sharedtypes.SessionSegmentWork
			return &next, nil
		}
		if settings.TimerMode == sharedtypes.TimerModeStructured && settings.BreaksEnabled {
			next := sharedtypes.SessionSegmentWork
			return &next, nil
		}
		return nil, nil
	}
	boundary, err := t.nextBoundary(ctx, sessionID, current, runtimeState)
	if err != nil || boundary == nil {
		return nil, err
	}
	next := boundary.NextSegment
	return &next, nil
}

func shouldAutoStart(
	segment sharedtypes.SessionSegmentType,
	settings *sharedtypes.CoreSettings,
) bool {
	if settings == nil {
		return false
	}
	if segment == sharedtypes.SessionSegmentWork {
		return settings.AutoStartWork
	}
	return settings.AutoStartBreaks
}

func (t *TimerService) runtimeStateForSession(
	sessionID string,
) (*runtimepkg.TimerRuntimeState, error) {
	prepared, err := runtimepkg.ReadTimerRuntimeState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
			strings.Contains(err.Error(), "invalid timer runtime state") {
			_ = runtimepkg.ClearTimerRuntimeState()
			return nil, nil
		}
		return nil, err
	}
	if prepared == nil {
		return nil, nil
	}
	if strings.TrimSpace(prepared.SessionID) != strings.TrimSpace(sessionID) {
		_ = runtimepkg.ClearTimerRuntimeState()
		return nil, nil
	}
	return prepared, nil
}

func (t *TimerService) activeRuntimeState(
	ctx context.Context,
) (*runtimepkg.TimerRuntimeState, error) {
	activeSession, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil || activeSession == nil {
		return nil, err
	}
	return t.runtimeStateForSession(activeSession.ID)
}

func (t *TimerService) writeOrClearRuntimeState(state *runtimepkg.TimerRuntimeState) error {
	if state == nil || (!state.HasHardLimit() && !state.HasPreparedSegment()) {
		return runtimepkg.ClearTimerRuntimeState()
	}
	state.RecordedAt = time.Now().UTC().Format(time.RFC3339)
	return runtimepkg.WriteTimerRuntimeState(*state)
}

func (t *TimerService) nextTimerDelay(
	sessionStart string,
	boundary *boundaryResult,
	runtimeState *runtimepkg.TimerRuntimeState,
) (time.Duration, bool, bool) {
	boundaryDelay := time.Duration(0)
	if boundary != nil && boundary.AfterSeconds > 0 {
		boundaryDelay = time.Duration(boundary.AfterSeconds) * time.Second
	}
	hardLimitDelay, hasHardLimit := hardLimitDelay(runtimeState, sessionStart, t.ctx.Now())
	switch {
	case hasHardLimit && boundaryDelay > 0:
		if hardLimitDelay <= boundaryDelay {
			return hardLimitDelay, true, true
		}
		return boundaryDelay, false, true
	case hasHardLimit:
		return hardLimitDelay, true, true
	case boundaryDelay > 0:
		return boundaryDelay, false, true
	default:
		return 0, false, false
	}
}

func hardLimitTimerState(
	state *runtimepkg.TimerRuntimeState,
	sessionStart string,
	now string,
) sharedtypes.TimerState {
	timerState := sharedtypes.TimerState{}
	if state == nil || !state.HasHardLimit() {
		return timerState
	}
	remaining, _ := hardLimitRemaining(state, sessionStart, now)
	timerState.HardLimitActive = true
	timerState.HardLimitExpired = state.HardLimitExpired
	timerState.HardLimitTotalSeconds = state.HardLimitTotalSeconds
	timerState.HardLimitRemainingSeconds = remaining
	timerState.HardLimitWorkSeconds = state.HardLimitWorkSeconds
	timerState.HardLimitBreakSeconds = state.HardLimitBreakSeconds
	timerState.HardLimitLongBreakSeconds = state.HardLimitLongBreakSeconds
	timerState.HardLimitCyclesBeforeLongBreak = state.HardLimitCyclesBeforeLongBreak
	return timerState
}

func hardLimitRemaining(
	state *runtimepkg.TimerRuntimeState,
	sessionStart string,
	now string,
) (int, bool) {
	if state == nil || !state.HasHardLimit() {
		return 0, false
	}
	remaining := state.HardLimitTotalSeconds - hardLimitConsumedSeconds(state, sessionStart, now)
	if remaining < 0 {
		remaining = 0
	}
	return remaining, true
}

func hardLimitConsumedSeconds(
	state *runtimepkg.TimerRuntimeState,
	sessionStart string,
	now string,
) int {
	if state == nil || !state.HasHardLimit() {
		return 0
	}
	consumed := state.HardLimitElapsedOffsetSeconds
	if !state.HardLimitExpired {
		startRef := state.HardLimitElapsedStartedAt
		if strings.TrimSpace(startRef) == "" {
			startRef = sessionStart
		}
		consumed += elapsedSeconds(startRef, now)
	}
	if consumed < 0 {
		return 0
	}
	if consumed > state.HardLimitTotalSeconds {
		return state.HardLimitTotalSeconds
	}
	return consumed
}

func hardLimitDelay(
	state *runtimepkg.TimerRuntimeState,
	sessionStart string,
	now string,
) (time.Duration, bool) {
	remaining, ok := hardLimitRemaining(state, sessionStart, now)
	if !ok {
		return 0, false
	}
	return time.Duration(max(1, remaining)) * time.Second, true
}

func segmentPtr(segment sharedtypes.SessionSegmentType) *sharedtypes.SessionSegmentType {
	return &segment
}
