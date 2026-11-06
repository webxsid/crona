package momentum

import (
	"slices"

	sharedtypes "crona.work/shared/types"
)

// ForHabit are the enabled habit momentums that count this habit.
func ForHabit(cards []sharedtypes.MomentumCard, habitID int64) []sharedtypes.MomentumCard {
	return enabled(cards, func(def sharedtypes.HabitStreakDefinition) bool {
		return def.TargetKind == sharedtypes.MomentumTargetKindHabit &&
			slices.Contains(def.HabitIDs, habitID)
	})
}

// ForContext are the enabled context momentums that sessions in this repo and stream count
// toward: a context matches the repo and either the stream or the whole repo, as the daemon
// counts them.
func ForContext(cards []sharedtypes.MomentumCard, repoID, streamID int64) []sharedtypes.MomentumCard {
	if repoID == 0 {
		return nil
	}
	return enabled(cards, func(def sharedtypes.HabitStreakDefinition) bool {
		if def.TargetKind != sharedtypes.MomentumTargetKindContext {
			return false
		}
		return slices.ContainsFunc(def.Contexts, func(context sharedtypes.MomentumContext) bool {
			return context.RepoID == repoID &&
				(context.StreamID == nil || *context.StreamID == streamID)
		})
	})
}

// PeriodPoint is the card's period that holds date, or its latest one.
func PeriodPoint(card sharedtypes.MomentumCard, date string) (sharedtypes.MomentumSeriesPoint, bool) {
	if len(card.Series) == 0 {
		return sharedtypes.MomentumSeriesPoint{}, false
	}
	for _, point := range card.Series {
		if point.StartDate <= date && date <= point.EndDate {
			return point, true
		}
	}
	return card.Series[len(card.Series)-1], true
}

func enabled(
	cards []sharedtypes.MomentumCard,
	matches func(sharedtypes.HabitStreakDefinition) bool,
) []sharedtypes.MomentumCard {
	out := []sharedtypes.MomentumCard{}
	for _, card := range cards {
		def := sharedtypes.NormalizeHabitStreakDefinition(card.Definition)
		if def.Enabled && matches(def) {
			card.Definition = def
			out = append(out, card)
		}
	}
	return out
}
