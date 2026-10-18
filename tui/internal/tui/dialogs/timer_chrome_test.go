package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

func TestTimerLifecycleDialogs(t *testing.T) {
	est := 60
	base := controllerpkg.State{}
	adhocKeep := controllerpkg.OpenAdhocEndSession(controllerpkg.State{AdhocElapsedSeconds: 91}, sharedtypes.AdhocSessionResolutionKeep, 0, 0, 0, "")
	transition := controllerpkg.OpenIssueSessionTransition(base, 4, "done")
	cases := []struct {
		c    formCase
		want []string
	}{
		{formCase{name: "end_session", state: controllerpkg.OpenSessionMessage(base, "end_session"), title: "End Session"}, []string{"Commit", "ctrl+s confirm", "ctrl+e details"}},
		{formCase{name: "end_session adhoc", state: adhocKeep, title: "End Session", context: "Elapsed 1m 31s"}, []string{"Summary", "Ad hoc / Sessions", "save ad hoc & finish"}},
		{formCase{name: "adhoc_resolution", state: controllerpkg.OpenAdhocResolution(controllerpkg.State{AdhocElapsedSeconds: 91}, nil), title: "Finish Ad hoc Session", context: "Elapsed 1m 31s"}, []string{"▶", "Keep as ad hoc", "enter choose", "esc back"}},
		{formCase{name: "adhoc_new_issue", state: controllerpkg.OpenAdhocNewIssue(base), title: "Create Issue"}, []string{"Title", "enter continue"}},
		{formCase{name: "timer_start_type", state: controllerpkg.OpenTimerStartType(base, 1, 2, 3, "Fix alerts", &est, 900), title: "Start Timer", context: "Fix alerts"}, []string{"▶", "Worked / Est", "enter choose"}},
		{formCase{name: "pomodoro_start", state: controllerpkg.OpenPomodoroStart(base, 1, 2, 3, "Fix alerts", &est, 900), title: "Pomodoro Session", context: "Fix alerts"}, []string{"Focus", "Short break", "Cycles", "ctrl+s start", "←/→ choose"}},
		{formCase{name: "timer_countdown_start", state: controllerpkg.OpenSingleTimerStart(base, 1, 2, 3, "Fix alerts", &est, 900), title: "Timer Session", context: "Fix alerts"}, []string{"Focus", "ctrl+s start"}},
		{formCase{name: "hard_limit_extend", state: controllerpkg.OpenHardLimitExtend(controllerpkg.OpenPomodoroStart(base, 1, 2, 3, "Fix alerts", &est, 900)), title: "Extend Pomodoro Session"}, []string{"Focus", "enter extend"}},
		{formCase{name: "timer_countdown_extend", state: controllerpkg.OpenCountdownExtend(base), title: "Extend Timer Session"}, []string{"Add", "ctrl+s extend"}},
		{formCase{name: "amend_session", state: controllerpkg.OpenAmendSession(base, "s1", "old message"), title: "Amend Session"}, []string{"Commit", "ctrl+s save"}},
		{formCase{name: "manual_session", state: controllerpkg.OpenManualSession(base, 3, "Fix alerts", &est, "2026-10-18"), title: "Log Session", context: "Fix alerts"}, []string{"Summary", "Date", "Work", "Break", "ctrl+s save"}},
		{formCase{name: "issue_session_transition", state: transition, title: "Complete Issue"}, []string{"Note", "ctrl+s confirm"}},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
		if bracketKey.MatchString(strings.ReplaceAll(out, "[s] ", "")) && tc.c.name != "timer_start_type" {
			t.Errorf("%s keeps bracketed keys:\n%s", tc.c.name, out)
		}
	}
	expired := controllerpkg.OpenHardLimitExpired(base, "Fix alerts")
	out := ansi.Strip(Render(testTheme(), sized(expired, 120, 36)))
	if !strings.HasPrefix(strings.Split(out, "\n")[0], "╭─ Pomodoro Session Complete") || !strings.Contains(out, "enter choose") || strings.Contains(out, "[enter]") {
		t.Errorf("hard_limit_expired not on the chrome:\n%s", out)
	}
}
