package viewhelpers

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

type GradientBarPalette struct {
	Start lipgloss.Color
	End   lipgloss.Color
	Track lipgloss.Color
}

func GradientRamp(start, end lipgloss.Color, steps int) []lipgloss.Color {
	if steps <= 0 {
		return nil
	}
	if steps == 1 {
		return []lipgloss.Color{start}
	}

	startColor := colorfulFromLipgloss(start)
	endColor := colorfulFromLipgloss(end)
	ramp := make([]lipgloss.Color, steps)
	for idx := range steps {
		t := float64(idx) / float64(steps-1)
		blended := startColor.BlendOkLab(endColor, t).Clamped()
		ramp[idx] = lipgloss.Color(blended.Hex())
	}
	return ramp
}

func GradientColorAt(ramp []lipgloss.Color, idx int) lipgloss.Color {
	if len(ramp) == 0 {
		return lipgloss.Color("0")
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(ramp) {
		idx = len(ramp) - 1
	}
	return ramp[idx]
}

func RenderGradientBar(width, filled int, palette GradientBarPalette) string {
	return RenderGradientBarWithMarker(width, filled, -1, palette, "┆")
}

func RenderGradientBarWithMarker(
	width, filled, markerPos int,
	palette GradientBarPalette,
	markerRune string,
) string {
	if width <= 0 {
		return ""
	}
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	ramp := GradientRamp(palette.Start, palette.End, width)
	var builder strings.Builder
	for idx := range width {
		switch {
		case idx == markerPos && markerPos >= 0:
			builder.WriteString(
				lipgloss.NewStyle().
					Background(lipgloss.Color("0")).
					Foreground(lipgloss.Color("208")).
					Bold(true).
					Render(markerRune),
			)
		case idx < filled:
			builder.WriteString(
				lipgloss.NewStyle().
					Foreground(GradientColorAt(ramp, idx)).
					Render("█"),
			)
		default:
			builder.WriteString(
				lipgloss.NewStyle().
					Foreground(palette.Track).
					Render("░"),
			)
		}
	}
	return builder.String()
}

func colorfulFromLipgloss(color lipgloss.Color) colorful.Color {
	if hex := ansiColorHex(string(color)); hex != "" {
		if parsed, err := colorful.Hex(hex); err == nil {
			return parsed
		}
	}
	if parsed, err := colorful.Hex(string(color)); err == nil {
		return parsed
	}
	return colorful.Color{}
}

func ansiColorHex(code string) string {
	switch strings.TrimSpace(code) {
	case "0":
		return "#111111"
	case "1":
		return "#8d5a5a"
	case "2":
		return "#4f7d5d"
	case "3":
		return "#9c8a4f"
	case "4":
		return "#5b6f99"
	case "5":
		return "#9b6aa8"
	case "6":
		return "#4f8f90"
	case "7":
		return "#bdbdbd"
	case "8":
		return "#676767"
	case "9":
		return "#b55b5b"
	case "10":
		return "#6fae74"
	case "11":
		return "#c3a34a"
	case "12":
		return "#5d7db2"
	case "13":
		return "#b07ac4"
	case "14":
		return "#58b0b8"
	case "15":
		return "#f0f0f0"
	case "208":
		return "#d48a45"
	default:
		return ""
	}
}
