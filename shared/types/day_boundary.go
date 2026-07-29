package types

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DayBoundarySchedule describes a local wall-clock boundary.
type DayBoundarySchedule struct {
	Enabled          bool           `json:"enabled"`
	DefaultTime      string         `json:"defaultTime"`
	WeekdayOverrides map[int]string `json:"weekdayOverrides,omitempty"`
}

var dayBoundaryTimePattern = regexp.MustCompile(`^(?:[01][0-9]|2[0-3]):[0-5][0-9]$`)

func DefaultStartOfDaySchedule() DayBoundarySchedule {
	return DayBoundarySchedule{Enabled: true, DefaultTime: "00:00"}
}

func DefaultEndOfDaySchedule() DayBoundarySchedule {
	return DayBoundarySchedule{Enabled: false, DefaultTime: "00:00"}
}

func NormalizeCoreSettingsDayBoundaries(settings *CoreSettings) {
	if settings == nil {
		return
	}
	if settings.StartOfDay.DefaultTime == "" {
		settings.StartOfDay = DefaultStartOfDaySchedule()
	}
	if settings.EndOfDay.DefaultTime == "" {
		settings.EndOfDay = DefaultEndOfDaySchedule()
	}
}

func (s DayBoundarySchedule) Normalized() DayBoundarySchedule {
	if strings.TrimSpace(s.DefaultTime) == "" {
		s.DefaultTime = "00:00"
	}
	if len(s.WeekdayOverrides) == 0 {
		s.WeekdayOverrides = nil
	}
	return s
}

func (s DayBoundarySchedule) Validate() error {
	s = s.Normalized()
	if !dayBoundaryTimePattern.MatchString(s.DefaultTime) {
		return fmt.Errorf("invalid defaultTime %q: expected HH:mm", s.DefaultTime)
	}
	if _, err := time.Parse("15:04", s.DefaultTime); err != nil {
		return fmt.Errorf("invalid defaultTime %q: %w", s.DefaultTime, err)
	}
	for weekday, value := range s.WeekdayOverrides {
		if weekday < 1 || weekday > 7 {
			return fmt.Errorf("invalid weekday override %d: expected 1-7", weekday)
		}
		if !dayBoundaryTimePattern.MatchString(value) {
			return fmt.Errorf("invalid weekday override %d time %q: expected HH:mm", weekday, value)
		}
	}
	return nil
}

func (s DayBoundarySchedule) TimeForWeekday(weekday time.Weekday) string {
	s = s.Normalized()
	isoWeekday := int(weekday)
	if isoWeekday == 0 {
		isoWeekday = 7
	}
	if value, ok := s.WeekdayOverrides[isoWeekday]; ok {
		return value
	}
	return s.DefaultTime
}

func (s DayBoundarySchedule) MinutesForWeekday(weekday time.Weekday) (int, error) {
	value := s.TimeForWeekday(weekday)
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, err
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

// ValidateJSONShape rejects values that would otherwise be silently coerced by
// map[string]any decoding in the settings RPC layer.
func ValidateDayBoundaryValue(value any) (DayBoundarySchedule, error) {
	var schedule DayBoundarySchedule
	switch typed := value.(type) {
	case DayBoundarySchedule:
		schedule = typed
	case *DayBoundarySchedule:
		if typed == nil {
			return schedule, fmt.Errorf("schedule must not be null")
		}
		schedule = *typed
	default:
		return schedule, fmt.Errorf("schedule must be an object")
	}
	if err := schedule.Validate(); err != nil {
		return schedule, err
	}
	return schedule.Normalized(), nil
}

func FormatDayBoundaryOverrides(overrides map[int]string) string {
	if len(overrides) == 0 {
		return ""
	}
	keys := make([]int, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, strconv.Itoa(key)+"="+overrides[key])
	}
	return strings.Join(parts, ",")
}
