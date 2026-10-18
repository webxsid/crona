package testsuite

import (
	"strings"
	"testing"

	"crona/tui/internal/api"
	"crona/tui/internal/tui/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func assertTinySummary(t *testing.T, rendered string) {
	t.Helper()
	// Collapse the strip's column padding; the date now lives on the sub-tab line.
	flat := strings.Join(strings.Fields(ansi.Strip(rendered)), " ")
	for _, want := range []string{"Tasks 0/1", "Habits 0/1", "─ Tasks", "─ Habits Due"} {
		if !strings.Contains(flat, want) {
			t.Fatalf("expected compact daily to contain %q, got %q", want, rendered)
		}
	}
}

func compactDefaultState(height int) views.ContentState {
	estimate1, estimate2, estimate3 := 60, 35, 25
	today := "2026-03-19"
	return views.ContentState{
		View:   "default",
		Pane:   "issues",
		Width:  92,
		Height: height,
		Cursors: map[string]int{
			"issues": 0,
		},
		Filters: map[string]string{
			"issues": "",
		},
		DefaultIssueSection: "open",
		DefaultIssues: []api.IssueWithMeta{
			{
				Issue: api.Issue{
					ID:              1,
					Title:           "Add keyboard-first command palette",
					Status:          "planned",
					EstimateMinutes: &estimate1,
					TodoForDate:     &today,
				},
				RepoName:   "Work",
				StreamName: "app",
			},
			{
				Issue: api.Issue{
					ID:              2,
					Title:           "Improve install docs for Linux",
					Status:          "planned",
					EstimateMinutes: &estimate2,
					TodoForDate:     &today,
				},
				RepoName:   "OSS",
				StreamName: "cli",
			},
			{
				Issue: api.Issue{
					ID:              3,
					Title:           "Research standing desk options",
					Status:          "abandoned",
					EstimateMinutes: &estimate3,
				},
				RepoName:   "Personal",
				StreamName: "home",
			},
		},
		Context: &api.ActiveContext{},
	}
}

func assertCompactDefault(t *testing.T, rendered string, height int) {
	t.Helper()
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "Open ") {
		t.Fatalf("expected two-row issue strip")
	}
	if !strings.Contains(plain, "─ Issues") {
		t.Fatalf("expected borderless issue list")
	}
	if !strings.Contains(plain, "Closed") {
		t.Fatalf("expected Closed count in issue strip")
	}
	if !strings.Contains(plain, "Add keyboard-first command palette") {
		t.Fatalf("expected open issue rows in compact default view")
	}
	if got := lipgloss.Height(rendered); got > height {
		t.Fatalf("default compact view height %d exceeds allocated height %d", got, height)
	}
}
