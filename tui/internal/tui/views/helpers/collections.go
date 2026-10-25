package viewhelpers

import "strings"

func FilteredStrings(items []string, filter string) []int {
	filter = normalizeFilter(filter)
	out := []int{}
	for i, item := range items {
		if filter == "" || strings.Contains(strings.ToLower(item), filter) {
			out = append(out, i)
		}
	}
	return out
}

// FilterMeta adds a pane's active filter to its header meta, as "/filter".
func FilterMeta(meta, filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return meta
	}
	if len([]rune(filter)) > 28 {
		filter = string([]rune(filter)[:27]) + "…"
	}
	if meta == "" {
		return "/" + filter
	}
	return meta + "  /" + filter
}
