package dialogs

import (
	"strings"
)

func dayBoundaryDaysLabel(days []int) string {
	labels := make([]string, 0, len(days))
	for _, day := range days {
		if day >= 1 && day <= 7 {
			labels = append(labels, []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[day-1])
		}
	}
	if len(labels) == 0 {
		return "None"
	}
	return strings.Join(labels, ", ")
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func fallbackTimezone(value string) string {
	if value == "" {
		return "daemon local time"
	}
	return value
}
