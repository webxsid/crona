package review

import (
	"math"
	"regexp"
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewaway "crona/tui/internal/tui/views/away"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func reviewState(width, height int) types.ContentState {
	end := "2026-10-14"
	dates := []string{"2026-10-08", "2026-10-09", "2026-10-10", "2026-10-11", "2026-10-12", "2026-10-13", "2026-10-14"}
	metrics := make([]api.DailyMetricsDay, 0, len(dates))
	days := make([]sharedtypes.DashboardWindowDay, 0, len(dates))
	for i, date := range dates {
		day := api.DailyMetricsDay{Date: date, WorkedSeconds: (i + 2) * 3600, TotalEstimatedMinutes: 300}
		if i != 2 {
			day.CheckIn = &api.DailyCheckIn{Date: date, Mood: 1 + i%5, Energy: 3, CreatedAt: date + "T09:00:00Z"}
		}
		metrics = append(metrics, day)
		days = append(days, sharedtypes.DashboardWindowDay{Date: date, Status: "done", PlannedCount: 5, CompletedCount: 4})
	}
	mood, energy, sleep := 3.8, 3.4, 7.1
	return types.ContentState{
		View: "review", Pane: "review_days", Width: width, Height: height,
		DashboardDate: end, SummaryDate: end,
		ReviewWindowDays: 7, ReviewWindowEnd: end,
		MetricsRange:       metrics,
		RollupMetricsRange: metrics,
		MetricsRollup: &api.MetricsRollup{
			Days: 7, CheckInDays: 6, WorkedSeconds: 17 * 3600,
			AverageMood: &mood, AverageEnergy: &energy, AverageSleepHours: &sleep,
			LatestBurnout: &api.BurnoutIndicator{Score: 20, Level: "low", Factors: map[string]float64{"breakDebt": 0.3, "recoveryBreaks": -0.2}},
		},
		DashboardWindow:  &api.DashboardWindowSummary{StartDate: dates[0], EndDate: end, PlannedCount: 27, CompletedCount: 22, FailedCount: 3, Days: days},
		WeeklyFocusScore: &api.FocusScoreSummary{Score: 72},
	}
}

func plain(s string) string { return ansi.Strip(s) }

var dayRowPattern = regexp.MustCompile(`(Mon|Tue|Wed|Thu|Fri|Sat|Sun) \d`)

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestStripTwoRowsWidthExact(t *testing.T) {
	for _, width := range []int{60, 100, 160} {
		rows := StripRows(types.Theme{}, reviewState(width, 40), width)
		if len(rows) != 2 {
			t.Fatalf("width %d: expected 2 rows, got %d", width, len(rows))
		}
		for i, row := range rows {
			if w := ansi.StringWidth(row); w != width {
				t.Fatalf("width %d: row %d is %d cells: %q", width, i, w, plain(row))
			}
		}
	}
	rows := StripRows(types.Theme{}, reviewState(160, 40), 160)
	top, bottom := squash(plain(rows[0])), squash(plain(rows[1]))
	for _, want := range []string{"Worked 17h", "Focus 72", "done 22", "failed 3"} {
		if !strings.Contains(top, want) {
			t.Fatalf("expected %q in %q", want, top)
		}
	}
	for _, want := range []string{"Mood 3.8", "Energy 3.4", "Sleep 7.1h", "Check-ins 6/7", "Risk low"} {
		if !strings.Contains(bottom, want) {
			t.Fatalf("expected %q in %q", want, bottom)
		}
	}
	narrow := StripRows(types.Theme{}, reviewState(60, 40), 60)
	if strings.Contains(plain(narrow[0]), "failed") || strings.Contains(plain(narrow[1]), "Risk") {
		t.Fatalf("expected the last groups dropped at 60: %q / %q", plain(narrow[0]), plain(narrow[1]))
	}
}

func TestStripBlankWithoutData(t *testing.T) {
	state := reviewState(160, 40)
	state.MetricsRollup = nil
	bottom := squash(plain(StripRows(types.Theme{}, state, 160)[1]))
	if !strings.Contains(bottom, "Mood –") || strings.Contains(bottom, "Mood 0") {
		t.Fatalf("expected a blank mood, got %q", bottom)
	}
}

func TestPanesByWidth(t *testing.T) {
	wide := plain(Render(types.Theme{}, reviewState(200, 50)))
	for _, want := range []string{"Days", "Trends", "Wed 14 Oct"} {
		if !strings.Contains(wide, want) {
			t.Fatalf("wide: expected %q", want)
		}
	}
	medium := plain(Render(types.Theme{}, reviewState(120, 36)))
	if !strings.Contains(medium, "Days") || !strings.Contains(medium, "Wed 14 Oct") || strings.Contains(medium, "Trends") {
		t.Fatalf("medium: expected Days and the day, no Trends:\n%s", medium)
	}
	narrow := plain(Render(types.Theme{}, reviewState(80, 30)))
	if !strings.Contains(narrow, "Days") || strings.Contains(narrow, "Trends") || strings.Contains(narrow, "Wed 14 Oct") {
		t.Fatalf("narrow: expected Days only:\n%s", narrow)
	}
}

func TestShowTimeSwapsSlot(t *testing.T) {
	state := reviewState(200, 50)
	if out := plain(Render(types.Theme{}, state)); !strings.Contains(out, "[S] time") {
		t.Fatalf("wide trends: expected [S] time meta")
	}
	state.ReviewShowTime = true
	out := plain(Render(types.Theme{}, state))
	if !strings.Contains(out, "─ Time ") || strings.Contains(out, "Trends") || !strings.Contains(out, "[S] trends") {
		t.Fatalf("wide time: expected Time with [S] trends:\n%s", out)
	}
	medium := reviewState(120, 36)
	if out := plain(Render(types.Theme{}, medium)); !strings.Contains(out, "[S] time") {
		t.Fatalf("medium day: expected [S] time meta")
	}
	medium.ReviewShowTime = true
	out = plain(Render(types.Theme{}, medium))
	if !strings.Contains(out, "─ Time ") || strings.Contains(out, "Wed 14 Oct") || !strings.Contains(out, "[S] day") {
		t.Fatalf("medium time: expected Time replacing the day with [S] day:\n%s", out)
	}
}

func TestDaysRowsAndSelection(t *testing.T) {
	out := plain(Render(types.Theme{}, reviewState(200, 50)))
	var rows []string
	for _, line := range strings.Split(out, "\n") {
		left := strings.SplitN(line, "│", 2)[0]
		if dayRowPattern.MatchString(left) {
			rows = append(rows, left)
		}
	}
	if len(rows) != 7 {
		t.Fatalf("expected 7 day rows, got %d: %q", len(rows), rows)
	}
	if !strings.Contains(rows[0], "Thu 8") || !strings.Contains(rows[6], "▶") || !strings.Contains(rows[6], "Wed 14") {
		t.Fatalf("expected oldest first and Wed 14 selected: %q", rows)
	}
	if strings.ContainsAny(rows[2], "▲▬▼") {
		t.Fatalf("expected no mood mark without a check-in: %q", rows[2])
	}
}

func TestDaysScrollsLongWindow(t *testing.T) {
	state := reviewState(120, 20)
	state.ReviewWindowDays = 90
	state.DashboardDate = "2026-09-04"
	out := plain(Render(types.Theme{}, state))
	if !strings.Contains(out, "▶ Fri 4 ") {
		t.Fatalf("expected the selected day visible:\n%s", out)
	}
}

func TestShortHeightFoldsTrends(t *testing.T) {
	if out := plain(Render(types.Theme{}, reviewState(200, 22))); strings.Contains(out, "Trends") {
		t.Fatalf("expected no Trends at height 22")
	}
	short := plain(Render(types.Theme{}, reviewState(200, 28)))
	tall := plain(Render(types.Theme{}, reviewState(200, 50)))
	recovery := regexp.MustCompile(`Recovery\s+\d+%`)
	if !strings.Contains(short, "Trends") || recovery.MatchString(short) || !recovery.MatchString(tall) {
		t.Fatalf("expected Trends at 28 with fewer charts than at 50 (Recovery only when tall)")
	}
}

func TestChartsSqueezeManyPoints(t *testing.T) {
	state := reviewState(160, 40)
	var many []api.DailyMetricsDay
	for i := range 90 {
		many = append(many, api.DailyMetricsDay{Date: "2026-07-01", WorkedSeconds: i * 60})
	}
	state.MetricsRange, state.RollupMetricsRange, state.ReviewWindowDays = many, many, 90
	_ = Render(types.Theme{}, state)
}

func TestHistoricalAwayDate(t *testing.T) {
	state := reviewState(120, 36)
	state.HistoricalAwayDate = "2026-10-14"
	if Render(types.Theme{}, state) != viewaway.RenderHistorical(types.Theme{}, state) {
		t.Fatalf("expected the away screen")
	}
}

func TestRenderExactSize(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {120, 36}, {200, 50}, {200, 22}} {
		lines := strings.Split(Render(types.Theme{}, reviewState(size[0], size[1])), "\n")
		if len(lines) != size[1] {
			t.Fatalf("%v: expected %d lines, got %d", size, size[1], len(lines))
		}
		for i, line := range lines {
			if w := ansi.StringWidth(line); w != size[0] {
				t.Fatalf("%v: line %d is %d cells: %q", size, i, w, plain(line))
			}
		}
	}
}

func TestNarrowDaysHints(t *testing.T) {
	narrow := plain(Render(types.Theme{}, reviewState(80, 30)))
	if !strings.Contains(narrow, "[enter] day") || !strings.Contains(narrow, "[S] time") {
		t.Fatalf("expected narrow Days hints:\n%s", narrow)
	}
	if wide := plain(Render(types.Theme{}, reviewState(200, 50))); strings.Contains(wide, "[enter]") {
		t.Fatalf("expected no [enter] hint when the detail is visible")
	}
}

func TestChartsGapDaysWithoutCheckIn(t *testing.T) {
	lines := renderBrailleCanvas([]float64{3, math.NaN(), 3}, 1, 5, 9, 1)
	if len(lines) != 1 || []rune(lines[0])[4] != ' ' {
		t.Fatalf("expected a gap where the day has no check-in, got %q", lines)
	}
	state := reviewState(200, 50)
	state.MetricsRange[len(state.MetricsRange)-1].CheckIn = nil
	out := plain(Render(types.Theme{}, state))
	if strings.Contains(out, "0.0/5") || !strings.Contains(out, "Mood      1.0/5") {
		t.Fatalf("expected Mood labelled with the last check-in, not 0:\n%s", out)
	}
}
