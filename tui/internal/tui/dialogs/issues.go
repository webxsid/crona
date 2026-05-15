package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	"github.com/charmbracelet/lipgloss"
)

func renderIssueDialog(theme Theme, state controllerpkg.State) string {
	const issueDialogWidth = 92
	switch state.Kind {
	case "create_issue_meta":
		contextRow := renderIssueContextColumns(theme, state.Width, issueDialogWidth, state.RepoName, state.StreamName)
		schedulingRow := renderInputColumns(state.Width, issueDialogWidth,
			theme.StyleDim.Render("Estimate (Optional)")+"\n"+state.Inputs[1].View(),
			theme.StyleDim.Render("Due (Optional)")+"\n"+state.Inputs[2].View(),
		)
		rows := []string{
			theme.StylePaneTitle.Render("New Issue"),
			"",
			theme.StyleDim.Render("Title"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Description (Optional)"),
			state.Description.View(),
			"",
			schedulingRow,
			"",
			contextRow,
			"",
		}
		rows = appendDialogFooter(theme, state, rows, issueDialogHint(state, "create"))
		return modal(theme, state.Width, issueDialogWidth, theme.ColorCyan, rows)
	case "create_issue_default":
		contextRow := renderDefaultIssueContextColumns(theme, state, state.Width, issueDialogWidth)
		schedulingRow := renderInputColumns(state.Width, issueDialogWidth,
			theme.StyleDim.Render("Estimate (Optional)")+"\n"+state.Inputs[3].View(),
			theme.StyleDim.Render("Due (Optional)")+"\n"+state.Inputs[4].View(),
		)
		rows := []string{
			theme.StylePaneTitle.Render("New Issue"),
			"",
			theme.StyleDim.Render("Title"),
			state.Inputs[2].View(),
			"",
			theme.StyleDim.Render("Description (Optional)"),
			state.Description.View(),
			"",
			schedulingRow,
			"",
			contextRow,
			"",
		}
		rows = appendDialogFooter(theme, state, rows, issueDialogHint(state, "create"))
		return modal(theme, state.Width, issueDialogWidth, theme.ColorCyan, rows)
	case "edit_issue":
		schedulingRow := renderInputColumns(state.Width, issueDialogWidth,
			theme.StyleDim.Render("Estimate (Optional)")+"\n"+state.Inputs[1].View(),
			theme.StyleDim.Render("Due (Optional)")+"\n"+state.Inputs[2].View(),
		)
		rows := []string{
			theme.StylePaneTitle.Render("Edit Issue"),
			"",
			theme.StyleDim.Render("Title"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("Description (Optional)"),
			state.Description.View(),
			"",
			schedulingRow,
			"",
		}
		rows = appendDialogFooter(theme, state, rows, issueDialogHint(state, "save"))
		return modal(theme, state.Width, issueDialogWidth, theme.ColorYellow, rows)
	case "issue_status":
		rows := []string{theme.StylePaneTitle.Render("Set Issue Status"), ""}
		if len(state.StatusItems) == 0 {
			rows = appendDialogFooter(theme, state, append(rows, theme.StyleDim.Render("No valid status transitions")), "[esc] close")
		} else {
			for i, status := range state.StatusItems {
				label := plainIssueStatus(string(status))
				if i == state.StatusCursor {
					rows = append(rows, theme.StyleSelected.Render("> "+label))
				} else {
					rows = append(rows, "  "+label)
				}
			}
			rows = appendDialogFooter(theme, state, rows, "[j/k] move   [enter] set   [esc] cancel")
		}
		return modal(theme, state.Width, 48, theme.ColorYellow, rows)
	case "issue_status_note":
		title := map[string]string{"blocked": "Block Issue", "in_review": "Send To Review", "done": "Complete Issue", "abandoned": "Abandon Issue"}[state.IssueStatus]
		if title == "" {
			title = "Status Note"
		}
		rows := []string{theme.StylePaneTitle.Render(title), "", theme.StyleDim.Render(state.StatusLabel), state.Inputs[0].View()}
		rows = appendDialogFooter(theme, state, rows, "[tab] next   "+dialogSubmitHint(state, "set")+"   [esc] cancel")
		return modal(theme, state.Width, 60, theme.ColorYellow, rows)
	default:
		return ""
	}
}

func issueDialogHint(state controllerpkg.State, submitLabel string) string {
	switch state.Kind {
	case "create_issue_default":
		switch state.FocusIdx {
		case 0, 1:
			return "[type] filter   [left/right] choose   [up/down/tab] move   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 5:
			return "[f2] calendar   [g] today   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	case "create_issue_meta", "edit_issue":
		switch state.FocusIdx {
		case 1:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 3:
			return "[f2] calendar   [g] today   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	default:
		return dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
	}
}

func renderIssueContextColumns(theme Theme, width, maxWidth int, repoName, streamName string) string {
	contentWidth := min(width-8, maxWidth) - 8
	if contentWidth < 28 {
		contentWidth = 28
	}
	colWidth := (contentWidth - 2) / 2
	if colWidth < 12 {
		colWidth = 12
	}
	repoCol := lipgloss.NewStyle().Width(colWidth).Render(
		theme.StyleDim.Render("Repo") + "\n" + theme.StyleHeader.Render(fallback(repoName, "-")),
	)
	streamCol := lipgloss.NewStyle().Width(colWidth).Render(
		theme.StyleDim.Render("Stream") + "\n" + theme.StyleHeader.Render(fallback(streamName, "-")),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, repoCol, "  ", streamCol)
}

func renderDefaultIssueContextColumns(theme Theme, state controllerpkg.State, width, maxWidth int) string {
	contentWidth := min(width-8, maxWidth) - 8
	if contentWidth < 28 {
		contentWidth = 28
	}
	colWidth := (contentWidth - 2) / 2
	if colWidth < 12 {
		colWidth = 12
	}
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
