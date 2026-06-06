package dialogs

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	viewchrome "crona/tui/internal/tui/views/chrome"
)

func renderHabitStreakDialog(theme Theme, state controllerpkg.State) string {
	momentumMode := isMomentumDialogKind(state.Kind)
	steps := []string{"Details", "Habits", "Review"}
	activeStep := max(0, state.HabitStreakStep-1)
	progress := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf("%d.%s", i+1, step)
		if i == activeStep {
			progress = append(progress, theme.StyleCursor.Render(label))
		} else {
			progress = append(progress, theme.StyleDim.Render(label))
		}
	}
	rows := []string{
		theme.StylePaneTitle.Render(habitStreakDialogTitle(momentumMode)),
		"",
		strings.Join(progress, "   "),
		"",
	}
	switch state.HabitStreakStep {
	case 1:
		nameLabel := "Name"
		if momentumMode {
			nameLabel = "Momentum name"
		}
		periodRowIdx := 1
		countRowIdx := 2
		if momentumMode {
			periodRowIdx = 2
			countRowIdx = 3
		}
		rows = append(rows,
			habitStreakRowLabel(theme, state, 0, nameLabel),
			dialogInputView(state, 0),
		)
		if momentumMode {
			rows = append(
				rows,
				"",
				habitStreakRowLabel(theme, state, 1, "Description (Optional)"),
				state.Description.View(),
			)
		}
		rows = append(
			rows,
			"",
			habitStreakRowLabel(theme, state, periodRowIdx, "Period"),
			renderHabitStreakPeriodChoice(theme, state),
			"",
			habitStreakRowLabel(theme, state, countRowIdx, "Required completions per bucket (daily = 1)"),
			dialogInputView(state, 1),
		)
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[tab] next field   [h/l] period   [enter] next   [esc] cancel",
		)
	case 2:
		rows = append(rows, theme.StyleDim.Render("Select contributing habits"))
		if len(state.HabitItems) == 0 {
			rows = append(rows, theme.StyleDim.Render("No habits available"))
		}
		for i, item := range state.HabitItems {
			prefix := "[ ] "
			if slices.Contains(state.HabitStreakDraft.HabitIDs, item.ID) {
				prefix = "[x] "
			}
			line := fmt.Sprintf("%s%s  %s / %s", prefix, item.Name, item.RepoName, item.StreamName)
			if i == state.HabitStreakCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			"[space] toggle   [a] all   [c] none   [tab] review",
		)
	case 3:
		description := ""
		if state.HabitStreakDraft.Description != nil {
			description = strings.TrimSpace(*state.HabitStreakDraft.Description)
		}
		rows = append(
			rows,
			theme.StyleDim.Render("Name"),
			theme.StyleHeader.Render(fallback(state.HabitStreakDraft.Name, "-")),
		)
		if description != "" {
			rows = append(
				rows,
				"",
				theme.StyleDim.Render("Description"),
				theme.StyleHeader.Render(description),
			)
		}
		rows = append(
			rows,
			"",
			theme.StyleDim.Render("Rule"),
			theme.StyleHeader.Render(
				fmt.Sprintf(
					"%d+ completions per %s",
					max(1, state.HabitStreakDraft.RequiredCount),
					strings.ToLower(habitStreakPeriodLabel(state.HabitStreakDraft.Period)),
				),
			),
			"",
			theme.StyleDim.Render("Habits"),
			theme.StyleHeader.Render(
				habitStreakHabitSummary(state.HabitStreakDraft.HabitIDs, state.HabitItems),
			),
		)
		submitLabel := "save streak"
		if momentumMode {
			submitLabel = "save momentum"
		}
		rows = appendDialogFooter(
			theme,
			state,
			rows,
			dialogSubmitHint(state, submitLabel)+"   [shift+tab] back   [esc] cancel",
		)
	}
	return modal(theme, state.Width, 88, theme.ColorCyan, rows)
}

func isMomentumDialogKind(kind string) bool {
	switch kind {
	case "create_momentum", "edit_momentum":
		return true
	default:
		return false
	}
}

func habitStreakDialogTitle(momentumMode bool) string {
	if momentumMode {
		return "Momentum"
	}
	return "Habit Streaks"
}

func renderHabitStreakPeriodChoice(theme Theme, state controllerpkg.State) string {
	options := []sharedtypes.HabitStreakPeriod{
		sharedtypes.HabitStreakPeriodDay,
		sharedtypes.HabitStreakPeriodWeek,
		sharedtypes.HabitStreakPeriodMonth,
	}
	parts := make([]string, 0, len(options))
	current := sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period)
	periodFocused := state.FocusIdx == 1
	if isMomentumDialogKind(state.Kind) {
		periodFocused = state.FocusIdx == 2
	}
	for _, option := range options {
		label := habitStreakPeriodLabel(option)
		if option == current {
			if periodFocused {
				parts = append(
					parts,
					theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+label),
				)
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
		return theme.StyleCursor.Render(viewchrome.SelectionCursor + " " + label)
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
