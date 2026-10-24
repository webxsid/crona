package dialogs

import controllerpkg "crona/tui/internal/tui/dialogs/controller"

func renderIssueDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "create_issue_meta", "create_issue_default":
		return renderIssueCreate(theme, state)
	case "edit_issue":
		return renderEditIssue(theme, state)
	case "issue_status":
		return renderIssueStatus(theme, state)
	case "issue_status_note":
		return renderStatusNote(theme, state)
	default:
		return ""
	}
}

func issueCreateHint(state controllerpkg.State) string {
	static := "[ctrl+s] save & close   [ctrl+a] create more   [ctrl+f] create & focus"
	if state.PromotingAdhoc {
		static = "[ctrl+s] convert issue"
	}
	if state.AdhocTimer {
		static = "[ctrl+s] create issue & finish session"
	}
	context := "[tab] next   [esc] cancel"
	if state.AdhocTimer {
		context = "[tab] next   [esc] back"
	}
	if state.Kind == "create_issue_default" {
		switch state.FocusIdx {
		case 0, 1:
			context = "[type] filter   [←/→] choose   [tab] move   " + context
		case 3:
			context = "[enter] newline   " + context
		case 5:
			context = "[ctrl+e] calendar   [g] today   " + context
		}
	} else {
		switch state.FocusIdx {
		case 1:
			context = "[enter] newline   " + context
		case 3:
			context = "[ctrl+e] calendar   [g] today   " + context
		}
	}
	return static + "\n" + context
}

// renderEditIssue draws Edit Issue: New Issue's fields without the pickers.
func renderEditIssue(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	column := len("Description")
	full, half := formColumns(inner, column)
	dueColumn := len("Due")
	if inner < 56 {
		dueColumn = column
	}
	return renderForm(theme, state, FormSpec{
		Title: "Edit Issue", Context: repoStreamContext(state), Border: theme.ColorYellow, FocusCount: 4,
		Primary: plainHint(dialogSubmitHint(state, "save")),
		Hints: func(focus int) []string {
			switch focus {
			case 1:
				return []string{"tab next", "enter newline"}
			case 3:
				return []string{"tab next", "ctrl+e calendar", "g today"}
			}
			return []string{"tab next"}
		},
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Title",
					column,
					bareInput(state.Inputs[0], full, "required"),
					state.FocusIdx == 0,
				),
			),
			textareaRow(
				1,
				Textarea(
					theme,
					"Description",
					column,
					bareTextarea(state),
					state.FocusIdx == 1,
					inner,
				),
			),
			pairRow(
				Field(
					theme,
					"Estimate",
					column,
					bareInput(state.Inputs[1], half, "45m, 1h30m"),
					state.FocusIdx == 2,
				),
				Field(
					theme,
					"Due",
					dueColumn,
					bareInput(state.Inputs[2], half, "YYYY-MM-DD"),
					state.FocusIdx == 3,
				),
				2,
				3,
				inner,
			),
		},
	})
}

// renderStatusNote draws the note asked for when an issue changes status; its prompt
// (StatusLabel) is the placeholder, since prompts are too long for the label column.
func renderStatusNote(theme Theme, state controllerpkg.State) string {
	title := map[string]string{"blocked": "Block Issue", "in_review": "Send To Review", "done": "Complete Issue", "abandoned": "Abandon Issue"}[state.IssueStatus]
	if title == "" {
		title = "Status Note"
	}
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Note"))
	return renderForm(theme, state, FormSpec{
		Title: title, Context: state.ViewName, Border: theme.ColorYellow, FocusCount: 1,
		Primary: plainHint(dialogSubmitHint(state, "set")),
		Hints:   func(int) []string { return []string{"tab next"} },
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Note",
					len("Note"),
					bareInput(state.Inputs[0], full, fallback(state.StatusLabel, "optional")),
					true,
				),
			),
		},
	})
}
