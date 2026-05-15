package calendar

import (
	"strings"
	"testing"

	"crona/tui/internal/tui/chrome"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderSingleDateMarksSelectionTodayAndCurrentWeek(t *testing.T) {
	rendered := ansi.Strip(strings.Join(Render(testTheme(), Selection{
		AnchorDate:   "2026-03-19",
		SelectedDate: "2026-03-19",
		Today:        "2026-05-14",
	}), "\n"))

	for _, want := range []string{"March 2026", "Week 12", "Today 14", "Wk 20", "19"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected calendar to contain %q, got %q", want, rendered)
		}
	}
}

func TestRenderRangeMarksVisibleRangeAndTodayCell(t *testing.T) {
	rendered := ansi.Strip(strings.Join(Render(testTheme(), Selection{
		AnchorDate: "2026-05-14",
		RangeStart: "2026-05-12",
		RangeEnd:   "2026-05-16",
		Today:      "2026-05-14",
	}), "\n"))

	for _, want := range []string{"May 2026", "Range W20-W20", "12", "13", "14", "15", "16", "20"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected range calendar to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"[14]", "[20]"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected calendar to avoid bracket marker %q, got %q", unwanted, rendered)
		}
	}
}

func TestRenderWindowKeepsAnchorWeekVisible(t *testing.T) {
	lines := Render(testTheme(), Selection{
		AnchorDate:   "2026-04-30",
		SelectedDate: "2026-04-30",
		MaxLines:     5,
		Today:        "2026-05-14",
	})
	rendered := ansi.Strip(strings.Join(lines, "\n"))

	if len(lines) != 5 {
		t.Fatalf("expected clipped calendar to use 5 lines, got %d: %q", len(lines), rendered)
	}
	if !strings.Contains(rendered, "30") {
		t.Fatalf("expected clipped calendar to keep anchor week visible, got %q", rendered)
	}
	if !strings.Contains(rendered, "Today 14") || !strings.Contains(rendered, "Wk 20") {
		t.Fatalf("expected clipped calendar metadata to keep today/current week visible, got %q", rendered)
	}
}

func TestRenderUsesStyledCellsWithoutBracketMarkers(t *testing.T) {
	rendered := ansi.Strip(strings.Join(Render(testTheme(), Selection{
		AnchorDate:   "2026-05-14",
		SelectedDate: "2026-05-14",
		Today:        "2026-05-14",
	}), "\n"))

	for _, unwanted := range []string{"[14]", "[20]"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected styled calendar cells without bracket marker %q, got %q", unwanted, rendered)
		}
	}
}

func testTheme() viewtypes.Theme {
	return viewtypes.Theme{
		ColorBlue:            chrome.ColorBlue,
		ColorCyan:            chrome.ColorCyan,
		ColorGreen:           chrome.ColorGreen,
		ColorMagenta:         chrome.ColorMagenta,
		ColorSubtle:          chrome.ColorSubtle,
		ColorYellow:          chrome.ColorYellow,
		ColorRed:             chrome.ColorRed,
		ColorDim:             chrome.ColorDim,
		ColorWhite:           chrome.ColorWhite,
		StyleActive:          chrome.StyleActive,
		StyleInactive:        chrome.StyleInactive,
		StylePaneTitle:       chrome.StylePaneTitle,
		StyleDim:             chrome.StyleDim,
		StyleCursor:          chrome.StyleCursor,
		StyleHeader:          chrome.StyleHeader,
		StyleError:           chrome.StyleError,
		StyleSelected:        chrome.StyleSelected,
		StyleSelectedInverse: chrome.StyleSelectedInverse,
		StyleNormal:          chrome.StyleNormal,
	}
}
