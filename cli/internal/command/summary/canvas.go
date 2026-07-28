package summary

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	sharedtypes "crona/shared/types"

	"github.com/charmbracelet/x/term"
)

type renderOptions struct {
	Width   int
	Color   bool
	Unicode bool
}

func summaryRenderOptions(w io.Writer) renderOptions {
	width := 100
	color := false
	unicode := false
	if columns, err := strconv.Atoi(strings.TrimSpace(os.Getenv("COLUMNS"))); err == nil && columns >= 40 {
		width = columns
	}
	if file, ok := w.(*os.File); ok && term.IsTerminal(file.Fd()) {
		unicode = true
		color = os.Getenv("NO_COLOR") == ""
		if terminalWidth, _, err := term.GetSize(file.Fd()); err == nil && terminalWidth >= 40 {
			width = terminalWidth
		}
	}
	return renderOptions{Width: max(40, width), Color: color, Unicode: unicode}
}

func renderDayCanvas(w io.Writer, data daySummaryData, opts renderOptions) error {
	width := opts.Width
	sections := []string{
		renderHero(opts, width, "DAILY SUMMARY", data.Date, data.Context, todayTimer(data.Date, data.Timer)),
		renderDayKPIs(opts, width, data),
	}
	if width >= 112 {
		leftWidth := width * 3 / 5
		agenda := renderDayAgenda(opts, leftWidth-4, data)
		signals := renderDaySignals(opts, width-leftWidth-4, data)
		sections = append(sections, joinBlocks(agendaBlock(opts, leftWidth, agenda), signalBlock(opts, width-leftWidth, signals)))
	} else {
		agenda := renderDayAgenda(opts, width-4, data)
		signals := renderDaySignals(opts, width-4, data)
		sections = append(sections, agendaBlock(opts, width, agenda), signalBlock(opts, width, signals))
	}
	_, err := fmt.Fprintln(w, strings.Join(sections, "\n"))
	return err
}

func renderRangeCanvas(w io.Writer, data rangeSummaryData, opts renderOptions) error {
	width := opts.Width
	period := data.Start + "  ->  " + data.End
	sections := []string{
		renderHero(opts, width, "SUMMARY", period, data.Context, nil),
		renderRangeKPIs(opts, width, data),
		renderRangeProgress(opts, width, data),
		renderRangeDays(opts, width, data),
	}
	_, err := fmt.Fprintln(w, strings.Join(sections, "\n"))
	return err
}

func renderHero(opts renderOptions, width int, title, period string, ctx *sharedtypes.ActiveContext, timer *sharedtypes.TimerState) string {
	lines := []string{
		paint(opts, ansiBold+ansiCyan, title) + "  " + paint(opts, ansiBold, period),
		paint(opts, ansiDim, summaryScope(ctx)),
	}
	if timer != nil && timer.State != "idle" && timer.State != "ready" {
		lines = append(lines, paint(opts, ansiYellow, strings.ToUpper(timer.State)+"  "+formatTimerCompact(timer)))
	}
	return box(opts, "", lines, width)
}

func renderDayKPIs(opts renderOptions, width int, data daySummaryData) string {
	total, resolved, worked, estimate := 0, 0, 0, 0
	if data.Summary != nil {
		total = data.Summary.TotalIssues
		resolved = data.Summary.CompletedIssues + data.Summary.AbandonedIssues
		worked = data.Summary.WorkedSeconds
		estimate = data.Summary.TotalEstimatedMinutes * 60
	}
	habitDone, habitFailed := cliHabitCounts(data.Habits)
	score := 0
	if data.Plan != nil {
		score = int(data.Plan.Summary.AccountabilityScore + .5)
	}
	cards := []cliKPI{
		{Label: "FOCUS", Value: formatSeconds(worked), Current: worked, Total: estimate, Color: ansiCyan},
		{Label: "ISSUES", Value: fmt.Sprintf("%d/%d resolved", resolved, total), Current: resolved, Total: total, Color: ansiGreen},
		{Label: "HABITS", Value: fmt.Sprintf("%d/%d done", habitDone, len(data.Habits)), Current: habitDone, Failed: habitFailed, Total: len(data.Habits), Color: ansiGreen},
		{Label: "PLAN", Value: fmt.Sprintf("%d%% accountable", score), Current: score, Total: 100, Color: ansiMagenta},
	}
	return renderKPIBox(opts, width, cards)
}

