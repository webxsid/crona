package viewruntime

import (
	"slices"
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
	if slices.Contains(settings.RestSpecificDates, day) {
		return true, false, "This is one of your configured rest days."
	}
	parsed, err := time.Parse("2006-01-02", day)
	if err != nil {
		return false, false, ""
	}
	if slices.Contains(settings.RestWeekdays, int(parsed.Weekday())) {
		return true, false, "This is one of your configured rest days."
	}
	return false, false, ""
}

func RestModeMessage(date string) string {
	messages := []string{
		"Enjoy your break",
		"Rest is part of the work",
		"Step away and reset",
		"Take the day lightly",
	}
	sum := 0
	for _, r := range date {
		sum += int(r)
	}
	return messages[sum%len(messages)]
}
