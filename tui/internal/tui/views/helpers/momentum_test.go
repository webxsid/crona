package viewhelpers

import (
	"testing"

	sharedtypes "crona/shared/types"
)

func TestVisibleMomentumCardsKeepsOrderAndMatches(t *testing.T) {
	cards := []sharedtypes.MomentumCard{{}, {}, {}}
	cards[0].Definition.Name, cards[1].Definition.Name, cards[2].Definition.Name = "Deep Work", "Walk", "Deep sleep"
	got := VisibleMomentumCards(cards, " DEEP ")
	if len(got) != 2 || got[0].Definition.Name != "Deep Work" ||
		got[1].Definition.Name != "Deep sleep" {
		t.Fatalf("expected the two deep cards in order, got %+v", got)
	}
	if len(VisibleMomentumCards(cards, "")) != 3 {
		t.Fatalf("expected every card with no filter")
	}
}
