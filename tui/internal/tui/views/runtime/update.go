package viewruntime

import (
	"strings"

	"crona/tui/internal/api"
)

func ShouldShowUpdatesView(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if !status.Enabled || !status.PromptEnabled || !status.UpdateAvailable {
		return false
	}
	return strings.TrimSpace(status.LatestVersion) != "" && strings.TrimSpace(status.LatestVersion) != strings.TrimSpace(status.DismissedVersion)
}

func WrapMultilineBlock(value string, width int) []string {
	if width < 8 {
		width = 8
	}
	rows := []string{}
	for _, raw := range strings.Split(value, "\n") {
		line := strings.TrimRight(raw, " ")
		if line == "" {
			rows = append(rows, "")
			continue
		}
		for len(line) > width {
			rows = append(rows, line[:width])
			line = line[width:]
		}
		rows = append(rows, line)
	}
	return rows
}
