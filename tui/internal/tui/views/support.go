package views

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
)

func renderSupportView(theme Theme, state ContentState) string {
	lines := []string{
		theme.StylePaneTitle.Render("Support"),
		renderActionLine(theme, state.Width-6, ContextualActions(theme, ActionsState{View: state.View, Pane: state.Pane})),
		"",
		theme.StyleHeader.Render("File bugs, open discussions, follow releases, or generate a support bundle."),
		theme.StyleDim.Render("Copy diagnostics is lightweight. Bundle is the full redacted artifact for bug reports."),
		theme.StyleDim.Render("Watch GitHub releases or discussions for updates; roadmap details live in docs/roadmap.md."),
		"",
		theme.StyleDim.Render("Bug reports"),
		theme.StyleNormal.Render(helperpkg.SupportIssuesURL()),
		"",
		theme.StyleDim.Render("Discussions"),
		theme.StyleNormal.Render(helperpkg.SupportDiscussionsURL()),
		"",
		theme.StyleDim.Render("Releases"),
		theme.StyleNormal.Render(helperpkg.SupportReleasesURL()),
		"",
		theme.StyleDim.Render("Roadmap"),
		theme.StyleNormal.Render(helperpkg.SupportRoadmapURL()),
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
