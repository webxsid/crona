package support

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewruntime "crona/tui/internal/tui/views/runtime"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderView(theme types.Theme, state types.ContentState) string {
	lines := []string{
		theme.StylePaneTitle.Render("Support"),
		viewchrome.RenderActionLine(theme, state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})),
		"",
		theme.StyleHeader.Render("File bugs, open discussions, follow releases, or generate a support bundle."),
		theme.StyleDim.Render("Copy diagnostics is lightweight. Bundle is the full redacted artifact for bug reports."),
		theme.StyleDim.Render("Beta builds expose [f9] for quick support actions. Use [v] plus a mnemonic to jump to views."),
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
	lines = append(lines, viewruntime.WrapMultilineBlock(helperpkg.SupportDiagnosticsSummary(state.KernelInfo, state.ExportAssets, state.UpdateStatus, state.Health, state.TUIExecutablePath, state.KernelExecutablePath), max(24, state.Width-6))...)
	if state.KernelInfo != nil {
		lines = append(lines,
			"",
			theme.StyleDim.Render("Transport"),
			theme.StyleNormal.Render(fmt.Sprintf("%s   %s", supportFallback(strings.TrimSpace(state.KernelInfo.Transport), "-"), supportFallback(strings.TrimSpace(state.KernelInfo.Endpoint), "-"))),
		)
	}
	return viewchrome.RenderPaneBox(theme, false, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}
