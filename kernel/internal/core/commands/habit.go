package commands

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"crona/kernel/internal/core"

	"github.com/google/uuid"

	sharedtypes "crona/shared/types"
	"crona/shared/utils"
)

func ListHabitsByStream(ctx context.Context, c *core.Context, streamID int64) ([]sharedtypes.Habit, error) {
	habits, err := c.Habits.ListByStream(ctx, streamID, c.UserID)
	if err != nil {
		return nil, err
	}
	sortHabits(habits, loadListSortSettings(ctx, c).habitSort)
	return habits, nil
}

func ListHabitsDueForDate(ctx context.Context, c *core.Context, date string) ([]sharedtypes.HabitDailyItem, error) {
	if !isISODate(date) {
		return nil, errors.New("date must be YYYY-MM-DD")
	}
	habits, err := c.Habits.ListDueWithMeta(ctx, date, c.UserID)
	if err != nil {
		return nil, err
	}
	completions, err := c.HabitCompletions.ListForDate(ctx, date, c.UserID)
	if err != nil {
		return nil, err
	}
	completionByHabit := make(map[int64]sharedtypes.HabitCompletion, len(completions))
	for _, completion := range completions {
		completionByHabit[completion.HabitID] = completion
	}
	addCompletion := func(item *sharedtypes.HabitDailyItem, completion sharedtypes.HabitCompletion) {
		item.Status = sharedtypes.NormalizeHabitCompletionStatus(completion.Status)
		item.Completed = item.Status == sharedtypes.HabitCompletionStatusCompleted
		item.CompletionID = &completion.ID
		item.CompletionDate = &completion.Date
		item.DurationMinutes = completion.DurationMinutes
		item.Notes = completion.Notes
		item.SnapshotName = completion.SnapshotName
		item.SnapshotDesc = completion.SnapshotDesc
		item.SnapshotType = completion.SnapshotType
		item.SnapshotDays = append([]int(nil), completion.SnapshotDays...)
		item.SnapshotTarget = completion.SnapshotTarget
		applyHabitCompletionSnapshot(item)
	}
	out := make([]sharedtypes.HabitDailyItem, 0, len(habits))
	seen := make(map[int64]struct{}, len(habits))
	for _, habit := range habits {
		item := sharedtypes.HabitDailyItem{HabitWithMeta: habit}
		if completion, ok := completionByHabit[habit.ID]; ok {
			addCompletion(&item, completion)
		}
		seen[habit.ID] = struct{}{}
		out = append(out, item)
	}
	for _, completion := range completions {
		if _, ok := seen[completion.HabitID]; ok {
			continue
		}
		habit, err := c.Habits.GetWithMetaByID(ctx, completion.HabitID, c.UserID)
		if err != nil {
			return nil, err
		}
		if habit == nil {
			continue
		}
		item := sharedtypes.HabitDailyItem{HabitWithMeta: *habit}
		addCompletion(&item, completion)
		out = append(out, item)
	}
	sortHabitDailyItems(out, loadListSortSettings(ctx, c).habitSort)
	return out, nil
}

func CreateHabit(ctx context.Context, c *core.Context, input struct {
	StreamID      int64
	Name          string
	Description   *string
	ScheduleType  string
	Weekdays      []int
	TargetMinutes *int
},
) (sharedtypes.Habit, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return sharedtypes.Habit{}, errors.New("habit name cannot be empty")
	}
	scheduleType, weekdays, err := normalizeHabitSchedule(input.ScheduleType, input.Weekdays)
	if err != nil {
		return sharedtypes.Habit{}, err
	}
	if input.TargetMinutes != nil && *input.TargetMinutes < 0 {
		return sharedtypes.Habit{}, errors.New("targetMinutes must be >= 0")
	}
	nextID, err := c.Habits.NextID(ctx)
	if err != nil {
		return sharedtypes.Habit{}, err
	}
	habit := sharedtypes.Habit{
		ID:            nextID,
		StreamID:      input.StreamID,
		Name:          name,
		Description:   normalizeOptionalString(input.Description),
		ScheduleType:  scheduleType,
		Weekdays:      weekdays,
		TargetMinutes: input.TargetMinutes,
		Active:        true,
	}
	now := c.Now()
	created, err := c.Habits.Create(ctx, habit, c.UserID, now)
	if err != nil {
		return sharedtypes.Habit{}, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityHabit,
		EntityID:  fmt.Sprintf("%d", created.ID),
		Action:    sharedtypes.OpActionCreate,
		Payload:   created,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return sharedtypes.Habit{}, err
	}
	emit(c, sharedtypes.EventTypeHabitCreated, created)
	return created, nil
}

