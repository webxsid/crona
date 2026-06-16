package updates

import (
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"
)

func renderView(theme types.Theme, state types.ContentState) string {
	lines := []string{theme.StylePaneTitle.Render("Updates")}
	actions := updateActions(theme, state)
	if len(actions) > 0 {
		lines = append(lines, viewchrome.RenderActionLine(theme, state.Width-6, actions))
		lines = append(lines, "")
	}
	if state.UpdateStatus == nil {
		lines = append(lines, theme.StyleDim.Render("Loading update status..."))
		return viewchrome.RenderPaneBox(
			theme,
			false,
			state.Width,
			state.Height,
			viewhelpers.StringsJoin(lines),
		)
	}

	status := state.UpdateStatus
	if banner := installScriptDeprecationBanner(theme, status); len(banner) > 0 {
		lines = append(lines, banner...)
		lines = append(lines, "")
	}
	if shouldShowMigrationChecklist(status) {
		lines = append(lines, migrationChecklistLines(theme, state)...)
		lines = append(lines, "")
	}
	lines = append(lines, statusCardLines(theme, state)...)
	lines = append(lines, "")
	lines = append(lines, theme.StyleDim.Render(separatorLine(state.Width-6)))
	lines = append(lines, theme.StyleHeader.Render(releaseNotesSectionTitle(status)))
	lines = append(lines, releaseNotesExcerptLines(theme, status.ReleaseNotes, state.Width-6)...)
	lines = append(lines, "")
	lines = append(lines, diagnosticsSectionLines(theme, state)...)
	if state.UpdateChecking {
		lines = append(
			lines,
			"",
			viewchrome.LipStyle(theme, theme.ColorYellow).Render("Checking for updates..."),
		)
	}
	if strings.TrimSpace(status.Error) != "" {
		lines = append(
			lines,
			"",
			theme.StyleError.Render("Check error: "+strings.TrimSpace(status.Error)),
		)
	}
	return viewchrome.RenderPaneBox(
		theme,
		false,
		state.Width,
		state.Height,
		viewhelpers.StringsJoin(lines),
	)
}

func updateActions(theme types.Theme, state types.ContentState) []string {
	actions := []string{
		theme.StyleHeader.Render("[r]") + theme.StyleDim.Render(" check again"),
	}
	status := state.UpdateStatus
	if status == nil {
		return actions
	}
	if strings.TrimSpace(status.UpdateCommand) != "" {
		copyLabel := "copy command"
		if viewchrome.IsMigrationCommand(status.UpdateCommand) {
			copyLabel = "copy migration command"
		}
		actions = append(
			actions,
			theme.StyleHeader.Render("[c]")+theme.StyleDim.Render(" "+copyLabel),
		)
	}
	if status.InstallScriptDeprecated {
		actions = append(
			actions,
			theme.StyleHeader.Render("[o]")+theme.StyleDim.Render(" open migration guide"),
		)
	} else if viewruntime.ShouldShowUpdatesView(status) {
		actions = append(
			actions,
			theme.StyleHeader.Render("[o]")+theme.StyleDim.Render(" open release notes"),
		)
	}
	actions = append(
		actions,
		theme.StyleHeader.Render("[d]")+
			theme.StyleDim.Render(" "+diagnosticsActionLabel(state.UpdateDiagnosticsExpanded)),
	)
	return actions
}

func statusCardLines(theme types.Theme, state types.ContentState) []string {
	status := state.UpdateStatus
	lines := []string{theme.StyleHeader.Render(statusHeadline(status))}
	if viewruntime.ShouldShowUpdatesView(status) {
		lines = append(lines, fieldBlock(theme, "Current Version", displayVersion(status.CurrentVersion), false)...)
		lines = append(lines, fieldBlock(theme, "Latest Version", displayVersion(status.LatestVersion), true)...)
		if strings.TrimSpace(status.UpdateCommand) != "" {
			lines = append(lines, fieldBlock(theme, "Update Command", strings.TrimSpace(status.UpdateCommand), true)...)
		}
		if source := strings.TrimSpace(string(status.InstallSource)); source != "" {
			lines = append(lines, fieldBlock(theme, "Install Source", sourceLabel(status.InstallSource), false)...)
		}
	} else {
		lines = append(lines, fieldBlock(theme, "Current Version", displayVersion(status.CurrentVersion), true)...)
		if source := strings.TrimSpace(string(status.InstallSource)); source != "" {
			lines = append(lines, fieldBlock(theme, "Install Source", sourceLabel(status.InstallSource), false)...)
		}
		if checked := relativeCheckedAt(status.CheckedAt); checked != "" {
			lines = append(lines, fieldBlock(theme, "Last Checked", checked, false)...)
		}
	}
	return lines
}