func renderRangeKPIs(opts renderOptions, width int, data rangeSummaryData) string {
	rollup := data.Rollup
	window := data.Window
	worked, sessions, focusDays, totalDays := 0, 0, 0, rangeDays(data.Start, data.End)
	if rollup != nil {
		worked, sessions, focusDays = rollup.WorkedSeconds, rollup.SessionCount, rollup.FocusDays
	}
	completed, planned, score := 0, 0, 0
	if window != nil {
		completed, planned = window.CompletedCount, window.PlannedCount
		score = int(window.AccountabilityScore + .5)
	}
	cards := []cliKPI{
		{Label: "FOCUS", Value: formatSeconds(worked), Current: focusDays, Total: totalDays, Color: ansiCyan},
		{Label: "SESSIONS", Value: strconv.Itoa(sessions), Current: focusDays, Total: totalDays, Color: ansiBlue},
		{Label: "PLAN", Value: fmt.Sprintf("%d/%d done", completed, planned), Current: completed, Total: planned, Color: ansiGreen},
		{Label: "SCORE", Value: fmt.Sprintf("%d%% accountable", score), Current: score, Total: 100, Color: ansiMagenta},
	}
	return renderKPIBox(opts, width, cards)
}

type cliKPI struct {
	Label   string
	Value   string
	Current int
	Failed  int
	Total   int
	Color   string
}

func renderKPIBox(opts renderOptions, width int, cards []cliKPI) string {
	inner := width - 4
	if width < 76 {
		lines := make([]string, 0, len(cards))
		for _, card := range cards {
			prefix := fmt.Sprintf("%-9s %-20s", card.Label, card.Value)
			barWidth := max(5, inner-visibleWidth(prefix)-1)
			lines = append(lines, paint(opts, ansiBold, fmt.Sprintf("%-9s", card.Label))+" "+
				fmt.Sprintf("%-20s", card.Value)+" "+progressBar(opts, card.Current, card.Failed, card.Total, barWidth, card.Color))
		}
		return box(opts, "AT A GLANCE", lines, width)
	}
	cellWidth := inner / len(cards)
	top, middle, bars := "", "", ""
	for idx, card := range cards {
		extra := 0
		if idx == len(cards)-1 {
			extra = inner - cellWidth*len(cards)
		}
		currentWidth := cellWidth + extra
		top += padRight(paint(opts, ansiBold, card.Label), currentWidth)
		middle += padRight(card.Value, currentWidth)
		bars += padRight(progressBar(opts, card.Current, card.Failed, card.Total, max(6, currentWidth-2), card.Color), currentWidth)
	}
	return box(opts, "AT A GLANCE", []string{top, middle, bars}, width)
}

func renderDayAgenda(opts renderOptions, width int, data daySummaryData) []string {
	lines := []string{paint(opts, ansiBold, fmt.Sprintf("ISSUES  %d", issueCount(data.Summary)))}
	if data.Summary == nil || len(data.Summary.Issues) == 0 {
		lines = append(lines, paint(opts, ansiDim, "No issues scheduled."))
	} else {
		for _, issue := range data.Summary.Issues {
			mark, color := issueMark(issue.Status)
			meta := issueCLIText(issue)
			available := max(8, width-visibleWidth(mark)-visibleWidth(meta)-7)
			lines = append(lines, paint(opts, color, mark)+" "+truncateText(issue.Title, available)+padGap(issue.Title, available)+"  "+meta)
		}
	}
	lines = append(lines, "", paint(opts, ansiBold, fmt.Sprintf("HABITS  %d", len(data.Habits))))
	if len(data.Habits) == 0 {
		lines = append(lines, paint(opts, ansiDim, "No habits due."))
	} else {
		for _, habit := range data.Habits {
			mark, color := habitMark(habit)
			meta := habitCLIText(habit)
			available := max(8, width-visibleWidth(mark)-visibleWidth(meta)-7)
			lines = append(lines, paint(opts, color, mark)+" "+truncateText(habit.Name, available)+padGap(habit.Name, available)+"  "+meta)
		}
	}
	return lines
}

