package updates

import (
	"strings"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

func firstUpdateSummary(status *api.UpdateStatus) string {
	if status == nil {
		return ""
	}
	if title := strings.TrimSpace(status.ReleaseName); title != "" {
		return title
	}
	for _, line := range strings.Split(status.ReleaseNotes, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}

func installActionLabel(status *api.UpdateStatus, installAvailable bool) string {
	if status != nil && strings.TrimSpace(status.UpdateCommand) != "" && !installAvailable {
		if viewchrome.IsMigrationCommand(status.UpdateCommand) {
			return "copy migration command"
		}
		return "copy update command"
	}
	if installAvailable {
		return "install"
	}
	return "install unavailable"
}
