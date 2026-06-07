package momentumview

import (
	"fmt"
	"math"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
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

func renderMomentumFocusTab(
	theme types.Theme,
	state types.ContentState,
	topSection []string,
) []string {
	lines := append([]string{}, topSection...)
	available := viewchrome.RemainingPaneHeight(state.Height, lines)
	graphLines, historyLines := momentumFocusSections(theme, state, max(1, state.Width-8), available)
	lines = append(lines, graphLines...)
	if len(graphLines) > 0 && len(historyLines) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines, historyLines...)
	return lines
}

func renderMomentumWellbeingPlaceholder(
	theme types.Theme,
	state types.ContentState,
	topSection []string,
) []string {
	lines := append([]string{}, topSection...)
	lines = append(
		lines,
		theme.StyleHeader.Render("Wellbeing Momentum"),
		"",
		theme.StyleDim.Render("This tab is reserved for the shared wellbeing momentum surface."),
		theme.StyleDim.Render(fmt.Sprintf("Window  %s  %dd days", state.MomentumDate, max(1, state.MomentumWindowDays))),
	)
	return lines
}

func momentumFocusSections(
	theme types.Theme,
	state types.ContentState,
	width int,
	available int,
) ([]string, []string) {
	window := state.MomentumMetricsRange
	if len(window) == 0 {
		return []string{theme.StyleDim.Render("No focus data for this window.")}, nil
	}
	graphHeight := 11
	if available >= 30 {
		graphHeight = 13
	}
	if available < 20 {
		graphHeight = 9
	}
	graph := renderMomentumFocusGraph(theme, window, width, graphHeight)
	historyHeader := []string{
		theme.StyleHeader.Render("Daily History"),
		theme.StyleDim.Render("estimated vs logged by day"),
	}
	bodyHeight := max(1, available-len(graph)-len(historyHeader)-1)
	history := renderMomentumFocusHistory(theme, window, state.MomentumHistoryY, width, bodyHeight)
	return graph, append(historyHeader, history...)
}

func renderMomentumFocusGraph(
	theme types.Theme,
	days []api.DailyMetricsDay,
	width int,
	height int,
) []string {
	if len(days) == 0 || width < 24 || height < 7 {
		return []string{theme.StyleDim.Render("No focus data")}
	}
	yLabelWidth := 4
	chartWidth := max(12, width-yLabelWidth-2)
	chartHeight := max(4, height-4)
	estimated := make([]float64, 0, len(days))
	actual := make([]float64, 0, len(days))
	maxValue := 1.0
	for _, day := range days {
		est := float64(day.TotalEstimatedMinutes) / 60.0
		act := float64(day.WorkedSeconds) / 3600.0
		estimated = append(estimated, est)
		actual = append(actual, act)
		maxValue = math.Max(maxValue, math.Max(est, act))
	}
	maxAxis := math.Ceil(maxValue)
	cells := make([][]styledRune, chartHeight)
	for y := range cells {
		cells[y] = make([]styledRune, chartWidth)
		for x := range cells[y] {
			cells[y][x] = styledRune{r: ' '}
		}
	}
	drawMomentumSeries(cells, estimated, maxAxis, theme.ColorYellow, '•')
	drawMomentumSeries(cells, actual, maxAxis, theme.ColorGreen, '•')
	lines := []string{
		theme.StyleHeader.Render("Focus Momentum"),
		theme.StyleDim.Render("estimated and logged hours by day"),
		renderMomentumFocusLegend(theme),
	}
	for row := 0; row < chartHeight; row++ {
		value := maxAxis * (1 - float64(row)/float64(max(1, chartHeight-1)))
		label := fmt.Sprintf("%3.0fh", value)
		var b strings.Builder
		for _, cell := range cells[row] {
			if !cell.set {
				b.WriteRune(cell.r)
				continue
			}
			b.WriteString(cell.style.Render(string(cell.r)))
		}
		lines = append(lines, fmt.Sprintf("%s │%s", theme.StyleDim.Render(label), b.String()))
	}
	lines = append(lines, fmt.Sprintf("%s └%s", strings.Repeat(" ", yLabelWidth), strings.Repeat("─", chartWidth)))
	lines = append(lines, renderMomentumFocusXAxis(theme, days, yLabelWidth+2, chartWidth))
	return lines
}

type styledRune struct {
	r     rune
	style lipgloss.Style
	set   bool
}

func drawMomentumSeries(
	grid [][]styledRune,
	values []float64,
	maxValue float64,
	color lipgloss.Color,
	marker rune,
) {
	if len(grid) == 0 || len(grid[0]) == 0 || len(values) == 0 {
		return
	}
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	prevX, prevY := -1, -1
	width := len(grid[0])
	height := len(grid)
	for i, value := range values {
		x := 0
		if len(values) > 1 {
			x = int(math.Round(float64(i) * float64(width-1) / float64(len(values)-1)))
		}
		y := height - 1 - int(math.Round(clamp01(value/maxValue)*float64(height-1)))
		if prevX >= 0 {
			drawMomentumLine(grid, prevX, prevY, x, y, style, marker)
		}
		setMomentumCell(grid, x, y, style, marker)
		prevX, prevY = x, y
	}
}

