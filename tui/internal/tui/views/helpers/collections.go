package viewhelpers

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
