package updates

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func updatesState(width, height int, expanded bool) types.ContentState {
	return types.ContentState{
		View: "updates", Width: width, Height: height, UpdateDiagnosticsExpanded: expanded,
		UpdateStatus: &api.UpdateStatus{CurrentVersion: "1.10.0", ReleaseTag: "v1.10.0", ReleaseNotes: "- Faster\n- Smaller", Enabled: true},
	}
}

func render(state types.ContentState) string { return ansi.Strip(renderView(types.Theme{}, state)) }

func TestUpdatesColumns(t *testing.T) {
	for _, line := range strings.Split(render(updatesState(160, 40, false)), "\n") {
		if strings.Contains(line, "Status") && strings.Contains(line, "Release Notes") {
			return
		}
	}
	t.Fatalf("expected Status and Release Notes side by side, got\n%s", render(updatesState(160, 40, false)))
}

func TestUpdatesDiagnosticsToggle(t *testing.T) {
	if strings.Contains(render(updatesState(160, 40, false)), "Release Tag") {
		t.Fatalf("expected diagnostics hidden")
	}
	if got := render(updatesState(160, 40, true)); !strings.Contains(got, "─ Diagnostics") || !strings.Contains(got, "Release Tag") {
		t.Fatalf("expected diagnostics shown, got\n%s", got)
	}
}

func TestUpdatesNarrowStacks(t *testing.T) {
	got := render(updatesState(60, 40, false))
	status, notes := strings.Index(got, "─ Status"), strings.Index(got, "─ Release Notes")
	if status < 0 || notes < status || strings.Contains(got, " │ ") {
		t.Fatalf("expected stacked sections, got\n%s", got)
	}
}

func TestUpdatesNoBox(t *testing.T) {
	if got := render(updatesState(120, 36, false)); strings.Contains(got, "╭") || !strings.Contains(got, "[r] check again") {
		t.Fatalf("expected no box and keys in the header, got\n%s", got)
	}
}
