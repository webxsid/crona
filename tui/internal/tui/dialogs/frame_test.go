package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func plain(s string) []string { return strings.Split(ansi.Strip(s), "\n") }

func TestFrameTitleInBorder(t *testing.T) {
	lines := plain(Frame(Theme{}, FrameSpec{Title: "New Issue", Context: "OSS › cli", Body: []string{"x"}, ScreenWidth: 120, ScreenHeight: 40, Focus: -1}))
	if !strings.HasPrefix(lines[0], "╭─ New Issue ") || !strings.HasSuffix(lines[0], " OSS › cli ─╮") {
		t.Fatalf("expected title and context in the border, got %q", lines[0])
	}
}

func TestFrameContextDroppedWhenTight(t *testing.T) {
	title, context := strings.Repeat("T", 40), strings.Repeat("c", 30)
	top := plain(Frame(Theme{}, FrameSpec{Title: title, Context: context, Body: []string{"x"}, ScreenWidth: 60, ScreenHeight: 40, Focus: -1}))[0]
	if !strings.Contains(top, title) || strings.Contains(top, "ccc") {
		t.Fatalf("expected the context dropped, got %q", top)
	}
}

func TestFrameWidthClamp(t *testing.T) {
	width := func(screen int, body string) int {
		return ansi.StringWidth(plain(Frame(Theme{}, FrameSpec{Title: "T", Body: []string{body}, ScreenWidth: screen, ScreenHeight: 40, Focus: -1}))[0])
	}
	if got := width(200, "x"); got != 44+6 {
		t.Fatalf("expected the minimum width, got %d", got)
	}
	if got := width(200, strings.Repeat("word ", 60)); got != 76+6 {
		t.Fatalf("expected the maximum width, got %d", got)
	}
	if got := width(60, strings.Repeat("word ", 60)); got > 52 {
		t.Fatalf("expected at most screen-8, got %d", got)
	}
}

func TestFrameNoTrailingBlank(t *testing.T) {
	lines := plain(Frame(Theme{}, FrameSpec{Title: "T", Body: []string{"body"}, Footer: []string{"esc cancel"}, ScreenWidth: 120, ScreenHeight: 40, Focus: -1}))
	if !strings.Contains(lines[len(lines)-3], "esc cancel") || strings.TrimSpace(strings.Trim(lines[len(lines)-2], "│")) != "" {
		t.Fatalf("expected the footer, one padding line, then the border, got %q", lines[len(lines)-3:])
	}
}

func TestFrameKeepsFocusVisible(t *testing.T) {
	body := make([]string, 40)
	for i := range body {
		body[i] = "line " + string(rune('A'+i%26)) + strings.Repeat("x", i)
	}
	text := ansi.Strip(Frame(Theme{}, FrameSpec{Title: "T", Body: body, Footer: []string{"esc cancel"}, ScreenWidth: 120, ScreenHeight: 20, Focus: 30}))
	if !strings.Contains(text, body[30]) || !strings.Contains(text, "esc cancel") || !strings.Contains(text, "↑ more") {
		t.Fatalf("expected the focused line, footer and ↑ more, got\n%s", text)
	}
	if n := len(strings.Split(text, "\n")); n > 16 {
		t.Fatalf("expected the frame within the screen, got %d lines", n)
	}
}

func TestFieldPadsLabel(t *testing.T) {
	if got := ansi.Strip(Field(Theme{}, "Due", 8, "x", false)); !strings.HasPrefix(got, "Due"+strings.Repeat(" ", 7)+"x") {
		t.Fatalf("expected a padded label, got %q", got)
	}
}

func TestFieldPairFallsBack(t *testing.T) {
	if got := FieldPair("a", "b", 50); len(got) != 2 {
		t.Fatalf("expected two rows at 50, got %q", got)
	}
	if got := FieldPair("a", "b", 60); len(got) != 1 {
		t.Fatalf("expected one row at 60, got %q", got)
	}
}

func TestPickerListOnlyWhenFocused(t *testing.T) {
	options := []string{"a", "b", "c", "d", "e", "f", "g"}
	unfocused := Picker(Theme{}, PickerSpec{Label: "Repo", Column: 6, Current: "b", Options: options, Selected: 1})
	if len(unfocused) != 1 || !strings.HasSuffix(ansi.Strip(unfocused[0]), "b ▾") {
		t.Fatalf("expected one collapsed line, got %q", unfocused)
	}
	focused := Picker(Theme{}, PickerSpec{Label: "Repo", Column: 6, Focused: true, Input: "flt", Current: "b", Options: options, Selected: 1})
	text := ansi.Strip(strings.Join(focused, "\n"))
	if len(focused) != 7 || !strings.Contains(text, "▶ b") || !strings.Contains(text, "+2 more") || !strings.Contains(text, "flt") {
		t.Fatalf("expected input, 5 options and +2 more, got\n%s", text)
	}
}

func TestPickerEmptyShowsPlaceholder(t *testing.T) {
	got := Picker(Theme{}, PickerSpec{Label: "Repo", Column: 6, Placeholder: "Type to search or create"})
	if len(got) != 1 || !strings.Contains(ansi.Strip(got[0]), "Type to search or create") {
		t.Fatalf("expected the placeholder, got %q", got)
	}
}

func TestFooterNeverSplits(t *testing.T) {
	hints := []string{"ctrl+a create more", "ctrl+f create & focus", "tab next"}
	lines := Footer(Theme{}, "ctrl+s save", hints, 30)
	text := ansi.Strip(strings.Join(lines, "\n"))
	for _, hint := range append(hints, "ctrl+s save", "esc cancel") {
		if !strings.Contains(text, hint) {
			t.Fatalf("expected %q whole, got\n%s", hint, text)
		}
	}
	if !strings.HasPrefix(text, "ctrl+s save") || !strings.HasSuffix(text, "esc cancel") {
		t.Fatalf("expected primary first and esc last, got\n%s", text)
	}
}
