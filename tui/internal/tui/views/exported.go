package views

import "github.com/charmbracelet/lipgloss"

func RenderDaily(theme Theme, state ContentState) string   { return renderDailyView(theme, state) }
func RenderDefault(theme Theme, state ContentState) string { return renderDefaultView(theme, state) }
func RenderWellbeing(theme Theme, state ContentState) string {
	return renderWellbeingView(theme, state)
}
func RenderReports(theme Theme, state ContentState) string  { return renderReportsView(theme, state) }
func RenderSettings(theme Theme, state ContentState) string { return renderSettingsView(theme, state) }
func RenderConfig(theme Theme, state ContentState) string   { return renderConfigView(theme, state) }
func RenderPaneBox(theme Theme, active bool, width, height int, content string) string {
	return renderPaneBox(theme, active, width, height, content)
}

func TestTheme() Theme {
	return Theme{
		ColorBlue:    lipgloss.Color("12"),
		ColorCyan:    lipgloss.Color("14"),
		ColorGreen:   lipgloss.Color("10"),
		ColorMagenta: lipgloss.Color("13"),
		ColorSubtle:  lipgloss.Color("7"),
		ColorYellow:  lipgloss.Color("11"),
		ColorRed:     lipgloss.Color("9"),
		ColorDim:     lipgloss.Color("8"),
		ColorWhite:   lipgloss.Color("15"),
		StyleActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("14")),
		StyleInactive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),
		StylePaneTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")),
		StyleDim:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		StyleCursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		StyleHeader:    lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		StyleError:     lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		StyleSelected:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		StyleNormal:    lipgloss.NewStyle(),
	}
}
