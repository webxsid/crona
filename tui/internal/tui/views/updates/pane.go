package updates

import (
	"github.com/charmbracelet/x/ansi"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"
)

// renderView lays Updates out as a status column (with diagnostics below when shown)
// beside full-height release notes; narrow screens stack the sections.
func renderView(theme types.Theme, state types.ContentState) string {
	width, height := state.Width, state.Height
	listWidth, notesWidth := viewchrome.LogPaneWidths(width)
	if notesWidth == 0 {
		listWidth = width
	}
	inner := state
	inner.Width = listWidth + 4
	status := statusSectionLines(theme, inner)
	meta := strings.Join(updateKeys(state), "  ")
	if ansi.StringWidth(meta)+len(" Status ")+3 > listWidth {
		keys := []string{}
		for _, line := range viewchrome.PackHints(updateKeys(state), max(1, listWidth-2)) {
			keys = append(keys, theme.StyleDim.Render(line))
		}
		status = append(append(keys, ""), status...)
		meta = ""
	}
	status = wrapLines(status, max(1, listWidth-2))
	notes := []string{theme.StyleDim.Render("Loading…")}
	notesTitle := "Release Notes"
	if state.UpdateStatus != nil {
		notesTitle = releaseNotesSectionTitle(state.UpdateStatus)
		notes = releaseNotesExcerptLines(theme, state.UpdateStatus.ReleaseNotes, max(1, max(notesWidth, listWidth)-2))
	}
	diagnostics := []string{}
	if state.UpdateDiagnosticsExpanded && state.UpdateStatus != nil {
		diagnostics = diagnosticsFieldLines(theme, inner)
	}
	section := func(title, meta string, focused bool, w, h int, lines []string) string {
		return viewchrome.RenderSection(theme, viewchrome.Section{Title: title, Meta: meta, Focused: focused, Width: w, Height: h, Lines: lines})
	}
	if notesWidth == 0 {
		statusHeight := min(len(status)+2, max(1, height/2))
		notesHeight := height - statusHeight
		diagHeight := 0
		if len(diagnostics) > 0 {
			notesHeight = min(len(notes)+2, max(1, (height-statusHeight)/2))
			diagHeight = height - statusHeight - notesHeight
		}
		parts := []string{section("Status", meta, true, width, statusHeight, status), section(notesTitle, "", false, width, notesHeight, notes)}
		if diagHeight > 0 {
			parts = append(parts, section("Diagnostics", "", false, width, diagHeight, diagnostics))
		}
		return strings.Join(parts, "\n")
	}
	left := section("Status", meta, true, listWidth, height, status)
	if len(diagnostics) > 0 {
		statusHeight := min(len(status)+2, max(1, height/2))
		left = section("Status", meta, true, listWidth, statusHeight, status) + "\n" +
			section("Diagnostics", "", false, listWidth, height-statusHeight, diagnostics)
	}
	return viewchrome.JoinColumns(theme, height, left, section(notesTitle, "", false, notesWidth, height, notes))
}

// statusSectionLines are the Status body: any banner or migration checklist, the status
// card, and the checking or error line.
func statusSectionLines(theme types.Theme, state types.ContentState) []string {
	status := state.UpdateStatus
	if status == nil {
		return []string{theme.StyleDim.Render("Loading…")}
	}
	lines := []string{}
	if banner := installScriptDeprecationBanner(theme, status); len(banner) > 0 {
		lines = append(append(lines, banner...), "")
	}
	if shouldShowMigrationChecklist(status) {
		lines = append(append(lines, migrationChecklistLines(theme, state)...), "")
	}
	lines = append(lines, statusCardLines(theme, state)...)
	if state.UpdateChecking {
		lines = append(lines, "", viewchrome.LipStyle(theme, theme.ColorYellow).Render("Checking for updates…"))
	}
	if strings.TrimSpace(status.Error) != "" {
		lines = append(lines, "", theme.StyleError.Render("Check error: "+strings.TrimSpace(status.Error)))
	}
	return lines
}

