package daily

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	viewmomentum "crona/tui/internal/tui/views/momentum"
	types "crona/tui/internal/tui/views/types"
)

func renderMomentumBlock(theme types.Theme, state types.ContentState, width, height int) []string {
	sections := make([][]string, 0, 2)
	if height >= 57 {
		if signals := renderSignalsBlock(theme, state, width); len(signals) > 0 {
			sections = append(sections, signals)
		}
	}
	if momentums := renderMomentumsBlock(theme, state, width); len(momentums) > 0 {
		sections = append(sections, momentums)
	}
	if len(sections) == 0 {
		return nil
	}

	lines := make([]string, 0, len(sections)*4)
	for idx, section := range sections {
		if idx > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, section...)
	}
	return lines
}

func renderSignalsBlock(theme types.Theme, state types.ContentState, width int) []string {
	rows := signalRows(theme, state)
	if len(rows) == 0 {
		return nil
	}

	lines := []string{theme.StyleHeader.Render("Signals")}
	lines = append(lines, pairSignalRows(rows, width)...)
	return lines
}

func renderMomentumsBlock(theme types.Theme, state types.ContentState, width int) []string {
	if state.DailyStreaks == nil {
		return nil
	}

	lines := []string{theme.StyleHeader.Render("Momentums")}
	lines = append(lines, viewmomentum.CompactComparison(
		theme,
		"Check-ins",
		sharedtypes.HabitStreakPeriodDay,
		state.DailyStreaks.CurrentCheckInDays,
		state.DailyStreaks.LongestCheckInDays,
		"Focus",
		sharedtypes.HabitStreakPeriodDay,
		state.DailyStreaks.CurrentFocusDays,
		state.DailyStreaks.LongestFocusDays,
		width,
	))
	return lines
}

func pairSignalRows(rows []string, width int) []string {
	switch len(rows) {
	case 0:
		return nil
	case 1:
		return []string{rows[0]}
	case 2:
		return []string{renderInlinePairRow(rows[0], rows[1], width)}
	case 3:
		return []string{renderInlineMultiRow(rows, width)}
	default:
		lines := []string{renderInlinePairRow(rows[0], rows[1], width)}
		if len(rows) > 3 {
			lines = append(lines, renderInlinePairRow(rows[2], rows[3], width))
		} else if len(rows) > 2 {
			lines = append(lines, rows[2])
		}
		return lines
	}
}

func renderInlinePairRow(left, right string, width int) string {
	return viewhelpers.Truncate(fmt.Sprintf("%s  |  %s", left, right), width)
}

func signalRows(theme types.Theme, state types.ContentState) []string {
	rows := make([]string, 0, 3)
	if energy := signalEnergyValue(state); energy != "" {
		rows = append(
			rows,
			fmt.Sprintf(
				"%s  %s",
				theme.StyleHeader.Render("Energy"),
				theme.StyleNormal.Render(energy),
			),
		)
	}
	if mood := signalMoodValue(state); mood != "" {
		rows = append(
			rows,
			fmt.Sprintf(
				"%s  %s",
				theme.StyleHeader.Render("Mood"),
				theme.StyleNormal.Render(mood),
			),
		)
	}
	if sleep := signalSleepValue(state); sleep != "" {
		rows = append(
			rows,
			fmt.Sprintf(
				"%s  %s",
				theme.StyleHeader.Render("Sleep"),
				theme.StyleNormal.Render(sleep),
			),
		)
	}
	return rows
}

func renderInlineMultiRow(rows []string, width int) string {
	return viewhelpers.Truncate(strings.Join(rows, "  |  "), width)
}

func signalMoodValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" {
		return fmt.Sprintf("today %d/5", state.DailyCheckIn.Mood)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageMood != nil {
		return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageMood)
	}
	return ""
}

func signalEnergyValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" {
		return fmt.Sprintf("today %d/5", state.DailyCheckIn.Energy)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageEnergy != nil {
		return fmt.Sprintf("avg %.1f/5", *state.MetricsRollup.AverageEnergy)
	}
	return ""
}

func signalSleepValue(state types.ContentState) string {
	if state.DailyCheckIn != nil && state.DailyCheckIn.Date != "" && state.DailyCheckIn.SleepHours != nil {
		return "today " + helperpkg.FormatCompactDurationHours(*state.DailyCheckIn.SleepHours)
	}
	if state.MetricsRollup != nil && state.MetricsRollup.AverageSleepHours != nil {
		return "avg " + helperpkg.FormatCompactDurationHours(*state.MetricsRollup.AverageSleepHours)
	}
	return ""
}
