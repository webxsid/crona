package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
)

func restProtection(step int) controllerpkg.State {
	state := controllerpkg.OpenEditRestProtection(
		controllerpkg.State{},
		nil,
		[]int{0, 6},
		[]string{"2026-12-25"},
	)
	state.ProtectionStep = step
	return state
}

func dayBoundary(step int) controllerpkg.State {
	state := controllerpkg.OpenEditDayBoundary(
		controllerpkg.State{},
		sharedtypes.CoreSettingsKey("startOfDay"),
		sharedtypes.DayBoundarySchedule{},
		"Asia/Kolkata",
	)
	state.DayBoundaryStep = step
	return state
}

func telemetry(step int) controllerpkg.State {
	state := controllerpkg.OpenEditTelemetrySettings(controllerpkg.State{}, true, false)
	state.TelemetryStep = step
	return state
}

func TestSettingsDialogs(t *testing.T) {
	cases := []struct {
		c    formCase
		want []string
	}{
		{
			formCase{
				name:  "date format",
				state: controllerpkg.OpenEditDateDisplayFormat(controllerpkg.State{}, "YYYY-MM-DD"),
				title: "Custom Date Format",
			},
			[]string{"Pattern", "ctrl+s save"},
		},
		{
			formCase{
				name:    "protection streaks",
				state:   restProtection(0),
				title:   "Rest & Streak Protection",
				context: "1/4 Streaks",
			},
			[]string{"▶", "space toggle"},
		},
		{
			formCase{
				name:    "protection weekdays",
				state:   restProtection(1),
				title:   "Rest & Streak Protection",
				context: "2/4 Weekdays",
			},
			[]string{"[x] Sun"},
		},
		{
			formCase{
				name:    "protection dates",
				state:   restProtection(2),
				title:   "Rest & Streak Protection",
				context: "3/4 Dates",
			},
			[]string{"2026-12-25", "a add date"},
		},
		{
			formCase{
				name:    "protection review",
				state:   restProtection(3),
				title:   "Rest & Streak Protection",
				context: "4/4 Review",
			},
			[]string{"Weekdays", "ctrl+s save"},
		},
		{
			formCase{
				name:    "boundary default",
				state:   dayBoundary(controllerpkg.DayBoundaryStepOverview),
				title:   "Start of Day Schedule",
				context: "Asia/Kolkata",
			},
			[]string{"Enabled", "Default", "ctrl+s save"},
		},
		{
			formCase{
				name:  "boundary days",
				state: dayBoundary(controllerpkg.DayBoundaryStepDays),
				title: "Start of Day Schedule",
			},
			[]string{"Monday", "enter next", "esc back"},
		},
		{
			formCase{
				name:  "boundary time",
				state: dayBoundary(controllerpkg.DayBoundaryStepTime),
				title: "Start of Day Schedule",
			},
			[]string{"Time", "ctrl+s apply", "esc back"},
		},
		{
			formCase{
				name:    "telemetry usage",
				state:   telemetry(0),
				title:   "Privacy & Diagnostics",
				context: "1/3 Usage",
			},
			[]string{"[x] Share usage signals", "space toggle"},
		},
		{
			formCase{
				name:    "telemetry review",
				state:   telemetry(2),
				title:   "Privacy & Diagnostics",
				context: "3/3 Review",
			},
			[]string{"Usage signals", "Save and Restart Now", "enter choose"},
		},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
		if strings.Contains(out, "[tab]") || strings.Contains(out, "[space") ||
			strings.Contains(out, "[a] ") {
			t.Errorf("%s keeps bracketed keys:\n%s", tc.c.name, out)
		}
	}
}

func TestWizardsKeepOneWidthAcrossSteps(t *testing.T) {
	steps := map[string][]controllerpkg.State{
		"telemetry": {telemetry(0), telemetry(1), telemetry(2)},
		"rest protection": {
			restProtection(0),
			restProtection(1),
			restProtection(2),
			restProtection(3),
		},
		"day boundary": {
			dayBoundary(controllerpkg.DayBoundaryStepOverview),
			dayBoundary(controllerpkg.DayBoundaryStepDays),
			dayBoundary(controllerpkg.DayBoundaryStepTime),
		},
		"momentum": {editMomentum(1), editMomentum(2), editMomentum(3)},
	}
	for name, states := range steps {
		width := -1
		for i, state := range states {
			top := strings.Split(Render(testTheme(), sized(state, 120, 40)), "\n")[0]
			if got := lipgloss.Width(top); width >= 0 && got != width {
				t.Errorf("%s step %d width %d, want %d", name, i, got, width)
			} else {
				width = got
			}
		}
	}
}
