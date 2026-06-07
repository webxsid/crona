package testsuite

import (
	"testing"

	"crona/tui/internal/api"
	dialogs "crona/tui/internal/tui/dialogs/controller"
	app "crona/tui/internal/tui/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultScopedIssues(t *testing.T) {
	repoID := int64(10)
	streamID := int64(20)
	otherRepoID := int64(11)
	otherStreamID := int64(21)

	model := app.NewDefaultScopeModel([]api.IssueWithMeta{
		{
			Issue:      api.Issue{ID: 1, StreamID: streamID, Title: "A"},
			RepoID:     repoID,
			RepoName:   "Repo",
			StreamName: "main",
		},
		{
			Issue:      api.Issue{ID: 2, StreamID: otherStreamID, Title: "B"},
			RepoID:     otherRepoID,
			RepoName:   "Other",
			StreamName: "dev",
		},
		{
			Issue:      api.Issue{ID: 3, StreamID: 22, Title: "C"},
			RepoID:     repoID,
			RepoName:   "Repo",
			StreamName: "next",
		},
	}, nil)

	if got := app.DefaultScopedIssuesForTest(model); len(got) != 3 {
		t.Fatalf("expected all issues without context, got %d", len(got))
	}

	model = app.NewDefaultScopeModel(model.AllIssuesForTest(), &api.ActiveContext{RepoID: &repoID})
	if got := app.DefaultScopedIssuesForTest(model); len(got) != 2 || got[0].RepoID != repoID ||
		got[1].RepoID != repoID {
		t.Fatalf("expected repo-scoped issues, got %+v", got)
	}

	model = app.NewDefaultScopeModel(
		model.AllIssuesForTest(),
		&api.ActiveContext{RepoID: &repoID, StreamID: &streamID},
	)
	if got := app.DefaultScopedIssuesForTest(model); len(got) != 1 || got[0].StreamID != streamID {
		t.Fatalf("expected stream-scoped issues, got %+v", got)
	}
}

func TestDefaultDialogsPrepopulateFromContext(t *testing.T) {
	repoID := int64(10)
	streamID := int64(20)
	repoName := "Repo"
	streamName := "main"

	model := app.NewDefaultScopeModel(nil, &api.ActiveContext{
		RepoID:     &repoID,
		RepoName:   &repoName,
		StreamID:   &streamID,
		StreamName: &streamName,
	})

	issueDialog := app.OpenCreateIssueDefaultDialogForTest(model)
	if issueDialog.DialogInputValue(0) != repoName {
		t.Fatalf("expected repo prefilled, got %q", issueDialog.DialogInputValue(0))
	}
	if issueDialog.DialogInputValue(1) != streamName {
		t.Fatalf("expected stream prefilled, got %q", issueDialog.DialogInputValue(1))
	}
	if issueDialog.DialogFocusIndex() != 2 {
		t.Fatalf(
			"expected title focus for stream-scoped issue dialog, got %d",
			issueDialog.DialogFocusIndex(),
		)
	}

	checkoutDialog := app.OpenCheckoutContextDialogForTest(model)
	if checkoutDialog.DialogInputValue(0) != repoName {
		t.Fatalf("expected checkout repo prefilled, got %q", checkoutDialog.DialogInputValue(0))
	}
	if checkoutDialog.DialogInputValue(1) != streamName {
		t.Fatalf("expected checkout stream prefilled, got %q", checkoutDialog.DialogInputValue(1))
	}
}

func TestCheckoutDialogSelectionUsesResolvedRepoAndStream(t *testing.T) {
	repos := []api.Repo{{ID: 10, Name: "Work"}}
	streams := []api.Stream{{ID: 20, RepoID: 10, Name: "app"}}
	allIssues := []api.IssueWithMeta{
		{
			Issue:      api.Issue{ID: 1, StreamID: 20, Title: "A"},
			RepoID:     10,
			RepoName:   "Work",
			StreamName: "app",
		},
	}
	repoInput := textinput.New()
	repoInput.SetValue("Wo")
	streamInput := textinput.New()
	streamInput.SetValue("ap")

	repoID, repoName, streamID, streamName := dialogs.CheckoutDialogSelection(
		[]textinput.Model{repoInput, streamInput},
		0,
		0,
		repos,
		allIssues,
		streams,
		nil,
	)
	if repoID != 10 || repoName != "Work" {
		t.Fatalf("expected repo Work, got %d %q", repoID, repoName)
	}
	if streamID == nil || *streamID != 20 || streamName != "app" {
		t.Fatalf("expected stream app, got %v %q", streamID, streamName)
	}
}

