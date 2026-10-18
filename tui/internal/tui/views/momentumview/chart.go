package momentumview

import (
	"fmt"
	"math"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	helperpkg "crona/tui/internal/tui/helpers"
	viewhelpers "crona/tui/internal/tui/views/helpers"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

func momentumCardTargetLabel(def sharedtypes.HabitStreakDefinition) string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		return "Contexts"
	case sharedtypes.MomentumTargetKindRepo:
		return "Repo"
	case sharedtypes.MomentumTargetKindStream:
		return "Stream"
	default:
		return "Habits"
	}
}

func renderMomentumSeries(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) string {
	if sharedtypes.NormalizeHabitStreakPeriod(def.Period) == sharedtypes.HabitStreakPeriodDay {
		if momentumDailySeriesUsesSquares(def, series) {
			return renderMomentumDailySquares(theme, series, width, enabled)
		}
		return renderMomentumDailyDistribution(theme, def, series, width, enabled)
	}
	return renderMomentumBucketTimeline(theme, def, series, width, enabled)
}

func momentumDailySeriesUsesSquares(
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
) bool {
	if len(series) == 0 {
		return false
	}
	if sharedtypes.NormalizeMomentumTargetKind(
		def.TargetKind,
	) == sharedtypes.MomentumTargetKindContext {
		return false
	}
	if max(1, def.RequiredCount) > 1 {
		return false
	}
	switch sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) {
	case sharedtypes.MomentumMatchModeAll:
		return true
	default:
		return len(def.HabitIDs) <= 1
	}
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

func momentumSeriesFootnote(
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	enabled bool,
) string {
	if !enabled {
		return "Momentum disabled"
	}
	if len(series) == 0 {
		return ""
	}
	last := series[len(series)-1]
	if last.Skipped {
		return fmt.Sprintf(
			"Latest bucket: skipped (%d protected days)",
			last.ProtectedDayCount,
		)
	}
	status := "missed"
	if last.MetTarget {
		status = "met"
	}
	parts := []string{
		fmt.Sprintf(
			"Latest bucket: %s/%s %s",
			momentumValueDisplay(last.Count, def),
			momentumValueDisplay(last.Target, def),
			status,
		),
	}
	if last.OriginalTarget > 0 && last.OriginalTarget != last.Target {
		parts = append(
			parts,
			fmt.Sprintf("original %s", momentumValueDisplay(last.OriginalTarget, def)),
		)
	}
	if last.ProtectedDayCount > 0 {
		parts = append(parts, fmt.Sprintf("%d protected days", last.ProtectedDayCount))
	}
	return strings.Join(parts, " · ")
}

func renderMomentumBucketTimeline(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
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

	scaleMax := momentumSeriesScale(series)
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
				def,
				point,
				labelWidth,
				ratioWidth,
				statusWidth,
				barWidth,
				scaleMax,
				enabled,
			),
		)
	}
	rowSeparator := "\n"
	return strings.Join(rows, rowSeparator)
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

type momentumChartCell struct {
	bar      bool
	barColor lipgloss.Color
}

func renderMomentumDailyDistribution(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) string {
	rows := momentumDailyDistributionRows(theme, def, series, width, enabled)
	if len(rows) == 0 {
		return theme.StyleDim.Render("no data")
	}
	return strings.Join(rows, "\n")
}

