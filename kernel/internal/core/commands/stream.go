package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

func CreateStream(ctx context.Context, c *core.Context, input struct {
	RepoID      int64
	Name        string
	Description *string
	Visibility  *sharedtypes.StreamVisibility
}) (sharedtypes.Stream, error) {
	visibility := sharedtypes.StreamVisibilityPersonal
	if input.Visibility != nil {
		visibility = *input.Visibility
	}
	return createStream(ctx, c, sharedtypes.Stream{
		RepoID: input.RepoID, Name: input.Name, Description: input.Description, Visibility: visibility,
	})
}

// createStream creates stream with a fresh ID; stream.System is kept for Crona's own streams.
func createStream(ctx context.Context, c *core.Context, stream sharedtypes.Stream) (sharedtypes.Stream, error) {
	if strings.TrimSpace(stream.Name) == "" {
		return sharedtypes.Stream{}, errors.New("stream name cannot be empty")
	}
	nextID, err := c.Streams.NextID(ctx)
	if err != nil {
		return sharedtypes.Stream{}, err
	}
	stream.ID = nextID
	stream.Name = strings.TrimSpace(stream.Name)
	stream.Description = normalizeOptionalString(stream.Description)
	now := c.Now()
	created, err := c.Streams.Create(ctx, stream, c.UserID, now)
	if err != nil {
		return sharedtypes.Stream{}, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityStream,
		EntityID:  fmt.Sprintf("%d", created.ID),
		Action:    sharedtypes.OpActionCreate,
		Payload:   created,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return sharedtypes.Stream{}, err
	}
	emit(c, sharedtypes.EventTypeStreamCreated, created)
	return created, nil
}

func UpdateStream(ctx context.Context, c *core.Context, streamID int64, updates struct {
	Name        *string
	Description sharedtypes.Patch[string]
	Visibility  *sharedtypes.StreamVisibility
}) (*sharedtypes.Stream, error) {
	if updates.Name != nil && strings.TrimSpace(*updates.Name) == "" {
		return nil, errors.New("stream name cannot be empty")
	}
	if updates.Name != nil {
		trimmed := strings.TrimSpace(*updates.Name)
		updates.Name = &trimmed
	}
	if updates.Description.Set {
		updates.Description.Value = normalizeOptionalString(updates.Description.Value)
	}
	if updates.Name != nil {
		if err := refuseSystemStream(ctx, c, streamID, *updates.Name); err != nil {
			return nil, err
		}
	}
	now := c.Now()
	updated, err := c.Streams.Update(ctx, streamID, c.UserID, now, updates)
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityStream,
		EntityID:  fmt.Sprintf("%d", streamID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   updates,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		emit(c, sharedtypes.EventTypeStreamUpdated, updated)
	}
	return updated, nil
}

func DeleteStream(ctx context.Context, c *core.Context, streamID int64) error {
	if err := refuseSystemStream(ctx, c, streamID, ""); err != nil {
		return err
	}
	now := c.Now()
	if err := c.Issues.CascadeSoftDeleteByStream(ctx, streamID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Habits.SoftDeleteByStream(ctx, streamID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Streams.SoftDelete(ctx, streamID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityStream,
		EntityID:  fmt.Sprintf("%d", streamID),
		Action:    sharedtypes.OpActionDelete,
		Payload:   nil,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeStreamDeleted, sharedtypes.IDEventPayload{ID: streamID})
	return nil
}

func ListStreamsByRepo(
	ctx context.Context,
	c *core.Context,
	repoID int64,
) ([]sharedtypes.Stream, error) {
	streams, err := c.Streams.ListByRepo(ctx, repoID, c.UserID)
	if err != nil {
		return nil, err
	}
	sortStreams(streams, loadListSortSettings(ctx, c).streamSort)
	return streams, nil
}

// refuseSystemStream is refuseSystemRepo for streams.
func refuseSystemStream(ctx context.Context, c *core.Context, streamID int64, newName string) error {
	stream, err := c.Streams.GetByID(ctx, streamID, c.UserID)
	if err != nil || stream == nil || !stream.System {
		return err
	}
	if newName == "" {
		return fmt.Errorf("%s is a system stream and cannot be deleted", stream.Name)
	}
	if newName != stream.Name {
		return fmt.Errorf("%s is a system stream and cannot be renamed", stream.Name)
	}
	return nil
}
