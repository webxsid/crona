package commands

import (
	"context"
	"encoding/json"
	"errors"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

type ContextPatch struct {
	RepoSet   bool
	RepoID    *int64
	StreamSet bool
	StreamID  *int64
	IssueSet  bool
	IssueID   *int64
}

func GetActiveContext(ctx context.Context, c *core.Context) (*sharedtypes.ActiveContext, error) {
	return c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
}

func SwitchRepo(
	ctx context.Context,
	c *core.Context,
	repoID int64,
) (*sharedtypes.ActiveContext, error) {
	now := c.Now()
	updated, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID: &repoID,
	})
	if err != nil {
		return nil, err
	}
	if err := appendContextOp(ctx, c, map[string]any{"repoId": repoID}, sharedtypes.OpActionUpdate, now); err != nil {
		return nil, err
	}
	return updated, emitContextSnapshot(c, sharedtypes.EventTypeContextRepoChanged, updated)
}

func SwitchStream(
	ctx context.Context,
	c *core.Context,
	streamID int64,
) (*sharedtypes.ActiveContext, error) {
	existing, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.RepoID == nil {
		return nil, errors.New("no active repo. switch repo first")
	}
	now := c.Now()
	updated, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID:   existing.RepoID,
		StreamID: &streamID,
	})
	if err != nil {
		return nil, err
	}
	if err := appendContextOp(ctx, c, map[string]any{"streamId": streamID}, sharedtypes.OpActionUpdate, now); err != nil {
		return nil, err
	}
	return updated, emitContextSnapshot(c, sharedtypes.EventTypeContextStreamChanged, updated)
}

func SwitchIssue(
	ctx context.Context,
	c *core.Context,
	issueID int64,
) (*sharedtypes.ActiveContext, error) {
	existing, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.StreamID == nil {
		return nil, errors.New("no active stream. switch stream first")
	}
	now := c.Now()
	updated, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID:   existing.RepoID,
		StreamID: existing.StreamID,
		IssueID:  &issueID,
	})
	if err != nil {
		return nil, err
	}
	if err := appendContextOp(ctx, c, map[string]any{"issueId": issueID}, sharedtypes.OpActionUpdate, now); err != nil {
		return nil, err
	}
	return updated, emitContextSnapshot(c, sharedtypes.EventTypeContextIssueChanged, updated)
}

func ClearIssue(ctx context.Context, c *core.Context) (*sharedtypes.ActiveContext, error) {
	existing, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("no active context")
	}
	now := c.Now()
	updated, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID:   existing.RepoID,
		StreamID: existing.StreamID,
	})
	if err != nil {
		return nil, err
	}
	if err := appendContextOp(ctx, c, map[string]any{"issueId": nil}, sharedtypes.OpActionUpdate, now); err != nil {
		return nil, err
	}
	return updated, emitContextSnapshot(c, sharedtypes.EventTypeContextIssueChanged, updated)
}

func ClearContext(ctx context.Context, c *core.Context) error {
	now := c.Now()
	if err := c.ActiveContext.Clear(ctx, c.UserID, c.DeviceID); err != nil {
		return err
	}
	if err := appendContextOp(ctx, c, map[string]any{}, sharedtypes.OpActionDelete, now); err != nil {
		return err
	}
	payload, _ := json.Marshal(sharedtypes.ContextClearedPayload{DeviceID: c.DeviceID})
	c.Events.Emit(sharedtypes.KernelEvent{
		Type:    sharedtypes.EventTypeContextCleared,
		Payload: payload,
	})
	return nil
}

func SetContext(
	ctx context.Context,
	c *core.Context,
	patch ContextPatch,
) (*sharedtypes.ActiveContext, error) {
	existing, err := c.ActiveContext.Get(ctx, c.UserID, c.DeviceID)
	if err != nil {
		return nil, err
	}

	var repoID, streamID, issueID *int64
	if existing != nil {
		repoID = existing.RepoID
		streamID = existing.StreamID
		issueID = existing.IssueID
	}

	eventType := ""
	if patch.RepoSet {
		repoID = patch.RepoID
		streamID = nil
		issueID = nil
		eventType = sharedtypes.EventTypeContextRepoChanged
	}
	if patch.StreamSet {
		if patch.StreamID != nil && repoID == nil {
			return nil, errors.New("no active repo. switch repo first")
		}
		streamID = patch.StreamID
		issueID = nil
		if eventType == "" {
			eventType = sharedtypes.EventTypeContextStreamChanged
		}
	}
	if patch.IssueSet {
		if patch.IssueID != nil && streamID == nil {
			return nil, errors.New("no active stream. switch stream first")
		}
		issueID = patch.IssueID
		if eventType == "" {
			eventType = sharedtypes.EventTypeContextIssueChanged
		}
	}

	updated, err := c.ActiveContext.Set(ctx, c.UserID, c.DeviceID, struct {
		RepoID   *int64
		StreamID *int64
		IssueID  *int64
	}{
		RepoID:   repoID,
		StreamID: streamID,
		IssueID:  issueID,
	})
	if err != nil {
		return nil, err
	}

	now := c.Now()
	payload := map[string]any{}
	if patch.RepoSet {
		payload["repoId"] = patch.RepoID
	}
	if patch.StreamSet {
		payload["streamId"] = patch.StreamID
	}
	if patch.IssueSet {
		payload["issueId"] = patch.IssueID
	}
	if err := appendContextOp(ctx, c, payload, sharedtypes.OpActionUpdate, now); err != nil {
		return nil, err
	}
	if eventType != "" {
		return updated, emitContextSnapshot(c, eventType, updated)
	}
	return updated, nil
}

func appendContextOp(
	ctx context.Context,
	c *core.Context,
	payload any,
	action sharedtypes.OpAction,
	now string,
) error {
	return c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityActiveContext,
		EntityID:  c.UserID,
		Action:    action,
		Payload:   payload,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	})
}

func emitContextSnapshot(
	c *core.Context,
	eventType string,
	active *sharedtypes.ActiveContext,
) error {
	payload, err := json.Marshal(sharedtypes.ContextChangedPayload{
		DeviceID: c.DeviceID,
		RepoID:   active.RepoID,
		StreamID: active.StreamID,
		IssueID:  active.IssueID,
	})
	if err != nil {
		return err
	}
	c.Events.Emit(sharedtypes.KernelEvent{
		Type:    eventType,
		Payload: payload,
	})
	return nil
}
