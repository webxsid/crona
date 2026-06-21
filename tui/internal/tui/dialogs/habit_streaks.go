package dialogs

import (
	"fmt"
	"slices"
	"strings"

	sharedtypes "crona/shared/types"
	sharedutils "crona/shared/utils"
	"crona/tui/internal/api"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	helperpkg "crona/tui/internal/tui/helpers"
	viewchrome "crona/tui/internal/tui/views/chrome"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func renderHabitStreakDialog(theme Theme, state controllerpkg.State) string {
	momentumMode := isMomentumDialogKind(state.Kind)
	steps := habitStreakWizardSteps(state.Kind)
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
		if state.Kind == "create_momentum" {
			rows = renderHabitStreakKindSelection(theme, state, rows)
			rows = appendDialogFooter(
				theme,
				state,
				rows,
				"[↑/↓] move   [enter] choose   [ctrl+s] continue   [esc] cancel",
			)
		} else {
			rows = renderHabitStreakTargets(theme, state, rows, momentumMode)
			rows = appendDialogFooter(
				theme,
				state,
				rows,
				targetSelectionFooter(state)+"   [esc] cancel",
			)
		}
	case 2:
		if state.Kind == "create_momentum" {
			rows = renderHabitStreakTargets(theme, state, rows, momentumMode)
			rows = appendDialogFooter(
				theme,
				state,
				rows,
				targetSelectionFooter(state)+"   [esc] back",
			)
		} else {
			rows = renderHabitStreakDetails(theme, state, rows, momentumMode)
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
	case 3:
		if state.Kind == "create_momentum" {
			rows = renderHabitStreakDetails(theme, state, rows, momentumMode)
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
		} else {
			rows = renderHabitStreakReview(theme, state, rows, momentumMode)
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
	case 4:
		rows = renderHabitStreakReview(theme, state, rows, momentumMode)
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

func habitStreakWizardSteps(kind string) []string {
	if kind == "create_momentum" {
		return []string{"Kind", "Targets", "Details", "Review"}
	}
	return []string{"Targets", "Details", "Review"}
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

func renderHabitStreakKindSelection(
	theme Theme,
	state controllerpkg.State,
	rows []string,
) []string {
	rows = append(rows, theme.StyleDim.Render("Choose momentum kind"))
	for i, item := range state.ChoiceItems {
		detail := ""
		if i < len(state.ChoiceDetails) {
			detail = state.ChoiceDetails[i]
		}
		line := item
		if detail != "" {
			line = fmt.Sprintf("%s  ·  %s", item, detail)
		}
		if i == state.ChoiceCursor {
			rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
		} else {
			rows = append(rows, theme.StyleNormal.Render("  "+line))
		}
	}
	return rows
}

func renderHabitStreakTargets(
	theme Theme,
	state controllerpkg.State,
	rows []string,
	momentumMode bool,
) []string {
	contextMode := sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
	if contextMode {
		rows = append(rows, theme.StyleDim.Render("Select contributing contexts"))
		rows = append(rows, renderMomentumContextColumns(theme, state, state.Width, 92))
		rows = append(rows, "")
		rows = append(rows, theme.StyleDim.Render("Selected contexts"))
		if len(state.HabitStreakDraft.Contexts) == 0 {
			rows = append(rows, theme.StyleDim.Render("No contexts selected"))
		}
		for i, contextItem := range state.HabitStreakDraft.Contexts {
			line := "[x] " + momentumContextLabel(
				contextItem,
				state.MomentumRepos,
				state.MomentumAllIssues,
				state.MomentumStreams,
			)
			if state.FocusIdx == 2 && i == state.HabitStreakCursor {
				rows = append(rows, theme.StyleCursor.Render(viewchrome.SelectionCursor+" "+line))
			} else {
				rows = append(rows, theme.StyleNormal.Render("  "+line))
			}
		}
		for _, warning := range momentumContextRedundancyWarnings(
			state.HabitStreakDraft.Contexts,
			state.MomentumRepos,
			state.MomentumAllIssues,
			state.MomentumStreams,
		) {
			rows = append(rows, lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(warning))
		}
		rows = append(rows, "")
		rows = append(rows, renderMomentumMatchMode(theme, state, state.FocusIdx == 3))
		return rows
	}
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
	rows = append(rows, "")
	rows = append(rows, renderMomentumMatchMode(theme, state, state.FocusIdx == 1))
	_ = momentumMode
	return rows
}

func renderHabitStreakDetails(
	theme Theme,
	state controllerpkg.State,
	rows []string,
	momentumMode bool,
) []string {
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
		habitStreakRowLabel(theme, state, countRowIdx, habitStreakRequirementLabel(state.HabitStreakDraft)),
		dialogInputView(state, 1),
	)
	if momentumMode &&
		sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
			sharedtypes.MomentumTargetKindHabit {
		capacity := habitMomentumViewCapacity(state.HabitStreakDraft, state.HabitItems)
		rows = append(rows, "")
		if capacity.Valid {
			rows = append(
				rows,
				theme.StyleDim.Render(
					fmt.Sprintf(
						"Allowed range: 1-%d per %s",
						capacity.MaxCount,
						habitStreakPeriodUnitLabel(state.HabitStreakDraft.Period),
					),
				),
			)
		} else if strings.TrimSpace(capacity.Reason) != "" {
			rows = append(rows, theme.StyleError.Render(capacity.Reason))
		}
	}
	rows = append(
		rows,
		"",
		theme.StyleDim.Render(habitStreakMatchTypeLabel(state.HabitStreakDraft)),
		theme.StyleHeader.Render(habitStreakMatchTypeSummary(state.HabitStreakDraft)),
		"",
		theme.StyleDim.Render(habitStreakTargetLabel(state.HabitStreakDraft)),
		theme.StyleHeader.Render(
			habitStreakTargetSelectionSummary(
				state.HabitStreakDraft,
				state.MomentumRepos,
				state.MomentumAllIssues,
				state.MomentumStreams,
				state.HabitItems,
			),
		),
	)
	return rows
}

func renderHabitStreakReview(
	theme Theme,
	state controllerpkg.State,
	rows []string,
	momentumMode bool,
) []string {
	description := ""
	if state.HabitStreakDraft.Description != nil {
		description = strings.TrimSpace(*state.HabitStreakDraft.Description)
	}
	if momentumMode {
		rows = append(
			rows,
			theme.StyleDim.Render("Kind"),
			theme.StyleHeader.Render(momentumTargetKindLabel(state.HabitStreakDraft.TargetKind)),
			"",
		)
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
				"%s per %s",
				habitStreakTargetSummary(state.HabitStreakDraft),
				strings.ToLower(habitStreakPeriodLabel(state.HabitStreakDraft.Period)),
			),
		),
		"",
		theme.StyleDim.Render(habitStreakTargetLabel(state.HabitStreakDraft)),
		theme.StyleHeader.Render(
			habitStreakTargetSelectionSummary(
				state.HabitStreakDraft,
				state.MomentumRepos,
				state.MomentumAllIssues,
				state.MomentumStreams,
				state.HabitItems,
			),
		),
	)
	return rows
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

func momentumTargetKindLabel(kind sharedtypes.MomentumTargetKind) string {
	switch sharedtypes.NormalizeMomentumTargetKind(kind) {
	case sharedtypes.MomentumTargetKindContext:
		return "Contexts"
	default:
		return "Habits"
	}
}

func habitStreakTargetLabel(def sharedtypes.HabitStreakDefinition) string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		return "Contexts"
	default:
		return "Habits"
	}
}

func habitStreakMatchTypeLabel(def sharedtypes.HabitStreakDefinition) string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		return "Matching contexts"
	default:
		return "Matching habits"
	}
}

