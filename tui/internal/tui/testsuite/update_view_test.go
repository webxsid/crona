package testsuite

import (
	"path/filepath"
	"strings"
	"testing"

	"crona/shared/config"
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	app "crona/tui/internal/tui/model"
	"crona/tui/internal/tui/testsuite/support"
	viewtypes "crona/tui/internal/tui/views/types"
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
		t.Fatalf(
			"expected dismissed updates view to stay on updates, got %s",
			updated.CurrentView(),
		)
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

func TestHomebrewInstallUsesPackageManagerCommand(t *testing.T) {
	rendered := support.RenderUpdates(viewtypes.ContentState{
		View:   "updates",
		Pane:   "issues",
		Width:  100,
		Height: 24,
		UpdateStatus: &api.UpdateStatus{
			Enabled:                  true,
			PromptEnabled:            true,
			UpdateAvailable:          true,
			CurrentVersion:           "1.5.1",
			LatestVersion:            "1.6.0",
			InstallAvailable:         false,
			InstallSource:            sharedtypes.InstallSourceBrew,
			BrewFormula:              "crona-beta",
			UpdateCommand:            "brew upgrade crona-beta",
			ReleaseURL:               "https://github.com/webxsid/crona/releases/tag/v1.6.0",
			ReleaseNotes:             "## Improvements\n- Faster startup\n",
			ChecksumsURL:             "https://example.com/checksums.txt",
			InstallScriptURL:         "https://example.com/install.sh",
			InstallUnavailableReason: "managed by Homebrew",
		},
	})
	for _, want := range []string{"Install source: brew", "Homebrew formula: crona-beta", "Update command: brew upgrade crona-beta", "[i] copy update command"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected brew update view to contain %q, got %q", want, rendered)
		}
	}
}

func TestHomebrewFormulaMismatchShowsMigrationCommand(t *testing.T) {
	rendered := support.RenderUpdates(viewtypes.ContentState{
		View:   "updates",
		Pane:   "issues",
		Width:  100,
		Height: 24,
		UpdateStatus: &api.UpdateStatus{
			Enabled:                  true,
			PromptEnabled:            true,
			UpdateAvailable:          true,
			CurrentVersion:           "1.6.0-beta.1",
			LatestVersion:            "1.6.0",
			InstallAvailable:         false,
			InstallSource:            sharedtypes.InstallSourceBrew,
			BrewFormula:              "crona",
			UpdateCommand:            "brew uninstall crona && brew install webxsid/tap/crona-beta",
			ReleaseURL:               "https://github.com/webxsid/crona/releases/tag/v1.6.0",
			ReleaseNotes:             "## Improvements\n- Faster startup\n",
			InstallUnavailableReason: "Homebrew formula mismatch: installed via crona while this build expects crona-beta. Run brew uninstall crona && brew install webxsid/tap/crona-beta.",
		},
	})
	for _, want := range []string{
		"Homebrew formula: crona",
		"Update command: brew uninstall crona && brew install webxsid/tap/crona-beta",
		"[i] copy migration command",
		"Install status: Homebrew formula mismatch:",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected migration view to contain %q, got %q", want, rendered)
		}
	}
}

func TestHomebrewInstallBlocksSelfUpdate(t *testing.T) {
	model := app.NewUpdateViewModel("", app.PaneIssues, "", &api.UpdateStatus{
		InstallAvailable: true,
		InstallSource:    sharedtypes.InstallSourceBrew,
		UpdateCommand:    "brew upgrade crona-beta",
	}, "", nil)

	if got := model.SelfUpdateUnsupportedReasonForTest(); got == "" {
		t.Fatalf("expected brew installs to report a manual update path")
	}
	if model.SelfUpdateInstallAvailableForTest() {
		t.Fatalf("expected brew installs to disable self-update")
	}
}
