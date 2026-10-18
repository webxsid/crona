package viewchrome

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// GroupRow is one selectable list row and the group it sits under.
type GroupRow struct {
	Group, Line string
	Danger      bool
}

// GroupList draws the cursor's window of rows, at most height lines, with a dim divider
// before each group; Danger groups' dividers are red.
func GroupList(theme Theme, rows []GroupRow, cursor, width, height int) []string {
	lines := make([]dayLine, 0, len(rows)+8)
	for i, row := range rows {
		if i == 0 || row.Group != rows[i-1].Group {
			style := theme.StyleDim
			if row.Danger {
				style = lipgloss.NewStyle().Foreground(theme.ColorRed)
			}
			text := ansi.Truncate("── "+row.Group+" ", width, "")
			lines = append(lines, dayLine{text: style.Render(text + strings.Repeat("─", max(0, width-ansi.StringWidth(text)))), row: -1})
		}
		line := row.Line
		if row.Danger {
			line = lipgloss.NewStyle().Foreground(theme.ColorRed).Render(line)
		}
		lines = append(lines, dayLine{text: line, row: i})
	}
	return dividedList(theme, lines, cursor, width, height)
}

// RowHelp explains one setting: what it does, its choices and its keys.
type RowHelp struct {
	Description string
	Options     []string
	Keys        []string
}

// HelpLines renders a setting's help: the description, then its options with the current
// one marked (or the whole current value when it has none), then its keys.
func HelpLines(theme Theme, help RowHelp, current string, width int) []string {
	width = max(1, width)
	lines := []string{}
	if help.Description != "" {
		lines = append(lines, strings.Split(ansi.Wrap(help.Description, width, ""), "\n")...)
		lines = append(lines, "")
	}
	if len(help.Options) > 0 {
		for _, option := range help.Options {
			mark := theme.StyleDim.Render("○ ")
			if strings.EqualFold(strings.TrimSpace(option), strings.TrimSpace(current)) {
				mark = "● "
			}
			lines = append(lines, mark+ansi.Truncate(option, max(1, width-2), "…"))
		}
	} else if current != "" {
		lines = append(lines, strings.Split(ansi.Hardwrap(current, width, false), "\n")...)
	}
	if len(help.Keys) > 0 {
		lines = append(lines, "", theme.StyleDim.Render(ansi.Truncate(strings.Join(help.Keys, "  "), width, "…")))
	}
	return lines
}

// LabelColumn is the width of the longest label, capped at limit.
func LabelColumn(labels []string, limit int) int {
	width := 0
	for _, label := range labels {
		width = max(width, ansi.StringWidth(label))
	}
	return min(width, limit)
}
