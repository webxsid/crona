package dialogs

import (
	shareddto "crona/shared/dto"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

func ParseEstimateInput(raw string) (*int, error) {
	return ParseOptionalDurationMinutes(raw, "Estimate")
}

func ParseDueDateInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, fmt.Errorf("due date must be YYYY-MM-DD")
	}
	return &raw, nil
}

func ValueToPointer(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func ValueOrEmpty(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func ParseNumericID(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func EndSessionRequest(inputs []textinput.Model) shareddto.EndSessionRequest {
	if len(inputs) == 0 {
		return shareddto.EndSessionRequest{}
	}
	req := shareddto.EndSessionRequest{
		CommitMessage: ValueToPointer(inputs[0].Value()),
	}
	if len(inputs) > 1 {
		req.WorkedOn = ValueToPointer(inputs[1].Value())
	}
	if len(inputs) > 2 {
		req.Outcome = ValueToPointer(inputs[2].Value())
	}
	if len(inputs) > 3 {
		req.NextStep = ValueToPointer(inputs[3].Value())
	}
	if len(inputs) > 4 {
		req.Blockers = ValueToPointer(inputs[4].Value())
	}
	if len(inputs) > 5 {
		req.Links = ValueToPointer(inputs[5].Value())
	}
	return req
}

func ParseDurationInput(raw string, required bool, label string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return 0, fmt.Errorf("%s is required", label)
		}
		return 0, nil
	}
	minutes, err := parseFlexibleDurationMinutes(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be like 90, 90m, 1h30m, or 1.5h", label)
	}
	seconds := minutes * 60
	if seconds < 0 {
		return 0, fmt.Errorf("%s must be non-negative", label)
	}
	if required && seconds == 0 {
		return 0, fmt.Errorf("%s must be positive", label)
	}
	return seconds, nil
}

func ParseClockInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("15:04", raw); err != nil {
		return nil, fmt.Errorf("time must be HH:MM")
	}
	return &raw, nil
}

func ParseOptionalDurationMinutes(raw string, label string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := parseFlexibleDurationMinutes(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be like 90, 90m, 1h30m, or 1.5h", strings.ToLower(label))
	}
	return &value, nil
}

func ParseOptionalDurationHours(raw string, label string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := parseFlexibleDurationHours(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be like 7.5h, 7h30m, or 450m", strings.ToLower(label))
	}
	return &value, nil
}

func FormatDurationMinutesInput(minutes *int) string {
	if minutes == nil {
		return ""
	}
	return formatFlexibleDurationMinutes(*minutes)
}

func FormatDurationHoursInput(hours *float64) string {
	if hours == nil {
		return ""
	}
	return formatFlexibleDurationHours(*hours)
}

func parseFlexibleDurationMinutes(raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, nil
	}
	if minutes, err := strconv.Atoi(raw); err == nil {
		if minutes < 0 {
			return 0, fmt.Errorf("duration must be non-negative")
		}
		return minutes, nil
	}
	if decimalHours, ok, err := parseBareHours(raw); err != nil {
		return 0, err
	} else if ok {
		return int(math.Round(decimalHours * 60)), nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	minutes := int(math.Round(d.Minutes()))
	if minutes < 0 {
		return 0, fmt.Errorf("duration must be non-negative")
	}
	return minutes, nil
}

func parseFlexibleDurationHours(raw string) (float64, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, nil
	}
	if hours, ok, err := parseBareHours(raw); err != nil {
		return 0, err
	} else if ok {
		if hours < 0 {
			return 0, fmt.Errorf("duration must be non-negative")
		}
		return hours, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	hours := d.Hours()
	if hours < 0 {
		return 0, fmt.Errorf("duration must be non-negative")
	}
	return hours, nil
}

func parseBareHours(raw string) (float64, bool, error) {
	if strings.ContainsAny(raw, "hm") {
		return 0, false, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false, nil
	}
	return value, true, nil
}

func formatFlexibleDurationMinutes(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	hours := minutes / 60
	remainder := minutes % 60
	switch {
	case hours > 0 && remainder > 0:
		return fmt.Sprintf("%dh%02dm", hours, remainder)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func formatFlexibleDurationHours(hours float64) string {
	if hours <= 0 {
		return "0h"
	}
	return formatFlexibleDurationMinutes(int(math.Round(hours * 60)))
}