func habitStreakMatchTypeSummary(def sharedtypes.HabitStreakDefinition) string {
	kind := "habits"
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) == sharedtypes.MomentumTargetKindContext {
		kind = "contexts"
	}
	switch sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) {
	case sharedtypes.MomentumMatchModeAll:
		return fmt.Sprintf("All of the selected %s", kind)
	default:
		return fmt.Sprintf("Any of the selected %s", kind)
	}
}

func habitStreakRequirementLabel(def sharedtypes.HabitStreakDefinition) string {
	period := habitStreakPeriodUnitLabel(def.Period)
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		return fmt.Sprintf("Required work per %s", period)
	default:
		return fmt.Sprintf("Required completions per %s", period)
	}
}

func habitStreakPeriodUnitLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "week"
	case sharedtypes.HabitStreakPeriodMonth:
		return "month"
	default:
		return "day"
	}
}

func habitStreakTargetSummary(def sharedtypes.HabitStreakDefinition) string {
	mode := cases.Title(language.Und).String(sharedtypes.MomentumMatchModeLabel(def.MatchMode))
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		if sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) == sharedtypes.MomentumMatchModeAll {
			return fmt.Sprintf("%s · %d contexts, %s each", mode, max(1, len(def.Contexts)), helperpkg.FormatCompactDurationSeconds(max(1, def.RequiredCount)))
		}
		return fmt.Sprintf("%s · %s work", mode, helperpkg.FormatCompactDurationSeconds(max(1, def.RequiredCount)))
	default:
		if sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) == sharedtypes.MomentumMatchModeAll {
			return fmt.Sprintf("%s · %d habits, %d each", mode, max(1, len(def.HabitIDs)), max(1, def.RequiredCount))
		}
		return fmt.Sprintf("%s · %d completions", mode, max(1, def.RequiredCount))
	}
}

