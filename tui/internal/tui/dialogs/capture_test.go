package dialogs

import (
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestCaptureRowEditorSaysEscGoesBack(t *testing.T) {
	rows := []shareddto.IssueCaptureProposal{{Line: 1, Title: "call bank", RepoName: "Inbox", StreamName: "Inbox", Inbox: true}}
	review := controllerpkg.OpenCaptureReview(controllerpkg.OpenCaptureIssues(controllerpkg.State{}, 0, "", nil), "- call bank", rows)
	editor, _, _ := controllerpkg.Update(review, controllerpkg.UpdateContext{}, "2026-10-27", tea.KeyMsg{Type: tea.KeyEnter})
	out := ansi.Strip(Render(testTheme(), sized(editor, 120, 36)))
	if !strings.Contains(out, "Edit Import Row") || !strings.Contains(out, "esc back") || strings.Contains(out, "esc cancel") {
		t.Fatalf("row editor should say esc goes back:\n%s", out)
	}
	if strings.Contains(out, "create more") {
		t.Fatalf("row editor should not offer create shortcuts:\n%s", out)
	}
}
