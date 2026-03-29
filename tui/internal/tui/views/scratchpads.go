package views

import (
	"strings"
)

func renderScratchpadView(theme Theme, state ContentState) string {
	if !state.ScratchpadOpen {
		return renderScratchpadPlaceholder(theme, state)
	}

	innerW := max(10, state.Width-6)
	contentH := max(1, state.Height-7)
	lines := []string{
		theme.StylePaneTitle.Render("Scratchpads"),
		theme.StyleHeader.Render(truncate(state.ScratchpadName, innerW)),
		theme.StyleDim.Render(truncate(state.ScratchpadPath, innerW)),
	}

	viewLines := strings.Split(state.ScratchpadRendered, "\n")
	if len(viewLines) > contentH {
		viewLines = viewLines[:contentH]
	}
	lines = append(lines, viewLines...)
	lines = append(lines, theme.StyleDim.Render("[h/l] switch  [j/k] scroll  [e] edit  [esc] close"))

	box := theme.StyleInactive
	if state.Pane == "scratchpads" {
		box = theme.StyleActive
	}
	return box.Width(state.Width-2).Height(state.Height-2).Padding(0, 1).Render(strings.Join(lines, "\n"))
}
