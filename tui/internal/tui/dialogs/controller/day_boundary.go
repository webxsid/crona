package controller

import (
	"fmt"
	"sort"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	DayBoundaryStepOverview = iota
	DayBoundaryStepDays
	DayBoundaryStepTime
)

// DayBoundaryOverride is a presentation grouping of weekdays that share a
// single override time. The persisted schedule remains a weekday-to-time map.
type DayBoundaryOverride struct {
	Days []int
	Time string
}

func OpenEditDayBoundary(
	state State,
	key sharedtypes.CoreSettingsKey,
	schedule sharedtypes.DayBoundarySchedule,
	timezone string,
) State {
	state = Close(state)
	schedule = schedule.Normalized()
	state.Kind = "edit_day_boundary"
	state.SettingKey = key
	state.DayBoundaryEnabled = schedule.Enabled
	state.DayBoundaryTimezone = timezone
	state.DayBoundaryStep = DayBoundaryStepOverview
	state.DayBoundarySchedule = schedule
	state.DayBoundaryOverrides = dayBoundaryOverrideGroups(schedule)
	state.DayBoundaryOverrideCursor = 0
	if AllDayBoundaryDaysOverridden(schedule) {
		state.DayBoundaryOverrideCursor = 1
	}
	state.DayBoundaryEditingTime = ""
	state.DayBoundaryEditingOverride = false
	state.Inputs = makeDayBoundaryInputs(schedule.DefaultTime, "")
	state.FocusIdx = state.DayBoundaryOverrideCursor
	return SyncDialogFocus(state)
}

func makeDayBoundaryInputs(defaultTime, overrideTime string) []textinput.Model {
	inputs := make([]textinput.Model, 2)
	for i, value := range []string{defaultTime, overrideTime} {
		input := textinput.New()
		input.Prompt = ""
		input.CharLimit = 5
		input.Width = 12
		input.Placeholder = "HH:mm"
		input.SetValue(value)
		inputs[i] = input
	}
	return inputs
}

func updateDayBoundary(state State, msg tea.KeyMsg) (State, *Action, string) {
	key := msg.String()
	if key == "esc" {
		if state.DayBoundaryStep != DayBoundaryStepOverview {
			return dayBoundaryOverview(state), nil, ""
		}
		return Close(state), nil, ""
	}

	switch state.DayBoundaryStep {
	case DayBoundaryStepOverview:
		return updateDayBoundaryOverview(state, key, msg)
	case DayBoundaryStepDays:
		return updateDayBoundaryDays(state, key)
	case DayBoundaryStepTime:
		return updateDayBoundaryTime(state, key, msg)
	default:
		return state, nil, ""
	}
}

func updateDayBoundaryOverview(state State, key string, msg tea.KeyMsg) (State, *Action, string) {
	switch key {
	case "up", "shift+tab":
		state.DayBoundaryOverrideCursor = shiftDayBoundaryOverviewCursor(state, -1)
		state.FocusIdx = state.DayBoundaryOverrideCursor
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case "down", "tab":
		state.DayBoundaryOverrideCursor = shiftDayBoundaryOverviewCursor(state, 1)
		state.FocusIdx = state.DayBoundaryOverrideCursor
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	case " ":
		if state.DayBoundaryOverrideCursor == 0 {
			state.DayBoundaryEnabled = !state.DayBoundaryEnabled
		}
		return clearDialogError(state), nil, ""
	case "a":
		if AllDayBoundaryDaysOverridden(state.DayBoundarySchedule) {
			return state, nil, ""
		}
		return openDayBoundaryDays(state, nil, ""), nil, ""
	case "enter":
		if state.DayBoundaryOverrideCursor > 0 {
			group := state.DayBoundaryOverrides[state.DayBoundaryOverrideCursor-1]
			return openDayBoundaryDays(state, group.Days, group.Time), nil, ""
		}
	case "D", "delete", "backspace":
		if state.DayBoundaryOverrideCursor > 0 {
			group := state.DayBoundaryOverrides[state.DayBoundaryOverrideCursor-1]
			for _, day := range group.Days {
				delete(state.DayBoundarySchedule.WeekdayOverrides, day)
			}
			state.DayBoundaryOverrides = dayBoundaryOverrideGroups(state.DayBoundarySchedule)
			state.DayBoundaryOverrideCursor = minInt(
				state.DayBoundaryOverrideCursor,
				len(state.DayBoundaryOverrides),
			)
			state.FocusIdx = state.DayBoundaryOverrideCursor
			return SyncDialogFocus(clearDialogError(state)), nil, ""
		}
	case "ctrl+s":
		schedule, err := dayBoundaryScheduleFromState(state)
		if err != nil {
			return state, nil, err.Error()
		}
		return Close(state), &Action{
			Kind:                "patch_setting",
			SettingKey:          state.SettingKey,
			DayBoundarySchedule: schedule,
		}, ""
	}
	if state.DayBoundaryOverrideCursor == 0 {
		var cmd tea.Cmd
		state.Inputs[0], cmd = state.Inputs[0].Update(msg)
		_ = cmd
	}
	return clearDialogError(state), nil, ""
}

func updateDayBoundaryDays(state State, key string) (State, *Action, string) {
	switch key {
	case "up":
		state.FocusIdx = nextAvailableDayBoundaryDay(state, -1)
	case "down":
		state.FocusIdx = nextAvailableDayBoundaryDay(state, 1)
	case " ", "x":
		if DayBoundaryDayAvailable(state, state.FocusIdx) {
			state.DayBoundarySelectedDays = toggleWeekday(
				state.DayBoundarySelectedDays,
				state.FocusIdx,
			)
		}
	case "enter", "ctrl+s":
		if len(state.DayBoundarySelectedDays) == 0 {
			return state, nil, "Select at least one day"
		}
		state.DayBoundaryStep = DayBoundaryStepTime
		state.Inputs[1].SetValue(state.DayBoundaryEditingTime)
		state.FocusIdx = 0
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	}
	return SyncDialogFocus(clearDialogError(state)), nil, ""
}

