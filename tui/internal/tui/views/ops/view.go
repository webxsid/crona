package ops

import (
	"strings"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
	viewui "crona/tui/internal/tui/views/ui"
)

func newView() viewui.View {
	return viewui.View{
		Panes: []viewui.Pane{
			viewui.NewRenderPane("ops", true, renderView),
		},
		RenderFunc: renderView,
	}
}

func renderView(theme types.Theme, state types.ContentState) string {
	return renderLog(theme, state)
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

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
