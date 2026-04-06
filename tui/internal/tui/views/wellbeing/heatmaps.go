package wellbeing

import (
	"strings"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
)

func Heatmap(size types.ViewSize, theme types.Theme, state types.ContentState) []string {
	switch size {
	case types.ViewSizeCompact:
		return wellbeingCompactHeatmap(theme, state)
	default:
		return wellbeingHeatmap(theme, state, state.Width)
	}
}

func wellbeingHeatmap(theme types.Theme, state types.ContentState, width int) []string {
	if len(state.MetricsRange) == 0 || width < 48 {
		return nil
	}
	rows := wellbeingHeatmapRows(state.MetricsRange)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, theme.StyleDim.Render(row))
	}
	return out
}

func wellbeingCompactHeatmap(theme types.Theme, state types.ContentState) []string {
	if len(state.MetricsRange) < 7 {
		return nil
	}
	glyphs := " .:-=+*#"
	window := state.MetricsRange
	if len(window) > 7 {
		window = window[len(window)-7:]
	}
	row := ""
	for _, day := range window {
		row += string(heatmapGlyph(glyphs, day)) + " "
	}
	row = strings.TrimSpace(row)
	if row == "" {
		return nil
	}
	return []string{
		theme.StyleDim.Render("Mon Tue Wed Thu Fri Sat Sun"),
		theme.StyleDim.Render(strings.ReplaceAll(row, " ", "   ")),
		theme.StyleDim.Render("low " + strings.TrimSpace(glyphs) + " high"),
	}
}

func wellbeingHeatmapRows(days []api.DailyMetricsDay) []string {
	const columns = 7
	glyphs := " .:-=+*#"
	rows := []string{"Mon Tue Wed Thu Fri Sat Sun"}
	line := ""
	weekdayCount := 0
	for _, day := range days {
		if weekdayCount == columns {
			rows = append(rows, strings.TrimSpace(line))
			line = ""
			weekdayCount = 0
		}
		line += string(heatmapGlyph(glyphs, day)) + "   "
		weekdayCount++
	}
	if strings.TrimSpace(line) != "" {
		rows = append(rows, strings.TrimSpace(line))
	}
	rows = append(rows, "Scale  low "+strings.TrimSpace(glyphs)+" high")
	return rows
}

func heatmapGlyph(glyphs string, day api.DailyMetricsDay) byte {
	score := 0
	if day.CheckIn != nil {
		score += 1
	}
	if day.SessionCount > 0 {
		score += min(4, day.SessionCount)
	}
	if day.WorkedSeconds >= 1800 {
		score += 1
	}
	if score < 0 {
		score = 0
	}
	if score >= len(glyphs) {
		score = len(glyphs) - 1
	}
	return glyphs[score]
}
