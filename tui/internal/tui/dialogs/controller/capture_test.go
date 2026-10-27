package controller

import (
	"slices"
	"testing"

	shareddto "crona/shared/dto"
	"crona/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
)

func captureKey(name string) tea.KeyMsg {
	switch name {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}

func capturePress(t *testing.T, state State, ctx UpdateContext, names ...string) (State, *Action, string) {
	t.Helper()
	var action *Action
	var status string
	for _, name := range names {
		state, action, status = Update(state, ctx, "2026-10-27", captureKey(name))
	}
	return state, action, status
}

func reviewRows() []shareddto.IssueCaptureProposal {
	return []shareddto.IssueCaptureProposal{
		{Line: 1, Title: "fix login", RepoID: 1, RepoName: "crona", StreamID: 10, StreamName: "api"},
		{Line: 2, Title: "broken", Flags: []string{"unknown #nope"}},
		{Line: 3, Title: "call bank", RepoName: "Inbox", StreamName: "Inbox", Inbox: true},
	}
}

func TestCaptureIssuesParsesTextWithDefaultStream(t *testing.T) {
	state := OpenCaptureIssues(State{}, 10, "crona/api", nil)
	state = typeInForm(state, "- fix login")
	state, action, _ := capturePress(t, state, UpdateContext{}, "ctrl+s")
	if action == nil || action.Kind != "capture_parse" || action.Title != "- fix login" || action.StreamID != 10 {
		t.Fatalf("unexpected action %+v", action)
	}
	if state.Kind != "capture_issues" {
		t.Fatalf("dialog should stay open while parsing, got %q", state.Kind)
	}
}

func TestCaptureReviewSelectionAndImport(t *testing.T) {
	state := OpenCaptureReview(OpenCaptureIssues(State{}, 0, "", nil), "- text", reviewRows())
	if !slices.Equal(state.CaptureSelected, []bool{true, false, true}) {
		t.Fatalf("flagged rows should start unselected: %v", state.CaptureSelected)
	}
	if _, _, status := capturePress(t, state, UpdateContext{}, "down", "space"); status == "" {
		t.Fatal("toggling a flagged row should explain why it cannot be selected")
	}
	state, _, _ = capturePress(t, state, UpdateContext{}, "space")
	if CaptureSelectedCount(state) != 1 {
		t.Fatalf("space should clear the first row: %v", state.CaptureSelected)
	}
	state, _, _ = capturePress(t, state, UpdateContext{}, "a")
	if !slices.Equal(state.CaptureSelected, []bool{true, false, true}) {
		t.Fatalf("a should select every unflagged row: %v", state.CaptureSelected)
	}
	state, action, _ := capturePress(t, state, UpdateContext{}, "ctrl+s")
	if action == nil || action.Kind != "capture_import" || len(action.CaptureRows) != 2 || action.CaptureLeft != 1 {
		t.Fatalf("unexpected import action %+v", action)
	}
	if !state.Processing {
		t.Fatal("review should show it is importing")
	}
}

func TestCaptureReviewEscKeepsTextAndRows(t *testing.T) {
	state := OpenCaptureReview(OpenCaptureIssues(State{}, 0, "", nil), "- text", reviewRows())
	state, _, _ = capturePress(t, state, UpdateContext{}, "space", "esc")
	if state.Kind != "capture_issues" || state.Description.Value() != "- text" {
		t.Fatalf("esc should return to the text: %q %q", state.Kind, state.Description.Value())
	}
	state, action, _ := capturePress(t, state, UpdateContext{}, "ctrl+s")
	if action != nil || state.Kind != "capture_review" || state.CaptureSelected[0] {
		t.Fatalf("unchanged text should reopen the review as it was: %+v %v", action, state.CaptureSelected)
	}
}

func TestCaptureRowEditorWritesTheRowBack(t *testing.T) {
	ctx := UpdateContext{
		Repos:   []api.Repo{{ID: 1, Name: "crona"}},
		Streams: []api.Stream{{ID: 10, RepoID: 1, Name: "api"}},
	}
	state := OpenCaptureReview(OpenCaptureIssues(State{}, 0, "", nil), "- text", reviewRows())
	state, _, _ = capturePress(t, state, ctx, "down", "down", "enter")
	if state.Kind != "create_issue_default" || !state.CaptureEditing || state.Inputs[2].Value() != "call bank" {
		t.Fatalf("enter should open the row in the New Issue form: %q %v", state.Kind, state.CaptureEditing)
	}
	state = typeInForm(state, " today")
	state, action, status := capturePress(t, state, ctx, "ctrl+s")
	if action != nil || status != "" || state.Kind != "capture_review" {
		t.Fatalf("saving should return to the review: %+v %q %q", action, status, state.Kind)
	}
	if row := state.CaptureRows[2]; row.Title != "call bank today" || !row.Inbox {
		t.Fatalf("row not written back: %+v", row)
	}
	if state.CaptureCursor != 2 {
		t.Fatalf("cursor should stay on the edited row, got %d", state.CaptureCursor)
	}

	// esc in the form leaves the row as it was.
	state, _, _ = capturePress(t, state, ctx, "enter")
	state = typeInForm(state, "!!")
	state, _, _ = capturePress(t, state, ctx, "esc")
	if state.Kind != "capture_review" || state.CaptureRows[2].Title != "call bank today" {
		t.Fatalf("esc should discard form edits: %q %+v", state.Kind, state.CaptureRows[2])
	}
}

func TestCaptureRowEditorStartsOnTheRowsTarget(t *testing.T) {
	ctx := UpdateContext{
		Repos:   []api.Repo{{ID: 1, Name: "crona-old"}, {ID: 2, Name: "crona"}},
		Streams: []api.Stream{{ID: 10, RepoID: 2, Name: "web"}, {ID: 11, RepoID: 2, Name: "api"}},
	}
	rows := []shareddto.IssueCaptureProposal{
		{Line: 1, Title: "fix login", RepoID: 2, RepoName: "crona", StreamID: 11, StreamName: "api"},
		{Line: 2, Title: "broken", Flags: []string{"unknown #nope"}},
		{Line: 3, Title: "half", RepoID: 2, RepoName: "crona", Flags: []string{"stream?"}},
	}
	review := OpenCaptureReview(OpenCaptureIssues(State{}, 0, "", nil), "- text", rows)

	state, _, _ := capturePress(t, review, ctx, "enter")
	repo, _ := selectedOption(DefaultRepoOptions(state.Inputs, ctx.Repos), state.RepoIndex)
	stream, _ := selectedOption(DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, nil, ctx.Streams, nil), state.StreamIndex)
	if repo.ID != "2" || stream.ID != "11" {
		t.Fatalf("resolved row should open on crona/api, got repo %q stream %q", repo.ID, stream.ID)
	}

	state, _, _ = capturePress(t, review, ctx, "down", "enter")
	if state.RepoIndex != -1 || state.StreamIndex != -1 {
		t.Fatalf("unknown target should start unchosen, got %d/%d", state.RepoIndex, state.StreamIndex)
	}
	if state, _, status := capturePress(t, state, ctx, "ctrl+s"); status == "" || state.Kind != "create_issue_default" {
		t.Fatalf("saving with no repo chosen should be refused, got %q in %q", status, state.Kind)
	}

	state, _, _ = capturePress(t, review, ctx, "down", "down", "enter")
	if state.RepoIndex < 0 || state.StreamIndex != -1 {
		t.Fatalf("stream? row should keep its repo and leave the stream unchosen, got %d/%d", state.RepoIndex, state.StreamIndex)
	}
}
