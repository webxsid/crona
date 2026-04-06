package wellbeing

import (
	"fmt"

	types "crona/tui/internal/tui/views/types"
)

func RiskSnapshot(size types.ViewSize, theme types.Theme, state types.ContentState) []string {
	switch size {
	case types.ViewSizeCompact:
		return wellbeingCompactRiskSnapshotLines(theme, state)
	default:
		return wellbeingRiskSnapshotLines(theme, state)
	}
}

func wellbeingRiskSnapshotLines(theme types.Theme, state types.ContentState) []string {
	burnout := latestBurnout(state)
	if burnout == nil && state.DailyPlan == nil {
		return nil
	}
	score, pressure, delayed, _ := dailyPlanSignals(state.DailyPlan)
	lines := []string{"", theme.StyleHeader.Render("Risk Snapshot")}
	if burnout != nil {
		lines = append(lines, fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)))
		lines = append(lines, theme.StyleDim.Render(burnoutSummary(burnout)))
	}
	lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("accountability %.1f   backlog %.1f   delayed %d", score, pressure, delayed)))
	return lines
}

func wellbeingCompactRiskSnapshotLines(theme types.Theme, state types.ContentState) []string {
	burnout := latestBurnout(state)
	if burnout == nil {
		return nil
	}
	return []string{
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Burnout"), burnoutBadge(theme, burnout)),
		theme.StyleDim.Render(burnoutSummary(burnout)),
	}
}
