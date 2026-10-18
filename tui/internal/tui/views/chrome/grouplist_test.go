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
