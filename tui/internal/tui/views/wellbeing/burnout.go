package wellbeing

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
	"github.com/charmbracelet/lipgloss"
)

func latestBurnout(state types.ContentState) *api.BurnoutIndicator {
	if state.MetricsRollup == nil || state.MetricsRollup.LatestBurnout == nil {
		return nil
	}
	return state.MetricsRollup.LatestBurnout
}

func burnoutBadge(theme types.Theme, burnout *api.BurnoutIndicator) string {
	style := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	switch burnout.Level {
	case "guarded":
		style = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "high":
		style = lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
	return style.Render(fmt.Sprintf("%d %s", burnout.Score, strings.ToUpper(string(burnout.Level))))
}

func burnoutSummary(burnout *api.BurnoutIndicator) string {
	switch burnout.Level {
	case "high":
		return "Risk is elevated. Reduce load and prioritize recovery today."
	case "guarded":
		return "Signals are mixed. Keep scope tight and protect breaks."
	default:
		return "Current signals look stable. Maintain your recovery rhythm."
	}
}

func burnoutContributorLines(burnout *api.BurnoutIndicator) (risks []string, recoveries []string) {
	type factor struct {
		name  string
		score float64
	}
	positive := make([]factor, 0, len(burnout.Factors))
	negative := make([]factor, 0, len(burnout.Factors))
	for name, score := range burnout.Factors {
		if score >= 0 {
			positive = append(positive, factor{name: name, score: score})
			continue
		}
		negative = append(negative, factor{name: name, score: score})
	}
	slices.SortFunc(positive, func(left, right factor) int {
		return cmp.Compare(right.score, left.score)
	})
	slices.SortFunc(negative, func(left, right factor) int {
		return cmp.Compare(left.score, right.score)
	})
	riskLimit := min(3, len(positive))
	recoveryLimit := min(2, len(negative))
	risks = make([]string, 0, riskLimit)
	recoveries = make([]string, 0, recoveryLimit)
	for i := range riskLimit {
		risks = append(
			risks,
			fmt.Sprintf(
				"- %s: +%d",
				prettifyBurnoutFactor(positive[i].name),
				int(math.Round(positive[i].score*100)),
			),
		)
	}
	for i := range recoveryLimit {
		recoveries = append(
			recoveries,
			fmt.Sprintf(
				"- %s: -%d",
				prettifyBurnoutFactor(negative[i].name),
				int(math.Round(math.Abs(negative[i].score)*100)),
			),
		)
	}
	return risks, recoveries
}

func prettifyBurnoutFactor(name string) string {
	switch name {
	case "workloadPressure":
		return "Workload pressure"
	case "breakDebt":
		return "Break debt"
	case "moodEnergyDrag":
		return "Mood/energy drag"
	case "sleepDebt":
		return "Sleep debt"
	case "recoveryConsistency":
		return "Recovery consistency"
	case "recoveryBreaks":
		return "Break recovery"
	case "loadStability":
		return "Load stability"
	default:
		return name
	}
}
