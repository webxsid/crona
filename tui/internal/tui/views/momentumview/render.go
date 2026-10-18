// Package momentumview renders the Insights tab's Momentum sub-tab: a strip, the momentum
// list and the selected momentum's detail.
package momentumview

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

const (
	minListWidth = 40
	maxListWidth = 56
)

// listWidthFor gives the list 3/10 of the screen, between the floor and the cap.
func listWidthFor(width int) int {
	return min(maxListWidth, max(minListWidth, width*3/10))
}

// Render shows the strip, the list, and the detail when the width allows.
func Render(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	rows := StripRows(theme, state, state.Width)
	rows = rows[:min(len(rows), state.Height)]
	height := state.Height - len(rows)
	if height == 0 {
		return strings.Join(rows, "\n")
	}
	cards := visibleCards(state)
	cursor := min(max(0, state.Cursors["momentum_cards"]), max(0, len(cards)-1))
	if viewhelpers.ClassFor(state.Width) == viewhelpers.WidthNarrow {
		meta := viewchrome.KeyHint(theme, "enter", "details")
		return strings.Join(append(rows, renderList(theme, cards, cursor, state.Width, height, meta)), "\n")
	}
	listWidth := listWidthFor(state.Width)
	list := renderList(theme, cards, cursor, listWidth, height, fmt.Sprintf("%d", len(cards)))
	detailWidth := state.Width - listWidth - 3
	var detail string
	if len(cards) == 0 {
		detail = viewchrome.RenderSection(theme, viewchrome.Section{Title: "Momentum", Width: detailWidth, Height: height})
	} else {
		card := cards[cursor]
		toggle := "disable"
		if !card.Definition.Enabled {
			toggle = "enable"
		}
		meta := viewchrome.FitHints([]string{viewchrome.KeyHint(theme, "e", "edit"), viewchrome.KeyHint(theme, "x", toggle)}, max(0, detailWidth-len(card.Definition.Name)-4))
		titleStyle := theme.StyleHeader.Bold(true)
		detail = viewchrome.RenderSection(theme, viewchrome.Section{
			Title: card.Definition.Name, Meta: meta, TitleStyle: &titleStyle, Width: detailWidth, Height: height,
			Lines: DetailLines(theme, card, detailWidth-2, height-1),
		})
	}
	return strings.Join(append(rows, viewchrome.JoinColumns(theme, height, list, detail)), "\n")
}

// visibleCards are the momentums the list filter keeps, in load order.
func visibleCards(state types.ContentState) []sharedtypes.MomentumCard {
	query := strings.ToLower(strings.TrimSpace(state.Filters["momentum_cards"]))
	if query == "" {
		return state.MomentumCards
	}
	out := []sharedtypes.MomentumCard{}
	for _, card := range state.MomentumCards {
		if strings.Contains(strings.ToLower(card.Definition.Name), query) {
			out = append(out, card)
		}
	}
	return out
}

func renderList(theme types.Theme, cards []sharedtypes.MomentumCard, cursor, width, height int, meta string) string {
	create := viewchrome.KeyHint(theme, "a", "new momentum")
	lines := []string{}
	if len(cards) == 0 {
		lines = append(lines, theme.StyleDim.Render("No momentums yet")+"  "+create)
	} else {
		lines = append(lines, create)
		start, end := viewchrome.ListWindow(cursor, len(cards), max(1, height-2))
		for i := start; i < end; i++ {
			row := listRow(theme, cards[i], max(1, width-4))
			if i == cursor {
				lines = append(lines, theme.StyleCursor.Render(viewchrome.SelectionCursor+" ")+row)
			} else {
				lines = append(lines, "  "+row)
			}
		}
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: "Momentums", Meta: meta, Focused: true, Width: width, Height: height, Lines: lines,
	})
}
