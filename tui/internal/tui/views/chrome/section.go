package viewchrome

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Section is a borderless pane: a rule-line header over an indented body.
type Section struct {
	Title   string
	Meta    string
	Focused bool
	// TitleStyle overrides the focused/unfocused title style when set.
	TitleStyle *lipgloss.Style
	Width      int
	Height     int
	Lines      []string
}

func RenderSection(theme Theme, s Section) string {
	if s.Width < 1 || s.Height < 1 {
		return ""
	}
	rows := []string{sectionHeader(theme, s)}
	bodyHeight := s.Height - 1
	body := s.Lines
	if len(body) > bodyHeight {
		body = append([]string{}, body[:max(0, bodyHeight-1)]...)
		if bodyHeight > 0 {
			body = append(body, theme.StyleDim.Render("↓ more"))
		}
	}
	for _, line := range body {
		rows = append(rows, fitCells(" "+line, s.Width))
	}
	for len(rows) < s.Height {
		rows = append(rows, strings.Repeat(" ", s.Width))
	}
	return strings.Join(rows, "\n")
}

func sectionHeader(theme Theme, s Section) string {
	titleStyle := theme.StyleDim.Bold(true)
	if s.Focused {
		titleStyle = theme.StylePaneTitle
	}
	if s.TitleStyle != nil {
		titleStyle = *s.TitleStyle
	}
	lead := "─"
	title := " " + s.Title + " "
	meta := ""
	if s.Meta != "" {
		meta = " " + s.Meta
	}
	fixed := ansi.StringWidth(lead) + ansi.StringWidth(title) + 1
	if fixed+ansi.StringWidth(meta) > s.Width {
		meta = ""
	}
	if fixed > s.Width {
		title = ansi.Truncate(title, max(0, s.Width-ansi.StringWidth(lead)-1), "…")
	}
	fill := s.Width - ansi.StringWidth(lead) - ansi.StringWidth(title) - ansi.StringWidth(meta)
	header := theme.StyleDim.Render(lead) +
		titleStyle.Render(title) +
		theme.StyleDim.Render(strings.Repeat("─", max(0, fill))+meta)
	return fitCells(header, s.Width)
}

// JoinColumns places rendered sections side by side with a dim gutter.
func JoinColumns(theme Theme, height int, columns ...string) string {
	gutterRows := make([]string, max(0, height))
	for i := range gutterRows {
		gutterRows[i] = " │ "
	}
	gutter := theme.StyleDim.Render(strings.Join(gutterRows, "\n"))
	parts := make([]string, 0, len(columns)*2)
	for i, column := range columns {
		if i > 0 {
			parts = append(parts, gutter)
		}
		parts = append(parts, column)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func fitCells(s string, width int) string {
	if ansi.StringWidth(s) > width {
		return ansi.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-ansi.StringWidth(s))
}
