package app

import (
	"path/filepath"
	"testing"

	"crona/shared/config"
	"crona/tui/internal/api"
)

func TestAvailableViewsAlwaysIncludeUpdates(t *testing.T) {
	model := Model{}

	found := false
	for _, view := range model.availableViews() {
		if view == ViewUpdates {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected updates view even without an available update")
	}

	model.updateStatus = &api.UpdateStatus{
		Enabled:         true,
		PromptEnabled:   true,
		UpdateAvailable: true,
		LatestVersion:   "0.3.0",
	}
	found = false
	for _, view := range model.availableViews() {
		if view == ViewUpdates {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected updates view when update is available")
	}
}

func TestUpdateStatusLoadedLeavesCurrentStatusToastAlone(t *testing.T) {
	model := Model{
		view:      ViewDaily,
		pane:      PaneIssues,
		statusMsg: "existing",
	}

	next, cmd := model.Update(updateStatusLoadedMsg{Status: &api.UpdateStatus{
		Enabled:         true,
		PromptEnabled:   true,
		UpdateAvailable: true,
		LatestVersion:   "0.3.0",
	}})
	if cmd != nil {
		t.Fatalf("expected no toast command for update availability")
	}
	updated := next.(Model)
	if updated.statusMsg != "existing" {
		t.Fatalf("expected status message to remain unchanged, got %q", updated.statusMsg)
	}
}

func TestDismissedUpdatesViewFallsBackToDaily(t *testing.T) {
	model := Model{
		view: ViewUpdates,
		pane: PaneIssues,
	}

	next, _ := model.Update(updateDismissedMsg{Status: &api.UpdateStatus{
		DismissedVersion: "0.3.0",
	}})
	updated := next.(Model)
	if updated.view != ViewUpdates {
		t.Fatalf("expected dismissed updates view to stay on updates, got %s", updated.view)
	}
}

func TestSelfUpdateDisabledForNonStandardRuntimeLocation(t *testing.T) {
	t.Setenv("CRONA_INSTALL_DIR", "/tmp/crona-install")

	model := Model{
		currentExecutablePath: "/Users/sm2101/Projects/crona-node/bin/crona-tui",
		kernelInfo: &api.KernelInfo{
			Env:            "prod",
			ExecutablePath: filepath.Join("/tmp/crona-install", config.KernelBinaryNameForMode("prod")),
		},
		updateStatus: &api.UpdateStatus{
			InstallAvailable: true,
		},
	}

	if got := model.selfUpdateUnsupportedReason(); got == "" {
		t.Fatalf("expected non-standard runtime reason")
	}
	if model.selfUpdateInstallAvailable() {
		t.Fatalf("expected self-update install to be disabled for non-standard runtime")
	}
}
