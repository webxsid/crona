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
		_ = runtimepkg.ClearPreparedTimerState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	prepared, err := t.preparedStateForSession(activeSession.ID)
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
		if prepared != nil {
			sessionID := activeSession.ID
			issueID := activeSession.IssueID
			segmentType := prepared.SegmentType
			return sharedtypes.TimerState{
				State:            "ready",
				SessionID:        &sessionID,
				IssueID:          &issueID,
				ReadySegmentType: &segmentType,
				NextSegmentType:  &segmentType,
				ElapsedSeconds:   0,
			}, nil
		}
		segmentType := sharedtypes.SessionSegmentWork
		return sharedtypes.TimerState{
			State:          "running",
			SessionID:      &activeSession.ID,
			IssueID:        &activeSession.IssueID,
			SegmentType:    &segmentType,
			ElapsedSeconds: elapsedSeconds(activeSession.StartTime, now),
		}, nil
	}
	state := "running"
	if activeSegment.SegmentType != sharedtypes.SessionSegmentWork {
		state = "paused"
	}
	nextSegment, err := t.nextSegmentForActiveState(
		ctx,
		activeSession.ID,
		activeSegment.SegmentType,
		settings,
	)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	elapsed := elapsedSeconds(activeSegment.StartTime, now)
	if activeSegment.ElapsedOffsetSeconds != nil {
		elapsed += *activeSegment.ElapsedOffsetSeconds
	}
	return sharedtypes.TimerState{
		State:           state,
		SessionID:       &activeSession.ID,
		IssueID:         &activeSession.IssueID,
		SegmentType:     &activeSegment.SegmentType,
		NextSegmentType: nextSegment,
		ElapsedSeconds:  elapsed,
	}, nil
}

func (t *TimerService) Start(
	ctx context.Context,
	repoID, streamID, issueID *int64,
) (sharedtypes.TimerState, error) {
	return t.start(ctx, repoID, streamID, issueID, false)
}

func (t *TimerService) StartIgnoringExistingStashes(
	ctx context.Context,
	repoID, streamID, issueID *int64,
) (sharedtypes.TimerState, error) {
	return t.start(ctx, repoID, streamID, issueID, true)
}

func (t *TimerService) start(
	ctx context.Context,
	repoID,
	streamID,
	issueID *int64,
	ignoreExistingStashes bool,
) (sharedtypes.TimerState, error) {
	if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
	if _, err := StartSession(ctx, t.ctx, resolvedIssueID); err != nil {
		return sharedtypes.TimerState{}, err
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
	if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
	if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
		_ = runtimepkg.ClearPreparedTimerState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	prepared, err := t.preparedStateForSession(activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if prepared != nil {
		if _, err := t.ctx.SessionSegments.StartSegment(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID, prepared.SegmentType); err != nil {
			return sharedtypes.TimerState{}, err
		}
		if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
	if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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

func (t *TimerService) End(
	ctx context.Context,
	input SessionEndInput,
) (sharedtypes.TimerState, error) {
	if _, err := StopSession(ctx, t.ctx, input); err != nil {
		return sharedtypes.TimerState{}, err
	}
	_ = runtimepkg.ClearPreparedTimerState()
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
		return runtimepkg.ClearPreparedTimerState()
	}
	prepared, err := t.preparedStateForSession(activeSession.ID)
	if err != nil {
		return err
	}
	if prepared != nil {
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
	if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
	prepared, err := t.preparedStateForSession(activeSession.ID)
	if err != nil {
		return err
	}
	if prepared != nil {
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
	boundary, err := t.nextBoundary(ctx, activeSession.ID, activeSegment.SegmentType)
	if err != nil {
		return err
	}
	if boundary == nil {
		return nil
	}
	t.scheduleBoundary(time.Duration(boundary.AfterMinutes)*time.Minute, func() {
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
		_ = runtimepkg.ClearPreparedTimerState()
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	autoStart := shouldAutoStart(boundary.NextSegment, settings)
	if autoStart {
		if boundary.NextSegment == sharedtypes.SessionSegmentWork {
			if err := runtimepkg.ClearPreparedTimerState(); err != nil {
				return sharedtypes.TimerState{}, err
			}
			if err := ResumeSession(ctx, t.ctx); err != nil {
				return sharedtypes.TimerState{}, err
			}
		} else {
			if err := runtimepkg.ClearPreparedTimerState(); err != nil {
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
		if err := runtimepkg.WritePreparedTimerState(runtimepkg.NewPreparedTimerState(sessionID, activeSession.IssueID, boundary.NextSegment)); err != nil {
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
	AfterMinutes int
}

func computeNextBoundary(
	current sharedtypes.SessionSegmentType,
	settings *sharedtypes.CoreSettings,
	completedWorkCycles int,
) *boundaryResult {
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
		return &boundaryResult{NextSegment: next, AfterMinutes: settings.WorkDurationMinutes}
	}
	if current == sharedtypes.SessionSegmentShortBreak {
		return &boundaryResult{
			NextSegment:  sharedtypes.SessionSegmentWork,
			AfterMinutes: settings.ShortBreakMinutes,
		}
	}
	if current == sharedtypes.SessionSegmentLongBreak {
		return &boundaryResult{
			NextSegment:  sharedtypes.SessionSegmentWork,
			AfterMinutes: settings.LongBreakMinutes,
		}
	}
	return nil
}

func (t *TimerService) nextBoundary(
	ctx context.Context,
	sessionID string,
	current sharedtypes.SessionSegmentType,
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
	return computeNextBoundary(current, settings, completedWorkCycles), nil
}

func (t *TimerService) nextSegmentForActiveState(
	ctx context.Context,
	sessionID string,
	current sharedtypes.SessionSegmentType,
	settings *sharedtypes.CoreSettings,
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
		if settings.TimerMode == sharedtypes.TimerModeStructured && settings.BreaksEnabled {
			next := sharedtypes.SessionSegmentWork
			return &next, nil
		}
		return nil, nil
	}
	boundary, err := t.nextBoundary(ctx, sessionID, current)
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

func (t *TimerService) preparedStateForSession(
	sessionID string,
) (*runtimepkg.PreparedTimerState, error) {
	prepared, err := runtimepkg.ReadPreparedTimerState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
			strings.Contains(err.Error(), "invalid prepared timer state") {
			_ = runtimepkg.ClearPreparedTimerState()
			return nil, nil
		}
		return nil, err
	}
	if prepared == nil {
		return nil, nil
	}
	if strings.TrimSpace(prepared.SessionID) != strings.TrimSpace(sessionID) {
		_ = runtimepkg.ClearPreparedTimerState()
		return nil, nil
	}
	return prepared, nil
}
