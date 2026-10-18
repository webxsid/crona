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
		contextRow := renderDefaultIssueContextColumns(theme, state, state.Width, 92)
		rows := []string{
			theme.StylePaneTitle.Render("New Habit"),
			"",
			theme.StyleDim.Render("Name"),
			state.Inputs[2].View(),
			"",
			theme.StyleDim.Render("Description (Optional)"),
			state.Description.View(),
			"",
			theme.StyleDim.Render("Schedule"),
			state.Inputs[3].View(),
			"",
			theme.StyleDim.Render("Target Duration (Optional)"),
			state.Inputs[4].View(),
			"",
			contextRow,
			"",
		}
		rows = appendDialogFooter(theme, state, rows, habitDialogHint(state, "create"))
		return modal(theme, state.Width, 92, theme.ColorCyan, rows)
	case "edit_habit":
		rows := []string{
			theme.StylePaneTitle.Render("Edit Habit"),
			"",
			theme.StyleDim.Render("Name"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Description (Optional)"),
			state.Description.View(),
			"",
			theme.StyleDim.Render("Schedule"),
			state.Inputs[1].View(),
			"",
			theme.StyleDim.Render("Target Duration (Optional)"),
			state.Inputs[2].View(),
			"",
		}
		rows = appendDialogFooter(theme, state, rows, habitDialogHint(state, "save"))
		return modal(theme, state.Width, 68, theme.ColorYellow, rows)
	default:
		return ""
	}
}

func habitDialogHint(state controllerpkg.State, submitLabel string) string {
	if state.Kind == "create_habit" {
		switch state.FocusIdx {
		case 0, 1:
			return "[type] filter   [←/→] choose   [tab] move   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	}
	if state.FocusIdx == state.DescriptionIndex {
		return "[enter] newline   [tab] next   " + dialogSubmitHint(
			state,
			submitLabel,
		) + "   [esc] cancel"
	}
	return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
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
