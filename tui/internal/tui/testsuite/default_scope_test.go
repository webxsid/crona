package testsuite

import (
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/app"
	dialogs "crona/tui/internal/tui/dialogs"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestDefaultScopedIssues(t *testing.T) {
	repoID := int64(10)
	streamID := int64(20)
	otherRepoID := int64(11)
	otherStreamID := int64(21)

	model := app.NewDefaultScopeModel([]api.IssueWithMeta{
		{Issue: api.Issue{ID: 1, StreamID: streamID, Title: "A"}, RepoID: repoID, RepoName: "Repo", StreamName: "main"},
		{Issue: api.Issue{ID: 2, StreamID: otherStreamID, Title: "B"}, RepoID: otherRepoID, RepoName: "Other", StreamName: "dev"},
		{Issue: api.Issue{ID: 3, StreamID: 22, Title: "C"}, RepoID: repoID, RepoName: "Repo", StreamName: "next"},
	}, nil)

	if got := app.DefaultScopedIssuesForTest(model); len(got) != 3 {
		t.Fatalf("expected all issues without context, got %d", len(got))
	}

	model = app.NewDefaultScopeModel(model.AllIssuesForTest(), &api.ActiveContext{RepoID: &repoID})
	if got := app.DefaultScopedIssuesForTest(model); len(got) != 2 || got[0].RepoID != repoID || got[1].RepoID != repoID {
		t.Fatalf("expected repo-scoped issues, got %+v", got)
	}

	model = app.NewDefaultScopeModel(model.AllIssuesForTest(), &api.ActiveContext{RepoID: &repoID, StreamID: &streamID})
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
		t.Fatalf("expected title focus for stream-scoped issue dialog, got %d", issueDialog.DialogFocusIndex())
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
		{Issue: api.Issue{ID: 1, StreamID: 20, Title: "A"}, RepoID: 10, RepoName: "Work", StreamName: "app"},
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
