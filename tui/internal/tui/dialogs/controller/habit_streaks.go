package controller

import (
	"fmt"
	"strconv"
	"strings"

	sharedtypes "crona/shared/types"
	sharedutils "crona/shared/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type habitStreakDetailRow int
type habitStreakTargetFocus int

const (
	habitStreakDetailRowName habitStreakDetailRow = iota
	habitStreakDetailRowDescription
	habitStreakDetailRowPeriod
	habitStreakDetailRowCount
)

const (
	habitTargetFocusList habitStreakTargetFocus = iota
	habitTargetFocusMatchMode
)

const (
	contextTargetFocusRepo habitStreakTargetFocus = iota
	contextTargetFocusStream
	contextTargetFocusSelected
	contextTargetFocusMatchMode
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
	state = syncMomentumStepFocus(state)
	if state.Kind == "create_momentum" {
		switch state.HabitStreakStep {
		case 1:
			return updateHabitStreakKindSelection(state, msg)
		case 2:
			return updateHabitStreakTargets(state, msg)
		case 3:
			return updateHabitStreakDetails(state, msg)
		case 4:
			return updateHabitStreakReview(state, msg)
		default:
			return state, nil, ""
		}
	}
	switch state.HabitStreakStep {
	case 1:
		return updateHabitStreakTargets(state, msg)
	case 2:
		return updateHabitStreakDetails(state, msg)
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
	count.CharLimit = 24
	count.Width = 24
	state.Inputs = []textinput.Model{name, count}
	state = habitStreakSetDetailFocus(state, habitStreakDetailRowName)
	state.HabitStreakEditIdx = idx
	state.HabitStreakDraft = def
	state = syncHabitStreakThresholdInput(state)
	state.HabitStreakStep = 1
	state.HabitStreakCursor = 0
	state.ErrorMessage = ""
	return state
}

func updateHabitStreakKindSelection(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		return Close(state), nil, ""
	case dialogActionMoveDown:
		state.ChoiceCursor = ShiftSelection(state.ChoiceCursor, len(state.ChoiceItems), 1)
	case dialogActionMoveUp:
		state.ChoiceCursor = ShiftSelection(state.ChoiceCursor, len(state.ChoiceItems), -1)
	case dialogActionFocusNext, dialogActionActivate, dialogActionToggle, dialogActionPrimary:
		if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
			return state, nil, ""
		}
		targetKind := sharedtypes.MomentumTargetKind(state.ChoiceValues[state.ChoiceCursor])
		state.HabitStreakDraft.TargetKind = targetKind
		state.HabitStreakDraft.HabitIDs = nil
		state.HabitStreakDraft.Contexts = nil
		state.HabitStreakDraft.RequiredCount = habitStreakDefaultRequiredCount(targetKind)
		state = syncHabitStreakThresholdInput(state)
		state.HabitStreakStep = 2
		state.HabitStreakCursor = 0
		state.ErrorMessage = ""
		return state, nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateHabitStreakDetails(state State, msg tea.KeyMsg) (State, *Action, string) {
	momentumMode := isMomentumDialogKind(state.Kind)
	row := habitStreakDetailRowForFocus(momentumMode, state.FocusIdx)
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		if state.Kind == "create_momentum" {
			state.HabitStreakStep = 2
			state.HabitStreakCursor = 0
			return state, nil, ""
		}
		return Close(state), nil, ""
	case dialogActionFocusNext:
		if row == habitStreakDetailRowCount {
			return commitHabitStreakDetails(state)
		}
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, 1))
		return state, nil, ""
	case dialogActionFocusPrev:
		if row == habitStreakDetailRowName {
			state.HabitStreakStep = detailsStepForState(state) - 1
			state.FocusIdx = momentumTargetMatchModeFocusIndex(state)
			return syncMomentumStepFocus(state), nil, ""
		}
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(momentumMode, row, -1))
		return state, nil, ""
	case dialogActionMoveLeft:
		if row == habitStreakDetailRowPeriod {
			state.HabitStreakDraft.Period = nextHabitStreakPeriod(state.HabitStreakDraft.Period, -1)
			return state, nil, ""
		}
	case dialogActionMoveRight:
		if row == habitStreakDetailRowPeriod {
			state.HabitStreakDraft.Period = nextHabitStreakPeriod(state.HabitStreakDraft.Period, 1)
			return state, nil, ""
		}
	case dialogActionActivate:
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
			return commitHabitStreakDetails(state)
		}
	case dialogActionPrimary:
		return commitHabitStreakDetails(state)
	}
	if sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) == sharedtypes.MomentumTargetKindContext &&
		row == habitStreakDetailRowCount && isMomentumThresholdTextInputKey(msg) {
		var cmd tea.Cmd
		state.Inputs[1], cmd = state.Inputs[1].Update(msg)
		_ = cmd
		return clearDialogError(state), nil, ""
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
	return clearDialogError(state), nil, ""
}

