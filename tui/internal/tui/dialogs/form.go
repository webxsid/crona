package dialogs

import (
	"strings"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
)

// formWidth is the content width forms lay their fields out to before clamping.
const formWidth = 72

// FormRow is a block of body lines; Focus maps a dialog FocusIdx to its line in Lines.
type FormRow struct {
	Lines []string
	Focus map[int]int
}

// FormSpec is one form dialog: its frame, rows and footer hints per focus.
type FormSpec struct {
	Title, Context string
	Border         lipgloss.Color
	Rows           []FormRow
	Primary        string
	Hints          func(focus int) []string
	FocusCount     int
	// Esc replaces the closing `esc cancel` hint, for forms where esc goes back.
	Esc string
}

func fieldRow(focus int, line string) FormRow {
	return FormRow{Lines: []string{line}, Focus: map[int]int{focus: 0}}
}

// pairRow puts two fields on one row, or two when FieldPair stacks them.
func pairRow(left, right string, leftFocus, rightFocus, inner int) FormRow {
	lines := FieldPair(left, right, inner)
	row := FormRow{Lines: lines, Focus: map[int]int{leftFocus: 0, rightFocus: 0}}
	if len(lines) > 1 {
		row.Focus[rightFocus] = 1
	}
	return row
}

func staticRow(lines ...string) FormRow { return FormRow{Lines: lines} }

func textareaRow(focus int, lines []string) FormRow {
	return FormRow{Lines: lines, Focus: map[int]int{focus: 0}}
}

// formBody lays rows out top to bottom and returns the focused row's line, or -1.
func formBody(rows []FormRow, focusIdx int) ([]string, int) {
	body, focusLine := []string{}, -1
	for _, row := range rows {
		if offset, ok := row.Focus[focusIdx]; ok {
			focusLine = len(body) + offset
		}
		body = append(body, row.Lines...)
	}
	return body, focusLine
}

// formFooter renders the focused field's hints, padded to the tallest focus so the
// dialog never changes height while tabbing.
func formFooter(
	theme Theme,
	primary string,
	hints func(focus int) []string,
	esc string,
	focus, focusCount, inner int,
) []string {
	if esc == "" {
		esc = "esc cancel"
	}
	hintsAt := func(f int) []string {
		if hints == nil {
			return nil
		}
		return hints(f)
	}
	footer := footerWithEsc(theme, primary, hintsAt(focus), esc, inner)
	tallest := len(footer)
	for f := 0; f < focusCount; f++ {
		tallest = max(tallest, len(footerWithEsc(theme, primary, hintsAt(f), esc, inner)))
	}
	for len(footer) < tallest {
		footer = append(footer, "")
	}
	return footer
}

func renderForm(theme Theme, state controllerpkg.State, spec FormSpec) string {
	inner := formInner(state.Width)
	body, focus := formBody(spec.Rows, state.FocusIdx)
	footer := formFooter(theme, spec.Primary, spec.Hints, spec.Esc, state.FocusIdx, spec.FocusCount, inner)
	return Frame(theme, FrameSpec{
		Title: spec.Title, Context: spec.Context, Border: spec.Border, Body: body,
		Footer: withError(
			theme,
			state,
			footer,
		), ScreenWidth: state.Width, ScreenHeight: state.Height,
		Focus: focus, Inner: inner,
	})
}

func formInner(width int) int {
	return FrameInner(width, []string{strings.Repeat(" ", formWidth)})
}

// formColumns gives a field's value width on its own row and inside a pair.
func formColumns(inner, column int) (full, half int) {
	full = inner - column - 3
	half = inner/2 - column - 6
	if inner < 56 {
		half = full
	}
	return full, half
}

// plainHint turns `[k] action` into `k action`.
func plainHint(hint string) string {
	return strings.Replace(strings.Replace(hint, "[", "", 1), "]", "", 1)
}
