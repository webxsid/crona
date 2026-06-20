package chrome

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBlue      = lipgloss.Color("12")
	ColorCyan      = lipgloss.Color("14")
	ColorGreen     = lipgloss.Color("10")
	ColorDullGreen = lipgloss.Color("2") // A duller green for less emphasis
	ColorMagenta   = lipgloss.Color("13")
	ColorSubtle    = lipgloss.Color("7")
	ColorYellow    = lipgloss.Color("11")
	ColorRed       = lipgloss.Color("9")
	ColorDullRed   = lipgloss.Color("1")   // A duller red for less emphasis
	ColorOrange    = lipgloss.Color("208") // A bright orange for warnings or highlights
	ColorDim       = lipgloss.Color("8")
	ColorWhite     = lipgloss.Color("15")
	ColorBlack     = lipgloss.Color("0")

	StyleActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan)

	StyleInactive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim)

	StylePaneTitle       = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	StyleDim             = lipgloss.NewStyle().Foreground(ColorDim)
	StyleCursor          = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	StyleHeader          = lipgloss.NewStyle().Foreground(ColorCyan)
	StyleError           = lipgloss.NewStyle().Foreground(ColorRed)
	StyleSelected        = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleSelectedInverse = lipgloss.NewStyle().
				Background(ColorGreen).
				Foreground(ColorBlack).
				Padding(0, 1)
	StyleNormal = lipgloss.NewStyle()
)
