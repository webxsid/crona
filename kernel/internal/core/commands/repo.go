package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
)

func CreateRepo(ctx context.Context, c *core.Context, input struct {
	Name        string
	Description *string
	Color       *string
}) (sharedtypes.Repo, error) {
	return createRepo(ctx, c, sharedtypes.Repo{Name: input.Name, Description: input.Description, Color: input.Color})
}

// createRepo creates repo with a fresh ID; repo.System is kept for Crona's own repos.
func createRepo(ctx context.Context, c *core.Context, repo sharedtypes.Repo) (sharedtypes.Repo, error) {
	if strings.TrimSpace(repo.Name) == "" {
		return sharedtypes.Repo{}, errors.New("repo name cannot be empty")
	}

	nextID, err := c.Repos.NextID(ctx)
	if err != nil {
		return sharedtypes.Repo{}, err
	}
	repo.ID = nextID
	repo.Name = strings.TrimSpace(repo.Name)
	repo.Description = normalizeOptionalString(repo.Description)
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
	Name        sharedtypes.Patch[string]
	Description sharedtypes.Patch[string]
	Color       sharedtypes.Patch[string]
}) (*sharedtypes.Repo, error) {
	if updates.Name.Set && updates.Name.Value != nil &&
		strings.TrimSpace(*updates.Name.Value) == "" {
		return nil, errors.New("repo name cannot be empty")
	}
	if updates.Name.Set && updates.Name.Value != nil {
		trimmed := strings.TrimSpace(*updates.Name.Value)
		updates.Name.Value = &trimmed
	}
	if updates.Description.Set {
		updates.Description.Value = normalizeOptionalString(updates.Description.Value)
	}
	if updates.Name.Set && updates.Name.Value != nil {
		if err := refuseSystemRepo(ctx, c, repoID, *updates.Name.Value); err != nil {
			return nil, err
		}
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
	if err := refuseSystemRepo(ctx, c, repoID, ""); err != nil {
		return err
	}
	now := c.Now()

	if err := c.Issues.CascadeSoftDeleteByRepo(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Habits.SoftDeleteByRepo(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Streams.SoftDeleteByRepo(ctx, repoID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Repos.SoftDelete(ctx, repoID, c.UserID, now); err != nil {
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

// refuseSystemRepo stops a system repo being renamed to newName, or deleted when newName
// is empty. Keeping its name is fine, so its description and color can still change.
func refuseSystemRepo(ctx context.Context, c *core.Context, repoID int64, newName string) error {
	repo, err := c.Repos.GetByID(ctx, repoID, c.UserID)
	if err != nil || repo == nil || !repo.System {
		return err
	}
	if newName == "" {
		return fmt.Errorf("%s is a system repo and cannot be deleted", repo.Name)
	}
	if newName != repo.Name {
		return fmt.Errorf("%s is a system repo and cannot be renamed", repo.Name)
	}
	return nil
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
