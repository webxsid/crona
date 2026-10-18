package daily

import (
	"fmt"
	"strings"

	shareddatefmt "crona/shared/datefmt"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
)

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
