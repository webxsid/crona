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
			uistate.PaneSettings: 11,
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
			uistate.PaneSettings: 11,
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
			ErrorReportingEnabled: false,
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 16,
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

func TestHandleActivateSelectedSettingNavigatesToMomentumView(t *testing.T) {
	state := State{
		ActiveView: uistate.ViewSettings,
		Settings: &sharedtypes.CoreSettings{
			HabitStreakDefs: []sharedtypes.HabitStreakDefinition{{Name: "Focus"}},
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 12,
		},
	}

	next, _, handled := handleActivateSelectedSetting(state, Deps{})
	if !handled {
		t.Fatalf("expected habit streaks row activation to be handled")
	}
	out := next.(State)
	if out.ActiveView != uistate.ViewMomentum {
		t.Fatalf("expected settings row to navigate to momentum view, got %q", out.ActiveView)
	}
	if out.ActivePane != uistate.PaneMomentumCards {
		t.Fatalf("expected momentum pane to be selected, got %q", out.ActivePane)
	}
}

func TestHandleAdjustSelectedSettingUpdatesWeekStart(t *testing.T) {
	var patchedKey sharedtypes.CoreSettingsKey
	var patchedValue any
	state := State{
		ActiveView: uistate.ViewSettings,
		Settings: &sharedtypes.CoreSettings{
			WeekStart: sharedtypes.WeekStartMonday,
		},
		Cursor: map[uistate.Pane]int{
			uistate.PaneSettings: 10,
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
		t.Fatalf("expected week start row to be handled")
	}
	if patchedKey != sharedtypes.CoreSettingsKeyWeekStart {
		t.Fatalf("expected week start patch, got %q", patchedKey)
	}
	if patchedValue != sharedtypes.WeekStartSunday {
		t.Fatalf("expected next week start Sunday, got %#v", patchedValue)
	}
}
