package views

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
)

func renderUpdatesView(theme Theme, state ContentState) string {
	actions := []string{
		theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" check now"),
		theme.StyleHeader.Render("[o]") + theme.StyleDim.Render(" open release"),
		theme.StyleHeader.Render("[U]") + theme.StyleDim.Render(" dismiss"),
	}
	if state.UpdateInstallAvailable {
		actions = append(actions, theme.StyleHeader.Render("[i]")+theme.StyleDim.Render(" install"))
	} else {
		actions = append(actions, theme.StyleDim.Render("[i] install unavailable"))
	}
	lines := []string{
		theme.StylePaneTitle.Render("Updates"),
		renderActionLine(theme, state.Width-6, actions),
		"",
	}
	if state.UpdateStatus == nil {
		lines = append(lines, theme.StyleDim.Render("Loading update status..."))
		return renderPaneBox(theme, false, state.Width, state.Height, stringsJoin(lines))
	}

	status := state.UpdateStatus
	title := "No update available"
	if shouldShowUpdatesView(status) {
		title = "Update ready: v" + strings.TrimSpace(status.LatestVersion)
		if summary := firstUpdateSummary(status); summary != "" {
			title += "  " + summary
		}
	}
	lines = append(lines, theme.StyleHeader.Render(title))

	if strings.TrimSpace(status.ReleaseTag) != "" {
		lines = append(lines, theme.StyleDim.Render("Release tag: "+strings.TrimSpace(status.ReleaseTag)))
	}
	if strings.TrimSpace(status.PublishedAt) != "" {
		lines = append(lines, theme.StyleDim.Render("Published: "+strings.TrimSpace(status.PublishedAt)))
	}
	if strings.TrimSpace(status.CheckedAt) != "" {
		lines = append(lines, theme.StyleDim.Render("Last checked: "+strings.TrimSpace(status.CheckedAt)))
	}
	if strings.TrimSpace(status.ReleaseURL) != "" {
		lines = append(lines, theme.StyleDim.Render("Release page: "+strings.TrimSpace(status.ReleaseURL)))
	}
	if strings.TrimSpace(status.InstallScriptURL) != "" {
		lines = append(lines, theme.StyleDim.Render("Installer: "+strings.TrimSpace(status.InstallScriptURL)))
	}
	if strings.TrimSpace(status.ChecksumsURL) != "" {
		lines = append(lines, theme.StyleDim.Render("Checksums: "+strings.TrimSpace(status.ChecksumsURL)))
	}
	if !state.UpdateInstallAvailable {
		reason := strings.TrimSpace(state.UpdateManualReason)
		if reason == "" {
			reason = strings.TrimSpace(status.InstallUnavailableReason)
		}
		if reason == "" {
			reason = "Install is unavailable for this release."
		}
		lines = append(lines, "", theme.StyleDim.Render("Install status: "+reason))
		if strings.TrimSpace(state.TUIExecutablePath) != "" {
			lines = append(lines, theme.StyleDim.Render("TUI path: "+strings.TrimSpace(state.TUIExecutablePath)))
		}
		if strings.TrimSpace(state.KernelExecutablePath) != "" {
			lines = append(lines, theme.StyleDim.Render("Kernel path: "+strings.TrimSpace(state.KernelExecutablePath)))
		}
	}
	if state.UpdateChecking {
		lines = append(lines, "", lipStyle(theme, theme.ColorYellow).Render("Checking for updates..."))
	}
	if state.UpdateInstalling {
		lines = append(lines, "", lipStyle(theme, theme.ColorYellow).Render("Installing update and relaunching Crona..."))
	}
	if strings.TrimSpace(state.UpdateInstallError) != "" {
		lines = append(lines, "", theme.StyleError.Render("Install error: "+strings.TrimSpace(state.UpdateInstallError)))
	}
	if strings.TrimSpace(status.Error) != "" {
		lines = append(lines, "", theme.StyleError.Render("Check error: "+strings.TrimSpace(status.Error)))
	}
	if output := strings.TrimSpace(state.UpdateInstallOutput); output != "" {
		lines = append(lines, "", theme.StyleDim.Render("Installer output:"), truncateMultiline(output, state.Width-6, 8))
	}

	lines = append(lines, "", theme.StyleDim.Render("Release notes:"))
	notes := strings.TrimSpace(status.ReleaseNotes)
	if notes == "" {
		notes = "No release notes were published for this release."
	}
	lines = append(lines, wrapMultiline(notes, state.Width-6)...)
	return renderPaneBox(theme, false, state.Width, state.Height, stringsJoin(lines))
}

func shouldShowUpdatesView(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if !status.Enabled || !status.PromptEnabled || !status.UpdateAvailable {
		return false
	}
	return strings.TrimSpace(status.LatestVersion) != "" && strings.TrimSpace(status.LatestVersion) != strings.TrimSpace(status.DismissedVersion)
}

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

func wrapMultiline(value string, width int) []string {
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

func truncateMultiline(value string, width, maxLines int) string {
	rows := wrapMultiline(value, width)
	if len(rows) <= maxLines {
		return stringsJoin(rows)
	}
	rows = append(rows[:maxLines-1], fmt.Sprintf("... %d more lines", len(rows)-maxLines+1))
	return stringsJoin(rows)
}
