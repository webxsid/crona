package daily

import (
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"
)

func timerIdle(state types.ContentState) bool {
	return state.Timer == nil || state.Timer.State == "idle"
}

// newItemLine offers the create key on its own dim line at the top of a pane.
func newItemLine(theme types.Theme, label string) string {
	return viewchrome.KeyHint(theme, "a", label)
}

// newTaskLine pairs the create key with the ad hoc timer, which needs no selected task;
// the timer key shows only while no timer runs.
func newTaskLine(theme types.Theme, state types.ContentState, width int) string {
	hints := []string{newItemLine(theme, "new task")}
	if timerIdle(state) {
		hints = append(hints, viewchrome.KeyHint(theme, "A", "quick timer"))
	}
	return viewchrome.FitHints(hints, width)
}

// taskActionLine stands in for the focused selected card's meta line. It offers what the
// pane's key list did: keys that start or edit work only while no timer runs.
func taskActionLine(theme types.Theme, state types.ContentState, width int) string {
	hints := []string{enterHint(theme, state)}
	if state.SelectedIssueAdhoc {
		if timerIdle(state) {
			hints = append(hints, viewchrome.KeyHint(theme, "e", "convert"))
		}
		return "  " + viewchrome.FitHints(hints, width-3)
	}
	hints = append(hints,
		viewchrome.KeyHint(theme, "s", "status"),
		viewchrome.KeyHint(theme, "d", "due"),
		viewchrome.KeyHint(theme, "P", "pin"),
		viewchrome.KeyHint(theme, "+", "tags"),
	)
	if state.DailyTaskSection == "overdue" {
		hints = append(hints, viewchrome.KeyHint(theme, "r", "carry"))
	}
	if timerIdle(state) {
		hints = append(hints,
			viewchrome.KeyHint(theme, "f", "focus"),
			viewchrome.KeyHint(theme, "e", "edit"),
			viewchrome.KeyHint(theme, "D", "delete"),
			viewchrome.KeyHint(theme, "m", "log"),
		)
	}
	// The section indents each line a cell, so the hints get width less three.
	return "  " + viewchrome.FitHints(hints, width-3)
}

// habitActionLine sits under the focused selected habit.
func habitActionLine(theme types.Theme, state types.ContentState, width int) string {
	hints := []string{
		enterHint(theme, state),
		viewchrome.KeyHint(theme, "x", "done"),
		viewchrome.KeyHint(theme, "F", "fail"),
		viewchrome.KeyHint(theme, "e", "edit"),
		viewchrome.KeyHint(theme, "+", "tags"),
	}
	if timerIdle(state) {
		hints = append(hints,
			viewchrome.KeyHint(theme, "D", "delete"),
			viewchrome.KeyHint(theme, "m", "log"),
		)
	}
	return "    " + viewchrome.FitHints(hints, width-4)
}

// TaskActionLine renders the selected issue card action hints.
func TaskActionLine(theme types.Theme, state types.ContentState, width int) string {
	return taskActionLine(theme, state, width)
}

// enterHint names what enter does on a selected task or habit: it shows the details in the
// detail pane on wide screens, and opens them in a dialog where that pane is hidden or folded.
func enterHint(theme types.Theme, state types.ContentState) string {
	if viewhelpers.ClassFor(state.Width) == viewhelpers.WidthWide {
		return viewchrome.KeyHint(theme, "enter", "details")
	}
	return viewchrome.KeyHint(theme, "enter", "open")
}
