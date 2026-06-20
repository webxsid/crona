package controller

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMomentumDialogAllowsQInNameField(t *testing.T) {
	state := OpenEditMomentumDirect(State{}, nil, nil, nil, nil, nil, sharedtypes.HabitStreakDefinition{
		ID:            "momentum-1",
		Name:          "Focus",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	})
	state.HabitStreakStep = 2

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if !strings.HasSuffix(next.Inputs[0].Value(), "q") {
		t.Fatalf("expected q to be inserted into the name field, got %q", next.Inputs[0].Value())
	}
}

func TestMomentumDialogAllowsQInDescriptionField(t *testing.T) {
	state := OpenEditMomentumDirect(State{}, nil, nil, nil, nil, nil, sharedtypes.HabitStreakDefinition{
		ID:            "momentum-1",
		Name:          "Focus",
		Description:   ValueToPointer("Keep it steady"),
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	})
	state.HabitStreakStep = 2
	state.FocusIdx = 1
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if !strings.HasSuffix(next.Description.Value(), "q") {
		t.Fatalf("expected q to be inserted into the description field, got %q", next.Description.Value())
	}
}

func TestMomentumCreateUsesCtrlSToAdvanceKindStep(t *testing.T) {
	state := OpenCreateMomentumDirect(State{}, nil, nil, nil, nil, nil)

	next, action, status := Update(state, UpdateContext{}, "2026-05-26", tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.HabitStreakStep != 2 {
		t.Fatalf("expected ctrl+s to advance to target step, got %d", next.HabitStreakStep)
	}
}

func TestMomentumDetailsUseCtrlSToAdvanceToReview(t *testing.T) {
	state := OpenEditMomentumDirect(State{}, nil, []api.HabitWithMeta{{
		Habit: sharedtypes.Habit{
			ID:           1,
			Name:         "Focus",
			ScheduleType: sharedtypes.HabitScheduleDaily,
		},
	}}, nil, nil, nil, sharedtypes.HabitStreakDefinition{
		ID:            "momentum-1",
		Name:          "Focus",
		Enabled:       true,
		TargetKind:    sharedtypes.MomentumTargetKindHabit,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
		HabitIDs:      []int64{1},
	})
	state.HabitStreakStep = 2

	next, action, status := Update(state, UpdateContext{}, "2026-05-26", tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.HabitStreakStep != 3 {
		t.Fatalf("expected ctrl+s to advance to review, got %d", next.HabitStreakStep)
	}
}

func TestMomentumDetailsShiftTabReturnsToTargets(t *testing.T) {
	state := OpenEditMomentumDirect(State{}, nil, nil, nil, nil, nil, sharedtypes.HabitStreakDefinition{
		ID:            "momentum-1",
		Name:          "Focus",
		Enabled:       true,
		Period:        sharedtypes.HabitStreakPeriodDay,
		RequiredCount: 1,
	})
	state.HabitStreakStep = 2
	state.FocusIdx = 0
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyShiftTab},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.HabitStreakStep != 1 {
		t.Fatalf("expected shift+tab to return to targets, got step %d", next.HabitStreakStep)
	}
}

func TestMomentumContextTargetsAllowRepoWideSelection(t *testing.T) {
	repos := []api.Repo{{ID: 7, Name: "Work"}}
	streams := []api.Stream{{ID: 8, RepoID: 7, Name: "App"}}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		streams,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Focus",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 1,
		},
	)
	state.HabitStreakStep = 1
	state.MomentumRepoInput.SetValue("Work")
	state.RepoIndex = 0
	state.StreamIndex = -1

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if len(next.HabitStreakDraft.Contexts) != 1 {
		t.Fatalf("expected one selected context, got %+v", next.HabitStreakDraft.Contexts)
	}
	if next.HabitStreakDraft.Contexts[0].RepoID != 7 || next.HabitStreakDraft.Contexts[0].StreamID != nil {
		t.Fatalf("expected repo-wide context, got %+v", next.HabitStreakDraft.Contexts[0])
	}
}

func TestMomentumContextTargetsLetPrintableHTypeIntoRepoSearch(t *testing.T) {
	repos := []api.Repo{{ID: 7, Name: "Work"}}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Focus",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Period:     sharedtypes.HabitStreakPeriodDay,
		},
	)
	state.HabitStreakStep = 1
	state.FocusIdx = int(contextTargetFocusRepo)
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.MomentumRepoInput.Value() != "h" {
		t.Fatalf("expected h to be inserted into repo search, got %q", next.MomentumRepoInput.Value())
	}
	if next.RepoIndex != 0 {
		t.Fatalf("expected repo selection to stay on first filtered result, got %d", next.RepoIndex)
	}
}

