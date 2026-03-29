package views

import "github.com/charmbracelet/lipgloss"

func newStyle(color interface{}) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if c, ok := color.(lipgloss.Color); ok {
		style = style.Foreground(c)
	}
	return style
}
