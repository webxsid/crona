package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/bubbles/textinput"
)

// issueFormInner is the inner width New Issue lays its fields out to.
const issueFormInner = 72

// renderIssueCreate draws New Issue (and its ad hoc variants) with the dialog chrome.
// create_issue_default focuses Repo, Stream, Title, Description, Estimate, Due, Summary;
// create_issue_meta focuses Title, Description, Estimate, Due.
func renderIssueCreate(theme Theme, state controllerpkg.State) string {
	inner := FrameInner(state.Width, []string{strings.Repeat(" ", issueFormInner)})
	column := len("Description")
	full := inner - column - 3
	half := inner/2 - column - 6
	dueColumn := len("Due")
	if inner < 56 {
		half, dueColumn = full, column
	}
	title, estimate, due := 0, 1, 2
	focusOf := map[string]int{"title": 0, "description": 1, "estimate": 2, "due": 3}
	if state.Kind == "create_issue_default" {
		title, estimate, due = 2, 3, 4
		focusOf = map[string]int{"repo": 0, "stream": 1, "title": 2, "description": 3, "estimate": 4, "due": 5, "summary": 6}
	}
	focused := func(name string) bool { idx, ok := focusOf[name]; return ok && state.FocusIdx == idx }

	body, focusLine := []string{}, -1
	add := func(name string, lines ...string) {
		if focused(name) {
			focusLine = len(body)
		}
		body = append(body, lines...)
	}
	frameTitle := "New Issue"
	if state.PromotingAdhoc {
		frameTitle = "Convert Ad hoc Issue"
	}
	if state.AdhocTimer {
		frameTitle = "Create Issue & Finish Session"
	}
	if (state.AdhocTimer || state.PromotingAdhoc) && state.ViewName != "" {
		body = append(body, theme.StyleDim.Render(state.ViewName), "")
	}
	add("title", Field(theme, "Title", column, bareInput(state.Inputs[title], full, "required"), focused("title")))
	add("description", Textarea(theme, "Description", column, bareTextarea(state), focused("description"), inner)...)
	pair := FieldPair(
		Field(theme, "Estimate", column, bareInput(state.Inputs[estimate], half, "45m, 1h30m"), focused("estimate")),
		Field(theme, "Due", dueColumn, bareInput(state.Inputs[due], half, "YYYY-MM-DD"), focused("due")),
		inner,
	)
	if focused("due") && len(pair) > 1 {
		body = append(body, pair[0])
		add("due", pair[1])
	} else {
		add("estimate", pair...)
		if focused("due") {
			focusLine = len(body) - 1
		}
	}
	if state.Kind == "create_issue_default" {
		body = append(body, "")
		add("repo", Picker(theme, PickerSpec{
			Label: "Repo", Column: column, Focused: focused("repo"), Input: bareInput(state.Inputs[0], full, "type to filter"),
			Current: state.RepoSelectorLabel, Options: state.RepoOptionLabels, Selected: state.RepoIndex,
			Placeholder: state.RepoSelectorLabel,
		})...)
		add("stream", Picker(theme, PickerSpec{
			Label: "Stream", Column: column, Focused: focused("stream"), Input: bareInput(state.Inputs[1], full, "type to filter"),
			Current: state.StreamSelectorLabel, Options: state.StreamOptionLabels, Selected: state.StreamIndex,
			Placeholder: state.StreamSelectorLabel,
		})...)
	}
	if state.Kind == "create_issue_meta" {
		body = append(body, Field(theme, "Repo", column, fallback(state.RepoName, "-"), false), Field(theme, "Stream", column, fallback(state.StreamName, "-"), false))
	}
	if state.AdhocTimer {
		add("summary", Field(theme, "Summary", column, bareInput(state.Inputs[5], full, "required"), focused("summary")))
	}
	primary, hints := splitHints(issueCreateHint(state))
	footer := Footer(theme, primary, hints, inner)
	tallest := 0
	for focus := 0; focus < len(focusOf); focus++ {
		probe := state
		probe.FocusIdx = focus
		p, h := splitHints(issueCreateHint(probe))
		tallest = max(tallest, len(Footer(theme, p, h, inner)))
	}
	for len(footer) < tallest {
		footer = append(footer, "")
	}
	context := ""
	if state.RepoName != "" && state.StreamName != "" {
		context = state.RepoName + " › " + state.StreamName
	}
	return Frame(theme, FrameSpec{
		Title: frameTitle, Context: context, Body: body, Footer: withError(theme, state, footer),
		ScreenWidth: state.Width, ScreenHeight: state.Height, Focus: focusLine, Inner: inner,
	})
}

// bareInput renders a text input without its prompt glyph at width cells, keeping its
// cursor; a non-empty placeholder replaces one that repeated the field's label.
func bareInput(input textinput.Model, width int, placeholder string) string {
	input.Prompt = ""
	input.Width = max(1, width)
	if placeholder != "" {
		input.Placeholder = placeholder
	}
	return input.View()
}

func bareTextarea(state controllerpkg.State) string {
	description := state.Description
	description.Prompt = ""
	description.ShowLineNumbers = false
	description.Placeholder = "optional"
	description.SetHeight(3)
	return description.View()
}

// splitHints turns a bracketed hint string into `key action` hints: the first is primary;
// `esc` hints are dropped because Footer always ends with esc cancel.
func splitHints(hint string) (string, []string) {
	hints := []string{}
	for _, line := range strings.Split(hint, "\n") {
		for _, part := range strings.Split(line, "   ") {
			part = strings.TrimSpace(part)
			if part == "" || strings.HasPrefix(part, "[esc]") {
				continue
			}
			hints = append(hints, strings.Replace(strings.Replace(part, "[", "", 1), "]", "", 1))
		}
	}
	if len(hints) == 0 {
		return "", nil
	}
	return hints[0], hints[1:]
}

// withError puts the dialog's validation error above its footer, where cutting the body
// for a short screen can never hide it.
func withError(theme Theme, state controllerpkg.State, footer []string) []string {
	if strings.TrimSpace(state.ErrorMessage) == "" {
		return footer
	}
	return append([]string{theme.StyleError.Render(state.ErrorMessage), ""}, footer...)
}

// renderCheckoutContext draws the context switcher: Repo and Stream pickers.
func renderCheckoutContext(theme Theme, state controllerpkg.State) string {
	inner := FrameInner(state.Width, []string{strings.Repeat(" ", 56)})
	column := len("Stream")
	full := inner - column - 3
	body := Picker(theme, PickerSpec{
		Label: "Repo", Column: column, Focused: state.FocusIdx == 0, Input: bareInput(state.Inputs[0], full, "type to filter"),
		Current: state.RepoSelectorLabel, Options: state.RepoOptionLabels, Selected: state.RepoIndex,
		Placeholder: state.RepoSelectorLabel,
	})
	focus := 0
	if state.FocusIdx == 1 {
		focus = len(body)
	}
	body = append(body, Picker(theme, PickerSpec{
		Label: "Stream", Column: column, Focused: state.FocusIdx == 1, Input: bareInput(state.Inputs[1], full, "type to filter"),
		Current: state.StreamSelectorLabel, Options: state.StreamOptionLabels, Selected: state.StreamIndex,
		Placeholder: state.StreamSelectorLabel,
	})...)
	footer := Footer(theme, "enter set context", []string{"c clear", "↑/↓/tab move", "type filter", "←/→ choose"}, inner)
	return Frame(theme, FrameSpec{
		Title: "Checkout Context", Body: body, Footer: withError(theme, state, footer),
		ScreenWidth: state.Width, ScreenHeight: state.Height, Focus: focus, Inner: inner,
	})
}
