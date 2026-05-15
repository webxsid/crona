package controller

import (
	"errors"
	"slices"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

func ParseHabitSchedule(raw string) (string, []int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "", "daily":
		return "daily", nil, nil
	case "weekdays":
		return "weekdays", nil, nil
	}
	parts := strings.Split(value, ",")
	weekdays := make([]int, 0, len(parts))
	for _, part := range parts {
		weekday, ok := parseWeekdayToken(strings.TrimSpace(part))
		if !ok {
			return "", nil, errors.New("schedule must be daily, weekdays, or comma-separated weekdays like mon,wed,fri")
		}
		weekdays = append(weekdays, weekday)
	}
	if len(weekdays) == 0 {
		return "", nil, errors.New("schedule must be daily, weekdays, or comma-separated weekdays like mon,wed,fri")
	}
	return "weekly", weekdays, nil
}

func ParseStreakKinds(raw string) ([]sharedtypes.StreakKind, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]sharedtypes.StreakKind, 0, len(parts))
	seen := map[sharedtypes.StreakKind]struct{}{}
	for _, part := range parts {
		var kind sharedtypes.StreakKind
		switch strings.TrimSpace(part) {
		case "focus", "focus_days":
			kind = sharedtypes.StreakKindFocusDays
		case "checkin", "checkins", "check-in", "check_in_days", "checkin_days":
			kind = sharedtypes.StreakKindCheckInDays
		default:
			return nil, errors.New("streaks must be comma-separated values from focus_days,checkin_days")
		}
		if _, ok := seen[kind]; ok {
			continue
		}
		seen[kind] = struct{}{}
		out = append(out, kind)
	}
	return out, nil
}

func ParseWeekdayList(raw string) ([]int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		weekday, ok := parseWeekdayToken(strings.TrimSpace(part))
		if !ok {
			return nil, errors.New("weekdays must be comma-separated tokens like mon,wed,fri")
		}
		if _, ok := seen[weekday]; ok {
			continue
		}
		seen[weekday] = struct{}{}
		out = append(out, weekday)
	}
	return out, nil
}

func parseWeekdayToken(value string) (int, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "sun", "sunday":
		return 0, true
	case "mon", "monday":
		return 1, true
	case "tue", "tues", "tuesday":
		return 2, true
	case "wed", "weds", "wednesday":
		return 3, true
	case "thu", "thur", "thurs", "thursday":
		return 4, true
	case "fri", "friday":
		return 5, true
	case "sat", "saturday":
		return 6, true
	default:
		return 0, false
	}
}

func ParseSpecificDates(raw string) ([]string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		date := strings.TrimSpace(part)
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("rest dates must use YYYY-MM-DD")
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		out = append(out, date)
	}
	return out, nil
}

func ParseRecurringDates(raw string) ([]string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		date := strings.TrimSpace(part)
		if _, err := time.Parse("01-02", date); err != nil {
			return nil, errors.New("recurring rest dates must use MM-DD")
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		out = append(out, date)
	}
	return out, nil
}

func streakKindStrings(values []sharedtypes.StreakKind) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func WeekdayTokens(days []int) []string {
	names := map[int]string{0: "sun", 1: "mon", 2: "tue", 3: "wed", 4: "thu", 5: "fri", 6: "sat"}
	out := make([]string, 0, len(days))
	for _, day := range days {
		if name, ok := names[day]; ok {
			out = append(out, name)
		}
	}
	return out
}

func normalizedWeekdays(values []int) []int {
	items := make([]int, 0, len(values))
	seen := map[int]struct{}{}
	for _, value := range values {
		if value < 0 || value > 6 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	slices.Sort(items)
	return items
}

func normalizedDateList(values []string) []string {
	items := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	slices.Sort(items)
	return items
}

func toggleWeekday(values []int, target int) []int {
	out := make([]int, 0, len(values))
	found := false
	for _, value := range values {
		if value == target {
			found = true
			continue
		}
		out = append(out, value)
	}
	if !found {
		out = append(out, target)
	}
	return normalizedWeekdays(out)
}

func toggleStreakKind(values []sharedtypes.StreakKind, target sharedtypes.StreakKind) []sharedtypes.StreakKind {
	out := make([]sharedtypes.StreakKind, 0, len(values))
	found := false
	for _, value := range values {
		if value == target {
			found = true
			continue
		}
		out = append(out, value)
	}
	if !found {
		out = append(out, target)
	}
	slices.Sort(out)
	return out
}
