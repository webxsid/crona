package momentumview

import (
	"regexp"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func renderState(width, height int, cards ...sharedtypes.MomentumCard) types.ContentState {
	return types.ContentState{
		View: "momentum", Pane: "momentum_cards", Width: width, Height: height,
		MomentumDate: "2026-10-18", MomentumWindowDays: 28,
		MomentumCards: cards,
		Cursors:       map[string]int{"momentum_cards": 0},
		Filters:       map[string]string{},
	}
}

func plain(s string) string { return ansi.Strip(s) }

func assertFits(t *testing.T, rendered string, width, height int) {
	t.Helper()
	lines := strings.Split(rendered, "\n")
	if len(lines) > height {
		t.Fatalf("expected at most %d lines, got %d", height, len(lines))
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > width {
			t.Fatalf("expected lines within %d cells, got %d: %q", width, w, plain(line))
		}
	}
}

func TestRenderNarrowListOnly(t *testing.T) {
	rendered := Render(types.Theme{}, renderState(60, 20, momentumTestCard("Deep work", "W1", "W2")))
	assertFits(t, rendered, 60, 20)
	text := plain(rendered)
	if strings.Contains(text, "Cadence") || !strings.Contains(text, "[enter]") {
		t.Fatalf("expected the list alone with an enter hint, got\n%s", text)
	}
}

func TestRenderMediumAndWideShowDetail(t *testing.T) {
	for _, size := range [][2]int{{120, 36}, {200, 50}} {
		rendered := Render(types.Theme{}, renderState(size[0], size[1], momentumTestCard("Deep work", "W1", "W2"), momentumTestCard("Reading", "W1")))
		assertFits(t, rendered, size[0], size[1])
		text := plain(rendered)
		for _, want := range []string{"Cadence", "Target", "Streak", "Deep work"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%v: expected %q in\n%s", size, want, text)
			}
		}
	}
}

func TestRenderInactiveDimmedAndPaused(t *testing.T) {
	off := momentumTestCard("Piano", "W1")
	off.Definition.Enabled = false
	state := renderState(120, 36, off)
	text := plain(Render(types.Theme{}, state))
	if !strings.Contains(text, "paused") || !strings.Contains(text, "[x] enable") {
		t.Fatalf("expected a paused row and an enable key, got\n%s", text)
	}
	if text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1")))); !strings.Contains(text, "[x] disable") {
		t.Fatalf("expected a disable key, got\n%s", text)
	}
}

func TestRenderListCreateLineAndEmpty(t *testing.T) {
	if text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1")))); !strings.Contains(text, "[a] new momentum") {
		t.Fatalf("expected a create line, got\n%s", text)
	}
	if text := squash(plain(Render(types.Theme{}, renderState(120, 36)))); !strings.Contains(text, "No momentums yet [a] new momentum") {
		t.Fatalf("expected the empty state, got\n%s", text)
	}
}

func TestRenderNoBordersOrHeader(t *testing.T) {
	text := plain(Render(types.Theme{}, renderState(120, 36, momentumTestCard("Deep work", "W1"))))
	if strings.Contains(text, "╭") || regexp.MustCompile(`Momentum\s+2026-`).MatchString(text) {
		t.Fatalf("expected no box or dated header, got\n%s", text)
	}
}

func TestRenderTinyHeight(t *testing.T) {
	for _, height := range []int{1, 4} {
		assertFits(t, Render(types.Theme{}, renderState(120, height, momentumTestCard("Deep work", "W1"))), 120, height)
	}
}

func TestRenderUsesFilteredRows(t *testing.T) {
	state := renderState(120, 36, momentumTestCard("Deep work", "W1"), momentumTestCard("Reading", "W1", "W2"))
	state.Filters["momentum_cards"] = "read"
	text := plain(Render(types.Theme{}, state))
	if strings.Contains(text, "Deep work") || !strings.Contains(text, "Reading") {
		t.Fatalf("expected only the filtered momentum, got\n%s", text)
	}
}

func TestDetailLinesWrapDescription(t *testing.T) {
	card := momentumTestCard("Deep work", "W1")
	description := strings.Repeat("steady focused work ", 4)
	card.Definition.Description = &description
	for _, line := range DetailLines(types.Theme{}, card, 30, 40) {
		if w := ansi.StringWidth(line); w > 30 {
			t.Fatalf("expected lines within 30 cells, got %q", plain(line))
		}
	}
	if !strings.Contains(squash(plain(strings.Join(DetailLines(types.Theme{}, card, 30, 40), " "))), strings.TrimSpace(description)) {
		t.Fatalf("expected the whole description")
	}
}
