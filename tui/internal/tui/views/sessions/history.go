package sessions

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	sessionmeta "crona/tui/internal/tui/views/sessionmeta"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func renderHistoryView(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "sessions"
	cur := state.Cursors["sessions"]
	indices := sessionmeta.FilteredSessionIndices(state.SessionHistory, state.Filters["sessions"])
	total := len(indices)
	title := state.SessionHistoryTitle
	if strings.TrimSpace(title) == "" {
		title = "Session History"
	}
	subtitle := state.SessionHistoryMeta
	if strings.TrimSpace(subtitle) == "" {
		subtitle = "Recent sessions across the workspace"
	}
	actionLine := viewchrome.RenderPaneActionLine(theme, state.Filters["sessions"], state.Width-6, viewchrome.PaneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render(title), theme.StyleDim.Render(subtitle), actionLine}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No sessions recorded"))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	dateW, durW := 16, 10
	issueW := state.Width - dateW - durW - 12
	if issueW < 18 {
		issueW = 18
	}
	header := fmt.Sprintf("%-2s %-*s %-*s %s", "", dateW, "Ended", durW, "Duration", "Notes")
	lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(header, state.Width-6)))
	inner := viewchrome.RemainingPaneHeight(state.Height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render("..."))
	}
	for pos := start; pos < end; pos++ {
		entry := state.SessionHistory[indices[pos]]
		ended := entry.StartTime
		if entry.EndTime != nil && *entry.EndTime != "" {
			ended = *entry.EndTime
		}
		ended = sessionmeta.FormatSessionTimestamp(ended)
		duration := sessionmeta.FormatSessionDuration(entry.DurationSeconds, entry.StartTime, entry.EndTime)
		note := sessionmeta.SessionHistorySummary(entry)
		row := fmt.Sprintf("%-2s %-*s %-*s %s", "", dateW, ended, durW, duration, viewhelpers.Truncate(note, issueW))
		if pos == cur && active {
			lines = append(lines, theme.StyleCursor.Render("▶ "+viewhelpers.Truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		} else if pos == cur {
			lines = append(lines, theme.StyleSelected.Render("  "+viewhelpers.Truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		} else {
			lines = append(lines, theme.StyleNormal.Render("  "+viewhelpers.Truncate(strings.TrimPrefix(row, "  "), state.Width-6)))
		}
	}
	if end < total {
		lines = append(lines, theme.StyleDim.Render("..."))
	}
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}