func commitHabitStreakDetails(state State) (State, *Action, string) {
	momentumMode := isMomentumDialogKind(state.Kind)
	name := strings.TrimSpace(state.Inputs[0].Value())
	if name == "" {
		state.ErrorMessage = "Streak name is required"
		return state, nil, ""
	}
	requiredKind := sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind)
	if requiredKind == sharedtypes.MomentumTargetKindContext {
		required, err := ParseDurationSecondsInput(state.Inputs[1].Value(), true, "Required work")
		if err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
		state.HabitStreakDraft.RequiredCount = required
	} else {
		required, err := ParsePositiveIntInput(state.Inputs[1].Value(), true, "Required completions")
		if err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
		state.HabitStreakDraft.RequiredCount = required
	}
	state.HabitStreakDraft.Name = name
	if momentumMode {
		state.HabitStreakDraft.Description = ValueToPointer(strings.TrimSpace(state.Description.Value()))
	}
	state.HabitStreakDraft.Period = sharedtypes.NormalizeHabitStreakPeriod(
		state.HabitStreakDraft.Period,
	)
	if sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindHabit {
		if err := validateHabitMomentumDraft(state.HabitStreakDraft, state.HabitItems); err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
	}
	state.HabitStreakStep = reviewStepForState(state)
	state.HabitStreakCursor = 0
	state.ErrorMessage = ""
	return state, nil, ""
}

func updateHabitStreakTargets(state State, msg tea.KeyMsg) (State, *Action, string) {
	contextMode := sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
	if contextMode {
		return updateMomentumContextTargets(state, msg)
	}
	return updateMomentumHabitTargets(state, msg)
}

func updateMomentumHabitTargets(state State, msg tea.KeyMsg) (State, *Action, string) {
	total := len(state.HabitItems)
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		if state.Kind == "create_momentum" {
			state.HabitStreakStep = 1
			state.ChoiceCursor = kindChoiceCursorForTargetKind(state.HabitStreakDraft.TargetKind)
			state.HabitStreakCursor = 0
			return state, nil, ""
		}
		return Close(state), nil, ""
	case dialogActionMoveDown:
		if state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, 1)
			return clearDialogError(state), nil, ""
		}
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, 1)
	case dialogActionMoveUp:
		if state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, -1)
			return clearDialogError(state), nil, ""
		}
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, -1)
	case dialogActionMoveLeft:
		if state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, -1)
		}
	case dialogActionMoveRight:
		if state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, 1)
		}
	case dialogActionFocusPrev:
		if state.FocusIdx == int(habitTargetFocusList) {
			return clearDialogError(state), nil, ""
		}
		state.FocusIdx--
		return syncMomentumStepFocus(clearDialogError(state)), nil, ""
	case dialogActionToggle:
		if state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, 1)
			return clearDialogError(state), nil, ""
		}
		if state.HabitStreakCursor >= 0 && state.HabitStreakCursor < len(state.HabitItems) {
			habitID := state.HabitItems[state.HabitStreakCursor].ID
			state.HabitStreakDraft.HabitIDs = toggleHabitMembership(
				state.HabitStreakDraft.HabitIDs,
				habitID,
			)
		}
	case dialogActionFocusNext, dialogActionActivate, dialogActionPrimary:
		if action == dialogActionPrimary || state.FocusIdx == int(habitTargetFocusMatchMode) {
			state.HabitStreakStep = detailsStepForState(state)
			state.HabitStreakCursor = 0
			state.FocusIdx = 0
			return syncMomentumStepFocus(clearDialogError(state)), nil, ""
		}
		state.FocusIdx = int(habitTargetFocusMatchMode)
		return syncMomentumStepFocus(clearDialogError(state)), nil, ""
	default:
		switch msg.String() {
		case "a":
			if state.FocusIdx == int(habitTargetFocusList) {
				ids := make([]int64, 0, len(state.HabitItems))
				for _, item := range state.HabitItems {
					ids = append(ids, item.ID)
				}
				state.HabitStreakDraft.HabitIDs = ids
			}
		case "c":
			if state.FocusIdx == int(habitTargetFocusList) {
				state.HabitStreakDraft.HabitIDs = nil
			}
		}
	}
	return clearDialogError(state), nil, ""
}

