package controller

import (
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAdhocTimerStartModesKeepAdhocMarker(t *testing.T) {
	state := OpenAdhocTimerStartType(State{})
	state.ChoiceCursor = 0
	_, action, status := updateTimerStartType(state, tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" || action == nil || action.TimerStart == nil {
		t.Fatalf("stopwatch start action = %+v, status %q", action, status)
	}
	if !action.TimerStart.Adhoc || action.TimerStart.IssueID != nil {
		t.Fatalf("stopwatch ad hoc request = %+v, want ad hoc with no issue", action.TimerStart)
	}

	state = OpenAdhocTimerStartType(State{})
	state.ChoiceCursor = 1
	state, _, status = updateTimerStartType(state, tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" || state.Kind != "pomodoro_start" || !state.AdhocTimer {
		t.Fatalf("pomodoro start state = %+v, status %q", state, status)
	}
	_, action, status = submitPomodoroStart(state)
	if status != "" || action == nil || action.TimerStart == nil || !action.TimerStart.Adhoc {
		t.Fatalf("pomodoro ad hoc request = %+v, status %q", action, status)
	}

	state = OpenAdhocTimerStartType(State{})
	state.ChoiceCursor = 2
	state, _, status = updateTimerStartType(state, tea.KeyMsg{Type: tea.KeyEnter})
	if status != "" || state.Kind != "timer_countdown_start" || !state.AdhocTimer {
		t.Fatalf("countdown start state = %+v, status %q", state, status)
	}
	_, action, status = updateTimerCountdownStart(state, tea.KeyMsg{Type: tea.KeyCtrlS})
	if status != "" || action == nil || action.TimerStart == nil || !action.TimerStart.Adhoc {
		t.Fatalf("countdown ad hoc request = %+v, status %q", action, status)
	}
}

func TestAdhocResolutionKeepBuildsEndPayload(t *testing.T) {
	state := OpenAdhocResolution(State{}, nil)
	state.ChoiceCursor = 2
	state, _, status := updateAdhocResolution(
		state,
		UpdateContext{},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || state.Kind != "end_session" ||
		state.AdhocResolution != sharedtypes.AdhocSessionResolutionKeep {
		t.Fatalf("keep resolution state = %+v, status %q", state, status)
	}
	state.Inputs[0].SetValue("miscellaneous work")
	_, action, status := updateSessionMessage(
		state,
		UpdateContext{},
		"2026-05-24",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" || action == nil {
		t.Fatalf("keep end action = %+v, status %q", action, status)
	}
	if action.Payload.AdhocResolution != sharedtypes.AdhocSessionResolutionKeep ||
		action.Payload.CommitMessage == nil || *action.Payload.CommitMessage != "miscellaneous work" {
		t.Fatalf("keep end payload = %+v", action.Payload)
	}
}

func TestEndSessionRejectsEmptyCommitMessageWithoutClosing(t *testing.T) {
	state := OpenSessionMessage(State{}, "end_session")
	next, action, status := updateSessionMessage(
		state,
		UpdateContext{HasActiveIssue: true, ActiveIssueStream: 5},
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "Commit message is required" || action != nil || next.Kind != "end_session" {
		t.Fatalf(
			"empty commit submission = state %q action %+v status %q; want dialog open with inline error",
			next.Kind,
			action,
			status,
		)
	}
}

func TestAdhocResolutionAttachBuildsTargetPayload(t *testing.T) {
	issues := []api.IssueWithMeta{
		{
			Issue: sharedtypes.Issue{
				ID:       41,
				StreamID: 5,
				Title:    "Eligible",
				Status:   sharedtypes.IssueStatusPlanned,
			},
			RepoID:     2,
			RepoName:   "Work",
			StreamName: "App",
		},
		{
			Issue: sharedtypes.Issue{
				ID:       42,
				StreamID: 6,
				Title:    "Done",
				Status:   sharedtypes.IssueStatusDone,
			},
			RepoID:     3,
			RepoName:   "Other",
			StreamName: "Archive",
		},
	}
	state := OpenAdhocIssuePicker(State{}, issues, "2026-10-08")
	if len(state.AdhocIssueIDs) != 1 || state.AdhocIssueIDs[0] != 41 || state.ChoiceCursor != 0 {
		t.Fatalf(
			"issue picker choices = %+v IDs=%+v cursor=%d, want eligible issue 41",
			state.ChoiceItems,
			state.AdhocIssueIDs,
			state.ChoiceCursor,
		)
	}
	state, _, status := updateAdhocIssuePicker(
		state,
		UpdateContext{AllIssues: issues},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || state.Kind != "end_session" || state.IssueID != 41 {
		t.Fatalf("selected issue state = %+v, status %q", state, status)
	}
	state.Inputs[0].SetValue("issue work")
	_, action, status := updateSessionMessage(
		state,
		UpdateContext{},
		"2026-05-24",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" || action == nil {
		t.Fatalf("attach end action = %+v, status %q", action, status)
	}
	if action.Payload.AdhocResolution != sharedtypes.AdhocSessionResolutionAttach ||
		action.Payload.IssueID == nil || *action.Payload.IssueID != 41 ||
		action.Payload.RepoID == nil || *action.Payload.RepoID != 2 ||
		action.Payload.StreamID == nil || *action.Payload.StreamID != 5 {
		t.Fatalf("attach end payload = %+v", action.Payload)
	}
}

func TestAdhocIssuePickerTabsFilterAndPreserveSelection(t *testing.T) {
	today, overdue, future := "2026-10-08", "2026-10-07", "2026-10-10"
	issues := []api.IssueWithMeta{
		{
			Issue: sharedtypes.Issue{
				ID:          1,
				Title:       "Due today",
				Status:      sharedtypes.IssueStatusPlanned,
				TodoForDate: &today,
			},
		},
		{
			Issue: sharedtypes.Issue{
				ID:          2,
				Title:       "Overdue",
				Status:      sharedtypes.IssueStatusPlanned,
				TodoForDate: &overdue,
			},
		},
		{
			Issue: sharedtypes.Issue{
				ID:          3,
				Title:       "Pinned",
				Status:      sharedtypes.IssueStatusReady,
				PinnedDaily: true,
			},
		},
		{
			Issue: sharedtypes.Issue{
				ID:          4,
				Title:       "Future",
				Status:      sharedtypes.IssueStatusPlanned,
				TodoForDate: &future,
			},
		},
	}
	state := OpenAdhocIssuePicker(State{}, issues, today)
	if len(state.ChoiceItems) != 1 || state.ChoiceItems[0] != "Due today" ||
		state.AdhocIssueTabCounts != [4]int{1, 1, 1, 4} {
		t.Fatalf(
			"Due today tab items/counts = %+v / %+v, want its issue and [1 1 1 4]",
			state.ChoiceItems,
			state.AdhocIssueTabCounts,
		)
	}
	ctx := UpdateContext{AllIssues: issues, CurrentDate: today}
	state.ChoiceCursor = 0
	next, _, _ := updateAdhocIssuePicker(state, ctx, tea.KeyMsg{Type: tea.KeyRight})
	if next.AdhocIssueTab != 1 || len(next.ChoiceItems) != 1 || next.ChoiceItems[0] != "Overdue" {
		t.Fatalf(
			"Overdue tab = %d %+v, want only overdue issue",
			next.AdhocIssueTab,
			next.ChoiceItems,
		)
	}
	next, _, _ = updateAdhocIssuePicker(next, ctx, tea.KeyMsg{Type: tea.KeyRight})
	if next.AdhocIssueTab != 2 || len(next.ChoiceItems) != 1 || next.ChoiceItems[0] != "Pinned" {
		t.Fatalf(
			"Pinned tab = %d %+v, want only pinned issue",
			next.AdhocIssueTab,
			next.ChoiceItems,
		)
	}
	next, _, _ = updateAdhocIssuePicker(next, ctx, tea.KeyMsg{Type: tea.KeyRight})
	if next.AdhocIssueTab != 3 || len(next.ChoiceItems) != 4 {
		t.Fatalf("All tab = %d %+v, want all four issues", next.AdhocIssueTab, next.ChoiceItems)
	}
	next, _, _ = updateAdhocIssuePicker(next, ctx, tea.KeyMsg{Type: tea.KeyRight})
	if next.AdhocIssueTab != 0 || len(next.ChoiceItems) != 1 || next.ChoiceItems[0] != "Due today" {
		t.Fatalf(
			"tab wrap = %d %+v, want Due today with its issue",
			next.AdhocIssueTab,
			next.ChoiceItems,
		)
	}
}

func TestAdhocResolutionCreateOpensStandardIssueDialogWithPrefills(t *testing.T) {
	repoID, streamID := int64(2), int64(5)
	repoName, streamName := "Work", "App"
	state := OpenAdhocResolution(
		State{AdhocElapsedSeconds: 31 * 60, AdhocDate: "2026-10-08"},
		&api.ActiveContext{
			RepoID: &repoID, StreamID: &streamID, RepoName: &repoName, StreamName: &streamName,
		},
	)
	state.ChoiceCursor = 1
	state, _, status := updateAdhocResolution(
		state,
		UpdateContext{},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" || state.Kind != "create_issue_default" {
		t.Fatalf("issue creation state kind = %q, status %q", state.Kind, status)
	}
	if state.Inputs[0].Value() != "Work" || state.Inputs[1].Value() != "App" ||
		state.Inputs[3].Value() != "45m" || state.Inputs[4].Value() != "2026-10-08" {
		t.Fatalf(
			"issue dialog prefills = repo:%q stream:%q estimate:%q date:%q",
			state.Inputs[0].Value(),
			state.Inputs[1].Value(),
			state.Inputs[3].Value(),
			state.Inputs[4].Value(),
		)
	}
	state.Inputs[2].SetValue("New issue")
	state.Inputs[5].SetValue("Implemented the requested change")
	state.RepoIndex, state.StreamIndex = 0, 0
	ctx := UpdateContext{
		Context: &api.ActiveContext{RepoName: &repoName, StreamName: &streamName},
		Repos:   []api.Repo{{ID: repoID, Name: repoName}},
		Streams: []api.Stream{{ID: streamID, RepoID: repoID, Name: streamName}},
	}
	_, action, status := updateCreateIssueDefault(
		state,
		ctx,
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" || action == nil || action.Kind != "create_issue_default" ||
		action.IssueCreateFollowUp != IssueCreateFollowUpAdhoc {
		t.Fatalf("issue creation action = %+v, status %q", action, status)
	}
	if action.Title != "New issue" || action.Estimate == nil || *action.Estimate != 45 ||
		action.DueDate == nil || *action.DueDate != "2026-10-08" {
		t.Fatalf("created issue attributes = %+v", action)
	}
	if action.CommitMessage != "Implemented the requested change" {
		t.Fatalf("session summary = %q, want preserved summary", action.CommitMessage)
	}
}

func TestPromoteAdhocIssueOpensExistingContextFormAndSubmitsConversion(t *testing.T) {
	issue := api.IssueWithMeta{Issue: sharedtypes.Issue{
		ID: 73, Title: "Session summary", Status: sharedtypes.IssueStatusDone, Adhoc: true,
	}}
	state := OpenPromoteAdhocIssue(State{}, issue, nil)
	if !state.PromotingAdhoc || state.IssueID != issue.ID ||
		state.Inputs[2].Value() != issue.Title {
		t.Fatalf("conversion dialog state = %+v", state)
	}
	state.Inputs[0].SetValue("Work")
	state.Inputs[1].SetValue("App")
	context := UpdateContext{
		Repos:   []api.Repo{{ID: 2, Name: "Work"}},
		Streams: []api.Stream{{ID: 5, RepoID: 2, Name: "App"}},
	}
	next, action, status := updateCreateIssueDefault(
		state,
		context,
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if status != "" || action == nil || action.Kind != "promote_adhoc_issue" ||
		action.IssueID != issue.ID ||
		action.RepoName != "Work" ||
		action.StreamName != "App" {
		t.Fatalf("conversion submit = state %+v action %+v status %q", next, action, status)
	}
}

func TestPromoteAdhocIssuePreservesHistoricalDate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		date  string
		valid bool
	}{
		{name: "original historical date", date: "2026-10-07", valid: true},
		{name: "changed historical date", date: "2026-10-06"},
		{name: "today", date: "2026-10-08", valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			originalDate := "2026-10-07"
			issue := api.IssueWithMeta{Issue: sharedtypes.Issue{
				ID: 73, Title: "Session summary", Status: sharedtypes.IssueStatusDone,
				Adhoc: true, TodoForDate: &originalDate,
			}}
			state := OpenPromoteAdhocIssue(State{}, issue, nil)
			state.Inputs[0].SetValue("Work")
			state.Inputs[1].SetValue("App")
			state.Inputs[4].SetValue(tc.date)
			context := UpdateContext{
				Repos:   []api.Repo{{ID: 2, Name: "Work"}},
				Streams: []api.Stream{{ID: 5, RepoID: 2, Name: "App"}},
			}
			next, action, status := updateCreateIssueDefault(
				state,
				context,
				"2026-10-08",
				tea.KeyMsg{Type: tea.KeyCtrlS},
			)
			if !tc.valid {
				if action != nil || status == "" || next.Kind != state.Kind {
					t.Fatalf(
						"expected date rejection with open dialog, got action %+v status %q kind %q",
						action,
						status,
						next.Kind,
					)
				}
				return
			}
			if status != "" || action == nil || action.DueDate == nil ||
				*action.DueDate != tc.date {
				t.Fatalf(
					"expected conversion preserving %s, got action %+v status %q",
					tc.date,
					action,
					status,
				)
			}
		})
	}
}

func TestAdhocIssueDetailOffersConversionAction(t *testing.T) {
	state := OpenViewEntity(
		State{},
		"Issue",
		"Session title",
		"Ad hoc / Sessions",
		"Reviewed session",
	)
	state.ViewEntityAdhoc = true
	state.IssueID = 73
	next, action, status := updateViewEntity(
		state,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
	)
	if status != "" || action == nil || action.Kind != "open_adhoc_conversion" ||
		action.IssueID != 73 ||
		next.Kind != "" {
		t.Fatalf("detail conversion action = state %+v action %+v status %q", next, action, status)
	}
}

func TestEmptyTodayTabRemainsReachable(t *testing.T) {
	issues := []api.IssueWithMeta{
		{Issue: sharedtypes.Issue{ID: 1, Title: "Future", Status: sharedtypes.IssueStatusPlanned}},
	}
	state := OpenAdhocIssuePicker(State{}, issues, "2026-10-08")
	if state.AdhocIssueTab != 3 {
		t.Fatal("initial empty today should fall back to All")
	}
	ctx := UpdateContext{AllIssues: issues, CurrentDate: "2026-10-08"}
	for _, want := range []int{0, 1, 2, 3} {
		state = switchAdhocIssueTab(state, ctx, 1)
		if state.AdhocIssueTab != want {
			t.Fatalf("tab %d, want %d", state.AdhocIssueTab, want)
		}
	}
	for _, want := range []int{2, 1, 0, 3} {
		state = switchAdhocIssueTab(state, ctx, -1)
		if state.AdhocIssueTab != want {
			t.Fatalf("tab %d, want %d", state.AdhocIssueTab, want)
		}
	}
}
func TestConversionIgnoresCreationShortcuts(t *testing.T) {
	state := OpenPromoteAdhocIssue(
		State{},
		api.IssueWithMeta{Issue: sharedtypes.Issue{ID: 73, Title: "Summary", Adhoc: true}},
		nil,
	)
	for _, key := range []tea.KeyType{tea.KeyCtrlA, tea.KeyCtrlF} {
		next, action, status := updateCreateIssueDefault(
			state,
			UpdateContext{},
			"2026-10-08",
			tea.KeyMsg{Type: key},
		)
		if action != nil || status != "" || next.Kind != state.Kind {
			t.Fatalf("unexpected conversion shortcut: %+v %q", action, status)
		}
	}
}
func TestAdhocAdvancedDetailsKeepDestination(t *testing.T) {
	state := OpenAdhocEndSession(
		State{AdhocElapsedSeconds: 123},
		sharedtypes.AdhocSessionResolutionAttach,
		41,
		2,
		3,
		"",
	)
	state.ViewName = "Title"
	state.RepoName = "Work"
	state.StreamName = "App"
	state.Inputs[0].SetValue("Summary")
	state = ToggleEndSessionAdvanced(ToggleEndSessionAdvanced(state))
	if state.IssueID != 41 || state.RepoID != 2 || state.StreamID != 3 ||
		state.ViewName != "Title" ||
		state.AdhocElapsedSeconds != 123 ||
		state.Inputs[0].Value() != "Summary" {
		t.Fatalf("lost confirmation state: %+v", state)
	}
}

func TestAdhocPickerNavigationAndSelectedIssuePreservation(t *testing.T) {
	today := "2026-10-08"
	issues := make([]api.IssueWithMeta, 30)
	for i := range issues {
		issues[i] = api.IssueWithMeta{
			Issue: sharedtypes.Issue{
				ID:          int64(i + 1),
				Title:       "Issue",
				Status:      sharedtypes.IssueStatusPlanned,
				TodoForDate: &today,
				PinnedDaily: true,
			},
		}
	}
	state := OpenAdhocIssuePicker(State{Height: 24}, issues, today)
	state.ChoiceCursor = 12
	ctx := UpdateContext{AllIssues: issues, CurrentDate: today}
	state = switchAdhocIssueTab(state, ctx, -1)
	if state.ChoiceCursor != 12 {
		t.Fatal("selection lost switching to All")
	}
	for _, tc := range []struct {
		key  tea.KeyType
		want int
	}{{tea.KeyHome, 0}, {tea.KeyPgDown, 2}, {tea.KeyPgUp, 0}, {tea.KeyEnd, 29}, {tea.KeyUp, 28}, {tea.KeyDown, 29}} {
		state, _, _ = updateAdhocIssuePicker(state, ctx, tea.KeyMsg{Type: tc.key})
		if state.ChoiceCursor != tc.want {
			t.Fatalf("key %v cursor %d want %d", tc.key, state.ChoiceCursor, tc.want)
		}
	}
}
func TestEmptyAdhocSummaryKeepsDialogOpen(t *testing.T) {
	state := OpenAdhocEndSession(State{}, sharedtypes.AdhocSessionResolutionKeep, 0, 0, 0, "")
	next, action, status := updateSessionMessage(
		state,
		UpdateContext{},
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyCtrlS},
	)
	if next.Kind != "end_session" || action != nil || status != "Session summary is required" {
		t.Fatalf("empty summary: %+v %q", action, status)
	}
}

func TestCancelAdhocIssueCreationReturnsToChoices(t *testing.T) {
	state := OpenAdhocIssueCreation(
		OpenAdhocResolution(State{AdhocElapsedSeconds: 185, AdhocDate: "2026-10-08"}, nil),
		nil,
	)
	next, action, status := updateCreateIssueDefault(
		state,
		UpdateContext{},
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyEsc},
	)
	if next.Kind != "adhoc_resolution" || len(next.ChoiceItems) != 3 || action != nil ||
		status != "" ||
		next.AdhocElapsedSeconds != 185 ||
		next.AdhocDate != "2026-10-08" {
		t.Fatalf(
			"cancel should return to attribution with timing intact: %+v, %+v, %q",
			next,
			action,
			status,
		)
	}
	next.ChoiceCursor = 2
	confirmation, _, _ := updateAdhocResolution(
		next,
		UpdateContext{},
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if confirmation.AdhocElapsedSeconds != 185 {
		t.Fatalf("keep lost elapsed time: %d", confirmation.AdhocElapsedSeconds)
	}
}
func TestCancelOrdinaryIssueCreationStillCloses(t *testing.T) {
	state := OpenCreateIssueDefault(State{})
	next, _, _ := updateCreateIssueDefault(
		state,
		UpdateContext{},
		"2026-10-08",
		tea.KeyMsg{Type: tea.KeyEsc},
	)
	if next.Kind != "" {
		t.Fatalf("ordinary cancel kind = %q", next.Kind)
	}
}
