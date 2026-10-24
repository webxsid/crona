package controller

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"

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

func updateEditSessionHistory(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab":
		return moveSessionHistoryFocus(state, 1), nil, ""
	case "shift+tab":
		return moveSessionHistoryFocus(state, -1), nil, ""
	case "ctrl+a":
		row := newSessionHistorySegmentRow(state, sharedtypes.SessionSegmentWork)
		if len(state.Inputs) >= 2 {
			row.StartTime.SetValue(sessionHistoryDatePrefix(state.Inputs[0].Value()))
			row.EndTime.SetValue(sessionHistoryDatePrefix(state.Inputs[1].Value()))
		}
		state.SessionHistorySegments = append(state.SessionHistorySegments, row)
		state.FocusIdx = sessionHistorySegmentFocus(len(state.SessionHistorySegments)-1, 0)
		return focusSessionHistoryControl(clearDialogError(state)), nil, ""
	case "ctrl+d":
		rowIndex, column, ok := sessionHistorySegmentAt(state.FocusIdx)
		if !ok || rowIndex >= len(state.SessionHistorySegments) {
			return state, nil, ""
		}
		state.SessionHistorySegments = slices.Delete(
			state.SessionHistorySegments,
			rowIndex,
			rowIndex+1,
		)
		if column > 0 {
			column--
		}
		state.FocusIdx = min(state.FocusIdx, historyControlCount(state)-1)
		if len(state.SessionHistorySegments) > 0 && state.FocusIdx >= 5 {
			state.FocusIdx = sessionHistorySegmentFocus(
				min(rowIndex, len(state.SessionHistorySegments)-1),
				column,
			)
		}
		return focusSessionHistoryControl(clearDialogError(state)), nil, ""
	}
	if msg.String() == "left" || msg.String() == "right" {
		if rowIndex, column, ok := sessionHistorySegmentAt(state.FocusIdx); ok && column == 0 {
			dir := 1
			if msg.String() == "left" {
				dir = -1
			}
			state.SessionHistorySegments[rowIndex].Type = nextSessionHistorySegmentType(
				state.SessionHistorySegments[rowIndex].Type,
				dir,
			)
			return clearDialogError(state), nil, ""
		}
	}
	if isDialogSubmitKey(state, msg.String()) {
		request, err := sessionHistoryUpdateRequest(state)
		if err != nil {
			return state, nil, err.Error()
		}
		return Close(state), &Action{Kind: state.Kind, SessionHistoryUpdate: request}, ""
	}
	if state.FocusIdx == 2 || state.FocusIdx == 3 || state.FocusIdx == 4 {
		index := state.FocusIdx - 2
		if index >= len(state.SessionHistoryNoteInputs) {
			return state, nil, ""
		}
		var cmd tea.Cmd
		state.SessionHistoryNoteInputs[index], cmd = state.SessionHistoryNoteInputs[index].Update(
			msg,
		)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	if state.FocusIdx >= 5 {
		rowIndex, column, ok := sessionHistorySegmentAt(state.FocusIdx)
		if !ok || rowIndex >= len(state.SessionHistorySegments) {
			return state, nil, ""
		}
		row := &state.SessionHistorySegments[rowIndex]
		switch column {
		case 1:
			var cmd tea.Cmd
			row.StartTime, cmd = row.StartTime.Update(msg)
			_ = cmd
		case 2:
			var cmd tea.Cmd
			row.EndTime, cmd = row.EndTime.Update(msg)
			_ = cmd
		}
		return clearDialogError(state), nil, ""
	}
	var cmd tea.Cmd
	if state.FocusIdx < 0 || state.FocusIdx >= len(state.Inputs) {
		return state, nil, ""
	}
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

func sessionHistoryUpdateRequest(state State) (*shareddto.UpdateSessionHistoryRequest, error) {
	if len(state.Inputs) < 2 || len(state.SessionHistoryNoteInputs) < 3 {
		return nil, errors.New("session editor is incomplete")
	}
	start, err := parseSessionHistoryTimestamp(state.Inputs[0].Value())
	if err != nil {
		return nil, errors.New("session start must use YYYY-MM-DD HH:mm:ss")
	}
	end, err := parseSessionHistoryTimestamp(state.Inputs[1].Value())
	if err != nil || !end.After(start) {
		return nil, errors.New("session end must be after its start")
	}
	segments := make([]sharedtypes.SessionSegment, 0, len(state.SessionHistorySegments))
	for i, row := range state.SessionHistorySegments {
		segmentStart, err := parseSessionHistoryTimestamp(row.StartTime.Value())
		if err != nil {
			return nil, fmt.Errorf("segment %d start must use YYYY-MM-DD HH:mm:ss", i+1)
		}
		segmentEnd, err := parseSessionHistoryTimestamp(row.EndTime.Value())
		if err != nil || !segmentEnd.After(segmentStart) {
			return nil, fmt.Errorf("segment %d end must be after its start", i+1)
		}
		if segmentStart.Before(start) || segmentEnd.After(end) {
			return nil, fmt.Errorf("segment %d must fall within the session", i+1)
		}
		endValue := segmentEnd.UTC().Format(time.RFC3339Nano)
		segments = append(segments, sharedtypes.SessionSegment{
			SegmentType: row.Type,
			StartTime:   segmentStart.UTC().Format(time.RFC3339Nano),
			EndTime:     &endValue,
		})
	}
	slices.SortFunc(segments, func(a, b sharedtypes.SessionSegment) int {
		startA, _ := time.Parse(time.RFC3339Nano, a.StartTime)
		startB, _ := time.Parse(time.RFC3339Nano, b.StartTime)
		return startA.Compare(startB)
	})
	for i := 1; i < len(segments); i++ {
		previousEnd, _ := time.Parse(time.RFC3339, *segments[i-1].EndTime)
		currentStart, _ := time.Parse(time.RFC3339, segments[i].StartTime)
		if currentStart.Before(previousEnd) {
			return nil, errors.New("session segments cannot overlap")
		}
	}
	notes := sharedtypes.ParsedSessionNotes{
		sharedtypes.SessionNoteSectionCommit: strings.TrimSpace(
			state.SessionHistoryNoteInputs[0].Value(),
		),
		sharedtypes.SessionNoteSectionWork: strings.TrimSpace(
			state.SessionHistoryNoteInputs[1].Value(),
		),
		sharedtypes.SessionNoteSectionNotes: strings.TrimSpace(
			state.SessionHistoryNoteInputs[2].Value(),
		),
	}
	return &shareddto.UpdateSessionHistoryRequest{
		ID: state.SessionID, StartTime: start.UTC().Format(time.RFC3339Nano), EndTime: end.UTC().Format(time.RFC3339Nano),
		EditableNotes: notes, Segments: segments,
	}, nil
}

func parseSessionHistoryTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse("2006-01-02 15:04:05.999999999 -07:00", value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05.999999999", value, time.Local); err == nil {
		return parsed, nil
	}
	return time.ParseInLocation("2006-01-02 15:04", value, time.Local)
}

