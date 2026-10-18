package momentumview

import (
	"fmt"

	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"
)

// StripRows sums the momentums: counts over streaks, column groups dropped right-first.
func StripRows(theme types.Theme, state types.ContentState, width int) []string {
	return viewchrome.WithWindowHint(theme, state.MomentumWindowDays, width, func(width int) []string { return stripRows(theme, state, width) })
}

func stripRows(theme types.Theme, state types.ContentState, width int) []string {
	if width <= 0 {
		return []string{"", ""}
	}
	label := func(s string) string { return theme.StyleHeader.Render(s) }
	active, inactive := 0, 0
	counts := map[string]int{}
	best, longest := "–", "–"
	bestValue, longestValue := -1, -1
	for _, card := range state.MomentumCards {
		if card.Definition.Enabled {
			active++
			counts[latestStatus(card)]++
			if card.Current > bestValue {
				bestValue, best = card.Current, card.Definition.Name+" "+streakLabel(card, card.Current)
			}
		} else {
			inactive++
		}
		if card.Longest > longestValue {
			longestValue, longest = card.Longest, card.Definition.Name+" "+streakLabel(card, card.Longest)
		}
	}
	groups := [][2]string{
		{fmt.Sprintf("%s %d  inactive %d", label("Active"), active, inactive), label("Best streak") + "  " + best},
		{fmt.Sprintf("%s  met %d  near %d  missed %d", label("Latest"), counts["met"], counts["near"], counts["missed"]), label("Longest ever") + "  " + longest},
	}
	return viewchrome.StripRows(groups, width)
}
