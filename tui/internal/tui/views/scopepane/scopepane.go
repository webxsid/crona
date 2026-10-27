// Package scopepane renders the Issues tab's scope tree: All, repos and their streams.
package scopepane

import (
	"fmt"
	"strings"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/scope"
	viewchrome "crona/tui/internal/tui/views/chrome"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// Render is the tree as a section: one row per node with its count and a ● on the
// checked-out repo and stream, the focused node's keys under it, and a create line.
func Render(
	theme types.Theme,
	state types.ContentState,
	counts map[scope.Scope]int,
	width, height int,
) string {
	focused := state.Pane == "scope"
	nodes := scope.Nodes(state.Repos, state.AllStreams)
	cur := min(max(0, state.Cursors["scope"]), len(nodes)-1)
	checked := scope.FromContext(state.Context)
	rowWidth := max(1, width-2)

	budget := max(1, height-2) // header and create line
	if focused && !nodes[cur].Scope.IsAll() {
		budget--
	}
	start, end := viewchrome.ListWindow(cur, len(nodes), budget)
	if start > 0 || end < len(nodes) {
		start, end = viewchrome.ListWindow(cur, len(nodes), max(1, budget-2))
	}
	lines := []string{}
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		lines = append(
			lines,
			nodeRow(
				theme,
				nodes[i],
				counts[nodes[i].Scope],
				isChecked(nodes[i], checked),
				focused && i == cur,
				i == cur,
				rowWidth,
			),
		)
		if focused && i == cur && !nodes[i].Scope.IsAll() {
			lines = append(lines, "    "+viewchrome.FitHints(nodeActions(theme, state), rowWidth-4))
		}
	}
	if end < len(nodes) {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", len(nodes)-end)))
	}
	if focused {
		lines = append(lines, createLine(theme, nodes[cur]))
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Scope", Focused: focused, Width: width, Height: height, Lines: lines,
	})
}

// RenderRow is the tree collapsed to one row for narrow screens: the browsed node between
// ‹ › markers, stepped with ←/→ while focused.
func RenderRow(theme types.Theme, state types.ContentState, width int) string {
	nodes := scope.Nodes(state.Repos, state.AllStreams)
	cur := min(max(0, state.Cursors["scope"]), len(nodes)-1)
	node := nodes[cur]
	path := node.Name
	if node.Scope.StreamID != 0 {
		path = repoName(state.Repos, node.Scope.RepoID) + " › " + node.Name
	}
	label := theme.StyleDim.Render("Scope")
	if state.Pane == "scope" {
		label = theme.StylePaneTitle.Render("Scope")
	}
	row := " " + label + " " + theme.StyleHeader.Render(
		"‹",
	) + " " + theme.StyleNormal.Render(
		path,
	) + " " + theme.StyleHeader.Render(
		"›",
	)
	if isChecked(node, scope.FromContext(state.Context)) {
		row += " " + theme.StyleHeader.Render("●")
	}
	if ansi.StringWidth(row) > width {
		return ansi.Truncate(row, width, "…")
	}
	return row + strings.Repeat(" ", width-ansi.StringWidth(row))
}

// IssueCounts counts open issues under All, each repo and each stream.
func IssueCounts(issues []api.IssueWithMeta) map[scope.Scope]int {
	counts := map[scope.Scope]int{}
	for _, issue := range issues {
		if issue.Status == "done" || issue.Status == "abandoned" {
			continue
		}
		add(counts, issue.RepoID, issue.StreamID)
	}
	return counts
}

// HabitCounts counts habit definitions under All, each repo and each stream.
func HabitCounts(habits []api.HabitWithMeta) map[scope.Scope]int {
	counts := map[scope.Scope]int{}
	for _, habit := range habits {
		add(counts, habit.RepoID, habit.StreamID)
	}
	return counts
}

func add(counts map[scope.Scope]int, repoID, streamID int64) {
	counts[scope.Scope{}]++
	counts[scope.Scope{RepoID: repoID}]++
	counts[scope.Scope{RepoID: repoID, StreamID: streamID}]++
}

func nodeRow(
	theme types.Theme,
	node scope.Node,
	count int,
	checked, cursor, selected bool,
	width int,
) string {
	marker := "  "
	if cursor {
		marker = viewchrome.SelectionCursor + " "
	}
	name := marker + strings.Repeat("  ", node.Depth) + node.Name
	dot := "  "
	if checked {
		dot = " " + theme.StyleHeader.Render("●")
	}
	right := theme.StyleDim.Render(fmt.Sprintf("%d", count)) + dot
	if node.System {
		right = theme.StyleDim.Render("system  ") + right
	}
	room := width - ansi.StringWidth(right) - 1
	name = ansi.Truncate(name, max(1, room), "…")
	style := theme.StyleNormal
	switch {
	case cursor:
		style = theme.StyleCursor
	case selected:
		style = theme.StyleSelected
	}
	return style.Render(
		name,
	) + strings.Repeat(
		" ",
		max(1, width-ansi.StringWidth(name)-ansi.StringWidth(right)),
	) + right
}

func nodeActions(theme types.Theme, state types.ContentState) []string {
	actions := []string{
		viewchrome.KeyHint(theme, "c", "check out"),
		viewchrome.KeyHint(theme, "e", "edit"),
	}
	if state.Timer == nil || state.Timer.State == "idle" {
		actions = append(actions, viewchrome.KeyHint(theme, "D", "delete"))
	}
	return append(actions, viewchrome.KeyHint(theme, "enter", "open"))
}

func createLine(theme types.Theme, node scope.Node) string {
	if node.Scope.IsAll() {
		return viewchrome.KeyHint(theme, "a", "new repo")
	}
	return viewchrome.KeyHint(theme, "a", "new stream")
}

func isChecked(node scope.Node, checked scope.Scope) bool {
	switch {
	case node.Scope.IsAll() || checked.IsAll():
		return false
	case node.Scope.StreamID != 0:
		return node.Scope.StreamID == checked.StreamID
	default:
		return node.Scope.RepoID == checked.RepoID
	}
}

func repoName(repos []api.Repo, id int64) string {
	for _, repo := range repos {
		if repo.ID == id {
			return repo.Name
		}
	}
	return "-"
}
