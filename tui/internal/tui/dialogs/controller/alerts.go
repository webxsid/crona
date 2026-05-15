package controller

import (
	"errors"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func OpenCreateAlertReminder(state State) State {
	inputs := []textinput.Model{
		newAlertReminderInput("daily | weekdays | mon,wed,fri"),
		withTimePrompt(state, newAlertReminderInput("20:00")),
	}
	inputs[0].SetValue("daily")
	inputs[0].Focus()
	state = Close(state)
	state.Kind = "create_alert_reminder"
	state.ReminderKind = sharedtypes.AlertReminderKindCheckIn
	state.Inputs = inputs
	state.FocusIdx = 0
	return SyncDialogFocus(state)
}

func OpenEditAlertReminder(state State, reminder sharedtypes.AlertReminder) State {
	next := OpenCreateAlertReminder(state)
	next.Kind = "edit_alert_reminder"
	next.ReminderID = reminder.ID
	next.ReminderKind = reminder.Kind
	next.Inputs[0].SetValue(formatReminderSchedule(reminder))
	next.Inputs[1].SetValue(strings.TrimSpace(reminder.TimeHHMM))
	next.FocusIdx = 0
	return SyncDialogFocus(next)
}

func updateAlertReminder(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		return SyncDialogFocus(clearDialogError(state)), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			scheduleType, weekdays, err := parseAlertReminderSchedule(state.Inputs[0].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			clock, err := ParseClockInput(state.Inputs[1].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			if clock == nil {
				return state, nil, "time is required"
			}
			action := &Action{
				Kind:             "create_alert_reminder",
				ReminderKind:     state.ReminderKind,
				ReminderSchedule: scheduleType,
				ReminderTimeHHMM: *clock,
				Weekdays:         weekdays,
			}
			if state.Kind == "edit_alert_reminder" {
				action.Kind = "edit_alert_reminder"
				action.ID = state.ReminderID
			}
			return Close(state), action, ""
		}
	}
	inputIdx, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func parseAlertReminderSchedule(raw string) (sharedtypes.AlertReminderScheduleType, []int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "", "daily":
		return sharedtypes.AlertReminderScheduleDaily, nil, nil
	case "weekdays":
		return sharedtypes.AlertReminderScheduleWeekly, []int{1, 2, 3, 4, 5}, nil
	default:
		weekdays, err := ParseWeekdayList(value)
		if err != nil {
			return "", nil, err
		}
		if len(weekdays) == 0 {
			return "", nil, errors.New("schedule must be daily, weekdays, or comma-separated weekdays like mon,wed,fri")
		}
		return sharedtypes.AlertReminderScheduleWeekly, weekdays, nil
	}
}

func formatReminderSchedule(reminder sharedtypes.AlertReminder) string {
	switch sharedtypes.NormalizeAlertReminderScheduleType(reminder.ScheduleType) {
	case sharedtypes.AlertReminderScheduleWeekly:
		if len(reminder.Weekdays) == 5 &&
			reminder.Weekdays[0] == 1 &&
			reminder.Weekdays[1] == 2 &&
			reminder.Weekdays[2] == 3 &&
			reminder.Weekdays[3] == 4 &&
			reminder.Weekdays[4] == 5 {
			return "weekdays"
		}
		return strings.Join(WeekdayTokens(reminder.Weekdays), ",")
	default:
		return "daily"
	}
}

func newAlertReminderInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 120
	input.Width = 36
	return input
}