func updateMomentumContextTargets(state State, msg tea.KeyMsg) (State, *Action, string) {
	state = syncMomentumStepFocus(state)
	if state.FocusIdx == int(contextTargetFocusRepo) ||
		state.FocusIdx == int(contextTargetFocusStream) {
		if isMomentumContextTextInputKey(msg) {
			return updateMomentumContextTextInput(state, msg)
		}
	}
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		if state.Kind == "create_momentum" {
			state.HabitStreakStep = 1
			state.ChoiceCursor = kindChoiceCursorForTargetKind(state.HabitStreakDraft.TargetKind)
			state.HabitStreakCursor = 0
			state.FocusIdx = 0
			return syncMomentumStepFocus(state), nil, ""
		}
		return Close(state), nil, ""
	case dialogActionFocusPrev:
		if state.FocusIdx == int(contextTargetFocusRepo) {
			if state.Kind == "create_momentum" {
				state.HabitStreakStep = 1
				state.ChoiceCursor = kindChoiceCursorForTargetKind(state.HabitStreakDraft.TargetKind)
				state.HabitStreakCursor = 0
				state.FocusIdx = 0
				return syncMomentumStepFocus(clearDialogError(state)), nil, ""
			}
			return clearDialogError(state), nil, ""
		}
		state.FocusIdx--
		return syncMomentumStepFocus(clearDialogError(state)), nil, ""
	case dialogActionFocusNext:
		if state.FocusIdx == momentumTargetMatchModeFocusIndex(state) {
			state.HabitStreakStep = detailsStepForState(state)
			state.HabitStreakCursor = 0
			state.FocusIdx = 0
			return syncMomentumStepFocus(clearDialogError(state)), nil, ""
		}
		state.FocusIdx++
		return syncMomentumStepFocus(clearDialogError(state)), nil, ""
	case dialogActionMoveDown:
		return updateMomentumContextSelection(state, 1), nil, ""
	case dialogActionMoveUp:
		return updateMomentumContextSelection(state, -1), nil, ""
	case dialogActionMoveLeft:
		return updateMomentumContextSelection(state, -1), nil, ""
	case dialogActionMoveRight:
		return updateMomentumContextSelection(state, 1), nil, ""
	case dialogActionToggle, dialogActionActivate:
		return handleMomentumContextActivation(state), nil, ""
	case dialogActionPrimary:
		state.HabitStreakStep = detailsStepForState(state)
		state.HabitStreakCursor = 0
		state.FocusIdx = 0
		return syncMomentumStepFocus(clearDialogError(state)), nil, ""
	default:
		switch msg.String() {
		case "c":
			if state.FocusIdx == int(contextTargetFocusSelected) {
				state.HabitStreakDraft.Contexts = nil
				state.HabitStreakCursor = 0
				return clearDialogError(state), nil, ""
			}
		}
	}
	return clearDialogError(state), nil, ""
}

