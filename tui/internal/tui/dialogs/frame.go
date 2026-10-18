package dialogs

import (
	"fmt"
	"strings"

	viewchrome "crona/tui/internal/tui/views/chrome"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	frameMinInner = 44
	frameMaxInner = 76
	// frameChrome is the border and padding around the inner width: 1+2 a side.
	frameChrome = 6
	// pickerShown is how many picker options a focused picker lists.
	pickerShown = 5
)

// FrameSpec is one dialog: its border title and context, body lines and key footer.
// Focus is the body line kept visible when the body is cut to fit; -1 means none.
type FrameSpec struct {
	Title, Context            string
	Danger                    bool
	Body, Footer              []string
	ScreenWidth, ScreenHeight int
	Focus                     int
	// Inner, when set, fixes the inner width (still clamped) for forms that lay out to it.
	Inner int
}

// FrameInner is the inner width a frame on this screen gives content of these lines.
func FrameInner(screenWidth int, content []string) int {
	widest := 0
	for _, line := range content {
		for _, part := range strings.Split(line, "\n") {
			widest = max(widest, ansi.StringWidth(part))
		}
	}
	inner := min(max(widest, frameMinInner), frameMaxInner)
	return max(1, min(inner, screenWidth-8-frameChrome))
}

// Frame draws a dialog: a rounded border with the title (and context, when it fits) in
// the top edge, padded body and footer, cut around Focus when taller than the screen.
func Frame(theme Theme, spec FrameSpec) string {
	inner := FrameInner(spec.ScreenWidth, append(append([]string{}, spec.Body...), spec.Footer...))
	if spec.Inner > 0 {
		inner = FrameInner(spec.ScreenWidth, []string{strings.Repeat(" ", spec.Inner)})
	}
	body, footer := wrapAll(spec.Body, inner), wrapAll(spec.Footer, inner)
	focus := spec.Focus
	if focus >= 0 {
		focus = wrappedIndex(spec.Body, inner, focus)
	}
	if limit := spec.ScreenHeight - 4; limit > 0 {
		budget := limit - 4 - len(footer)
		if len(footer) > 0 {
			budget--
		}
		if len(body) > budget {
			body = windowLines(theme, body, max(0, focus), max(1, budget))
		}
	}
	border := lipgloss.NewStyle().Foreground(theme.ColorCyan)
	if spec.Danger {
		border = lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
	side := border.Render("│")
	row := func(line string) string {
		return side + "  " + line + strings.Repeat(" ", max(0, inner-ansi.StringWidth(line))) + "  " + side
	}
	rows := []string{frameTop(theme, border, spec.Title, spec.Context, inner+frameChrome), row("")}
	for _, line := range body {
		rows = append(rows, row(line))
	}
	if len(footer) > 0 {
		rows = append(rows, row(""))
		for _, line := range footer {
			rows = append(rows, row(line))
		}
	}
	rows = append(rows, row(""), border.Render("╰"+strings.Repeat("─", inner+frameChrome-2)+"╯"))
	return strings.Join(rows, "\n")
}

func frameTop(theme Theme, border lipgloss.Style, title, context string, width int) string {
	title = ansi.Truncate(title, max(1, width-6), "…")
	left := "╭─ " + title + " "
	right := " " + context + " ─╮"
	if context == "" || ansi.StringWidth(left)+ansi.StringWidth(right)+1 > width {
		context, right = "", "─╮"
	}
	fill := strings.Repeat("─", max(0, width-ansi.StringWidth(left)-ansi.StringWidth(right)))
	out := border.Render("╭─ ") + theme.StylePaneTitle.Render(title) + border.Render(" "+fill)
	if context != "" {
		return out + " " + theme.StyleDim.Render(context) + border.Render(" ─╮")
	}
	return out + border.Render("─╮")
}

func wrapAll(lines []string, width int) []string {
	out := []string{}
	for _, line := range lines {
		out = append(out, strings.Split(ansi.Wrap(line, width, ""), "\n")...)
	}
	return out
}

// wrappedIndex maps a source line index to the first wrapped line it produced.
func wrappedIndex(lines []string, width, index int) int {
	out := 0
	for i, line := range lines {
		if i == index {
			return out
		}
		out += len(strings.Split(ansi.Wrap(line, width, ""), "\n"))
	}
	return out
}

// windowLines keeps budget lines around focus, marking cut edges with ↑ more and ↓ more.
func windowLines(theme Theme, lines []string, focus, budget int) []string {
	if budget < 3 {
		return lines[:min(len(lines), budget)]
	}
	start, end := viewchrome.ListWindow(focus, len(lines), budget-2)
	out := []string{}
	if start > 0 {
		out = append(out, theme.StyleDim.Render("↑ more"))
	}
	out = append(out, lines[start:end]...)
	if end < len(lines) {
		out = append(out, theme.StyleDim.Render("↓ more"))
	}
	return out
}

// Field is one `label  value` row, the label padded to column and accented when focused.
func Field(theme Theme, label string, column int, value string, focused bool) string {
	style := theme.StyleDim
	if focused {
		style = theme.StylePaneTitle
	}
	return style.Render(label) + strings.Repeat(" ", max(0, column-ansi.StringWidth(label))+2) + value
}

// FieldPair puts two field rows side by side, or stacks them below an inner width of 56.
func FieldPair(left, right string, inner int) []string {
	if inner < 56 {
		return []string{left, right}
	}
	half := inner / 2
	left = ansi.Truncate(left, half-2, "…")
	return []string{left + strings.Repeat(" ", max(2, half-ansi.StringWidth(left))) + right}
}

// Textarea is a label row over the body, indented by 2.
func Textarea(theme Theme, label string, column int, body string, focused bool, inner int) []string {
	lines := []string{Field(theme, label, column, "", focused)}
	for _, line := range strings.Split(body, "\n") {
		lines = append(lines, "  "+ansi.Truncate(line, max(1, inner-2), ""))
	}
	return lines
}

// PickerSpec is a searchable choice: Input is the rendered filter input shown while focused.
type PickerSpec struct {
	Label       string
	Column      int
	Focused     bool
	Input       string
	Current     string
	Options     []string
	Selected    int
	Placeholder string
}

// Picker shows `label  current ▾` collapsed, or the filter input over up to 5 matches.
func Picker(theme Theme, spec PickerSpec) []string {
	if !spec.Focused || len(spec.Options) == 0 {
		value := theme.StyleDim.Render(spec.Placeholder)
		if spec.Current != "" && len(spec.Options) > 0 {
			value = spec.Current + theme.StyleDim.Render(" ▾")
		}
		if spec.Focused {
			value = spec.Input + "  " + value
		}
		return []string{Field(theme, spec.Label, spec.Column, value, spec.Focused)}
	}
	lines := []string{Field(theme, spec.Label, spec.Column, spec.Input, true)}
	indent := strings.Repeat(" ", spec.Column+2)
	selected := min(max(0, spec.Selected), len(spec.Options)-1)
	start, end := viewchrome.ListWindow(selected, len(spec.Options), pickerShown)
	for i := start; i < end; i++ {
		if i == selected {
			lines = append(lines, indent+theme.StyleCursor.Render("▶ "+spec.Options[i]))
			continue
		}
		lines = append(lines, indent+theme.StyleDim.Render("  "+spec.Options[i]))
	}
	if more := len(spec.Options) - (end - start); more > 0 {
		lines = append(lines, indent+theme.StyleDim.Render(fmt.Sprintf("  +%d more", more)))
	}
	return lines
}

// Choice is one fixed option: its key, label and the help shown while it is selected.
type Choice struct {
	Key, Label, Help string
}

// Choices lists fixed options with ▶ on the selected one, then its help, dim.
func Choices(theme Theme, choices []Choice, selected, inner int) []string {
	lines := []string{}
	for i, choice := range choices {
		row := "[" + choice.Key + "] " + choice.Label
		if i == selected {
			lines = append(lines, theme.StyleCursor.Render("▶ "+row))
			continue
		}
		lines = append(lines, "  "+row)
	}
	if selected >= 0 && selected < len(choices) && choices[selected].Help != "" {
		lines = append(lines, "")
		for _, line := range strings.Split(ansi.Wrap(choices[selected].Help, max(1, inner), ""), "\n") {
			lines = append(lines, theme.StyleDim.Render(line))
		}
	}
	return lines
}

// Footer packs the primary action, the other hints and `esc cancel` into whole-hint lines.
func Footer(theme Theme, primary string, hints []string, inner int) []string {
	all := append(append([]string{primary}, hints...), "esc cancel")
	lines := viewchrome.PackHints(all, inner)
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, primary) {
			lines[i] = theme.StyleHeader.Render(primary) + theme.StyleDim.Render(strings.TrimPrefix(line, primary))
			continue
		}
		lines[i] = theme.StyleDim.Render(line)
	}
	return lines
}
