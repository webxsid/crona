package chrome

import "github.com/charmbracelet/lipgloss"

var (
	ColorBlue    = lipgloss.Color("12")
	ColorCyan    = lipgloss.Color("14")
	ColorGreen   = lipgloss.Color("10")
	ColorMagenta = lipgloss.Color("13")
	ColorSubtle  = lipgloss.Color("7")
	ColorYellow  = lipgloss.Color("11")
	ColorRed     = lipgloss.Color("9")
	ColorDim     = lipgloss.Color("8")
	ColorWhite   = lipgloss.Color("15")

	StyleActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan)

	StyleInactive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim)

	StylePaneTitle = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	StyleDim       = lipgloss.NewStyle().Foreground(ColorDim)
	StyleCursor    = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	StyleHeader    = lipgloss.NewStyle().Foreground(ColorCyan)
	StyleError     = lipgloss.NewStyle().Foreground(ColorRed)
	StyleSelected  = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleNormal    = lipgloss.NewStyle()
)
