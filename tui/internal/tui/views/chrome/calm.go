package viewchrome

import (
	"strings"

	helperpkg "crona/tui/internal/tui/helpers"

	"github.com/charmbracelet/lipgloss"
)

// SettingLine is a settings-style row: the label dim in its column, then the value styled
// by CalmValue.
func SettingLine(theme Theme, label, value string, column int) string {
	return theme.StyleDim.Render(LabelCell(label, column)) + "  " + CalmValue(theme, value)
}

// CalmValue colours a setting's value by what it means: on as a soft green dot, off and
// defaults dim, paths shortened, and • joins softened to a dim ·.
func CalmValue(theme Theme, value string) string {
	trimmed := strings.TrimSpace(value)
	lower := strings.ToLower(trimmed)
	green := lipgloss.NewStyle().Foreground(theme.ColorDullGreen)
	switch {
	case lower == "enabled" || lower == "on":
		return green.Render("● on")
	case lower == "disabled" || lower == "off":
		return theme.StyleDim.Render("○ off")
	case lower == "" || lower == "-" || lower == "–" || lower == "default" || lower == "none":
		return theme.StyleDim.Render(orDashValue(trimmed))
	case strings.HasPrefix(lower, "enabled ·") || strings.HasPrefix(lower, "enabled •"):
		return green.Render("● on") + calmJoin(theme, trimmed[len("enabled"):])
	case strings.HasPrefix(lower, "disabled ·") || strings.HasPrefix(lower, "disabled •"):
		return theme.StyleDim.Render("○ off") + calmJoin(theme, trimmed[len("disabled"):])
	case strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "~/"):
		return theme.StyleNormal.Render(helperpkg.ShortPath(trimmed, 40))
	case lower == "destructive":
		return lipgloss.NewStyle().Foreground(theme.ColorDullRed).Render(trimmed)
	}
	return calmJoin(theme, trimmed)
}

// calmJoin renders text with its • and · separators dim.
func calmJoin(theme Theme, text string) string {
	text = strings.ReplaceAll(text, " • ", " · ")
	parts := strings.Split(text, " · ")
	for i, part := range parts {
		parts[i] = theme.StyleNormal.Render(part)
	}
	return strings.Join(parts, theme.StyleDim.Render(" · "))
}

func orDashValue(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
