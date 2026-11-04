package viewchrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func sectionTheme() Theme {
	return Theme{
		StylePaneTitle: lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		StyleDim:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

func TestRenderSectionHeaderFormat(t *testing.T) {
	out := RenderSection(sectionTheme(), Section{Title: "Tasks", Meta: "3/5", Width: 40, Height: 3})
	header := ansi.Strip(strings.Split(out, "\n")[0])
	if !strings.HasPrefix(header, "─ Tasks ") || !strings.HasSuffix(header, " 3/5") {
		t.Fatalf("unexpected header %q", header)
	}
	if got := ansi.StringWidth(header); got != 40 {
		t.Fatalf("header width = %d, want 40", got)
	}
}

func TestRenderSectionExactSize(t *testing.T) {
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "row"
	}
	out := RenderSection(sectionTheme(), Section{Title: "List", Width: 30, Height: 4, Lines: lines})
	rows := strings.Split(out, "\n")
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}
	for i, row := range rows {
		if got := ansi.StringWidth(row); got != 30 {
			t.Fatalf("row %d width = %d, want 30", i, got)
		}
	}
	if !strings.Contains(rows[3], "↓ more") {
		t.Fatalf("expected clipped section to end with ↓ more, got %q", rows[3])
	}
}

func TestRenderSectionFocusedTitleStyle(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(previousProfile)
	base := Section{Title: "Tasks", Width: 30, Height: 2}
	focused := base
	focused.Focused = true
	a := strings.Split(RenderSection(sectionTheme(), base), "\n")[0]
	b := strings.Split(RenderSection(sectionTheme(), focused), "\n")[0]
	if a == b {
		t.Fatal("expected focused title to render differently")
	}
}

func TestJoinColumnsGutter(t *testing.T) {
	left := RenderSection(sectionTheme(), Section{Title: "A", Width: 10, Height: 3})
	right := RenderSection(sectionTheme(), Section{Title: "B", Width: 10, Height: 3})
	rows := strings.Split(JoinColumns(sectionTheme(), 3, left, right), "\n")
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	for i, row := range rows {
		if got := ansi.StringWidth(row); got != 23 || !strings.Contains(row, "│") {
			t.Fatalf("row %d = %q (width %d)", i, row, got)
		}
	}
}

func TestRenderSectionFadesOnlyDimmedBodies(t *testing.T) {
	theme := sectionTheme()
	line := "\x1b[92mready\x1b[0m done"
	dimmed := RenderSection(theme, Section{Title: "Issues", Dimmed: true, Width: 30, Height: 3, Lines: []string{line}})
	body := strings.Split(dimmed, "\n")[1]
	if !strings.HasPrefix(body, "\x1b[2m") || !strings.Contains(body, "\x1b[0m\x1b[2m done") {
		t.Fatalf("a dimmed body should stay faint across inner resets, got %q", body)
	}
	focused := RenderSection(theme, Section{Title: "Issues", Focused: true, Width: 30, Height: 3, Lines: []string{line}})
	if strings.Contains(focused, "\x1b[2m") {
		t.Fatalf("a focused body should not be faint, got %q", focused)
	}
}