func momentumDailyDistributionRows(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
) []string {
	if len(series) == 0 || width < 1 {
		return nil
	}

	chartHeight := 7
	seriesMax := momentumSeriesScale(series)
	seriesMax = max(seriesMax, 1)
	yLabelWidth := 5
	axisMax := seriesMax
	tickRows, tickLabels := momentumChartTicks(seriesMax, chartHeight, def)
	if sharedtypes.NormalizeMomentumTargetKind(
		def.TargetKind,
	) == sharedtypes.MomentumTargetKindContext {
		ctxAxis := momentumContextAxis(seriesMax)
		axisMax = ctxAxis.MaxAxis
		yLabelWidth = ctxAxis.LabelWidth
		tickRows, tickLabels = momentumContextTickRows(
			axisMax,
			chartHeight,
			ctxAxis.Step,
			ctxAxis.Format,
		)
	}
	chartWidth := max(12, width-yLabelWidth-2)

	grid := make([][]momentumChartCell, chartHeight)
	for y := range grid {
		grid[y] = make([]momentumChartCell, chartWidth)
	}

	sharedTarget, sharedTargetOK := momentumSharedTarget(series)
	xPositions := momentumXPositions(len(series), chartWidth)
	for i, point := range series {
		x := xPositions[i]
		barStatus := momentumStatusForPoint(point, enabled)
		barPalette := momentumVerticalPalette(theme, def, barStatus, enabled)
		barRamp := viewhelpers.GradientRamp(barPalette.Start, barPalette.End, chartHeight)
		barRow := momentumChartValueRow(point.Count, axisMax, chartHeight)

		for y := barRow; y < chartHeight; y++ {
			grid[y][x].bar = true
			grid[y][x].barColor = viewhelpers.GradientColorAt(barRamp, y)
		}
	}

	targetRow := -1
	if sharedTargetOK {
		targetRow = momentumChartValueRow(sharedTarget, axisMax, chartHeight)
	}
	lines := make([]string, 0, chartHeight+2)
	lineStyle := lipgloss.NewStyle().Foreground(theme.ColorYellow).Bold(true)
	axisStyle := lipgloss.NewStyle().Foreground(theme.ColorDim)
	for row := range chartHeight {
		label := strings.Repeat(" ", yLabelWidth)
		if tick, ok := tickLabels[row]; ok {
			label = fmt.Sprintf("%*s", yLabelWidth, tick)
		}
		var b strings.Builder
		for col := range chartWidth {
			cell := grid[row][col]
			switch {
			case row == targetRow:
				b.WriteString(lineStyle.Render("─"))
			case cell.bar:
				barStyle := lipgloss.NewStyle().Foreground(cell.barColor)
				if !enabled {
					barStyle = theme.StyleDim
				}
				b.WriteString(barStyle.Render("█"))
			case tickRows[row]:
				b.WriteString(axisStyle.Render("┈"))
			default:
				b.WriteRune(' ')
			}
		}
		lines = append(lines, fmt.Sprintf("%s │%s", theme.StyleDim.Render(label), b.String()))
	}
	lines = append(
		lines,
		fmt.Sprintf("%s └%s", strings.Repeat(" ", yLabelWidth), strings.Repeat("─", chartWidth)),
	)
	lines = append(lines, momentumChartXAxis(theme, series, yLabelWidth+2, chartWidth, xPositions))
	return lines
}

type momentumContextAxisSpec struct {
	MaxAxis    int
	Step       int
	LabelWidth int
	Format     func(int) string
}

func momentumContextAxis(maxAxis int) momentumContextAxisSpec {
	step := 15 * 60
	labelWidth := 3
	format := func(value int) string {
		return fmt.Sprintf("%02dm", value/60)
	}
	if maxAxis >= int(time.Hour.Seconds()) {
		step = 30 * 60
		labelWidth = 6
		format = func(value int) string {
			hours := value / 3600
			minutes := (value % 3600) / 60
			return fmt.Sprintf("%02dh%02dm", hours, minutes)
		}
	}
	maxAxis = momentumRoundUpAxis(maxAxis, step)

	maxAxis = max(maxAxis, step)

	return momentumContextAxisSpec{
		MaxAxis:    maxAxis,
		Step:       step,
		LabelWidth: labelWidth,
		Format:     format,
	}
}

func momentumRoundUpAxis(value, step int) int {
	if step <= 0 {
		return value
	}
	if value <= 0 {
		return step
	}
	remainder := value % step
	if remainder == 0 {
		return value
	}
	return value + step - remainder
}

func momentumContextTickRows(
	maxAxis, chartHeight, step int,
	format func(int) string,
) (map[int]bool, map[int]string) {
	rows := map[int]bool{}
	labels := map[int]string{}
	seenRows := map[int]bool{}
	seenLabels := map[string]bool{}
	for value := 0; value <= maxAxis; value += step {
		row := momentumChartValueRow(value, maxAxis, chartHeight)
		label := format(value)
		if seenRows[row] || seenLabels[label] {
			continue
		}
		seenRows[row] = true
		seenLabels[label] = true
		rows[row] = true
		labels[row] = label
	}
	return rows, labels
}

func momentumSharedTarget(series []sharedtypes.MomentumSeriesPoint) (int, bool) {
	if len(series) == 0 {
		return 0, false
	}
	target := series[0].Target
	for _, point := range series[1:] {
		if point.Target != target || point.Skipped {
			return 0, false
		}
	}
	return target, true
}