func TestMomentumContextTargetsLetPrintableHTypeIntoDurationInput(t *testing.T) {
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		[]api.Repo{{ID: 7, Name: "Work"}},
		[]api.Stream{{ID: 8, RepoID: 7, Name: "App"}},
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Focus",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 3600,
		},
	)
	state.HabitStreakStep = 2
	state = habitStreakSetDetailFocus(state, habitStreakDetailRowCount)
	state.Inputs[1].SetValue("")

	for _, r := range []rune("1h34m23s") {
		next, action, status := Update(
			state,
			UpdateContext{},
			"2026-05-26",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}},
		)
		if status != "" {
			t.Fatalf("unexpected status %q", status)
		}
		if action != nil {
			t.Fatalf("expected no action, got %+v", action)
		}
		state = next
	}
	if got := state.Inputs[1].Value(); got != "1h34m23s" {
		t.Fatalf("expected duration text to be accepted, got %q", got)
	}
}

func TestMomentumDetailsAcceptDurationAndKeepDailyThreshold(t *testing.T) {
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		[]api.Repo{{ID: 7, Name: "Work"}},
		[]api.Stream{{ID: 8, RepoID: 7, Name: "App"}},
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Focus",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindContext,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 3600,
		},
	)
	state.HabitStreakStep = 2
	state = habitStreakSetDetailFocus(state, habitStreakDetailRowCount)
	state.Inputs[1].SetValue("1h34m23s")

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.HabitStreakDraft.RequiredCount != 5663 {
		t.Fatalf("expected duration to be stored as seconds, got %d", next.HabitStreakDraft.RequiredCount)
	}
	if next.HabitStreakDraft.Period != sharedtypes.HabitStreakPeriodDay {
		t.Fatalf("expected daily period to remain unchanged, got %q", next.HabitStreakDraft.Period)
	}
}

func TestMomentumDetailsBlockInvalidDailyAllHabitSelection(t *testing.T) {
	state := OpenEditMomentumDirect(
		State{},
		nil,
		[]api.HabitWithMeta{
			{
				Habit: sharedtypes.Habit{
					ID:           1,
					Name:         "Journal",
					ScheduleType: sharedtypes.HabitScheduleDaily,
				},
			},
			{
				Habit: sharedtypes.Habit{
					ID:           2,
					Name:         "Walk",
					ScheduleType: sharedtypes.HabitScheduleWeekdays,
				},
			},
		},
		nil,
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:            "momentum-1",
			Name:          "Daily all",
			Enabled:       true,
			TargetKind:    sharedtypes.MomentumTargetKindHabit,
			MatchMode:     sharedtypes.MomentumMatchModeAll,
			Period:        sharedtypes.HabitStreakPeriodDay,
			RequiredCount: 1,
			HabitIDs:      []int64{1, 2},
		},
	)
	state.HabitStreakStep = 2

	next, action, status := Update(state, UpdateContext{}, "2026-05-26", tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.HabitStreakStep != 2 {
		t.Fatalf("expected to stay on details step, got %d", next.HabitStreakStep)
	}
	if !strings.Contains(next.ErrorMessage, "repeat daily") {
		t.Fatalf("expected daily repeat validation, got %q", next.ErrorMessage)
	}
}

func TestMomentumContextTargetsUseSelectedStreamIndexEvenWhenSearchIsBlank(t *testing.T) {
	repos := []api.Repo{{ID: 7, Name: "Work"}}
	streams := []api.Stream{
		{ID: 8, RepoID: 7, Name: "App"},
		{ID: 9, RepoID: 7, Name: "Docs"},
	}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		streams,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Focus",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Period:     sharedtypes.HabitStreakPeriodDay,
		},
	)
	state.HabitStreakStep = 1
	state.FocusIdx = int(contextTargetFocusStream)
	state.RepoIndex = 0
	state.StreamIndex = 1
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if len(next.HabitStreakDraft.Contexts) != 1 {
		t.Fatalf("expected one selected context, got %+v", next.HabitStreakDraft.Contexts)
	}
	if next.HabitStreakDraft.Contexts[0].RepoID != 7 {
		t.Fatalf("expected repo id 7, got %+v", next.HabitStreakDraft.Contexts[0])
	}
	if next.HabitStreakDraft.Contexts[0].StreamID == nil || *next.HabitStreakDraft.Contexts[0].StreamID != 9 {
		t.Fatalf("expected selected stream id 9, got %+v", next.HabitStreakDraft.Contexts[0].StreamID)
	}
}

