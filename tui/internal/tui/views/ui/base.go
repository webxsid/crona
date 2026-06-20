package ui

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewtypes "crona/tui/internal/tui/views/types"
)

type PaneBase struct {
	ID       string
	Title    string
	Subtitle string
	Focused  bool
	Width    int
	Height   int
	Cursor   int
	Scroll   int
}

func (p *PaneBase) Resize(width, height int) {
	p.Width = width
	p.Height = height
}

func (p *PaneBase) Focus(focused bool) {
	p.Focused = focused
}

func (p PaneBase) TitleLine(theme viewtypes.Theme, title string) string {
	return theme.StylePaneTitle.Render(title)
}

func (p PaneBase) HeaderLines(titleLine, contextLine, subtitle string) []string {
	lines := []string{titleLine}
	if strings.TrimSpace(contextLine) != "" {
		lines = append(lines, contextLine)
	}
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, subtitle)
	}
	return lines
}

func (p PaneBase) ControlLine(
	theme viewtypes.Theme,
	width int,
	active bool,
	actions []string,
	showFilter bool,
) string {
	if active {
		return viewchrome.RenderPaneActionLine(theme, actions, width)
	}
	if showFilter {
		return ""
	}
	return theme.StyleDim.Render("")
}

func (p PaneBase) MoreAbove(theme viewtypes.Theme, count int) string {
	if count <= 0 {
		return ""
	}
	return theme.StyleDim.Render(fmt.Sprintf("↑ %d more", count))
}

func (p PaneBase) MoreBelow(theme viewtypes.Theme, count int) string {
	if count <= 0 {
		return ""
	}
	return theme.StyleDim.Render(fmt.Sprintf("↓ %d more", count))
}

func (p PaneBase) Render(theme viewtypes.Theme, content string) string {
	return viewchrome.RenderPaneBox(
		theme,
		p.Focused,
		p.Width,
		p.Height,
		viewhelpers.StringsJoin([]string{content}),
	)
}
