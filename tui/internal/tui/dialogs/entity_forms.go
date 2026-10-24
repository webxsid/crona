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
		return Field(
			theme,
			label,
			column,
			bareInput(state.Inputs[input], width, placeholder),
			state.FocusIdx == input,
		)
	}
	return renderForm(theme, state, FormSpec{
		Title: title, Context: displayDate, Border: border, FocusCount: 6,
		Primary: plainHint(dialogSubmitHint(state, "save")), Hints: tabNextHints,
		Rows: []FormRow{
			pairRow(field("Mood", 0, half, "1-5"), field("Energy", 1, half, "1-5"), 0, 1, inner),
			pairRow(
				field("Sleep", 2, half, "7h30m"),
				field("Sleep score", 3, half, "0-100"),
				2,
				3,
				inner,
			),
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
		Primary: plainHint(
			dialogSubmitHint(state, "save"),
		), Hints: textareaHints(state.DescriptionIndex),
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Duration",
					column,
					bareInput(state.Inputs[0], full, "optional"),
					state.FocusIdx == 0,
				),
			),
			textareaRow(
				state.DescriptionIndex,
				Textarea(
					theme,
					"Notes",
					column,
					bareTextarea(state),
					state.FocusIdx == state.DescriptionIndex,
					inner,
				),
			),
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
	column := len("Cadence")
	full, _ := formColumns(inner, column)
	return renderForm(theme, state, FormSpec{
		Title: title, Border: border, FocusCount: 3,
		Primary: plainHint(dialogSubmitHint(state, "save")),
		Hints: func(focus int) []string {
			if hints := cadenceHints(state, focus); hints != nil {
				return hints
			}
			return tabNextHints(focus)
		},
		Rows: []FormRow{
			fieldRow(0, cadenceField(theme, state, column, 0)),
			fieldRow(2, daysField(theme, state, column, 2)),
			fieldRow(
				1,
				Field(
					theme,
					"Time",
					column,
					bareInput(state.Inputs[1], full, "HH:MM"),
					state.FocusIdx == 1,
				),
			),
			staticRow("", theme.StyleDim.Render("Time uses 24-hour HH:MM, for example 20:00.")),
		},
	})
}

// cadenceField is the one-row Daily / Weekdays / Custom days picker.
func cadenceField(theme Theme, state controllerpkg.State, column, focus int) string {
	return Picker(theme, PickerSpec{
		Label: "Cadence", Column: column, Focused: state.FocusIdx == focus,
		Options: controllerpkg.CadenceLabels, Selected: state.CadenceIdx,
	})[0]
}

// daysField is the Mo–Su row: ticked days in brackets, the cursor accented while focused,
// and the whole row dim unless Custom days is chosen.
func daysField(theme Theme, state controllerpkg.State, column, focus int) string {
	custom := state.CadenceIdx == len(controllerpkg.CadenceLabels)-1
	focused := custom && state.FocusIdx == focus
	ticked := map[int]bool{}
	for _, day := range state.CadenceDays {
		ticked[day] = true
	}
	cells := []string{}
	for i, name := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		cell := " " + name + " "
		if ticked[(i+1)%7] {
			cell = "[" + name + "]"
		}
		switch {
		case !custom:
			cell = theme.StyleDim.Render(cell)
		case focused && i == state.CadenceCursor:
			cell = theme.StyleCursor.Render(cell)
		}
		cells = append(cells, cell)
	}
	return Field(theme, "Days", column, strings.Join(cells, ""), focused)
}

// cadenceHints are the extra keys while the cadence picker or Days row is focused.
func cadenceHints(state controllerpkg.State, focus int) []string {
	cadence, days, _, _ := controllerpkg.CadenceLayout(state)
	switch focus {
	case cadence:
		return []string{"tab next", "←/→ choose"}
	case days:
		return []string{"tab next", "←/→ move", "space toggle"}
	}
	return nil
}
