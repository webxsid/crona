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
		Height: 40,
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
			InstallScriptDeprecated:  true,
			MigrationGuideURL:        "https://crona.work/migration",
		},
	})
	for _, want := range []string{
		"Important",
		"Install script deprecation",
		"Migration guide: https://crona.work/migration",
		"Migration steps",
		"crona backup",
		"crona restore <backup-path>",
		"Current Version",
		"Latest Version",
		"Install Source",
		"Homebrew",
		"Update Command",
		"brew upgrade crona-beta",
		"What's New",
		"[o] open migration guide",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected brew update view to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Homebrew formula:", "Release Page", "Checksums", "Configured Channel", "Install Status", "Installer", "TUI Path", "Engine Path"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected brew update view to keep %q hidden by default, got %q", unwanted, rendered)
		}
	}
	if strings.Contains(rendered, "[i]") {
		t.Fatalf("expected brew update view to hide install action, got %q", rendered)
	}
}

func TestWingetInstallUsesPackageManagerCommand(t *testing.T) {
	rendered := support.RenderUpdates(viewtypes.ContentState{
		View:   "updates",
		Pane:   "issues",
		Width:  100,
		Height: 40,
		UpdateStatus: &api.UpdateStatus{
			Enabled:                  true,
			PromptEnabled:            true,
			UpdateAvailable:          true,
			CurrentVersion:           "1.5.1",
			LatestVersion:            "1.6.0",
			InstallAvailable:         false,
			InstallSource:            sharedtypes.InstallSourceWinget,
			UpdateCommand:            "winget upgrade --id Webxsid.Crona -e",
			ReleaseURL:               "https://github.com/webxsid/crona/releases/tag/v1.6.0",
			ReleaseNotes:             "## Improvements\n- Faster startup\n",
			ChecksumsURL:             "https://example.com/checksums.txt",
			InstallScriptURL:         "https://example.com/install.ps1",
			InstallUnavailableReason: "managed by winget",
			InstallScriptDeprecated:  true,
			MigrationGuideURL:        "https://crona.work/migration",
		},
	})
	for _, want := range []string{
		"Important",
		"Install script deprecation",
		"Migration guide: https://crona.work/migration",
		"[c] copy command",
		"Migration steps",
		"crona backup",
		"crona restore <backup-path>",
		"Current Version",
		"Latest Version",
		"Install Source",
		"winget",
		"Update Command",
		"winget upgrade --id Webxsid.Crona -e",
		"What's New",
		"[o] open migration guide",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected winget update view to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Release Page", "Checksums", "Configured Channel", "Install Status", "Installer", "TUI Path", "Engine Path"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected winget update view to keep %q hidden by default, got %q", unwanted, rendered)
		}
	}
	if strings.Contains(rendered, "[i]") {
		t.Fatalf("expected winget update view to hide install action, got %q", rendered)
	}
}

func TestHomebrewFormulaMismatchShowsMigrationCommand(t *testing.T) {
	rendered := support.RenderUpdates(viewtypes.ContentState{
		View:   "updates",
		Pane:   "issues",
		Width:  100,
		Height: 40,
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
			InstallScriptDeprecated:  true,
			MigrationGuideURL:        "https://crona.work/migration",
		},
	})
	for _, want := range []string{
		"Important",
		"Install script deprecation",
		"Migration guide: https://crona.work/migration",
		"[c] copy migration command",
		"Migration steps",
		"crona backup",
		"crona restore <backup-path>",
		"Current Version",
		"Latest Version",
		"Install Source",
		"Homebrew",
		"Update Command",
		"brew uninstall crona && brew install webxsid/tap/crona-beta",
		"What's New",
		"[o] open migration guide",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected migration view to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Release Page", "Checksums", "Configured Channel", "Install Status", "Installer", "TUI Path", "Engine Path", "Homebrew formula:", "Install status:"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected migration view to keep %q hidden by default, got %q", unwanted, rendered)
		}
	}
	if strings.Contains(rendered, "[i]") {
		t.Fatalf("expected migration view to hide install action, got %q", rendered)
	}
}

func TestUpdatesViewExpandedDiagnosticsRevealInternalFields(t *testing.T) {
	rendered := support.RenderUpdates(viewtypes.ContentState{
		View:                      "updates",
		Pane:                      "issues",
		Width:                     100,
		Height:                    60,
		UpdateDiagnosticsExpanded: true,
		UpdateStatus: &api.UpdateStatus{
			Enabled:                  true,
			PromptEnabled:            true,
			UpdateAvailable:          true,
			CurrentVersion:           "1.5.1",
			LatestVersion:            "1.6.0",
			InstallSource:            sharedtypes.InstallSourceBrew,
			BrewFormula:              "crona",
			UpdateCommand:            "brew upgrade crona",
			ReleaseTag:               "v1.6.0",
			ReleaseURL:               "https://github.com/webxsid/crona/releases/tag/v1.6.0",
			ReleaseNotes:             "## Improvements\n- Faster startup\n",
			InstallUnavailableReason: "managed by Homebrew",
			InstallScriptDeprecated:  true,
			MigrationGuideURL:        "https://crona.work/migration",
			Channel:                  sharedtypes.UpdateChannelStable,
			RunningChannel:           sharedtypes.UpdateChannelStable,
			RunningIsBeta:            false,
			ReleaseIsPrerelease:      false,
			LatestIsBeta:             false,
			CheckedAt:                "2026-06-16T06:20:10Z",
		},
		UpdateManualReason:   "managed by Homebrew",
		TUIExecutablePath:    "/opt/homebrew/bin/crona-tui",
		KernelExecutablePath: "/opt/homebrew/bin/crona-kernel",
	})
	for _, want := range []string{
		"Diagnostics",
		"[d] hide diagnostics",
		"Release Tag",
		"v1.6.0",
		"Release Page",
		"Configured Channel",
		"Install Source",
		"Brew Formula",
		"Last Checked",
		"Update Command",
		"TUI Path",
		"Engine Path",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected expanded diagnostics to contain %q, got %q", want, rendered)
		}
	}
	if !strings.Contains(rendered, "[d] hide diagnostics") {
		t.Fatalf("expected expanded diagnostics toggle, got %q", rendered)
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

func TestWingetInstallBlocksSelfUpdate(t *testing.T) {
	model := app.NewUpdateViewModel("", app.PaneIssues, "", &api.UpdateStatus{
		InstallAvailable: true,
		InstallSource:    sharedtypes.InstallSourceWinget,
		UpdateCommand:    "winget upgrade --id Webxsid.Crona -e",
	}, "", nil)

	if got := model.SelfUpdateUnsupportedReasonForTest(); got == "" {
		t.Fatalf("expected winget installs to report a manual update path")
	}
	if model.SelfUpdateInstallAvailableForTest() {
		t.Fatalf("expected winget installs to disable self-update")
	}
}
