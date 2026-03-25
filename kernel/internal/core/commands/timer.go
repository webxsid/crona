package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"crona/kernel/internal/core"

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
		return sharedtypes.TimerState{State: "idle"}, nil
	}
	activeSegment, err := t.ctx.SessionSegments.GetActive(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if activeSegment == nil {
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
	elapsed := elapsedSeconds(activeSegment.StartTime, now)
	if activeSegment.ElapsedOffsetSeconds != nil {
		elapsed += *activeSegment.ElapsedOffsetSeconds
	}
	return sharedtypes.TimerState{
		State:          state,
		SessionID:      &activeSession.ID,
		IssueID:        &activeSession.IssueID,
		SegmentType:    &activeSegment.SegmentType,
		ElapsedSeconds: elapsed,
	}, nil
}

func (t *TimerService) Start(ctx context.Context, issueID *int64) (sharedtypes.TimerState, error) {
	active, err := t.ctx.Sessions.GetActiveSession(ctx, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if active != nil {
		return sharedtypes.TimerState{}, errors.New("cannot start a new session while another session is active")
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
		return sharedtypes.TimerState{}, errors.New("no issue specified and no active issue in context")
	}
	issue, err := t.ctx.Issues.GetByID(ctx, resolvedIssueID, t.ctx.UserID)
	if err != nil {
		return sharedtypes.TimerState{}, err
	}
	if issue == nil {
		return sharedtypes.TimerState{}, errors.New("issue not found")
	}
	if !sharedtypes.CanStartFocus(issue.Status) {
		return sharedtypes.TimerState{}, errors.New("focus sessions cannot be started for the current issue status")
	}
	if _, err := StartSession(ctx, t.ctx, resolvedIssueID); err != nil {
		return sharedtypes.TimerState{}, err
	}
	if nextStatus := sharedtypes.AutoStatusOnFocusStart(issue.Status); nextStatus != sharedtypes.NormalizeIssueStatus(issue.Status) {
		if _, err := changeIssueStatus(ctx, t.ctx, resolvedIssueID, nextStatus, nil, true); err != nil {
			return sharedtypes.TimerState{}, err
		}
	}
	if issueID != nil {
		if _, err := t.ctx.ActiveContext.Set(ctx, t.ctx.UserID, t.ctx.DeviceID, struct {
			RepoID   *int64
			StreamID *int64
			IssueID  *int64
		}{IssueID: &resolvedIssueID}); err == nil {
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

func (t *TimerService) Pause(ctx context.Context) (sharedtypes.TimerState, error) {
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

func (t *TimerService) End(ctx context.Context, input SessionEndInput) (sharedtypes.TimerState, error) {
	if _, err := StopSession(ctx, t.ctx, input); err != nil {
		return sharedtypes.TimerState{}, err
	}
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
	return t.ScheduleNextBoundary(ctx)
}

func (t *TimerService) RestoreFromStash(ctx context.Context, input struct {
	IssueID        int64
	SegmentType    sharedtypes.SessionSegmentType
	ElapsedSeconds int
},
) error {
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
	activeSegment, err := t.ctx.SessionSegments.GetActive(ctx, t.ctx.UserID, t.ctx.DeviceID, activeSession.ID)
	if err != nil || activeSegment == nil {
		return err
	}
	if activeSegment.SegmentType == sharedtypes.SessionSegmentRest {
		return nil
	}
	settings, err := t.ctx.CoreSettings.Get(ctx, t.ctx.UserID)
	if err != nil {
		return err
	}

	completedCycles, err := t.ctx.SessionSegments.CountWorkSegments(ctx, activeSession.ID)
	if err != nil {
		return err
	}
	boundary := computeNextBoundary(activeSegment.SegmentType, settings, completedCycles)
	if boundary == nil {
		return nil
	}
	t.scheduleBoundary(time.Duration(boundary.AfterMinutes)*time.Minute, func() {
		current, err := t.ctx.SessionSegments.GetActive(context.Background(), t.ctx.UserID, t.ctx.DeviceID, activeSession.ID)
		if err != nil || current == nil || current.SegmentType != activeSegment.SegmentType {
			return
		}
		if boundary.NextSegment == sharedtypes.SessionSegmentWork {
			_, _ = t.Resume(context.Background())
		} else {
			_ = PauseSession(context.Background(), t.ctx, boundary.NextSegment)
		}
		emit(t.ctx, sharedtypes.EventTypeTimerBoundary, t.boundaryPayload(activeSegment.SegmentType, boundary.NextSegment))
		_ = t.ScheduleNextBoundary(context.Background())
	})
	return nil
}

func (t *TimerService) boundaryPayload(from, to sharedtypes.SessionSegmentType) sharedtypes.TimerBoundaryPayload {
	payload := sharedtypes.TimerBoundaryPayload{
		From:    from,
		To:      to,
		Title:   boundaryTitle(to),
		Message: boundaryMessage(from, to),
	}
	activeContext, err := t.ctx.ActiveContext.Get(context.Background(), t.ctx.UserID, t.ctx.DeviceID)
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

func boundaryTitle(segment sharedtypes.SessionSegmentType) string {
	switch segment {
	case sharedtypes.SessionSegmentShortBreak:
		return "Short break started"
	case sharedtypes.SessionSegmentLongBreak:
		return "Long break started"
	case sharedtypes.SessionSegmentWork:
		return "Focus block started"
	default:
		return "Timer boundary reached"
	}
}

func boundaryMessage(from, to sharedtypes.SessionSegmentType) string {
	switch {
	case from == sharedtypes.SessionSegmentWork && to == sharedtypes.SessionSegmentShortBreak:
		return "Work block complete. Time for a short break"
	case from == sharedtypes.SessionSegmentWork && to == sharedtypes.SessionSegmentLongBreak:
		return "Work cycle complete. Time for a long break"
	case to == sharedtypes.SessionSegmentWork:
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

func computeNextBoundary(current sharedtypes.SessionSegmentType, settings *sharedtypes.CoreSettings, completedWorkCycles int) *boundaryResult {
	if settings.TimerMode != "structured" || !settings.BreaksEnabled {
		return nil
	}
	if current == sharedtypes.SessionSegmentWork {
		isLongBreak := settings.LongBreakEnabled && completedWorkCycles > 0 && completedWorkCycles%settings.CyclesBeforeLongBreak == 0
		next := sharedtypes.SessionSegmentShortBreak
		if isLongBreak {
			next = sharedtypes.SessionSegmentLongBreak
		}
		return &boundaryResult{NextSegment: next, AfterMinutes: settings.WorkDurationMinutes}
	}
	if current == sharedtypes.SessionSegmentShortBreak {
		return &boundaryResult{NextSegment: sharedtypes.SessionSegmentWork, AfterMinutes: settings.ShortBreakMinutes}
	}
	if current == sharedtypes.SessionSegmentLongBreak {
		return &boundaryResult{NextSegment: sharedtypes.SessionSegmentWork, AfterMinutes: settings.LongBreakMinutes}
	}
	return nil
}