func TestCheckoutContextArrowCyclesBlankSelections(t *testing.T) {
	state := dialogs.OpenCheckoutContext(dialogs.State{})
	ctx := dialogs.UpdateContext{
		Repos: []api.Repo{
			{ID: 10, Name: "Work"},
			{ID: 11, Name: "Personal"},
		},
		Streams: []api.Stream{
			{ID: 20, RepoID: 10, Name: "main"},
			{ID: 21, RepoID: 10, Name: "dev"},
		},
	}

	next, action, status := dialogs.Update(
		state,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyRight},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected repo cycling to stay in dialog, got %+v", action)
	}
	if next.RepoIndex != 0 {
		t.Fatalf("expected right from blank repo to select first option, got %d", next.RepoIndex)
	}

	repoLabel, streamLabel := dialogs.CheckoutDialogLabels(
		next.Inputs,
		next.RepoIndex,
		next.StreamIndex,
		ctx.Repos,
		ctx.AllIssues,
		ctx.Streams,
		ctx.Context,
	)
	if repoLabel != "Work" {
		t.Fatalf("expected highlighted repo label, got %q", repoLabel)
	}
	if streamLabel != "Select a stream" {
		t.Fatalf("expected empty stream placeholder, got %q", streamLabel)
	}

	next, action, status = dialogs.Update(
		next,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyTab},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected tab to stay in dialog, got %+v", action)
	}
	next, action, status = dialogs.Update(
		next,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyRight},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected stream cycling to stay in dialog, got %+v", action)
	}
	if next.StreamIndex != 0 {
		t.Fatalf("expected right from blank stream to select first option, got %d", next.StreamIndex)
	}

	repoLabel, streamLabel = dialogs.CheckoutDialogLabels(
		next.Inputs,
		next.RepoIndex,
		next.StreamIndex,
		ctx.Repos,
		ctx.AllIssues,
		ctx.Streams,
		ctx.Context,
	)
	if repoLabel != "Work" {
		t.Fatalf("expected repo label to remain selected, got %q", repoLabel)
	}
	if streamLabel != "main" {
		t.Fatalf("expected highlighted stream label, got %q", streamLabel)
	}

	next, action, status = dialogs.Update(
		next,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close after enter, got %q", next.Kind)
	}
	if action == nil || action.Kind != "checkout_context" || action.RepoID != 10 ||
		action.RepoName != "Work" || action.StreamID != 20 || action.StreamName != "main" {
		t.Fatalf("unexpected checkout action %+v", action)
	}
}

func TestCheckoutContextLeftCyclesBlankRepoSelection(t *testing.T) {
	state := dialogs.OpenCheckoutContext(dialogs.State{})
	ctx := dialogs.UpdateContext{
		Repos: []api.Repo{
			{ID: 10, Name: "Work"},
			{ID: 11, Name: "Personal"},
		},
	}

	next, action, status := dialogs.Update(
		state,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyLeft},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if action != nil {
		t.Fatalf("expected repo cycling to stay in dialog, got %+v", action)
	}
	if next.RepoIndex != 1 {
		t.Fatalf("expected left from blank repo to select last option, got %d", next.RepoIndex)
	}
	repoLabel, streamLabel := dialogs.CheckoutDialogLabels(
		next.Inputs,
		next.RepoIndex,
		next.StreamIndex,
		ctx.Repos,
		ctx.AllIssues,
		ctx.Streams,
		ctx.Context,
	)
	if repoLabel != "Personal" {
		t.Fatalf("expected highlighted repo label, got %q", repoLabel)
	}
	if streamLabel != "Select a stream" {
		t.Fatalf("expected empty stream placeholder, got %q", streamLabel)
	}

	next, action, status = dialogs.Update(
		next,
		ctx,
		"2026-04-10",
		tea.KeyMsg{Type: tea.KeyEnter},
	)
	if status != "" {
		t.Fatalf("unexpected status %q", status)
	}
	if next.Kind != "" {
		t.Fatalf("expected dialog to close after enter, got %q", next.Kind)
	}
	if action == nil || action.Kind != "checkout_context" || action.RepoID != 11 ||
		action.RepoName != "Personal" || action.StreamID != 0 || action.StreamName != "" {
		t.Fatalf("unexpected checkout action %+v", action)
	}
}
