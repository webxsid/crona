package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"
	uistate "crona/tui/internal/tui/state"

	"github.com/charmbracelet/x/ansi"
)

// assertListChrome is assertFormChrome for dialogs whose lists may show [k] item keys:
// it checks the frame title, the context, and a `key action` footer ending esc cancel.
func assertListChrome(t *testing.T, c formCase) string {
	t.Helper()
	out := ansi.Strip(Render(testTheme(), sized(c.state, 120, 36)))
	lines := strings.Split(out, "\n")
	if !strings.HasPrefix(lines[0], "╭─ "+c.title) {
		t.Errorf("%s: top border %q lacks title %q", c.name, lines[0], c.title)
	}
	if c.context != "" && !strings.Contains(lines[0], c.context) {
		t.Errorf("%s: top border %q lacks context %q", c.name, lines[0], c.context)
	}
	for _, old := range []string{"[esc]", "[enter]", "[↑/↓]"} {
		if strings.Contains(out, old) {
			t.Errorf("%s: old footer %q remains:\n%s", c.name, old, out)
		}
	}
	last := ""
	for i := len(lines) - 2; i > 0; i-- {
		if inside := strings.TrimSpace(strings.Trim(lines[i], "│")); inside != "" {
			last = inside
			break
		}
	}
	if !strings.HasSuffix(last, "esc cancel") {
		t.Errorf("%s: footer ends %q, want esc cancel", c.name, last)
	}
	return out
}

func TestConfirmAndListDialogs(t *testing.T) {
	status := controllerpkg.OpenIssueStatus(controllerpkg.State{}, "in_progress", "Fix alerts", "OSS", "cli", nil, 0)
	cases := []formCase{
		{name: "confirm_delete", state: controllerpkg.OpenConfirmDelete(controllerpkg.State{}, "issue", "1", "Fix alerts", 1, 2), title: "Delete Issue"},
		{name: "confirm_wipe", state: controllerpkg.OpenConfirmWipeData(controllerpkg.State{}), title: "Wipe All Data"},
		{name: "issue_status", state: status, title: "Set Issue Status", context: "OSS › cli"},
		{name: "view_jump", state: controllerpkg.OpenViewJump(controllerpkg.State{}, []uistate.View{uistate.ViewDaily, uistate.ViewDefault}), title: "Jump To View"},
		{name: "beta_support", state: controllerpkg.OpenBetaSupport(controllerpkg.State{}), title: "Beta Support"},
	}
	for _, c := range cases {
		out := assertListChrome(t, c)
		if c.name == "confirm_delete" && !strings.Contains(out, "enter delete") {
			t.Errorf("confirm_delete footer lacks enter delete:\n%s", out)
		}
		if c.name == "issue_status" && !strings.Contains(out, "▶") {
			t.Errorf("issue_status lacks a cursor on its list:\n%s", out)
		}
	}
}
