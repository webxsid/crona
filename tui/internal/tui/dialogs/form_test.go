package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestFormBodyFocusLine(t *testing.T) {
	rows := []FormRow{staticRow("a"), fieldRow(0, "x"), pairRow("l", "r", 1, 2, 50)}
	for focus, want := range map[int]int{0: 1, 2: 3, 9: -1} {
		if _, got := formBody(rows, focus); got != want {
			t.Errorf("focus %d: line %d, want %d", focus, got, want)
		}
	}
}

func TestFormFooterPaddedToTallest(t *testing.T) {
	hints := func(focus int) []string {
		if focus == 1 {
			return []string{"aaaa one", "bbbb two", "cccc three", "dddd four", "eeee five", "ffff six"}
		}
		return nil
	}
	zero := formFooter(testTheme(), "ctrl+s save", hints, 0, 2, 44)
	one := formFooter(testTheme(), "ctrl+s save", hints, 1, 2, 44)
	if len(zero) != len(one) || len(one) < 2 {
		t.Fatalf("footer heights differ: %d vs %d", len(zero), len(one))
	}
}

func TestFrameBorderColour(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)
	theme := testTheme()
	top := func(spec FrameSpec) string {
		spec.Title, spec.Body, spec.ScreenWidth, spec.ScreenHeight = "T", []string{"x"}, 120, 40
		return strings.Split(Frame(theme, spec), "\n")[0]
	}
	yellow := strings.SplitAfter(lipgloss.NewStyle().Foreground(theme.ColorYellow).Render("╭"), "╭")[0]
	if !strings.HasPrefix(top(FrameSpec{Border: theme.ColorYellow}), yellow) {
		t.Fatalf("expected a yellow border")
	}
	red := strings.SplitAfter(lipgloss.NewStyle().Foreground(theme.ColorRed).Render("╭"), "╭")[0]
	if !strings.HasPrefix(top(FrameSpec{Border: theme.ColorYellow, Danger: true}), red) {
		t.Fatalf("expected Danger to win")
	}
}