// updateKeys are the Status section's keys, plain so the header can fit them.
func updateKeys(state types.ContentState) []string {
	keys := []string{"[r] check again"}
	if status := state.UpdateStatus; status != nil {
		if strings.TrimSpace(status.UpdateCommand) != "" {
			label := "[c] copy command"
			if viewchrome.IsMigrationCommand(status.UpdateCommand) {
				label = "[c] copy migration command"
			}
			keys = append(keys, label)
		}
		if status.InstallScriptDeprecated {
			keys = append(keys, "[o] open migration guide")
		} else if viewruntime.ShouldShowUpdatesView(status) {
			keys = append(keys, "[o] open release notes")
		}
	}
	return append(keys, "[d] "+diagnosticsActionLabel(state.UpdateDiagnosticsExpanded))
}

func statusCardLines(theme types.Theme, state types.ContentState) []string {
	status := state.UpdateStatus
	lines := []string{theme.StyleHeader.Render(statusHeadline(status))}
	if viewruntime.ShouldShowUpdatesView(status) {
		lines = append(
			lines,
			fieldBlock(theme, "Current Version", displayVersion(status.CurrentVersion), false)...)
		lines = append(
			lines,
			fieldBlock(theme, "Latest Version", displayVersion(status.LatestVersion), true)...)
		if source := sharedtypes.NormalizeInstallSource(status.InstallSource); source == sharedtypes.InstallSourceScoop {
			lines = append(
				lines,
				fieldBlock(theme, "Channel", channelLabel(status.ReleaseChannel), false)...)
		}
		if strings.TrimSpace(status.UpdateCommand) != "" {
			lines = append(
				lines,
				fieldBlock(
					theme,
					"Update Command",
					strings.TrimSpace(status.UpdateCommand),
					true,
				)...)
		}
		if source := strings.TrimSpace(string(status.InstallSource)); source != "" {
			lines = append(
				lines,
				fieldBlock(theme, "Install Source", sourceLabel(status.InstallSource), false)...)
		}
	} else {
		lines = append(lines, fieldBlock(theme, "Current Version", displayVersion(status.CurrentVersion), true)...)
		if source := strings.TrimSpace(string(status.InstallSource)); source != "" {
			lines = append(lines, fieldBlock(theme, "Install Source", sourceLabel(status.InstallSource), false)...)
		}
		if source := sharedtypes.NormalizeInstallSource(status.InstallSource); source == sharedtypes.InstallSourceScoop {
			lines = append(lines, fieldBlock(theme, "Channel", channelLabel(status.ReleaseChannel), false)...)
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
	case sharedtypes.InstallSourceScoop:
		return "Scoop"
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

func channelLabel(channel sharedtypes.UpdateChannel) string {
	switch sharedtypes.NormalizeUpdateChannel(channel) {
	case sharedtypes.UpdateChannelBeta:
		return "Beta"
	default:
		return "Stable"
	}
}

func releaseNotesExcerptLines(theme types.Theme, notes string, width int) []string {
	excerpt, more := releaseNotesExcerpt(notes, 40)
	lines := []string{}
	if len(excerpt) == 0 {
		lines = append(
			lines,
			theme.StyleDim.Render("No release notes were published for this release."),
		)
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

func diagnosticsFieldLines(theme types.Theme, state types.ContentState) []string {
	status := state.UpdateStatus
	lines := []string{}

	addField := func(label, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		lines = append(lines, theme.StyleDim.Render(label))
		for _, row := range viewruntime.WrapMultilineBlock(value, max(1, state.Width-8)) {
			lines = append(lines, theme.StyleNormal.Render("  "+row))
		}
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
	addField("Install Channel", string(status.ReleaseChannel))
	addField("Brew Formula", status.BrewFormula)
	addField("Last Checked", strings.TrimSpace(status.CheckedAt))
	addField("Update Command", status.UpdateCommand)
	if strings.TrimSpace(state.TUIExecutablePath) != "" {
		addField("TUI Path", state.TUIExecutablePath)
	}
	if strings.TrimSpace(state.KernelExecutablePath) != "" {
		addField("Engine Path", state.KernelExecutablePath)
	}
	return lines
}

func shouldShowMigrationChecklist(status *api.UpdateStatus) bool {
	if status == nil {
		return false
	}
	if status.InstallScriptDeprecated || viewchrome.IsMigrationCommand(status.UpdateCommand) {
		return true
	}
	return strings.Contains(
		strings.ToLower(strings.TrimSpace(status.InstallUnavailableReason)),
		"mismatch",
	)
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

// wrapLines wraps each styled line to width, keeping its style.
func wrapLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.Split(ansi.Wrap(line, width, ""), "\n")...)
	}
	return out
}
