package viewchrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func groupRows() []GroupRow {
	return []GroupRow{{Group: "Dates", Line: "Date Format"}, {Group: "Dates", Line: "Week Start"}, {Group: "Danger", Line: "Reset", Danger: true}}
}

func TestGroupListDividesGroups(t *testing.T) {
	lines := GroupList(Theme{}, groupRows(), 1, 30, 10)
	text := ansi.Strip(strings.Join(lines, "\n"))
	if strings.Count(text, "── ") != 2 || !strings.Contains(text, "── Dates ") || !strings.Contains(text, "▶ Week Start") {
		t.Fatalf("expected two dividers and the cursor row, got\n%s", text)
	}
}

func TestGroupListKeepsDividerInView(t *testing.T) {
	lines := GroupList(Theme{}, groupRows(), 2, 30, 2)
	if got := ansi.Strip(lines[0]); !strings.HasPrefix(got, "── Danger") {
		t.Fatalf("expected the group's divider first, got %q", got)
	}
}

func TestHelpLinesMarksCurrent(t *testing.T) {
	text := ansi.Strip(strings.Join(HelpLines(Theme{}, RowHelp{Description: "First day.", Options: []string{"Monday", "Sunday"}, Keys: []string{"[←/→] change"}}, "sunday", 30), "\n"))
	for _, want := range []string{"First day.", "○ Monday", "● Sunday", "[←/→] change"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in\n%s", want, text)
		}
	}
}

func TestHelpLinesWrapsValueWithoutOptions(t *testing.T) {
	path := "/Users/someone/Library/Application Support/Crona Dev/reports/x"
	lines := HelpLines(Theme{}, RowHelp{Description: "Where reports go."}, path, 20)
	joined := ""
	for _, line := range lines {
		plain := ansi.Strip(line)
		if ansi.StringWidth(plain) > 20 {
			t.Fatalf("expected lines within 20 cells, got %q", plain)
		}
		if strings.HasPrefix(plain, "/") || (joined != "" && !strings.HasPrefix(plain, "[")) {
			if plain != "" && !strings.Contains(plain, "reports go") {
				joined += plain
			}
		}
	}
	if joined != path {
		t.Fatalf("expected the whole path, got %q", joined)
	}
}

func TestLabelColumn(t *testing.T) {
	if got := LabelColumn([]string{"ab", strings.Repeat("x", 40)}, 24); got != 24 {
		t.Fatalf("expected the cap, got %d", got)
	}
}

func TestPackHintsKeepsHintsWhole(t *testing.T) {
	got := PackHints([]string{"[r] check again", "[o] open migration guide", "[d] diagnostics"}, 30)
	if len(got) != 3 || got[1] != "[o] open migration guide" {
		t.Fatalf("expected whole hints per line, got %q", got)
	}
	if got := PackHints([]string{"[a] one", "[b] two"}, 30); len(got) != 1 || got[0] != "[a] one  [b] two" {
		t.Fatalf("expected hints packed onto one line, got %q", got)
	}
}

func TestLabelCellFitsColumn(t *testing.T) {
	if got := LabelCell("Add Plan-the-Day Reminder", 24); ansi.StringWidth(got) != 24 || !strings.HasSuffix(got, "…") {
		t.Fatalf("expected a 24-cell truncated label, got %q", got)
	}
	if got := LabelCell("Sound", 8); got != "Sound   " {
		t.Fatalf("expected a padded label, got %q", got)
	}
}
