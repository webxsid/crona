package viewhelpers

import "crona/tui/internal/api"

func ScratchpadItems(in []api.ScratchPad) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.Name)
	}
	return out
}

func FilteredStrings(items []string, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, item := range items {
		if filter == "" || containsFold(item, filter) {
			out = append(out, i)
		}
	}
	return out
}
