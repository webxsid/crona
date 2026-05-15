package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func renderHabitStreakDialog(theme Theme, state controllerpkg.State) string {
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
			habitStreakRowLabel(theme, state, 0, "Name"),
			dialogInputView(state, 0),
			"",
			habitStreakRowLabel(theme, state, 1, "Period"),
			renderHabitStreakPeriodChoice(theme, state),
			"",
			habitStreakRowLabel(theme, state, 2, "Required completions per bucket (daily = 1)"),
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

func renderHabitStreakPeriodChoice(theme Theme, state controllerpkg.State) string {
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
			if state.FocusIdx == 1 {
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

func habitStreakRowLabel(theme Theme, state controllerpkg.State, row int, label string) string {
	if state.FocusIdx == row {
		return theme.StyleCursor.Render("▶ " + label)
	}
	return theme.StyleDim.Render(label)
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
