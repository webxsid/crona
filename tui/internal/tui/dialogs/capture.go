package dialogs

import (
	"fmt"
	"strings"

	shareddto "crona/shared/dto"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"

	"github.com/charmbracelet/x/ansi"
)

// renderCaptureIssues draws Add issues: one textarea of `- ` lines and the default target.
func renderCaptureIssues(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	description := state.Description
	description.SetWidth(max(20, inner))
	description.SetHeight(min(10, max(4, state.Height-14)))
	body := []string{
		theme.StyleDim.Render("Lines without # go to " + fallback(state.CaptureTarget, "Inbox")),
		"",
		description.View(),
		"",
	}
	hints, esc := []string{"enter newline"}, ""
	if len(state.CaptureSuggestions) > 0 {
		body = append(body, captureSuggestionLines(theme, state, inner)...)
		hints, esc = []string{"tab insert", "↑/↓ choose"}, "esc dismiss"
	} else {
		body = append(body, theme.StyleDim.Render("#repo/stream  ~45m  @fri  +tag  · indent for a description"))
	}
	return renderListDialog(theme, state, listDialog{
		title: "Add Issues", wide: true, body: body, focus: -1,
		primary: "ctrl+s review", hints: hints, esc: esc,
	})
}

// captureSuggestionLines draws the annotation popup, the cursor entry highlighted.
func captureSuggestionLines(theme Theme, state controllerpkg.State, inner int) []string {
	lines := make([]string, 0, len(state.CaptureSuggestions))
	for i, suggestion := range state.CaptureSuggestions {
		text := ansi.Truncate(suggestion.Insert, max(8, inner-24), "…")
		pad := strings.Repeat(" ", max(2, 28-ansi.StringWidth(text)))
		if i == state.CaptureSuggestCursor {
			lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+text)+pad+theme.StyleDim.Render(suggestion.Detail))
			continue
		}
		lines = append(lines, theme.StyleNormal.Render("  "+text)+pad+theme.StyleDim.Render(suggestion.Detail))
	}
	return lines
}

// renderCaptureReview draws the parsed rows: one line each, the cursor row expanded.
func renderCaptureReview(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	body := []string{}
	focus := -1
	for i, row := range state.CaptureRows {
		mark := "[ ]"
		if len(row.Flags) > 0 {
			mark = "[?]"
		} else if i < len(state.CaptureSelected) && state.CaptureSelected[i] {
			mark = "[x]"
		}
		if i != state.CaptureCursor {
			body = append(body, captureRowLine(theme, mark, row, inner))
			continue
		}
		focus = len(body)
		body = append(body, theme.StyleCursor.Render(ansi.Truncate(viewchrome.SelectionCursor+mark+" "+row.Title, inner, "…")))
		body = append(body, captureRowDetail(theme, row, inner)...)
	}
	selected := controllerpkg.CaptureSelectedCount(state)
	primary := fmt.Sprintf("ctrl+s import %d", selected)
	if state.Processing {
		primary = fallback(state.ProcessingLabel, "Importing…")
	}
	return renderListDialog(theme, state, listDialog{
		title: "Import Issues", context: fmt.Sprintf("%d of %d selected", selected, len(state.CaptureRows)),
		wide: true, body: body, focus: focus, esc: "esc back",
		primary: primary, hints: []string{"space toggle", "a all", "enter edit"},
	})
}

func captureRowLine(theme Theme, mark string, row shareddto.IssueCaptureProposal, inner int) string {
	meta := captureRowMeta(row)
	// "  [x] " leads, and two spaces keep the meta off the title.
	room := max(8, inner-6-2-ansi.StringWidth(meta))
	title := ansi.Truncate(row.Title, room, "…")
	pad := strings.Repeat(" ", room-ansi.StringWidth(title)+2)
	line := "  " + mark + " " + title + pad
	meta = ansi.Truncate(meta, max(1, inner-ansi.StringWidth(line)), "…")
	if len(row.Flags) > 0 {
		return theme.StyleNormal.Render(line) + theme.StyleError.Render(meta)
	}
	return theme.StyleNormal.Render(line) + theme.StyleDim.Render(meta)
}

// captureRowMeta is the compact tail of a collapsed row: why it is flagged, or its target,
// estimate and plan date.
func captureRowMeta(row shareddto.IssueCaptureProposal) string {
	if len(row.Flags) > 0 {
		return strings.Join(row.Flags, ", ")
	}
	parts := []string{row.RepoName + "/" + row.StreamName}
	if row.EstimateMinutes != nil {
		parts = append(parts, controllerpkg.FormatDurationMinutesInput(row.EstimateMinutes))
	}
	if row.TodoForDate != nil {
		parts = append(parts, *row.TodoForDate)
	}
	return strings.Join(parts, "  ")
}

func captureRowDetail(theme Theme, row shareddto.IssueCaptureProposal, inner int) []string {
	field := func(label, value string) string {
		if value == "" {
			value = "—"
		}
		return theme.StyleDim.Render(fmt.Sprintf("     %-11s ", label)) + ansi.Truncate(value, max(4, inner-17), "…")
	}
	lines := []string{}
	if len(row.Flags) > 0 {
		lines = append(lines, theme.StyleError.Render("     "+strings.Join(row.Flags, ", ")+" · enter to fix"))
	}
	lines = append(lines,
		field("Repo", row.RepoName),
		field("Stream", row.StreamName),
		field("Estimate", controllerpkg.FormatDurationMinutesInput(row.EstimateMinutes)),
		field("Due", controllerpkg.ValueOrEmpty(row.TodoForDate)),
	)
	tags := make([]string, len(row.Tags))
	for i, tag := range row.Tags {
		tags[i] = "+" + tag
	}
	lines = append(lines, field("Tags", strings.Join(tags, " ")))
	if row.Description != nil {
		first, _, more := strings.Cut(*row.Description, "\n")
		if more {
			first += " …"
		}
		lines = append(lines, field("Description", first))
	}
	return lines
}
