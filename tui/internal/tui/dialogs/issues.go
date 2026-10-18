package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	issuecore "crona/tui/internal/tui/views/issuecore"
	"github.com/charmbracelet/lipgloss"
)

func renderIssueDialog(theme Theme, state controllerpkg.State) string {
	const issueDialogWidth = 92
	switch state.Kind {
	case "create_issue_meta", "create_issue_default":
		return renderIssueCreate(theme, state)
	case "edit_issue":
		return renderEditIssue(theme, state)
	case "issue_status":
		contextRow := renderIssueContextColumns(
			theme,
			state.Width,
			64,
			state.RepoName,
			state.StreamName,
		)
		statusSummary := theme.StyleDim.Render("Current status") + "\n" +
			theme.StyleHeader.Render(plainIssueStatus(state.IssueStatus))
		workSummary := theme.StyleDim.Render("Effort") + "\n" +
			theme.StyleHeader.Render(
				issuecore.IssueWorkedEstimateSummary(
					state.IssueWorkedSeconds,
					state.IssueEstimateMins,
				),
			)
		metaRow := renderInputColumns(state.Width, 64, statusSummary, workSummary)
		rows := []string{
			theme.StylePaneTitle.Render("Set Issue Status"),
			"",
			theme.StyleDim.Render("Issue"),
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
			"",
			contextRow,
			"",
			metaRow,
			"",
			theme.StyleDim.Render("Change status to"),
		}
		if len(state.StatusItems) == 0 {
			rows = appendDialogFooter(
				theme,
				state,
				append(rows, theme.StyleDim.Render("No valid status transitions")),
				"[esc] close",
			)
		} else {
			for i, status := range state.StatusItems {
				label := plainIssueStatus(string(status))
				if i == state.StatusCursor {
					rows = append(rows, theme.StyleSelected.Render("> "+label))
				} else {
					rows = append(rows, "  "+label)
				}
			}
			rows = appendDialogFooter(theme, state, rows, "[↑/↓] move   [enter] set   [esc] cancel")
		}
		return modal(theme, state.Width, 64, theme.ColorYellow, rows)
	case "issue_status_note":
		return renderStatusNote(theme, state)
	default:
		return ""
	}
}

func issueDialogHint(state controllerpkg.State, submitLabel string) string {
	if state.Kind == "create_issue_default" || state.Kind == "create_issue_meta" {
		return issueCreateHint(state)
	}
	switch state.Kind {
	case "edit_issue":
		switch state.FocusIdx {
		case 1:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		case 3:
			return "[ctrl+e] calendar   [g] today   [tab] next   " + dialogSubmitHint(
				state,
				submitLabel,
			) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	default:
		return dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
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

func renderIssueContextColumns(
	theme Theme,
	width, maxWidth int,
	repoName, streamName string,
) string {
	contentWidth := min(width-8, maxWidth) - 8
	contentWidth = max(28, contentWidth)
	colWidth := (contentWidth - 2) / 2
	colWidth = max(colWidth, 12)
	repoCol := lipgloss.NewStyle().Width(colWidth).Render(
		theme.StyleDim.Render("Repo") + "\n" + theme.StyleHeader.Render(fallback(repoName, "-")),
	)
	streamCol := lipgloss.NewStyle().Width(colWidth).Render(
		theme.StyleDim.Render("Stream") +
			"\n" + theme.StyleHeader.Render(fallback(streamName, "-")),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, repoCol, "  ", streamCol)
}

func renderDefaultIssueContextColumns(
	theme Theme,
	state controllerpkg.State,
	width, maxWidth int,
) string {
	contentWidth := min(width-8, maxWidth) - 8
	contentWidth = max(28, contentWidth)
	colWidth := (contentWidth - 2) / 2
	colWidth = max(colWidth, 12)
	repoCol := lipgloss.NewStyle().Width(colWidth).Render(strings.Join([]string{
		theme.StyleDim.Render("Repo"),
		state.Inputs[0].View(),
		"",
		renderSelector(theme, state, state.RepoSelectorLabel, state.FocusIdx == 0),
	}, "\n"))
	streamCol := lipgloss.NewStyle().Width(colWidth).Render(strings.Join([]string{
		theme.StyleDim.Render("Stream"),
		state.Inputs[1].View(),
		"",
		renderSelector(theme, state, state.StreamSelectorLabel, state.FocusIdx == 1),
	}, "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, repoCol, "  ", streamCol)
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
			fieldRow(0, Field(theme, "Title", column, bareInput(state.Inputs[0], full, "required"), state.FocusIdx == 0)),
			textareaRow(1, Textarea(theme, "Description", column, bareTextarea(state), state.FocusIdx == 1, inner)),
			pairRow(
				Field(theme, "Estimate", column, bareInput(state.Inputs[1], half, "45m, 1h30m"), state.FocusIdx == 2),
				Field(theme, "Due", dueColumn, bareInput(state.Inputs[2], half, "YYYY-MM-DD"), state.FocusIdx == 3),
				2, 3, inner,
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
			fieldRow(0, Field(theme, "Note", len("Note"), bareInput(state.Inputs[0], full, fallback(state.StatusLabel, "optional")), true)),
		},
	})
}