func habitStreakTargetSelectionSummary(
	def sharedtypes.HabitStreakDefinition,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	habits []sharedtypes.HabitWithMeta,
) string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		if len(def.Contexts) == 0 {
			return "None"
		}
		labels := make([]string, 0, len(def.Contexts))
		for _, contextItem := range def.Contexts {
			labels = append(labels, momentumContextLabel(contextItem, repos, allIssues, streams))
		}
		if len(labels) <= 3 {
			return strings.Join(labels, ", ")
		}
		return fmt.Sprintf("%s, %s, %s +%d", labels[0], labels[1], labels[2], len(labels)-3)
	default:
		return habitStreakHabitSummary(def.HabitIDs, habits)
	}
}

func targetSelectionFooter(state controllerpkg.State) string {
	contextMode := sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
	if contextMode {
		return "[type] filter   [←/→] cycle   [enter] add/remove   [tab] next   [ctrl+s] details"
	}
	return "[space] toggle   [a] all   [c] none   [←/→] mode   [tab] next   [ctrl+s] details"
}

func renderMomentumMatchMode(
	theme Theme,
	state controllerpkg.State,
	active bool,
) string {
	label := "Match mode"
	if active {
		label = viewchrome.SelectionCursor + " " + label
	}
	return theme.StyleDim.Render(label) + "\n" +
		renderSelector(
			theme,
			state,
			cases.Title(language.Und).String(sharedtypes.MomentumMatchModeLabel(state.HabitStreakDraft.MatchMode)),
			active,
		)
}

func renderTargetFieldLabel(theme Theme, active bool, label string) string {
	if active {
		return theme.StyleCursor.Render(label)
	}
	return theme.StyleDim.Render(label)
}

func renderMomentumContextColumns(
	theme Theme,
	state controllerpkg.State,
	width, maxWidth int,
) string {
	contentWidth := min(width-8, maxWidth) - 8
	if contentWidth < 28 {
		contentWidth = 28
	}
	colWidth := (contentWidth - 2) / 2
	if colWidth < 12 {
		colWidth = 12
	}
	repoCol := lipgloss.NewStyle().Width(colWidth).Render(strings.Join([]string{
		renderTargetFieldLabel(theme, state.FocusIdx == 0, "Repo"),
		state.MomentumRepoInput.View(),
		"",
		renderSelector(theme, state, state.RepoSelectorLabel, state.FocusIdx == 0),
	}, "\n"))
	streamCol := lipgloss.NewStyle().Width(colWidth).Render(strings.Join([]string{
		renderTargetFieldLabel(theme, state.FocusIdx == 1, "Stream (Optional)"),
		state.MomentumStreamInput.View(),
		"",
		renderSelector(theme, state, state.StreamSelectorLabel, state.FocusIdx == 1),
	}, "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, repoCol, "  ", streamCol)
}

func momentumContextLabel(
	contextItem sharedtypes.MomentumContext,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
) string {
	repoLabel := ""
	for _, repo := range repos {
		if repo.ID == contextItem.RepoID {
			repoLabel = strings.TrimSpace(repo.Name)
			break
		}
	}
	if repoLabel == "" {
		repoLabel = fmt.Sprintf("%d", contextItem.RepoID)
	}
	if contextItem.StreamID == nil {
		return repoLabel + " / Any stream"
	}
	streamID := *contextItem.StreamID
	streamLabel := ""
	for _, stream := range streams {
		if stream.ID == streamID {
			streamLabel = strings.TrimSpace(stream.Name)
			break
		}
	}
	if streamLabel == "" {
		for _, issue := range allIssues {
			if issue.RepoID == contextItem.RepoID && issue.StreamID == streamID {
				streamLabel = strings.TrimSpace(issue.StreamName)
				if streamLabel != "" {
					break
				}
			}
		}
	}
	if streamLabel == "" {
		streamLabel = fmt.Sprintf("%d", streamID)
	}
	return repoLabel + " / " + streamLabel
}

func momentumContextRedundancyWarnings(
	contexts []sharedtypes.MomentumContext,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
) []string {
	redundancies := sharedtypes.MomentumContextRedundancies(contexts)
	if len(redundancies) == 0 {
		return nil
	}
	out := make([]string, 0, len(redundancies))
	for _, redundancy := range redundancies {
		repoWideLabel := momentumContextLabel(redundancy.RepoWideContext, repos, allIssues, streams)
		redundantLabels := make([]string, 0, len(redundancy.RedundantContexts))
		for _, contextItem := range redundancy.RedundantContexts {
			redundantLabels = append(
				redundantLabels,
				momentumContextLabel(contextItem, repos, allIssues, streams),
			)
		}
		label := ""
		switch len(redundantLabels) {
		case 0:
			continue
		case 1:
			label = redundantLabels[0]
		case 2:
			label = redundantLabels[0] + " and " + redundantLabels[1]
		default:
			label = fmt.Sprintf("%d specific stream selections", len(redundantLabels))
		}
		out = append(out, fmt.Sprintf("Warning: %s already covers %s", repoWideLabel, label))
	}
	return out
}

func habitMomentumViewCapacity(
	def sharedtypes.HabitStreakDefinition,
	habits []sharedtypes.HabitWithMeta,
) sharedutils.HabitMomentumCapacity {
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
