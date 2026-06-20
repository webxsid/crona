package utils

import (
	"testing"

	"crona/shared/types"
)

func TestHabitMomentumCapacityDailyAllRequiresDailyHabits(t *testing.T) {
	capacity := HabitMomentumCapacityForSelection([]types.Habit{
		{ScheduleType: types.HabitScheduleDaily},
		{ScheduleType: types.HabitScheduleWeekdays},
	}, types.HabitStreakPeriodDay, types.MomentumMatchModeAll)
	if capacity.Valid {
		t.Fatalf("expected invalid capacity")
	}
	if capacity.Reason != "daily all-match momentum requires every selected habit to repeat daily" {
		t.Fatalf("unexpected reason %q", capacity.Reason)
	}
}

func TestHabitMomentumCapacityDailyAnyUsesLargestStack(t *testing.T) {
	capacity := HabitMomentumCapacityForSelection([]types.Habit{
		{ScheduleType: types.HabitScheduleDaily},
		{ScheduleType: types.HabitScheduleDaily},
		{ScheduleType: types.HabitScheduleWeekdays},
	}, types.HabitStreakPeriodDay, types.MomentumMatchModeAny)
	if !capacity.Valid || capacity.MaxCount != 3 {
		t.Fatalf("expected max 3, got %+v", capacity)
	}
}

func TestHabitMomentumCapacityWeeklyAllUsesOverlapCount(t *testing.T) {
	capacity := HabitMomentumCapacityForSelection([]types.Habit{
		{ScheduleType: types.HabitScheduleWeekly, Weekdays: []int{1, 3, 5}},
		{ScheduleType: types.HabitScheduleWeekly, Weekdays: []int{3, 5}},
	}, types.HabitStreakPeriodWeek, types.MomentumMatchModeAll)
	if !capacity.Valid || capacity.MaxCount != 2 {
		t.Fatalf("expected max 2, got %+v", capacity)
	}
}

func TestHabitMomentumCapacityWeeklyAnyUsesSummedDueCounts(t *testing.T) {
	capacity := HabitMomentumCapacityForSelection([]types.Habit{
		{ScheduleType: types.HabitScheduleWeekly, Weekdays: []int{1, 3, 5}},
		{ScheduleType: types.HabitScheduleWeekdays},
	}, types.HabitStreakPeriodWeek, types.MomentumMatchModeAny)
	if !capacity.Valid || capacity.MaxCount != 8 {
		t.Fatalf("expected max 8, got %+v", capacity)
	}
}

func TestHabitMomentumCapacityMonthlyUsesRealMonthWorstCase(t *testing.T) {
	capacity := HabitMomentumCapacityForSelection([]types.Habit{
		{ScheduleType: types.HabitScheduleWeekly, Weekdays: []int{1, 3, 5}},
	}, types.HabitStreakPeriodMonth, types.MomentumMatchModeAny)
	if !capacity.Valid || capacity.MaxCount != 14 {
		t.Fatalf("expected max 14, got %+v", capacity)
	}
}
