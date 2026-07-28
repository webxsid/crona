package summary

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"crona/cli/internal/flags"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

var nowFn = time.Now

func Usage() string {
	return "Usage: crona summary [DATE | --date DATE | --start DATE --end DATE | --yesterday | --week | --month | --last-x-days N]\n"
}

func Run(args []string, deps Deps) error {
	if len(args) > 0 && flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}

	date, start, end, err := parseArgs(args)
	if err != nil {
		return err
	}
	if start != "" || end != "" {
		return runRange(start, end, deps)
	}
	return runDay(date, deps)
}

func parseArgs(args []string) (date, start, end string, err error) {
	if len(args) == 1 && !strings.HasPrefix(strings.TrimSpace(args[0]), "-") {
		date = strings.TrimSpace(args[0])
		if err := validateDate(date); err != nil {
			return "", "", "", err
		}
		return date, "", "", nil
	}
	if len(args) == 2 &&
		!strings.HasPrefix(strings.TrimSpace(args[0]), "-") &&
		!strings.HasPrefix(strings.TrimSpace(args[1]), "-") {
		start = strings.TrimSpace(args[0])
		end = strings.TrimSpace(args[1])
		if err := validateDate(start); err != nil {
			return "", "", "", err
		}
		if err := validateDate(end); err != nil {
			return "", "", "", err
		}
		if start > end {
			return "", "", "", errors.New("start date must be on or before end date")
		}
		return "", start, end, nil
	}

	fs := flags.New("summary")
	dateFlag := fs.String("date", "", "")
	startFlag := fs.String("start", "", "")
	endFlag := fs.String("end", "", "")
	yesterdayFlag := fs.Bool("yesterday", false, "")
	weekFlag := fs.Bool("week", false, "")
	monthFlag := fs.Bool("month", false, "")
	lastXDaysFlag := fs.Int("last-x-days", 0, "")
	if err := fs.Parse(args); err != nil {
		return "", "", "", err
	}
	if len(fs.Args()) > 0 {
		return "", "", "", fmt.Errorf("unexpected argument: %s", fs.Args()[0])
	}

	date = strings.TrimSpace(*dateFlag)
	start = strings.TrimSpace(*startFlag)
	end = strings.TrimSpace(*endFlag)
	yesterday := *yesterdayFlag
	week := *weekFlag
	month := *monthFlag
	lastXDays := *lastXDaysFlag
	hasLastXDays := hasArg(args, "--last-x-days")

	if sumBool(yesterday, week, month, hasLastXDays) > 1 {
		return "", "", "", errors.New("use only one of --yesterday, --week, --month, or --last-x-days")
	}
	if hasLastXDays && lastXDays <= 0 {
		return "", "", "", errors.New("--last-x-days requires a positive integer")
	}

	if date != "" && (start != "" || end != "") {
		return "", "", "", errors.New("use either --date or --start/--end, not both")
	}
	if date == "" && start == "" && end == "" && !yesterday && !week && !month && !hasLastXDays {
		date = nowFn().Format("2006-01-02")
	}
	if yesterday || week || month || hasLastXDays {
		if date != "" || start != "" || end != "" {
			return "", "", "", errors.New("preset flags cannot be combined with --date or --start/--end")
		}
		today := nowFn()
		switch {
		case yesterday:
			date = today.AddDate(0, 0, -1).Format("2006-01-02")
		case week:
			start = startOfCalendarWeek(today).Format("2006-01-02")
			end = today.Format("2006-01-02")
		case month:
			start = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location()).Format("2006-01-02")
			end = today.Format("2006-01-02")
		case hasLastXDays:
			start = today.AddDate(0, 0, -(lastXDays - 1)).Format("2006-01-02")
			end = today.Format("2006-01-02")
		}
	}
	if date != "" {
		if err := validateDate(date); err != nil {
			return "", "", "", err
		}
		return date, "", "", nil
	}
	if start == "" || end == "" {
		return "", "", "", errors.New("both --start and --end are required for a range")
	}
	if err := validateDate(start); err != nil {
		return "", "", "", err
	}
	if err := validateDate(end); err != nil {
		return "", "", "", err
	}
	if start > end {
		return "", "", "", errors.New("start date must be on or before end date")
	}
	return "", start, end, nil
}