func renderDaySignals(opts renderOptions, width int, data daySummaryData) []string {
	lines := []string{paint(opts, ansiBold, "ACCOUNTABILITY")}
	if data.Plan == nil || strings.TrimSpace(data.Plan.Date) == "" {
		lines = append(lines, paint(opts, ansiDim, "No committed plan."))
	} else {
		s := data.Plan.Summary
		lines = append(lines,
			fmt.Sprintf("%.0f%% score  ·  pressure %.1f", s.AccountabilityScore, s.BacklogPressure),
			progressBar(opts, s.CompletedCount, s.FailedCount+s.AbandonedCount, max(1, s.PlannedCount), max(10, width-4), ansiGreen),
			paint(opts, ansiDim, fmt.Sprintf("done %d  pending %d  failed %d  abandoned %d", s.CompletedCount, s.PendingRollbackCount, s.FailedCount, s.AbandonedCount)),
		)
	}
	lines = append(lines, "", paint(opts, ansiBold, "WELLBEING"))
	if data.CheckIn == nil || strings.TrimSpace(data.CheckIn.Date) == "" {
		lines = append(lines, paint(opts, ansiDim, "No check-in recorded."))
	} else {
		checkIn := data.CheckIn
		values := []string{fmt.Sprintf("mood %d/5", checkIn.Mood), fmt.Sprintf("energy %d/5", checkIn.Energy)}
		if checkIn.SleepHours != nil {
			values = append(values, "sleep "+formatFloatHours(*checkIn.SleepHours))
		}
		if checkIn.ScreenTimeMinutes != nil {
			values = append(values, "screen "+formatMinutes(*checkIn.ScreenTimeMinutes))
		}
		lines = append(lines, strings.Join(values, "  ·  "))
		if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
			lines = append(lines, paint(opts, ansiDim, truncateText(strings.TrimSpace(*checkIn.Notes), max(12, width-4))))
		}
	}
	lines = append(lines, "", paint(opts, ansiBold, "MOMENTUM"))
	if data.Streaks == nil {
		lines = append(lines, paint(opts, ansiDim, "No streak data."))
	} else {
		lines = append(lines, fmt.Sprintf("focus %d/%d  ·  check-ins %d/%d  ·  habits %d/%d",
			data.Streaks.CurrentFocusDays, data.Streaks.LongestFocusDays,
			data.Streaks.CurrentCheckInDays, data.Streaks.LongestCheckInDays,
			data.Streaks.CurrentHabitDays, data.Streaks.LongestHabitDays,
		))
	}
	return lines
}

func renderRangeProgress(opts renderOptions, width int, data rangeSummaryData) string {
	rollup := data.Rollup
	window := data.Window
	lines := make([]string, 0, 4)
	barWidth := max(10, width-34)
	if window != nil {
		lines = append(lines, ratioLine(opts, "Plan completion", window.CompletedCount, window.FailedCount+window.AbandonedCount, window.PlannedCount, barWidth, ansiGreen))
	}
	if rollup != nil {
		totalIssues := 0
		for _, day := range data.Metrics {
			totalIssues += day.TotalIssues
		}
		lines = append(lines,
			ratioLine(opts, "Issue resolution", rollup.CompletedIssues+rollup.AbandonedIssues, 0, totalIssues, barWidth, ansiGreen),
			ratioLine(opts, "Habit completion", rollup.HabitCompletedCount, rollup.HabitFailedCount, rollup.HabitDueCount, barWidth, ansiGreen),
			ratioLine(opts, "Check-in coverage", rollup.CheckInDays, 0, max(1, rollup.Days), barWidth, ansiMagenta),
		)
	}
	return box(opts, "OUTCOMES", lines, width)
}

func ratioLine(opts renderOptions, label string, current, failed, total, barWidth int, color string) string {
	ratio := 0
	if total > 0 {
		ratio = current * 100 / total
	}
	return fmt.Sprintf("%-19s %3d%%  %s", label, ratio, progressBar(opts, current, failed, total, barWidth, color))
}

