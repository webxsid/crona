package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"
)

func TestMomentumDetailsShowAllowedRangeForHabitSelection(t *testing.T) {
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		[]api.HabitWithMeta{
			{
				Habit: sharedtypes.Habit{
					ID:           1,
					Name:         "Train",
					ScheduleType: sharedtypes.HabitScheduleWeekly,
					Weekdays:     []int{1, 3, 5},
				},
			},
		},
		nil,
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Weekly train",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindHabit,
			MatchMode:     sharedtypes.MomentumMatchModeAny,
			Period:        sharedtypes.HabitStreakPeriodWeek,
			RequiredCount: 2,
			HabitIDs:      []int64{1},
		},
	)
	state.HabitStreakStep = 2

	rendered := renderHabitStreakDialog(testTheme(), state)
	if !strings.Contains(rendered, "Allowed range: 1-3 per week") {
		t.Fatalf("expected allowed range in dialog, got %q", rendered)
	}
}

func TestMomentumTargetsShowRedundancyWarningForRepoWideAndSpecificContexts(t *testing.T) {
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		nil,
		[]api.Repo{{ID: 7, Name: "Work"}},
		[]api.Stream{{ID: 8, RepoID: 7, Name: "App"}, {ID: 9, RepoID: 7, Name: "Infra"}},
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Delivery",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: 7},
				{RepoID: 7, StreamID: int64PtrTest(8)},
				{RepoID: 7, StreamID: int64PtrTest(9)},
			},
		},
	)
	state.HabitStreakStep = 1

	rendered := renderHabitStreakDialog(testTheme(), state)
	if !strings.Contains(rendered, "Warning: Work / Any stream already covers Work / App and Work / Infra") {
		t.Fatalf("expected redundancy warning in dialog, got %q", rendered)
	}
}

func TestMomentumTargetsSkipRedundancyWarningForDistinctContexts(t *testing.T) {
	state := controllerpkg.OpenEditMomentumDirect(
		controllerpkg.State{},
		nil,
		nil,
		[]api.Repo{{ID: 7, Name: "Work"}},
		[]api.Stream{{ID: 8, RepoID: 7, Name: "App"}},
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Delivery",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Contexts: []sharedtypes.MomentumContext{
				{RepoID: 7, StreamID: int64PtrTest(8)},
			},
		},
	)
	state.HabitStreakStep = 1

	rendered := renderHabitStreakDialog(testTheme(), state)
	if strings.Contains(rendered, "already covers") {
		t.Fatalf("did not expect redundancy warning, got %q", rendered)
	}
}

func testTheme() Theme {
	return Theme{}
}

func int64PtrTest(v int64) *int64 {
	return &v
}
