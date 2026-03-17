package dialogs

func renderRepoStreamDialog(theme Theme, state State) string {
	switch state.Kind {
	case "create_repo":
		rows := []string{theme.StylePaneTitle.Render("New Repo"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View(), "", theme.StyleDim.Render("[enter] newline in description   [ctrl+s] create   [tab] next   [esc] cancel")}
		return modal(theme, state.Width, 56, theme.ColorCyan, rows)
	case "edit_repo":
		rows := []string{theme.StylePaneTitle.Render("Edit Repo"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View(), "", theme.StyleDim.Render("[enter] newline in description   [ctrl+s] save   [tab] next   [esc] cancel")}
		return modal(theme, state.Width, 56, theme.ColorYellow, rows)
	case "create_stream":
		rows := []string{theme.StylePaneTitle.Render("New Stream"), "", theme.StyleDim.Render("Repo"), theme.StyleHeader.Render(state.RepoName), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View(), "", theme.StyleDim.Render("[enter] newline in description   [ctrl+s] create   [tab] next   [esc] cancel")}
		return modal(theme, state.Width, 56, theme.ColorCyan, rows)
	case "edit_stream":
		rows := []string{theme.StylePaneTitle.Render("Edit Stream"), "", theme.StyleDim.Render("Repo"), theme.StyleHeader.Render(state.RepoName), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View(), "", theme.StyleDim.Render("[enter] newline in description   [ctrl+s] save   [tab] next   [esc] cancel")}
		return modal(theme, state.Width, 56, theme.ColorYellow, rows)
	case "checkout_context":
		rows := []string{
			theme.StylePaneTitle.Render("Checkout Context"),
			"",
			theme.StyleDim.Render("Repo"),
			state.Inputs[0].View(),
			"",
			renderSelector(theme, state.RepoSelectorLabel, true),
			"",
			theme.StyleDim.Render("Stream"),
			state.Inputs[1].View(),
			"",
			renderSelector(theme, state.StreamSelectorLabel, false),
			"",
			theme.StyleDim.Render("[left/right] choose   [up/down/tab] move   [enter] checkout/create   [c] clear   [esc] cancel"),
		}
		return modal(theme, state.Width, 72, theme.ColorCyan, rows)
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
			theme.StyleDim.Render("Target Minutes (Optional)"),
			state.Inputs[4].View(),
			"",
			contextRow,
			"",
			theme.StyleDim.Render(habitDialogHint(state, "create")),
		}
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
			theme.StyleDim.Render("Target Minutes (Optional)"),
			state.Inputs[2].View(),
			"",
			theme.StyleDim.Render("[enter] newline in description   [ctrl+s] save   [tab] next   [esc] cancel"),
		}
		return modal(theme, state.Width, 68, theme.ColorYellow, rows)
	default:
		return ""
	}
}

func habitDialogHint(state State, submitLabel string) string {
	if state.Kind == "create_habit" {
		switch state.FocusIdx {
		case 0, 1:
			return "[type] filter   [left/right] choose   [up/down/tab] move   [ctrl+s] " + submitLabel + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   [ctrl+s] " + submitLabel + "   [esc] cancel"
		default:
			return "[tab] next   [ctrl+s] " + submitLabel + "   [esc] cancel"
		}
	}
	return "[tab] next   [ctrl+s] " + submitLabel + "   [esc] cancel"
}
