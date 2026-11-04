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
	ColorLavender  = lipgloss.Color("141") // Carried-over issues
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

	// StyleTabActive is the active tab's pill; StyleTabIdle the label of the others.
	StyleTabActive = lipgloss.NewStyle().Bold(true).Foreground(ColorBlack).Background(ColorCyan)
	StyleTabIdle   = lipgloss.NewStyle().Foreground(ColorSubtle)
	// StyleSubTabActive marks the open sub-tab with an underline instead of a frame.
	StyleSubTabActive = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(ColorCyan)
	StyleDevBadge     = lipgloss.NewStyle().Bold(true).Foreground(ColorYellow)
)
