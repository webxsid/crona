package dialogs

import (
	"fmt"
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

func renderSessionDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "end_session":
		title := "End Session"
		hint := dialogSubmitHint(state, "confirm") + "   [ctrl+e] open details dialog   [esc] cancel"
		labels := []string{
			"Commit message",
			"Worked on",
			"Outcome",
			"Next step",
			"Blockers",
			"Links",
		}
		rows := []string{theme.StylePaneTitle.Render(title)}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "timer_start_type":
		rows := []string{theme.StylePaneTitle.Render("Start Timer")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		for i, item := range state.ChoiceItems {
			line := timerStartChoiceLabel(i, item)
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose   [esc] cancel")
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	case "pomodoro_start":
		vm := controllerpkg.BuildPomodoroDialogViewModel(state)
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if totalPreview := vm.EstimatedTotalDuration; totalPreview != "" {
			rows = append(rows, "", theme.StyleSelected.Render(totalPreview))
		}
		if vm.ShowSummary {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render(
					fmt.Sprintf(
						"%s  ·  %s  ·  %s",
						vm.FocusDisplay,
						vm.ShortBreakDisplay,
						vm.LongBreakDisplay,
					),
				),
			)
			rows = append(rows, theme.StyleDim.Render(
				vm.CyclesSummary,
			))
		}
		rows = append(rows, "", pomodoroRowLabel(theme, "Focus", vm.FocusRowActive || vm.FocusCustomActive))
		rows = append(rows, pomodoroChoiceRow(
			theme,
			vm.FocusChoices,
			state.PomodoroFocusChoice,
			vm.FocusCustomChoice,
			vm.FocusCustomActive,
			state.Inputs[0].View(),
		))
		rows = append(rows, "", pomodoroRowLabel(theme, "Short Break", vm.BreakRowActive || vm.BreakCustomActive))
		rows = append(rows, pomodoroChoiceRow(
			theme,
			vm.BreakChoices,
			state.PomodoroBreakChoice,
			vm.BreakCustomChoice,
			vm.BreakCustomActive,
			state.Inputs[1].View(),
		))
		if vm.LongBreakForcedOff {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break", vm.LongBreakRowActive || vm.LongBreakCustomActive))
			rows = append(rows, pomodoroChoiceRow(
				theme,
				vm.LongBreakChoices,
				vm.LongBreakSelected,
				vm.LongBreakCustomChoice,
				vm.LongBreakCustomActive,
				state.Inputs[2].View(),
			))
		}
		if vm.CyclesDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Cycles"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Cycles", vm.CyclesRowActive))
			rows = append(rows, state.Inputs[3].View())
		}
		if vm.LongBreakDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break After", vm.LongBreakCyclesActive))
			rows = append(rows, state.Inputs[4].View())
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[←/→] choose   [↑/↓] move   "+dialogSubmitHint(state, "start")+"   [esc] back",
		)
		return modal(theme, state.Width, 72, theme.ColorGreen, rows)
	case "hard_limit_expired":
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Session Complete")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		rows = append(rows, "", theme.StyleDim.Render("Choose how to finish this pomodoro session."))
		for i, item := range state.ChoiceItems {
			line := "  " + item
			if i == state.ChoiceCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
			} else {
				rows = append(rows, theme.StyleNormal.Render(line))
			}
		}
		if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
			rows = append(rows, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
		}
		rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] choose")
		return modal(theme, state.Width, 68, theme.ColorYellow, rows)
	case "hard_limit_extend":
		vm := controllerpkg.BuildPomodoroDialogViewModel(state)
		rows := []string{theme.StylePaneTitle.Render("Extend Pomodoro Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if totalPreview := controllerpkg.PreviewPomodoroExtendDuration(state); totalPreview != "" {
			rows = append(rows, "", theme.StyleSelected.Render(totalPreview))
		}
		if vm.ShowSummary {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render(
					fmt.Sprintf(
						"%s  ·  %s  ·  %s",
						vm.FocusDisplay,
						vm.ShortBreakDisplay,
						vm.LongBreakDisplay,
					),
				),
			)
			rows = append(rows, theme.StyleDim.Render(vm.CyclesSummary))
		}
		rows = append(rows, "", pomodoroRowLabel(theme, "Focus", vm.FocusRowActive || vm.FocusCustomActive))
		rows = append(rows, pomodoroChoiceRow(theme, vm.FocusChoices, state.PomodoroFocusChoice, vm.FocusCustomChoice, vm.FocusCustomActive, state.Inputs[0].View()))
		rows = append(rows, "", pomodoroRowLabel(theme, "Short Break", vm.BreakRowActive || vm.BreakCustomActive))
		rows = append(rows, pomodoroChoiceRow(theme, vm.BreakChoices, state.PomodoroBreakChoice, vm.BreakCustomChoice, vm.BreakCustomActive, state.Inputs[1].View()))
		if vm.LongBreakForcedOff {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break", vm.LongBreakRowActive || vm.LongBreakCustomActive))
			rows = append(rows, pomodoroChoiceRow(theme, vm.LongBreakChoices, vm.LongBreakSelected, vm.LongBreakCustomChoice, vm.LongBreakCustomActive, state.Inputs[2].View()))
		}
		if vm.CyclesDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Cycles"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Cycles", vm.CyclesRowActive))
			rows = append(rows, state.Inputs[3].View())
		}
		if vm.LongBreakDisabled {
			rows = append(rows, "", pomodoroDisabledValueRow(theme, "Long Break"))
		} else {
			rows = append(rows, "", pomodoroRowLabel(theme, "Long Break After", vm.LongBreakCyclesActive))
			rows = append(rows, state.Inputs[4].View())
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[←/→] choose   [↑/↓] move   [enter] extend   [esc] back",
		)
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "amend_session":
		rows := []string{
			theme.StylePaneTitle.Render("Amend Session"),
			"",
			theme.StyleDim.Render("Commit message"),
			state.Inputs[0].View(),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "save")+"   [esc] cancel",
		)
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
	case "manual_session":
		rows := []string{theme.StylePaneTitle.Render("Log Session")}
		if state.IssueID != 0 {
			label := "Issue #" + itoa(state.IssueID)
			if state.ViewName != "" {
				label += "  " + state.ViewName
			}
			if state.IssueEstimateMins != nil && *state.IssueEstimateMins > 0 {
				label += fmt.Sprintf(
					"  · estimate %s",
					controllerpkg.FormatDurationMinutesInput(state.IssueEstimateMins),
				)
			}
			rows = append(rows, "", theme.StyleDim.Render(label))
		}
		labels := []string{
			"Summary",
			"Date",
			"Work duration",
			"Break duration",
			"Start time",
			"End time",
			"Notes",
		}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = appendDialogFooter(theme, state, rows, manualSessionHint(state))
		return modal(theme, state.Width, 72, theme.ColorGreen, rows)
	case "issue_session_transition":
		title := "End Session?"
		body := "Mark this issue and end the active session?"
		hint := "[y/enter] confirm   [n/esc] cancel"
		border := theme.ColorYellow
		switch state.IssueStatus {
		case "done":
			title, body, border = "Complete Issue", "Mark the issue done and end the active session.", theme.ColorGreen
		case "abandoned":
			title, body, border = "Abandon Issue", "Abandon the issue and end the active session.", theme.ColorRed
		}
		rows := []string{theme.StylePaneTitle.Render(title), "", body}
		if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") &&
			len(state.Inputs) > 0 {
			hint = dialogSubmitHint(state, "confirm") + "   [esc] cancel"
			label := "Abandon reason"
			if state.IssueStatus == "done" {
				label = "Completion note"
			}
			rows = append(rows, "", theme.StyleDim.Render(label), state.Inputs[0].View())
		}
		rows = appendDialogFooter(theme, state, rows, hint)
		return modal(theme, state.Width, 68, border, rows)
	default:
		return ""
	}
}

