package layout

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	viewchrome "crona/tui/internal/tui/views/chrome"

	"github.com/charmbracelet/x/ansi"
)

func TestStatusLineKeepsKeysWhileATimerRuns(t *testing.T) {
	timer := &api.TimerState{State: "running", SegmentType: new(sharedtypes.SessionSegmentWork)}
	for _, view := range []string{"daily", "session_active"} {
		line := ansi.Strip(renderStatusLine(State{
			Width: 200, TimerActive: true, PaneActions: []string{"[a] new task"},
			HeaderState: viewchrome.HeaderState{Width: 200, View: view, Timer: timer},
		}))
		if !strings.Contains(line, "[a] new task") {
			t.Fatalf("%s: expected pane keys beside the session chip, got %q", view, line)
		}
		if hasChip := strings.Contains(line, "FOCUS"); hasChip != (view == "daily") {
			t.Fatalf("%s: chip shown=%v, got %q", view, hasChip, line)
		}
	}
}
