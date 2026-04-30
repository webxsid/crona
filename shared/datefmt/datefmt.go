package datefmt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
)

const isoLayout = "2006-01-02"

func FormatISODate(raw string, settings *sharedtypes.CoreSettings) string {
	parsed, err := time.Parse(isoLayout, strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return FormatDate(parsed, settings)
}

func FormatRFC3339Date(raw string, settings *sharedtypes.CoreSettings) string {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		if len(raw) >= len(isoLayout) {
			return FormatISODate(raw[:10], settings)
		}
		return strings.TrimSpace(raw)
	}
	return FormatDate(parsed, settings)
}

func FormatDate(value time.Time, settings *sharedtypes.CoreSettings) string {
	pattern := effectivePattern(settings)
	rendered, ok := formatMomentDate(value, pattern)
	if ok {
		return rendered
	}
	rendered, ok = formatMomentDate(value, presetPattern(sharedtypes.DateDisplayPresetISO))
	if ok {
		return rendered
	}
	return value.Format(isoLayout)
}

func Preview(settings *sharedtypes.CoreSettings, now time.Time) string {
	return FormatDate(now, settings)
}

func DisplayPattern(settings *sharedtypes.CoreSettings) string {
	return effectivePattern(settings)
}

func PresetPattern(preset sharedtypes.DateDisplayPreset) string {
	return presetPattern(preset)
}

func effectivePattern(settings *sharedtypes.CoreSettings) string {
	preset := sharedtypes.DateDisplayPresetISO
	custom := ""
	if settings != nil {
		preset = sharedtypes.NormalizeDateDisplayPreset(settings.DateDisplayPreset)
		custom = strings.TrimSpace(settings.DateDisplayFormat)
	}
	if preset == sharedtypes.DateDisplayPresetCustom && custom != "" {
		if _, ok := formatMomentDate(time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC), custom); ok {
			return custom
		}
	}
	return presetPattern(preset)
}

func presetPattern(preset sharedtypes.DateDisplayPreset) string {
	switch sharedtypes.NormalizeDateDisplayPreset(preset) {
	case sharedtypes.DateDisplayPresetUS:
		return "MM/DD/YYYY"
	case sharedtypes.DateDisplayPresetEurope:
		return "DD/MM/YYYY"
	case sharedtypes.DateDisplayPresetLong:
		return "D MMM YYYY"
	case sharedtypes.DateDisplayPresetCustom:
		return "YYYY-MM-DD"
	default:
		return "YYYY-MM-DD"
	}
}

func formatMomentDate(value time.Time, pattern string) (string, bool) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return "", false
	}
	var out strings.Builder
	for i := 0; i < len(trimmed); {
		if trimmed[i] == '[' {
			end := strings.IndexByte(trimmed[i+1:], ']')
			if end < 0 {
				return "", false
			}
			out.WriteString(trimmed[i+1 : i+1+end])
			i += end + 2
			continue
		}
		token, width := longestToken(trimmed[i:])
		if token == "" {
			out.WriteByte(trimmed[i])
			i++
			continue
		}
		out.WriteString(renderToken(value, token))
		i += width
	}
	return out.String(), true
}

func longestToken(input string) (string, int) {
	tokens := []string{
		"YYYY", "dddd", "MMMM", "ddd", "MMM", "Do", "YY", "MM", "DD", "WW", "LL", "L", "M", "D", "W", "d",
	}
	for _, token := range tokens {
		if strings.HasPrefix(input, token) {
			return token, len(token)
		}
	}
	return "", 0
}

func renderToken(value time.Time, token string) string {
	switch token {
	case "YYYY":
		return fmt.Sprintf("%04d", value.Year())
	case "YY":
		return fmt.Sprintf("%02d", value.Year()%100)
	case "MMMM":
		return value.Month().String()
	case "MMM":
		name := value.Month().String()
		if len(name) > 3 {
			return name[:3]
		}
		return name
	case "MM":
		return fmt.Sprintf("%02d", int(value.Month()))
	case "M":
		return strconv.Itoa(int(value.Month()))
	case "DD":
		return fmt.Sprintf("%02d", value.Day())
	case "D":
		return strconv.Itoa(value.Day())
	case "Do":
		return ordinal(value.Day())
	case "dddd":
		return value.Weekday().String()
	case "ddd":
		name := value.Weekday().String()
		if len(name) > 3 {
			return name[:3]
		}
		return name
	case "d":
		return strconv.Itoa(int(value.Weekday()))
	case "WW":
		week := isoWeek(value)
		return fmt.Sprintf("%02d", week)
	case "W":
		return strconv.Itoa(isoWeek(value))
	case "L":
		return renderToken(value, "MM") + "/" + renderToken(value, "DD") + "/" + renderToken(value, "YYYY")
	case "LL":
		return renderToken(value, "D") + " " + renderToken(value, "MMM") + " " + renderToken(value, "YYYY")
	default:
		return token
	}
}

func ordinal(day int) string {
	if day%100 >= 11 && day%100 <= 13 {
		return strconv.Itoa(day) + "th"
	}
	switch day % 10 {
	case 1:
		return strconv.Itoa(day) + "st"
	case 2:
		return strconv.Itoa(day) + "nd"
	case 3:
		return strconv.Itoa(day) + "rd"
	default:
		return strconv.Itoa(day) + "th"
	}
}

func isoWeek(value time.Time) int {
	_, week := value.ISOWeek()
	return week
}
