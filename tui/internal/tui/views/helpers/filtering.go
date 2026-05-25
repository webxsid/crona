package viewhelpers

import "strings"

func normalizeFilter(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
