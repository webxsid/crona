package layout

import (
	"strings"
	"testing"

	versionpkg "crona/shared/version"
	uistate "crona/tui/internal/tui/state"
)

func statusRow(state State) string {
	lines := plainLines(Render(state))
	return lines[len(lines)-1]
}

func TestStatusLineEndsWithMarkDevAndVersion(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 160, 36)
	state.IsDevMode = true
	line := strings.TrimSpace(statusRow(state))
	if !strings.HasSuffix(line, "[ >_ ] DEV v"+versionpkg.Current()) {
		t.Fatalf("expected the mark, dev badge and version at the end of the status line, got %q", line)
	}
	if top := plainLines(Render(state))[0]; strings.Contains(top, "[ >_ ]") || strings.Contains(top, versionpkg.Current()) {
		t.Fatalf("expected the mark and version off the tab line, got %q", top)
	}
}

func TestStatusLineDropsMarkBeforeKeysWhenNarrow(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 60, 24)
	state.GlobalActions = []string{"[tab] next pane", "[c] context"}
	line := statusRow(state)
	if strings.Contains(line, versionpkg.Current()) || !strings.Contains(line, "[tab] next pane") {
		t.Fatalf("expected keys kept and the version dropped at 60, got %q", line)
	}
}
