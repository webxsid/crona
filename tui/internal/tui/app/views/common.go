package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderScratchpadPlaceholder(theme Theme, state ContentState) string {
	actions := paneActionsForState(theme, state, state.Pane == "scratchpads")
	return renderSimplePaneWithActions(theme, "Scratchpads", state.Filters["scratchpads"], state.Cursors["scratchpads"], scratchpadItems(state.Scratchpads), state.Pane == "scratchpads", state.Width, state.Height, "No scratchpads — [a] create new", actions)
}

func renderSimplePane(theme Theme, title, filter string, cursor int, items []string, active bool, width, height int, empty string) string {
	return renderSimplePaneWithActions(theme, title, filter, cursor, items, active, width, height, empty, nil)
}

func renderSimplePaneWithActions(theme Theme, title, filter string, cursor int, items []string, active bool, width, height int, empty string, actions []string) string {
	indices := filteredStrings(items, filter)
	total := len(indices)
	inner := height - 5
	if inner < 1 {
		inner = 1
	}
	if !active {
		actions = nil
	}
	lines := []string{theme.StylePaneTitle.Render(title), renderPaneActionLine(theme, filter, width-6, actions)}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render(empty))
	} else {
		start, end := listWindow(cursor, total, inner)
		if start > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			lines = append(lines, renderPaneRowStyled(theme, i, cursor, active, items[indices[i]], nil, width))
		}
		if remaining := total - end; remaining > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
		}
	}
	return renderPaneBox(theme, active, width, height, stringsJoin(lines))
}

func renderFilterLine(theme Theme, filter string, width int) string {
	return renderPaneActionLine(theme, filter, width, nil)
}

func renderPaneActionLine(theme Theme, filter string, width int, actions []string) string {
	segments := []string{filterToken(theme, filter)}
	segments = append(segments, actions...)
	return joinActionSegments(segments, width)
}

func RenderPaneActionLine(theme Theme, filter string, width int, actions []string) string {
	return renderPaneActionLine(theme, filter, width, actions)
}

func renderActionLine(theme Theme, width int, actions []string) string {
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
	kept := make([]string, 0, len(segments))
	used := 0
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			continue
		}
		segmentWidth := lipgloss.Width(segment)
		additional := segmentWidth
		if len(kept) > 0 {
			additional += 3
		}
		if used+additional > width {
			break
		}
		kept = append(kept, segment)
		used += additional
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, "   ")
}

func paneActionsForState(theme Theme, state ContentState, active bool) []string {
	if !active {
		return nil
	}
	return ContextualActions(theme, ActionsState{
		View:           state.View,
		Pane:           state.Pane,
		ScratchpadOpen: state.ScratchpadOpen,
		TimerState:     timerStateFromContent(state),
	})
}

func timerStateFromContent(state ContentState) string {
	if state.Timer == nil {
		return ""
	}
	return state.Timer.State
}

func renderPaneRowStyled(theme Theme, i, cur int, active bool, text string, contentStyle *lipgloss.Style, width int) string {
	line := truncate(text, width-6)
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

func renderPaneBox(theme Theme, active bool, width, height int, content string) string {
	box := theme.StyleInactive
	if active {
		box = theme.StyleActive
	}
	return box.Width(width-2).Height(height-2).Padding(0, 1).Render(content)
}
