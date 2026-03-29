package dialogs

func renderSessionDialog(theme Theme, state State) string {
	switch state.Kind {
	case "stash_list":
		rows := []string{theme.StylePaneTitle.Render("Stashes"), ""}
		if len(state.Stashes) == 0 {
			rows = append(rows, theme.StyleDim.Render("No stashes available"))
		} else {
			for i, stash := range state.Stashes {
				if i == state.StashCursor {
					rows = append(rows, theme.StyleCursor.Render("▶ "+stash.Label))
				} else {
					rows = append(rows, theme.StyleNormal.Render("  "+stash.Label))
				}
				rows = append(rows, theme.StyleDim.Render("  "+stash.Meta))
			}
		}
		rows = append(rows, "", theme.StyleDim.Render("[j/k] move   [enter] pop   [x] drop   [esc] cancel"))
		return modal(theme, state.Width, 60, theme.ColorYellow, rows)
	case "end_session", "stash_session":
		title := "End Session"
		hint := "[enter] confirm   [ctrl+e] details   [esc] cancel"
		labels := []string{"Commit message", "Worked on", "Outcome", "Next step", "Blockers", "Links"}
		if state.Kind == "stash_session" {
			title = "Stash Session"
			hint = "[enter] confirm   [esc] cancel"
			labels = []string{"Stash note"}
		}
		rows := []string{theme.StylePaneTitle.Render(title)}
		for i := range state.Inputs {
			rows = append(rows, "", theme.StyleDim.Render(labels[i]), state.Inputs[i].View())
		}
		rows = append(rows, "", theme.StyleDim.Render(hint))
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
	case "amend_session":
		rows := []string{
			theme.StylePaneTitle.Render("Amend Session"),
			"",
			theme.StyleDim.Render("Commit message"),
			state.Inputs[0].View(),
			"",
			theme.StyleDim.Render("[ctrl+s] save   [esc] cancel"),
		}
		return modal(theme, state.Width, 68, theme.ColorCyan, rows)
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
		if (state.IssueStatus == "done" || state.IssueStatus == "abandoned") && len(state.Inputs) > 0 {
			hint = "[enter] confirm   [esc] cancel"
			label := "Abandon reason"
			if state.IssueStatus == "done" {
				label = "Completion note"
			}
			rows = append(rows, "", theme.StyleDim.Render(label), state.Inputs[0].View())
		}
		rows = append(rows, "", theme.StyleDim.Render(hint))
		return modal(theme, state.Width, 68, border, rows)
	default:
		return ""
	}
}
