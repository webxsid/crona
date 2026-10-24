package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

// timerMetaRows splits renderTimerMeta's rows into body lines.
func timerMetaRows(
	theme Theme,
	state controllerpkg.State,
	totalLabel string,
	totalSeconds int,
	endsAt bool,
) []string {
	meta := renderTimerMeta(theme, state, totalLabel, totalSeconds, endsAt)
	if meta == "" {
		return nil
	}
	return append(strings.Split(meta, "\n"), "")
}

func renderEndSession(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	column := len("Worked on")
	full, _ := formColumns(inner, column)
	labels := []string{"Commit", "Worked on", "Outcome", "Next step", "Blockers", "Links"}
	placeholders := []string{"required", "optional", "optional", "optional", "optional", "optional"}
	primary := plainHint(dialogSubmitHint(state, "confirm"))
	rows, context := []FormRow{}, ""
	if state.AdhocTimer {
		labels[0] = "Summary"
		context = adhocElapsedLabel(state)
		if state.AdhocResolution == sharedtypes.AdhocSessionResolutionKeep {
			rows = append(rows, staticRow(
				Field(
					theme,
					"Records to",
					column,
					theme.StyleHeader.Render("Ad hoc / Sessions"),
					false,
				),
				theme.StyleDim.Render(
					"Your summary becomes the issue title, shortened if needed. The issue is marked Done and assigned today.",
				),
				"",
			))
			primary = "ctrl+s save ad hoc & finish"
		} else {
			rows = append(rows, staticRow(
				Field(theme, "Attach to", column, theme.StyleHeader.Render(state.ViewName), false),
				Field(theme, "", column, theme.StyleDim.Render(state.RepoName+" / "+state.StreamName), false),
				"",
			))
			primary = "ctrl+s attach & finish"
		}
	}
	for i := range state.Inputs {
		if i >= len(labels) {
			break
		}
		rows = append(
			rows,
			fieldRow(
				i,
				Field(
					theme,
					labels[i],
					column,
					bareInput(state.Inputs[i], full, placeholders[i]),
					state.FocusIdx == i,
				),
			),
		)
	}
	return renderForm(theme, state, FormSpec{
		Title: "End Session", Context: context, Rows: rows, Primary: primary, FocusCount: len(state.Inputs),
		Hints: func(int) []string { return []string{"tab next", "ctrl+e details"} },
	})
}

// renderTwoLineChoices draws a list whose items each carry a dim detail line.
func renderTwoLineChoices(theme Theme, state controllerpkg.State, body []string) ([]string, int) {
	focus := -1
	for i, item := range state.ChoiceItems {
		if i > 0 {
			body = append(body, "")
		}
		cursor := -1
		if i == state.ChoiceCursor {
			focus, cursor = len(body), 0
		}
		body = append(body, selectList(theme, []string{item}, cursor)...)
		if i < len(state.ChoiceDetails) && state.ChoiceDetails[i] != "" {
			body = append(body, theme.StyleDim.Render("    "+state.ChoiceDetails[i]))
		}
	}
	return body, focus
}

func renderAdhocResolution(theme Theme, state controllerpkg.State) string {
	body := []string{}
	if state.ViewName != "" {
		body = append(body, theme.StyleDim.Render(state.ViewName), "")
	}
	body, focus := renderTwoLineChoices(theme, state, body)
	if len(state.ChoiceItems) == 0 {
		body = append(body, theme.StyleDim.Render("No eligible issues are available."))
	}
	return renderListDialog(theme, state, listDialog{
		title: state.ViewTitle, context: adhocElapsedLabel(state), body: body, focus: focus,
		primary: "enter choose", hints: []string{"↑/↓ move"}, esc: "esc back",
	})
}

func renderAdhocNewIssue(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Title"))
	note := theme.StyleDim.Render("Issue will be created in the active repo and stream.")
	if state.RepoID <= 0 || state.StreamID <= 0 {
		note = theme.StyleError.Render("Select an active repo and stream first.")
	}
	rows := []FormRow{}
	if state.ViewName != "" {
		rows = append(rows, staticRow(theme.StyleDim.Render(state.ViewName)))
	}
	rows = append(
		rows,
		staticRow(note, ""),
		fieldRow(
			0,
			Field(theme, "Title", len("Title"), bareInput(state.Inputs[0], full, "required"), true),
		),
	)
	return renderForm(
		theme,
		state,
		FormSpec{Title: state.ViewTitle, Rows: rows, Primary: "enter continue", FocusCount: 1},
	)
}

