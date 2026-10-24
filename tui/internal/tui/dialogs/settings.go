package dialogs

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

// stepStrip draws a wizard's numbered steps with the active one selected.
func stepStrip(theme Theme, steps []string, active int) string {
	cells := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf(" %d %s ", i+1, step)
		if i == active {
			cells = append(cells, theme.StyleSelected.Render(label))
		} else {
			cells = append(cells, theme.StyleDim.Render(label))
		}
	}
	return strings.Join(cells, " ")
}

// stepContext is the `n/total Step` label a wizard shows in its border.
func stepContext(steps []string, active int) string {
	return fmt.Sprintf("%d/%d %s", active+1, len(steps), steps[active])
}

func renderDateFormatForm(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Pattern"))
	return renderForm(theme, state, FormSpec{
		Title: "Custom Date Format", FocusCount: 1, Primary: plainHint(dialogSubmitHint(state, "save")),
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Pattern",
					len("Pattern"),
					bareInput(state.Inputs[0], full, "YYYY-MM-DD"),
					true,
				),
			),
			staticRow(
				"",
				theme.StyleDim.Render("Examples: YYYY-MM-DD   Do MMM YYYY   MM/DD/YYYY   [Week] W"),
			),
		},
	})
}

func checkbox(on bool, label string) string {
	if on {
		return "[x] " + label
	}
	return "[ ] " + label
}

func renderRestProtectionDialog(theme Theme, state controllerpkg.State) string {
	steps := []string{"Streaks", "Weekdays", "Dates", "Review"}
	active := min(max(0, state.ProtectionStep), len(steps)-1)
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{
		wide:    true,
		title:   "Rest & Streak Protection",
		context: stepContext(steps, active),
		focus:   -1,
	}
	switch active {
	case 0:
		body = append(body, theme.StyleDim.Render("Select which streaks are protected"))
		labels := []string{}
		for _, kind := range sharedtypes.AvailableStreakKinds() {
			label := "Focus Days"
			if kind == sharedtypes.StreakKindCheckInDays {
				label = "Check-In Days"
			}
			labels = append(labels, checkbox(hasStreakKind(state.ProtectionStreaks, kind), label))
		}
		d.focus = len(body) + state.ProtectionCursor
		body = append(body, selectList(theme, labels, state.ProtectionCursor)...)
		d.primary, d.hints = "tab next", []string{"↑/↓ move", "space toggle", "a all", "c none"}
	case 1:
		body = append(body, theme.StyleDim.Render("Select default rest weekdays"))
		labels := []string{}
		for i, label := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
			labels = append(labels, checkbox(slices.Contains(state.ProtectionWeekdays, i), label))
		}
		d.focus = len(body) + state.ProtectionCursor
		body = append(body, selectList(theme, labels, state.ProtectionCursor)...)
		d.primary, d.hints = "tab next", []string{"↑/↓ move", "space toggle", "c clear"}
	case 2:
		body = append(body, theme.StyleDim.Render("Manage one-off rest dates"))
		if len(state.ProtectionDates) == 0 {
			body = append(body, theme.StyleDim.Render("No rest dates added"))
		}
		d.focus = len(body) + state.ProtectionCursor
		body = append(body, selectList(theme, state.ProtectionDates, state.ProtectionCursor)...)
		d.primary, d.hints = "tab next", []string{"a add date", "D remove selected"}
	default:
		column := len("Weekdays")
		body = append(
			body,
			Field(
				theme,
				"Streaks",
				column,
				fallback(streakKindsSummary(state.ProtectionStreaks), "None"),
				false,
			),
			Field(theme, "Weekdays", column, weekdaysSummary(state.ProtectionWeekdays), false),
			Field(theme, "Dates", column, restDatesSummary(state.ProtectionDates), false),
		)
		d.primary, d.hints = plainHint(dialogSubmitHint(state, "save")), []string{"shift+tab back"}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}

