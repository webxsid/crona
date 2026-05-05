package wellbeing

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

type LabelTypes string

const (
	LabelTypeMood   LabelTypes = "mood"
	LabelTypeEnergy LabelTypes = "energy"
	LabelTypeSleep  LabelTypes = "sleep"
	LabelTypeWorked LabelTypes = "worked"
	LabelTypeStreak LabelTypes = "streak"
)

func label(label_type LabelTypes, state types.ContentState) string {
	switch label_type {
	case LabelTypeMood:
		return wellbeingMoodLabel(state)
	case LabelTypeEnergy:
		return wellbeingEnergyLabel(state)
	case LabelTypeSleep:
		return wellbeingSleepLabel(state)
	case LabelTypeWorked:
		return wellbeingWorkedLabel(state)
	case LabelTypeStreak:
		return wellbeingStreakLabel(state)
	default:
		return "-"
	}
}

func wellbeingMoodLabel(state types.ContentState) string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		if state.MetricsRollup != nil && state.MetricsRollup.AverageMood != nil {
			return fmt.Sprintf("7d avg %.1f/5", *state.MetricsRollup.AverageMood)
		}
		return "-"
	}
	return fmt.Sprintf("today %d/5", state.DailyCheckIn.Mood)
}

func wellbeingEnergyLabel(state types.ContentState) string {
	if state.DailyCheckIn == nil || state.DailyCheckIn.Date == "" {
		if state.MetricsRollup != nil && state.MetricsRollup.AverageEnergy != nil {
			return fmt.Sprintf("7d avg %.1f/5", *state.MetricsRollup.AverageEnergy)
		}
		return "-"
	}
	return fmt.Sprintf("today %d/5", state.DailyCheckIn.Energy)
}

func wellbeingSleepLabel(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.SleepHours != nil {
		return "today " + helperpkg.FormatCompactDurationHours(*state.DailyCheckIn.SleepHours)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageSleepHours != nil {
		return "7d avg " + helperpkg.FormatCompactDurationHours(*state.MetricsRollup.AverageSleepHours)
	}
	return "-"
}

func wellbeingWorkedLabel(state types.ContentState) string {
	if state.MetricsRollup == nil {
		return "-"
	}
	return viewhelpers.FormatClock(state.MetricsRollup.WorkedSeconds)
}

func wellbeingStreakLabel(state types.ContentState) string {
	if state.Streaks == nil {
		return "-"
	}
	return fmt.Sprintf("C%d/%d F%d/%d H%d/%d", state.Streaks.CurrentCheckInDays, state.Streaks.LongestCheckInDays, state.Streaks.CurrentFocusDays, state.Streaks.LongestFocusDays, state.Streaks.CurrentHabitDays, state.Streaks.LongestHabitDays)
}