func TestMomentumContextRepoSearchFiltersCandidates(t *testing.T) {
	repos := []api.Repo{
		{ID: 7, Name: "Work"},
		{ID: 8, Name: "Personal"},
	}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		nil,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Focus",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Period:     sharedtypes.HabitStreakPeriodDay,
		},
	)
	state.HabitStreakStep = 1
	state.MomentumRepoInput.SetValue("Per")
	state.RepoIndex = 0

	options := momentumRepoOptions(state)
	if len(options) != 1 {
		t.Fatalf("expected one filtered repo candidate, got %+v", options)
	}
	if options[0].Label != "Personal" {
		t.Fatalf("expected Personal to match repo search, got %+v", options[0])
	}
}

func TestMomentumContextStreamSearchFiltersCandidates(t *testing.T) {
	repos := []api.Repo{{ID: 7, Name: "Work"}}
	streams := []api.Stream{
		{ID: 8, RepoID: 7, Name: "App"},
		{ID: 9, RepoID: 7, Name: "Docs"},
	}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		streams,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Focus",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Period:     sharedtypes.HabitStreakPeriodDay,
		},
	)
	state.HabitStreakStep = 1
	state.MomentumRepoInput.SetValue("Work")
	state.RepoIndex = 0
	state.MomentumStreamInput.SetValue("Ap")
	state.StreamIndex = 0

	options := momentumStreamOptions(state)
	if len(options) != 2 {
		t.Fatalf("expected Any Stream plus one filtered candidate, got %+v", options)
	}
	if options[0].Label != "Any Stream" {
		t.Fatalf("expected Any Stream to remain first, got %+v", options[0])
	}
	if options[1].Label != "App" {
		t.Fatalf("expected App to match stream search, got %+v", options[1])
	}
}

func TestMomentumDialogLabelsUseMomentumSearchInputs(t *testing.T) {
	repos := []api.Repo{
		{ID: 7, Name: "Work"},
		{ID: 8, Name: "Personal"},
	}
	streams := []api.Stream{
		{ID: 9, RepoID: 8, Name: "Mobile"},
	}
	repoInput := textinput.New()
	repoInput.SetValue("Per")
	streamInput := textinput.New()
	streamInput.SetValue("Mob")

	repoLabel, streamLabel := MomentumDialogLabels(
		[]textinput.Model{repoInput, streamInput},
		0,
		0,
		repos,
		nil,
		streams,
		nil,
	)

	if repoLabel != "Personal" {
		t.Fatalf("expected repo search to resolve Personal, got %q", repoLabel)
	}
	if streamLabel != "Mobile" {
		t.Fatalf("expected stream search to resolve Mobile, got %q", streamLabel)
	}
}

func TestMomentumDialogLabelsHandleClearedSearchWithoutPanic(t *testing.T) {
	repos := []api.Repo{{ID: 7, Name: "Work"}}
	repoInput := textinput.New()
	streamInput := textinput.New()

	repoLabel, streamLabel := MomentumDialogLabels(
		[]textinput.Model{repoInput, streamInput},
		-1,
		-1,
		repos,
		nil,
		nil,
		nil,
	)

	if repoLabel != "Type to search" {
		t.Fatalf("expected cleared repo search to render placeholder, got %q", repoLabel)
	}
	if streamLabel != "Any Stream" {
		t.Fatalf("expected cleared stream search to render Any Stream, got %q", streamLabel)
	}
}

func TestMomentumContextTargetsCycleRepoAndStreamHorizontally(t *testing.T) {
	repos := []api.Repo{
		{ID: 7, Name: "Work"},
		{ID: 8, Name: "Personal"},
	}
	streams := []api.Stream{
		{ID: 9, RepoID: 7, Name: "App"},
		{ID: 10, RepoID: 7, Name: "Docs"},
	}
	state := OpenEditMomentumDirect(
		State{},
		nil,
		nil,
		repos,
		streams,
		nil,
		sharedtypes.HabitStreakDefinition{
			ID:         "momentum-1",
			Name:       "Focus",
			Enabled:    true,
			TargetKind: sharedtypes.MomentumTargetKindContext,
			Period:     sharedtypes.HabitStreakPeriodDay,
		},
	)
	state.HabitStreakStep = 1
	state.FocusIdx = 0
	state.RepoIndex = 0
	state.StreamIndex = 0
	state = SyncDialogFocus(state)

	next, action, status := Update(
		state,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRight},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if next.RepoIndex != 1 {
		t.Fatalf("expected repo selection to cycle horizontally, got %d", next.RepoIndex)
	}

	streamState := state
	streamState.FocusIdx = 1
	streamState = SyncDialogFocus(streamState)
	streamState, action, status = Update(
		streamState,
		UpdateContext{},
		"2026-05-26",
		tea.KeyMsg{Type: tea.KeyRight},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected no action, got %+v", action)
	}
	if streamState.StreamIndex != 1 {
		t.Fatalf("expected stream selection to cycle horizontally, got %d", streamState.StreamIndex)
	}
}
