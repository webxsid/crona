package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"crona/kernel/internal/core"
	"crona/kernel/internal/store"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

func CreateRepo(ctx context.Context, c *core.Context, input struct {
	Name        string
	Description *string
	Color       *string
}) (sharedtypes.Repo, error) {
	if strings.TrimSpace(input.Name) == "" {
		return sharedtypes.Repo{}, errors.New("repo name cannot be empty")
	}

	nextID, err := c.Repos.NextID(ctx)
	if err != nil {
		return sharedtypes.Repo{}, err
	}
	repo := sharedtypes.Repo{
		ID:          nextID,
		Name:        strings.TrimSpace(input.Name),
		Description: normalizeOptionalString(input.Description),
		Color:       input.Color,
	}
	now := c.Now()

	created, err := c.Repos.Create(ctx, repo, c.UserID, now)
	if err != nil {
		return sharedtypes.Repo{}, err
	}

	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityRepo,
		EntityID:  fmt.Sprintf("%d", created.ID),
		Action:    sharedtypes.OpActionCreate,
		Payload:   created,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return sharedtypes.Repo{}, err
	}

	emit(c, sharedtypes.EventTypeRepoCreated, created)
	return created, nil
}

func UpdateRepo(ctx context.Context, c *core.Context, repoID int64, updates struct {
	Name        store.Patch[string]
	Description store.Patch[string]
	Color       store.Patch[string]
}) (*sharedtypes.Repo, error) {
	if updates.Name.Set && updates.Name.Value != nil && strings.TrimSpace(*updates.Name.Value) == "" {
		return nil, errors.New("repo name cannot be empty")
	}
	if updates.Name.Set && updates.Name.Value != nil {
		trimmed := strings.TrimSpace(*updates.Name.Value)
		updates.Name.Value = &trimmed
	}
	if updates.Description.Set {
		updates.Description.Value = normalizeOptionalString(updates.Description.Value)
	}

	now := c.Now()
	updated, err := c.Repos.Update(ctx, repoID, c.UserID, now, updates)
	if err != nil {
		return nil, err
	}

	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityRepo,
		EntityID:  fmt.Sprintf("%d", repoID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   updates,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}

	if updated != nil {
		emit(c, sharedtypes.EventTypeRepoUpdated, updated)
	}
	return updated, nil
}

func DeleteRepo(ctx context.Context, c *core.Context, repoID int64) error {
	now := c.Now()

	if err := c.Repos.SoftDelete(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Streams.SoftDeleteByRepo(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Issues.CascadeSoftDeleteByRepo(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityRepo,
		EntityID:  fmt.Sprintf("%d", repoID),
		Action:    sharedtypes.OpActionDelete,
		Payload:   nil,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}

	emit(c, sharedtypes.EventTypeRepoDeleted, sharedtypes.IDEventPayload{ID: repoID})
	return nil
}

func ListRepos(ctx context.Context, c *core.Context) ([]sharedtypes.Repo, error) {
	repos, err := c.Repos.List(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	sortRepos(repos, loadListSortSettings(ctx, c).repoSort)
	return repos, nil
}

func emit(c *core.Context, eventType string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	c.Events.Emit(sharedtypes.KernelEvent{
		Type:    eventType,
		Payload: body,
	})
}
