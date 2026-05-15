package testsuite

import (
	"path/filepath"
	"testing"

	"crona/shared/config"
	"crona/tui/internal/api"
	app "crona/tui/internal/tui/model"
)

func TestAvailableViewsAlwaysIncludeUpdates(t *testing.T) {
	model := app.NewUpdateViewModel("", app.PaneIssues, "", nil, "", nil)

	found := false
	for _, view := range app.AvailableViewsForTest(model) {
		if view == app.ViewUpdates {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected updates view even without an available update")
	}

	model = app.NewUpdateViewModel("", app.PaneIssues, "", &api.UpdateStatus{
		Enabled:         true,
		PromptEnabled:   true,
		UpdateAvailable: true,
		LatestVersion:   "0.3.0",
	}, "", nil)
	found = false
	for _, view := range app.AvailableViewsForTest(model) {
		if view == app.ViewUpdates {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected updates view when update is available")
	}
}

func TestUpdateStatusLoadedLeavesCurrentStatusToastAlone(t *testing.T) {
	model := app.NewUpdateViewModel(app.ViewDaily, app.PaneIssues, "existing", nil, "", nil)

	updated, cmd := app.ApplyUpdateStatusLoadedForTest(model, &api.UpdateStatus{
		Enabled:         true,
		PromptEnabled:   true,
		UpdateAvailable: true,
		LatestVersion:   "0.3.0",
	})
	if cmd != nil {
		t.Fatalf("expected no toast command for update availability")
	}
	if updated.StatusMessage() != "existing" {
		t.Fatalf("expected status message to remain unchanged, got %q", updated.StatusMessage())
	}
}

func TestDismissedUpdatesViewFallsBackToDaily(t *testing.T) {
	model := app.NewUpdateViewModel(app.ViewUpdates, app.PaneIssues, "", nil, "", nil)

	updated, _ := app.ApplyUpdateDismissedForTest(model, &api.UpdateStatus{
		DismissedVersion: "0.3.0",
	})
	if updated.CurrentView() != app.ViewUpdates {
		t.Fatalf("expected dismissed updates view to stay on updates, got %s", updated.CurrentView())
	}
}

func TestSelfUpdateDisabledForNonStandardRuntimeLocation(t *testing.T) {
	t.Setenv("CRONA_INSTALL_DIR", "/tmp/crona-install")

	model := app.NewUpdateViewModel("", app.PaneIssues, "", &api.UpdateStatus{
		InstallAvailable: true,
	}, "/Users/sm2101/Projects/crona-node/bin/crona-tui", &api.KernelInfo{
		Env:            "prod",
		ExecutablePath: filepath.Join("/tmp/crona-install", config.KernelBinaryNameForMode("prod")),
	})

	if got := model.SelfUpdateUnsupportedReasonForTest(); got == "" {
		t.Fatalf("expected non-standard runtime reason")
	}
	if model.SelfUpdateInstallAvailableForTest() {
		t.Fatalf("expected self-update install to be disabled for non-standard runtime")
	}
}
