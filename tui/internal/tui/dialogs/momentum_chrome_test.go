package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

func editMomentum(step int) controllerpkg.State {
	state := controllerpkg.OpenEditMomentumDirect(controllerpkg.State{}, nil, nil, nil, nil, nil,
		sharedtypes.HabitStreakDefinition{
			ID: "momentum-1", Name: "Recovery Mix", Enabled: true,
			TargetKind: sharedtypes.MomentumTargetKindHabit, Period: sharedtypes.HabitStreakPeriodWeek, RequiredCount: 2,
		})
	state.HabitStreakStep = step
	return state
}

func TestMomentumWizardChrome(t *testing.T) {
	create := controllerpkg.OpenCreateMomentumDirect(controllerpkg.State{}, nil, nil, nil, nil, nil)
	cases := []struct {
		c    formCase
		want []string
	}{
		{formCase{name: "kind", state: create, title: "New Momentum", context: "1/4 Kind"}, []string{"Choose momentum kind", "▶", "enter choose"}},
		{formCase{name: "targets", state: editMomentum(1), title: "Edit Momentum", context: "1/3 Targets"}, []string{"Select contributing habits", "Match", "ctrl+s details"}},
		{formCase{name: "details", state: editMomentum(2), title: "Edit Momentum", context: "2/3 Details"}, []string{"Name", "Period", "Required", "per week", "ctrl+s save momentum"}},
		{formCase{name: "review", state: editMomentum(3), title: "Edit Momentum", context: "3/3 Review"}, []string{"Name", "Recovery Mix", "Rule"}},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
		if strings.Contains(out, "(Optional)") || strings.Contains(out, "[ > ") {
			t.Errorf("%s keeps old chrome:\n%s", tc.c.name, out)
		}
	}
}

func TestMomentumReviewRuleReadsPerWeek(t *testing.T) {
	out := ansi.Strip(Render(testTheme(), sized(editMomentum(3), 120, 36)))
	if !strings.Contains(out, "per week") || strings.Contains(out, "per weekly") {
		t.Fatalf("rule should read per week:\n%s", out)
	}
}
