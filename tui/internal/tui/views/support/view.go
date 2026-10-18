package support

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("support", false, renderView),
		},
		RenderFunc: renderView,
	}
}

// Action is one Support row: its key, label and the link it opens, if any.
type Action struct {
	Key, Label, URL string
}

// Actions are Support's rows in the order input's supportActions runs them.
func Actions() []Action {
	return []Action{
		{"o", "Feedback & roadmap", helperpkg.SupportFeedbackRoadmapURL()},
		{"r", "Releases", helperpkg.SupportReleasesURL()},
		{"g", "Documentation", helperpkg.SupportDocumentationURL()},
		{"c", "Copy diagnostics", ""},
		{"b", "Generate bundle", ""},
	}
}

// renderView lays Support out as an action list beside the diagnostics; narrow screens stack them.
func renderView(theme types.Theme, state types.ContentState) string {
	width, height := state.Width, state.Height
	listWidth, detailWidth := viewchrome.LogPaneWidths(width)
	if detailWidth == 0 {
		listWidth = width
	}
	actions := Actions()
	cursor := min(max(0, state.SupportCursor), len(actions)-1)
	actionLines := []string{theme.StyleDim.Render("Report bugs, share feedback, or build a support bundle."), ""}
	for i, action := range actions {
		line := fmt.Sprintf("[%s] %s", action.Key, action.Label)
		if i == cursor {
			actionLines = append(actionLines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
		} else {
			actionLines = append(actionLines, "  "+line)
		}
		if action.URL != "" {
			for _, part := range strings.Split(ansi.Wrap(action.URL, max(1, listWidth-7), "/"), "\n") {
				actionLines = append(actionLines, "    "+theme.StyleDim.Render(part))
			}
		}
	}
	diagnosticsWidth := listWidth
	if detailWidth > 0 {
		diagnosticsWidth = detailWidth
	}
	diagnostics := diagnosticsLines(theme, state, max(1, diagnosticsWidth-2))
	section := func(title string, focused bool, w, h int, lines []string) string {
		return viewchrome.RenderSection(theme, viewchrome.Section{Title: title, Focused: focused, Width: w, Height: h, Lines: lines})
	}
	if detailWidth == 0 {
		actionsHeight := min(len(actionLines)+1, height)
		return section("Actions", true, width, actionsHeight, actionLines) + "\n" +
			section("Diagnostics", false, width, max(1, height-actionsHeight), diagnostics)
	}
	return viewchrome.JoinColumns(theme, height,
		section("Actions", true, listWidth, height, actionLines),
		section("Diagnostics", false, detailWidth, height, diagnostics))
}

func diagnosticsLines(theme types.Theme, state types.ContentState, width int) []string {
	summary := helperpkg.SupportDiagnosticsSummary(
		state.KernelInfo, state.ExportAssets, state.UpdateStatus, state.Health,
		state.TUIExecutablePath, state.KernelExecutablePath,
	)
	lines := strings.Split(ansi.Wrap(summary, width, "/"), "\n")
	if state.KernelInfo != nil {
		lines = append(lines, "", theme.StyleDim.Render("Transport"),
			supportFallback(strings.TrimSpace(state.KernelInfo.Transport), "-")+"  "+theme.StyleDim.Render(supportFallback(strings.TrimSpace(state.KernelInfo.Endpoint), "-")))
	}
	return lines
}
