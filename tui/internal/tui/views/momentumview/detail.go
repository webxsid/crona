package momentumview

import (
	"strings"

	sharedtypes "crona/shared/types"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	viewsummary "crona/tui/internal/tui/views/summary"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// DetailLines is the momentum's settings and streak, its chart in the room left, then the
// latest-bucket footnote, at most height lines.
func DetailLines(theme types.Theme, card sharedtypes.MomentumCard, width, height int) []string {
	def := card.Definition
	status := latestStatus(card)
	if status == "" {
		status = "–"
	}
	lines := []string{
		viewsummary.Row(theme, "Cadence", momentumhelpers.CadenceLabel(def.Period), width),
		viewsummary.Row(theme, "Target", momentumTargetSummary(def), width),
		viewsummary.Row(theme, momentumCardTargetLabel(def), momentumTargetSummaryNames(card), width),
		viewsummary.Row(theme, "Streak", streakLabel(card, card.Current)+" now · best "+streakLabel(card, card.Longest), width),
		viewsummary.Row(theme, "Status", momentumStatusStyle(theme, latestStatus(card)).Render(status), width),
	}
	if def.Description != nil && strings.TrimSpace(*def.Description) != "" {
		lines = append(lines, "")
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*def.Description), max(1, width), ""), "\n") {
			lines = append(lines, theme.StyleDim.Render(line))
		}
	}
	footnote := theme.StyleDim.Render(viewhelpers.Truncate(momentumSeriesFootnote(def, card.Series, def.Enabled), width))
	if room := height - len(lines) - 2; room > 0 {
		chart := strings.Split(renderMomentumSeries(theme, def, card.Series, width, def.Enabled), "\n")
		lines = append(lines, "")
		for _, line := range chart[:min(len(chart), room)] {
			lines = append(lines, ansi.Truncate(line, width, ""))
		}
	}
	lines = append(lines, footnote)
	return lines[:min(len(lines), max(0, height))]
}
