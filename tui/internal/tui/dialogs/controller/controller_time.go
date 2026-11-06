package controller

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	shareddatefmt "crona.work/shared/datefmt"
	shareddto "crona.work/shared/dto"
	sharedtypes "crona.work/shared/types"

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
		row.StartTime.SetValue(nextSessionHistorySegmentStart(state))
		state.SessionHistorySegments = append(state.SessionHistorySegments, row)
		state.FocusIdx = sessionHistorySegmentFocus(len(state.SessionHistorySegments)-1, 0)
		state = syncSessionHistoryBounds(state)
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
		state = syncSessionHistoryBounds(state)
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
		return syncSessionHistoryBounds(clearDialogError(state)), nil, ""
	}
	var cmd tea.Cmd
	if state.FocusIdx < 0 || state.FocusIdx >= len(state.Inputs) {
		return state, nil, ""
	}
	state.Inputs[state.FocusIdx], cmd = state.Inputs[state.FocusIdx].Update(msg)
	_ = cmd
	return clearDialogError(state), nil, ""
}

// syncSessionHistoryBounds keeps the session's start and end on the first segment's start
// and the last segment's end. A time still being typed is skipped, so the session keeps its
// last valid bound until the segment's time parses again.
func syncSessionHistoryBounds(state State) State {
	segments := state.SessionHistorySegments
	if len(segments) == 0 || len(state.Inputs) < 2 {
		return state
	}
	start, err := resolveSessionHistoryTime(state.Inputs[0].Value(), state.SessionHistoryBounds[0], time.Time{}, state.SessionHistoryDatePattern)
	if err != nil {
		return state
	}
	if first, err := resolveSessionHistoryTime(segments[0].StartTime.Value(), segments[0].OriginalStart, start, state.SessionHistoryDatePattern); err == nil {
		start = first
		state.Inputs[0].SetValue(sessionHistoryStamp(start, state.SessionHistoryDatePattern))
	}
	if last, err := resolveSessionHistoryTime(segments[len(segments)-1].EndTime.Value(), segments[len(segments)-1].OriginalEnd, start, state.SessionHistoryDatePattern); err == nil {
		state.Inputs[1].SetValue(sessionHistoryTimeText(last, start, state.SessionHistoryDatePattern))
	}
	return state
}

// nextSessionHistorySegmentStart starts a new segment where the last one ends, or at the
// session's start when there is none.
func nextSessionHistorySegmentStart(state State) string {
	if n := len(state.SessionHistorySegments); n > 0 {
		return state.SessionHistorySegments[n-1].EndTime.Value()
	}
	if len(state.Inputs) == 0 {
		return ""
	}
	start, err := resolveSessionHistoryTime(state.Inputs[0].Value(), state.SessionHistoryBounds[0], time.Time{}, state.SessionHistoryDatePattern)
	if err != nil {
		return ""
	}
	return sessionHistoryClock(start)
}

func sessionHistoryUpdateRequest(state State) (*shareddto.UpdateSessionHistoryRequest, error) {
	if len(state.Inputs) < 2 || len(state.SessionHistoryNoteInputs) < 3 {
		return nil, errors.New("session editor is incomplete")
	}
	start, err := resolveSessionHistoryTime(state.Inputs[0].Value(), state.SessionHistoryBounds[0], time.Time{}, state.SessionHistoryDatePattern)
	if err != nil {
		return nil, fmt.Errorf(
			"session start needs a date and time, like %s",
			sessionHistoryStamp(time.Date(2026, 11, 4, 9, 30, 0, 0, time.Local), state.SessionHistoryDatePattern),
		)
	}
	end, err := resolveSessionHistoryTime(state.Inputs[1].Value(), state.SessionHistoryBounds[1], start, state.SessionHistoryDatePattern)
	if err != nil || !end.After(start) {
		return nil, errors.New("session end must be after its start")
	}
	segments := make([]sharedtypes.SessionSegment, 0, len(state.SessionHistorySegments))
	for i, row := range state.SessionHistorySegments {
		segmentStart, err := resolveSessionHistoryTime(row.StartTime.Value(), row.OriginalStart, start, state.SessionHistoryDatePattern)
		if err != nil {
			return nil, fmt.Errorf("segment %d start must be a time, like 09:30", i+1)
		}
		segmentEnd, err := resolveSessionHistoryTime(row.EndTime.Value(), row.OriginalEnd, start, state.SessionHistoryDatePattern)
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

// parseSessionHistoryTime reads a time typed in the session editor, in local time. A date
// in the user's pattern (or ISO) followed by `HH:mm[:ss]` stands alone. A bare `HH:mm[:ss]`
// lands on whichever of the day before, the day of, or the day after base is nearest to
// base, so a session can run past midnight in either direction. The zero base accepts
// only a full date and time.
func parseSessionHistoryTime(value string, base time.Time, pattern string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -07:00", value, time.Local); err == nil {
		return parsed, nil
	}
	if cut := strings.LastIndexByte(value, ' '); cut > 0 {
		if clock, ok := parseSessionHistoryClock(value[cut+1:]); ok {
			if day, err := shareddatefmt.ParseDate(value[:cut], pattern, time.Local); err == nil {
				return atClock(day, clock), nil
			}
		}
	}
	clock, ok := parseSessionHistoryClock(value)
	if !ok {
		return time.Time{}, fmt.Errorf("unrecognised time %q", value)
	}
	if base.IsZero() {
		return time.Time{}, errors.New("a date is required")
	}
	best := atClock(base.In(time.Local), clock)
	for _, offset := range []int{-1, 1} {
		candidate := atClock(base.In(time.Local).AddDate(0, 0, offset), clock)
		if candidate.Sub(base).Abs() < best.Sub(base).Abs() {
			best = candidate
		}
	}
	return best, nil
}

func parseSessionHistoryClock(value string) (time.Time, bool) {
	for _, layout := range []string{"15:04:05", "15:04"} {
		if clock, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return clock, true
		}
	}
	return time.Time{}, false
}

// atClock is day's date at clock's time of day, in local time.
func atClock(day, clock time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), clock.Hour(), clock.Minute(), clock.Second(), 0, time.Local)
}

// sessionHistoryDatePlaceholder writes the user's date pattern for a placeholder.
func sessionHistoryDatePlaceholder(pattern string) string {
	if strings.TrimSpace(pattern) == "" {
		return "YYYY-MM-DD"
	}
	return pattern
}

// resolveSessionHistoryTime is the loaded time while the field still shows it, and the
// parsed text once it has been edited.
func resolveSessionHistoryTime(
	value string,
	original SessionHistoryTime,
	base time.Time,
	pattern string,
) (time.Time, error) {
	if !original.At.IsZero() && strings.TrimSpace(value) == original.Text {
		return original.At, nil
	}
	return parseSessionHistoryTime(value, base, pattern)
}

// sessionHistoryClock is `HH:mm`, with seconds only when they are not zero.
func sessionHistoryClock(t time.Time) string {
	t = t.In(time.Local)
	if t.Second() != 0 {
		return t.Format("15:04:05")
	}
	return t.Format("15:04")
}

func sessionHistoryStamp(t time.Time, pattern string) string {
	return shareddatefmt.FormatWithPattern(t.In(time.Local), pattern) + " " + sessionHistoryClock(t)
}

// sessionHistoryTimeText is the clock alone on base's day, and the full stamp otherwise.
func sessionHistoryTimeText(t, base time.Time, pattern string) string {
	if t.In(time.Local).Format("2006-01-02") == base.In(time.Local).Format("2006-01-02") {
		return sessionHistoryClock(t)
	}
	return sessionHistoryStamp(t, pattern)
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