func sessionHistoryDatePrefix(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < len("2006-01-02") {
		return ""
	}
	if _, err := time.Parse("2006-01-02", value[:len("2006-01-02")]); err != nil {
		return ""
	}
	return value[:len("2006-01-02")] + " "
}

func historyControlCount(state State) int {
	return 5 + len(state.SessionHistorySegments)*3
}

func sessionHistorySegmentFocus(rowIndex, column int) int {
	return 5 + rowIndex*3 + column
}

func sessionHistorySegmentAt(focus int) (int, int, bool) {
	if focus < 5 {
		return 0, 0, false
	}
	index := focus - 5
	return index / 3, index % 3, true
}

func moveSessionHistoryFocus(state State, dir int) State {
	count := historyControlCount(state)
	state.FocusIdx = (state.FocusIdx + dir + count) % count
	return focusSessionHistoryControl(clearDialogError(state))
}

func focusSessionHistoryControl(state State) State {
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	for i := range state.SessionHistoryNoteInputs {
		state.SessionHistoryNoteInputs[i].Blur()
	}
	for i := range state.SessionHistorySegments {
		state.SessionHistorySegments[i].StartTime.Blur()
		state.SessionHistorySegments[i].EndTime.Blur()
	}
	switch state.FocusIdx {
	case 0, 1:
		state.Inputs[state.FocusIdx].Focus()
	case 2, 3, 4:
		state.SessionHistoryNoteInputs[state.FocusIdx-2].Focus()
	default:
		rowIndex, column, ok := sessionHistorySegmentAt(state.FocusIdx)
		if ok && rowIndex < len(state.SessionHistorySegments) {
			switch column {
			case 1:
				state.SessionHistorySegments[rowIndex].StartTime.Focus()
			case 2:
				state.SessionHistorySegments[rowIndex].EndTime.Focus()
			}
		}
	}
	return state
}

func nextSessionHistorySegmentType(
	current sharedtypes.SessionSegmentType,
	dir int,
) sharedtypes.SessionSegmentType {
	types := []sharedtypes.SessionSegmentType{
		sharedtypes.SessionSegmentWork,
		sharedtypes.SessionSegmentShortBreak,
		sharedtypes.SessionSegmentLongBreak,
		sharedtypes.SessionSegmentRest,
	}
	index := slices.Index(types, current)
	if index < 0 {
		index = 0
	}
	return types[(index+dir+len(types))%len(types)]
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
