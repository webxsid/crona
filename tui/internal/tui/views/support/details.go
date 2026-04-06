package support

import "strings"

func supportFallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return strings.TrimSpace(value)
}
