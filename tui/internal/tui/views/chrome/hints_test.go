package viewchrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFitHintsKeepsWholeHints(t *testing.T) {
	hints := []string{"[a] new", "[s] status", "[d] due"}
	if got := ansi.Strip(FitHints(hints, 100)); got != "[a] new  [s] status  [d] due" {
		t.Fatalf("got %q", got)
	}
	got := ansi.Strip(FitHints(hints, 20))
	if got != "[a] new  [s] status" || strings.Contains(got, "…") {
		t.Fatalf("expected only whole hints within 20, got %q", got)
	}
}
