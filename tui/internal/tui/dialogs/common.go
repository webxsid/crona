package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func plainIssueStatus(status string) string {
	switch status {
	case "in_progress":
		return "in progress"
	case "in_review":
		return "in review"
	default:
		return status
	}
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func dialogSubmitChord(controllerpkg.State) string {
	return "ctrl+s"
}

func dialogSubmitHint(state controllerpkg.State, label string) string {
	return "[" + dialogSubmitChord(state) + "] " + label
}
