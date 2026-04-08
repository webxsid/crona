package alertsmeta

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewhelpers "crona/tui/internal/tui/views/helpers"
)

type RowKey string

const (
	RowNotifications    RowKey = "notifications"
	RowSound            RowKey = "sound"
	RowSoundPreset      RowKey = "sound_preset"
	RowUrgency          RowKey = "urgency"
	RowLogoIcon         RowKey = "logo_icon"
	RowTestNotification RowKey = "test_notification"
	RowTestSound        RowKey = "test_sound"
	RowAddReminder      RowKey = "add_checkin_reminder"
)

type Row struct {
	Section      string
	Key          RowKey
	Label        string
	Value        string
	Selectable   bool
	ReminderID   string
	ReminderKind sharedtypes.AlertReminderKind
}

type VisibleRow struct {
	Header       bool
	Text         string
	SelectableAt int
	Key          RowKey
}

func Rows(settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder) []Row {
	if settings == nil {
		return nil
	}
	rows := []Row{
		{Section: "Delivery", Key: RowNotifications, Label: "Notifications", Value: enabledDisabled(settings.BoundaryNotifications), Selectable: true},
		{Section: "Delivery", Key: RowSound, Label: "Sound", Value: enabledDisabled(settings.BoundarySound), Selectable: true},
		{Section: "Delivery", Key: RowSoundPreset, Label: "Sound Preset", Value: soundPresetLabel(settings.AlertSoundPreset), Selectable: true},
		{Section: "Delivery", Key: RowUrgency, Label: "Urgency", Value: urgencyLabel(settings.AlertUrgency), Selectable: true},
		{Section: "Delivery", Key: RowLogoIcon, Label: "Logo Icon", Value: enabledDisabled(settings.AlertIconEnabled), Selectable: true},
	}
	for _, reminder := range reminders {
		rows = append(rows, Row{
			Section:      "Reminders",
			Key:          reminderRowKey(reminder.ID),
			Label:        reminderLabel(reminder),
			Value:        reminderSummary(reminder),
			Selectable:   true,
			ReminderID:   reminder.ID,
			ReminderKind: reminder.Kind,
		})
	}
	if status != nil {
		rows = append(rows,
			Row{Section: "Backend", Label: "Notification Backend", Value: backendLabel(status.NotificationsAvailable, status.NotificationBackend)},
			Row{Section: "Backend", Label: "Sound Backend", Value: backendLabel(status.SoundAvailable, status.SoundBackend)},
			Row{Section: "Backend", Label: "Subtitle", Value: supportedLabel(status.SubtitleSupported)},
			Row{Section: "Backend", Label: "Urgency", Value: supportedLabel(status.UrgencySupported)},
			Row{Section: "Backend", Label: "Icon", Value: supportedLabel(status.IconSupported)},
			Row{Section: "Backend", Label: "Bundled Sound", Value: supportedLabel(status.BundledSoundSupported)},
		)
	}
	rows = append(rows,
		Row{Section: "Actions", Key: RowAddReminder, Label: "Add Check-In Reminder", Value: "Create a scheduled nightly reminder", Selectable: true},
		Row{Section: "Actions", Key: RowTestNotification, Label: "Test Notification", Value: "Send sample alert", Selectable: true},
		Row{Section: "Actions", Key: RowTestSound, Label: "Test Sound", Value: "Play selected preset", Selectable: true},
	)
	return rows
}

func ItemLabels(settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder) []string {
	rows := Rows(settings, status, reminders)
	labels := make([]string, 0, len(rows))
	for _, row := range rows {
		label := row.Label
		if strings.TrimSpace(row.Value) != "" {
			label += " " + row.Value
		}
		labels = append(labels, label)
	}
	return labels
}

func FilteredIndices(filter string, settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder) []int {
	if settings == nil {
		return nil
	}
	return viewhelpers.FilteredStrings(ItemLabels(settings, status, reminders), filter)
}

