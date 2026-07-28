package summary

import (
	"bytes"
	"strings"
	"testing"
	"time"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestRunDayDefaultsToToday(t *testing.T) {
	oldNowFn := nowFn
	nowFn = func() time.Time {
		return time.Date(2026, 7, 28, 9, 30, 0, 0, time.UTC)
	}
	t.Cleanup(func() { nowFn = oldNowFn })

	var out bytes.Buffer
	var gotMethods []string
	err := Run(nil, Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			gotMethods = append(gotMethods, method)
			switch method {
			case protocol.MethodContextGet:
				ctx := target.(*sharedtypes.ActiveContext)
				repo := "alpha"
				stream := "main"
				issue := "ship it"
				ctx.RepoName = &repo
				ctx.StreamName = &stream
				ctx.IssueTitle = &issue
				return nil
			case protocol.MethodTimerGetState:
				timer := target.(*sharedtypes.TimerState)
				timer.State = "running"
				timer.ElapsedSeconds = 372
				return nil
			case protocol.MethodIssueDailySummary:
				query := params.(shareddto.DailyIssueSummaryQuery)
				if query.Date == nil || *query.Date != "2026-07-28" {
					t.Fatalf("unexpected daily summary query: %#v", query)
				}
				sum := target.(*sharedtypes.DailyIssueSummary)
				sum.Date = "2026-07-28"
				sum.TotalIssues = 4
				sum.CompletedIssues = 3
				sum.AbandonedIssues = 1
				sum.TotalEstimatedMinutes = 180
				sum.WorkedSeconds = 150*60 + 30
				return nil
			case protocol.MethodHabitListDue:
				query := params.(shareddto.ListHabitsDueQuery)
				if query.Date != "2026-07-28" {
					t.Fatalf("unexpected habit query: %#v", query)
				}
				habits := target.(*[]sharedtypes.HabitDailyItem)
				*habits = []sharedtypes.HabitDailyItem{{Completed: true}}
				return nil
			case protocol.MethodDailyPlanGet:
				query := params.(shareddto.DailyPlanQuery)
				if query.Date != "2026-07-28" {
					t.Fatalf("unexpected daily plan query: %#v", query)
				}
				plan := target.(*sharedtypes.DailyPlan)
				plan.Date = "2026-07-28"
				plan.Summary.PlannedCount = 5
				plan.Summary.CompletedCount = 4
				plan.Summary.FailedCount = 1
				plan.Summary.AbandonedCount = 0
				plan.Summary.PendingRollbackCount = 1
				plan.Summary.AccountabilityScore = 82.5
				plan.Summary.BacklogPressure = 1.4
				return nil
			case protocol.MethodCheckInGet:
				query := params.(shareddto.DailyCheckInQuery)
				if query.Date != "2026-07-28" {
					t.Fatalf("unexpected check-in query: %#v", query)
				}
				checkIn := target.(*sharedtypes.DailyCheckIn)
				checkIn.Date = "2026-07-28"
				checkIn.Mood = 4
				checkIn.Energy = 3
				sleep := 7.5
				checkIn.SleepHours = &sleep
				screen := 90
				checkIn.ScreenTimeMinutes = &screen
				notes := "Clear head, low distraction"
				checkIn.Notes = &notes
				return nil
			case protocol.MethodMetricsRollup:
				query := params.(shareddto.DateRangeQuery)
				if query.Start != "2026-07-22" || query.End != "2026-07-28" {
					t.Fatalf("unexpected rollup query: %#v", query)
				}
				rollup := target.(*sharedtypes.MetricsRollup)
				avgMood := 3.8
				avgEnergy := 3.2
				avgSleep := 7.1
				rollup.AverageMood = &avgMood
				rollup.AverageEnergy = &avgEnergy
				rollup.AverageSleepHours = &avgSleep
				return nil
			case protocol.MethodMetricsStreaksLifetime:
				query := params.(shareddto.DailyCheckInQuery)
				if query.Date != "2026-07-28" {
					t.Fatalf("unexpected streak query: %#v", query)
				}
				streaks := target.(*sharedtypes.StreakSummary)
				streaks.CurrentFocusDays = 6
				streaks.LongestFocusDays = 11
				streaks.CurrentCheckInDays = 9
				streaks.LongestCheckInDays = 13
				streaks.CurrentHabitDays = 4
				streaks.LongestHabitDays = 7
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("summary run: %v", err)
	}

	for _, want := range []string{"DAILY SUMMARY", "AT A GLANCE", "AGENDA", "SIGNALS", "ACCOUNTABILITY", "WELLBEING", "MOMENTUM", "2026-07-28"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
		}
	}
	if len(gotMethods) != 8 {
		t.Fatalf("unexpected method count: %+v", gotMethods)
	}
}

func TestRunRangeUsesWindowQueries(t *testing.T) {
	var out bytes.Buffer
	var gotMethods []string
	err := Run([]string{"--start", "2026-07-01", "--end", "2026-07-07"}, Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			gotMethods = append(gotMethods, method)
			switch method {
			case protocol.MethodContextGet:
				ctx := target.(*sharedtypes.ActiveContext)
				repoID := int64(91)
				streamID := int64(92)
				issueID := int64(93)
				ctx.RepoID = &repoID
				ctx.StreamID = &streamID
				ctx.IssueID = &issueID
				return nil
			case protocol.MethodTimerGetState:
				timer := target.(*sharedtypes.TimerState)
				timer.State = "paused"
				return nil
			case protocol.MethodDashboardWindow:
				query := params.(shareddto.DashboardWindowQuery)
				if query.Start != "2026-07-01" || query.End != "2026-07-07" {
					t.Fatalf("unexpected dashboard query: %#v", query)
				}
				if query.RepoID == nil || *query.RepoID != 91 || query.StreamID == nil || *query.StreamID != 92 || query.IssueID == nil || *query.IssueID != 93 {
					t.Fatalf("expected context identifiers in dashboard query, got %#v", query)
				}
				window := target.(*sharedtypes.DashboardWindowSummary)
				window.StartDate = "2026-07-01"
				window.EndDate = "2026-07-07"
				window.PlannedCount = 8
				window.CompletedCount = 5
				window.FailedCount = 1
				window.AbandonedCount = 2
				window.MissedCount = 1
				window.CarryOverCount = 3
				window.AccountabilityScore = 62.5
				return nil
			case protocol.MethodMetricsRange:
				query := params.(shareddto.DateRangeQuery)
				if query.Start != "2026-07-01" || query.End != "2026-07-07" {
					t.Fatalf("unexpected range query: %#v", query)
				}
				days := target.(*[]sharedtypes.DailyMetricsDay)
				*days = []sharedtypes.DailyMetricsDay{
					{Date: "2026-07-01", WorkedSeconds: 60, RestSeconds: 20, SessionCount: 1, CompletedIssues: 1, TotalIssues: 2, HabitCompletedCount: 1, HabitDueCount: 2, HabitFailedCount: 0},
				}
				return nil
			case protocol.MethodMetricsRollup:
				query := params.(shareddto.DateRangeQuery)
				if query.Start != "2026-07-01" || query.End != "2026-07-07" {
					t.Fatalf("unexpected rollup query: %#v", query)
				}
				rollup := target.(*sharedtypes.MetricsRollup)
				rollup.Days = 7
				rollup.CheckInDays = 4
				rollup.FocusDays = 3
				rollup.WorkedSeconds = 3600
				rollup.RestSeconds = 900
				rollup.CompletedIssues = 5
				rollup.AbandonedIssues = 2
				rollup.HabitCompletedCount = 6
				rollup.HabitFailedCount = 1
				return nil
			case protocol.MethodMetricsStreaks:
				query := params.(shareddto.DateRangeQuery)
				if query.Start != "2026-07-01" || query.End != "2026-07-07" {
					t.Fatalf("unexpected streak query: %#v", query)
				}
				streaks := target.(*sharedtypes.StreakSummary)
				streaks.CurrentFocusDays = 2
				streaks.LongestFocusDays = 8
				streaks.CurrentCheckInDays = 5
				streaks.LongestCheckInDays = 9
				streaks.CurrentHabitDays = 3
				streaks.LongestHabitDays = 6
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("summary range run: %v", err)
	}

	for _, want := range []string{"SUMMARY", "AT A GLANCE", "OUTCOMES", "DAILY RHYTHM", "2026-07-01", "2026-07-07"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
		}
	}
	if len(gotMethods) != 6 {
		t.Fatalf("unexpected method count: %+v", gotMethods)
	}
}

func TestRunYesterdayPresetUsesPreviousDay(t *testing.T) {
	oldNowFn := nowFn
	nowFn = func() time.Time {
		return time.Date(2026, 7, 28, 9, 30, 0, 0, time.UTC)
	}
	t.Cleanup(func() { nowFn = oldNowFn })

	var out bytes.Buffer
	err := Run([]string{"--yesterday"}, Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			switch method {
			case protocol.MethodContextGet:
				return nil
			case protocol.MethodTimerGetState:
				return nil
			case protocol.MethodIssueDailySummary:
				query := params.(shareddto.DailyIssueSummaryQuery)
				if query.Date == nil || *query.Date != "2026-07-27" {
					t.Fatalf("unexpected yesterday query: %#v", query)
				}
				return nil
			case protocol.MethodHabitListDue:
				query := params.(shareddto.ListHabitsDueQuery)
				if query.Date != "2026-07-27" {
					t.Fatalf("unexpected yesterday habit query: %#v", query)
				}
				return nil
			case protocol.MethodDailyPlanGet, protocol.MethodCheckInGet, protocol.MethodMetricsRollup, protocol.MethodMetricsStreaksLifetime:
				return nil
			default:
				t.Fatalf("unexpected method: %s", method)
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("summary yesterday run: %v", err)
	}
}

func TestRunWeekAndMonthPresetsUseCalendarRanges(t *testing.T) {
	oldNowFn := nowFn
	nowFn = func() time.Time {
		return time.Date(2026, 7, 28, 9, 30, 0, 0, time.UTC)
	}
	t.Cleanup(func() { nowFn = oldNowFn })

	tests := []struct {
		name      string
		args      []string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "week",
			args:      []string{"--week"},
			wantStart: "2026-07-27",
			wantEnd:   "2026-07-28",
		},
		{
			name:      "month",
			args:      []string{"--month"},
			wantStart: "2026-07-01",
			wantEnd:   "2026-07-28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := Run(tt.args, Deps{
				Stdout: &out,
				CallKernel: func(method string, params, target any) error {
					switch method {
					case protocol.MethodContextGet, protocol.MethodTimerGetState:
						return nil
					case protocol.MethodDashboardWindow:
						query := params.(shareddto.DashboardWindowQuery)
						if query.Start != tt.wantStart || query.End != tt.wantEnd {
							t.Fatalf("unexpected dashboard query: %#v", query)
						}
						return nil
					case protocol.MethodMetricsRange, protocol.MethodMetricsRollup, protocol.MethodMetricsStreaks:
						query := params.(shareddto.DateRangeQuery)
						if query.Start != tt.wantStart || query.End != tt.wantEnd {
							t.Fatalf("unexpected date range query: %#v", query)
						}
						return nil
					default:
						return nil
					}
				},
			})
			if err != nil {
				t.Fatalf("summary %s run: %v", tt.name, err)
			}
		})
	}
}

