package dispatch

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/commands"
	uistate "crona/tui/internal/tui/state"
)

func momentumCard(name string, enabled bool) api.MomentumCard {
	card := api.MomentumCard{}
	card.Definition.Name = name
	card.Definition.Enabled = enabled
	return card
}

func TestMomentumCardsLoadActiveFirst(t *testing.T) {
	loaded := []api.MomentumCard{momentumCard("A", false), momentumCard("B", true), momentumCard("C", false), momentumCard("D", true)}
	state, _, _ := HandleMessage(MessageState{}, commands.MomentumRangeLoadedMsg{Cards: loaded}, MessageDeps{})
	var names []string
	for _, card := range state.MomentumCards {
		names = append(names, card.Definition.Name)
	}
	if got := len(names); got != 4 || names[0] != "B" || names[1] != "D" || names[2] != "A" || names[3] != "C" {
		t.Fatalf("expected active momentums first in their order, got %v", names)
	}
	if loaded[0].Definition.Name != "A" {
		t.Fatalf("expected the loaded slice untouched, got %v", loaded[0].Definition.Name)
	}
}

func TestMomentumCardsLoadStillClamps(t *testing.T) {
	clamped := false
	HandleMessage(MessageState{Cursor: map[uistate.Pane]int{uistate.PaneMomentumCards: 9}},
		commands.MomentumRangeLoadedMsg{Cards: []api.MomentumCard{momentumCard("A", true), momentumCard("B", false)}},
		MessageDeps{ClampFiltered: func(_ *MessageState, pane uistate.Pane) { clamped = clamped || pane == uistate.PaneMomentumCards }})
	if !clamped {
		t.Fatalf("expected the momentum cursor to be clamped")
	}
}
