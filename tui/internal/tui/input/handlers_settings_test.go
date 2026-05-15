package input

import (
	"testing"

	sharedtypes "crona/shared/types"
	uistate "crona/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleAdjustSelectedSettingUpdatesPromptGlyphMode(t *testing.T) {
	var patchedKey sharedtypes.CoreSettingsKey
	var patchedValue any
	state := State{
		ActiveView: uistate.ViewSettings,
		Settings: &sharedtypes.CoreSettings{
			PromptGlyphMode: sharedtypes.PromptGlyphModeEmoji,
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 19,
		},
	}
	deps := Deps{
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			patchedKey = key
			patchedValue = value
			return nil
		},
	}

	_, _, handled := handleAdjustSelectedSetting(state, deps, 1)
	if !handled {
		t.Fatalf("expected settings row to be handled")
	}
	if patchedKey != sharedtypes.CoreSettingsKeyPromptGlyphMode {
		t.Fatalf("expected prompt glyph mode patch, got %q", patchedKey)
	}
	if patchedValue != sharedtypes.PromptGlyphModeUnicode {
		t.Fatalf("expected next prompt glyph mode unicode, got %#v", patchedValue)
	}
}

func TestHandleActivateSelectedSettingUpdatesPromptGlyphMode(t *testing.T) {
	var patchedKey sharedtypes.CoreSettingsKey
	state := State{
		ActiveView: uistate.ViewSettings,
		Settings: &sharedtypes.CoreSettings{
			PromptGlyphMode: sharedtypes.PromptGlyphModeUnicode,
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 19,
		},
	}
	deps := Deps{
		PatchSetting: func(key sharedtypes.CoreSettingsKey, value any, repoID, streamID int64, dashboardDate string) tea.Cmd {
			patchedKey = key
			return nil
		},
	}

	_, _, handled := handleActivateSelectedSetting(state, deps)
	if !handled {
		t.Fatalf("expected settings row activation to be handled")
	}
	if patchedKey != sharedtypes.CoreSettingsKeyPromptGlyphMode {
		t.Fatalf("expected prompt glyph mode patch on activation, got %q", patchedKey)
	}
}

func TestHandleActivateSelectedSettingOpensTelemetrySettingsDialog(t *testing.T) {
	state := State{
		ActiveView: uistate.ViewSettings,
		Settings: &sharedtypes.CoreSettings{
			UsageTelemetryEnabled: true,
			ErrorReportingEnabled:  false,
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 24,
		},
	}
	deps := Deps{
		OpenEditTelemetrySettingsDialog: func(state *State) bool {
			state.Dialog = "edit_telemetry_settings"
			return true
		},
	}

	next, _, handled := handleActivateSelectedSetting(state, deps)
	if !handled {
		t.Fatalf("expected telemetry settings row activation to be handled")
	}
	out := next.(State)
	if out.Dialog != "edit_telemetry_settings" {
		t.Fatalf("expected telemetry settings dialog to open, got %q", out.Dialog)
	}
}
