package momentumview

import (
	"fmt"
	"math"
	"strings"

	sharedtypes "crona/shared/types"
	viewchrome "crona/tui/internal/tui/views/chrome"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	momentumhelpers "crona/tui/internal/tui/views/momentum"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func Render(theme types.Theme, state types.ContentState) string {
	active := state.Pane == "momentum_cards"
	topSection := renderMomentumTopSection(theme, state)
	if len(state.MomentumCards) == 0 {
		return renderMomentumEmptyState(theme, state, active, topSection)
	}
	lines := renderMomentumCards(theme, state, active, topSection)
	return viewchrome.RenderPaneBox(theme, active, state.Width, state.Height, strings.Join(lines, "\n"))
}

func momentumHeader(state types.ContentState) string {
	return fmt.Sprintf("Momentum  %s  %dd window", state.MomentumDate, max(1, state.MomentumWindowDays))
}

func renderMomentumTopSection(theme types.Theme, state types.ContentState) []string {
	topSection := []string{theme.StylePaneTitle.Render(momentumHeader(state))}
	if state.Height < 24 {
		return topSection
	}
	return append(
		topSection,
		viewchrome.RenderActionLine(
			theme,
			state.Width-6,
			viewchrome.ContextualActions(
				theme,
				viewchrome.ActionsState{
					View:        state.View,
					Pane:        state.Pane,
					MomentumTab: state.MomentumTab,
				},
			),
		),
		"",
	)
}

func renderMomentumEmptyState(
	theme types.Theme,
	state types.ContentState,
	active bool,
	topSection []string,
) string {
	lines := append([]string{}, topSection...)
	lines = append(
		lines,
		theme.StyleHeader.Render("No momentum definitions"),
		theme.StyleDim.Render("Press [a] to create your first momentum."),
	)
	return viewchrome.RenderPaneBox(
		theme,
		active,
		state.Width,
		state.Height,
		strings.Join(lines, "\n"),
	)
}

func renderMomentumCards(
	theme types.Theme,
	state types.ContentState,
	active bool,
	topSection []string,
) []string {
	cursor := state.Cursors["momentum_cards"]
	lines := append([]string{}, topSection...)
	availableCardLines := viewchrome.RemainingPaneHeight(state.Height, lines)
	mode := chooseMomentumCardMode(theme, state, cursor, active, availableCardLines)
	cardBodies, cardHeights := renderMomentumCardBodies(theme, state, cursor, active, mode)
	if len(cardBodies) == 0 {
		return lines
	}
	windowStart, windowEnd := momentumCardWindow(cursor, cardHeights, availableCardLines, len(cardBodies))
	return appendMomentumCardWindow(lines, theme, cardBodies, windowStart, windowEnd)
}

func renderMomentumCardBodies(
	theme types.Theme,
	state types.ContentState,
	cursor int,
	active bool,
	mode momentumCardMode,
) ([]string, []int) {
	cardBodies := make([]string, 0, len(state.MomentumCards))
	cardHeights := make([]int, 0, len(state.MomentumCards))
	for idx, card := range state.MomentumCards {
		rendered := renderCard(
			theme,
			state,
			card,
			idx == cursor,
			active,
			mode,
		)
		cardBodies = append(cardBodies, rendered)
		cardHeights = append(cardHeights, lipgloss.Height(rendered))
	}
	return cardBodies, cardHeights
}

func momentumCardWindow(cursor int, cardHeights []int, availableCardLines int, totalCards int) (int, int) {
	windowStart, windowEnd := 0, 0
	overflowHints := 0
	for range 4 {
		inner := max(availableCardLines-overflowHints, 1)
		windowStart, windowEnd = visibleMomentumCardWindow(cursor, cardHeights, inner)
		nextOverflowHints := 0
		if windowStart > 0 {
			nextOverflowHints++
		}
		if windowEnd < totalCards {
			nextOverflowHints++
		}
		if nextOverflowHints == overflowHints {
			break
		}
		overflowHints = nextOverflowHints
	}
	return windowStart, windowEnd
}

func appendMomentumCardWindow(
	lines []string,
	theme types.Theme,
	cardBodies []string,
	windowStart int,
	windowEnd int,
) []string {
	if windowStart > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", windowStart)))
	}
	for idx := windowStart; idx < windowEnd; idx++ {
		if idx > windowStart {
			lines = append(lines, "")
		}
		lines = append(lines, cardBodies[idx])
	}
	if remaining := len(cardBodies) - windowEnd; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return lines
}