func renderRangeDays(opts renderOptions, width int, data rangeSummaryData) string {
	if len(data.Metrics) == 0 {
		return box(opts, "DAILY RHYTHM", []string{paint(opts, ansiDim, "No activity in this period.")}, width)
	}
	maxWorked := 1
	for _, day := range data.Metrics {
		maxWorked = max(maxWorked, day.WorkedSeconds)
	}
	wide := width >= 100
	barWidth := max(8, width-79)
	if !wide {
		barWidth = max(6, width-55)
	}
	lines := make([]string, 0, len(data.Metrics)+1)
	if wide {
		lines = append(lines, paint(opts, ansiBold, fmt.Sprintf("%-10s  %-8s %-7s %-8s %-9s %-10s %-9s %s", "DATE", "WORKED", "REST", "SESSIONS", "ISSUES", "HABITS", "MOOD", "ACTIVITY")))
	} else {
		lines = append(lines, paint(opts, ansiBold, fmt.Sprintf("%-10s  %-8s %-8s %-9s %s", "DATE", "WORKED", "ISSUES", "HABITS", "ACTIVITY")))
	}
	for _, day := range data.Metrics {
		issues := fmt.Sprintf("%d/%d", day.CompletedIssues+day.AbandonedIssues, day.TotalIssues)
		habits := fmt.Sprintf("%d/%d", day.HabitCompletedCount, day.HabitDueCount)
		activity := progressBar(opts, day.WorkedSeconds, 0, maxWorked, barWidth, ansiCyan)
		if wide {
			mood := "-"
			if day.CheckIn != nil {
				mood = fmt.Sprintf("%d/5", day.CheckIn.Mood)
			}
			lines = append(lines, fmt.Sprintf("%-10s  %-8s %-7s %-8d %-9s %-10s %-9s %s",
				day.Date, formatSeconds(day.WorkedSeconds), formatSeconds(day.RestSeconds), day.SessionCount, issues, habits, mood, activity))
		} else {
			lines = append(lines, fmt.Sprintf("%-10s  %-8s %-8s %-9s %s", day.Date, formatSeconds(day.WorkedSeconds), issues, habits, activity))
		}
	}
	return box(opts, "DAILY RHYTHM", lines, width)
}

func agendaBlock(opts renderOptions, width int, lines []string) string {
	return box(opts, "AGENDA", lines, width)
}

func signalBlock(opts renderOptions, width int, lines []string) string {
	return box(opts, "SIGNALS", lines, width)
}

func box(opts renderOptions, title string, lines []string, width int) string {
	width = max(10, width)
	inner := width - 4
	topLabel := ""
	if title != "" {
		topLabel = " " + title + " "
	}
	top := "+" + topLabel + strings.Repeat("-", max(0, width-visibleWidth(topLabel)-2)) + "+"
	out := []string{paint(opts, ansiDim, top)}
	for _, line := range lines {
		for _, wrapped := range wrapLine(line, inner) {
			out = append(out, paint(opts, ansiDim, "|")+" "+padRight(wrapped, inner)+" "+paint(opts, ansiDim, "|"))
		}
	}
	out = append(out, paint(opts, ansiDim, "+"+strings.Repeat("-", width-2)+"+"))
	return strings.Join(out, "\n")
}

func joinBlocks(left, right string) string {
	leftLines, rightLines := strings.Split(left, "\n"), strings.Split(right, "\n")
	leftWidth := 0
	for _, line := range leftLines {
		leftWidth = max(leftWidth, visibleWidth(line))
	}
	height := max(len(leftLines), len(rightLines))
	out := make([]string, 0, height)
	for idx := 0; idx < height; idx++ {
		l, r := "", ""
		if idx < len(leftLines) {
			l = leftLines[idx]
		}
		if idx < len(rightLines) {
			r = rightLines[idx]
		}
		out = append(out, padRight(l, leftWidth)+r)
	}
	return strings.Join(out, "\n")
}

func progressBar(opts renderOptions, current, failed, total, width int, color string) string {
	if width < 1 {
		return ""
	}
	full, empty := "#", "-"
	if opts.Unicode {
		full, empty = "█", "░"
	}
	if total < 1 {
		return paint(opts, ansiDim, strings.Repeat(empty, width))
	}
	filled := clampInt(current*width/total, 0, width)
	failedWidth := clampInt(failed*width/total, 0, width-filled)
	return paint(opts, color, strings.Repeat(full, filled)) +
		paint(opts, ansiRed, strings.Repeat(full, failedWidth)) +
		paint(opts, ansiDim, strings.Repeat(empty, width-filled-failedWidth))
}

