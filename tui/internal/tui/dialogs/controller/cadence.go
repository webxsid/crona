package controller

import "strings"

// CadenceLabels are the cadence picker's choices, in order.
var CadenceLabels = []string{"Daily", "Weekdays", "Custom days"}

const cadenceCustom = 2

// cadenceLayout places a dialog's cadence controls: the picker's focus, the Days row's
// focus (always the last stop), the schedule input they write, and the focus count.
type cadenceLayout struct{ cadence, days, input, count int }

func cadenceLayoutFor(state State) (cadenceLayout, bool) {
	switch state.Kind {
	case "create_habit":
		return cadenceLayout{cadence: 4, days: 6, input: 3, count: 7}, true
	case "edit_habit":
		return cadenceLayout{cadence: 2, days: 4, input: 1, count: 5}, true
	case "create_alert_reminder", "edit_alert_reminder":
		return cadenceLayout{cadence: 0, days: 2, input: 0, count: 3}, true
	}
	return cadenceLayout{}, false
}

// CadenceLayout reports where a dialog's cadence picker and Days row sit in its focus order.
func CadenceLayout(state State) (cadence, days, count int, ok bool) {
	layout, ok := cadenceLayoutFor(state)
	return layout.cadence, layout.days, layout.count, ok
}

// initCadence reads the schedule input into the cadence picker.
func initCadence(state State) State {
	layout, ok := cadenceLayoutFor(state)
	if !ok {
		return state
	}
	value := strings.TrimSpace(strings.ToLower(state.Inputs[layout.input].Value()))
	state.CadenceIdx, state.CadenceDays = 0, nil
	switch value {
	case "", "daily":
	case "weekdays":
		state.CadenceIdx = 1
	default:
		if days, err := ParseWeekdayList(value); err == nil && len(days) > 0 {
			state.CadenceIdx, state.CadenceDays = cadenceCustom, days
		}
	}
	return writeCadence(state, layout)
}

// writeCadence stores the picker's choice in the schedule input, which saving parses.
func writeCadence(state State, layout cadenceLayout) State {
	value := "daily"
	switch state.CadenceIdx {
	case 1:
		value = "weekdays"
	case cadenceCustom:
		value = strings.Join(WeekdayTokens(normalizedWeekdays(state.CadenceDays)), ",")
	}
	state.Inputs[layout.input].SetValue(value)
	return state
}

// cadenceMissingDays is true when Custom days is chosen with no day ticked.
func cadenceMissingDays(state State) bool {
	_, ok := cadenceLayoutFor(state)
	return ok && state.CadenceIdx == cadenceCustom && len(state.CadenceDays) == 0
}

// updateCadenceKey moves focus with tab (skipping Days unless Custom days is chosen) and
// handles the picker and Days row; it reports whether it consumed the key.
func updateCadenceKey(state State, key string) (State, bool) {
	layout, ok := cadenceLayoutFor(state)
	if !ok || key == "esc" || isDialogSubmitKey(state, key) {
		return state, false
	}
	switch {
	case key == "tab" || key == "shift+tab":
		dir := ternaryDir(key)
		next := (state.FocusIdx + dir + layout.count) % layout.count
		if next == layout.days && state.CadenceIdx != cadenceCustom {
			next = (next + dir + layout.count) % layout.count
		}
		state.FocusIdx = next
		return SyncDialogFocus(clearDialogError(state)), true
	case state.FocusIdx == layout.cadence:
		switch key {
		case "left":
			state.CadenceIdx = (state.CadenceIdx + len(CadenceLabels) - 1) % len(CadenceLabels)
		case "right":
			state.CadenceIdx = (state.CadenceIdx + 1) % len(CadenceLabels)
		default:
			return state, true
		}
		return clearDialogError(writeCadence(state, layout)), true
	case state.FocusIdx == layout.days:
		switch key {
		case "left":
			state.CadenceCursor = (state.CadenceCursor + 6) % 7
		case "right":
			state.CadenceCursor = (state.CadenceCursor + 1) % 7
		case " ", "space":
			state.CadenceDays = toggleWeekday(state.CadenceDays, (state.CadenceCursor+1)%7)
			return clearDialogError(writeCadence(state, layout)), true
		}
		return state, true
	}
	return state, false
}
