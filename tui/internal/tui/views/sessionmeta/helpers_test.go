package sessionmeta

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderResponsiveClockSelectsSizeByPaneSpace(t *testing.T) {
	cases := []struct {
		name   string
		width  int
		height int
		want   string
	}{
		{name: "tiny", width: 8, height: 2, want: RenderClockTiny("00:12:34")},
		{name: "small", width: 47, height: 7, want: RenderClockSmall("00:12:34")},
		{name: "medium", width: 79, height: 11, want: RenderClockMedium("00:12:34")},
		{name: "large", width: 111, height: 15, want: RenderClockLarge("00:12:34")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderResponsiveClock("00:12:34", tc.width, tc.height, lipgloss.Color(""), lipgloss.Color(""))
			if got != tc.want {
				t.Fatalf("expected responsive clock to select %q, got %q", tc.want, got)
			}
		})
	}
}

func TestClockRenderVariantsHaveExpectedShape(t *testing.T) {
	if got := RenderClockTiny("00:12:34"); got != "00:12:34" {
		t.Fatalf("expected tiny clock to be plain text, got %q", got)
	}
	if got := RenderClockSmall("00:12:34"); countLines(got) != 5 || !strings.Contains(got, "█") || strings.Contains(got, "░") {
		t.Fatalf("expected small clock to render a 5-line tile display, got %q", got)
	}
	if got := RenderClockMedium("00:12:34"); countLines(got) != 10 || !strings.Contains(got, "█") || strings.Contains(got, "░") {
		t.Fatalf("expected medium clock to render a 10-line tile display, got %q", got)
	}
	if got := RenderClockLarge("00:12:34"); countLines(got) != 15 || !strings.Contains(got, "█") || strings.Contains(got, "░") {
		t.Fatalf("expected large clock to render a 15-line tile display, got %q", got)
	}
}

func countLines(value string) int {
	if value == "" {
		return 0
	}
	lines := 1
	for _, r := range value {
		if r == '\n' {
			lines++
		}
	}
	return lines
}
