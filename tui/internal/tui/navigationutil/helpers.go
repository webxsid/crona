package navigationutil

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

func NextView[T comparable](order []T, current T, dir int) T {
	for i, candidate := range order {
		if candidate == current {
			return order[(i+dir+len(order))%len(order)]
		}
	}
	return current
}

func NextPane[T comparable](panes []T, current T, dir int) T {
	if len(panes) == 0 {
		return current
	}
	for i, pane := range panes {
		if pane == current {
			return panes[(i+dir+len(panes))%len(panes)]
		}
	}
	return panes[0]
}

func OptionalText(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "-"
	}
	return strings.TrimSpace(*value)
}

func FormatHabitSchedule(scheduleType sharedtypes.HabitScheduleType, weekdays []int) string {
	switch scheduleType {
	case sharedtypes.HabitScheduleWeekdays:
		return "weekdays"
	case sharedtypes.HabitScheduleWeekly:
		return FormatHabitWeekdays(weekdays)
	default:
		return "daily"
	}
}

func FormatHabitWeekdays(weekdays []int) string {
	if len(weekdays) == 0 {
		return "-"
	}
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := make([]string, 0, len(weekdays))
	for _, day := range weekdays {
		if day >= 0 && day < len(names) {
			out = append(out, names[day])
		}
	}
	return strings.Join(out, ", ")
}

func FormatHabitTarget(target *int) string {
	if target == nil {
		return "-"
	}
	return fmt.Sprintf("%dm", *target)
}

func ShiftISODate(date string, days int) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now().AddDate(0, 0, days).Format("2006-01-02")
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}
