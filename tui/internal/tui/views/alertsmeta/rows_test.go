package alertsmeta

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestRowsIncludeInactivityControls(t *testing.T) {
	settings := &sharedtypes.CoreSettings{
		BoundaryNotifications: true,
		BoundarySound:         true,
		AlertSoundPreset:      sharedtypes.AlertSoundPresetChime,
		AlertUrgency:          sharedtypes.AlertUrgencyNormal,
		AlertIconEnabled:      true,
		InactivityAlerts:      true,
		InactivityThreshold:   60,
		InactivityRepeat:      90,
	}

	rows := Rows(settings, nil, nil)
	seen := map[RowKey]string{}
	for _, row := range rows {
		seen[row.Key] = row.Value
	}
	if seen[RowInactivityAlerts] != "Enabled" {
		t.Fatalf("expected inactivity toggle row, got %q", seen[RowInactivityAlerts])
	}
	if seen[RowInactivityAfter] != "60m" {
		t.Fatalf("expected inactivity threshold row, got %q", seen[RowInactivityAfter])
	}
	if seen[RowInactivityRepeat] != "90m" {
		t.Fatalf("expected inactivity repeat row, got %q", seen[RowInactivityRepeat])
	}
}
