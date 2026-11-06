package issues

import (
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"
)

func newIssueLine(theme types.Theme, state types.ContentState, width int) string {
	hints := []string{viewchrome.KeyHint(theme, "a", "new issue")}
	if state.Timer == nil || state.Timer.State == "idle" {
		hints = append(hints, viewchrome.KeyHint(theme, "A", "quick timer"))
	}
	return viewchrome.FitHints(hints, width)
}
