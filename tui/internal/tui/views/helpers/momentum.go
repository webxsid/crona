package viewhelpers

import sharedtypes "crona/shared/types"

// VisibleMomentumCards are the cards the list filter keeps, in load order, as the cursor indexes
// them.
func VisibleMomentumCards(
	cards []sharedtypes.MomentumCard,
	filter string,
) []sharedtypes.MomentumCard {
	names := make([]string, len(cards))
	for i, card := range cards {
		names[i] = card.Definition.Name
	}
	out := []sharedtypes.MomentumCard{}
	for _, i := range FilteredStrings(names, filter) {
		out = append(out, cards[i])
	}
	return out
}
