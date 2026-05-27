package momentum

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func Row(
	theme types.Theme,
	name, cadence string,
	period sharedtypes.HabitStreakPeriod,
	current, longest, width int,
) []string {
	label := viewhelpers.Truncate(name, max(8, width-10))
	if cadence != "" {
		label = fmt.Sprintf("%s  %s", label, theme.StyleDim.Render("["+cadence+"]"))
	}
	return []string{
		theme.StyleHeader.Render(label),
		fmt.Sprintf(
			"%s  %s current · %s best",
			Ladder(theme, period, current, longest),
			FormatLength(current, Unit(period)),
			FormatLength(longest, Unit(period)),
		),
	}
}

func CompactRow(
	theme types.Theme,
	name, cadence string,
	period sharedtypes.HabitStreakPeriod,
	current, longest, width int,
) string {
	label := name
	if cadence != "" {
		label = fmt.Sprintf("%s [%s]", label, cadence)
	}
	return viewhelpers.Truncate(
		fmt.Sprintf(
			"%s  %s  %s",
			label,
			Ladder(theme, period, current, longest),
			FormatLength(current, Unit(period)),
		),
		max(12, width-6),
	)
}

func CompactComparison(
	theme types.Theme,
	leftName string,
	leftPeriod sharedtypes.HabitStreakPeriod,
	leftCurrent, leftLongest int,
	rightName string,
	rightPeriod sharedtypes.HabitStreakPeriod,
	rightCurrent, rightLongest, width int,
) string {
	left := fmt.Sprintf(
		"%s %s %s current · %s best",
		leftName,
		Ladder(theme, leftPeriod, leftCurrent, leftLongest),
		FormatLength(leftCurrent, Unit(leftPeriod)),
		FormatLength(leftLongest, Unit(leftPeriod)),
	)
	right := fmt.Sprintf(
		"%s %s %s current · %s best",
		rightName,
		Ladder(theme, rightPeriod, rightCurrent, rightLongest),
		FormatLength(rightCurrent, Unit(rightPeriod)),
		FormatLength(rightLongest, Unit(rightPeriod)),
	)
	return viewhelpers.Truncate(fmt.Sprintf("%s  |  %s", left, right), width)
}

func Ladder(
	theme types.Theme,
	period sharedtypes.HabitStreakPeriod,
	current, longest int,
) string {
	thresholds := Thresholds(period)
	filled := TierCount(period, current)
	ladder := strings.Repeat("▰", filled) + strings.Repeat("▱", len(thresholds)-filled)
	style := lipgloss.NewStyle().Foreground(theme.ColorDim)
	switch {
	case current > 0 && current == longest:
		style = lipgloss.NewStyle().Foreground(theme.ColorCyan)
	case filled >= 4:
		style = lipgloss.NewStyle().Foreground(theme.ColorGreen)
	case filled > 0:
		style = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	}
	return style.Render(ladder)
}

func TierCount(period sharedtypes.HabitStreakPeriod, current int) int {
	if current <= 0 {
		return 0
	}
	filled := 0
	for _, threshold := range Thresholds(period) {
		if current < threshold {
			break
		}
		filled++
	}
	return filled
}

func Thresholds(period sharedtypes.HabitStreakPeriod) []int {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return []int{1, 2, 4, 8, 13, 26, 52}
	case sharedtypes.HabitStreakPeriodMonth:
		return []int{1, 2, 3, 6, 12, 24}
	default:
		return []int{1, 3, 7, 14, 30, 60, 100}
	}
}

func Unit(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "w"
	case sharedtypes.HabitStreakPeriodMonth:
		return "mo"
	default:
		return "d"
	}
}

func FormatLength(value int, unit string) string {
	return fmt.Sprintf("%d%s", value, unit)
}

func CadenceLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "weekly"
	case sharedtypes.HabitStreakPeriodMonth:
		return "monthly"
	default:
		return "daily"
	}
}
