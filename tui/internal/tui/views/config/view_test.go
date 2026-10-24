package config

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	configitems "crona/tui/internal/tui/configitems"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

var longDir = "/Users/someone/Library/Application Support/Crona Dev/a-much-longer-folder-name/reports"

func configState(width, height int, assets *api.ExportAssetStatus) types.ContentState {
	return types.ContentState{
		View: "config", Pane: "config", Width: width, Height: height, ExportAssets: assets,
		Cursors: map[string]int{"config": 0}, Filters: map[string]string{},
	}
}

func assets() *api.ExportAssetStatus {
	return &api.ExportAssetStatus{
		ReportsDir:        longDir,
		DefaultReportsDir: longDir,
		ICSDir:            "/tmp/ics",
		DefaultICSDir:     "/tmp/ics",
	}
}

func TestConfigEveryRowHasHelp(t *testing.T) {
	for _, item := range configitems.Build(assets()) {
		if strings.TrimSpace(item.DetailBody) == "" {
			t.Fatalf("expected help for %q", item.Label)
		}
	}
}

func TestConfigDetailShowsWholePath(t *testing.T) {
	text := ansi.Strip(renderView(types.Theme{}, configState(200, 40, assets())))
	joined := ""
	for _, line := range strings.Split(text, "\n") {
		if _, right, ok := strings.Cut(line, " │ "); ok {
			joined += strings.TrimSpace(right)
		}
	}
	if !strings.Contains(joined, longDir) {
		t.Fatalf("expected the whole path in the detail, got\n%s", text)
	}
}

func TestConfigLayout(t *testing.T) {
	text := ansi.Strip(renderView(types.Theme{}, configState(120, 36, assets())))
	for _, want := range []string{"── Directories", "▶ Reports directory", "─ Reports directory ", "[R] rescan tools"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "╭") {
		t.Fatalf("expected no box")
	}
}

func TestConfigLoading(t *testing.T) {
	if text := ansi.Strip(renderView(types.Theme{}, configState(120, 36, nil))); !strings.Contains(
		text,
		"Loading…",
	) {
		t.Fatalf("expected loading, got\n%s", text)
	}
}
