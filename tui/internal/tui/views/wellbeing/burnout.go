package wellbeing

import (
	"fmt"
	"math"
	"sort"
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
	sort.Slice(positive, func(i, j int) bool { return positive[i].score > positive[j].score })
	sort.Slice(negative, func(i, j int) bool { return negative[i].score < negative[j].score })
	riskLimit := min(3, len(positive))
	recoveryLimit := min(2, len(negative))
	risks = make([]string, 0, riskLimit)
	recoveries = make([]string, 0, recoveryLimit)
	for i := 0; i < riskLimit; i++ {
		risks = append(risks, fmt.Sprintf("- %s: +%d", prettifyBurnoutFactor(positive[i].name), int(math.Round(positive[i].score*100))))
	}
	for i := 0; i < recoveryLimit; i++ {
		recoveries = append(recoveries, fmt.Sprintf("- %s: -%d", prettifyBurnoutFactor(negative[i].name), int(math.Round(math.Abs(negative[i].score)*100))))
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
