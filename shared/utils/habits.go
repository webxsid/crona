package utils

import (
	"time"

	"crona/shared/types"
)

func HabitMatchesDate(habit types.Habit, date string) bool {
	if !habit.Active {
		return false
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	switch types.NormalizeHabitScheduleType(habit.ScheduleType) {
	case types.HabitScheduleWeekdays:
		wd := int(parsed.Weekday())
		return wd >= 1 && wd <= 5
	case types.HabitScheduleWeekly:
		if len(habit.Weekdays) == 0 {
			return false
		}
		wd := int(parsed.Weekday())
		for _, day := range habit.Weekdays {
			if day == wd {
				return true
			}
		}
		return false
	default:
		return true
	}
}
