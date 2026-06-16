package updates

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	versionpkg "crona/shared/version"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewtypes "crona/tui/internal/tui/views/types"
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
	if versionpkg.InstallScriptDeprecationEnabled() || (status != nil && status.InstallScriptDeprecated) {
		return "open migration guide"
	}
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

func updateOpenActionLabel(status *api.UpdateStatus) string {
	if versionpkg.InstallScriptDeprecationEnabled() || (status != nil && status.InstallScriptDeprecated) {
		return "open migration guide"
	}
	return "open release page"
}

func installScriptDeprecationBanner(theme viewtypes.Theme, status *api.UpdateStatus) []string {
	if !versionpkg.InstallScriptDeprecationEnabled() && (status == nil || !status.InstallScriptDeprecated) {
		return nil
	}
	lines := []string{
		viewchrome.LipStyle(theme, theme.ColorYellow).Render("Important"),
		"Install script deprecation",
		versionpkg.InstallScriptDeprecationMessage(),
	}
	url := strings.TrimSpace(status.MigrationGuideURL)
	if url == "" {
		url = versionpkg.InstallScriptMigrationURL
	}
	if url = strings.TrimSpace(url); url != "" {
		lines = append(lines, "Migration guide: "+url)
	}
	return lines
}

func displayVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "v") {
		return value
	}
	return "v" + value
}

func separatorLine(width int) string {
	if width < 8 {
		width = 8
	}
	return strings.Repeat("─", width)
}

func statusHeadline(status *api.UpdateStatus) string {
	if status != nil && status.UpdateAvailable {
		return "↑ Update Available"
	}
	return "✓ You're up to date"
}

func diagnosticsHeadline(expanded bool) string {
	if expanded {
		return "Diagnostics"
	}
	return "Diagnostics (hidden)"
}

func diagnosticsActionLabel(expanded bool) string {
	if expanded {
		return "hide diagnostics"
	}
	return "show diagnostics"
}

func releaseNotesSectionTitle(status *api.UpdateStatus) string {
	if status != nil && status.UpdateAvailable {
		return "What's New"
	}
	return "Release Notes"
}

func releaseNotesExcerpt(notes string, limit int) ([]string, bool) {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return nil, false
	}
	var lines []string
	rawLines := strings.Split(notes, "\n")
	sawMore := false
	for _, raw := range rawLines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		cleaned, ok := normalizeReleaseNoteLine(line)
		if !ok {
			continue
		}
		if len(lines) >= limit {
			sawMore = true
			continue
		}
		lines = append(lines, cleaned)
	}
	if len(lines) == 0 {
		for _, raw := range rawLines {
			line := strings.TrimSpace(raw)
			if line == "" {
				continue
			}
			if len(lines) >= limit {
				sawMore = true
				continue
			}
			lines = append(lines, line)
		}
	}
	if len(lines) > 0 && len(lines) < countNonEmptyReleaseLines(rawLines) {
		sawMore = true
	}
	return lines, sawMore
}

func normalizeReleaseNoteLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", false
	}
	for strings.HasPrefix(trimmed, "#") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
	}
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "• ") {
		return strings.TrimSpace(trimmed[2:]), true
	}
	if idx := strings.IndexFunc(trimmed, func(r rune) bool { return !unicode.IsDigit(r) }); idx > 0 {
		if strings.HasPrefix(trimmed[idx:], ". ") || strings.HasPrefix(trimmed[idx:], ") ") {
			return strings.TrimSpace(trimmed[idx+2:]), true
		}
	}
	return trimmed, true
}

func countNonEmptyReleaseLines(lines []string) int {
	total := 0
	for _, raw := range lines {
		if strings.TrimSpace(raw) != "" {
			total++
		}
	}
	return total
}

func relativeCheckedAt(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	d := time.Since(ts)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
