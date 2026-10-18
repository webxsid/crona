package alerts

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func alertsState(width, height int) types.ContentState {
	return types.ContentState{
		View: "alerts", Pane: "alerts", Width: width, Height: height,
		Cursors: map[string]int{"alerts": 0}, Filters: map[string]string{},
		Settings:    &sharedtypes.CoreSettings{BoundaryNotifications: true},
		AlertStatus: &api.AlertStatus{NotificationsAvailable: true, NotificationBackend: "terminal-notifier"},
		AlertReminders: []api.AlertReminder{
			{ID: "r1", Kind: sharedtypes.AlertReminderKindCheckIn},
			{ID: "r2", Kind: sharedtypes.AlertReminderKindDailyPlan},
		},
	}
}

func TestAlertsEveryRowHasHelp(t *testing.T) {
	state := alertsState(120, 36)
	for _, row := range alertsmeta.Rows(state.Settings, state.AlertStatus, state.AlertReminders) {
		if row.Selectable && helpFor(row).Description == "" {
			t.Fatalf("expected help for %q", row.Label)
		}
	}
}

func TestAlertsBackendInStrip(t *testing.T) {
	lines := strings.Split(ansi.Strip(renderView(types.Theme{}, alertsState(160, 36))), "\n")
	if !strings.Contains(strings.Join(lines[:2], "\n"), "terminal-notifier") {
		t.Fatalf("expected the backend in the strip, got\n%s", strings.Join(lines, "\n"))
	}
	if strings.Contains(strings.Join(lines, "\n"), "── Backend") {
		t.Fatalf("expected no Backend group in the list")
	}
}

func TestAlertsLayout(t *testing.T) {
	text := ansi.Strip(renderView(types.Theme{}, alertsState(120, 36)))
	for _, want := range []string{"── Delivery", "▶ Notifications", "● Enabled", "[←/→] change"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "╭") {
		t.Fatalf("expected no box")
	}
	for _, line := range strings.Split(ansi.Strip(renderView(types.Theme{}, alertsState(60, 20))), "\n") {
		if ansi.StringWidth(line) > 60 {
			t.Fatalf("expected lines within 60, got %q", line)
		}
	}
}
