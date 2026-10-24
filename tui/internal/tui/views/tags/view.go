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

// renderView lists the tag vocabulary with usage beside the selected tag's detail.
func renderView(theme types.Theme, state types.ContentState) string {
	filter := state.Filters["tags"]
	indices := tagsmeta.FilteredIndices(state.Tags, filter)
	cursor := min(max(0, state.Cursors["tags"]), max(0, len(indices)-1))
	listWidth, detailWidth := viewchrome.LogPaneWidths(state.Width)

	var body []string
	if len(indices) == 0 {
		empty := "No tags yet. Press + on an issue or habit to add one."
		if filter != "" {
			empty = "No tags match the current filter"
		}
		body = []string{theme.StyleDim.Render(empty)}
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
			chip := issuecore.TagChip(theme, tag.Name, state.Tags)
			pad := max(0, column-ansi.StringWidth("+"+tag.Name))
			rows = append(rows, viewchrome.GroupRow{
				Group: group,
				Line:  chip + strings.Repeat(" ", pad) + "  " + tagsmeta.Usage(tag),
			})
		}
		body = viewchrome.GroupList(theme, rows, cursor, listWidth-1, state.Height-1)
	}
	meta := fmt.Sprint(len(indices))
	if unused := tagsmeta.Unused(state.Tags); unused > 0 {
		meta += fmt.Sprintf("  %d unused", unused)
	}
	list := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Tags", Meta: viewhelpers.FilterMeta(meta, filter), Focused: true,
		Width: listWidth, Height: state.Height, Lines: body,
	})
	if detailWidth == 0 {
		return list
	}
	title, lines := "Detail", []string{}
	if tag, ok := tagsmeta.Selected(state.Tags, filter, cursor); ok {
		title = "+" + tag.Name
		help := viewchrome.RowHelp{
			Description: "Used on " + tagsmeta.Usage(tag) + ". Renaming onto an existing tag merges the two.",
			Keys:        []string{"[enter] rename", "[←/→] color", "[D] delete", "[U] remove unused"},
		}
		if tag.System {
			help = viewchrome.RowHelp{
				Description: "A system tag, used on " + tagsmeta.Usage(tag) + ". It can't be renamed or deleted; only its color changes.",
				Keys:        []string{"[←/→] color", "[U] remove unused"},
			}
		}
		color := "none"
		if tag.Color != nil && *tag.Color != "" {
			color = *tag.Color
		}
		lines = viewchrome.HelpLines(theme, help, "color "+color, detailWidth-2)
	}
	detail := viewchrome.RenderSection(theme, viewchrome.Section{
		Title: title, Width: detailWidth, Height: state.Height, Lines: lines,
	})
	return viewchrome.JoinColumns(theme, state.Height, list, detail)
}

// labelWidth caps the tag-name column; names are at most 32 characters plus the +.
const labelWidth = 34
