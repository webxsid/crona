//go:build e2e

package e2e

import (
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

func TestSortSettingsAffectKernelLists(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repoB := createRepo(t, kernel, "Beta")
	repoA := createRepo(t, kernel, "Alpha")

	streamZ := createStream(t, kernel, repoA.ID, "zeta")
	streamA := createStream(t, kernel, repoA.ID, "alpha")

	issueLater := createIssue(t, kernel, streamA.ID, "Later", nil)
	issueSoon := createIssue(t, kernel, streamA.ID, "Soon", nil)
	issueNone := createIssue(t, kernel, streamA.ID, "None", nil)
	target15 := 15
	target45 := 45
	var habitLong, habitShort, habitNone sharedtypes.Habit
	kernel.call(t, protocol.MethodHabitCreate, shareddto.CreateHabitRequest{
		StreamID:      streamA.ID,
		Name:          "Deep Work",
		ScheduleType:  "daily",
		TargetMinutes: &target45,
	}, &habitLong)
	kernel.call(t, protocol.MethodHabitCreate, shareddto.CreateHabitRequest{
		StreamID:      streamA.ID,
		Name:          "Inbox Zero",
		ScheduleType:  "daily",
		TargetMinutes: &target15,
	}, &habitShort)
	kernel.call(t, protocol.MethodHabitCreate, shareddto.CreateHabitRequest{
		StreamID:     streamA.ID,
		Name:         "Journal",
		ScheduleType: "daily",
	}, &habitNone)

	setIssueTodoDate(t, kernel, issueLater.ID, "2026-03-20")
	setIssueTodoDate(t, kernel, issueSoon.ID, "2026-03-18")

	var ok shareddto.OKResponse
	kernel.call(t, protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   sharedtypes.CoreSettingsKeyRepoSort,
		Value: sharedtypes.RepoSortAlphabeticalAsc,
	}, &ok)
	if !ok.OK {
		t.Fatalf("expected repo sort patch ok")
	}
	kernel.call(t, protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   sharedtypes.CoreSettingsKeyStreamSort,
		Value: sharedtypes.StreamSortAlphabeticalDesc,
	}, &ok)
	if !ok.OK {
		t.Fatalf("expected stream sort patch ok")
	}
	kernel.call(t, protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   sharedtypes.CoreSettingsKeyIssueSort,
		Value: sharedtypes.IssueSortDueDateAsc,
	}, &ok)
	if !ok.OK {
		t.Fatalf("expected issue sort patch ok")
	}
	kernel.call(t, protocol.MethodSettingsPatch, shareddto.PatchCoreSettingRequest{
		Key:   sharedtypes.CoreSettingsKeyHabitSort,
		Value: sharedtypes.HabitSortTargetMinutesDesc,
	}, &ok)
	if !ok.OK {
		t.Fatalf("expected habit sort patch ok")
	}

	var repoSort sharedtypes.RepoSort
	kernel.call(t, protocol.MethodSettingsGet, shareddto.GetCoreSettingRequest{
		Key: sharedtypes.CoreSettingsKeyRepoSort,
	}, &repoSort)
	if repoSort != sharedtypes.RepoSortAlphabeticalAsc {
		t.Fatalf("expected repoSort=%q, got %q", sharedtypes.RepoSortAlphabeticalAsc, repoSort)
	}
	var habitSort sharedtypes.HabitSort
	kernel.call(t, protocol.MethodSettingsGet, shareddto.GetCoreSettingRequest{
		Key: sharedtypes.CoreSettingsKeyHabitSort,
	}, &habitSort)
	if habitSort != sharedtypes.HabitSortTargetMinutesDesc {
		t.Fatalf("expected habitSort=%q, got %q", sharedtypes.HabitSortTargetMinutesDesc, habitSort)
	}

	var repos []sharedtypes.Repo
	kernel.call(t, protocol.MethodRepoList, nil, &repos)
	if len(repos) != 2 || repos[0].ID != repoA.ID || repos[1].ID != repoB.ID {
		t.Fatalf("unexpected repo order: %+v", repos)
	}

	var streams []sharedtypes.Stream
	kernel.call(t, protocol.MethodStreamList, shareddto.ListStreamsQuery{RepoID: repoA.ID}, &streams)
	if len(streams) != 2 || streams[0].ID != streamZ.ID || streams[1].ID != streamA.ID {
		t.Fatalf("unexpected stream order: %+v", streams)
	}

	var issues []sharedtypes.Issue
	kernel.call(t, protocol.MethodIssueList, shareddto.ListIssuesQuery{StreamID: streamA.ID}, &issues)
	if len(issues) != 3 || issues[0].ID != issueSoon.ID || issues[1].ID != issueLater.ID || issues[2].ID != issueNone.ID {
		t.Fatalf("unexpected issue order: %+v", issues)
	}

	var allIssues []sharedtypes.IssueWithMeta
	kernel.call(t, protocol.MethodIssueListAll, nil, &allIssues)
	if len(allIssues) != 3 || allIssues[0].ID != issueSoon.ID || allIssues[1].ID != issueLater.ID || allIssues[2].ID != issueNone.ID {
		t.Fatalf("unexpected all issue order: %+v", allIssues)
	}

	var habits []sharedtypes.Habit
	kernel.call(t, protocol.MethodHabitList, shareddto.ListHabitsQuery{StreamID: streamA.ID}, &habits)
	if len(habits) != 3 || habits[0].ID != habitLong.ID || habits[1].ID != habitShort.ID || habits[2].ID != habitNone.ID {
		t.Fatalf("unexpected habit order: %+v", habits)
	}

	var dueHabits []sharedtypes.HabitDailyItem
	kernel.call(t, protocol.MethodHabitListDue, shareddto.ListHabitsDueQuery{Date: "2026-03-19"}, &dueHabits)
	if len(dueHabits) != 3 || dueHabits[0].ID != habitLong.ID || dueHabits[1].ID != habitShort.ID || dueHabits[2].ID != habitNone.ID {
		t.Fatalf("unexpected due habit order: %+v", dueHabits)
	}
}

func setIssueTodoDate(t *testing.T, k *testKernel, issueID int64, date string) {
	t.Helper()
	var issue sharedtypes.Issue
	k.call(t, protocol.MethodIssueSetTodo, map[string]any{
		"id":   issueID,
		"date": date,
	}, &issue)
	if issue.ID != issueID || issue.TodoForDate == nil || *issue.TodoForDate != date {
		t.Fatalf("expected issue %d due date %q, got %+v", issueID, date, issue)
	}
}
