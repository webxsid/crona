package layout

import (
	"strings"
	"testing"

	"crona/tui/internal/tui/views"
)

func TestRenderShowsDedicatedUpdateInstallScreen(t *testing.T) {
	rendered := Render(State{
		Width:             150,
		Height:            40,
		UpdateInstallOpen: true,
		ContentState: views.ContentState{
			UpdateInstalling:    true,
			UpdateInstallPhase:  "installing",
			UpdateInstallDetail: "Installing update...",
			UpdateInstallOutput: "Downloaded checksums\nVerified checksum",
		},
	})
	for _, want := range []string{"Updating Crona", "Installing update", "Recent output"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected update install screen to contain %q, got %q", want, rendered)
		}
	}
	for _, unwanted := range []string{"Views", "repo:", "stream:"} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("expected dedicated update screen to hide normal shell chrome %q, got %q", unwanted, rendered)
		}
	}
}