func FilteredSelectableRows(filter string, settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder) []Row {
	rows := Rows(settings, status, reminders)
	indices := FilteredIndices(filter, settings, status, reminders)
	out := make([]Row, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(rows) && rows[idx].Selectable {
			out = append(out, rows[idx])
		}
	}
	return out
}

func FilteredSelectableCount(filter string, settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder) int {
	return len(FilteredSelectableRows(filter, settings, status, reminders))
}

func SelectedRow(filter string, settings *sharedtypes.CoreSettings, status *api.AlertStatus, reminders []api.AlertReminder, cursor int) (Row, bool) {
	rows := FilteredSelectableRows(filter, settings, status, reminders)
	if cursor < 0 || cursor >= len(rows) {
		return Row{}, false
	}
	return rows[cursor], true
}

func GroupedVisibleRows(indices []int, rows []Row, selected int) ([]VisibleRow, int) {
	visible := make([]VisibleRow, 0, len(indices)+4)
	lastSection := ""
	selectedVisible := 0
	selectableIndex := -1
	for _, idx := range indices {
		row := rows[idx]
		if row.Section != "" && row.Section != lastSection {
			visible = append(visible, VisibleRow{Header: true, Text: row.Section, SelectableAt: -1})
			lastSection = row.Section
		}
		if row.Selectable {
			selectableIndex++
		}
		selectableAt := -1
		if row.Selectable {
			selectableAt = selectableIndex
		}
		text := row.Label
		if strings.TrimSpace(row.Value) != "" {
			text += "  " + row.Value
		}
		visible = append(visible, VisibleRow{Text: text, SelectableAt: selectableAt, Key: row.Key})
		if selectableAt == selected {
			selectedVisible = len(visible) - 1
		}
	}
	return visible, selectedVisible
}

func reminderRowKey(id string) RowKey {
	return RowKey("reminder:" + strings.TrimSpace(id))
}

func reminderLabel(reminder api.AlertReminder) string {
	switch reminder.Kind {
	case sharedtypes.AlertReminderKindCheckIn:
		return "Check-In Reminder"
	default:
		return "Reminder"
	}
}

func reminderSummary(reminder api.AlertReminder) string {
	status := "Off"
	if reminder.Enabled {
		status = "On"
	}
	return status + "  " + reminderScheduleSummary(reminder)
}

func reminderScheduleSummary(reminder api.AlertReminder) string {
	switch sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType) {
	case sharedtypes.AlertReminderScheduleWeekly:
		days := weekdayListLabel(reminder.Weekdays)
		if days == "" {
			days = "weekdays"
		}
		return fmt.Sprintf("%s at %s", days, reminder.TimeHHMM)
	default:
		return fmt.Sprintf("daily at %s", reminder.TimeHHMM)
	}
}

func weekdayListLabel(values []int) string {
	names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value >= 0 && value < len(names) {
			out = append(out, names[value])
		}
	}
	return strings.Join(out, ",")
}

func enabledDisabled(v bool) string {
	if v {
		return "Enabled"
	}
	return "Disabled"
}

func supportedLabel(v bool) string {
	if v {
		return "Supported"
	}
	return "Unavailable"
}

func backendLabel(available bool, name string) string {
	if !available {
		return "Unavailable"
	}
	if strings.TrimSpace(name) == "" {
		return "Available"
	}
	return name
}

func urgencyLabel(value sharedtypes.AlertUrgency) string {
	switch sharedtypes.NormalizeAlertUrgency(value) {
	case sharedtypes.AlertUrgencyLow:
		return "Low"
	case sharedtypes.AlertUrgencyHigh:
		return "High"
	default:
		return "Normal"
	}
}

func soundPresetLabel(value sharedtypes.AlertSoundPreset) string {
	switch sharedtypes.NormalizeAlertSoundPreset(value) {
	case sharedtypes.AlertSoundPresetSoftBell:
		return "Soft Bell"
	case sharedtypes.AlertSoundPresetFocusGong:
		return "Focus Gong"
	case sharedtypes.AlertSoundPresetMinimalClick:
		return "Minimal Click"
	default:
		return "Chime"
	}
}
