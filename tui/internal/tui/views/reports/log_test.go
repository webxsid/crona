package reports

import (
	"strings"
	"testing"
	"time"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func modified(day, hour int) string {
	return time.Date(2026, 10, day, hour, 0, 0, 0, time.Local).UTC().Format(time.RFC3339)
}

func reportsState(
	width int,
	reports []api.ExportReportFile,
	assets *api.ExportAssetStatus,
) types.ContentState {
	return types.ContentState{
		View: "reports", Pane: "export_reports", Width: width, Height: 30,
		Cursors: map[string]int{"export_reports": 0}, Filters: map[string]string{},
		ExportReports: reports, ExportAssets: assets,
	}
}

func sampleReports() []api.ExportReportFile {
	return []api.ExportReportFile{
		{
			Name:       "daily.md",
			Path:       "/r/reports/daily-2026-10-17.md",
			Kind:       "daily",
			DateLabel:  "Sat 17 Oct",
			Format:     "markdown",
			SizeBytes:  1234,
			ModifiedAt: modified(17, 18),
		},
		{
			Name:       "weekly.pdf",
			Path:       "/r/reports/weekly.pdf",
			Kind:       "weekly",
			ScopeLabel: "Crona",
			DateLabel:  "W42",
			Format:     "pdf",
			SizeBytes:  900,
			ModifiedAt: modified(17, 9),
		},
		{
			Name:       "daily.md",
			Path:       "/r/reports/daily-2026-10-16.md",
			Kind:       "daily",
			DateLabel:  "Fri 16 Oct",
			Format:     "markdown",
			SizeBytes:  800,
			ModifiedAt: modified(16, 20),
		},
	}
}

func text(state types.ContentState) string { return ansi.Strip(renderView(types.Theme{}, state)) }

var dir = &api.ExportAssetStatus{
	ReportsDir: "/Users/someone/Library/Application Support/Crona Dev/reports",
}

func TestReportsGroupsByDay(t *testing.T) {
	got := text(reportsState(120, sampleReports(), dir))
	for _, want := range []string{"── Sat 17 Oct · 2", "── Fri 16 Oct · 1", "▶ daily", "Crona  W42"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "╭") {
		t.Fatalf("expected no box")
	}
}

func TestReportsDetail(t *testing.T) {
	got := text(reportsState(160, sampleReports(), dir))
	for _, want := range []string{"daily-2026-10-17.md", "[o] open", "Kind", "Period", "Format", "1.2 KB", "Created", "/r/reports/daily-2026-10-17.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in\n%s", want, got)
		}
	}
}

func TestReportsStripKeepsDirTail(t *testing.T) {
	lines := strings.Split(text(reportsState(40, sampleReports(), dir)), "\n")
	if !strings.HasSuffix(strings.TrimRight(lines[1], " "), "/reports") ||
		!strings.Contains(lines[1], "…") {
		t.Fatalf("expected the dir tail with a leading ellipsis, got %q", lines[1])
	}
}

func TestReportsEmpty(t *testing.T) {
	if got := text(reportsState(120, nil, dir)); !strings.Contains(got, "No exported reports yet") {
		t.Fatalf("expected empty state, got\n%s", got)
	}
}

func TestReportsLoading(t *testing.T) {
	if got := text(reportsState(120, nil, nil)); !strings.Contains(got, "Loading…") {
		t.Fatalf("expected loading, got\n%s", got)
	}
}
