package settings

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	settingsmeta "crona/tui/internal/tui/views/settingsmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func testSettings() *api.CoreSettings {
	return &sharedtypes.CoreSettings{ManageHistory: true, WeekStart: sharedtypes.WeekStartMonday}
}

func TestSettingsEveryRowHasHelp(t *testing.T) {
	for _, row := range settingsmeta.Rows(testSettings()) {
		if settingsHelp[row.Label].Description == "" {
			t.Fatalf("expected help for %q", row.Label)
		}
	}
}

func TestSettingsCurrentIsAnOption(t *testing.T) {
	for _, row := range settingsmeta.Rows(testSettings()) {
		options := settingsHelp[row.Label].Options
		if len(options) == 0 {
			continue
		}
		found := false
		for _, option := range options {
			found = found || strings.EqualFold(option, row.Value)
		}
		if !found {
			t.Fatalf("%q: value %q not among %v", row.Label, row.Value, options)
		}
	}
}

func settingsText(width, height, cursor int) string {
	state := types.ContentState{
		View: "settings", Pane: "settings", Width: width, Height: height, Settings: testSettings(),
		Cursors: map[string]int{"settings": cursor}, Filters: map[string]string{},
	}
	return ansi.Strip(renderView(types.Theme{}, state))
}

func TestSettingsLayout(t *testing.T) {
	text := settingsText(120, 36, 0)
	for _, want := range []string{"── History", "▶ Manage History", "─ Manage History ", "● Enabled", "[←/→] change"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "╭") {
		t.Fatalf("expected no box, got\n%s", text)
	}
	if narrow := settingsText(60, 20, 0); strings.Contains(narrow, "● Enabled") {
		t.Fatalf("expected no detail when narrow, got\n%s", narrow)
	}
}
