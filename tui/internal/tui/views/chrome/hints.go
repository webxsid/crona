package viewchrome

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// KeyHint is one key hint: the key in brackets, then its label.
func KeyHint(theme Theme, key, label string) string {
	return theme.StyleHeader.Render("["+key+"]") + theme.StyleDim.Render(" "+label)
}

// FitHints joins hints two spaces apart, keeping as many whole hints as fit in width.
func FitHints(hints []string, width int) string {
	out := ""
	for _, hint := range hints {
		next := hint
		if out != "" {
			next = out + "  " + hint
		}
		if ansi.StringWidth(next) > width {
			break
		}
		out = next
	}
	return strings.TrimRight(out, " ")
}
