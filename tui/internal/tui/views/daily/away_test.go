package daily

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderViewShowsHistoricalAwayModePanel(t *testing.T) {
	state := dailySummaryTestState()
	state.Width = 100
	state.Height = 30
	state.HistoricalAwayDate = "2026-05-27"
	rendered := ansi.Strip(renderView(testTheme(), state))
	if !strings.Contains(rendered, "Away Mode") ||
		!strings.Contains(rendered, "You chose to rest and recover on") {
		t.Fatalf("expected historical away panel, got %q", rendered)
	}
	if strings.Contains(rendered, "Signals") {
		t.Fatalf("expected historical away panel to replace daily content, got %q", rendered)
	}
}