func sumBool(values ...bool) int {
	total := 0
	for _, value := range values {
		if value {
			total++
		}
	}
	return total
}

func startOfCalendarWeek(value time.Time) time.Time {
	offset := int(value.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return value.AddDate(0, 0, -offset)
}

func hasArg(args []string, prefix string) bool {
	for _, arg := range args {
		if strings.HasPrefix(strings.TrimSpace(arg), prefix) {
			return true
		}
	}
	return false
}

func runDay(date string, deps Deps) error {
	ctx, timer, err := loadContextAndTimer(deps)
	if err != nil {
		return err
	}
	summary, err := loadDailySummary(deps, date)
	if err != nil {
		return err
	}
	habits, err := loadDueHabits(deps, date)
	if err != nil {
		return err
	}
	plan, err := loadDailyPlan(deps, date)
	if err != nil {
		return err
	}
	checkIn, err := loadDailyCheckIn(deps, date)
	if err != nil {
		return err
	}
	rollup, err := loadMetricsRollup(deps, shiftDate(date, -6), date)
	if err != nil {
		return err
	}
	streaks, err := loadLifetimeStreaks(deps, date)
	if err != nil {
		return err
	}
	data := daySummaryData{
		Date:    date,
		Context: ctx,
		Timer:   timer,
		Summary: summary,
		Habits:  habits,
		Plan:    plan,
		CheckIn: checkIn,
		Rollup:  rollup,
		Streaks: streaks,
	}
	return renderDaySummary(deps.Stdout, data)
}

func runRange(start, end string, deps Deps) error {
	ctx, timer, err := loadContextAndTimer(deps)
	if err != nil {
		return err
	}
	window, err := loadDashboardWindow(deps, start, end, ctx)
	if err != nil {
		return err
	}
	metrics, err := loadMetricsRange(deps, start, end)
	if err != nil {
		return err
	}
	rollup, err := loadMetricsRollup(deps, start, end)
	if err != nil {
		return err
	}
	streaks, err := loadRangeStreaks(deps, start, end)
	if err != nil {
		return err
	}
	data := rangeSummaryData{
		Start:   start,
		End:     end,
		Context: ctx,
		Timer:   timer,
		Window:  window,
		Metrics: metrics,
		Rollup:  rollup,
		Streaks: streaks,
	}
	return renderRangeSummary(deps.Stdout, data)
}

type daySummaryData struct {
	Date    string
	Context *sharedtypes.ActiveContext
	Timer   *sharedtypes.TimerState
	Summary *sharedtypes.DailyIssueSummary
	Habits  []sharedtypes.HabitDailyItem
	Plan    *sharedtypes.DailyPlan
	CheckIn *sharedtypes.DailyCheckIn
	Rollup  *sharedtypes.MetricsRollup
	Streaks *sharedtypes.StreakSummary
}

type rangeSummaryData struct {
	Start   string
	End     string
	Context *sharedtypes.ActiveContext
	Timer   *sharedtypes.TimerState
	Window  *sharedtypes.DashboardWindowSummary
	Metrics []sharedtypes.DailyMetricsDay
	Rollup  *sharedtypes.MetricsRollup
	Streaks *sharedtypes.StreakSummary
}

func renderDaySummary(w io.Writer, data daySummaryData) error {
	if summaryRenderOptions(w).Width >= 50 {
		return renderDayCanvas(w, data, summaryRenderOptions(w))
	}
	if _, err := fmt.Fprintln(w, "SUMMARY"); err != nil {
		return err
	}
	if err := printMetaSection(w, []kvRow{
		{Key: "mode", Value: "day"},
		{Key: "date", Value: data.Date},
	}); err != nil {
		return err
	}
	if err := printContextSection(w, data.Context, data.Timer); err != nil {
		return err
	}
	if err := printDailyIssueSection(w, data.Summary); err != nil {
		return err
	}
	if err := printDailyPlanSection(w, data.Plan); err != nil {
		return err
	}
	if err := printCheckInSection(w, data.CheckIn, data.Rollup); err != nil {
		return err
	}
	return printStreakSection(w, data.Streaks)
}

func renderRangeSummary(w io.Writer, data rangeSummaryData) error {
	if summaryRenderOptions(w).Width >= 50 {
		return renderRangeCanvas(w, data, summaryRenderOptions(w))
	}
	if _, err := fmt.Fprintln(w, "SUMMARY"); err != nil {
		return err
	}
	if err := printMetaSection(w, []kvRow{
		{Key: "mode", Value: "range"},
		{Key: "start", Value: data.Start},
		{Key: "end", Value: data.End},
	}); err != nil {
		return err
	}
	if err := printContextSection(w, data.Context, data.Timer); err != nil {
		return err
	}
	if err := printRangeWindowSection(w, data.Window); err != nil {
		return err
	}
	if err := printRangeMetricsSection(w, data.Rollup); err != nil {
		return err
	}
	if err := printRangeStreakSection(w, data.Streaks); err != nil {
		return err
	}
	return printRangeDaysSection(w, data.Metrics)
}

type kvRow struct {
	Key   string
	Value string
}

func printMetaSection(w io.Writer, rows []kvRow) error {
	if len(rows) == 0 {
		return nil
	}
	return printTable(w, "meta", []string{"field", "value"}, rowsToCells(rows))
}

func printContextSection(w io.Writer, ctx *sharedtypes.ActiveContext, timer *sharedtypes.TimerState) error {
	rows := []kvRow{}
	if ctx != nil {
		rows = append(rows,
			kvRow{Key: "repo", Value: optionalString(ctx.RepoName, "-")},
			kvRow{Key: "stream", Value: optionalString(ctx.StreamName, "-")},
			kvRow{Key: "issue", Value: optionalString(ctx.IssueTitle, "-")},
		)
	}
	if timer != nil {
		rows = append(rows, kvRow{Key: "timer", Value: formatTimer(timer)})
	}
	if len(rows) == 0 {
		return nil
	}
	return printTable(w, "context", []string{"item", "value"}, rowsToCells(rows))
}

func printDailyIssueSection(w io.Writer, summary *sharedtypes.DailyIssueSummary) error {
	if summary == nil {
		return nil
	}
	rows := []kvRow{
		{Key: "total issues", Value: fmt.Sprintf("%d", summary.TotalIssues)},
		{Key: "resolved", Value: fmt.Sprintf("%d", summary.CompletedIssues+summary.AbandonedIssues)},
		{Key: "completed", Value: fmt.Sprintf("%d", summary.CompletedIssues)},
		{Key: "abandoned", Value: fmt.Sprintf("%d", summary.AbandonedIssues)},
		{Key: "estimated", Value: formatMinutes(summary.TotalEstimatedMinutes)},
		{Key: "worked", Value: formatSeconds(summary.WorkedSeconds)},
		{Key: "delta", Value: formatSignedSeconds(summary.WorkedSeconds - summary.TotalEstimatedMinutes*60)},
	}
	return printTable(w, "day", []string{"metric", "value"}, rowsToCells(rows))
}

func printDailyPlanSection(w io.Writer, plan *sharedtypes.DailyPlan) error {
	if plan == nil {
		return nil
	}
	summary := plan.Summary
	rows := []kvRow{
		{Key: "planned", Value: fmt.Sprintf("%d", summary.PlannedCount)},
		{Key: "completed", Value: fmt.Sprintf("%d", summary.CompletedCount)},
		{Key: "failed", Value: fmt.Sprintf("%d", summary.FailedCount)},
		{Key: "abandoned", Value: fmt.Sprintf("%d", summary.AbandonedCount)},
		{Key: "pending", Value: fmt.Sprintf("%d", summary.PendingRollbackCount)},
		{Key: "score", Value: fmt.Sprintf("%.1f", summary.AccountabilityScore)},
		{Key: "pressure", Value: fmt.Sprintf("%.1f", summary.BacklogPressure)},
	}
	return printTable(w, "plan", []string{"metric", "value"}, rowsToCells(rows))
}

func printCheckInSection(w io.Writer, checkIn *sharedtypes.DailyCheckIn, rollup *sharedtypes.MetricsRollup) error {
	rows := []kvRow{}
	if checkIn != nil {
		rows = append(rows,
			kvRow{Key: "mood", Value: fmt.Sprintf("%d/5", checkIn.Mood)},
			kvRow{Key: "energy", Value: fmt.Sprintf("%d/5", checkIn.Energy)},
		)
		if checkIn.SleepHours != nil {
			rows = append(rows, kvRow{Key: "sleep", Value: formatFloatHours(*checkIn.SleepHours)})
		}
		if checkIn.ScreenTimeMinutes != nil {
			rows = append(rows, kvRow{Key: "screen", Value: formatMinutes(*checkIn.ScreenTimeMinutes)})
		}
		if notes := strings.TrimSpace(optionalString(checkIn.Notes, "")); notes != "" {
			rows = append(rows, kvRow{Key: "notes", Value: notes})
		}
	}
	if rollup != nil {
		if rollup.AverageMood != nil {
			rows = append(rows, kvRow{Key: "avg mood", Value: fmt.Sprintf("%.1f/5", *rollup.AverageMood)})
		}
		if rollup.AverageEnergy != nil {
			rows = append(rows, kvRow{Key: "avg energy", Value: fmt.Sprintf("%.1f/5", *rollup.AverageEnergy)})
		}
		if rollup.AverageSleepHours != nil {
			rows = append(rows, kvRow{Key: "avg sleep", Value: formatFloatHours(*rollup.AverageSleepHours)})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return printTable(w, "check-in", []string{"metric", "value"}, rowsToCells(rows))
}

func printStreakSection(w io.Writer, streaks *sharedtypes.StreakSummary) error {
	if streaks == nil {
		return nil
	}
	rows := []kvRow{
		{Key: "focus", Value: fmt.Sprintf("%d / %d", streaks.CurrentFocusDays, streaks.LongestFocusDays)},
		{Key: "check-ins", Value: fmt.Sprintf("%d / %d", streaks.CurrentCheckInDays, streaks.LongestCheckInDays)},
		{Key: "habits", Value: fmt.Sprintf("%d / %d", streaks.CurrentHabitDays, streaks.LongestHabitDays)},
	}
	return printTable(w, "streaks", []string{"metric", "current / longest"}, rowsToCells(rows))
}

func printRangeWindowSection(w io.Writer, window *sharedtypes.DashboardWindowSummary) error {
	if window == nil {
		return nil
	}
	rows := []kvRow{
		{Key: "planned", Value: fmt.Sprintf("%d", window.PlannedCount)},
		{Key: "completed", Value: fmt.Sprintf("%d", window.CompletedCount)},
		{Key: "failed", Value: fmt.Sprintf("%d", window.FailedCount)},
		{Key: "abandoned", Value: fmt.Sprintf("%d", window.AbandonedCount)},
		{Key: "missed", Value: fmt.Sprintf("%d", window.MissedCount)},
		{Key: "carry over", Value: fmt.Sprintf("%d", window.CarryOverCount)},
		{Key: "score", Value: fmt.Sprintf("%.1f", window.AccountabilityScore)},
	}
	return printTable(w, "window", []string{"metric", "value"}, rowsToCells(rows))
}

func printRangeMetricsSection(w io.Writer, rollup *sharedtypes.MetricsRollup) error {
	if rollup == nil {
		return nil
	}
	rows := []kvRow{
		{Key: "days", Value: fmt.Sprintf("%d", rollup.Days)},
		{Key: "check-ins", Value: fmt.Sprintf("%d", rollup.CheckInDays)},
		{Key: "focus", Value: fmt.Sprintf("%d", rollup.FocusDays)},
		{Key: "worked", Value: formatSeconds(rollup.WorkedSeconds)},
		{Key: "rest", Value: formatSeconds(rollup.RestSeconds)},
		{Key: "issues done", Value: fmt.Sprintf("%d", rollup.CompletedIssues)},
		{Key: "issues abandoned", Value: fmt.Sprintf("%d", rollup.AbandonedIssues)},
		{Key: "habit done", Value: fmt.Sprintf("%d", rollup.HabitCompletedCount)},
		{Key: "habit failed", Value: fmt.Sprintf("%d", rollup.HabitFailedCount)},
	}
	return printTable(w, "rollup", []string{"metric", "value"}, rowsToCells(rows))
}

func printRangeStreakSection(w io.Writer, streaks *sharedtypes.StreakSummary) error {
	if streaks == nil {
		return nil
	}
	rows := []kvRow{
		{Key: "focus", Value: fmt.Sprintf("%d / %d", streaks.CurrentFocusDays, streaks.LongestFocusDays)},
		{Key: "check-ins", Value: fmt.Sprintf("%d / %d", streaks.CurrentCheckInDays, streaks.LongestCheckInDays)},
		{Key: "habits", Value: fmt.Sprintf("%d / %d", streaks.CurrentHabitDays, streaks.LongestHabitDays)},
	}
	return printTable(w, "streaks", []string{"metric", "current / longest"}, rowsToCells(rows))
}

func printRangeDaysSection(w io.Writer, days []sharedtypes.DailyMetricsDay) error {
	if len(days) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(days))
	for _, day := range days {
		rows = append(rows, []string{
			day.Date,
			formatSeconds(day.WorkedSeconds),
			formatSeconds(day.RestSeconds),
			fmt.Sprintf("%d", day.SessionCount),
			fmt.Sprintf("%d/%d", day.CompletedIssues, day.TotalIssues),
			fmt.Sprintf("%d/%d/%d", day.HabitCompletedCount, day.HabitDueCount, day.HabitFailedCount),
		})
	}
	return printTable(w, "days", []string{"date", "worked", "rest", "sessions", "issues", "habits"}, rows)
}

func rowsToCells(rows []kvRow) [][]string {
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, []string{row.Key, row.Value})
	}
	return out
}