func UpdateHabit(ctx context.Context, c *core.Context, habitID int64, updates struct {
	Name          sharedtypes.Patch[string]
	Description   sharedtypes.Patch[string]
	ScheduleType  *string
	Weekdays      []int
	WeekdaysSet   bool
	TargetMinutes sharedtypes.Patch[int]
	Active        *bool
},
) (*sharedtypes.Habit, error) {
	if updates.Name.Set && updates.Name.Value != nil {
		trimmed := strings.TrimSpace(*updates.Name.Value)
		if trimmed == "" {
			return nil, errors.New("habit name cannot be empty")
		}
		updates.Name.Value = &trimmed
	}
	if updates.Description.Set {
		updates.Description.Value = normalizeOptionalString(updates.Description.Value)
	}
	if updates.TargetMinutes.Set && updates.TargetMinutes.Value != nil && *updates.TargetMinutes.Value < 0 {
		return nil, errors.New("targetMinutes must be >= 0")
	}
	if updates.ScheduleType != nil || updates.WeekdaysSet {
		current, err := c.Habits.GetByID(ctx, habitID, c.UserID)
		if err != nil {
			return nil, err
		}
		if current == nil {
			return nil, errors.New("habit not found")
		}
		scheduleRaw := string(current.ScheduleType)
		if updates.ScheduleType != nil {
			scheduleRaw = *updates.ScheduleType
		}
		weekdays := current.Weekdays
		if updates.WeekdaysSet {
			weekdays = updates.Weekdays
		}
		normalizedType, normalizedDays, err := normalizeHabitSchedule(scheduleRaw, weekdays)
		if err != nil {
			return nil, err
		}
		value := string(normalizedType)
		updates.ScheduleType = &value
		updates.Weekdays = normalizedDays
		updates.WeekdaysSet = true
	}
	now := c.Now()
	updated, err := c.Habits.Update(ctx, habitID, c.UserID, now, struct {
		Name          sharedtypes.Patch[string]
		Description   sharedtypes.Patch[string]
		ScheduleType  sharedtypes.Patch[string]
		Weekdays      []int
		WeekdaysSet   bool
		TargetMinutes sharedtypes.Patch[int]
		Active        *bool
	}{
		Name:          updates.Name,
		Description:   updates.Description,
		ScheduleType:  sharedtypes.Patch[string]{Set: updates.ScheduleType != nil, Value: updates.ScheduleType},
		Weekdays:      updates.Weekdays,
		WeekdaysSet:   updates.WeekdaysSet,
		TargetMinutes: updates.TargetMinutes,
		Active:        updates.Active,
	})
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityHabit,
		EntityID:  fmt.Sprintf("%d", habitID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   updates,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	if updated != nil {
		emit(c, sharedtypes.EventTypeHabitUpdated, updated)
	}
	return updated, nil
}

func DeleteHabit(ctx context.Context, c *core.Context, habitID int64) error {
	now := c.Now()
	if err := c.Habits.SoftDelete(ctx, habitID, c.UserID, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityHabit,
		EntityID:  fmt.Sprintf("%d", habitID),
		Action:    sharedtypes.OpActionDelete,
		Payload:   nil,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeHabitDeleted, sharedtypes.IDEventPayload{ID: habitID})
	return nil
}

func CompleteHabit(ctx context.Context, c *core.Context, habitID int64, date string, status sharedtypes.HabitCompletionStatus, durationMinutes *int, notes *string) (*sharedtypes.HabitCompletion, error) {
	if !isISODate(date) {
		return nil, errors.New("date must be YYYY-MM-DD")
	}
	status = sharedtypes.NormalizeHabitCompletionStatus(status)
	if durationMinutes != nil && *durationMinutes < 0 {
		return nil, errors.New("durationMinutes must be >= 0")
	}
	habit, err := c.Habits.GetByID(ctx, habitID, c.UserID)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, errors.New("habit not found")
	}
	if !utils.HabitMatchesDate(*habit, date) {
		return nil, errors.New("habit is not due for this date")
	}
	nextID, err := c.HabitCompletions.NextID(ctx)
	if err != nil {
		return nil, err
	}
	now := c.Now()
	completion, err := c.HabitCompletions.Upsert(ctx, sharedtypes.HabitCompletion{
		ID:              nextID,
		HabitID:         habitID,
		Date:            date,
		Status:          status,
		DurationMinutes: durationMinutes,
		Notes:           normalizeOptionalString(notes),
		SnapshotName:    stringPtr(habit.Name),
		SnapshotDesc:    habit.Description,
		SnapshotType:    habitScheduleTypePtr(habit.ScheduleType),
		SnapshotDays:    append([]int(nil), habit.Weekdays...),
		SnapshotTarget:  habit.TargetMinutes,
	}, c.UserID, now)
	if err != nil {
		return nil, err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityHabitCompletion,
		EntityID:  fmt.Sprintf("%d", completion.ID),
		Action:    sharedtypes.OpActionUpdate,
		Payload:   completion,
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return nil, err
	}
	emit(c, sharedtypes.EventTypeHabitCompleted, completion)
	return &completion, nil
}