func renderTimerStartType(theme Theme, state controllerpkg.State) string {
	body := timerMetaRows(theme, state, "", 0, false)
	labels := make([]string, 0, len(state.ChoiceItems))
	for i, item := range state.ChoiceItems {
		labels = append(labels, timerStartChoiceLabel(i, item))
	}
	focus := len(body) + state.ChoiceCursor
	body = append(body, selectList(theme, labels, state.ChoiceCursor)...)
	if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
		body = append(body, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
	}
	return renderListDialog(theme, state, listDialog{
		title: "Start Timer", context: state.ViewName, border: theme.ColorGreen, body: body, focus: focus,
		primary: "enter choose", hints: []string{"↑/↓ move"},
	})
}

// renderPomodoro draws Pomodoro Session and Extend Pomodoro Session: preset rows for
// focus and breaks, then cycles.
func renderPomodoro(theme Theme, state controllerpkg.State) string {
	vm := controllerpkg.BuildPomodoroDialogViewModel(state)
	extend := state.Kind == "hard_limit_extend"
	column := len("Short break")
	body := []string{}
	if extend {
		if preview := controllerpkg.PreviewPomodoroExtendDuration(state); preview != "" {
			body = append(body, theme.StyleSelected.Render(preview), "")
		}
	} else {
		body = append(body, timerMetaRows(theme, state, vm.EstimatedTotalDuration, vm.EstimatedTotalSeconds, true)...)
	}
	if vm.ShowSummary {
		body = append(
			body,
			theme.StyleDim.Render(
				fmt.Sprintf(
					"%s  ·  %s  ·  %s",
					vm.FocusDisplay,
					vm.ShortBreakDisplay,
					vm.LongBreakDisplay,
				),
			),
			theme.StyleDim.Render(vm.CyclesSummary),
			"",
		)
	}
	focus := -1
	add := func(label, value string, active bool) {
		if active {
			focus = len(body)
		}
		body = append(body, Field(theme, label, column, value, active))
	}
	disabled := theme.StyleDim.Render("disabled")
	add(
		"Focus",
		pomodoroChoiceRow(
			theme,
			vm.FocusChoices,
			state.PomodoroFocusChoice,
			vm.FocusCustomChoice,
			vm.FocusCustomActive,
			bareInput(state.Inputs[0], 8, "25m"),
		),
		vm.FocusRowActive || vm.FocusCustomActive,
	)
	add(
		"Short break",
		pomodoroChoiceRow(
			theme,
			vm.BreakChoices,
			state.PomodoroBreakChoice,
			vm.BreakCustomChoice,
			vm.BreakCustomActive,
			bareInput(state.Inputs[1], 8, "5m"),
		),
		vm.BreakRowActive || vm.BreakCustomActive,
	)
	if vm.LongBreakForcedOff {
		add("Long break", disabled, false)
	} else {
		add("Long break", pomodoroChoiceRow(theme, vm.LongBreakChoices, vm.LongBreakSelected, vm.LongBreakCustomChoice, vm.LongBreakCustomActive, bareInput(state.Inputs[2], 8, "15m")), vm.LongBreakRowActive || vm.LongBreakCustomActive)
	}
	if vm.CyclesDisabled {
		add("Cycles", disabled, false)
	} else {
		add("Cycles", bareInput(state.Inputs[3], 8, "4"), vm.CyclesRowActive)
	}
	if vm.LongBreakDisabled {
		add("Long after", disabled, false)
	} else {
		add("Long after", bareInput(state.Inputs[4], 3, "4")+theme.StyleDim.Render(" cycles"), vm.LongBreakCyclesActive)
	}
	d := listDialog{
		title:   "Pomodoro Session",
		context: state.ViewName,
		border:  theme.ColorGreen,
		body:    body,
		focus:   focus,
		primary: plainHint(
			dialogSubmitHint(state, "start"),
		),
		hints: []string{"←/→ choose", "↑/↓ move"},
		esc:   "esc back",
		wide:  true,
	}
	if extend {
		d.title, d.context, d.border, d.primary = "Extend Pomodoro Session", "", "", "enter extend"
	}
	return renderListDialog(theme, state, d)
}

