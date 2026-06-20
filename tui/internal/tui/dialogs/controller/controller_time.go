package controller

import (
	"strconv"
	"strings"
	"time"

	shareddto "crona/shared/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func updateCheckIn(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		return Close(state), nil, ""
	case dialogActionFocusNext, dialogActionFocusPrev, dialogActionMoveDown, dialogActionMoveUp:
		dir := dialogFocusMoveDir(action)
		if dir == 0 {
			dir = dialogVerticalMoveDir(action)
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(state.Inputs)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if action == dialogActionPrimary {
			mood, err := strconv.Atoi(strings.TrimSpace(state.Inputs[0].Value()))
			if err != nil || mood < 1 || mood > 5 {
				return state, nil, "Mood must be between 1 and 5"
			}
			energy, err := strconv.Atoi(strings.TrimSpace(state.Inputs[1].Value()))
			if err != nil || energy < 1 || energy > 5 {
				return state, nil, "Energy must be between 1 and 5"
			}
			sleepHours, err := ParseOptionalDurationHours(state.Inputs[2].Value(), "Sleep")
			if err != nil {
				return state, nil, err.Error()
			}
			sleepScore, status := parseOptionalIntRange(
				strings.TrimSpace(state.Inputs[3].Value()),
				0,
				100,
				"Sleep score must be between 0 and 100",
			)
			if status != "" {
				return state, nil, status
			}
			screenTime, err := ParseOptionalDurationMinutes(state.Inputs[4].Value(), "Screen time")
			if err != nil {
				return state, nil, err.Error()
			}
			note := ValueToPointer(strings.TrimSpace(state.Inputs[5].Value()))
			return Close(state), &Action{
				Kind:              state.Kind,
				CheckInDate:       state.CheckInDate,
				Mood:              mood,
				Energy:            energy,
				SleepHours:        sleepHours,
				SleepScore:        sleepScore,
				ScreenTimeMinutes: screenTime,
				Note:              note,
			}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func parseOptionalIntRange(raw string, min int, max int, message string) (*int, string) {
	if raw == "" {
		return nil, ""
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return nil, message
	}
	return &value, ""
}

func updateAmendSession(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		return Close(state), nil, ""
	default:
		if action == dialogActionPrimary {
			note := strings.TrimSpace(state.Inputs[0].Value())
			if note == "" {
				return state, nil, "Commit message is required"
			}
			return Close(
					state,
				), &Action{
					Kind: "amend_session",
					ID:   state.SessionID,
					Note: ValueToPointer(note),
				}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateManualSession(state State, msg tea.KeyMsg) (State, *Action, string) {
	action := dialogActionForKey(state, msg.String())
	switch action {
	case dialogActionCancel:
		return Close(state), nil, ""
	}
	switch msg.String() {
	case "ctrl+e", "ctrl+y":
		if state.FocusIdx == 1 {
			return OpenDatePicker(
				state,
				"manual_session",
				state.IssueID,
				1,
				ValueToPointer(state.Inputs[1].Value()),
				strings.TrimSpace(state.Inputs[1].Value()),
			), nil, ""
		}
	case "g":
		if state.FocusIdx == 1 {
			state.Inputs[1].SetValue(time.Now().Format("2006-01-02"))
			return state, nil, ""
		}
	}
	switch action {
	case dialogActionFocusNext, dialogActionFocusPrev, dialogActionMoveDown, dialogActionMoveUp:
		dir := dialogFocusMoveDir(action)
		if dir == 0 {
			dir = dialogVerticalMoveDir(action)
		}
		state.FocusIdx = (state.FocusIdx + dir + len(state.Inputs)) % len(
			state.Inputs,
		)
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	default:
		if action == dialogActionPrimary {
			date := strings.TrimSpace(state.Inputs[1].Value())
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return state, nil, "Date must be YYYY-MM-DD"
			}
			workSeconds, err := ParseDurationInput(state.Inputs[2].Value(), true, "Work duration")
			if err != nil {
				return state, nil, err.Error()
			}
			breakSeconds, err := ParseDurationInput(
				state.Inputs[3].Value(),
				false,
				"Break duration",
			)
			if err != nil {
				return state, nil, err.Error()
			}
			startTime, err := ParseClockInput(state.Inputs[4].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			endTime, err := ParseClockInput(state.Inputs[5].Value())
			if err != nil {
				return state, nil, err.Error()
			}
			req := shareddto.ManualSessionLogRequest{
				IssueID:              state.IssueID,
				Date:                 date,
				WorkDurationSeconds:  workSeconds,
				BreakDurationSeconds: breakSeconds,
				StartTime:            startTime,
				EndTime:              endTime,
				CommitMessage:        ValueToPointer(state.Inputs[0].Value()),
				Notes:                ValueToPointer(state.Inputs[6].Value()),
			}
			return Close(
					state,
				), &Action{
					Kind:          "manual_session",
					IssueID:       state.IssueID,
					ManualSession: &req,
				}, ""
		}
	}
	var cmd tea.Cmd
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func updateHabitCompletion(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		return Close(state), nil, ""
	case dialogActionFocusNext, dialogActionFocusPrev, dialogActionMoveDown, dialogActionMoveUp:
		dir := dialogFocusMoveDir(action)
		if dir == 0 {
			dir = dialogVerticalMoveDir(action)
		}
		state.FocusIdx = (state.FocusIdx + dir + 2) % 2
		state = SyncDialogFocus(state)
		return clearDialogError(state), nil, ""
	}
	if dialogActionForKey(state, msg.String()) == dialogActionPrimary {
		duration, err := ParseOptionalDurationMinutes(state.Inputs[0].Value(), "Duration")
		if err != nil {
			return state, nil, err.Error()
		}
		return Close(state), &Action{
			Kind:        state.Kind,
			HabitID:     state.HabitID,
			CheckInDate: state.CheckInDate,
			Estimate:    duration,
			Note:        ValueToPointer(strings.TrimSpace(state.Description.Value())),
		}, ""
	}
	if state.DescriptionEnabled && state.FocusIdx == state.DescriptionIndex {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	state.Inputs[0], cmd = state.Inputs[0].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}
