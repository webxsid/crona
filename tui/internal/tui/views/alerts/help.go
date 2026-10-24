package alerts

import (
	sharedtypes "crona/shared/types"
	alertsmeta "crona/tui/internal/tui/views/alertsmeta"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

var (
	toggle = []string{"Enabled", "Disabled"}
	change = []string{"[←/→] change"}
	run    = []string{"[enter] run"}
)

// alertsHelp explains each fixed Alerts row, keyed by its row key.
var alertsHelp = map[alertsmeta.RowKey]viewchrome.RowHelp{
	alertsmeta.RowNotifications: {
		Description: "Desktop notifications when a focus block or break ends.",
		Options:     toggle,
		Keys:        change,
	},
	alertsmeta.RowSound: {
		Description: "Plays a sound with each notification.",
		Options:     toggle,
		Keys:        change,
	},
	alertsmeta.RowSoundPreset: {
		Description: "The sound played with notifications. Sound Effect by Universfield from Pixabay.",
		Options: []string{
			"Chime",
			"Soft Bell",
			"Notification Ping",
			"Focus Gong",
			"Minimal Click",
		},
		Keys: change,
	},
	alertsmeta.RowUrgency: {
		Description: "How insistently the system shows notifications, where the backend supports it.",
		Options:     []string{"Low", "Normal", "High"},
		Keys:        change,
	},
	alertsmeta.RowLogoIcon: {
		Description: "Shows the Crona logo on notifications, where the backend supports it.",
		Options:     toggle,
		Keys:        change,
	},
	alertsmeta.RowInactivityAlerts: {
		Description: "Nudges you when no focus session has run for a while during the day.",
		Options:     toggle,
		Keys:        change,
	},
	alertsmeta.RowInactivityAfter: {
		Description: "How long without a session before the first nudge.",
		Keys:        change,
	},
	alertsmeta.RowInactivityRepeat: {
		Description: "How often the nudge repeats while you stay inactive.",
		Keys:        change,
	},
	alertsmeta.RowAddCheckInReminder: {
		Description: "Adds a daily reminder to fill in your check-in.", Keys: []string{"[enter] add"},
	},
	alertsmeta.RowAddDailyPlanReminder: {
		Description: "Adds a daily reminder to plan the day.", Keys: []string{"[enter] add"},
	},
	alertsmeta.RowTestNotification: {
		Description: "Sends a test notification with the current settings.",
		Keys:        run,
	},
	alertsmeta.RowTestSound: {Description: "Plays the selected sound once.", Keys: run},
}

var reminderKeys = []string{"[space] toggle", "[enter] edit", "[D] delete"}

// reminderHelp explains a reminder row by its kind, since reminders are added at runtime.
var reminderHelp = map[sharedtypes.AlertReminderKind]viewchrome.RowHelp{
	sharedtypes.AlertReminderKindCheckIn: {
		Description: "Reminds you to fill in the day's check-in at the scheduled time.",
		Keys:        reminderKeys,
	},
	sharedtypes.AlertReminderKindDailyPlan: {
		Description: "Reminds you to plan the day at the scheduled time.",
		Keys:        reminderKeys,
	},
}

func helpFor(row alertsmeta.Row) viewchrome.RowHelp {
	if row.ReminderID != "" {
		return reminderHelp[row.ReminderKind]
	}
	return alertsHelp[row.Key]
}
