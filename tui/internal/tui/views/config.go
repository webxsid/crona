package views

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
)

func renderConfigView(theme Theme, state ContentState) string {
	active := state.Pane == "config"
	cur := state.Cursors["config"]
	items := configItems(state.ExportAssets)
	indices := filteredStrings(items, state.Filters["config"])
	total := len(indices)
	actionLine := renderPaneActionLine(theme, state.Filters["config"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render("Config"), actionLine}
	if state.ExportAssets == nil {
		lines = append(lines, theme.StyleDim.Render("Loading export assets..."))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No config items match the current filter"))
		return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
	}
	inner := remainingPaneHeight(state.Height, lines)
	start, end := listWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(lines, renderPaneRowStyled(theme, i, cur, active, items[indices[i]], nil, state.Width))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}

func configItems(status *api.ExportAssetStatus) []string {
	if status == nil {
		return nil
	}
	items := make([]string, 0, len(status.TemplateAssets)+3)
	for _, asset := range status.TemplateAssets {
		items = append(items, fmt.Sprintf("%-24s %s", asset.Label, configAssetValue(asset)))
	}
	pdfRenderer := "unavailable"
	if status.PDFRendererAvailable {
		pdfRenderer = status.PDFRendererName
	}
	items = append(items, fmt.Sprintf("%-24s %s", "Reports directory", status.ReportsDir))
	items = append(items, fmt.Sprintf("%-24s %s", "ICS export directory", status.ICSDir))
	items = append(items, fmt.Sprintf("%-24s %s", "PDF renderer", pdfRenderer))
	return items
}

func configAssetValue(asset api.ExportTemplateAsset) string {
	if asset.Resettable {
		switch {
		case asset.Customized:
			return "[customized]"
		case asset.UpdateAvailable:
			return "[new default available]"
		default:
			return "[default]"
		}
	}
	path := strings.TrimSpace(asset.UserPath)
	if path == "" {
		return "-"
	}
	return path
}
