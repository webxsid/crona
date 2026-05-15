package dialogs

import (
	"fmt"
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
)

func renderOnboardingScreen(theme Theme, state controllerpkg.State) string {
	contentWidth := state.Width - 16

	contentWidth = min(contentWidth, 72)
	contentWidth = max(contentWidth, 100)

	progress := onboardingProgress(theme, state.TelemetryStep)
	rows := []string{
		centerBlock(renderOnboardingLogo(theme), contentWidth),
		"",
		centerLine(theme.StyleHeader.Render("Let's get things set up."), contentWidth),
		"",
		centerLine(progress, contentWidth),
		"",
	}
	switch state.TelemetryStep {
	case 0:
		rows = append(rows,
			centerLine(theme.StyleDim.Render("Your work stays on this machine, and you can change these choices later."), contentWidth),
			"",
			centerLine(theme.StyleHeader.Render("What to expect"), contentWidth),
			centerLine(theme.StyleNormal.Render("Keep issues, sessions, habits, reports, and wellbeing in one place."), contentWidth),
			centerLine(theme.StyleNormal.Render("Move through your work without leaving the terminal."), contentWidth),
			centerLine(theme.StyleNormal.Render("Set your privacy preferences before you continue."), contentWidth),
		)
		rows = appendDialogFooter(theme, state, rows, "[l] next section   [h] previous section")
	case 1:
		rows = append(rows,
			centerLine(theme.StyleHeader.Render("Your workspace"), contentWidth),
			centerLine(theme.StyleDim.Render("A few things become easier from here."), contentWidth),
			"",
			centerLine(theme.StyleNormal.Render("Daily work, focus sessions, habits, and reports all stay connected."), contentWidth),
			centerLine(theme.StyleNormal.Render("Wellbeing and momentum stay visible without extra setup."), contentWidth),
			centerLine(theme.StyleNormal.Render("Everything stays local, searchable, and quick."), contentWidth),
		)
		rows = appendDialogFooter(theme, state, rows, "[space] toggle	[l] next section   [h] previous section")
	case 2:
		usageLine := toggleLine(state.TelemetryPrivacyCursor == 0, state.TelemetryUsage, "Share usage signals")
		diagnosticsLine := toggleLine(state.TelemetryPrivacyCursor == 1, state.TelemetryErrors, "Share diagnostics")
		rows = append(rows,
			centerLine(theme.StyleHeader.Render("Privacy choices"), contentWidth),
			centerLine(theme.StyleDim.Render("Choose how much anonymous feedback you want to share."), contentWidth),
			"",
			centerLine(usageLine, contentWidth),
			centerLine(diagnosticsLine, contentWidth),
			"",
			centerLine(theme.StyleDim.Render("Usage signals help improve the app and catch issues early."), contentWidth),
			centerLine(theme.StyleDim.Render("Diagnostics help us investigate failures without your work content."), contentWidth),
		)
		rows = appendDialogFooter(theme, state, rows, "[space] toggle   [j/k] choose   [l] next section   [h] previous section")
	default:
		startLabel := reviewChoiceLine(state.TelemetryReviewCursor == 0, "Start Crona")
		restartLabel := reviewChoiceLine(state.TelemetryReviewCursor == 1, "Start and Restart Now")
		backLabel := reviewChoiceLine(state.TelemetryReviewCursor == 2, "Back")
		rows = append(rows,
			centerLine(theme.StyleHeader.Render("Review your choices"), contentWidth),
			centerLine(theme.StyleDim.Render("You can update these later in Settings."), contentWidth),
			"",
			centerLine(theme.StyleDim.Render("Usage signals"), contentWidth),
			centerLine(theme.StyleHeader.Render(telemetryStateLabel(state.TelemetryUsage)), contentWidth),
			"",
			centerLine(theme.StyleDim.Render("Diagnostics"), contentWidth),
			centerLine(theme.StyleHeader.Render(telemetryStateLabel(state.TelemetryErrors)), contentWidth),
			"",
			centerLine(theme.StyleError.Render("Changes take effect after restart."), contentWidth),
			centerLine(theme.StyleDim.Render("Choose how you want to finish."), contentWidth),
			"",
			centerLine(startLabel, contentWidth),
			centerLine(restartLabel, contentWidth),
			centerLine(backLabel, contentWidth),
		)
		rows = appendDialogFooter(theme, state, rows, "[j/k] choose   [enter] finish   [l] next section   [h] previous section")
	}
	return lipgloss.NewStyle().Width(contentWidth).Render(strings.Join(rows, "\n"))
}

func renderOnboardingLogo(theme Theme) string {
	lines := []string{
		" ██████╗██████╗  ██████╗ ███╗   ██╗ █████╗ ",
		"██╔════╝██╔══██╗██╔═══██╗████╗  ██║██╔══██╗",
		"██║     ██████╔╝██║   ██║██╔██╗ ██║███████║",
		"██║     ██╔══██╗██║   ██║██║╚██╗██║██╔══██║",
		"╚██████╗██║  ██║╚██████╔╝██║ ╚████║██║  ██║",
		" ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝",
		"",
	}
	return strings.Join(lines, "\n")
}

func onboardingProgress(theme Theme, step int) string {
	steps := []string{"Welcome", "Workspace", "Privacy", "Review"}
	progress := make([]string, 0, len(steps))
	for i, label := range steps {
		piece := fmt.Sprintf("%d.%s", i+1, label)
		if i == step {
			progress = append(progress, theme.StyleCursor.Render(piece))
			continue
		}
		progress = append(progress, theme.StyleDim.Render(piece))
	}
	return strings.Join(progress, "   ")
}

func centerLine(line string, width int) string {
	if width <= 0 {
		return line
	}
	if lipgloss.Width(line) >= width {
		return line
	}
	return strings.Repeat(" ", (width-lipgloss.Width(line))/2) + line
}

func centerBlock(block string, width int) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = centerLine(line, width)
	}
	return strings.Join(lines, "\n")
}

func toggleLine(selected, enabled bool, label string) string {
	prefix := "  "
	if selected {
		prefix = "▶ "
	}
	return prefix + toggleLabel(enabled, label)
}

func reviewChoiceLine(selected bool, label string) string {
	prefix := "  "
	if selected {
		prefix = "▶ "
	}
	return prefix + label
}
