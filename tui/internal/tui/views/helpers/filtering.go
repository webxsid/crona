package viewhelpers

import "strings"

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func containsFold(value, filter string) bool {
	return strings.Contains(strings.ToLower(value), filter)
}
