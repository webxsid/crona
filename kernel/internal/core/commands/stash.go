package commands

import (
	"context"
	"encoding/json"
	"errors"

	"crona/kernel/internal/core"
	"crona/kernel/internal/sessionnotes"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

func StashPush(ctx context.Context, c *core.Context, stashNote *string) (sharedtypes.Stash, error) {
	now := c.Now()
	activeContext, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
	if err != nil {
		return sharedtypes.Stash{}, err
	}
	if activeContext == nil {
		return sharedtypes.Stash{}, errors.New("no active context to stash")
	}
	activeSession, err := c.Sessions.GetActiveSession(ctx, c.UserID)
	if err != nil {
		return sharedtypes.Stash{}, err
	}

	var sessionID *string
	var pausedType *sharedtypes.SessionSegmentType
	var elapsed *int

	if activeSession != nil {
		var endedSegments []sharedtypes.SessionSegment
		activeSegment, err := c.SessionSegments.GetActive(
			ctx,
			c.UserID,
			c.DeviceID,
			activeSession.ID,
		)
		if err != nil {
			return sharedtypes.Stash{}, err
		}
		if activeSegment != nil {
			elapsedValue := elapsedSeconds(activeSegment.StartTime, now)
			sessionID = &activeSession.ID
			pausedType = &activeSegment.SegmentType
			elapsed = &elapsedValue
			if err := c.SessionSegments.EndActiveSegment(ctx, c.UserID, c.DeviceID, activeSession.ID); err != nil {
				return sharedtypes.Stash{}, err
			}
		}
		endedSegments, err = c.SessionSegments.ListBySession(ctx, activeSession.ID)
		if err != nil {
			return sharedtypes.Stash{}, err
		}
		if _, err := c.Sessions.Stop(ctx, activeSession.ID, struct {
			EndTime         string
			DurationSeconds int
			Notes           *string
		}{
			EndTime:         now,
			DurationSeconds: sessionnotes.TotalSegmentDurationSeconds(endedSegments),
		}, c.UserID, c.DeviceID, now); err != nil {
			return sharedtypes.Stash{}, err
		}
	}

	stash := sharedtypes.Stash{
		ID:                uuid.NewString(),
		UserID:            c.UserID,
		DeviceID:          c.DeviceID,
		RepoID:            activeContext.RepoID,
		StreamID:          activeContext.StreamID,
		IssueID:           activeContext.IssueID,
		SessionID:         sessionID,
		PausedSegmentType: pausedType,
		ElapsedSeconds:    elapsed,
		Note:              stashNote,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := c.Stash.Save(ctx, stash); err != nil {
		return sharedtypes.Stash{}, err
	}
	if err := c.ActiveContext.Clear(ctx, c.UserID, c.DeviceID); err != nil {
		return sharedtypes.Stash{}, err
	}
	emit(c, sharedtypes.EventTypeStashCreated, stash)
	payload, _ := json.Marshal(sharedtypes.ContextClearedPayload{DeviceID: c.DeviceID})
	c.Events.Emit(
		sharedtypes.KernelEvent{Type: sharedtypes.EventTypeContextCleared, Payload: payload},
	)
	return stash, nil
}

func StashPop(ctx context.Context, c *core.Context, timer *TimerService, stashID string) error {
	stash, err := c.Stash.Get(ctx, stashID, c.UserID)
	if err != nil {
		return err
	}
	if stash == nil {
		return errors.New("stash not found")
	}
	activeSession, err := c.Sessions.GetActiveSession(ctx, c.UserID)
	if err != nil {
		return err
	}
	if activeSession != nil {
		return errors.New("cannot apply stash while a focus session is active")
	}
	if _, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID:   stash.RepoID,
		StreamID: stash.StreamID,
		IssueID:  stash.IssueID,
	}); err != nil {
		return err
	}

	if stash.SessionID != nil && stash.IssueID != nil && stash.PausedSegmentType != nil &&
		timer != nil {
		if err := timer.RestoreFromStash(ctx, struct {
			IssueID        int64
			SegmentType    sharedtypes.SessionSegmentType
			ElapsedSeconds int
		}{
			IssueID:        *stash.IssueID,
			SegmentType:    *stash.PausedSegmentType,
			ElapsedSeconds: derefInt(stash.ElapsedSeconds),
		}); err != nil {
			return err
		}
	}
	if err := c.Stash.Delete(ctx, stashID, c.UserID); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeStashApplied, stash)
	if active, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID); err == nil && active != nil {
		return emitContextSnapshot(c, sharedtypes.EventTypeContextRepoChanged, active)
	}
	return nil
}

func CommitStashWithoutPop(ctx context.Context, c *core.Context, stashID string) error {
	stash, err := c.Stash.Get(ctx, stashID, c.UserID)
	if err != nil {
		return err
	}
	if stash == nil {
		return errors.New("stash not found")
	}
	if err := c.Stash.Delete(ctx, stashID, c.UserID); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityStash,
		EntityID:  stashID,
		Action:    sharedtypes.OpActionDelete,
		Payload:   map[string]any{},
		Timestamp: c.Now(),
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeStashDropped, sharedtypes.StashEventPayload{
		ID:       stashID,
		DeviceID: c.DeviceID,
	})
	return nil
}

func StashDrop(ctx context.Context, c *core.Context, stashID string) error {
	if err := c.Stash.Delete(ctx, stashID, c.UserID); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityStash,
		EntityID:  stashID,
		Action:    sharedtypes.OpActionDelete,
		Payload:   map[string]any{},
		Timestamp: c.Now(),
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeStashDropped, sharedtypes.StashEventPayload{
		ID:       stashID,
		DeviceID: c.DeviceID,
	})
	return nil
}

func ListStashes(ctx context.Context, c *core.Context) ([]sharedtypes.Stash, error) {
	return c.Stash.List(ctx, c.UserID)
}

func GetStash(ctx context.Context, c *core.Context, id string) (*sharedtypes.Stash, error) {
	return c.Stash.Get(ctx, id, c.UserID)
}

func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
