package issuecore

import (
	"testing"

	"crona/tui/internal/api"

	"github.com/charmbracelet/x/ansi"
)

func TestTagChipsTruncateWithCount(t *testing.T) {
	theme := Theme{}
	tags := []string{"quick", "release", "urgent"}
	if got := ansi.Strip(TagChips(theme, tags, nil, 40)); got != "+quick +release +urgent" {
		t.Fatalf("expected all chips, got %q", got)
	}
	if got := ansi.Strip(TagChips(theme, tags, nil, 16)); got != "+quick +2" {
		t.Fatalf("expected a count for the rest, got %q", got)
	}
	if got := ansi.Strip(TagChips(theme, tags, []api.Tag{{Name: "quick"}}, 2)); got != "+3" {
		t.Fatalf("expected only the count, got %q", got)
	}
}
