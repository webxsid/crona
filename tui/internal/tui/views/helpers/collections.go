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
