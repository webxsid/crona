package viewchrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestGlobalActionsUseCheckInAndAwayLabels(t *testing.T) {
	daily := strings.Join(GlobalActions(Theme{}, ActionsState{View: "daily"}), "\n")
	if !strings.Contains(daily, "[?]") || !strings.Contains(daily, "keys") {
		t.Fatalf("expected global actions to advertise the help dialog, got %q", daily)
	}
	if !strings.Contains(daily, "[w]") || !strings.Contains(daily, "check-in") {
		t.Fatalf("expected daily global actions to advertise check-in, got %q", daily)
	}
	if !strings.Contains(daily, "[W]") || !strings.Contains(daily, "away") {
		t.Fatalf("expected daily global actions to advertise away on W, got %q", daily)
	}

	wellbeing := strings.Join(GlobalActions(Theme{}, ActionsState{View: "wellbeing"}), "\n")
	if !strings.Contains(wellbeing, "[w]") || !strings.Contains(wellbeing, "check-in") {
		t.Fatalf("expected wellbeing global actions to advertise check-in on w, got %q", wellbeing)
	}
	if !strings.Contains(wellbeing, "[W]") || !strings.Contains(wellbeing, "away") {
		t.Fatalf("expected wellbeing global actions to advertise away on W, got %q", wellbeing)
	}

	away := strings.Join(GlobalActions(Theme{}, ActionsState{View: "away", AwayModeActive: true}), "\n")
	if !strings.Contains(away, "[W]") || !strings.Contains(away, "disable away") {
		t.Fatalf("expected away global actions to advertise disable away on W, got %q", away)
	}
}

func TestContextualActionsUseCheckInAndAwayLabels(t *testing.T) {
	wellbeing := strings.Join(ContextualActions(Theme{}, ActionsState{View: "wellbeing"}), "\n")
	if !strings.Contains(wellbeing, "[w]") || !strings.Contains(wellbeing, "check-in") {
		t.Fatalf("expected wellbeing contextual actions to advertise check-in, got %q", wellbeing)
	}

	away := strings.Join(ContextualActions(Theme{}, ActionsState{
		View:           "away",
		RestModeActive: true,
		AwayModeActive: true,
	}), "\n")
	if !strings.Contains(away, "[W]") || !strings.Contains(away, "disable away") {
		t.Fatalf("expected away contextual actions to advertise disable away on W, got %q", away)
	}
}

func TestUpdateActionsLabelMigrationCommands(t *testing.T) {
	actions := strings.Join(ContextualActions(Theme{}, ActionsState{
		View:                   "updates",
		UpdateCommand:          "brew uninstall crona-beta && brew install webxsid/tap/crona",
		UpdateVisible:          true,
		UpdateInstallAvailable: false,
	}), "\n")
	if !strings.Contains(actions, "copy migration command") {
		t.Fatalf("expected migration command label, got %q", actions)
	}
	if !IsMigrationCommand("brew uninstall crona-beta && brew install webxsid/tap/crona") {
		t.Fatalf("expected migration command helper to detect uninstall/install flow")
	}
}

func TestRenderPaneActionLineCompactsWrappedActions(t *testing.T) {
	rendered := RenderPaneActionLine(
		Theme{},
		[]string{"[enter] open issue details", "[a] create issue"},
		32,
	)
	plain := ansi.Strip(rendered)
	if strings.Contains(plain, "\n") {
		t.Fatalf("expected compact action line to stay on one row, got %q", rendered)
	}
	for _, want := range []string{"[enter] details", "[a] create"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected compact action line to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"open issue details", "create issue"} {
		if strings.Contains(plain, unwanted) {
			t.Fatalf("expected compact action line to shorten %q, got %q", unwanted, rendered)
		}
	}
}

func TestRenderPaneActionLineKeepsWideLabels(t *testing.T) {
	rendered := RenderPaneActionLine(
		Theme{},
		[]string{"[enter] open issue details", "[a] create issue"},
		120,
	)
	if strings.Contains(rendered, "\n") {
		t.Fatalf("expected wide action line to stay on one row, got %q", rendered)
	}
	for _, want := range []string{"open issue details", "create issue"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected wide action line to keep %q, got %q", want, rendered)
		}
	}
}
