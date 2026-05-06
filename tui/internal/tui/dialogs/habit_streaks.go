package dialogs

import (
	"fmt"
	"strconv"
	"strings"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type habitStreakDetailRow int

const (
	habitStreakDetailRowName habitStreakDetailRow = iota
	habitStreakDetailRowPeriod
	habitStreakDetailRowCount
)

func updateHabitStreaks(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch state.HabitStreakStep {
	case 0:
		return updateHabitStreakManager(state, msg)
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

func updateHabitStreakManager(state State, msg tea.KeyMsg) (State, *Action, string) {
	total := len(state.HabitStreakDefs) + 1
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "j", "down":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, 1)
	case "k", "up":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, -1)
	case "n", "a":
		return openHabitStreakEditor(state, -1, sharedtypes.HabitStreakDefinition{
			Enabled:       true,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 1,
		}), nil, ""
	case "x", " ":
		if state.HabitStreakCursor < len(state.HabitStreakDefs) {
			state.HabitStreakDefs[state.HabitStreakCursor].Enabled = !state.HabitStreakDefs[state.HabitStreakCursor].Enabled
		}
	case "d", "backspace", "delete":
		if state.HabitStreakCursor < len(state.HabitStreakDefs) {
			idx := state.HabitStreakCursor
			state.HabitStreakDefs = append(state.HabitStreakDefs[:idx], state.HabitStreakDefs[idx+1:]...)
			if state.HabitStreakCursor >= len(state.HabitStreakDefs) && state.HabitStreakCursor > 0 {
				state.HabitStreakCursor--
			}
		}
	case "enter":
		if state.HabitStreakCursor >= len(state.HabitStreakDefs) {
			return openHabitStreakEditor(state, -1, sharedtypes.HabitStreakDefinition{
				Enabled:       true,
				Period:        sharedtypes.HabitStreakPeriodDay,
				RequiredCount: 1,
			}), nil, ""
		}
		return openHabitStreakEditor(state, state.HabitStreakCursor, state.HabitStreakDefs[state.HabitStreakCursor]), nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			return Close(state), &Action{
				Kind:            "patch_setting",
				SettingKey:      sharedtypes.CoreSettingsKeyHabitStreakDefs,
				HabitStreakDefs: sharedtypes.NormalizeHabitStreakDefinitions(state.HabitStreakDefs),
			}, ""
		}
	}
	return clearDialogError(state), nil, ""
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
	row := habitStreakDetailRow(state.FocusIdx)
	switch msg.String() {
	case "esc":
		return habitStreakBackToManager(state), nil, ""
	case "tab":
		if row == habitStreakDetailRowCount {
			return moveHabitStreakToHabitSelection(state)
		}
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(row, 1))
		return state, nil, ""
	case "shift+tab":
		state = habitStreakSetDetailFocus(state, habitStreakDetailRowNext(row, -1))
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
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowPeriod)
			return state, nil, ""
		case habitStreakDetailRowPeriod:
			state = habitStreakSetDetailFocus(state, habitStreakDetailRowCount)
			return state, nil, ""
		default:
			return moveHabitStreakToHabitSelection(state)
		}
	}
	if inputIdx, ok := habitStreakInputIndex(row); ok && inputIdx >= 0 && inputIdx < len(state.Inputs) {
		var cmd tea.Cmd
		state.Inputs[inputIdx], cmd = state.Inputs[inputIdx].Update(msg)
		_ = cmd
	}
	state = habitStreakApplyPeriodRule(state)
	return clearDialogError(state), nil, ""
}

