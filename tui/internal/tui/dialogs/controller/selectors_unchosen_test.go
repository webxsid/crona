package controller

import (
	"testing"

	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/textinput"
)

// An unresolved import row opens New Issue with nothing chosen; labels and names must
// cope with the negative index instead of indexing with it.
func TestDefaultIssueDialogHandlesUnchosenIndexes(t *testing.T) {
	inputs := []textinput.Model{textinput.New(), textinput.New()}
	repos := []api.Repo{{ID: 1, Name: "crona"}}
	streams := []api.Stream{{ID: 10, RepoID: 1, Name: "api"}}
	if repo, _ := DefaultIssueDialogLabels(inputs, -1, -1, repos, nil, streams, nil); repo == "" {
		t.Fatal("expected a prompt for an unchosen repo")
	}
	if _, stream := DefaultIssueDialogLabels(inputs, 0, -1, repos, nil, streams, nil); stream == "" {
		t.Fatal("expected a prompt for an unchosen stream")
	}
	if repo, stream := DefaultIssueDialogNames(inputs, 0, -1, repos, nil, streams, nil); repo != "" || stream != "" {
		t.Fatalf("an unchosen stream should give no names, got %q/%q", repo, stream)
	}
}
