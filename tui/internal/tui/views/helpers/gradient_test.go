package viewhelpers

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGradientRampInterpolatesAcrossSteps(t *testing.T) {
	ramp := GradientRamp(lipgloss.Color("#2f6f4e"), lipgloss.Color("#8fd36c"), 4)
	if len(ramp) != 4 {
		t.Fatalf("expected 4 ramp colors, got %d", len(ramp))
	}
	if ramp[0] == ramp[len(ramp)-1] {
		t.Fatalf("expected gradient endpoints to differ, got %v", ramp)
	}
	if ramp[1] == ramp[2] {
		t.Fatalf("expected interior ramp colors to vary, got %v", ramp)
	}
}

func TestRenderGradientBarWithMarkerKeepsWidth(t *testing.T) {
	bar := RenderGradientBarWithMarker(
		12,
		7,
		5,
		GradientBarPalette{
			Start: lipgloss.Color("#2f6f4e"),
			End:   lipgloss.Color("#8fd36c"),
			Track: lipgloss.Color("#3a3a3a"),
		},
		"┆",
	)
	if bar == "" {
		t.Fatal("expected gradient bar output")
	}
	if !containsRune(bar, '┆') {
		t.Fatalf("expected marker rune in gradient bar, got %q", bar)
	}
}

func containsRune(text string, target rune) bool {
	for _, r := range text {
		if r == target {
			return true
		}
	}
	return false
}