func TestRunLastXDaysPresetUsesRollingRange(t *testing.T) {
	oldNowFn := nowFn
	nowFn = func() time.Time {
		return time.Date(2026, 7, 28, 9, 30, 0, 0, time.UTC)
	}
	t.Cleanup(func() { nowFn = oldNowFn })

	var out bytes.Buffer
	err := Run([]string{"--last-x-days", "7"}, Deps{
		Stdout: &out,
		CallKernel: func(method string, params, target any) error {
			switch method {
			case protocol.MethodContextGet, protocol.MethodTimerGetState:
				return nil
			case protocol.MethodDashboardWindow:
				query := params.(shareddto.DashboardWindowQuery)
				if query.Start != "2026-07-22" || query.End != "2026-07-28" {
					t.Fatalf("unexpected rolling dashboard query: %#v", query)
				}
				return nil
			case protocol.MethodMetricsRange, protocol.MethodMetricsRollup, protocol.MethodMetricsStreaks:
				query := params.(shareddto.DateRangeQuery)
				if query.Start != "2026-07-22" || query.End != "2026-07-28" {
					t.Fatalf("unexpected rolling date range query: %#v", query)
				}
				return nil
			default:
				return nil
			}
		},
	})
	if err != nil {
		t.Fatalf("summary last-x-days run: %v", err)
	}
}
