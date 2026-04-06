package wellbeing

import (
	"fmt"

	types "crona/tui/internal/tui/views/types"
)

func Accountability(size types.ViewSize, theme types.Theme, state types.ContentState) []string {
	switch size {
	case types.ViewSizeCompact:
		return wellbeingCompactAccountabilityLines(theme, state)
	default:
		return wellbeingAccountabilityLines(theme, state)
	}
}

func wellbeingAccountabilityLines(theme types.Theme, state types.ContentState) []string {
	planned, completed, failed, abandoned, pending := dailyPlanCounts(state.DailyPlan)
	score, pressure, delayed, highRisk := dailyPlanSignals(state.DailyPlan)
	if planned == 0 && completed == 0 && failed == 0 && abandoned == 0 && pending == 0 && delayed == 0 && score == 0 {
		return nil
	}
	lines := []string{
		"",
		fmt.Sprintf("%s  planned %d   completed %d   failed %d", theme.StyleHeader.Render("Accountability"), planned, completed, failed),
		theme.StyleDim.Render(fmt.Sprintf("pending rollback %d   abandoned %d", pending, abandoned)),
	}
	if score > 0 || delayed > 0 || highRisk > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("score %.1f   backlog %.1f   delayed %d   high risk %d", score, pressure, delayed, highRisk)))
	}
	for _, item := range recentPlanFailureLines(state.DailyPlan, 2) {
		lines = append(lines, theme.StyleDim.Render(item))
	}
	return lines
}

func wellbeingCompactAccountabilityLines(theme types.Theme, state types.ContentState) []string {
	planned, completed, failed, _, pending := dailyPlanCounts(state.DailyPlan)
	score, _, delayed, _ := dailyPlanSignals(state.DailyPlan)
	if planned == 0 && completed == 0 && failed == 0 && pending == 0 && delayed == 0 && score == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("%s  Planned %d  Completed %d  Failed %d  Pending %d", theme.StyleHeader.Render("Accountability"), planned, completed, failed, pending),
		theme.StyleDim.Render(fmt.Sprintf("Score %.1f  Delayed %d", score, delayed)),
	}
}
