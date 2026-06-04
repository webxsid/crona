package support

import (
	"fmt"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewruntime "crona/tui/internal/tui/views/runtime"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("support", false, renderView),
		},
	}
}

func renderView(theme types.Theme, state types.ContentState) string {
	lines := []string{
		theme.StylePaneTitle.Render("Support"),
		viewchrome.RenderActionLine(
			theme,
			state.Width-6,
			viewchrome.ContextualActions(
				theme,
				viewchrome.ActionsState{View: state.View, Pane: state.Pane},
			),
		),
		"",
		theme.StyleHeader.Render(
			"Report bugs, open discussions, check releases, or generate a support bundle.",
		),
		theme.StyleDim.Render(
			"Copy diagnostics is the quick clipboard version. Bundle is the full redacted ZIP for bug reports.",
		),
		theme.StyleDim.Render(
			"Beta builds expose [f9] support actions. Use [v] plus a mnemonic to jump to views.",
		),
		theme.StyleDim.Render(
			"Releases and discussions track updates; roadmap details live in docs/roadmap.md.",
		),
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
	lines = append(
		lines,
		viewruntime.WrapMultilineBlock(
			helperpkg.SupportDiagnosticsSummary(
				state.KernelInfo,
				state.ExportAssets,
				state.UpdateStatus,
				state.Health,
				state.TUIExecutablePath,
				state.KernelExecutablePath,
			),
			max(24, state.Width-6),
		)...)
	if state.KernelInfo != nil {
		lines = append(
			lines,
			"",
			theme.StyleDim.Render("Transport"),
			theme.StyleNormal.Render(
				fmt.Sprintf(
					"%s   %s",
					supportFallback(strings.TrimSpace(state.KernelInfo.Transport), "-"),
					supportFallback(strings.TrimSpace(state.KernelInfo.Endpoint), "-"),
				),
			),
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