func updateHabitStreakReview(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch action := dialogActionForKey(state, msg.String()); action {
	case dialogActionCancel:
		state.HabitStreakStep = detailsStepForState(state)
		state.HabitStreakCursor = 0
		return state, nil, ""
	case dialogActionFocusPrev, dialogActionMoveLeft:
		state.HabitStreakStep = detailsStepForState(state)
		return state, nil, ""
	default:
		if action == dialogActionPrimary {
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

func detailsStepForState(state State) int {
	if state.Kind == "create_momentum" {
		return 3
	}
	return 2
}

func reviewStepForState(state State) int {
	if state.Kind == "create_momentum" {
		return 4
	}
	return 3
}

func kindChoiceCursorForTargetKind(kind sharedtypes.MomentumTargetKind) int {
	switch sharedtypes.NormalizeMomentumTargetKind(kind) {
	case sharedtypes.MomentumTargetKindContext:
		return 1
	default:
		return 0
	}
}

func toggleMomentumContextMembership(
	values []sharedtypes.MomentumContext,
	value sharedtypes.MomentumContext,
) []sharedtypes.MomentumContext {
	for i, current := range values {
		if current == value {
			return append(values[:i], values[i+1:]...)
		}
	}
	return append(values, value)
}

type MomentumContextChoice struct {
	Context sharedtypes.MomentumContext
	Label   string
}

func MomentumContextOptions(habits []sharedtypes.HabitWithMeta) []MomentumContextChoice {
	if len(habits) == 0 {
		return nil
	}
	seen := map[sharedtypes.MomentumContext]struct{}{}
	out := make([]MomentumContextChoice, 0, len(habits))
	for _, habit := range habits {
		context := sharedtypes.MomentumContext{
			RepoID:   habit.RepoID,
			StreamID: int64Ptr(habit.StreamID),
		}
		if _, ok := seen[context]; ok {
			continue
		}
		seen[context] = struct{}{}
		label := strings.TrimSpace(habit.RepoName)
		if stream := strings.TrimSpace(habit.StreamName); stream != "" {
			if label != "" {
				label += " / " + stream
			} else {
				label = stream
			}
		}
		if label == "" {
			label = fmt.Sprintf("%d / %d", habit.RepoID, habit.StreamID)
		}
		out = append(out, MomentumContextChoice{
			Context: context,
			Label:   label,
		})
	}
	return out
}

func nextMomentumMatchMode(
	current sharedtypes.MomentumMatchMode,
	dir int,
) sharedtypes.MomentumMatchMode {
	options := []sharedtypes.MomentumMatchMode{
		sharedtypes.MomentumMatchModeAny,
		sharedtypes.MomentumMatchModeAll,
	}
	return options[nextIndex(sharedtypes.NormalizeMomentumMatchMode(current), options, dir)]
}

func updateMomentumContextSelection(state State, dir int) State {
	switch state.FocusIdx {
	case int(contextTargetFocusRepo):
		options := momentumRepoOptions(state)
		state.RepoIndex = ShiftSelection(state.RepoIndex, len(options), dir)
		if len(options) == 0 {
			state.RepoIndex = -1
			state.StreamIndex = -1
			return clearDialogError(state)
		}
		state.StreamIndex = -1
	case int(contextTargetFocusStream):
		options := momentumStreamOptions(state)
		state.StreamIndex = shiftMomentumStreamIndex(state.StreamIndex, len(options)-1, dir)
		if len(options) == 0 {
			state.StreamIndex = -1
		}
	case int(contextTargetFocusSelected):
		state.HabitStreakCursor = ShiftSelection(
			state.HabitStreakCursor,
			len(state.HabitStreakDraft.Contexts),
			dir,
		)
	case int(contextTargetFocusMatchMode):
		state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, dir)
	}
	return clearDialogError(state)
}

func handleMomentumContextActivation(state State) State {
	switch state.FocusIdx {
	case int(contextTargetFocusSelected):
		if state.HabitStreakCursor >= 0 && state.HabitStreakCursor < len(state.HabitStreakDraft.Contexts) {
			state.HabitStreakDraft.Contexts = append(
				append([]sharedtypes.MomentumContext(nil), state.HabitStreakDraft.Contexts[:state.HabitStreakCursor]...),
				state.HabitStreakDraft.Contexts[state.HabitStreakCursor+1:]...,
			)
			if state.HabitStreakCursor >= len(state.HabitStreakDraft.Contexts) && state.HabitStreakCursor > 0 {
				state.HabitStreakCursor--
			}
		}
	case int(contextTargetFocusMatchMode):
		state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, 1)
	default:
		if candidate, ok := currentMomentumContextCandidate(state); ok {
			state.HabitStreakDraft.Contexts = toggleMomentumContextMembership(
				state.HabitStreakDraft.Contexts,
				candidate,
			)
		}
	}
	return clearDialogError(state)
}

func momentumRepoOptions(state State) []SelectorOption {
	options := DefaultRepoOptions(
		[]textinput.Model{state.MomentumRepoInput},
		state.MomentumRepos,
	)
	return stripCreateOptions(options)
}

func momentumStreamOptions(state State) []SelectorOption {
	options := []SelectorOption{{ID: "__any__", Label: "Any Stream"}}
	options = append(options, DefaultStreamOptions(
		[]textinput.Model{state.MomentumRepoInput, state.MomentumStreamInput},
		state.RepoIndex,
		state.MomentumRepos,
		state.MomentumAllIssues,
		state.MomentumStreams,
		nil,
	)...)
	return stripCreateOptions(options)
}

func stripCreateOptions(options []SelectorOption) []SelectorOption {
	out := make([]SelectorOption, 0, len(options))
	for _, option := range options {
		if option.ID == "__new__" {
			continue
		}
		out = append(out, option)
	}
	return out
}

func currentMomentumContextCandidate(state State) (sharedtypes.MomentumContext, bool) {
	repoOption, ok := selectedOption(momentumRepoOptions(state), state.RepoIndex)
	if !ok {
		return sharedtypes.MomentumContext{}, false
	}
	repoID, err := strconv.ParseInt(repoOption.ID, 10, 64)
	if err != nil || repoID <= 0 {
		return sharedtypes.MomentumContext{}, false
	}
	if state.StreamIndex < 0 {
		return sharedtypes.MomentumContext{RepoID: repoID}, true
	}
	streamOption, ok := selectedMomentumStreamOption(momentumStreamOptions(state), state.StreamIndex)
	if !ok {
		return sharedtypes.MomentumContext{RepoID: repoID}, true
	}
	if streamOption.ID == "__any__" {
		return sharedtypes.MomentumContext{RepoID: repoID}, true
	}
	streamID, err := strconv.ParseInt(streamOption.ID, 10, 64)
	if err != nil || streamID <= 0 {
		return sharedtypes.MomentumContext{RepoID: repoID}, true
	}
	return sharedtypes.MomentumContext{RepoID: repoID, StreamID: int64Ptr(streamID)}, true
}

func syncMomentumStepFocus(state State) State {
	state = SyncDialogFocus(state)
	return syncMomentumTargetFocus(state)
}

func syncMomentumTargetFocus(state State) State {
	state.MomentumRepoInput.Blur()
	state.MomentumStreamInput.Blur()
	if !isMomentumDialogKind(state.Kind) {
		return state
	}
	if state.HabitStreakStep == detailsStepForState(state) {
		return syncHabitStreakDetailFocus(state)
	}
	if state.HabitStreakStep != detailsStepForState(state)-1 {
		return state
	}
	if sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) != sharedtypes.MomentumTargetKindContext {
		return state
	}
	switch state.FocusIdx {
	case int(contextTargetFocusRepo):
		state.MomentumRepoInput.Focus()
	case int(contextTargetFocusStream):
		state.MomentumStreamInput.Focus()
	}
	return state
}

