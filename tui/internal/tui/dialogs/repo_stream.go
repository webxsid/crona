package dialogs

import controllerpkg "crona/tui/internal/tui/dialogs/controller"

func renderRepoStreamDialog(theme Theme, state controllerpkg.State) string {
	switch state.Kind {
	case "create_repo":
		rows := []string{theme.StylePaneTitle.Render("New Repo"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View()}
		rows = appendDialogFooter(theme, state, rows, nameDescriptionHint(state, "create"))
		return modal(theme, state.Width, 56, theme.ColorCyan, rows)
	case "edit_repo":
		rows := []string{theme.StylePaneTitle.Render("Edit Repo"), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View()}
		rows = appendDialogFooter(theme, state, rows, nameDescriptionHint(state, "save"))
		return modal(theme, state.Width, 56, theme.ColorYellow, rows)
	case "create_stream":
		rows := []string{theme.StylePaneTitle.Render("New Stream"), "", theme.StyleDim.Render("Repo"), theme.StyleHeader.Render(state.RepoName), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View()}
		rows = appendDialogFooter(theme, state, rows, nameDescriptionHint(state, "create"))
		return modal(theme, state.Width, 56, theme.ColorCyan, rows)
	case "edit_stream":
		rows := []string{theme.StylePaneTitle.Render("Edit Stream"), "", theme.StyleDim.Render("Repo"), theme.StyleHeader.Render(state.RepoName), "", theme.StyleDim.Render("Name"), state.Inputs[0].View(), "", theme.StyleDim.Render("Description (Optional)"), state.Description.View()}
		rows = appendDialogFooter(theme, state, rows, nameDescriptionHint(state, "save"))
		return modal(theme, state.Width, 56, theme.ColorYellow, rows)
	case "checkout_context":
		rows := []string{
			theme.StylePaneTitle.Render("Checkout Context"),
			"",
			theme.StyleDim.Render("Repo"),
			state.Inputs[0].View(),
			"",
			renderSelector(theme, state, state.RepoSelectorLabel, true),
			"",
			theme.StyleDim.Render("Stream"),
			state.Inputs[1].View(),
			"",
			renderSelector(theme, state, state.StreamSelectorLabel, false),
			"",
		}
		rows = appendDialogFooter(theme, state, rows, checkoutHint())
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

func nameDescriptionHint(state controllerpkg.State, submitLabel string) string {
	if state.FocusIdx == state.DescriptionIndex {
		return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
	}
	return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
}

func checkoutHint() string {
	return "[type] filter   [left/right] choose   [up/down/tab] move   [enter] checkout/create   [c] clear   [esc] cancel"
}

func habitDialogHint(state controllerpkg.State, submitLabel string) string {
	if state.Kind == "create_habit" {
		switch state.FocusIdx {
		case 0, 1:
			return "[type] filter   [left/right] choose   [up/down/tab] move   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		case 3:
			return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		default:
			return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
		}
	}
	if state.FocusIdx == state.DescriptionIndex {
		return "[enter] newline   [tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
	}
	return "[tab] next   " + dialogSubmitHint(state, submitLabel) + "   [esc] cancel"
}
