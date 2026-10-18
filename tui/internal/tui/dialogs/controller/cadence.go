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

// updateCadenceKey moves focus with tab in cadenceFocusOrder and
// handles the picker and Days row; it reports whether it consumed the key.
func updateCadenceKey(state State, key string) (State, bool) {
	layout, ok := cadenceLayoutFor(state)
	if !ok || key == "esc" || isDialogSubmitKey(state, key) {
		return state, false
	}
	switch {
	case key == "tab" || key == "shift+tab":
		order := cadenceFocusOrder(state, layout)
		at := 0
		for i, focus := range order {
			if focus == state.FocusIdx {
				at = i
			}
		}
		state.FocusIdx = order[(at+ternaryDir(key)+len(order))%len(order)]
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

// cadenceFocusOrder is the tab order: Days sits right after the cadence picker, and only
// while Custom days is chosen.
func cadenceFocusOrder(state State, layout cadenceLayout) []int {
	order := make([]int, 0, layout.count)
	for focus := 0; focus < layout.count; focus++ {
		if focus == layout.days {
			continue
		}
		order = append(order, focus)
		if focus == layout.cadence && state.CadenceIdx == cadenceCustom {
			order = append(order, layout.days)
		}
	}
	return order
}
