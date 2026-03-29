package views

import (
	"fmt"
	"strings"
)

func renderOpsView(theme Theme, state ContentState) string {
	active := state.Pane == "ops"
	cur := state.Cursors["ops"]
	indices := reverseIndices(filteredOpIndices(state.Ops, state.Filters["ops"]))
	total := len(indices)
	actionLine := renderPaneActionLine(theme, state.Filters["ops"], state.Width-6, paneActionsForState(theme, state, active))
	lines := []string{theme.StylePaneTitle.Render("Ops Log"), theme.StyleDim.Render(fmt.Sprintf("limit: %d", currentOpsLimit(state))), actionLine}
	if total == 0 {
		lines = append(lines, theme.StyleDim.Render("No operations recorded"))
	} else {
		timeW, entityW, actionW := 19, max(10, state.Width/8), 10
		targetW := state.Width - timeW - entityW - actionW - 12
		if targetW < 12 {
			targetW = 12
		}
		header := fmt.Sprintf("%-2s %-19s %-*s %-*s %s", "", "Time", entityW, "Entity", actionW, "Action", "Target")
		lines = append(lines, theme.StyleDim.Render(truncate(header, state.Width-6)))
		inner := remainingPaneHeight(state.Height, lines)
		start, end := listWindow(cur, total, inner)
		if start > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			op := state.Ops[indices[i]]
			ts := op.Timestamp
			if len(ts) >= 19 {
				ts = strings.Replace(ts[:19], "T", " ", 1)
			}
			target := op.EntityID
			if len(target) > 8 {
				target = target[:8]
			}
			row := fmt.Sprintf("%-2s %-19s %-*s %-*s %s", "", ts, entityW, truncate(string(op.Entity), entityW), actionW, truncate(string(op.Action), actionW), truncate(target, targetW))
			if i == cur && active {
				lines = append(lines, theme.StyleCursor.Render("▶ "+truncate(row[2:], state.Width-6)))
			} else if i == cur {
				lines = append(lines, theme.StyleSelected.Render("  "+truncate(row[2:], state.Width-6)))
			} else {
				lines = append(lines, theme.StyleNormal.Render("  "+truncate(row[2:], state.Width-6)))
			}
		}
		if remaining := total - end; remaining > 0 {
			lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("   ↓ %d more", remaining)))
		}
	}
	return renderPaneBox(theme, active, state.Width, state.Height, stringsJoin(lines))
}
