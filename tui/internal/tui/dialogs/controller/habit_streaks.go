package controller

import (
	"strconv"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type habitStreakDetailRow int

const (
	habitStreakDetailRowName habitStreakDetailRow = iota
	habitStreakDetailRowDescription
	habitStreakDetailRowPeriod
	habitStreakDetailRowCount
)

func isMomentumDialogKind(kind string) bool {
	switch kind {
	case "create_momentum", "edit_momentum":
		return true
	default:
		return false
	}
}

func updateHabitStreaks(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch state.HabitStreakStep {
	case 1:
		return updateHabitStreakDetails(state, msg)
	case 2:
		return updateHabitStreakHabits(state, msg)
	case 3:
		return updateHabitStreakReview(state, msg)
	default:
		return state, nil, ""
	}
}

func openHabitStreakEditor(state State, idx int, def sharedtypes.HabitStreakDefinition) State {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	name := textinput.New()
	name.Placeholder = "Health streak"
	name.SetValue(def.Name)
	name.CharLimit = 80
	name.Width = 36
	name.Focus()
	count := textinput.New()
	count.Placeholder = "1"
	count.SetValue(strconv.Itoa(max(1, def.RequiredCount)))
	count.CharLimit = 3
	count.Width = 8
	state.Inputs = []textinput.Model{name, count}
	state = habitStreakSetDetailFocus(state, habitStreakDetailRowName)
	state.HabitStreakEditIdx = idx
	state.HabitStreakDraft = def
	state.HabitStreakStep = 1
	state.HabitStreakCursor = 0
	state.ErrorMessage = ""
	return state
}

func updateHabitStreakDetails(state State, msg tea.KeyMsg) (State, *Action, string) {
	momentumMode := isMomentumDialogKind(state.Kind)
	row := habitStreakDetailRowForFocus(momentumMode, state.FocusIdx)
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "tab":
		if row == habitStreakDetailRowCount {
			return moveHabitStreakToHabitSelection(state)
		}
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, 1))
		return state, nil, ""
	case "shift+tab":
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, -1))
		return state, nil, ""
	case "left", "h":
		if row == habitStreakDetailRowPeriod {
			state.HabitStreakDraft.Period = nextHabitStreakPeriod(state.HabitStreakDraft.Period, -1)
			state = habitStreakApplyPeriodRule(state)
			return state, nil, ""
		}
	case "right", "l":
		if row == habitStreakDetailRowPeriod {
			state.HabitStreakDraft.Period = nextHabitStreakPeriod(state.HabitStreakDraft.Period, 1)
			state = habitStreakApplyPeriodRule(state)
			return state, nil, ""
		}
	case "enter":
		switch row {
		case habitStreakDetailRowName:
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, 1))
			return state, nil, ""
		case habitStreakDetailRowDescription:
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, 1))
			return state, nil, ""
		case habitStreakDetailRowPeriod:
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, 1))
			return state, nil, ""
		default:
			return moveHabitStreakToHabitSelection(state)
		}
	}
	if momentumMode && row == habitStreakDetailRowDescription {
		var cmd tea.Cmd
		state.Description, cmd = state.Description.Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
	}
	if inputIdx, ok := habitStreakInputIndex(row); ok && inputIdx >= 0 &&
		inputIdx < len(state.Inputs) {
		var cmd tea.Cmd
		state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
		_ = cmd
	}
	state = habitStreakApplyPeriodRule(state)
	return clearDialogError(state), nil, ""
}

func moveHabitStreakToHabitSelection(state State) (State, *Action, string) {
	momentumMode := isMomentumDialogKind(state.Kind)
	name := strings.TrimSpace(state.Inputs[0].Value())
	if name == "" {
		state.ErrorMessage = "Streak name is required"
		return state, nil, ""
	}
	required, err := strconv.Atoi(strings.TrimSpace(state.Inputs[1].Value()))
	if err != nil || required <= 0 {
		state.ErrorMessage = "Required count must be a positive integer"
		return state, nil, ""
	}
	state.HabitStreakDraft.Name = name
	if momentumMode {
		state.HabitStreakDraft.Description = ValueToPointer(strings.TrimSpace(state.Description.Value()))
	}
	state.HabitStreakDraft.RequiredCount = required
	state.HabitStreakDraft.Period = sharedtypes.NormalizeHabitStreakPeriod(
		state.HabitStreakDraft.Period,
	)
	state = habitStreakApplyPeriodRule(state)
	state.HabitStreakStep = 2
	state.HabitStreakCursor = 0
	state.ErrorMessage = ""
	return state, nil, ""
}

