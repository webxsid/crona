package daily

import (
	"fmt"
	"strings"

	shareddatefmt "crona/shared/datefmt"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
	viewcalendar "crona/tui/internal/tui/views/calendar"
	types "crona/tui/internal/tui/views/types"
)

func renderCalendarForMode(
	theme types.Theme,
	rawDate string,
	weekStart sharedtypes.WeekStart,
	maxLines int,
	mode viewcalendar.Mode,
	awayDates []string,
) []string {
	if mode == viewcalendar.ModeAuto {
		return nil
	}
	return viewcalendar.Render(theme, viewcalendar.Selection{
		AnchorDate:   rawDate,
		SelectedDate: rawDate,
		MaxLines:     maxLines,
		WeekStart:    weekStart,
		Mode:         mode,
		AwayDates:    awayDates,
	})
}

func displayPatternIncludesWeek(settings *api.CoreSettings) bool {
	pattern := shareddatefmt.DisplayPattern(settings)
	pattern = strings.ReplaceAll(pattern, "[Week]", "")
	pattern = strings.ReplaceAll(pattern, "[week]", "")
	return strings.Contains(pattern, "W")
}

func issueProgressText(workedSeconds, estimateMinutes int, compact bool) string {
	worked := helperpkg.FormatCompactDurationSeconds(workedSeconds)
	estimate := helperpkg.FormatCompactDurationMinutes(estimateMinutes)
	if compact {
		if estimateMinutes > 0 {
			return fmt.Sprintf("%s/%s", worked, estimate)
		}
		return worked
	}
	if estimateMinutes > 0 {
		return fmt.Sprintf("worked %s / est. %s", worked, estimate)
	}
	return "worked " + worked
}
