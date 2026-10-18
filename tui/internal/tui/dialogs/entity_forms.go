package dialogs

import (
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func tabNextHints(int) []string { return []string{"tab next"} }

// renderCheckInForm draws New/Edit Check-In: two pairs, then Screen time and Notes.
func renderCheckInForm(theme Theme, state controllerpkg.State, displayDate string) string {
	title, border := "New Check-In", theme.ColorCyan
	if state.Kind == "edit_checkin" {
		title, border = "Edit Check-In", theme.ColorYellow
	}
	inner := formInner(state.Width)
	column := len("Screen time")
	full, half := formColumns(inner, column)
	field := func(label string, input int, width int, placeholder string) string {
		return Field(theme, label, column, bareInput(state.Inputs[input], width, placeholder), state.FocusIdx == input)
	}
	return renderForm(theme, state, FormSpec{
		Title: title, Context: displayDate, Border: border, FocusCount: 6,
		Primary: plainHint(dialogSubmitHint(state, "save")), Hints: tabNextHints,
		Rows: []FormRow{
			pairRow(field("Mood", 0, half, "1-5"), field("Energy", 1, half, "1-5"), 0, 1, inner),
			pairRow(field("Sleep", 2, half, "7h30m"), field("Sleep score", 3, half, "0-100"), 2, 3, inner),
			fieldRow(4, field("Screen time", 4, full, "2h")),
			fieldRow(5, field("Notes", 5, full, "optional")),
		},
	})
}

// renderHabitLogForm draws a habit completion: Duration, then a Notes textarea.
func renderHabitLogForm(theme Theme, state controllerpkg.State) string {
	title := "Habit Log"
	if strings.TrimSpace(state.ViewTitle) != "" {
		title = state.ViewTitle
	}
	inner := formInner(state.Width)
	column := len("Duration")
	full, _ := formColumns(inner, column)
	return renderForm(theme, state, FormSpec{
		Title: title, Context: state.CheckInDate, Border: theme.ColorGreen, FocusCount: 2,
		Primary: plainHint(dialogSubmitHint(state, "save")), Hints: textareaHints(state.DescriptionIndex),
		Rows: []FormRow{
			fieldRow(0, Field(theme, "Duration", column, bareInput(state.Inputs[0], full, "optional"), state.FocusIdx == 0)),
			textareaRow(state.DescriptionIndex, Textarea(theme, "Notes", column, bareTextarea(state), state.FocusIdx == state.DescriptionIndex, inner)),
		},
	})
}

// renderReminderForm draws Add/Edit Reminder: Schedule and Time, then the format help.
func renderReminderForm(theme Theme, state controllerpkg.State) string {
	plan := state.ReminderKind == sharedtypes.AlertReminderKindDailyPlan
	title, border := "Add Check-In Reminder", theme.ColorCyan
	if plan {
		title = "Add Plan-the-Day Reminder"
	}
	if state.Kind == "edit_alert_reminder" {
		title, border = "Edit Check-In Reminder", theme.ColorYellow
		if plan {
			title = "Edit Plan-the-Day Reminder"
		}
	}
	inner := formInner(state.Width)
	column := len("Schedule")
	_, half := formColumns(inner, column)
	timeColumn := len("Time")
	if inner < 56 {
		timeColumn = column
	}
	return renderForm(theme, state, FormSpec{
		Title: title, Border: border, FocusCount: 2,
		Primary: plainHint(dialogSubmitHint(state, "save")), Hints: tabNextHints,
		Rows: []FormRow{
			pairRow(
				Field(theme, "Schedule", column, bareInput(state.Inputs[0], half, "daily"), state.FocusIdx == 0),
				Field(theme, "Time", timeColumn, bareInput(state.Inputs[1], half, "HH:MM"), state.FocusIdx == 1),
				0, 1, inner,
			),
			staticRow("",
				theme.StyleDim.Render("Use daily, weekdays, or weekday lists like mon,wed,fri."),
				theme.StyleDim.Render("Time uses 24-hour HH:MM, for example 20:00."),
			),
		},
	})
}
