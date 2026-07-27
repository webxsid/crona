package alertsmeta

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
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
	if got := soundPresetLabel(sharedtypes.AlertSoundPresetNotificationPing); got != "Notification Ping" {
		t.Fatalf("expected notification ping label, got %q", got)
	}
}

func TestRowsIncludeBothReminderActionsAndLabels(t *testing.T) {
	settings := &sharedtypes.CoreSettings{}
	reminders := []api.AlertReminder{
		{ID: "checkin", Kind: sharedtypes.AlertReminderKindCheckIn},
		{ID: "plan", Kind: sharedtypes.AlertReminderKindDailyPlan},
	}

	rows := Rows(settings, nil, reminders)
	labels := map[RowKey]string{}
	for _, row := range rows {
		labels[row.Key] = row.Label
	}

	if labels[RowAddCheckInReminder] != "Add Check-In Reminder" {
		t.Fatalf("missing check-in reminder action: %+v", labels)
	}
	if labels[RowAddDailyPlanReminder] != "Add Plan-the-Day Reminder" {
		t.Fatalf("missing daily-plan reminder action: %+v", labels)
	}
	if labels[reminderRowKey("plan")] != "Plan-the-Day Reminder" {
		t.Fatalf("unexpected daily-plan reminder label: %+v", labels)
	}
}
