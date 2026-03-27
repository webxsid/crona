package api

import (
	"encoding/json"
	"testing"
)

func TestDecodeSettingsReadsBoundarySettingsFromPublicShape(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"local": map[string]any{
			"userId":                       "local",
			"deviceId":                     "device-1",
			"timerMode":                    "structured",
			"breaksEnabled":                true,
			"workDurationMinutes":          25,
			"shortBreakMinutes":            5,
			"longBreakMinutes":             15,
			"longBreakEnabled":             true,
			"cyclesBeforeLongBreak":        4,
			"autoStartBreaks":              false,
			"autoStartWork":                false,
			"boundaryNotificationsEnabled": false,
			"boundarySoundEnabled":         true,
			"updateChecksEnabled":          true,
			"updatePromptEnabled":          true,
			"repoSort":                     "chronological_asc",
			"streamSort":                   "chronological_asc",
			"issueSort":                    "priority",
			"habitSort":                    "schedule",
			"createdAt":                    "1",
			"updatedAt":                    "2",
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	settings, err := decodeSettings(raw)
	if err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if settings == nil {
		t.Fatalf("expected settings, got nil")
	}
	if settings.BoundaryNotifications {
		t.Fatalf("expected boundary notifications false, got true")
	}
	if !settings.BoundarySound {
		t.Fatalf("expected boundary sound true, got false")
	}
}