func momentumVerticalPalette(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	status string,
	enabled bool,
) viewhelpers.GradientBarPalette {
	if !enabled {
		return viewhelpers.GradientBarPalette{
			Start: theme.ColorDim,
			End:   theme.ColorDim,
			Track: theme.ColorDim,
		}
	}
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		switch status {
		case "met":
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullGreen,
				End:   theme.ColorCyan,
				Track: theme.ColorDim,
			}
		case "near":
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullGreen,
				End:   theme.ColorYellow,
				Track: theme.ColorDim,
			}
		default:
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullRed,
				End:   theme.ColorOrange,
				Track: theme.ColorDim,
			}
		}
	default:
		switch status {
		case "met":
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullGreen,
				End:   theme.ColorGreen,
				Track: theme.ColorDim,
			}
		case "near":
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullGreen,
				End:   theme.ColorYellow,
				Track: theme.ColorDim,
			}
		default:
			return viewhelpers.GradientBarPalette{
				Start: theme.ColorDullRed,
				End:   theme.ColorOrange,
				Track: theme.ColorDim,
			}
		}
	}
}

func momentumXPositions(points int, chartWidth int) []int {
	out := make([]int, points)
	if points == 1 {
		out[0] = min(chartWidth-1, chartWidth/2)
		return out
	}
	for i := range points {
		out[i] = int(math.Round(float64(i) * float64(chartWidth-1) / float64(points-1)))
	}
	return out
}

func momentumChartValueRow(value, maxAxis, chartHeight int) int {
	if maxAxis <= 0 {
		return chartHeight - 1
	}
	row := chartHeight - 1 - int(math.Round(float64(value)/float64(maxAxis)*float64(chartHeight-1)))
	if row < 0 {
		return 0
	}
	if row >= chartHeight {
		return chartHeight - 1
	}
	return row
}

func momentumChartTicks(
	maxAxis, chartHeight int,
	def sharedtypes.HabitStreakDefinition,
) (map[int]bool, map[int]string) {
	rows := map[int]bool{}
	labels := map[int]string{}
	steps := min(chartHeight, 5)
	steps = max(steps, 2)
	seen := map[int]bool{}
	seenLabels := map[string]bool{}
	for i := 0; i < steps; i++ {
		value := float64(maxAxis) * (1 - float64(i)/float64(steps-1))
		row := int(math.Round(float64(i) * float64(chartHeight-1) / float64(steps-1)))
		row = min(chartHeight-1, max(0, row))
		if seen[row] {
			continue
		}
		label := momentumValueDisplay(int(math.Round(value)), def)
		if seenLabels[label] {
			continue
		}
		seen[row] = true
		seenLabels[label] = true
		rows[row] = true
		labels[row] = label
	}
	return rows, labels
}

