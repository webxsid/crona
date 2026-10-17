package daily

import (
	"fmt"

	helperpkg "crona/tui/internal/tui/helpers"
	types "crona/tui/internal/tui/views/types"
)

func signalMoodValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" {
		return fmt.Sprintf("today %d/5", state.DailyCheckIn.Mood)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageMood != nil {
		return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageMood)
	}
	return ""
}

func signalEnergyValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" {
		return fmt.Sprintf("today %d/5", state.DailyCheckIn.Energy)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageEnergy != nil {
		return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageEnergy)
	}
	return ""
}

func signalSleepValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" &&
		state.DailyCheckIn.SleepHours != nil {
		return "today " + helperpkg.FormatCompactDurationHours(*state.DailyCheckIn.SleepHours)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageSleepHours != nil {
		return "avg " + helperpkg.FormatCompactDurationHours(*state.MetricsRollup.AverageSleepHours)
	}
	return ""
}