func moveHabitStreakToHabitSelection(state State) (State, *Action, string) {
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
	state.HabitStreakDraft.RequiredCount = required
	state.HabitStreakDraft.Period = sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period)
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
			return habitStreakBackToManager(state), nil, ""
		case "enter", "tab":
			state.HabitStreakStep = 3
			return state, nil, ""
		}
		return state, nil, ""
	}
	switch msg.String() {
	case "esc":
		return habitStreakBackToManager(state), nil, ""
	case "j", "down":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, 1)
	case "k", "up":
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, total, -1)
	case " ", "x":
		habitID := state.HabitItems[state.HabitStreakCursor].ID
		state.HabitStreakDraft.HabitIDs = toggleHabitMembership(state.HabitStreakDraft.HabitIDs, habitID)
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
		return habitStreakBackToManager(state), nil, ""
	case "shift+tab", "left", "h":
		state.HabitStreakStep = 2
		return state, nil, ""
	default:
		if isDialogSubmitKey(state, msg.String()) {
			defs := append([]sharedtypes.HabitStreakDefinition(nil), state.HabitStreakDefs...)
			draft := sharedtypes.NormalizeHabitStreakDefinition(state.HabitStreakDraft)
			if state.HabitStreakEditIdx >= 0 && state.HabitStreakEditIdx < len(defs) {
				defs[state.HabitStreakEditIdx] = draft
			} else {
				defs = append(defs, draft)
			}
			state.HabitStreakDefs = sharedtypes.NormalizeHabitStreakDefinitions(defs)
			state.HabitStreakStep = 0
			state.HabitStreakCursor = 0
			state.HabitStreakEditIdx = -1
			state.Inputs = nil
			state.FocusIdx = 0
			return clearDialogError(state), nil, ""
		}
	}
	return clearDialogError(state), nil, ""
}

func habitStreakBackToManager(state State) State {
	state.HabitStreakStep = 0
	state.HabitStreakCursor = 0
	state.HabitStreakEditIdx = -1
	state.Inputs = nil
	state.FocusIdx = 0
	state.ErrorMessage = ""
	return state
}

