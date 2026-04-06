package viewchrome

import "strings"

func ClipSummaryLines(theme Theme, lines []string, height int) []string {
	maxLines := height - 4
	flattened := make([]string, 0, len(lines))
	for _, line := range lines {
		flattened = append(flattened, strings.Split(line, "\n")...)
	}
	if maxLines < 1 || len(flattened) <= maxLines {
		return flattened
	}
	if maxLines == 1 {
		return []string{theme.StyleDim.Render("...")}
	}
	clipped := append([]string{}, flattened[:maxLines-1]...)
	return append(clipped, theme.StyleDim.Render("..."))
}
