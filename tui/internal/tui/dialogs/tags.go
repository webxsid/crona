package dialogs

import (
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

// renderTagPicker draws the tag picker: the filter-or-create input, then the checklist.
func renderTagPicker(theme Theme, state controllerpkg.State) string {
	body := []string{}
	if state.ViewName != "" {
		body = append(body, theme.StyleDim.Render(state.ViewName), "")
	}
	if len(state.Inputs) > 0 {
		body = append(body, state.Inputs[0].View(), "")
	}
	rows := controllerpkg.TagPickerRows(state)
	focus := len(body) + state.TagCursor
	if len(rows) == 0 {
		body = append(body, theme.StyleDim.Render("  No tags yet. Type a name to create one."))
		focus = -1
	}
	for i, row := range rows {
		mark := "[ ]"
		if row.Checked {
			mark = "[x]"
		}
		text := mark + " +" + row.Name
		if row.Create {
			text = "[+] create +" + row.Name
		}
		if i == state.TagCursor {
			body = append(body, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+text))
			continue
		}
		body = append(body, theme.StyleNormal.Render("  "+text))
	}
	return renderListDialog(theme, state, listDialog{
		title: "Tags", body: body, focus: focus,
		primary: "enter save", hints: []string{"↑/↓ move", "space toggle", "type to filter or create"},
	})
}

// renderPaneFilter draws the pane filter: free text, +tag to keep and -tag to drop.
func renderPaneFilter(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Filter"))
	return renderForm(theme, state, FormSpec{
		Title: "Filter " + titleWords(state.ViewName), FocusCount: 1, Primary: "enter apply",
		Rows: []FormRow{
			fieldRow(0, Field(theme, "Filter", len("Filter"), bareInput(state.Inputs[0], full, "text  +tag  -tag"), true)),
			staticRow("", theme.StyleDim.Render("Words match the text, +tag keeps a tag (any of several), -tag drops it. Empty clears.")),
		},
	})
}
