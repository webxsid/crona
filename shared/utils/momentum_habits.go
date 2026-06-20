package utils

import (
	"fmt"
	"slices"

	"crona/shared/types"
)

type HabitMomentumCapacity struct {
	MaxCount int
	Valid    bool
	Reason   string
}

func HabitMomentumCapacityForSelection(
	habits []types.Habit,
	period types.HabitStreakPeriod,
	matchMode types.MomentumMatchMode,
) HabitMomentumCapacity {
	if len(habits) == 0 {
		return HabitMomentumCapacity{Reason: "habit target requires at least one habit"}
	}
	period = types.NormalizeHabitStreakPeriod(period)
	matchMode = types.NormalizeMomentumMatchMode(matchMode)
	weeksets := make([][]int, 0, len(habits))
	for _, habit := range habits {
		weeksets = append(weeksets, habitDueWeekdays(habit))
	}
	switch matchMode {
	case types.MomentumMatchModeAll:
		return habitMomentumAllCapacity(habits, weeksets, period)
	default:
		return habitMomentumAnyCapacity(weeksets, period)
	}
}

func habitMomentumAnyCapacity(
	weeksets [][]int,
	period types.HabitStreakPeriod,
) HabitMomentumCapacity {
	maxCount := 0
	switch period {
	case types.HabitStreakPeriodWeek:
		for _, weekdays := range weeksets {
			maxCount += len(weekdays)
		}
	case types.HabitStreakPeriodMonth:
		for _, weekdays := range weeksets {
			maxCount += maxMonthOccurrences(weekdays)
		}
	default:
		stack := map[int]int{}
		for _, weekdays := range weeksets {
			for _, day := range weekdays {
				stack[day]++
			}
		}
		for _, count := range stack {
			if count > maxCount {
				maxCount = count
			}
		}
	}
	if maxCount <= 0 {
		return HabitMomentumCapacity{Reason: "selected habits do not repeat in the chosen period"}
	}
	return HabitMomentumCapacity{MaxCount: maxCount, Valid: true}
}

func habitMomentumAllCapacity(
	habits []types.Habit,
	weeksets [][]int,
	period types.HabitStreakPeriod,
) HabitMomentumCapacity {
	if period == types.HabitStreakPeriodDay {
		for _, habit := range habits {
			if types.NormalizeHabitScheduleType(habit.ScheduleType) != types.HabitScheduleDaily {
				return HabitMomentumCapacity{
					Reason: "daily all-match momentum requires every selected habit to repeat daily",
				}
			}
		}
		return HabitMomentumCapacity{MaxCount: 1, Valid: true}
	}
	overlap := intersectWeekdays(weeksets)
	if len(overlap) == 0 {
		reason := "selected habits do not overlap for weekly all matching"
		if period == types.HabitStreakPeriodMonth {
			reason = "selected habits do not overlap for monthly all matching"
		}
		return HabitMomentumCapacity{Reason: reason}
	}
	maxCount := len(overlap)
	if period == types.HabitStreakPeriodMonth {
		maxCount = maxMonthOccurrences(overlap)
	}
	if maxCount <= 0 {
		return HabitMomentumCapacity{Reason: "selected habits do not overlap in the chosen period"}
	}
	return HabitMomentumCapacity{MaxCount: maxCount, Valid: true}
}

func habitDueWeekdays(habit types.Habit) []int {
	switch types.NormalizeHabitScheduleType(habit.ScheduleType) {
	case types.HabitScheduleWeekdays:
		return []int{1, 2, 3, 4, 5}
	case types.HabitScheduleWeekly:
		out := make([]int, 0, len(habit.Weekdays))
		seen := map[int]struct{}{}
		for _, day := range habit.Weekdays {
			if day < 0 || day > 6 {
				continue
			}
			if _, ok := seen[day]; ok {
				continue
			}
			seen[day] = struct{}{}
			out = append(out, day)
		}
		slices.Sort(out)
		return out
	default:
		return []int{0, 1, 2, 3, 4, 5, 6}
	}
}

func intersectWeekdays(weeksets [][]int) []int {
	if len(weeksets) == 0 {
		return nil
	}
	counts := map[int]int{}
	for _, set := range weeksets {
		local := map[int]struct{}{}
		for _, day := range set {
			if day < 0 || day > 6 {
				continue
			}
			local[day] = struct{}{}
		}
		for day := range local {
			counts[day]++
		}
	}
	out := make([]int, 0, 7)
	for day, count := range counts {
		if count == len(weeksets) {
			out = append(out, day)
		}
	}
	slices.Sort(out)
	return out
}

func maxMonthOccurrences(weekdays []int) int {
	best := 0
	for monthLen := 28; monthLen <= 31; monthLen++ {
		for startWeekday := 0; startWeekday <= 6; startWeekday++ {
			count := monthOccurrences(weekdays, monthLen, startWeekday)
			if count > best {
				best = count
			}
		}
	}
	return best
}

func monthOccurrences(weekdays []int, monthLen int, startWeekday int) int {
	if monthLen <= 0 || len(weekdays) == 0 {
		return 0
	}
	target := map[int]struct{}{}
	for _, day := range weekdays {
		target[day] = struct{}{}
	}
	count := 0
	for dayOfMonth := 0; dayOfMonth < monthLen; dayOfMonth++ {
		weekday := (startWeekday + dayOfMonth) % 7
		if _, ok := target[weekday]; ok {
			count++
		}
	}
	return count
}

func HabitMomentumCapacityError(
	habits []types.Habit,
	period types.HabitStreakPeriod,
	matchMode types.MomentumMatchMode,
	requiredCount int,
) error {
	capacity := HabitMomentumCapacityForSelection(habits, period, matchMode)
	if !capacity.Valid {
		return fmt.Errorf("%s", capacity.Reason)
	}
	if requiredCount < 1 {
		return fmt.Errorf("required completions must be positive")
	}
	if requiredCount > capacity.MaxCount {
		return fmt.Errorf(
			"required completions exceed the maximum possible for the selected habits (%d)",
			capacity.MaxCount,
		)
	}
	return nil
}
