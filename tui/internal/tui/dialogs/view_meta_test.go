package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const issueViewMeta = "Repo OSS   Stream cli   Status in_progress   worked 19m / est. 35m   Due 2026-10-18   ID 10"

func TestRenderViewMetaStacksColumns(t *testing.T) {
	lines := renderViewMeta(Theme{}, issueViewMeta, 70)
	if len(lines) != 2 {
		t.Fatalf("expected 6 fields in 3 columns over 2 rows, got %q", lines)
	}
	if !strings.HasPrefix(lines[0], "repo") || !strings.Contains(lines[0], "stream") || !strings.Contains(lines[0], "status") {
		t.Fatalf("expected repo, stream and status on the first row, got %q", lines[0])
	}
	// Columns line up: "worked" starts where "repo" does, "due" where "stream" does.
	if strings.Index(lines[1], "due") != strings.Index(lines[0], "stream") {
		t.Fatalf("expected aligned columns, got %q", lines)
	}
	for _, line := range lines {
		if ansi.StringWidth(line) > 70 {
			t.Fatalf("line %q is wider than 70", line)
		}
	}
}

func TestRenderViewMetaFallsBackToOneColumn(t *testing.T) {
	if lines := renderViewMeta(Theme{}, issueViewMeta, 30); len(lines) != 6 {
		t.Fatalf("expected one field per row when narrow, got %q", lines)
	}
}
