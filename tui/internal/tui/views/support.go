package views

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
)

const (
	SupportIssueURL   = "https://github.com/webxsid/crona/issues/new/choose"
	SupportProjectURL = "https://github.com/webxsid/crona"
)

func renderSupportView(theme Theme, state ContentState) string {
	lines := []string{
		theme.StylePaneTitle.Render("Support"),
		renderActionLine(theme, state.Width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})),
		"",
		theme.StyleHeader.Render("Report bugs, open project links, or copy diagnostics."),
		"",
		theme.StyleDim.Render("Issue tracker"),
		theme.StyleNormal.Render(SupportIssueURL),
		"",
		theme.StyleDim.Render("Project"),
		theme.StyleNormal.Render(SupportProjectURL),
		"",
		theme.StyleDim.Render("Diagnostics"),
	}
	lines = append(lines, wrapMultiline(helperpkg.SupportDiagnosticsSummary(state.KernelInfo, state.ExportAssets, state.UpdateStatus, state.Health, state.TUIExecutablePath, state.KernelExecutablePath), max(24, state.Width-6))...)
	if state.KernelInfo != nil {
		lines = append(lines,
			"",
			theme.StyleDim.Render("Transport"),
			theme.StyleNormal.Render(fmt.Sprintf("%s   %s", supportFallback(strings.TrimSpace(state.KernelInfo.Transport), "-"), supportFallback(strings.TrimSpace(state.KernelInfo.Endpoint), "-"))),
		)
	}
	return renderPaneBox(theme, false, state.Width, state.Height, stringsJoin(lines))
}

func supportFallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return strings.TrimSpace(value)
}