func UncompleteHabit(ctx context.Context, c *core.Context, habitID int64, date string) error {
	if !isISODate(date) {
		return errors.New("date must be YYYY-MM-DD")
	}
	now := c.Now()
	if err := c.HabitCompletions.DeleteByHabitAndDate(ctx, habitID, date, c.UserID, now); err != nil {
		return err
	}
	if err := c.Ops.Append(ctx, sharedtypes.Op{
		ID:        uuid.NewString(),
		Entity:    sharedtypes.OpEntityHabitCompletion,
		EntityID:  fmt.Sprintf("%d", habitID),
		Action:    sharedtypes.OpActionDelete,
		Payload:   map[string]any{"habitId": habitID, "date": date},
		Timestamp: now,
		UserID:    c.UserID,
		DeviceID:  c.DeviceID,
	}); err != nil {
		return err
	}
	emit(c, sharedtypes.EventTypeHabitUncompleted, map[string]any{"habitId": habitID, "date": date})
	return nil
}

func ListHabitHistory(ctx context.Context, c *core.Context, habitID int64) ([]sharedtypes.HabitCompletion, error) {
	return c.HabitCompletions.ListByHabit(ctx, habitID, c.UserID)
}

func normalizeHabitSchedule(raw string, weekdays []int) (sharedtypes.HabitScheduleType, []int, error) {
	value := sharedtypes.NormalizeHabitScheduleType(sharedtypes.HabitScheduleType(strings.TrimSpace(raw)))
	switch value {
	case sharedtypes.HabitScheduleWeekdays:
		return value, nil, nil
	case sharedtypes.HabitScheduleWeekly:
		if len(weekdays) == 0 {
			return "", nil, errors.New("weekly habits require at least one weekday")
		}
		out := make([]int, 0, len(weekdays))
		seen := map[int]struct{}{}
		for _, day := range weekdays {
			if day < 0 || day > 6 {
				return "", nil, errors.New("weekday values must be between 0 and 6")
			}
			if _, ok := seen[day]; ok {
				continue
			}
			seen[day] = struct{}{}
			out = append(out, day)
		}
		slices.Sort(out)
		return value, out, nil
	default:
		return sharedtypes.HabitScheduleDaily, nil, nil
	}
}

func isISODate(value string) bool {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	return err == nil
}

func applyHabitCompletionSnapshot(item *sharedtypes.HabitDailyItem) {
	if item == nil {
		return
	}
	if item.SnapshotName != nil && strings.TrimSpace(*item.SnapshotName) != "" {
		item.Name = strings.TrimSpace(*item.SnapshotName)
	}
	if item.SnapshotDesc != nil {
		item.Description = item.SnapshotDesc
	}
	if item.SnapshotType != nil {
		item.ScheduleType = sharedtypes.NormalizeHabitScheduleType(*item.SnapshotType)
		item.Weekdays = append([]int(nil), item.SnapshotDays...)
	}
	if item.SnapshotTarget != nil {
		item.TargetMinutes = item.SnapshotTarget
	}
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func habitScheduleTypePtr(value sharedtypes.HabitScheduleType) *sharedtypes.HabitScheduleType {
	normalized := sharedtypes.NormalizeHabitScheduleType(value)
	return &normalized
}
