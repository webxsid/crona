package viewchrome

import (
	"strings"
	"testing"
)

func TestLogoSizesAreDistinct(t *testing.T) {
	if got := LogoTiny(); got != "[ CRONA ]" {
		t.Fatalf("expected tiny logo wordmark, got %q", got)
	}
	if got := strings.Count(LogoSmall(), "\n") + 1; got != 2 {
		t.Fatalf("expected small logo to span 2 lines, got %d", got)
	}
	if got := strings.Count(LogoMedium(), "\n") + 1; got != 4 {
		t.Fatalf("expected medium logo to span 4 lines, got %d", got)
	}
	if got := strings.Count(LogoLarge(), "\n") + 1; got != 6 {
		t.Fatalf("expected large logo to span 6 lines, got %d", got)
	}
}