func updateDayBoundaryTime(state State, key string, msg tea.KeyMsg) (State, *Action, string) {
	if key == "ctrl+s" || key == "enter" {
		timeValue := strings.TrimSpace(state.Inputs[1].Value())
		if !validDayBoundaryTime(timeValue) {
			return state, nil, fmt.Sprintf("invalid override time %q: expected HH:mm", timeValue)
		}
		if state.DayBoundarySchedule.WeekdayOverrides == nil {
			state.DayBoundarySchedule.WeekdayOverrides = map[int]string{}
		}
		if state.DayBoundaryEditingOverride {
			for day, existing := range state.DayBoundarySchedule.WeekdayOverrides {
				if existing == state.DayBoundaryEditingTime && containsInt(state.DayBoundarySelectedDays, day) {
					delete(state.DayBoundarySchedule.WeekdayOverrides, day)
				}
			}
		}
		for _, day := range state.DayBoundarySelectedDays {
			state.DayBoundarySchedule.WeekdayOverrides[day] = timeValue
		}
		state.DayBoundaryOverrides = dayBoundaryOverrideGroups(state.DayBoundarySchedule)
		return dayBoundaryOverview(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[1], cmd = state.Inputs[1].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func openDayBoundaryDays(state State, selected []int, timeValue string) State {
	state.DayBoundaryStep = DayBoundaryStepDays
	state.DayBoundarySelectedDays = append([]int(nil), selected...)
	state.DayBoundaryEditingTime = timeValue
	state.DayBoundaryEditingOverride = len(selected) > 0
	if len(selected) == 0 {
		state.DayBoundarySelectedDays = nil
	}
	state.FocusIdx = firstAvailableDayBoundaryDay(state)
	return SyncDialogFocus(clearDialogError(state))
}

func dayBoundaryOverview(state State) State {
	state.DayBoundaryStep = DayBoundaryStepOverview
	state.DayBoundaryEditingTime = ""
	state.DayBoundaryEditingOverride = false
	state.DayBoundarySelectedDays = nil
	state.DayBoundaryOverrideCursor = minInt(state.DayBoundaryOverrideCursor, len(state.DayBoundaryOverrides))
	state.FocusIdx = state.DayBoundaryOverrideCursor
	return SyncDialogFocus(clearDialogError(state))
}

func shiftDayBoundaryOverviewCursor(state State, direction int) int {
	if AllDayBoundaryDaysOverridden(state.DayBoundarySchedule) && len(state.DayBoundaryOverrides) > 0 {
		current := state.DayBoundaryOverrideCursor - 1
		current = (current + direction + len(state.DayBoundaryOverrides)) % len(state.DayBoundaryOverrides)
		return current + 1
	}
	total := len(state.DayBoundaryOverrides) + 1
	return (state.DayBoundaryOverrideCursor + direction + total) % total
}

func dayBoundaryScheduleFromState(state State) (sharedtypes.DayBoundarySchedule, error) {
	if len(state.Inputs) < 1 {
		return sharedtypes.DayBoundarySchedule{}, fmt.Errorf("schedule editor is unavailable")
	}
	schedule := state.DayBoundarySchedule
	schedule.Enabled = state.DayBoundaryEnabled
	schedule.DefaultTime = strings.TrimSpace(state.Inputs[0].Value())
	if err := schedule.Validate(); err != nil {
		return sharedtypes.DayBoundarySchedule{}, err
	}
	return schedule.Normalized(), nil
}

func dayBoundaryOverrideGroups(schedule sharedtypes.DayBoundarySchedule) []DayBoundaryOverride {
	byTime := map[string][]int{}
	for day, value := range schedule.WeekdayOverrides {
		byTime[value] = append(byTime[value], day)
	}
	times := make([]string, 0, len(byTime))
	for value := range byTime {
		times = append(times, value)
	}
	sort.Strings(times)
	groups := make([]DayBoundaryOverride, 0, len(times))
	for _, value := range times {
		days := byTime[value]
		sort.Ints(days)
		groups = append(groups, DayBoundaryOverride{Days: days, Time: value})
	}
	return groups
}

func DayBoundaryDayAvailable(state State, day int) bool {
	if day < 1 || day > 7 {
		return false
	}
	if state.DayBoundaryEditingOverride && containsInt(state.DayBoundarySelectedDays, day) {
		return true
	}
	_, overridden := state.DayBoundarySchedule.WeekdayOverrides[day]
	return !overridden
}

func firstAvailableDayBoundaryDay(state State) int {
	for day := 1; day <= 7; day++ {
		if DayBoundaryDayAvailable(state, day) {
			return day
		}
	}
	return 1
}

func nextAvailableDayBoundaryDay(state State, direction int) int {
	for offset := 1; offset <= 7; offset++ {
		day := ((state.FocusIdx-1+direction*offset)%7+7)%7 + 1
		if DayBoundaryDayAvailable(state, day) {
			return day
		}
	}
	return state.FocusIdx
}

func AllDayBoundaryDaysOverridden(schedule sharedtypes.DayBoundarySchedule) bool {
	return len(schedule.WeekdayOverrides) == 7
}

func validDayBoundaryTime(value string) bool {
	return sharedtypes.DayBoundarySchedule{DefaultTime: value}.Validate() == nil
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