// renderCountdownForm draws Timer Session and Extend Timer Session: one duration field.
func renderCountdownForm(theme Theme, state controllerpkg.State) string {
	column := len("Focus")
	title, label, help, primary, border := "Timer Session", "Focus", "Single countdown. No breaks. Reuses the hard-limit session completion flow.", "start", theme.ColorGreen
	rows := []FormRow{}
	if state.Kind == "timer_countdown_extend" {
		title, label, help, primary, border = "Extend Timer Session", "Add", "Adds time to the current countdown. No breaks or cycles.", "extend", ""
	} else if meta := timerMetaRows(theme, state, "", state.TimerCountdownSeconds, true); meta != nil {
		rows = append(rows, staticRow(meta...))
	}
	full, _ := formColumns(formInner(state.Width), column)
	rows = append(rows,
		fieldRow(0, Field(theme, label, column, bareInput(state.Inputs[0], full, "25m"), true)),
		staticRow("", theme.StyleDim.Render(help)),
	)
	context := state.ViewName
	if state.Kind == "timer_countdown_extend" {
		context = ""
	}
	return renderForm(theme, state, FormSpec{
		Title: title, Context: context, Border: border, Rows: rows, FocusCount: 1,
		Primary: plainHint(dialogSubmitHint(state, primary)),
	})
}

func renderHardLimitExpired(theme Theme, state controllerpkg.State) string {
	title, description := "Pomodoro Session Complete", "Choose how to finish this pomodoro session."
	if state.HardLimitKind == sharedtypes.TimerHardLimitKindCountdown {
		title, description = "Timer Session Complete", "Choose how to finish this timer session."
	}
	body := []string{theme.StyleDim.Render(description), ""}
	focus := len(body) + state.ChoiceCursor
	body = append(body, selectList(theme, state.ChoiceItems, state.ChoiceCursor)...)
	if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
		body = append(body, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
	}
	inner := FrameInner(state.Width, body)
	// The session has to end one way or another, so this dialog has no esc.
	return Frame(theme, FrameSpec{
		Title: title, Context: state.ViewName, Border: theme.ColorYellow, Body: body,
		Footer: withError(
			theme,
			state,
			footerWithEsc(theme, "enter choose", []string{"↑/↓ move"}, "", inner),
		),
		ScreenWidth: state.Width, ScreenHeight: state.Height, Focus: focus, Inner: inner,
	})
}

func renderAmendSession(theme Theme, state controllerpkg.State) string {
	full, _ := formColumns(formInner(state.Width), len("Commit"))
	return renderForm(theme, state, FormSpec{
		Title: "Amend Session", FocusCount: 1, Primary: plainHint(dialogSubmitHint(state, "save")),
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Commit",
					len("Commit"),
					bareInput(state.Inputs[0], full, "required"),
					true,
				),
			),
		},
	})
}

func renderManualSession(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	column := len("Summary")
	full, half := formColumns(inner, column)
	context := state.ViewName
	if state.IssueEstimateMins != nil && *state.IssueEstimateMins > 0 {
		context += " · est. " + controllerpkg.FormatDurationMinutesInput(state.IssueEstimateMins)
	}
	field := func(label string, i, width int, placeholder string) string {
		return Field(
			theme,
			label,
			column,
			bareInput(state.Inputs[i], width, placeholder),
			state.FocusIdx == i,
		)
	}
	rows := []FormRow{
		fieldRow(0, field("Summary", 0, full, "optional")),
		fieldRow(1, field("Date", 1, full, "YYYY-MM-DD")),
		pairRow(field("Work", 2, half, "90m"), field("Break", 3, half, "optional"), 2, 3, inner),
		pairRow(field("Start", 4, half, "HH:mm"), field("End", 5, half, "HH:mm"), 4, 5, inner),
	}
	if len(state.Inputs) > 6 {
		rows = append(rows, fieldRow(6, field("Notes", 6, full, "optional")))
	}
	return renderForm(theme, state, FormSpec{
		Title: "Log Session", Context: strings.TrimSpace(context), Border: theme.ColorGreen, Rows: rows,
		FocusCount: len(state.Inputs), Primary: plainHint(dialogSubmitHint(state, "save")),
		Hints: func(focus int) []string {
			if focus == 1 {
				return []string{"tab next", "ctrl+e pick date", "g today"}
			}
			return []string{"tab next", "durations: 90, 90m, 1h30m"}
		},
	})
}

