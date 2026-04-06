package viewruntime

import (
	"strings"
	"time"

	"crona/tui/internal/api"
)

func ProtectedRestMode(settings *api.CoreSettings, date string) (bool, bool, string) {
	if settings == nil {
		return false, false, ""
	}
	if settings.AwayModeEnabled {
		return true, true, "Away mode is active."
	}
	day := strings.TrimSpace(date)
	if day == "" {
		day = time.Now().Format("2006-01-02")
	}
	if containsDate(settings.RestSpecificDates, day) {
		return true, false, "This is one of your configured rest days."
	}
	parsed, err := time.Parse("2006-01-02", day)
	if err != nil {
		return false, false, ""
	}
	if containsWeekday(settings.RestWeekdays, int(parsed.Weekday())) {
		return true, false, "This is one of your configured rest days."
	}
	return false, false, ""
}

func RestModeMessage(date string) string {
	messages := []string{"Enjoy your break", "Rest is part of the work", "Step away and reset", "Take the day lightly"}
	sum := 0
	for _, r := range date {
		sum += int(r)
	}
	return messages[sum%len(messages)]
}

func containsDate(values []string, value string) bool {
	for _, candidate := range values {
		if strings.TrimSpace(candidate) == value {
			return true
		}
	}
	return false
}

func containsWeekday(values []int, weekday int) bool {
	for _, value := range values {
		if value == weekday {
			return true
		}
	}
	return false
}
