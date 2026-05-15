package controller

import (
	"strings"
	"time"

	sharedtypes "crona/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func updateDatePicker(state State, ctx UpdateContext, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Parent == "rollup_start" || state.Parent == "rollup_end" {
			return Close(state), nil, ""
		}
		return closeDatePicker(state), nil, ""
	case "enter", " ":
		selected := state.DateCursorValue
		if state.Parent == "create_issue_meta" || state.Parent == "create_issue_default" || state.Parent == "edit_issue" {
			if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
				state.Inputs[idx].SetValue(selected)
			}
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "edit_rest_protection" {
			state.ProtectionDates = normalizedDateList(append(state.ProtectionDates, selected))
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "rollup_start" {
			return Close(state), &Action{Kind: "set_rollup_start_date", DueDate: ValueToPointer(selected)}, ""
		}
		if state.Parent == "rollup_end" {
			return Close(state), &Action{Kind: "set_rollup_end_date", DueDate: ValueToPointer(selected)}, ""
		}
		return Close(state), &Action{Kind: "set_issue_todo_date", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, DueDate: ValueToPointer(selected)}, ""
	case "backspace", "delete", "c":
		if state.Parent == "create_issue_meta" || state.Parent == "create_issue_default" || state.Parent == "edit_issue" {
			if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
				state.Inputs[idx].SetValue("")
			}
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "edit_rest_protection" {
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "rollup_start" || state.Parent == "rollup_end" {
			return Close(state), nil, ""
		}
		return Close(state), &Action{Kind: "set_issue_todo_date", IssueID: state.IssueID, StreamID: ctx.ActiveIssueStream, DueDate: ValueToPointer("")}, ""
	case "left", "h":
		return shiftDatePicker(state, 0, 0, -1), nil, ""
	case "right", "l":
		return shiftDatePicker(state, 0, 0, 1), nil, ""
	case "up", "k":
		return shiftDatePicker(state, 0, 0, -7), nil, ""
	case "down", "j":
		return shiftDatePicker(state, 0, 0, 7), nil, ""
	case ",":
		return shiftDatePicker(state, 0, -1, 0), nil, ""
	case ".":
		return shiftDatePicker(state, 0, 1, 0), nil, ""
	case "g":
		return OpenDatePicker(state, state.Parent, state.IssueID, state.FocusIdx, ValueToPointer(time.Now().Format("2006-01-02")), currentDate), nil, ""
	}
	return state, nil, ""
}

func ternaryDir(key string) int {
	if key == "shift+tab" || key == "up" {
		return -1
	}
	return 1
}

func closeDatePicker(state State) State {
	parent := state.Parent
	dateField := state.FocusIdx
	state.Kind = parent
	state.Parent = ""
	state.DateMonthValue = ""
	state.DateCursorValue = ""
	if parent == "create_issue_meta" || parent == "create_issue_default" || parent == "edit_issue" {
		state.FocusIdx = dateField
		state = SyncDialogFocus(state)
		return state
	}
	for i := range state.Inputs {
		if i == state.FocusIdx {
			state.Inputs[i].Focus()
		} else {
			state.Inputs[i].Blur()
		}
	}
	return state
}

func shiftDatePicker(state State, years, months, days int) State {
	selected := DialogDate(state, time.Now().Format("2006-01-02")).AddDate(years, months, days)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	state.DateCursorValue = selected.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	return state
}

func updateRestProtection(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab", "right", "l":
		state.ProtectionStep = (state.ProtectionStep + 1) % 4
		state.ProtectionCursor = 0
		return state, nil, ""
	case "shift+tab", "left", "h":
		state.ProtectionStep = (state.ProtectionStep + 3) % 4
		state.ProtectionCursor = 0
		return state, nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			return Close(state), &Action{
				Kind:        "patch_rest_protection",
				StreakKinds: streakKindStrings(state.ProtectionStreaks),
				IntList:     normalizedWeekdays(state.ProtectionWeekdays),
				RestDates:   normalizedDateList(state.ProtectionDates),
			}, ""
		}
	}
	switch state.ProtectionStep {
	case 0:
		return updateRestProtectionStreaks(state, msg)
	case 1:
		return updateRestProtectionWeekdays(state, msg)
	case 2:
		return updateRestProtectionDates(state, currentDate, msg)
	case 3:
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateRestProtectionStreaks(state State, msg tea.KeyMsg) (State, *Action, string) {
	total := len(sharedtypes.AvailableStreakKinds())
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, total, 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, total, -1)
	case " ", "x":
		current := sharedtypes.AvailableStreakKinds()[state.ProtectionCursor]
		state.ProtectionStreaks = toggleStreakKind(state.ProtectionStreaks, current)
	case "a":
		state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), sharedtypes.AvailableStreakKinds()...)
	case "c":
		state.ProtectionStreaks = nil
	case "enter":
		state.ProtectionStep = 1
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}

func updateRestProtectionWeekdays(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, 7, 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, 7, -1)
	case " ", "x":
		state.ProtectionWeekdays = toggleWeekday(state.ProtectionWeekdays, state.ProtectionCursor)
	case "c":
		state.ProtectionWeekdays = nil
	case "enter":
		state.ProtectionStep = 2
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}

func updateRestProtectionDates(state State, currentDate string, msg tea.KeyMsg) (State, *Action, string) {
	totalDates := len(state.ProtectionDates)
	if totalDates < 1 {
		totalDates = 1
	}
	switch msg.String() {
	case "j", "down":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, totalDates, 1)
	case "k", "up":
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, totalDates, -1)
	case "a":
		initial := currentDate
		if strings.TrimSpace(initial) == "" {
			initial = time.Now().Format("2006-01-02")
		}
		return OpenDatePicker(state, "edit_rest_protection", 0, 0, ValueToPointer(initial), currentDate), nil, ""
	case "d", "backspace", "delete":
		if len(state.ProtectionDates) == 0 {
			return state, nil, ""
		}
		idx := state.ProtectionCursor
		if idx >= len(state.ProtectionDates) {
			idx = len(state.ProtectionDates) - 1
		}
		state.ProtectionDates = append(state.ProtectionDates[:idx], state.ProtectionDates[idx+1:]...)
		if state.ProtectionCursor >= len(state.ProtectionDates) && state.ProtectionCursor > 0 {
			state.ProtectionCursor--
		}
	case "enter":
		state.ProtectionStep = 3
		state.ProtectionCursor = 0
	}
	return state, nil, ""
}