func renderHabitStreakDialog(theme Theme, state State) string {
	steps := []string{"Manage", "Details", "Habits", "Review"}
	progress := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf("%d.%s", i+1, step)
		if i == state.HabitStreakStep {
			progress = append(progress, theme.StyleCursor.Render(label))
		} else {
			progress = append(progress, theme.StyleDim.Render(label))
		}
	}
	rows := []string{
		theme.StylePaneTitle.Render("Habit Streaks"),
		"",
		strings.Join(progress, "   "),
		"",
	}
	switch state.HabitStreakStep {
	case 0:
		rows = append(rows, theme.StyleDim.Render("Create and manage named habit streaks"))
		if len(state.HabitStreakDefs) == 0 {
			rows = append(rows, theme.StyleDim.Render("No custom streaks configured"))
		}
		for i, def := range state.HabitStreakDefs {
			prefix := "  "
			if i == state.HabitStreakCursor {
				prefix = "▶ "
			}
			status := "off"
			if def.Enabled {
				status = "on"
			}
			line := fmt.Sprintf("%s%s  %s  %s  %d/%s  %d habits", prefix, fallback(def.Name, "(unnamed)"), status, habitStreakPeriodLabel(def.Period), max(1, def.RequiredCount), strings.ToLower(habitStreakPeriodLabel(def.Period)), len(def.HabitIDs))
			if i == state.HabitStreakCursor {
				rows = append(rows, theme.StyleCursor.Render(line))
			} else {
				rows = append(rows, theme.StyleNormal.Render(line))
			}
		}
		createLine := "  + Create new streak"
		if state.HabitStreakCursor == len(state.HabitStreakDefs) {
			createLine = "▶ + Create new streak"
			rows = append(rows, theme.StyleCursor.Render(createLine))
		} else {
			rows = append(rows, theme.StyleNormal.Render(createLine))
		}
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save")+"   [enter] edit   [n] new   [x] toggle   [d] delete")
	case 1:
		rows = append(rows,
			habitStreakRowLabel(theme, state, habitStreakDetailRowName, "Name"),
			dialogInputView(state, 0),
			"",
			habitStreakRowLabel(theme, state, habitStreakDetailRowPeriod, "Period"),
			renderHabitStreakPeriodChoice(theme, state),
			"",
			habitStreakRowLabel(theme, state, habitStreakDetailRowCount, "Required completions per bucket (daily = 1)"),
			dialogInputView(state, 1),
		)
		rows = appendDialogFooter(theme, state, rows, "[tab] next field   [h/l] period   [enter] next   [esc] cancel")
	case 2:
		rows = append(rows, theme.StyleDim.Render("Select contributing habits"))
		if len(state.HabitItems) == 0 {
			rows = append(rows, theme.StyleDim.Render("No habits available"))
		}
		for i, item := range state.HabitItems {
			prefix := "[ ] "
			if containsHabitID(state.HabitStreakDraft.HabitIDs, item.ID) {
				prefix = "[x] "
			}
			line := fmt.Sprintf("%s%s  %s / %s", prefix, item.Name, item.RepoName, item.StreamName)
			if i == state.HabitStreakCursor {
				rows = append(rows, theme.StyleCursor.Render("▶ "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		rows = appendDialogFooter(theme, state, rows, "[space] toggle   [a] all   [c] none   [tab] review")
	case 3:
		rows = append(rows,
			theme.StyleDim.Render("Name"),
			theme.StyleHeader.Render(fallback(state.HabitStreakDraft.Name, "-")),
			"",
			theme.StyleDim.Render("Rule"),
			theme.StyleHeader.Render(fmt.Sprintf("%d+ completions per %s", max(1, state.HabitStreakDraft.RequiredCount), strings.ToLower(habitStreakPeriodLabel(state.HabitStreakDraft.Period)))),
			"",
			theme.StyleDim.Render("Habits"),
			theme.StyleHeader.Render(habitStreakHabitSummary(state.HabitStreakDraft.HabitIDs, state.HabitItems)),
		)
		rows = appendDialogFooter(theme, state, rows, dialogSubmitHint(state, "save streak")+"   [shift+tab] back   [esc] cancel")
	}
	return modal(theme, state.Width, 88, theme.ColorCyan, rows)
}

func renderHabitStreakPeriodChoice(theme Theme, state State) string {
	options := []sharedtypes.HabitStreakPeriod{
		sharedtypes.HabitStreakPeriodDay,
		sharedtypes.HabitStreakPeriodWeek,
		sharedtypes.HabitStreakPeriodMonth,
	}
	parts := make([]string, 0, len(options))
	current := sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period)
	for _, option := range options {
		label := habitStreakPeriodLabel(option)
		if option == current {
			if habitStreakDetailRow(state.FocusIdx) == habitStreakDetailRowPeriod {
				parts = append(parts, theme.StyleCursor.Render("▶ "+label))
			} else {
				parts = append(parts, theme.StyleCursor.Render(label))
			}
		} else {
			parts = append(parts, theme.StyleDim.Render(label))
		}
	}
	return strings.Join(parts, "   ")
}

func habitStreakRowLabel(theme Theme, state State, row habitStreakDetailRow, label string) string {
	if habitStreakDetailRow(state.FocusIdx) == row {
		return theme.StyleCursor.Render("▶ " + label)
	}
	return theme.StyleDim.Render(label)
}

func nextHabitStreakPeriod(current sharedtypes.HabitStreakPeriod, dir int) sharedtypes.HabitStreakPeriod {
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

func habitStreakDetailRowNext(row habitStreakDetailRow, dir int) habitStreakDetailRow {
	rows := []habitStreakDetailRow{
		habitStreakDetailRowName,
		habitStreakDetailRowPeriod,
		habitStreakDetailRowCount,
	}
	return rows[nextIndex(row, rows, dir)]
}

func habitStreakSetDetailFocus(state State, row habitStreakDetailRow) State {
	state.FocusIdx = int(row)
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	if inputIdx, ok := habitStreakInputIndex(row); ok && inputIdx >= 0 && inputIdx < len(state.Inputs) {
		state.Inputs[inputIdx].Focus()
	}
	return state
}

func habitStreakApplyPeriodRule(state State) State {
	if sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period) == sharedtypes.HabitStreakPeriodDay {
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

func habitStreakPeriodLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "Weekly"
	case sharedtypes.HabitStreakPeriodMonth:
		return "Monthly"
	default:
		return "Daily"
	}
}

func toggleHabitMembership(values []int64, habitID int64) []int64 {
	for i, value := range values {
		if value == habitID {
			return append(values[:i], values[i+1:]...)
		}
	}
	return append(values, habitID)
}

func containsHabitID(values []int64, habitID int64) bool {
	for _, value := range values {
		if value == habitID {
			return true
		}
	}
	return false
}

func habitStreakHabitSummary(ids []int64, habits []sharedtypes.HabitWithMeta) string {
	if len(ids) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		for _, habit := range habits {
			if habit.ID == id {
				labels = append(labels, habit.Name)
				break
			}
		}
	}
	if len(labels) == 0 {
		return fmt.Sprintf("%d habits", len(ids))
	}
	if len(labels) <= 3 {
		return strings.Join(labels, ", ")
	}
	return fmt.Sprintf("%s, %s, %s +%d", labels[0], labels[1], labels[2], len(labels)-3)
}
