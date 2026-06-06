package commands

import (
	"context"
	"errors"
	"strings"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
)

func ListHabitStreakDefinitions(
	ctx context.Context,
	c *core.Context,
) ([]sharedtypes.HabitStreakDefinition, error) {
	return c.HabitStreakDefinitions.List(ctx, c.UserID)
}

func CreateHabitStreakDefinition(
	ctx context.Context,
	c *core.Context,
	def sharedtypes.HabitStreakDefinition,
) (sharedtypes.HabitStreakDefinition, error) {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	if strings.TrimSpace(def.Name) == "" {
		return sharedtypes.HabitStreakDefinition{}, errors.New("habit streak name cannot be empty")
	}
	created, err := c.HabitStreakDefinitions.Create(ctx, c.UserID, c.Now(), def)
	if err != nil {
		return sharedtypes.HabitStreakDefinition{}, err
	}
	if err := InvalidateCustomHabitMomentumSnapshotsFrom(ctx, c, extractISODate(c.Now())); err != nil {
		return sharedtypes.HabitStreakDefinition{}, err
	}
	return created, nil
}

func UpdateHabitStreakDefinition(
	ctx context.Context,
	c *core.Context,
	def sharedtypes.HabitStreakDefinition,
) (*sharedtypes.HabitStreakDefinition, error) {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	if strings.TrimSpace(def.ID) == "" {
		return nil, errors.New("habit streak id is required")
	}
	if strings.TrimSpace(def.Name) == "" {
		return nil, errors.New("habit streak name cannot be empty")
	}
	updated, err := c.HabitStreakDefinitions.Update(ctx, c.UserID, c.Now(), def)
	if err != nil {
		return nil, err
	}
	if err := InvalidateCustomHabitMomentumSnapshotsFrom(ctx, c, extractISODate(c.Now())); err != nil {
		return nil, err
	}
	return updated, nil
}

func DeleteHabitStreakDefinition(ctx context.Context, c *core.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("habit streak id is required")
	}
	if err := c.HabitStreakDefinitions.Delete(ctx, c.UserID, id, c.Now()); err != nil {
		return err
	}
	return InvalidateCustomHabitMomentumSnapshotsFrom(ctx, c, extractISODate(c.Now()))
}
