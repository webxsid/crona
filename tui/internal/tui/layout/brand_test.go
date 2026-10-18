package layout

import (
	"strings"
	"testing"

	versionpkg "crona/shared/version"
	uistate "crona/tui/internal/tui/state"
)

func TestTabLineShowsMarkAndVersion(t *testing.T) {
	line := plainLines(Render(shellStateForTest(uistate.ViewDaily, 160, 36)))[0]
	if !strings.HasPrefix(strings.TrimSpace(line), "[ >_ ]") {
		t.Fatalf("expected the crona mark first, got %q", line)
	}
	if !strings.HasSuffix(strings.TrimSpace(line), "v"+versionpkg.Current()) {
		t.Fatalf("expected the version last, got %q", line)
	}
	if !strings.Contains(line, "crona › tui") {
		t.Fatalf("expected the context to stay, got %q", line)
	}
}

func TestTabLineDropsVersionThenMarkWhenNarrow(t *testing.T) {
	line := plainLines(Render(shellStateForTest(uistate.ViewDaily, 60, 24)))[0]
	if !strings.Contains(line, "[5] System") || strings.Contains(line, "v"+versionpkg.Current()) {
		t.Fatalf("expected tabs kept and version dropped at 60, got %q", line)
	}
}
