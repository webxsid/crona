package support

import (
	"github.com/charmbracelet/x/ansi"
	"strings"

	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"
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
		{"o", "Feedback & discussions", helperpkg.SupportFeedbackRoadmapURL()},
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
	actionLines := []string{
		theme.StyleDim.Render("Report bugs, share feedback, or build a support bundle."),
		"",
	}
	labels := make([]string, 0, len(actions))
	for _, action := range actions {
		labels = append(labels, action.Label)
	}
	column := viewchrome.LabelColumn(labels, 24)
	for i, action := range actions {
		line := viewchrome.LipStyle(theme, theme.ColorCyan).Render(action.Key) + "  " + viewchrome.LabelCell(action.Label, column)
		if action.URL != "" {
			url := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(action.URL, "https://"), "http://"), "/")
			line += "  " + theme.StyleDim.Render(url)
		}
		line = ansi.Truncate(line, max(1, listWidth-4), "…")
		if i == cursor {
			actionLines = append(actionLines, viewchrome.SelectedRow(theme, line))
		} else {
			actionLines = append(actionLines, "  "+viewchrome.Faint(line))
		}
	}
	diagnosticsWidth := listWidth
	if detailWidth > 0 {
		diagnosticsWidth = detailWidth
	}
	diagnostics := diagnosticsLines(theme, state, max(1, diagnosticsWidth-2))
	section := func(title string, focused bool, w, h int, lines []string) string {
		return viewchrome.RenderSection(
			theme,
			viewchrome.Section{Title: title, Focused: focused, Width: w, Height: h, Lines: lines},
		)
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

// diagnosticsTable lays the diagnostics summary's "Key: value" lines out as a table: keys
// dim in a column, values beside them with paths shortened to fit.
func diagnosticsTable(theme types.Theme, summary string, width int) []string {
	type pair struct{ key, value string }
	pairs := []pair{}
	column := 0
	for _, line := range strings.Split(summary, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			pairs = append(pairs, pair{value: line})
			continue
		}
		pairs = append(pairs, pair{key, strings.TrimSpace(value)})
		column = max(column, ansi.StringWidth(key))
	}
	column = min(column, max(8, width*2/5))
	lines := make([]string, 0, len(pairs))
	for _, p := range pairs {
		if p.key == "" {
			lines = append(lines, ansi.Truncate(p.value, width, "…"))
			continue
		}
		room := max(4, width-column-2)
		value := p.value
		if strings.HasPrefix(value, "/") {
			value = helperpkg.ShortPath(value, room)
		}
		lines = append(lines, theme.StyleDim.Render(viewchrome.LabelCell(p.key, column))+"  "+
			viewchrome.CalmValue(theme, ansi.Truncate(value, room, "…")))
	}
	return lines
}

func diagnosticsLines(theme types.Theme, state types.ContentState, width int) []string {
	summary := helperpkg.SupportDiagnosticsSummary(
		state.KernelInfo, state.ExportAssets, state.UpdateStatus, state.Health,
		state.TUIExecutablePath, state.KernelExecutablePath,
	)
	lines := diagnosticsTable(theme, summary, width)
	if state.KernelInfo != nil {
		lines = append(
			lines,
			"",
			theme.StyleDim.Render("Transport"),
			supportFallback(
				strings.TrimSpace(state.KernelInfo.Transport),
				"-",
			)+"  "+theme.StyleDim.Render(
				helperpkg.ShortPath(supportFallback(strings.TrimSpace(state.KernelInfo.Endpoint), "-"), max(8, width-14)),
			),
		)
	}
	return lines
}
