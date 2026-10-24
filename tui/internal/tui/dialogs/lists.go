package dialogs

import (
	"fmt"
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
	issuecore "crona/tui/internal/tui/views/issuecore"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	border         lipgloss.Color
	esc            string
	// wide pins the frame to the form width, so a wizard keeps one width across steps.
	wide    bool
	body    []string
	focus   int
	primary string
	hints   []string
}

func renderListDialog(theme Theme, state controllerpkg.State, d listDialog) string {
	inner := FrameInner(state.Width, d.body)
	if d.wide {
		inner = formInner(state.Width)
	}
	esc := d.esc
	if esc == "" {
		esc = "esc cancel"
	}
	return Frame(theme, FrameSpec{
		Title: d.title, Context: d.context, Danger: d.danger, Border: d.border, Body: d.body,
		Footer:      withError(theme, state, footerWithEsc(theme, d.primary, d.hints, esc, inner)),
		ScreenWidth: state.Width, ScreenHeight: state.Height, Focus: d.focus, Inner: inner,
	})
}

func renderConfirmDelete(theme Theme, state controllerpkg.State) string {
	return renderListDialog(theme, state, listDialog{
		title: "Delete " + titleWords(fallback(state.DeleteKind, "item")), danger: true, focus: -1,
		body: []string{
			theme.StyleError.Render("Delete " + fallback(state.DeleteLabel, "this item") + "?"),
		},
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
		Field(
			theme,
			"Issue",
			column,
			theme.StyleHeader.Render(fallback(state.ViewName, "-")),
			false,
		),
		Field(theme, "Status", column, plainIssueStatus(state.IssueStatus), false),
		Field(
			theme,
			"Effort",
			column,
			issuecore.IssueWorkedEstimateSummary(state.IssueWorkedSeconds, state.IssueEstimateMins),
			false,
		),
	}
	body = append(body, "", theme.StyleDim.Render("Change status to"))
	if len(state.StatusItems) == 0 {
		body = append(body, theme.StyleDim.Render("No valid status transitions"))
		return renderListDialog(
			theme,
			state,
			listDialog{
				title:   "Set Issue Status",
				context: repoStreamContext(state),
				body:    body,
				focus:   -1,
			},
		)
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
	if state.PenaltyHint != "" {
		body = append(body, "", theme.StyleDim.Render(state.PenaltyHint))
	}
	hints := []string{"h/j/k/l move", ",/. month", "c clear"}
	if state.Kind == "summary_calendar" {
		hints = []string{"←/→/↑/↓ move", ",/. month", "g today"}
	}
	return renderListDialog(theme, state, listDialog{
		title: fallback(
			state.DateTitle,
			"Pick Date",
		), context: state.DateMonth, body: body, focus: -1,
		primary: "enter choose", hints: hints,
	})
}

// renderAdhocIssuePicker draws Attach to Issue: section tabs, then two-line issue rows. The
// list is padded to the fullest tab, so switching tabs never changes the dialog's height.
func renderAdhocIssuePicker(theme Theme, state controllerpkg.State) string {
	body := []string{}
	if state.ViewName != "" {
		body = append(body, theme.StyleDim.Render(state.ViewName), "")
	}
	tabs := []string{}
	for i, label := range []string{"Due today", "Overdue", "Pinned", "All"} {
		label = fmt.Sprintf("%s %d", label, state.AdhocIssueTabCounts[i])
		if i == state.AdhocIssueTab {
			tabs = append(tabs, theme.StyleSelected.Render(" "+label+" "))
		} else {
			tabs = append(tabs, theme.StyleDim.Render(" "+label+" "))
		}
	}
	body = append(body, strings.Join(tabs, " "), "")
	width := FrameInner(state.Width, []string{strings.Repeat(" ", frameMaxInner)}) - 2
	start, focus := len(body), -1
	for i, item := range state.ChoiceItems {
		if i > 0 {
			body = append(body, "")
		}
		if i == state.ChoiceCursor {
			focus = len(body)
		}
		body = append(
			body,
			selectList(
				theme,
				[]string{ansi.Truncate(strings.ReplaceAll(item, "\n", " "), width, "…")},
				boolIndex(i == state.ChoiceCursor),
			)...)
		if i < len(state.ChoiceDetails) {
			body = append(
				body,
				theme.StyleDim.Render("  "+ansi.Truncate(state.ChoiceDetails[i], width, "…")),
			)
		}
	}
	if len(state.ChoiceItems) == 0 {
		message := "No issues in this tab. Switch tabs to see other eligible issues."
		if state.AdhocIssueTabCounts[3] == 0 {
			message = "No eligible issues are available."
		}
		body = append(body, theme.StyleDim.Render(message))
	}
	fullest := 0
	for _, count := range state.AdhocIssueTabCounts {
		fullest = max(fullest, count)
	}
	for len(body) < start+max(1, fullest*3-1) {
		body = append(body, "")
	}
	context := ""
	if len(state.ChoiceItems) > 0 {
		context = fmt.Sprintf("%d of %d", state.ChoiceCursor+1, len(state.ChoiceItems))
	}
	return renderListDialog(theme, state, listDialog{
		title: fallback(
			state.ViewTitle,
			"Attach to Issue",
		), context: context, body: body, focus: focus,
		primary: "enter continue", hints: []string{"←/→ tab", "↑/↓ move"},
	})
}

// boolIndex is 0 when on (selectList's cursor on the single row) and -1 otherwise.
func boolIndex(on bool) int {
	if on {
		return 0
	}
	return -1
}
