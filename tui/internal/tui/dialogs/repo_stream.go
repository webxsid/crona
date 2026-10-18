package dialogs

import (
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
)

func renderRepoStreamDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "create_repo":
		return renderNameDescriptionForm(theme, state, "New Repo", "", "", "create")
	case "edit_repo":
		return renderNameDescriptionForm(theme, state, "Edit Repo", "", theme.ColorYellow, "save")
	case "create_stream":
		return renderNameDescriptionForm(theme, state, "New Stream", state.RepoName, "", "create")
	case "edit_stream":
		return renderNameDescriptionForm(theme, state, "Edit Stream", state.RepoName, theme.ColorYellow, "save")
	case "checkout_context":
		return renderCheckoutContext(theme, state)
	case "create_habit":
		return renderHabitForm(theme, state, true)
	case "edit_habit":
		return renderHabitForm(theme, state, false)
	default:
		return ""
	}
}

// renderHabitForm draws New Habit (with Repo/Stream pickers at the bottom) and Edit Habit.
func renderHabitForm(theme Theme, state controllerpkg.State, create bool) string {
	inner := formInner(state.Width)
	column := len("Description")
	full, _ := formColumns(inner, column)
	name, schedule, target := 0, 1, 2
	if create {
		name, schedule, target = 2, 3, 4
	}
	focusOf := func(input int) int {
		if input >= state.DescriptionIndex {
			return input + 1
		}
		return input
	}
	focused := func(input int) bool { return state.FocusIdx == focusOf(input) }
	_, days, count, _ := controllerpkg.CadenceLayout(state)
	rows := []FormRow{
		fieldRow(focusOf(name), Field(theme, "Name", column, bareInput(state.Inputs[name], full, "required"), focused(name))),
		textareaRow(state.DescriptionIndex, Textarea(theme, "Description", column, bareTextarea(state), state.FocusIdx == state.DescriptionIndex, inner)),
		fieldRow(focusOf(schedule), cadenceField(theme, state, column, focusOf(schedule))),
		fieldRow(days, daysField(theme, state, column, days)),
		fieldRow(focusOf(target), Field(theme, "Target", column, bareInput(state.Inputs[target], full, "optional"), focused(target))),
	}
	title, border, submit := "Edit Habit", lipgloss.Color(theme.ColorYellow), "save"
	if create {
		title, border, submit = "New Habit", "", "create"
		rows = append(rows, staticRow(""), repoPickerRow(theme, state, column, full, 0), streamPickerRow(theme, state, column, full, 1))
	}
	textarea := textareaHints(state.DescriptionIndex)
	return renderForm(theme, state, FormSpec{
		Title: title, Context: repoStreamContext(state), Border: border, Rows: rows, FocusCount: count,
		Primary: plainHint(dialogSubmitHint(state, submit)),
		Hints: func(focus int) []string {
			if create && focus <= 1 {
				return []string{"tab next", "type filter", "←/→ choose"}
			}
			if hints := cadenceHints(state, focus); hints != nil {
				return hints
			}
			return textarea(focus)
		},
	})
}

// renderNameDescriptionForm draws the repo and stream forms: Name, then Description.
func renderNameDescriptionForm(theme Theme, state controllerpkg.State, title, context string, border lipgloss.Color, submit string) string {
	inner := formInner(state.Width)
	column := len("Description")
	full, _ := formColumns(inner, column)
	return renderForm(theme, state, FormSpec{
		Title: title, Context: context, Border: border, FocusCount: 2,
		Primary: plainHint(dialogSubmitHint(state, submit)),
		Hints:   textareaHints(state.DescriptionIndex),
		Rows: []FormRow{
			fieldRow(0, Field(theme, "Name", column, bareInput(state.Inputs[0], full, "required"), state.FocusIdx == 0)),
			textareaRow(state.DescriptionIndex, Textarea(theme, "Description", column, bareTextarea(state), state.FocusIdx == state.DescriptionIndex, inner)),
		},
	})
}

// textareaHints is `tab next`, plus `enter newline` while the textarea is focused.
func textareaHints(descriptionIndex int) func(int) []string {
	return func(focus int) []string {
		if focus == descriptionIndex {
			return []string{"tab next", "enter newline"}
		}
		return []string{"tab next"}
	}
}
