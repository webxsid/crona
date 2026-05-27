package dialogs

import (
	"fmt"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

func renderSessionDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "stash_list":
		rows := []string{theme.StylePaneTitle.Render("Stashes"), ""}
		if len(state.Stashes) == 0 {
			rows = append(rows, theme.StyleDim.Render("No stashes available"))
		} else {
			for i, stash := range state.Stashes {
				if i == state.StashCursor {
					rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+stash.Label))
				} else {
					rows = append(rows, theme.StyleNormal.Render("  "+stash.Label))
				}
				rows = append(rows, theme.StyleDim.Render("  "+stash.Meta))
			}
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[j/k] move   [enter] pop   [x] drop   [esc] cancel",
		)
		return modal(theme, state.Width, 60, theme.ColorYellow, rows)
	case "end_session", "stash_session":
		title := "End Session"
		hint := dialogSubmitHint(state, "confirm") + "   [ctrl+e] details   [esc] cancel"
		labels := []string{
			"Commit message",
			"Worked on",
			"Outcome",
			"Next step",
			"Blockers",
			"Links",
		}
		if state.Kind == "stash_session" {
			title = "Stash Session"
			hint = dialogSubmitHint(state, "confirm") + "   [esc] cancel"
			labels = []string{"Stash note"}
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
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] select   [esc] cancel")
		return modal(theme, state.Width, 68, theme.ColorGreen, rows)
	case "pomodoro_focus_presets":
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Focus Presets")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
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
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] select   [esc] back")
		return modal(theme, state.Width, 70, theme.ColorGreen, rows)
	case "pomodoro_focus_custom":
		rows := []string{
			theme.StylePaneTitle.Render("Custom Focus Duration"),
			"",
			theme.StyleDim.Render("Focus duration"),
			state.Inputs[0].View(),
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "continue")+"   [esc] back")
		return modal(theme, state.Width, 64, theme.ColorGreen, rows)
	case "pomodoro_break_presets":
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Break Presets")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
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
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] select   [esc] back")
		return modal(theme, state.Width, 70, theme.ColorGreen, rows)
	case "pomodoro_break_custom":
		rows := []string{
			theme.StylePaneTitle.Render("Custom Break Duration"),
			"",
			theme.StyleDim.Render("Break duration"),
			state.Inputs[0].View(),
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "continue")+"   [esc] back")
		return modal(theme, state.Width, 64, theme.ColorGreen, rows)
	case "pomodoro_start":
		rows := []string{theme.StylePaneTitle.Render("Pomodoro Session")}
		if state.ViewName != "" {
			rows = append(rows, "", theme.StyleDim.Render(state.ViewName))
		}
		if state.PomodoroFocusSeconds > 0 || state.PomodoroBreakSeconds > 0 {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render(
					fmt.Sprintf(
						"Focus %s  ·  Break %s",
						helperpkg.FormatCompactDurationSeconds(state.PomodoroFocusSeconds),
						helperpkg.FormatCompactDurationSeconds(state.PomodoroBreakSeconds),
					),
				),
			)
		}
		labels := []string{
			"Total duration",
			"Short breaks until long break",
			"Long break duration",
		}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[tab] next   "+dialogSubmitHint(state, "start")+"   [esc] back",
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
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] select")
		return modal(theme, state.Width, 68, theme.ColorYellow, rows)
	case "hard_limit_extend":
		rows := []string{
			theme.StylePaneTitle.Render("Extend Pomodoro Session"),
			"",
			theme.StyleDim.Render("Add more time to the active pomodoro session."),
		}
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
		rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] select   [esc] back")
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
	case "hard_limit_extend_custom":
		rows := []string{
			theme.StylePaneTitle.Render("Custom Extension"),
			"",
			theme.StyleDim.Render("Extension duration"),
			state.Inputs[0].View(),
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, "extend")+"   [esc] back",
		)
		return modal(theme, state.Width, 64, theme.ColorCyan, rows)
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

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

func manualSessionHint(state controllerpkg.State) string {
	switch state.FocusIdx {
	case 1:
		return "[f2] pick date   [g] today   [tab] next   " + dialogSubmitHint(
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
