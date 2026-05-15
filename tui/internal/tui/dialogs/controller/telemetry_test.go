package controller

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnboardingDialogAdvancesAndCompletes(t *testing.T) {
	state := OpenOnboarding(State{}, true, false)
	if state.Kind != "onboarding" {
		t.Fatalf("expected onboarding dialog, got %q", state.Kind)
	}

	next, action, status := Update(state, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyTab})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryStep != 1 {
		t.Fatalf("expected telemetry step 1 after tab, got %d", next.TelemetryStep)
	}
	if action != nil {
		t.Fatalf("expected no action on first step advance, got %+v", action)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyTab})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryStep != 2 {
		t.Fatalf("expected telemetry step 2 after tab, got %d", next.TelemetryStep)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyDown})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryPrivacyCursor != 1 {
		t.Fatalf("expected privacy cursor to move to 1, got %d", next.TelemetryPrivacyCursor)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryStep != 2 {
		t.Fatalf("expected enter to be ignored outside review, got step %d", next.TelemetryStep)
	}
	if action != nil {
		t.Fatalf("expected no action outside review, got %+v", action)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyTab})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryStep != 3 {
		t.Fatalf("expected review step after tab, got %d", next.TelemetryStep)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyDown})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryReviewCursor != 1 {
		t.Fatalf("expected review cursor to move to 1, got %d", next.TelemetryReviewCursor)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close, got %q", next.Kind)
	}
	if action == nil || action.Kind != "complete_onboarding" {
		t.Fatalf("unexpected action %+v", action)
	}
}

func TestTelemetrySettingsDialogEnterTogglesPrivacyChoices(t *testing.T) {
	state := OpenEditTelemetrySettings(State{}, true, false)
	if state.Kind != "edit_telemetry_settings" {
		t.Fatalf("expected telemetry settings dialog, got %q", state.Kind)
	}

	next, action, status := Update(state, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryUsage {
		t.Fatal("expected usage setting to toggle with enter on the usage step")
	}
	if next.TelemetryStep != 0 {
		t.Fatalf("expected to stay on the usage step, got %d", next.TelemetryStep)
	}
	if action != nil {
		t.Fatalf("expected no action on usage step enter, got %+v", action)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyTab})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.TelemetryStep != 1 {
		t.Fatalf("expected diagnostics step after tab, got %d", next.TelemetryStep)
	}

	next, action, status = Update(next, UpdateContext{}, "2026-05-15", tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if !next.TelemetryErrors {
		t.Fatal("expected diagnostics setting to toggle with enter on the diagnostics step")
	}
	if next.TelemetryStep != 1 {
		t.Fatalf("expected to stay on the diagnostics step, got %d", next.TelemetryStep)
	}
	if action != nil {
		t.Fatalf("expected no action on diagnostics step enter, got %+v", action)
	}
}