func summaryScope(ctx *sharedtypes.ActiveContext) string {
	if ctx == nil {
		return "Workspace"
	}
	parts := make([]string, 0, 3)
	for _, value := range []*string{ctx.RepoName, ctx.StreamName, ctx.IssueTitle} {
		if value != nil && strings.TrimSpace(*value) != "" {
			parts = append(parts, strings.TrimSpace(*value))
		}
	}
	if len(parts) == 0 {
		return "Workspace"
	}
	return strings.Join(parts, " > ")
}

func todayTimer(date string, timer *sharedtypes.TimerState) *sharedtypes.TimerState {
	if date != nowFn().Format("2006-01-02") {
		return nil
	}
	return timer
}

func formatTimerCompact(timer *sharedtypes.TimerState) string {
	segment := "session"
	if timer.SegmentType != nil {
		segment = string(*timer.SegmentType)
	}
	return segment + "  " + formatSeconds(timer.ElapsedSeconds)
}

func issueCount(summary *sharedtypes.DailyIssueSummary) int {
	if summary == nil {
		return 0
	}
	return len(summary.Issues)
}

func issueMark(status sharedtypes.IssueStatus) (string, string) {
	switch status {
	case sharedtypes.IssueStatusDone:
		return "[x]", ansiGreen
	case sharedtypes.IssueStatusAbandoned:
		return "[-]", ansiRed
	case sharedtypes.IssueStatusBlocked:
		return "[!]", ansiRed
	case sharedtypes.IssueStatusInProgress, sharedtypes.IssueStatusInReview:
		return "[>]", ansiCyan
	default:
		return "[ ]", ansiYellow
	}
}

func habitMark(habit sharedtypes.HabitDailyItem) (string, string) {
	if habit.Status == sharedtypes.HabitCompletionStatusFailed {
		return "[-]", ansiRed
	}
	if habit.Completed {
		return "[x]", ansiGreen
	}
	return "[ ]", ansiYellow
}

func issueCLIText(issue sharedtypes.Issue) string {
	parts := []string{strings.ReplaceAll(string(issue.Status), "_", " ")}
	if issue.EstimateMinutes != nil {
		parts = append(parts, formatMinutes(*issue.EstimateMinutes))
	}
	if issue.WorkedSeconds > 0 {
		parts = append(parts, formatSeconds(issue.WorkedSeconds)+" worked")
	}
	return strings.Join(parts, " · ")
}

func habitCLIText(habit sharedtypes.HabitDailyItem) string {
	parts := []string{"due"}
	if habit.Completed {
		parts[0] = string(habit.Status)
	}
	if habit.DurationMinutes != nil {
		parts = append(parts, formatMinutes(*habit.DurationMinutes))
	}
	target := habit.TargetMinutes
	if habit.SnapshotTarget != nil {
		target = habit.SnapshotTarget
	}
	if target != nil {
		parts = append(parts, "target "+formatMinutes(*target))
	}
	return strings.Join(parts, " · ")
}

func cliHabitCounts(habits []sharedtypes.HabitDailyItem) (completed, failed int) {
	for _, habit := range habits {
		if habit.Status == sharedtypes.HabitCompletionStatusFailed {
			failed++
		} else if habit.Completed {
			completed++
		}
	}
	return completed, failed
}

func rangeDays(start, end string) int {
	first, errFirst := timeParse(start)
	last, errLast := timeParse(end)
	if errFirst != nil || errLast != nil || last.Before(first) {
		return 1
	}
	return int(last.Sub(first).Hours()/24) + 1
}

func timeParse(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func wrapLine(value string, width int) []string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
	if visibleWidth(value) <= width {
		return []string{value}
	}
	return []string{truncateText(value, width)}
}

func truncateText(value string, width int) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
	if width < 1 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width < 4 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func padGap(value string, width int) string {
	return strings.Repeat(" ", max(0, width-visibleWidth(truncateText(value, width))))
}

func padRight(value string, width int) string {
	return value + strings.Repeat(" ", max(0, width-visibleWidth(value)))
}

func visibleWidth(value string) int {
	width := 0
	inEscape := false
	for _, r := range value {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			width++
		}
	}
	return width
}

func paint(opts renderOptions, code, value string) string {
	if !opts.Color || value == "" {
		return value
	}
	return code + value + ansiReset
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)
