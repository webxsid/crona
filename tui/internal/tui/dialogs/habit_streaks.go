package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	sharedutils "crona/shared/utils"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
	if sharedtypes.NormalizeMomentumTargetKind(
		def.TargetKind,
	) == sharedtypes.MomentumTargetKindContext {
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
		if sharedtypes.NormalizeMomentumMatchMode(
			def.MatchMode,
		) == sharedtypes.MomentumMatchModeAll {
			return fmt.Sprintf(
				"%s · %d contexts, %s each",
				mode,
				max(1, len(def.Contexts)),
				helperpkg.FormatCompactDurationSeconds(max(1, def.RequiredCount)),
			)
		}
		return fmt.Sprintf(
			"%s · %s work",
			mode,
			helperpkg.FormatCompactDurationSeconds(max(1, def.RequiredCount)),
		)
	default:
		if sharedtypes.NormalizeMomentumMatchMode(
			def.MatchMode,
		) == sharedtypes.MomentumMatchModeAll {
			return fmt.Sprintf(
				"%s · %d habits, %d each",
				mode,
				max(1, len(def.HabitIDs)),
				max(1, def.RequiredCount),
			)
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
			return sharedutils.HabitMomentumCapacity{
				Reason: "selected habits are no longer available",
			}
		}
		selected = append(selected, habit)
	}
	return sharedutils.HabitMomentumCapacityForSelection(selected, def.Period, def.MatchMode)
}
