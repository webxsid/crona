package localipc

import (
	"strings"
	"testing"
)

func TestDefaultEndpointWindowsUsesRuntimeScopedPipeName(t *testing.T) {
	base := `C:\Users\alice\AppData\Local\Crona`
	got := windowsPipeEndpoint(base, "prod")
	if got == `\\.\pipe\crona-daemon` {
		t.Fatalf("expected runtime-scoped pipe name, got %q", got)
	}
	if !strings.HasPrefix(got, `\\.\pipe\crona`) {
		t.Fatalf("expected windows pipe prefix, got %q", got)
	}
}

func TestWindowsPipeEndpointDiffersAcrossRuntimeBases(t *testing.T) {
	a := windowsPipeEndpoint(`C:\Users\alice\AppData\Local\Crona`, "prod")
	b := windowsPipeEndpoint(`C:\Users\bob\AppData\Local\Crona`, "prod")
	if a == b {
		t.Fatalf("expected distinct endpoints for distinct runtime bases, got %q", a)
	}
}

func TestWindowsPipeEndpointKeepsDevSuffix(t *testing.T) {
	got := windowsPipeEndpoint(`C:\Users\alice\AppData\Local\Crona Dev`, "dev")
	if !strings.HasPrefix(got, `\\.\pipe\crona-daemon-dev`) {
		t.Fatalf("expected dev pipe prefix, got %q", got)
	}
}
