package ops

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func Render(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "ops"
	cur := state.Cursors["ops"]
	indices := reverseIndices(filteredOpIndices(state.Ops, state.Filters["ops"]))
	total := len(indices)
	lines := []string{
		theme.StylePaneTitle.Render("Ops Log"),
		theme.StyleDim.Render(fmt.Sprintf("limit: %d", currentOpsLimit(state))),
		viewchrome.RenderPaneActionLine(theme, state.Filters["ops"], state.Width-6, viewchrome.ContextualActions(theme, viewchrome.ActionsState{View: state.View, Pane: state.Pane})),
	}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No operations recorded"))
		return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
	}
	timeW, entityW, actionW := 19, max(10, state.Width/8), 10
	targetW := state.Width - timeW - entityW - actionW - 12
	if targetW < 12 {
		targetW = 12
	}
	header := fmt.Sprintf("%-2s %-19s %-*s %-*s %s", "", "Time", entityW, "Entity", actionW, "Action", "Target")
	lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate(header, state.Width-6)))
	inner := viewchrome.RemainingPaneHeight(state.Height, lines)
	start, end := viewchrome.ListWindow(cur, total, inner)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(lines, renderOpRow(theme, state, i, cur, active, state.Ops[indices[i]], entityW, actionW, targetW))
	}
	if remaining := total - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
	}
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, viewhelpers.StringsJoin(lines))
}

func renderOpRow(theme types.Theme, state types.ContentState, i, cur int, active bool, op api.Op, entityW, actionW, targetW int) string {
	ts := op.Timestamp
	if len(ts) >= 19 {
		ts = strings.Replace(ts[:19], "T", " ", 1)
	}
	target := op.EntityID
	if len(target) > 8 {
		target = target[:8]
	}
	row := fmt.Sprintf("%-2s %-19s %-*s %-*s %s", "", ts, entityW, viewhelpers.Truncate(string(op.Entity), entityW), actionW, viewhelpers.Truncate(string(op.Action), actionW), viewhelpers.Truncate(target, targetW))
	switch {
	case i == cur && active:
		return theme.StyleCursor.Render("▶ " + viewhelpers.Truncate(row[2:], state.Width-6))
	case i == cur:
		return theme.StyleSelected.Render("  " + viewhelpers.Truncate(row[2:], state.Width-6))
	default:
		return theme.StyleNormal.Render("  " + viewhelpers.Truncate(row[2:], state.Width-6))
	}
}

func filteredOpIndices(ops []api.Op, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, op := range ops {
		text := strings.ToLower(string(op.Entity) + " " + string(op.Action) + " " + op.EntityID)
		if filter == "" || strings.Contains(text, filter) {
			out = append(out, i)
		}
	}
	return out
}

func reverseIndices(in []int) []int {
	out := make([]int, len(in))
	for i := range in {
		out[i] = in[len(in)-1-i]
	}
	return out
}

func currentOpsLimit(state types.ContentState) int {
	visibleRows := state.Height - 6
	if visibleRows < 10 {
		visibleRows = 10
	}
	return visibleRows
}

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
