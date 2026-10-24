package tags

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	issuecore "crona/tui/internal/tui/views/issuecore"
	tagsmeta "crona/tui/internal/tui/views/tagsmeta"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"

	"github.com/charmbracelet/x/ansi"
)

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("tags", true, renderView),
		},
	}
}

// labelWidth caps the tag-name column; names are at most 32 characters plus the +.
const labelWidth = 34

// paneWidths keeps the list just wide enough for its columns and gives the rest to the detail.
func paneWidths(width int) (list, detail int) {
	if width < 100 {
		return width, 0
	}
	list = min(72, max(64, width*2/5))
	return list, width - list - 3
}

// renderView lists the tag vocabulary with its usage beside the selected tag's issues and habits.
func renderView(theme types.Theme, state types.ContentState) string {
	listWidth, detailWidth := paneWidths(state.Width)
	list := renderList(theme, state, listWidth)
	if detailWidth == 0 {
		return list
	}
	return viewchrome.JoinColumns(theme, state.Height, list, renderDetail(theme, state, detailWidth))
}

func renderList(theme types.Theme, state types.ContentState, width int) string {
	filter := state.Filters["tags"]
	indices := tagsmeta.FilteredIndices(state.Tags, filter)
	cursor := min(max(0, state.Cursors["tags"]), max(0, len(indices)-1))
	hints := viewchrome.FitHints([]string{
		viewchrome.KeyHint(theme, "enter", "items"),
		viewchrome.KeyHint(theme, "r", "rename"),
		viewchrome.KeyHint(theme, "←/→", "color"),
		viewchrome.KeyHint(theme, "D", "delete"),
		viewchrome.KeyHint(theme, "U", "unused"),
	}, width-2)
	body := []string{hints}
	if len(indices) == 0 {
		empty := "No tags yet. Press + on an issue or habit to add one."
		if filter != "" {
			empty = "No tags match the current filter"
		}
		body = append(body, theme.StyleDim.Render(empty))
	} else {
		labels := make([]string, 0, len(indices))
		for _, idx := range indices {
			labels = append(labels, "+"+state.Tags[idx].Name)
		}
		column := viewchrome.LabelColumn(labels, labelWidth)
		rows := make([]viewchrome.GroupRow, 0, len(indices))
		for _, idx := range indices {
			tag := state.Tags[idx]
			group := "Yours"
			if tag.System {
				group = "System"
			}
			open, closed := tagsmeta.OpenDone(tag.Name, state.AllIssues)
			pad := max(0, column-ansi.StringWidth("+"+tag.Name))
			rows = append(rows, viewchrome.GroupRow{
				Group: group,
				Line: issuecore.TagChip(theme, tag.Name, state.Tags) + strings.Repeat(" ", pad) +
					fmt.Sprintf("  %3d open  %3d done  %3d habits", open, closed, tag.HabitCount),
			})
		}
		body = append(body, viewchrome.GroupList(theme, rows, cursor, width-1, state.Height-4)...)
	}
	issues, habits := tagsmeta.Untagged(state.AllIssues, state.AllHabits)
	for len(body) < state.Height-2 {
		body = append(body, "")
	}
	body = append(body, theme.StyleDim.Render(ansi.Truncate(
		fmt.Sprintf("Untagged: %d open issues, %d active habits", issues, habits), width-2, "…")))
	meta := fmt.Sprint(len(indices))
	if unused := tagsmeta.Unused(state.Tags); unused > 0 {
		meta += fmt.Sprintf("  %d unused", unused)
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Tags", Meta: viewhelpers.FilterMeta(meta, filter), Focused: true,
		Width: width, Height: state.Height, Lines: body,
	})
}

func renderDetail(theme types.Theme, state types.ContentState, width int) string {
	tag, ok := tagsmeta.Selected(state.Tags, state.Filters["tags"], state.Cursors["tags"])
	if !ok {
		return viewchrome.RenderSection(theme, viewchrome.Section{Title: "Detail", Width: width, Height: state.Height})
	}
	room := max(1, width-2)
	open, closed := tagsmeta.OpenDone(tag.Name, state.AllIssues)
	color := "no color"
	if tag.Color != nil && *tag.Color != "" {
		color = *tag.Color
	}
	kind := "your tag"
	if tag.System {
		kind = "system tag, color only"
	}
	lines := []string{
		fmt.Sprintf("%d open · %d done · %d habits", open, closed, tag.HabitCount),
		theme.StyleDim.Render(color + " · " + kind),
	}
	issues := tagsmeta.TaggedIssues(tag.Name, state.AllIssues)
	habits := tagsmeta.TaggedHabits(tag.Name, state.AllHabits)
	if len(issues) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Issues"))
		for _, issue := range issues {
			style := theme.StyleNormal
			if statusStyle := issuecore.IssueStatusStyle(theme, string(issue.Status)); statusStyle != nil {
				style = *statusStyle
			}
			meta := "  " + issuecore.PlainIssueStatus(string(issue.Status)) + " · " +
				issuecore.IssueContextLabel(issue.RepoName, issue.StreamName)
			title := ansi.Truncate(issue.Title, max(1, room-min(ansi.StringWidth(meta), room/2)), "…")
			lines = append(lines, style.Render(title)+theme.StyleDim.Render(ansi.Truncate(meta, max(0, room-ansi.StringWidth(title)), "…")))
		}
	}
	if len(habits) > 0 {
		lines = append(lines, "", theme.StyleHeader.Render("Habits"))
		for _, habit := range habits {
			meta := "  " + issuecore.IssueContextLabel(habit.RepoName, habit.StreamName)
			if !habit.Active {
				meta = "  paused ·" + meta
			}
			name := ansi.Truncate(habit.Name, max(1, room-min(ansi.StringWidth(meta), room/2)), "…")
			lines = append(lines, name+theme.StyleDim.Render(ansi.Truncate(meta, max(0, room-ansi.StringWidth(name)), "…")))
		}
	}
	if len(issues)+len(habits) == 0 {
		lines = append(lines, "", theme.StyleDim.Render("Nothing carries this tag yet."))
	}
	if limit := max(1, state.Height-2); len(lines) > limit {
		hidden := len(lines) - limit + 1
		lines = append(lines[:limit-1], theme.StyleDim.Render(fmt.Sprintf("… %d more", hidden)))
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "+" + tag.Name, Width: width, Height: state.Height, Lines: lines,
	})
}
