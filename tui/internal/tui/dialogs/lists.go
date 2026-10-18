package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
	issuecore "crona/tui/internal/tui/views/issuecore"
)

// selectList draws a list with ▶ on the cursor row.
func selectList(theme Theme, items []string, cursor int) []string {
	lines := make([]string, 0, len(items))
	for i, item := range items {
		if i == cursor {
			lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+item))
			continue
		}
		lines = append(lines, theme.StyleNormal.Render("  "+item))
	}
	return lines
}

// listDialog is a dialog that is not a form: a body (with an optional focused line kept in
// view) and a footer of hints.
type listDialog struct {
	title, context string
	danger         bool
	body           []string
	focus          int
	primary        string
	hints          []string
}

func renderListDialog(theme Theme, state controllerpkg.State, d listDialog) string {
	inner := FrameInner(state.Width, d.body)
	return Frame(theme, FrameSpec{
		Title: d.title, Context: d.context, Danger: d.danger, Body: d.body,
		Footer:      withError(theme, state, Footer(theme, d.primary, d.hints, inner)),
		ScreenWidth: state.Width, ScreenHeight: state.Height, Focus: d.focus, Inner: inner,
	})
}

func renderConfirmDelete(theme Theme, state controllerpkg.State) string {
	return renderListDialog(theme, state, listDialog{
		title: "Delete " + titleWords(fallback(state.DeleteKind, "item")), danger: true, focus: -1,
		body:    []string{theme.StyleError.Render("Delete " + fallback(state.DeleteLabel, "this item") + "?")},
		primary: "enter delete",
	})
}

func renderConfirmWipe(theme Theme, state controllerpkg.State) string {
	return renderListDialog(theme, state, listDialog{
		title: "Wipe All Data", danger: true, focus: -1,
		body: []string{
			theme.StyleError.Render("This deletes all Crona runtime data."),
			theme.StyleDim.Render("Issues, sessions, habits, reports, and settings will be reset."),
			theme.StyleDim.Render("Installed binaries will remain untouched."),
		},
		primary: "enter wipe data",
	})
}

func renderIssueStatus(theme Theme, state controllerpkg.State) string {
	column := len("Status")
	body := []string{
		Field(theme, "Issue", column, theme.StyleHeader.Render(fallback(state.ViewName, "-")), false),
		Field(theme, "Status", column, plainIssueStatus(state.IssueStatus), false),
		Field(theme, "Effort", column, issuecore.IssueWorkedEstimateSummary(state.IssueWorkedSeconds, state.IssueEstimateMins), false),
	}
	body = append(body, "", theme.StyleDim.Render("Change status to"))
	if len(state.StatusItems) == 0 {
		body = append(body, theme.StyleDim.Render("No valid status transitions"))
		return renderListDialog(theme, state, listDialog{title: "Set Issue Status", context: repoStreamContext(state), body: body, focus: -1})
	}
	labels := make([]string, 0, len(state.StatusItems))
	for _, status := range state.StatusItems {
		labels = append(labels, plainIssueStatus(string(status)))
	}
	focus := len(body) + state.StatusCursor
	body = append(body, selectList(theme, labels, state.StatusCursor)...)
	return renderListDialog(theme, state, listDialog{
		title: "Set Issue Status", context: repoStreamContext(state), body: body, focus: focus,
		primary: "enter set", hints: []string{"↑/↓ move"},
	})
}

// renderMenuList draws Jump To View and Beta Support: a keyed list and the selected
// item's help.
func renderMenuList(theme Theme, state controllerpkg.State) string {
	body := []string{}
	if state.ViewName != "" {
		body = append(body, theme.StyleDim.Render(state.ViewName), "")
	}
	focus := len(body) + state.ChoiceCursor
	body = append(body, selectList(theme, state.ChoiceItems, state.ChoiceCursor)...)
	if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) &&
		strings.TrimSpace(state.ChoiceDetails[state.ChoiceCursor]) != "" {
		body = append(body, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
	}
	verb := "jump"
	if state.Kind == "beta_support" {
		verb = "run"
	}
	return renderListDialog(theme, state, listDialog{
		title: state.ViewTitle, body: body, focus: focus,
		primary: "enter " + verb, hints: []string{"↑/↓ move", "key " + verb},
	})
}

// titleWords turns a snake_case kind into Title Words.
func titleWords(kind string) string {
	words := strings.Fields(strings.ReplaceAll(kind, "_", " "))
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

// renderDateGrid draws the due-date picker and the summary calendar: the selected day, then
// the month grid, with the month in the top border.
func renderDateGrid(theme Theme, state controllerpkg.State) string {
	body := []string{theme.StyleHeader.Render(state.DateHeader), ""}
	body = append(body, strings.Split(state.DateGrid, "\n")...)
	hints := []string{"h/j/k/l move", ",/. month", "c clear"}
	if state.Kind == "summary_calendar" {
		hints = []string{"←/→/↑/↓ move", ",/. month", "g today"}
	}
	return renderListDialog(theme, state, listDialog{
		title: fallback(state.DateTitle, "Pick Date"), context: state.DateMonth, body: body, focus: -1,
		primary: "enter choose", hints: hints,
	})
}
