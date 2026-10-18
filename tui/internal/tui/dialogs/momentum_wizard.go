package dialogs

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// renderHabitStreakDialog draws the momentum (and habit streak) wizard: a step strip, the
// step's body and its keys, with the step in the top border.
func renderHabitStreakDialog(theme Theme, state controllerpkg.State) string {
	momentumMode := isMomentumDialogKind(state.Kind)
	steps := habitStreakWizardSteps(state.Kind)
	active := min(max(0, state.HabitStreakStep-1), len(steps)-1)
	title := "Habit Streaks"
	if momentumMode {
		title = "New Momentum"
		if state.Kind == "edit_momentum" {
			title = "Edit Momentum"
		}
	}
	body := []string{stepStrip(theme, steps, active), ""}
	submit := "ctrl+s save streak"
	if momentumMode {
		submit = "ctrl+s save momentum"
	}
	d := listDialog{wide: true, title: title, context: stepContext(steps, active), focus: -1}
	switch steps[active] {
	case "Kind":
		d.body, d.focus = momentumKindStep(theme, state, body)
		d.primary, d.hints = "enter choose", []string{"↑/↓ move", "ctrl+s continue"}
	case "Targets":
		d.body, d.focus = momentumTargetsStep(theme, state, body)
		d.primary = "ctrl+s details"
		d.hints = []string{"space toggle", "a all", "c none", "←/→ mode", "tab next"}
		if momentumContextMode(state) {
			d.hints = []string{"type filter", "←/→ cycle", "enter add/remove", "tab next"}
		}
	case "Details":
		d.body, d.focus = momentumDetailsStep(theme, state, body, momentumMode)
		d.primary, d.hints = submit, []string{"tab next", "shift+tab back"}
	default:
		d.body = momentumReviewStep(theme, state, body, momentumMode)
		d.primary, d.hints = submit, []string{"shift+tab back"}
	}
	return renderListDialog(theme, state, d)
}

func momentumContextMode(state controllerpkg.State) bool {
	return sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
}

func momentumKindStep(theme Theme, state controllerpkg.State, body []string) ([]string, int) {
	body = append(body, theme.StyleDim.Render("Choose momentum kind"))
	focus := len(body) + state.ChoiceCursor
	body = append(body, selectList(theme, state.ChoiceItems, state.ChoiceCursor)...)
	if state.ChoiceCursor >= 0 && state.ChoiceCursor < len(state.ChoiceDetails) {
		body = append(body, "", theme.StyleDim.Render(state.ChoiceDetails[state.ChoiceCursor]))
	}
	return body, focus
}

func momentumTargetsStep(theme Theme, state controllerpkg.State, body []string) ([]string, int) {
	column := len("Stream")
	full := formInner(state.Width) - column - 3
	matchFocus, focus := 1, -1
	if momentumContextMode(state) {
		matchFocus = 3
		body = append(body, theme.StyleDim.Render("Select contributing contexts"))
		repo := state.MomentumRepoInput
		stream := state.MomentumStreamInput
		body = append(body,
			Picker(theme, PickerSpec{Label: "Repo", Column: column, Focused: state.FocusIdx == 0,
				Input: bareInput(repo, full, "type to filter"), Placeholder: state.RepoSelectorLabel})[0],
			Picker(theme, PickerSpec{Label: "Stream", Column: column, Focused: state.FocusIdx == 1,
				Input: bareInput(stream, full, "any stream"), Placeholder: state.StreamSelectorLabel})[0],
			"", theme.StyleDim.Render("Selected contexts"),
		)
		if state.FocusIdx < 2 {
			focus = len(body) - 4 + state.FocusIdx
		}
		if len(state.HabitStreakDraft.Contexts) == 0 {
			body = append(body, theme.StyleDim.Render("No contexts selected"))
		}
		labels := []string{}
		for _, item := range state.HabitStreakDraft.Contexts {
			labels = append(labels, "[x] "+momentumContextLabel(item, state.MomentumRepos, state.MomentumAllIssues, state.MomentumStreams))
		}
		cursor := -1
		if state.FocusIdx == 2 {
			cursor = state.HabitStreakCursor
			focus = len(body) + cursor
		}
		body = append(body, selectList(theme, labels, cursor)...)
		for _, warning := range momentumContextRedundancyWarnings(state.HabitStreakDraft.Contexts, state.MomentumRepos, state.MomentumAllIssues, state.MomentumStreams) {
			body = append(body, lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(warning))
		}
	} else {
		body = append(body, theme.StyleDim.Render("Select contributing habits"))
		if len(state.HabitItems) == 0 {
			body = append(body, theme.StyleDim.Render("No habits available"))
		}
		labels := []string{}
		for _, item := range state.HabitItems {
			box := "[ ] "
			if slices.Contains(state.HabitStreakDraft.HabitIDs, item.ID) {
				box = "[x] "
			}
			labels = append(labels, fmt.Sprintf("%s%s  %s / %s", box, item.Name, item.RepoName, item.StreamName))
		}
		if state.FocusIdx == 0 {
			focus = len(body) + state.HabitStreakCursor
		}
		body = append(body, selectList(theme, labels, state.HabitStreakCursor)...)
	}
	matchFocused := state.FocusIdx == matchFocus
	if matchFocused {
		focus = len(body) + 1
	}
	mode := cases.Title(language.Und).String(sharedtypes.MomentumMatchModeLabel(state.HabitStreakDraft.MatchMode))
	body = append(body, "", Picker(theme, PickerSpec{Label: "Match", Column: column, Focused: matchFocused, Options: []string{mode}})[0])
	return body, focus
}