func fieldBlock(theme types.Theme, label, value string, highlight bool) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	lines := []string{
		theme.StyleDim.Render(label),
	}
	if highlight {
		lines = append(lines, theme.StyleHeader.Render("  "+value))
	} else {
		lines = append(lines, theme.StyleNormal.Render("  "+value))
	}
	return lines
}

func sourceLabel(source sharedtypes.InstallSource) string {
	switch source {
	case sharedtypes.InstallSourceBrew:
		return "Homebrew"
	case sharedtypes.InstallSourceWinget:
		return "winget"
	case sharedtypes.InstallSourceGo:
		return "go install"
	case sharedtypes.InstallSourceScript:
		return "GitHub install script"
	case sharedtypes.InstallSourceManual:
		return "manual"
	default:
		return strings.TrimSpace(string(source))
	}
}

func releaseNotesExcerptLines(theme types.Theme, notes string, width int) []string {
	excerpt, more := releaseNotesExcerpt(notes, 5)
	lines := []string{}
	if len(excerpt) == 0 {
		lines = append(lines, theme.StyleDim.Render("No release notes were published for this release."))
		return lines
	}
	for _, item := range excerpt {
		for _, row := range viewruntime.WrapMultilineBlock("• "+item, width) {
			lines = append(lines, theme.StyleNormal.Render(row))
		}
	}
	if more {
		lines = append(lines, theme.StyleDim.Render("[more]"))
	}
	return lines
}

func diagnosticsSectionLines(theme types.Theme, state types.ContentState) []string {
	status := state.UpdateStatus
	lines := []string{
		theme.StyleDim.Render(separatorLine(state.Width - 6)),
		theme.StyleHeader.Render(diagnosticsHeadline(state.UpdateDiagnosticsExpanded)),
	}
	if !state.UpdateDiagnosticsExpanded {
		lines = append(lines, theme.StyleDim.Render("[d] "+diagnosticsActionLabel(false)))
		return lines
	}

	addField := func(label, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		lines = append(lines, theme.StyleDim.Render(label))
		lines = append(lines, theme.StyleNormal.Render("  "+value))
	}

	addField("Release Tag", status.ReleaseTag)
	if status.ReleaseIsPrerelease {
		addField("Release Kind", "prerelease")
	} else if strings.TrimSpace(status.ReleaseTag) != "" {
		addField("Release Kind", "stable")
	}
	addField("Release Page", status.ReleaseURL)
	addField("Configured Channel", string(status.Channel))
	addField("Install Source", string(status.InstallSource))
	addField("Brew Formula", status.BrewFormula)
	addField("Last Checked", strings.TrimSpace(status.CheckedAt))
	addField("Update Command", status.UpdateCommand)
	if strings.TrimSpace(state.TUIExecutablePath) != "" {
		addField("TUI Path", state.TUIExecutablePath)
	}
	if strings.TrimSpace(state.KernelExecutablePath) != "" {
		addField("Engine Path", state.KernelExecutablePath)
	}
	lines = append(lines, theme.StyleDim.Render("[d] "+diagnosticsActionLabel(true)))
	return lines
}

func shouldShowMigrationChecklist(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if status.InstallScriptDeprecated || viewchrome.IsMigrationCommand(status.UpdateCommand) {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(status.InstallUnavailableReason)), "mismatch")
}

func migrationChecklistLines(theme types.Theme, state types.ContentState) []string {
	lines := []string{
		theme.StyleHeader.Render("Migration steps"),
	}
	steps := []string{
		"1. Run `crona backup`.",
		"2. Remove the old install with your current package manager.",
		"3. Install Crona again with the method you want to keep.",
		"4. Run `crona restore <backup-path>`.",
	}
	for _, step := range steps {
		for _, row := range viewruntime.WrapMultilineBlock(step, state.Width-6) {
			lines = append(lines, theme.StyleNormal.Render(row))
		}
	}
	return lines
}
