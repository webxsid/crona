package terminaltitle

import (
	"strings"
	"testing"
)

func TestSanitizeStripsControlCharacters(t *testing.T) {
	got := Sanitize(" Crona\x1b]0;bad\a\nTitle\t ")
	if strings.ContainsAny(got, "\x1b\a\n\t") {
		t.Fatalf("expected control characters stripped, got %q", got)
	}
	if got != "Crona]0;badTitle" {
		t.Fatalf("unexpected sanitized title %q", got)
	}
}

func TestSequenceWrapsSanitizedTitle(t *testing.T) {
	got := Sequence("Crona")
	if got != "\x1b]0;Crona\a" {
		t.Fatalf("unexpected title sequence %q", got)
	}
}

func TestSanitizeTruncatesLongTitle(t *testing.T) {
	got := Sanitize(strings.Repeat("x", 120))
	if len([]rune(got)) != maxTitleRunes {
		t.Fatalf("expected %d runes, got %d", maxTitleRunes, len([]rune(got)))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}