func drawMomentumLine(
	grid [][]styledRune,
	x0, y0, x1, y1 int,
	style lipgloss.Style,
	marker rune,
) {
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		setMomentumCell(grid, x0, y0, style, marker)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func setMomentumCell(grid [][]styledRune, x, y int, style lipgloss.Style, marker rune) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
		return
	}
	cell := grid[y][x]
	if cell.r != ' ' && cell.r != marker {
		grid[y][x] = styledRune{
			r:     '◆',
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true),
			set:   true,
		}
		return
	}
	if cell.r != ' ' {
		grid[y][x] = styledRune{
			r:     '◆',
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true),
			set:   true,
		}
		return
	}
	grid[y][x] = styledRune{r: marker, style: style, set: true}
}

func renderMomentumFocusLegend(theme types.Theme) string {
	estimate := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true).Render("•")
	actual := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("•")
	return fmt.Sprintf("%s estimated  %s logged", estimate, actual)
}

func renderMomentumFocusXAxis(
	theme types.Theme,
	days []api.DailyMetricsDay,
	leftPad int,
	chartWidth int,
) string {
	if len(days) == 0 {
		return ""
	}
	runes := make([]rune, chartWidth)
	for i := range runes {
		runes[i] = ' '
	}
	samples := []int{0}
	if len(days) > 1 {
		samples = append(samples, len(days)/3, (2*len(days))/3, len(days)-1)
	}
	seen := map[int]bool{}
	for _, idx := range samples {
		if idx < 0 || idx >= len(days) || seen[idx] {
			continue
		}
		seen[idx] = true
		label := momentumShortDate(days[idx].Date)
		x := 0
		if len(days) > 1 {
			x = int(math.Round(float64(idx) * float64(chartWidth-1) / float64(len(days)-1)))
		}
		start := max(0, min(x-len(label)/2, chartWidth-len(label)))
		for i, r := range label {
			if start+i >= 0 && start+i < len(runes) {
				runes[start+i] = r
			}
		}
	}
	return strings.Repeat(" ", leftPad) + theme.StyleDim.Render(string(runes))
}

func renderMomentumFocusHistory(
	theme types.Theme,
	days []api.DailyMetricsDay,
	cursor int,
	width int,
	height int,
) []string {
	if len(days) == 0 || height < 1 {
		return []string{theme.StyleDim.Render("No history")}
	}
	rows := momentumFocusHistoryRows(days)
	start, end := visibleMomentumHistoryWindow(cursor, len(rows), height)
	lines := make([]string, 0, end-start+2)
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		prefix := "  "
		style := theme.StyleNormal
		if i == cursor {
			prefix = viewchrome.SelectionCursor + " "
			style = theme.StyleCursor
		}
		lines = append(lines, style.Render(viewhelpers.Truncate(prefix+rows[i], width)))
	}
	if remaining := len(rows) - end; remaining > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", remaining)))
	}
	return lines
}

func momentumFocusHistoryRows(days []api.DailyMetricsDay) []string {
	rows := make([]string, 0, len(days))
	for _, day := range days {
		estimate := float64(day.TotalEstimatedMinutes) / 60.0
		actual := float64(day.WorkedSeconds) / 3600.0
		delta := actual - estimate
		status := "on target"
		switch {
		case actual == 0 && estimate == 0:
			status = "empty"
		case delta > 0.01:
			status = "over"
		case delta < -0.01:
			status = "under"
		}
		rows = append(rows, fmt.Sprintf(
			"%s   est %.1fh   logged %.1fh   %s%.1fh   %s",
			momentumLongDate(day.Date),
			estimate,
			actual,
			map[bool]string{true: "+", false: ""}[delta >= 0],
			delta,
			status,
		))
	}
	return rows
}

func visibleMomentumHistoryWindow(cursor, total, inner int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	inner = max(1, inner)
	start, end := viewchrome.ListWindow(cursor, total, inner)
	for {
		used := end - start
		if start > 0 {
			used++
		}
		if end < total {
			used++
		}
		if used <= inner || end-start <= 1 {
			return start, end
		}
		if end < total {
			end--
			continue
		}
		if start > 0 {
			start++
			continue
		}
		return start, end
	}
}

func momentumShortDate(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("Jan 2")
}

func momentumLongDate(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("Mon Jan 2")
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
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
	naturalWidth := target
	naturalWidth = max(target, point.Count)
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