func renderDayBoundaryDialog(theme Theme, state controllerpkg.State) string {
	title := "Start of Day Schedule"
	if state.SettingKey == "endOfDay" {
		title = "End of Day Schedule"
	}
	steps := []string{"Default", "Add Override", "Time"}
	active := min(max(0, state.DayBoundaryStep), len(steps)-1)
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{
		wide:    true,
		title:   title,
		context: fallbackTimezone(state.DayBoundaryTimezone),
		focus:   -1,
	}
	column := len("Enabled")
	switch state.DayBoundaryStep {
	case controllerpkg.DayBoundaryStepDays:
		body = append(body, theme.StyleDim.Render("Select available days"))
		labels, cursor := []string{}, -1
		for day, label := range []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"} {
			day++
			line := checkbox(containsInt(state.DayBoundarySelectedDays, day), label)
			if !controllerpkg.DayBoundaryDayAvailable(state, day) {
				line = theme.StyleDim.Render(line + " (configured)")
			} else if day == state.FocusIdx {
				cursor = day - 1
			}
			labels = append(labels, line)
		}
		if cursor >= 0 {
			d.focus = len(body) + cursor
		}
		body = append(body, selectList(theme, labels, cursor)...)
		d.primary, d.hints, d.esc = "enter next", []string{"↑/↓ move", "space toggle"}, "esc back"
	case controllerpkg.DayBoundaryStepTime:
		full, _ := formColumns(formInner(state.Width), column)
		body = append(
			body,
			Field(
				theme,
				"Days",
				column,
				dayBoundaryDaysLabel(state.DayBoundarySelectedDays),
				false,
			),
			Field(theme, "Time", column, bareInput(state.Inputs[1], full, "HH:mm"), true),
		)
		d.focus = len(body) - 1
		d.primary, d.esc = "ctrl+s apply", "esc back"
	default:
		allOverridden := controllerpkg.AllDayBoundaryDaysOverridden(state.DayBoundarySchedule)
		full, _ := formColumns(formInner(state.Width), column)
		value := bareInput(state.Inputs[0], full, "HH:mm")
		if allOverridden {
			value = theme.StyleDim.Render(state.Inputs[0].Value() + "  (all days have overrides)")
		}
		body = append(body,
			Field(theme, "Enabled", column, enabledLabel(state.DayBoundaryEnabled), false),
			Field(theme, "Default", column, value, state.DayBoundaryOverrideCursor == 0),
			"", theme.StyleDim.Render("Configured overrides"),
		)
		if state.DayBoundaryOverrideCursor == 0 {
			d.focus = len(body) - 3
		}
		if len(state.DayBoundaryOverrides) == 0 {
			body = append(body, theme.StyleDim.Render("None"))
		}
		labels := []string{}
		for _, override := range state.DayBoundaryOverrides {
			labels = append(
				labels,
				fmt.Sprintf("%-22s %s", dayBoundaryDaysLabel(override.Days), override.Time),
			)
		}
		if state.DayBoundaryOverrideCursor > 0 {
			d.focus = len(body) + state.DayBoundaryOverrideCursor - 1
		}
		body = append(body, selectList(theme, labels, state.DayBoundaryOverrideCursor-1)...)
		add := "a add override"
		if allOverridden {
			add = "a add override (unavailable)"
		}
		d.primary, d.hints = "ctrl+s save", []string{"space toggle", add, "enter edit", "D remove"}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}

func renderTelemetrySettingsDialog(theme Theme, state controllerpkg.State) string {
	steps := []string{"Usage", "Diagnostics", "Review"}
	active := min(max(0, state.TelemetryStep), len(steps)-1)
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{
		wide:    true,
		title:   "Privacy & Diagnostics",
		context: stepContext(steps, active),
		focus:   -1,
	}
	switch active {
	case 0:
		body = append(
			body,
			theme.StyleDim.Render(
				"Share anonymous usage signals so Crona can spot problems early and understand which parts of the app people use.",
			),
			theme.StyleDim.Render(
				"This does not include your work content, paths, notes, or anything you type into Crona.",
			),
			"",
		)
		body = append(
			body,
			selectList(
				theme,
				[]string{checkbox(state.TelemetryUsage, "Share usage signals")},
				0,
			)...)
		d.primary, d.hints = "space toggle", []string{"tab next"}
	case 1:
		body = append(
			body,
			theme.StyleDim.Render(
				"Share sanitized diagnostic reports when Crona runs into an error it can explain or recover from.",
			),
			theme.StyleDim.Render(
				"The report is stripped down to help us debug the problem without collecting your work content.",
			),
			"",
		)
		body = append(
			body,
			selectList(theme, []string{checkbox(state.TelemetryErrors, "Share diagnostics")}, 0)...)
		d.primary, d.hints = "space toggle", []string{"tab next", "shift+tab back"}
	default:
		column := len("Usage signals")
		body = append(body,
			Field(theme, "Usage signals", column, telemetryStateLabel(state.TelemetryUsage), false),
			Field(theme, "Diagnostics", column, telemetryStateLabel(state.TelemetryErrors), false),
			"",
			theme.StyleError.Render("Changes take effect after restart."),
			theme.StyleDim.Render("Save now, or restart immediately to apply them right away."),
			"",
		)
		d.focus = len(body) + state.ChoiceCursor
		body = append(
			body,
			selectList(
				theme,
				[]string{"Save", "Save and Restart Now", "Cancel"},
				state.ChoiceCursor,
			)...)
		d.primary, d.hints = "enter choose", []string{"↑/↓ move", "shift+tab back"}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}