func syncHabitStreakDetailFocus(state State) State {
	if !isMomentumDialogKind(state.Kind) {
		return state
	}
	row := habitStreakDetailRowForFocus(true, state.FocusIdx)
	return habitStreakSetDetailFocus(state, row)
}

func updateMomentumContextTextInput(state State, msg tea.KeyMsg) (State, *Action, string) {
	if state.FocusIdx == int(contextTargetFocusRepo) {
		var cmd tea.Cmd
		before := strings.TrimSpace(state.MomentumRepoInput.Value())
		state.MomentumRepoInput, cmd = state.MomentumRepoInput.Update(msg)
		_ = cmd
		if strings.TrimSpace(state.MomentumRepoInput.Value()) != before {
			if strings.TrimSpace(state.MomentumRepoInput.Value()) == "" {
				state.RepoIndex = -1
				state.StreamIndex = -1
			} else {
				state.RepoIndex = 0
				state.StreamIndex = -1
			}
		}
		return clearDialogError(state), nil, ""
	}
	if state.FocusIdx == int(contextTargetFocusStream) {
		var cmd tea.Cmd
		before := strings.TrimSpace(state.MomentumStreamInput.Value())
		state.MomentumStreamInput, cmd = state.MomentumStreamInput.Update(msg)
		_ = cmd
		if strings.TrimSpace(state.MomentumStreamInput.Value()) != before {
			if strings.TrimSpace(state.MomentumStreamInput.Value()) == "" {
				state.StreamIndex = -1
			} else {
				state.StreamIndex = 0
			}
		}
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func shiftMomentumStreamIndex(current, totalReal, dir int) int {
	if totalReal <= 0 {
		return -1
	}
	positionCount := totalReal + 1
	position := current + 1
	if position < 0 || position >= positionCount {
		position = 0
	}
	next := position + dir
	if next < 0 {
		next = positionCount - 1
	}
	if next >= positionCount {
		next = 0
	}
	return next - 1
}

func selectedMomentumStreamOption(options []SelectorOption, index int) (SelectorOption, bool) {
	if len(options) == 0 {
		return SelectorOption{}, false
	}
	if index < 0 {
		return options[0], true
	}
	optionIndex := index + 1
	if optionIndex >= len(options) {
		optionIndex = len(options) - 1
	}
	return options[optionIndex], true
}

func isMomentumContextTextInputKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace, tea.KeyBackspace, tea.KeyDelete:
		return true
	default:
		return false
	}
}

