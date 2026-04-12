package updates

import (
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
	types "crona/tui/internal/tui/views/types"
)

func renderView(theme types.Theme, state types.ContentState) string {
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
		viewchrome.RenderActionLine(theme, state.Width-6, actions),
		"",
	}
	if state.UpdateStatus == nil {
		lines = append(lines, theme.StyleDim.Render("Loading update status..."))
		return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}

	status := state.UpdateStatus
	title := "No update available"
	if viewruntime.ShouldShowUpdatesView(status) {
		title = "Update ready: v" + strings.TrimSpace(status.LatestVersion)
		if summary := firstUpdateSummary(status); summary != "" {
			title += "  " + summary
		}
	}
	lines = append(lines, theme.StyleHeader.Render(title))
	lines = append(lines, theme.StyleDim.Render("Running channel: "+settingsmeta.UpdateChannelLabel(status.RunningChannel)))
	lines = append(lines, theme.StyleDim.Render("Configured update channel: "+settingsmeta.UpdateChannelLabel(status.Channel)))
	if status.ReleaseIsPrerelease {
		lines = append(lines, theme.StyleDim.Render("Release type: beta prerelease"))
	}
	if status.LatestVersion != "" {
		latestKind := "stable release"
		if status.LatestIsBeta {
			latestKind = "beta release"
		}
		lines = append(lines, theme.StyleDim.Render("Latest release kind: "+latestKind))
	}

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
			lines = append(lines, theme.StyleDim.Render("Engine path: "+strings.TrimSpace(state.KernelExecutablePath)))
		}
	}
	if state.UpdateChecking {
		lines = append(lines, "", viewchrome.LipStyle(theme, theme.ColorYellow).Render("Checking for updates..."))
	}
	if state.UpdateInstalling {
		lines = append(lines, "", viewchrome.LipStyle(theme, theme.ColorYellow).Render("Installing update and relaunching Crona..."))
	}
	if strings.TrimSpace(state.UpdateInstallError) != "" {
		lines = append(lines, "", theme.StyleError.Render("Install error: "+strings.TrimSpace(state.UpdateInstallError)))
	}
	if strings.TrimSpace(status.Error) != "" {
		lines = append(lines, "", theme.StyleError.Render("Check error: "+strings.TrimSpace(status.Error)))
	}

	lines = append(lines, "", theme.StyleDim.Render("Release notes:"))
	notes := strings.TrimSpace(status.ReleaseNotes)
	if notes == "" {
		notes = "No release notes were published for this release."
	}
	lines = append(lines, viewruntime.WrapMultilineBlock(notes, state.Width-6)...)
	return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}
