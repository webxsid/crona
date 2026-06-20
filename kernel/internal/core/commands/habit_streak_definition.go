package commands

import (
	"context"
	"errors"
	"strings"

	"crona/kernel/internal/core"
	sharedtypes "crona/shared/types"
	sharedutils "crona/shared/utils"
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
	if err := validateHabitStreakTarget(ctx, c, def); err != nil {
		return sharedtypes.HabitStreakDefinition{}, err
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
	if err := validateHabitStreakTarget(ctx, c, def); err != nil {
		return nil, err
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

func validateHabitStreakTarget(
	ctx context.Context,
	c *core.Context,
	def sharedtypes.HabitStreakDefinition,
) error {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		if len(def.Contexts) == 0 {
			return errors.New("context target requires at least one context")
		}
	default:
		if len(def.HabitIDs) == 0 {
			return errors.New("habit target requires at least one habit")
		}
		allHabits, err := c.Habits.ListAllWithMeta(ctx, c.UserID)
		if err != nil {
			return err
		}
		selected := make([]sharedtypes.Habit, 0, len(def.HabitIDs))
		byID := make(map[int64]sharedtypes.Habit, len(allHabits))
		for _, habit := range allHabits {
			byID[habit.ID] = habit.Habit
		}
		for _, habitID := range def.HabitIDs {
			habit, ok := byID[habitID]
			if !ok {
				return errors.New("habit target contains unknown habits")
			}
			selected = append(selected, habit)
		}
		if err := sharedutils.HabitMomentumCapacityError(
			selected,
			def.Period,
			def.MatchMode,
			def.RequiredCount,
		); err != nil {
			return err
		}
	}
	return nil
}