func timerStartChoiceLabel(idx int, item string) string {
	switch idx {
	case 0:
		return "[s] " + item
	case 1:
		return "[p] " + item
	default:
		return item
	}
}

func pomodoroChoiceRow(
	theme Theme,
	choices []string,
	selected int,
	customIdx int,
	customActive bool,
	customView string,
) string {
	parts := make([]string, 0, len(choices)+1)
	for i, choice := range choices {
		style := theme.StyleNormal
		if i == selected {
			style = theme.StyleCursor
		}
		parts = append(parts, style.Render(choice))
	}
	if selected == customIdx {
		style := theme.StyleNormal
		if customActive {
			style = theme.StyleCursor
		}
		parts = append(parts, style.Render(customView))
	}
	return strings.Join(parts, "   ")
}

func pomodoroRowLabel(theme Theme, label string, active bool) string {
	if active {
		return theme.StyleCursor.Render("> " + label)
	}
	return theme.StyleDim.Render("  " + label)
}

func pomodoroDisabledValueRow(theme Theme, label string) string {
	return theme.StyleDim.Render(label + ": disabled")
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

func manualSessionHint(state controllerpkg.State) string {
	switch state.FocusIdx {
	case 1:
		return "[ctrl+e] pick date   [g] today   [tab] next   " + dialogSubmitHint(
			state,
			"save",
		) + "   [esc] cancel"
	default:
		return "[tab] next   durations: 90, 90m, 1h30m   " + dialogSubmitHint(
			state,
			"save",
		) + "   [esc] cancel"
	}
}
