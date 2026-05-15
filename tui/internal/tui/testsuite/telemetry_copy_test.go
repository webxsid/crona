package testsuite

import (
	"strings"
	"testing"

	"crona/tui/internal/tui/dialogs"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	layoutpkg "crona/tui/internal/tui/layout"
)

func TestTelemetryDialogCopyUsesProductLanguage(t *testing.T) {
	onboarding := dialogs.Render(layoutpkg.DialogTheme(), dialogstate.OpenOnboarding(dialogstate.State{}, true, false))
	for _, want := range []string{
		"Welcome",
		"Let's get things set up.",
		"What to expect",
		"Keep issues, sessions, habits, reports, and wellbeing in one place.",
		"Your work stays on this machine",
		"change these choices later",
	} {
		if !strings.Contains(onboarding, want) {
			t.Fatalf("expected onboarding copy to contain %q, got %q", want, onboarding)
		}
	}
	if strings.Contains(onboarding, "Crona") {
		t.Fatalf("expected onboarding body to avoid repeating Crona, got %q", onboarding)
	}
	for _, unwanted := range []string{
		"tui_started",
		"daemon_started",
		"daemon_stopped",
		"error_reported",
		"running_channel",
		"entrypoint",
	} {
		if strings.Contains(onboarding, unwanted) {
			t.Fatalf("expected onboarding copy to avoid %q, got %q", unwanted, onboarding)
		}
	}

	onboardingPrivacy := dialogstate.OpenOnboarding(dialogstate.State{}, true, false)
	onboardingPrivacy.TelemetryStep = 2
	onboardingPrivacy.TelemetryUsage = true
	onboardingPrivacy.TelemetryPrivacyCursor = 1
	privacy := dialogs.Render(layoutpkg.DialogTheme(), onboardingPrivacy)
	for _, want := range []string{
		"Privacy choices",
		"▶ [ ] Share diagnostics",
		"[x] Share usage signals",
	} {
		if !strings.Contains(privacy, want) {
			t.Fatalf("expected onboarding privacy copy to contain %q, got %q", want, privacy)
		}
	}
	if strings.Contains(privacy, "▶ [x] Share usage signals") {
		t.Fatalf("expected onboarding privacy cursor to select only one toggle, got %q", privacy)
	}

	onboardingReview := dialogstate.OpenOnboarding(dialogstate.State{}, true, false)
	onboardingReview.TelemetryStep = 3
	onboardingReview.TelemetryReviewCursor = 1
	review := dialogs.Render(layoutpkg.DialogTheme(), onboardingReview)
	for _, want := range []string{
		"Review your choices",
		"▶ Start and Restart Now",
		"Start Crona",
		"Back",
	} {
		if !strings.Contains(review, want) {
			t.Fatalf("expected onboarding review copy to contain %q, got %q", want, review)
		}
	}

	settingsState := dialogstate.OpenEditTelemetrySettings(dialogstate.State{}, true, false)
	settingsState.TelemetryStep = 2
	settingsState.ChoiceCursor = 1
	settings := dialogs.Render(layoutpkg.DialogTheme(), settingsState)
	for _, want := range []string{
		"Privacy & Diagnostics",
		"Usage signals",
		"Diagnostics",
		"Save and Restart Now",
	} {
		if !strings.Contains(settings, want) {
			t.Fatalf("expected telemetry settings copy to contain %q, got %q", want, settings)
		}
	}
}
