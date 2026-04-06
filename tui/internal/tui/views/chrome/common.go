package viewchrome

import (
	"fmt"
	"strings"

	viewruntime "crona/tui/internal/tui/views/runtime"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	"github.com/charmbracelet/lipgloss"
)

func RenderScratchpadPlaceholder(theme Theme, state ContentState) string {
	actions := PaneActionsForState(theme, state, state.Pane == "scratchpads")
	return RenderSimplePaneWithActions(theme, "Scratchpads", state.Filters["scratchpads"], state.Cursors["scratchpads"], viewhelpers.ScratchpadItems(state.Scratchpads), state.Pane == "scratchpads", state.Width, state.Height, "No scratchpads — [a] create new", actions)
}

func RenderSimplePane(theme Theme, title, filter string, cursor int, items []string, active bool, width, height int, empty string) string {
	return RenderSimplePaneWithActions(theme, title, filter, cursor, items, active, width, height, empty, nil)
}

func RenderSimplePaneWithActions(theme Theme, title, filter string, cursor int, items []string, active bool, width, height int, empty string, actions []string) string {
	indices := viewhelpers.FilteredStrings(items, filter)
	total := len(indices)
	if !active {
		actions = nil
	}
	actionLine := RenderPaneActionLine(theme, filter, width-6, actions)
	lines := []string{theme.StylePaneTitle.Render(title), actionLine}
	inner := RemainingPaneHeight(height, lines)
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render(empty))
	} else {
		start, end := ListWindow(cursor, total, inner)
		if start > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			lines = append(lines, RenderPaneRowStyled(theme, i, cursor, active, items[indices[i]], nil, width))
		}
		if remaining := total - end; remaining > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
		}
	}
	return RenderPaneBox(theme, active, width, height, viewhelpers.StringsJoin(lines))
}

func RenderFilterLine(theme Theme, filter string, width int) string {
	return RenderPaneActionLine(theme, filter, width, nil)
}

func RenderPaneActionLine(theme Theme, filter string, width int, actions []string) string {
	segments := []string{filterToken(theme, filter)}
	segments = append(segments, actions...)
	return joinActionSegments(segments, width)
}

func RenderActionLine(theme Theme, width int, actions []string) string {
	return joinActionSegments(actions, width)
}

func filterToken(theme Theme, filter string) string {
	if strings.TrimSpace(filter) == "" {
		return theme.StyleDim.Render("[/] filter")
	}
	return theme.StyleHeader.Render("/ " + filter)
}

func joinActionSegments(segments []string, width int) string {
	if width < 1 {
		return ""
	}
	rows := []string{}
	current := make([]string, 0, len(segments))
	used := 0
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			continue
		}
		segmentWidth := lipgloss.Width(segment)
		additional := segmentWidth
		if len(current) > 0 {
			additional += 3
		}
		if used+additional > width && len(current) > 0 {
			rows = append(rows, strings.Join(current, "   "))
			current = []string{segment}
			used = segmentWidth
			continue
		}
		if segmentWidth > width && len(current) == 0 {
			rows = append(rows, segment)
			used = 0
			continue
		}
		current = append(current, segment)
		used += additional
	}
	if len(current) > 0 {
		rows = append(rows, strings.Join(current, "   "))
	}
	if len(rows) == 0 {
		return ""
	}
	return strings.Join(rows, "\n")
}

func RemainingPaneHeight(height int, lines []string) int {
	inner := height - 2 - RenderedLineCount(lines)
	if inner < 1 {
		return 1
	}
	return inner
}

func RenderedLineCount(lines []string) int {
	total := 0
	for _, line := range lines {
		h := lipgloss.Height(line)
		if h < 1 {
			h = 1
		}
		total += h
	}
	return total
}

func PaneActionsForState(theme Theme, state ContentState, active bool) []string {
	if !active {
		return nil
	}
	return ContextualActions(theme, ActionsState{
		View:                   state.View,
		Pane:                   state.Pane,
		ScratchpadOpen:         state.ScratchpadOpen,
		TimerState:             timerStateFromContent(state),
		RestModeActive:         state.RestModeActive,
		AwayModeActive:         state.AwayModeActive,
		UpdateVisible:          viewruntime.ShouldShowUpdatesView(state.UpdateStatus),
		UpdateInstallAvailable: state.UpdateInstallAvailable,
	})
}

func timerStateFromContent(state ContentState) string {
	if state.Timer == nil {
		return ""
	}
	return state.Timer.State
}

func RenderPaneRowStyled(theme Theme, i, cur int, active bool, text string, contentStyle *lipgloss.Style, width int) string {
	line := viewhelpers.Truncate(text, width-6)
	if contentStyle != nil {
		line = contentStyle.Render(line)
	}
	if i == cur && active {
		return theme.StyleCursor.Render("▶ " + line)
	}
	if i == cur {
		return theme.StyleSelected.Render("  " + line)
	}
	return theme.StyleNormal.Render("  " + line)
}

func RenderPaneBox(theme Theme, active bool, width, height int, content string) string {
	box := theme.StyleInactive
	if active {
		box = theme.StyleActive
	}
	return box.Width(width-2).Height(height-2).Padding(0, 1).Render(clipBoxContent(content, height-2))
}

func clipBoxContent(content string, maxLines int) string {
	if maxLines < 1 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	if maxLines == 1 {
		return "..."
	}
	clipped := append([]string{}, lines[:maxLines-1]...)
	clipped = append(clipped, "...")
	return strings.Join(clipped, "\n")
}