func momentumTargetMatchModeFocusIndex(state State) int {
	if sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) == sharedtypes.MomentumTargetKindContext {
		return int(contextTargetFocusMatchMode)
	}
	return int(habitTargetFocusMatchMode)
}

func habitStreakDefaultRequiredCount(kind sharedtypes.MomentumTargetKind) int {
	switch sharedtypes.NormalizeMomentumTargetKind(kind) {
	case sharedtypes.MomentumTargetKindContext:
		return 3600
	default:
		return 1
	}
}

func syncHabitStreakThresholdInput(state State) State {
	if len(state.Inputs) < 2 {
		return state
	}
	input := state.Inputs[1]
	switch sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		input.Placeholder = "Required work (e.g. 1h34m23s)"
		input = withTimePrompt(state, input)
		input.SetValue(FormatDurationSecondsInput(max(1, state.HabitStreakDraft.RequiredCount)))
	default:
		input.Placeholder = "Required completions"
		input.SetValue(strconv.Itoa(max(1, state.HabitStreakDraft.RequiredCount)))
	}
	state.Inputs[1] = input
	return state
}

func isMomentumThresholdTextInputKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace, tea.KeyBackspace, tea.KeyDelete:
		return true
	default:
		return false
	}
}

func habitMomentumDraftCapacity(
	def sharedtypes.HabitStreakDefinition,
	habits []sharedtypes.HabitWithMeta,
) sharedutils.HabitMomentumCapacity {
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) != sharedtypes.MomentumTargetKindHabit {
		return sharedutils.HabitMomentumCapacity{}
	}
	habitMap := make(map[int64]sharedtypes.Habit, len(habits))
	for _, habit := range habits {
		habitMap[habit.ID] = habit.Habit
	}
	selected := make([]sharedtypes.Habit, 0, len(def.HabitIDs))
	for _, habitID := range def.HabitIDs {
		habit, ok := habitMap[habitID]
		if !ok {
			return sharedutils.HabitMomentumCapacity{Reason: "selected habits are no longer available"}
		}
		selected = append(selected, habit)
	}
	return sharedutils.HabitMomentumCapacityForSelection(selected, def.Period, def.MatchMode)
}

func validateHabitMomentumDraft(
	def sharedtypes.HabitStreakDefinition,
	habits []sharedtypes.HabitWithMeta,
) error {
	capacity := habitMomentumDraftCapacity(def, habits)
	if !capacity.Valid {
		return fmt.Errorf("%s", capacity.Reason)
	}
	if def.RequiredCount < 1 {
		return fmt.Errorf("required completions must be positive")
	}
	if def.RequiredCount > capacity.MaxCount {
		return fmt.Errorf(
			"required completions exceed the maximum possible for the selected habits (%d)",
			capacity.MaxCount,
		)
	}
	return nil
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
