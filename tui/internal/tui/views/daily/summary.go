package daily

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
)

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
