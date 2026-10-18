package support

import (
	"strings"
	"testing"

	types "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func supportText(width, cursor int) string {
	return ansi.Strip(renderView(types.Theme{}, types.ContentState{View: "support", Width: width, Height: 30, SupportCursor: cursor}))
}

func TestSupportLayout(t *testing.T) {
	wide := supportText(120, 2)
	sideBySide := false
	for _, line := range strings.Split(wide, "\n") {
		sideBySide = sideBySide || (strings.Contains(line, "─ Actions") && strings.Contains(line, "─ Diagnostics"))
	}
	if !sideBySide || !strings.Contains(wide, "▶ [g] Documentation") || strings.Contains(wide, "╭") {
		t.Fatalf("expected actions beside diagnostics with the cursor on docs, got\n%s", wide)
	}
	narrow := supportText(60, 0)
	if a, d := strings.Index(narrow, "─ Actions"), strings.Index(narrow, "─ Diagnostics"); a < 0 || d < a || strings.Contains(narrow, " │ ") {
		t.Fatalf("expected stacked sections when narrow, got\n%s", narrow)
	}
}

func TestSupportActionsOrder(t *testing.T) {
	keys := []string{}
	for _, action := range Actions() {
		keys = append(keys, action.Key)
	}
	if strings.Join(keys, "") != "orgcb" {
		t.Fatalf("expected actions o r g c b, got %v", keys)
	}
}
