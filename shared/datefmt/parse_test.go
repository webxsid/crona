package datefmt

import (
	"testing"
	"time"
)

func TestParseDateReadsBackEveryParseablePattern(t *testing.T) {
	day := time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC)
	for _, pattern := range []string{"YYYY-MM-DD", "MM/DD/YYYY", "DD/MM/YYYY", "D MMM YYYY", "ddd, D MMMM YYYY", "L", "LL", "[on] D.M.YY"} {
		rendered := FormatWithPattern(day, pattern)
		got, err := ParseDate(rendered, pattern, time.UTC)
		if err != nil || !got.Equal(day) {
			t.Fatalf("%s: %q parsed to %v, %v", pattern, rendered, got, err)
		}
		if got, err := ParseDate("2026-11-04", pattern, time.UTC); err != nil || !got.Equal(day) {
			t.Fatalf("%s: ISO should always be accepted, got %v, %v", pattern, got, err)
		}
	}
	if _, err := ParseDate("4th Nov 2026", "Do MMM YYYY", time.UTC); err == nil {
		t.Fatal("ordinals cannot be read back, so only ISO should parse")
	}
}
