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

func FormatClockText(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func StringsJoin(rows []string) string {
	return strings.Join(rows, "\n")
}
