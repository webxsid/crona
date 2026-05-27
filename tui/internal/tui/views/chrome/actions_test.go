package viewchrome

import (
	"strings"
	"testing"
)

func TestGlobalActionsUseCheckInAndAwayLabels(t *testing.T) {
	daily := strings.Join(GlobalActions(Theme{}, ActionsState{View: "daily"}), "\n")
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
		View:            "away",
		RestModeActive:  true,
		AwayModeActive:  true,
	}), "\n")
	if !strings.Contains(away, "[W]") || !strings.Contains(away, "disable away") {
		t.Fatalf("expected away contextual actions to advertise disable away on W, got %q", away)
	}
}