func momentumDetailsStep(theme Theme, state controllerpkg.State, body []string, momentumMode bool) ([]string, int) {
	column := len("Description")
	inner := formInner(state.Width)
	full, _ := formColumns(inner, column)
	periodFocus, countFocus := 1, 2
	if momentumMode {
		periodFocus, countFocus = 2, 3
	}
	rows := []FormRow{fieldRow(0, Field(theme, "Name", column, bareInput(state.Inputs[0], full, "required"), state.FocusIdx == 0))}
	if momentumMode {
		rows = append(rows, textareaRow(1, Textarea(theme, "Description", column, bareTextarea(state), state.FocusIdx == 1, inner)))
	}
	current := sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period)
	periods := []string{}
	for _, option := range []sharedtypes.HabitStreakPeriod{sharedtypes.HabitStreakPeriodDay, sharedtypes.HabitStreakPeriodWeek, sharedtypes.HabitStreakPeriodMonth} {
		label := habitStreakPeriodLabel(option)
		if option == current {
			periods = append(periods, theme.StyleSelected.Render(" "+label+" "))
		} else {
			periods = append(periods, theme.StyleDim.Render(" "+label+" "))
		}
	}
	unit := strings.TrimPrefix(habitStreakRequirementLabel(state.HabitStreakDraft), "Required ")
	rows = append(rows,
		fieldRow(periodFocus, Field(theme, "Period", column, strings.Join(periods, " "), state.FocusIdx == periodFocus)),
		fieldRow(countFocus, Field(theme, "Required", column, bareInput(state.Inputs[1], 6, "1")+theme.StyleDim.Render("  "+unit), state.FocusIdx == countFocus)),
	)
	if momentumMode && !momentumContextMode(state) {
		capacity := habitMomentumViewCapacity(state.HabitStreakDraft, state.HabitItems)
		if capacity.Valid {
			rows = append(rows, staticRow(theme.StyleDim.Render(strings.Repeat(" ", column+2)+fmt.Sprintf("Allowed range: 1-%d per %s", capacity.MaxCount, habitStreakPeriodUnitLabel(state.HabitStreakDraft.Period)))))
		} else if strings.TrimSpace(capacity.Reason) != "" {
			rows = append(rows, staticRow(theme.StyleError.Render(capacity.Reason)))
		}
	}
	rows = append(rows, staticRow("",
		Field(theme, "Matching", column, habitStreakMatchTypeSummary(state.HabitStreakDraft), false),
		Field(theme, habitStreakTargetLabel(state.HabitStreakDraft), column, momentumTargetsSummary(state), false),
	))
	lines, focus := formBody(rows, state.FocusIdx)
	if focus >= 0 {
		focus += len(body)
	}
	return append(body, lines...), focus
}

func momentumReviewStep(theme Theme, state controllerpkg.State, body []string, momentumMode bool) []string {
	column := len("Description")
	if momentumMode {
		body = append(body, Field(theme, "Kind", column, momentumTargetKindLabel(state.HabitStreakDraft.TargetKind), false))
	}
	body = append(body, Field(theme, "Name", column, theme.StyleHeader.Render(fallback(state.HabitStreakDraft.Name, "-")), false))
	if state.HabitStreakDraft.Description != nil && strings.TrimSpace(*state.HabitStreakDraft.Description) != "" {
		body = append(body, Field(theme, "Description", column, strings.TrimSpace(*state.HabitStreakDraft.Description), false))
	}
	rule := fmt.Sprintf("%s per %s", habitStreakTargetSummary(state.HabitStreakDraft), habitStreakPeriodUnitLabel(state.HabitStreakDraft.Period))
	return append(body,
		Field(theme, "Rule", column, rule, false),
		Field(theme, habitStreakTargetLabel(state.HabitStreakDraft), column, momentumTargetsSummary(state), false),
	)
}

func momentumTargetsSummary(state controllerpkg.State) string {
	return habitStreakTargetSelectionSummary(state.HabitStreakDraft, state.MomentumRepos, state.MomentumAllIssues, state.MomentumStreams, state.HabitItems)
}
