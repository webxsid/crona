package issuecore

import (
	"fmt"

	"crona/tui/internal/api"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// TagColor is the palette color for a tag color name, and false for none or an unknown name.
func TagColor(theme Theme, name string) (lipgloss.Color, bool) {
	switch name {
	case "blue":
		return theme.ColorBlue, true
	case "cyan":
		return theme.ColorCyan, true
	case "green":
		return theme.ColorGreen, true
	case "magenta":
		return theme.ColorMagenta, true
	case "yellow":
		return theme.ColorYellow, true
	case "red":
		return theme.ColorRed, true
	case "orange":
		return theme.ColorOrange, true
	case "lavender":
		return theme.ColorLavender, true
	default:
		return "", false
	}
}

// TagChip renders one tag as "+name" in its color, or muted when it has none.
func TagChip(theme Theme, name string, vocab []api.Tag) string {
	style := theme.StyleDim
	for _, tag := range vocab {
		if tag.Name == name && tag.Color != nil {
			if color, ok := TagColor(theme, *tag.Color); ok {
				style = lipgloss.NewStyle().Foreground(color)
			}
			break
		}
	}
	return style.Render("+" + name)
}

// TagChips renders tags as chips within width cells. When they do not all fit it shows as many
// as it can and a "+N" count of the rest; with no room for even that it is "".
func TagChips(theme Theme, tags []string, vocab []api.Tag, width int) string {
	out := ""
	for i, name := range tags {
		chip := TagChip(theme, name, vocab)
		next := chip
		if out != "" {
			next = out + " " + chip
		}
		rest := ""
		if i < len(tags)-1 {
			rest = fmt.Sprintf(" +%d", len(tags)-1-i)
		}
		if ansi.StringWidth(next)+len(rest) > width {
			more := theme.StyleDim.Render(fmt.Sprintf("+%d", len(tags)-i))
			if out == "" {
				if ansi.StringWidth(more) > width {
					return ""
				}
				return more
			}
			if ansi.StringWidth(out)+1+ansi.StringWidth(more) > width {
				return out
			}
			return out + " " + more
		}
		out = next
	}
	return out
}
