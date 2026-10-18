package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCheckoutChrome(t *testing.T) {
	state := issueState("checkout_context", 0, 120, 36)
	state.RepoSelectorLabel, state.StreamSelectorLabel = "OSS", "cli"
	text := ansi.Strip(renderRepoStreamDialog(testTheme(), state))
	if !strings.Contains(text, "╭─ Checkout Context") || strings.Contains(text, "[ > ") || !strings.Contains(text, "▶ OSS") || !strings.Contains(text, "Work") {
		t.Fatalf("expected the chrome with repo options, got\n%s", text)
	}
	streamLine := false
	for _, line := range strings.Split(text, "\n") {
		streamLine = streamLine || (strings.Contains(line, "Stream") && strings.Contains(line, "cli ▾"))
	}
	if !streamLine {
		t.Fatalf("expected a collapsed stream picker, got\n%s", text)
	}
	state.Width = 60
	narrow := ansi.Strip(renderRepoStreamDialog(testTheme(), state))
	for _, hint := range []string{"enter set context", "c clear", "↑/↓/tab move", "esc cancel"} {
		if !strings.Contains(narrow, hint) {
			t.Fatalf("expected %q whole at width 60, got\n%s", hint, narrow)
		}
	}
}