func renderCard(
	theme types.Theme,
	state types.ContentState,
	card sharedtypes.MomentumCard,
	selected bool,
	active bool,
	mode momentumCardMode,
) string {
	def := sharedtypes.NormalizeHabitStreakDefinition(card.Definition)
	cardWidth, graphWidth := momentumCardWidths(state.Width)
	body := renderCardBody(
		theme,
		def,
		card,
		renderMomentumCardTitle(theme, def, selected, graphWidth, mode),
		graphWidth,
		mode,
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(momentumCardBorderColor(theme, selected, active)).
		Padding(1, 2).
		Width(cardWidth).
		Render(body)
}

func momentumCardWidths(totalWidth int) (int, int) {
	cardWidth := max(28, totalWidth-8)
	innerWidth := max(24, cardWidth-4)
	return cardWidth, max(20, innerWidth-2)
}

func renderMomentumCardTitle(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	selected bool,
	graphWidth int,
	mode momentumCardMode,
) string {
	prefix := "  "
	if selected {
		prefix = viewchrome.SelectionCursor + " "
	}
	title := theme.StyleHeader.Render(viewhelpers.Truncate(prefix+def.Name, graphWidth))
	if mode != momentumCardModeNormal {
		return title
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		"  ",
		momentumCardStatus(theme, def.Enabled),
	)
}

func momentumCardStatus(theme types.Theme, enabled bool) string {
	if enabled {
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("Active")
	}
	return theme.StyleDim.Render("Inactive")
}

func momentumCardBorderColor(theme types.Theme, selected bool, active bool) lipgloss.Color {
	if selected && active {
		return theme.ColorCyan
	}
	if selected {
		return theme.ColorSubtle
	}
	return theme.ColorDim
}

type momentumCardMode int

const (
	momentumCardModeNormal momentumCardMode = iota
	momentumCardModeCompact
	momentumCardModeUltraCompact
)

func chooseMomentumCardMode(
	theme types.Theme,
	state types.ContentState,
	cursor int,
	active bool,
	availableLines int,
) momentumCardMode {
	if len(state.MomentumCards) == 0 {
		return momentumCardModeNormal
	}

	cursor = max(0, min(cursor, len(state.MomentumCards)-1))

	selectedCard := state.MomentumCards[cursor]
	for _, mode := range []momentumCardMode{
		momentumCardModeNormal,
		momentumCardModeCompact,
		momentumCardModeUltraCompact,
	} {
		rendered := renderCard(
			theme,
			state,
			selectedCard,
			true,
			active,
			mode,
		)
		if lipgloss.Height(rendered) <= availableLines {
			return mode
		}
	}
	return momentumCardModeUltraCompact
}

func renderCardBody(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	card sharedtypes.MomentumCard,
	titleRow string,
	graphWidth int,
	mode momentumCardMode,
) string {
	meta := renderMomentumCardMeta(theme, def, card, graphWidth)
	habits := renderMomentumCardHabits(theme, card.HabitNames, graphWidth)
	timeline := renderMomentumSeries(theme, def.Period, card.Series, graphWidth, def.Enabled)
	footnote := theme.StyleDim.Render(viewhelpers.Truncate(momentumSeriesFootnote(card.Series, def.Enabled), graphWidth))
	return renderMomentumCardLayout(
		titleRow,
		renderMomentumCardDescription(theme, def.Description, graphWidth),
		meta,
		habits,
		timeline,
		footnote,
		mode,
	)
}

func renderMomentumCardMeta(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	card sharedtypes.MomentumCard,
	graphWidth int,
) string {
	return theme.StyleNormal.Render(
		viewhelpers.Truncate(
			fmt.Sprintf(
				"%s · target %d/%s · current streak %s · best %s",
				momentumhelpers.CadenceLabel(def.Period),
				def.RequiredCount,
				momentumBucketUnit(def.Period),
				momentumhelpers.FormatLength(card.Current, momentumhelpers.Unit(def.Period)),
				momentumhelpers.FormatLength(card.Longest, momentumhelpers.Unit(def.Period)),
			),
			graphWidth,
		),
	)
}

func renderMomentumCardDescription(theme types.Theme, description *string, graphWidth int) string {
	if description == nil {
		return ""
	}
	text := strings.TrimSpace(*description)
	if text == "" {
		return ""
	}
	return theme.StyleDim.Render(viewhelpers.Truncate(text, graphWidth))
}

func renderMomentumCardHabits(theme types.Theme, habitNames []string, graphWidth int) string {
	return theme.StyleDim.Render(
		viewhelpers.Truncate("Habits: "+momentumHabitSummary(habitNames), graphWidth),
	)
}

func renderMomentumCardLayout(
	titleRow string,
	descriptionRow string,
	meta string,
	habits string,
	timeline string,
	footnote string,
	mode momentumCardMode,
) string {
	switch mode {
	case momentumCardModeCompact:
		rows := []string{titleRow}
		if descriptionRow != "" {
			rows = append(rows, descriptionRow)
		}
		rows = append(rows, meta, habits, timeline, footnote)
		return strings.Join(rows, "\n")
	case momentumCardModeUltraCompact:
		return strings.Join([]string{titleRow, meta, timeline, footnote}, "\n")
	default:
		rows := []string{titleRow}
		if descriptionRow != "" {
			rows = append(rows, descriptionRow, "")
		} else {
			rows = append(rows, "")
		}
		rows = append(rows, meta, "", habits, "", timeline, footnote)
		return strings.Join(rows, "\n")
	}
}

func renderMomentumSeries(
	theme types.Theme,
	period sharedtypes.HabitStreakPeriod,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) string {
	if sharedtypes.NormalizeHabitStreakPeriod(period) == sharedtypes.HabitStreakPeriodDay {
		return renderMomentumDailySquares(theme, series, width, enabled)
	}
	return renderMomentumBucketTimeline(theme, series, width, enabled)
}

func visibleMomentumCardWindow(cursor int, heights []int, inner int) (int, int) {
	total := len(heights)
	if total == 0 {
		return 0, 0
	}

	inner = max(1, inner)
	cursor = max(0, min(cursor, total-1))

	start := cursor
	end := cursor + 1
	used := max(1, heights[cursor])
	for {
		canAbove := start > 0 && used+1+max(1, heights[start-1]) <= inner
		canBelow := end < total && used+1+max(1, heights[end]) <= inner
		if !canAbove && !canBelow {
			break
		}
		aboveRemaining := start
		belowRemaining := total - end
		if canAbove && (!canBelow || aboveRemaining >= belowRemaining) {
			start--
			used += 1 + max(1, heights[start])
			continue
		}
		if canBelow {
			used += 1 + max(1, heights[end])
			end++
			continue
		}
		break
	}
	return start, end
}

func momentumHabitSummary(names []string) string {
	if len(names) == 0 {
		return "No habits linked"
	}
	if len(names) <= 3 {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s, %s, %s +%d", names[0], names[1], names[2], len(names)-3)
}

func momentumSeriesFootnote(series []sharedtypes.MomentumSeriesPoint, enabled bool) string {
	if !enabled {
		return "Momentum disabled"
	}
	if len(series) == 0 {
		return ""
	}
	last := series[len(series)-1]
	status := "missed"
	if last.MetTarget {
		status = "met"
	}
	return fmt.Sprintf("Latest bucket: %d/%d %s", last.Count, last.Target, status)
}

func renderMomentumBucketTimeline(
	theme types.Theme,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) string {
	if len(series) == 0 {
		return theme.StyleDim.Render("no data")
	}
	if width < 1 {
		return theme.StyleDim.Render("no data")
	}

	labelWidth := max(10, min(16, width/4))
	ratioWidth := 7
	statusWidth := 7
	spacerWidth := 4
	barWidth := width - labelWidth - ratioWidth - statusWidth - spacerWidth
	if barWidth < 8 {
		barWidth = max(4, width-ratioWidth-statusWidth-spacerWidth-6)
		labelWidth = max(6, width-ratioWidth-statusWidth-spacerWidth-barWidth)
	}

	labelWidth = max(6, labelWidth)
	barWidth = max(4, barWidth)

	rows := make([]string, 0, len(series))
	for _, point := range series {
		rows = append(
			rows,
			renderMomentumBucketRow(
				theme,
				point,
				labelWidth,
				ratioWidth,
				statusWidth,
				barWidth,
				enabled,
			),
		)
	}
	return strings.Join(rows, "\n")
}

func renderMomentumDailySquares(
	theme types.Theme,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) string {
	if len(series) == 0 || width < 1 {
		return theme.StyleDim.Render("no data")
	}
	rows := momentumDailySquareRows(theme, series, width, enabled)
	if len(rows) == 0 {
		return theme.StyleDim.Render("no data")
	}
	return strings.Join(rows, "\n")
}

func momentumDailySquareRows(
	theme types.Theme,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) []string {
	rows := make([]string, 0, len(series)/max(1, width)+2)
	row := ""
	for _, point := range series {
		cell := momentumDailySquareCell(theme, point, enabled)
		cellWidth := lipgloss.Width(cell)
		if cellWidth <= 0 {
			cellWidth = 1
		}
		if row != "" && lipgloss.Width(row)+1+cellWidth > width {
			rows = append(rows, row)
			row = ""
		}
		if row != "" {
			row += " "
		}
		row += cell
	}
	if strings.TrimSpace(row) != "" {
		rows = append(rows, row)
	}
	return rows
}

func momentumDailySquareCell(theme types.Theme, point sharedtypes.MomentumSeriesPoint, enabled bool) string {
	if !enabled {
		if point.Count > 0 || point.MetTarget {
			return theme.StyleDim.Render("■")
		}
		return theme.StyleDim.Render("□")
	}
	if point.Count > 0 || point.MetTarget {
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("■")
	}
	return lipgloss.NewStyle().Foreground(theme.ColorRed).Render("□")
}

func renderMomentumBucketRow(
	theme types.Theme,
	point sharedtypes.MomentumSeriesPoint,
	labelWidth, ratioWidth, statusWidth, barWidth int,
	enabled bool,
) string {
	label := padRight(
		viewhelpers.Truncate(momentumDisplayLabel(point), labelWidth),
		labelWidth,
	)
	ratioText := padRight(fmt.Sprintf("%d/%d", point.Count, point.Target), ratioWidth)
	status := momentumStatusForPoint(point, enabled)
	statusText := momentumStatusStyle(theme, status).Render(padRight(status, statusWidth))
	bar := momentumBucketBar(theme, point, barWidth, enabled)
	return strings.Join([]string{label, ratioText, statusText, bar}, "  ")
}

func momentumDisplayLabel(point sharedtypes.MomentumSeriesPoint) string {
	label := strings.TrimSpace(point.Label)
	if label == "" {
		return label
	}
	if strings.HasPrefix(label, "[") {
		return label
	}
	if !strings.Contains(point.BucketKey, "-W") {
		return label
	}
	if _, week, err := momentumWeekNumber(point.BucketKey); err == nil {
		return fmt.Sprintf("[%d] %s", week, label)
	}
	return label
}

func momentumWeekNumber(bucketKey string) (int, int, error) {
	var year, week int
	if _, err := fmt.Sscanf(bucketKey, "%4d-W%2d", &year, &week); err != nil {
		return 0, 0, err
	}
	return year, week, nil
}

func momentumBucketBar(theme types.Theme, point sharedtypes.MomentumSeriesPoint, width int, enabled bool) string {
	width = max(1, width)
	target := max(point.Target, 1)
	status := momentumStatusForPoint(point, enabled)
	naturalWidth := max(target, point.Count)
	renderWidth := min(width, max(1, naturalWidth))
	filled := min(point.Count, renderWidth)
	markerPos := 0
	if point.Target > 0 {
		markerPos = min(target-1, renderWidth-1)
	}
	var builder strings.Builder
	fillStyle := momentumStatusStyle(theme, status)
	if !enabled {
		fillStyle = theme.StyleDim
	}
	markerStyle := lipgloss.NewStyle().Foreground(theme.ColorWhite).Bold(true)
	if !enabled {
		markerStyle = theme.StyleDim
	}
	for idx := range renderWidth {
		switch {
		case idx == markerPos && point.Target > 0:
			builder.WriteString(markerStyle.Render("┆"))
		case idx < filled:
			builder.WriteString(fillStyle.Render("█"))
		default:
			builder.WriteString("░")
		}
	}
	return builder.String()
}

func momentumStatusForPoint(point sharedtypes.MomentumSeriesPoint, enabled bool) string {
	if !enabled {
		return "paused"
	}
	if point.Count >= point.Target {
		return "met"
	}
	if point.Target > 0 && float64(point.Count) >= math.Ceil(float64(point.Target)*0.75) {
		return "near"
	}
	return "missed"
}

func momentumStatusStyle(theme types.Theme, status string) lipgloss.Style {
	switch status {
	case "met":
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true)
	case "near":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "paused":
		return theme.StyleDim
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
}

func momentumBucketUnit(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "week"
	case sharedtypes.HabitStreakPeriodMonth:
		return "month"
	default:
		return "day"
	}
}

func padRight(value string, width int) string {
	if width <= 0 {
		return ""
	}
	valueWidth := lipgloss.Width(value)
	if valueWidth >= width {
		return viewhelpers.Truncate(value, width)
	}
	return value + strings.Repeat(" ", width-valueWidth)
}