func renderIssueSessionTransition(theme Theme, state controllerpkg.State) string {
	title, message, border := "End Session?", "Mark this issue and end the active session?", theme.ColorYellow
	switch state.IssueStatus {
	case "done":
		title, message, border = "Complete Issue", "Mark the issue done and end the active session.", theme.ColorGreen
	case "abandoned":
		title, message, border = "Abandon Issue", "Abandon the issue and end the active session.", theme.ColorRed
	}
	rows := []FormRow{staticRow(message)}
	if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") && len(state.Inputs) > 0 {
		label := "Reason"
		if state.IssueStatus == "done" {
			label = "Note"
		}
		full, _ := formColumns(formInner(state.Width), len("Reason"))
		rows = append(
			rows,
			staticRow(""),
			fieldRow(
				0,
				Field(
					theme,
					label,
					len("Reason"),
					bareInput(state.Inputs[0], full, "optional"),
					true,
				),
			),
		)
		return renderForm(
			theme,
			state,
			FormSpec{
				Title:      title,
				Border:     border,
				Rows:       rows,
				FocusCount: 1,
				Primary:    plainHint(dialogSubmitHint(state, "confirm")),
			},
		)
	}
	body, _ := formBody(rows, -1)
	return renderListDialog(
		theme,
		state,
		listDialog{
			title:   title,
			border:  border,
			body:    body,
			focus:   -1,
			primary: "y confirm",
			hints:   []string{"enter confirm"},
			esc:     "n/esc cancel",
		},
	)
}

// renderEditSessionHistory draws Edit Session History: start and end, notes, then the
// segment rows (type, start → end). FocusIdx 0–1 are the times, 2–4 the notes, and each
// segment takes three stops from 5.
func renderEditSessionHistory(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	column := len("Work summary")
	_, half := formColumns(inner, column)
	rows := []FormRow{}
	if state.SessionHistoryContext != "" {
		rows = append(rows, staticRow(theme.StyleDim.Render(state.SessionHistoryContext), ""))
	}
	if len(state.Inputs) >= 2 {
		rows = append(rows, pairRow(
			Field(
				theme,
				"Start",
				column,
				bareInput(state.Inputs[0], half, "YYYY-MM-DD HH:mm:ss"),
				state.FocusIdx == 0,
			),
			Field(
				theme,
				"End",
				column,
				bareInput(state.Inputs[1], half, "YYYY-MM-DD HH:mm:ss"),
				state.FocusIdx == 1,
			),
			0,
			1,
			inner,
		))
	}
	for i, label := range []string{"Commit", "Work summary", "Notes"} {
		if i < len(state.SessionHistoryNoteInputs) {
			rows = append(
				rows,
				textareaRow(
					2+i,
					Textarea(
						theme,
						label,
						column,
						state.SessionHistoryNoteInputs[i].View(),
						state.FocusIdx == 2+i,
						inner,
					),
				),
			)
		}
	}
	rows = append(rows, staticRow("", theme.StyleDim.Render("Work and break segments")))
	for i, segment := range state.SessionHistorySegments {
		first := 5 + i*3
		rows = append(rows, FormRow{
			Lines: []string{
				historySegmentTypeRow(theme, state, segment, i),
				"    " + segment.StartTime.View() + "  →  " + segment.EndTime.View(),
			},
			Focus: map[int]int{first: 0, first + 1: 1, first + 2: 1},
		})
	}
	if len(state.SessionHistorySegments) == 0 {
		rows = append(
			rows,
			staticRow(theme.StyleDim.Render("No segments. Press ctrl+a to add one.")),
		)
	}
	return renderForm(theme, state, FormSpec{
		Title: "Edit Session History", Rows: rows, Primary: plainHint(dialogSubmitHint(state, "save")),
		Hints: func(int) []string {
			return []string{"tab next", "ctrl+a add segment", "ctrl+d remove", "←/→ change type"}
		},
	})
}
