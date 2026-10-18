package ops

import (
	"strings"
	"testing"
	"time"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func at(day, hour, minute int) string {
	return time.Date(2026, 10, day, hour, minute, 0, 0, time.Local).UTC().Format(time.RFC3339)
}

func opsState(width int) types.ContentState {
	return types.ContentState{
		View: "ops", Pane: "ops", Width: width, Height: 30,
		Cursors: map[string]int{"ops": 0}, Filters: map[string]string{},
		Ops: []api.Op{
			{Entity: "issue", EntityID: "10", Action: "create", Timestamp: at(16, 16, 17), UserID: "u1", DeviceID: "d1"},
			{Entity: "issue", EntityID: "10", Action: "update", Timestamp: at(17, 9, 0), UserID: "u1", DeviceID: "d1"},
			{Entity: "habit", EntityID: "3", Action: "delete", Timestamp: at(17, 10, 0), UserID: "u1", DeviceID: "d1"},
			{Entity: "habit_completion", EntityID: "2026-10-17", Action: "update", Timestamp: at(17, 11, 5),
				UserID: "u1", DeviceID: "laptop", Payload: map[string]any{"status": "completed", "minutes": 9}},
		},
	}
}

func opsText(width int) string { return ansi.Strip(renderView(types.Theme{}, opsState(width))) }

func TestOpsStripCountsActions(t *testing.T) {
	text := opsText(200)
	for _, want := range []string{"Ops 4", "create 1  update 2  delete 1", "Since 2026-10-16 16:17"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
}

func TestOpsRowsHumaniseAndKeepTarget(t *testing.T) {
	text := opsText(120)
	if !strings.Contains(text, "habit completion") || !strings.Contains(text, "2026-10-17") || strings.Contains(text, "habit_completion") {
		t.Fatalf("expected humanised entity and whole target, got\n%s", text)
	}
}

func TestOpsListMetaCarriesLimitKeys(t *testing.T) {
	if text := opsText(120); !strings.Contains(text, "[+/-] limit") {
		t.Fatalf("expected limit keys in list meta, got\n%s", text)
	}
}

func TestPayloadLines(t *testing.T) {
	lines := PayloadLines(map[string]any{"b": 1, "a": map[string]any{"x": "y"}}, 30)
	got := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.HasPrefix(got, "a\n  x  y\nb  1") {
		t.Fatalf("expected sorted, nested payload lines, got %q", got)
	}
	if got := ansi.Strip(strings.Join(PayloadLines("hello", 30), "")); got != `"hello"` {
		t.Fatalf("expected compact JSON for a string, got %q", got)
	}
	if got := ansi.Strip(strings.Join(PayloadLines([]any{1, 2}, 30), "")); got != "[1,2]" {
		t.Fatalf("expected compact JSON for an array, got %q", got)
	}
	if got := ansi.Strip(strings.Join(PayloadLines(nil, 30), "")); got != "–" {
		t.Fatalf("expected a dash for nil, got %q", got)
	}
	long := PayloadLines(map[string]any{"note": strings.Repeat("word ", 30)}, 30)
	for _, line := range long {
		if ansi.StringWidth(line) > 30 {
			t.Fatalf("expected lines within 30 cells, got %q", line)
		}
	}
}

func TestOpsDetail(t *testing.T) {
	text := opsText(200)
	for _, want := range []string{"habit completion update", "Target", "Time", "Device", "laptop", "User", "Payload", "status  completed"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
}
