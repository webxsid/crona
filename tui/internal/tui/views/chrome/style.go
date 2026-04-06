package viewchrome

import "github.com/charmbracelet/lipgloss"

func newStyle(color interface{}) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if c, ok := color.(lipgloss.Color); ok {
		style = style.Foreground(c)
	}
	return style
}

func LipStyle(theme Theme, color interface{}) styleLike {
	return styleLike{theme: theme, color: color}
}

type styleLike struct {
	theme Theme
	color interface{}
}

func (s styleLike) Render(text string) string {
	return newStyle(s.color).Render(text)
}
