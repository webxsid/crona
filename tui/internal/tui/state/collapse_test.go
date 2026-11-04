package state

import "testing"

func TestKeyLabelUsesOptionGlyphOnMac(t *testing.T) {
	if got := keyLabelFor("darwin", "alt+1"); got != "⌥1" {
		t.Fatalf("darwin: got %q, want ⌥1", got)
	}
	if got := keyLabelFor("linux", "alt+1"); got != "alt+1" {
		t.Fatalf("linux: got %q, want alt+1", got)
	}
	if got := FoldKey(CollapseDailyDetail); got != "alt+3" {
		t.Fatalf("FoldKey(detail) = %q, want alt+3", got)
	}
}
