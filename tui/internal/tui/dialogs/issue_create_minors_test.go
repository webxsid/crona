package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

func TestNewIssueMetaNamesRepoAndStreamOnce(t *testing.T) {
	state := controllerpkg.OpenCreateIssueMeta(
		controllerpkg.State{Width: 100, Height: 40},
		3,
		"cli",
		"OSS",
	)
	rendered := ansi.Strip(renderSessionDialogOrIssue(state))
	if n := strings.Count(rendered, "OSS"); n != 1 {
		t.Fatalf("repo shown %d times:\n%s", n, rendered)
	}
}

func renderSessionDialogOrIssue(state controllerpkg.State) string {
	return renderIssueDialog(Theme{}, state)
}
