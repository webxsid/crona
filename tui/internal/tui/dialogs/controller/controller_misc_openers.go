package controller

import (
	"strconv"
	"strings"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
)

func OpenEditTelemetrySettings(state State, usageEnabled, errorReportingEnabled bool) State {
	state = Close(state)
	state.Kind = "edit_telemetry_settings"
	state.TelemetryStep = 0
	state.TelemetryUsage = usageEnabled
	state.TelemetryErrors = errorReportingEnabled
	return state
}

func OpenOnboarding(state State, usageEnabled, errorReportingEnabled bool) State {
	state = Close(state)
	state.Kind = "onboarding"
	state.TelemetryStep = 0
	state.TelemetryUsage = usageEnabled
	state.TelemetryErrors = errorReportingEnabled
	return state
}

func OpenCreateMomentumDirect(
	state State,
	defs []api.HabitStreakDefinition,
	habits []api.HabitWithMeta,
) State {
	state = closeAndPrimeMomentumState(state, defs, habits)
	state.Kind = "create_momentum"
	return openMomentumEditor(state, -1, sharedtypes.HabitStreakDefinition{
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	})
}

func OpenEditMomentumDirect(
	state State,
	defs []api.HabitStreakDefinition,
	habits []api.HabitWithMeta,
	def sharedtypes.HabitStreakDefinition,
) State {
	state = closeAndPrimeMomentumState(state, defs, habits)
	state.Kind = "edit_momentum"
	idx := -1
	for i, item := range state.HabitStreakDefs {
		if item.ID == def.ID {
			idx = i
			break
		}
	}
	return openMomentumEditor(state, idx, def)
}

func openMomentumEditor(state State, idx int, def sharedtypes.HabitStreakDefinition) State {
	state = openHabitStreakEditor(state, idx, def)
	state.Inputs[0].Placeholder = "Momentum name"
	state.Description = newDescriptionInput(36, 4)
	if def.Description != nil {
		state.Description.SetValue(strings.TrimSpace(*def.Description))
	}
	state.DescriptionEnabled = true
	state.DescriptionIndex = 1
	state = SyncDialogFocus(state)
	return state
}

func closeAndPrimeMomentumState(
	state State,
	defs []api.HabitStreakDefinition,
	habits []api.HabitWithMeta,
) State {
	state = Close(state)
	state.HabitItems = append([]sharedtypes.HabitWithMeta(nil), habits...)
	state.HabitStreakOriginalDefs = append([]sharedtypes.HabitStreakDefinition(nil), defs...)
	state.HabitStreakDefs = append([]sharedtypes.HabitStreakDefinition(nil), defs...)
	state.HabitStreakOriginalDefs = sharedtypes.NormalizeHabitStreakDefinitions(
		state.HabitStreakOriginalDefs,
	)
	state.HabitStreakDefs = sharedtypes.NormalizeHabitStreakDefinitions(state.HabitStreakDefs)
	state.HabitStreakStep = 1
	state.HabitStreakCursor = 0
	state.HabitStreakEditIdx = -1
	state.HabitStreakDraft = sharedtypes.HabitStreakDefinition{
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	}
	return state
}

func OpenCreateCheckIn(state State, date string) State {
	return openCheckInDialog(state, "create_checkin", date, nil)
}

func OpenEditCheckIn(state State, checkIn *api.DailyCheckIn, date string) State {
	return openCheckInDialog(state, "edit_checkin", date, checkIn)
}

func openCheckInDialog(state State, kind string, date string, checkIn *api.DailyCheckIn) State {
	mood := textinput.New()
	mood.Placeholder = "Mood 1-5"
	mood.CharLimit = 1
	mood.Width = 12
	mood.Focus()
	energy := textinput.New()
	energy.Placeholder = "Energy 1-5"
	energy.CharLimit = 1
	energy.Width = 12
	sleepHours := textinput.New()
	sleepHours.Placeholder = "7.5h | 7h30m | 450m"
	sleepHours.CharLimit = 8
	sleepHours.Width = 20
	sleepHours = withTimePrompt(state, sleepHours)
	sleepScore := textinput.New()
	sleepScore.Placeholder = "Sleep score"
	sleepScore.CharLimit = 3
	sleepScore.Width = 16
	screenTime := textinput.New()
	screenTime.Placeholder = "45m | 1h20m"
	screenTime.CharLimit = 8
	screenTime.Width = 20
	screenTime = withTimePrompt(state, screenTime)
	notes := textinput.New()
	notes.Placeholder = "Notes (optional)"
	notes.CharLimit = 200
	notes.Width = 52
	if checkIn != nil {
		mood.SetValue(strconv.Itoa(checkIn.Mood))
		energy.SetValue(strconv.Itoa(checkIn.Energy))
		if checkIn.SleepHours != nil {
			sleepHours.SetValue(FormatDurationHoursInput(checkIn.SleepHours))
		}
		if checkIn.SleepScore != nil {
			sleepScore.SetValue(strconv.Itoa(*checkIn.SleepScore))
		}
		if checkIn.ScreenTimeMinutes != nil {
			screenTime.SetValue(FormatDurationMinutesInput(checkIn.ScreenTimeMinutes))
		}
		if checkIn.Notes != nil {
			notes.SetValue(strings.TrimSpace(*checkIn.Notes))
		}
	}
	state = Close(state)
	state.Kind = kind
	state.CheckInDate = date
	state.Inputs = []textinput.Model{mood, energy, sleepHours, sleepScore, screenTime, notes}
	return state
}