func printTable(w io.Writer, title string, headers []string, rows [][]string) error {
	if _, err := fmt.Fprintf(w, "\n%s\n", strings.ToUpper(title)); err != nil {
		return err
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = displayWidth(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if got := displayWidth(cell); got > widths[i] {
				widths[i] = got
			}
		}
	}
	if err := printBorder(w, widths); err != nil {
		return err
	}
	if err := printRow(w, headers, widths); err != nil {
		return err
	}
	if err := printBorder(w, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := printRow(w, row, widths); err != nil {
			return err
		}
	}
	return printBorder(w, widths)
}

func printBorder(w io.Writer, widths []int) error {
	if _, err := fmt.Fprint(w, "+"); err != nil {
		return err
	}
	for _, width := range widths {
		if _, err := fmt.Fprint(w, strings.Repeat("-", width+2), "+"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func printRow(w io.Writer, cells []string, widths []int) error {
	if _, err := fmt.Fprint(w, "|"); err != nil {
		return err
	}
	for i, width := range widths {
		cell := ""
		if i < len(cells) {
			cell = strings.ReplaceAll(strings.TrimSpace(cells[i]), "\n", " ")
		}
		padding := width - displayWidth(cell)
		if padding < 0 {
			padding = 0
		}
		if _, err := fmt.Fprintf(w, " %s%s |", cell, strings.Repeat(" ", padding)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func displayWidth(value string) int {
	return len([]rune(strings.TrimSpace(value)))
}

func validateDate(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("date is required")
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("invalid date %q", value)
	}
	return nil
}

func shiftDate(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func optionalString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	if trimmed := strings.TrimSpace(*value); trimmed != "" {
		return trimmed
	}
	return fallback
}

func formatTimer(timer *sharedtypes.TimerState) string {
	segment := "-"
	if timer.SegmentType != nil {
		segment = string(*timer.SegmentType)
	} else if timer.ReadySegmentType != nil {
		segment = string(*timer.ReadySegmentType)
	}
	issue := "-"
	if timer.IssueID != nil {
		issue = fmt.Sprintf("%d", *timer.IssueID)
	}
	session := "-"
	if timer.SessionID != nil {
		session = *timer.SessionID
	}
	return fmt.Sprintf("%s %s issue=%s session=%s elapsed=%s", timer.State, segment, issue, session, formatSeconds(timer.ElapsedSeconds))
}

func formatSeconds(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func formatSignedSeconds(totalSeconds int) string {
	sign := "+"
	if totalSeconds < 0 {
		sign = "-"
		totalSeconds = -totalSeconds
	}
	return sign + formatSeconds(totalSeconds)
}

func formatMinutes(totalMinutes int) string {
	if totalMinutes < 0 {
		totalMinutes = 0
	}
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func formatFloatHours(hours float64) string {
	if hours < 0 {
		hours = 0
	}
	totalMinutes := int(hours*60 + 0.5)
	return formatMinutes(totalMinutes)
}

func loadContextAndTimer(deps Deps) (*sharedtypes.ActiveContext, *sharedtypes.TimerState, error) {
	var ctx sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextGet, nil, &ctx); err != nil {
		return nil, nil, err
	}
	var timer sharedtypes.TimerState
	if err := deps.CallKernel(protocol.MethodTimerGetState, nil, &timer); err != nil {
		return &ctx, nil, err
	}
	return &ctx, &timer, nil
}

func loadDailySummary(deps Deps, date string) (*sharedtypes.DailyIssueSummary, error) {
	var out sharedtypes.DailyIssueSummary
	if err := deps.CallKernel(protocol.MethodIssueDailySummary, shareddto.DailyIssueSummaryQuery{Date: optionalDate(date)}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func loadDailyPlan(deps Deps, date string) (*sharedtypes.DailyPlan, error) {
	var out sharedtypes.DailyPlan
	if err := deps.CallKernel(protocol.MethodDailyPlanGet, shareddto.DailyPlanQuery{Date: strings.TrimSpace(date)}, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.Date) == "" {
		return nil, nil
	}
	return &out, nil
}

func loadDueHabits(deps Deps, date string) ([]sharedtypes.HabitDailyItem, error) {
	var out []sharedtypes.HabitDailyItem
	if err := deps.CallKernel(protocol.MethodHabitListDue, shareddto.ListHabitsDueQuery{
		Date: strings.TrimSpace(date),
	}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func loadDailyCheckIn(deps Deps, date string) (*sharedtypes.DailyCheckIn, error) {
	var out sharedtypes.DailyCheckIn
	if err := deps.CallKernel(protocol.MethodCheckInGet, shareddto.DailyCheckInQuery{Date: strings.TrimSpace(date)}, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.Date) == "" {
		return nil, nil
	}
	return &out, nil
}

func loadMetricsRollup(deps Deps, start, end string) (*sharedtypes.MetricsRollup, error) {
	var out sharedtypes.MetricsRollup
	if err := deps.CallKernel(protocol.MethodMetricsRollup, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func loadLifetimeStreaks(deps Deps, date string) (*sharedtypes.StreakSummary, error) {
	var out sharedtypes.StreakSummary
	if err := deps.CallKernel(protocol.MethodMetricsStreaksLifetime, shareddto.DailyCheckInQuery{
		Date: strings.TrimSpace(date),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func loadDashboardWindow(deps Deps, start, end string, ctx *sharedtypes.ActiveContext) (*sharedtypes.DashboardWindowSummary, error) {
	var out sharedtypes.DashboardWindowSummary
	req := shareddto.DashboardWindowQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}
	if ctx != nil {
		req.RepoID = ctx.RepoID
		req.StreamID = ctx.StreamID
		req.IssueID = ctx.IssueID
	}
	if err := deps.CallKernel(protocol.MethodDashboardWindow, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func loadMetricsRange(deps Deps, start, end string) ([]sharedtypes.DailyMetricsDay, error) {
	var out []sharedtypes.DailyMetricsDay
	if err := deps.CallKernel(protocol.MethodMetricsRange, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func loadRangeStreaks(deps Deps, start, end string) (*sharedtypes.StreakSummary, error) {
	var out sharedtypes.StreakSummary
	if err := deps.CallKernel(protocol.MethodMetricsStreaks, shareddto.DateRangeQuery{
		Start: strings.TrimSpace(start),
		End:   strings.TrimSpace(end),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func optionalDate(date string) *string {
	trimmed := strings.TrimSpace(date)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
