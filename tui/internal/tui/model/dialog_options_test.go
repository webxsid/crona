package model

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"crona/tui/internal/api"
)

func TestDialogRenderStateCarriesPickerOptions(t *testing.T) {
	m := Model{
		dialog: "checkout_context", width: 120, height: 36,
		cursor: map[Pane]int{}, filters: map[Pane]string{},
		repos:        []api.Repo{{ID: 1, Name: "OSS"}, {ID: 2, Name: "Work"}},
		dialogInputs: []textinput.Model{textinput.New(), textinput.New()},
	}
	if got := strings.Join(m.dialogRenderState().RepoOptionLabels, ","); got != "OSS,Work" {
		t.Fatalf("expected both repos, got %q", got)
	}
	m.dialogInputs[0].SetValue("wo")
	if got := strings.Join(m.dialogRenderState().RepoOptionLabels, ","); got != "Work,Create New Repo: wo" {
		t.Fatalf("expected the filtered repos, got %q", got)
	}
}