func momentumChartXAxis(
	theme types.Theme,
	series []sharedtypes.MomentumSeriesPoint,
	leftPad int,
	chartWidth int,
	xPositions []int,
) string {
	if len(series) == 0 {
		return ""
	}
	runes := make([]rune, chartWidth)
	for i := range runes {
		runes[i] = ' '
	}
	samples := []int{0}
	if len(series) > 1 {
		samples = append(samples, len(series)/3, (2*len(series))/3, len(series)-1)
	}
	seen := map[int]bool{}
	for _, idx := range samples {
		if idx < 0 || idx >= len(series) || seen[idx] {
			continue
		}
		seen[idx] = true
		label := viewhelpers.Truncate(
			momentumDisplayLabel(series[idx]),
			max(4, min(chartWidth, 12)),
		)
		x := xPositions[idx]
		start := max(0, min(x-lipgloss.Width(label)/2, chartWidth-lipgloss.Width(label)))
		for i, r := range label {
			if start+i >= 0 && start+i < len(runes) {
				runes[start+i] = r
			}
		}
	}
	return strings.Repeat(" ", leftPad) + theme.StyleDim.Render(string(runes))
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

func momentumDailySquareCell(
	theme types.Theme,
	point sharedtypes.MomentumSeriesPoint,
	enabled bool,
) string {
	if !enabled {
		if point.Count > 0 || point.MetTarget {
			return theme.StyleDim.Render("■")
		}
		return theme.StyleDim.Render("□")
	}
	if point.Count > 0 || point.MetTarget {
		return lipgloss.NewStyle().Foreground(theme.ColorDullGreen).Bold(true).Render("■")
	}
	return lipgloss.NewStyle().Foreground(theme.ColorRed).Render("□")
}

func renderMomentumBucketRow(
	theme types.Theme,
	def sharedtypes.HabitStreakDefinition,
	point sharedtypes.MomentumSeriesPoint,
	labelWidth, ratioWidth, statusWidth, barWidth, scaleMax int,
	enabled bool,
) string {
	label := padRight(
		viewhelpers.Truncate(momentumDisplayLabel(point), labelWidth),
		labelWidth,
	)
	ratioText := padRight(
		momentumRatioText(point, def),
		ratioWidth,
	)
	status := momentumStatusForPoint(point, enabled)
	statusText := momentumStatusStyle(theme, status).Render(padRight(status, statusWidth))
	bar := momentumBucketBar(theme, point, barWidth, enabled, scaleMax)
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

func momentumBucketBar(
	theme types.Theme,
	point sharedtypes.MomentumSeriesPoint,
	width int,
	enabled bool,
	scaleMax int,
) string {
	width = max(1, width)
	scaleMax = max(1, scaleMax)
	status := momentumStatusForPoint(point, enabled)
	if point.Skipped {
		return theme.StyleDim.Render(strings.Repeat("·", width))
	}
	renderWidth := width
	filled := int(math.Round(float64(point.Count) / float64(scaleMax) * float64(renderWidth)))
	filled = min(max(0, filled), renderWidth)

	markerPos := 0
	if point.Target > 0 {
		markerPos = int(
			math.Round(float64(point.Target) / float64(scaleMax) * float64(renderWidth-1)),
		)
		markerPos = min(max(0, markerPos), renderWidth-1)
	}
	palette := momentumBarPalette(theme, status, enabled)
	return viewhelpers.RenderGradientBarWithMarker(renderWidth, filled, markerPos, palette, "┆")
}

func momentumBarPalette(
	theme types.Theme,
	status string,
	enabled bool,
) viewhelpers.GradientBarPalette {
	if !enabled {
		return viewhelpers.GradientBarPalette{
			Start: theme.ColorDim,
			End:   theme.ColorDim,
			Track: theme.ColorDim,
		}
	}
	switch status {
	case "met":
		return viewhelpers.GradientBarPalette{
			Start: theme.ColorDullGreen,
			End:   theme.ColorGreen,
			Track: theme.ColorDim,
		}
	case "near":
		return viewhelpers.GradientBarPalette{
			Start: theme.ColorDullGreen,
			End:   theme.ColorYellow,
			Track: theme.ColorDim,
		}
	default:
		return viewhelpers.GradientBarPalette{
			Start: theme.ColorDullRed,
			End:   theme.ColorOrange,
			Track: theme.ColorDim,
		}
	}
}

func momentumSeriesScale(series []sharedtypes.MomentumSeriesPoint) int {
	scaleMax := 1
	for _, point := range series {
		scaleMax = max(scaleMax, point.Count, point.Target)
	}
	return scaleMax
}

func momentumStatusForPoint(point sharedtypes.MomentumSeriesPoint, enabled bool) string {
	if !enabled {
		return "paused"
	}
	if point.Skipped {
		return "skipped"
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
		return lipgloss.NewStyle().Foreground(theme.ColorDullGreen).Bold(true)
	case "near":
		return lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "paused":
		return theme.StyleDim
	case "skipped":
		return theme.StyleDim
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorDullRed)
	}
}

func momentumRatioText(
	point sharedtypes.MomentumSeriesPoint,
	def sharedtypes.HabitStreakDefinition,
) string {
	if point.Skipped {
		return "skip"
	}
	return fmt.Sprintf(
		"%s/%s",
		momentumValueDisplay(point.Count, def),
		momentumValueDisplay(point.Target, def),
	)
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

func momentumTargetSummary(def sharedtypes.HabitStreakDefinition) string {
	mode := momentumModeLabel(def.MatchMode)
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
			helperpkg.FormatCompactDurationSeconds(def.RequiredCount),
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
		return fmt.Sprintf(
			"%s · %d/%s",
			mode,
			max(1, def.RequiredCount),
			momentumBucketUnit(def.Period),
		)
	}
}

func momentumModeLabel(mode sharedtypes.MomentumMatchMode) string {
	value := sharedtypes.MomentumMatchModeLabel(mode)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func momentumTargetSummaryNames(card sharedtypes.MomentumCard) string {
	if len(card.TargetNames) > 0 {
		return momentumHabitSummary(card.TargetNames)
	}
	if len(card.HabitNames) > 0 {
		return momentumHabitSummary(card.HabitNames)
	}
	return "Not set"
}

func momentumValueDisplay(value int, def sharedtypes.HabitStreakDefinition) string {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		return helperpkg.FormatCompactDurationSeconds(value)
	default:
		return fmt.Sprintf("%d", value)
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
