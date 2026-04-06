package viewhelpers

import (
	"fmt"
	"strings"
)

func Truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func FormatClock(totalSeconds int) string {
	return fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
}

func StringsJoin(rows []string) string {
	return strings.Join(rows, "\n")
}