func updateHabitStreakHabits(state State, msg tea.KeyMsg) (State, *Action, string) {
	total := len(state.HabitItems)
	if total == 0 {
		switch msg.String() {
		case "esc":
			state.HabitStreakStep = 1
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowCount)
			return state, nil, ""
		case "enter", "tab":
			state.HabitStreakStep = 3
			return state, nil, ""
		}
		return state, nil, ""
	}
	switch msg.String() {
	case "esc":
		state.HabitStreakStep = 1
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowCount)
		return state, nil, ""
	case "j", "down":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, 1)
	case "k", "up":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, -1)
	case " ", "x":
		habitID := state.HabitItems[state.HabitStreakCursor].ID
		state.HabitStreakDraft.HabitIDs = toggleHabitMembership(
			state.HabitStreakDraft.HabitIDs,
			habitID,
		)
	case "a":
		ids := make([]int64, 0, len(state.HabitItems))
		for _, item := range state.HabitItems {
			ids = append(ids, item.ID)
		}
		state.HabitStreakDraft.HabitIDs = ids
	case "c":
		state.HabitStreakDraft.HabitIDs = nil
	case "tab", "enter":
		state.HabitStreakStep = 3
		state.HabitStreakCursor = 0
	}
	return clearDialogError(state), nil, ""
}

func updateHabitStreakReview(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		state.HabitStreakStep = 2
		state.HabitStreakCursor = 0
		return state, nil, ""
	case "shift+tab", "left", "h":
		state.HabitStreakStep = 2
		return state, nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			draft := sharedtypes.NormalizeHabitStreakDefinition(state.HabitStreakDraft)
			if state.Kind == "create_momentum" || state.Kind == "edit_momentum" {
				kind := "update_momentum"
				if state.Kind == "create_momentum" {
					kind = "create_momentum"
				}
				return Close(state), &Action{
					Kind:            kind,
					HabitStreakDefs: []sharedtypes.HabitStreakDefinition{draft},
				}, ""
			}
			return Close(state), nil, ""
		}
	}
	return clearDialogError(state), nil, ""
}

func nextHabitStreakPeriod(
	current sharedtypes.HabitStreakPeriod,
	dir int,
) sharedtypes.HabitStreakPeriod {
	options := []sharedtypes.HabitStreakPeriod{
		sharedtypes.HabitStreakPeriodDay,
		sharedtypes.HabitStreakPeriodWeek,
		sharedtypes.HabitStreakPeriodMonth,
	}
	current = sharedtypes.NormalizeHabitStreakPeriod(current)
	return options[nextIndex(current, options, dir)]
}

func nextIndex[T comparable](current T, options []T, dir int) int {
	currentIdx := 0
	for i, option := range options {
		if option == current {
			currentIdx = i
			break
		}
	}
	next := currentIdx + dir
	if next < 0 {
		next = len(options) - 1
	}
	if next >= len(options) {
		next = 0
	}
	return next
}

func habitStreakDetailRowNext(
	momentumMode bool,
	row habitStreakDetailRow,
	dir int,
) habitStreakDetailRow {
	rows := habitStreakDetailRows(momentumMode)
	return rows[nextIndex(row, rows, dir)]
}

func habitStreakSetDetailFocus(state State, row habitStreakDetailRow) State {
	momentumMode := isMomentumDialogKind(state.Kind)
	state.FocusIdx = habitStreakFocusForRow(momentumMode, row)
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	state.Description.Blur()
	if momentumMode && row == habitStreakDetailRowDescription && state.DescriptionEnabled {
		state.Description.Focus()
		return state
	}
	if inputIdx, ok := habitStreakInputIndex(row); ok && inputIdx >= 0 &&
		inputIdx < len(state.Inputs) {
		state.Inputs[inputIdx].Focus()
	}
	return state
}

func habitStreakApplyPeriodRule(state State) State {
	if sharedtypes.NormalizeHabitStreakPeriod(
		state.HabitStreakDraft.Period,
	) == sharedtypes.HabitStreakPeriodDay {
		state.HabitStreakDraft.RequiredCount = 1
		if len(state.Inputs) > 1 {
			state.Inputs[1].SetValue("1")
		}
	}
	return state
}

func habitStreakInputIndex(row habitStreakDetailRow) (int, bool) {
	switch row {
	case habitStreakDetailRowName:
		return 0, true
	case habitStreakDetailRowCount:
		return 1, true
	default:
		return 0, false
	}
}

func habitStreakDetailRows(momentumMode bool) []habitStreakDetailRow {
	if momentumMode {
		return []habitStreakDetailRow{
			habitStreakDetailRowName,
			habitStreakDetailRowDescription,
			habitStreakDetailRowPeriod,
			habitStreakDetailRowCount,
		}
	}
	return []habitStreakDetailRow{
		habitStreakDetailRowName,
		habitStreakDetailRowPeriod,
		habitStreakDetailRowCount,
	}
}

func habitStreakDetailRowForFocus(momentumMode bool, focusIdx int) habitStreakDetailRow {
	rows := habitStreakDetailRows(momentumMode)
	if focusIdx < 0 {
		focusIdx = 0
	}
	if focusIdx >= len(rows) {
		focusIdx = len(rows) - 1
	}
	return rows[focusIdx]
}

func habitStreakFocusForRow(momentumMode bool, row habitStreakDetailRow) int {
	rows := habitStreakDetailRows(momentumMode)
	for i, candidate := range rows {
		if candidate == row {
			return i
		}
	}
	return 0
}

func toggleHabitMembership(values []int64, habitID int64) []int64 {
	for i, value := range values {
		if value == habitID {
			return append(values[:i], values[i+1:]...)
		}
	}
	return append(values, habitID)
}
