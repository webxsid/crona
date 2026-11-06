package viewchrome

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// CenterLines centres lines in a width by height box, each line on its own and the block
// a little above the middle.
func CenterLines(lines []string, width, height int) []string {
	out := make([]string, 0, height)
	top := max(0, (height-len(lines))*2/5)
	for range top {
		out = append(out, "")
	}
	for _, line := range lines {
		line = ansi.Truncate(line, width, "…")
		out = append(out, strings.Repeat(" ", max(0, (width-ansi.StringWidth(line))/2))+line)
	}
	return out
}
