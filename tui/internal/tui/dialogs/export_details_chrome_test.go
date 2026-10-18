package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/x/ansi"
)

func listState(kind string) controllerpkg.State {
	return controllerpkg.State{
		Kind: kind, CheckInDate: "2026-10-18", ChoiceItems: []string{"Daily", "Weekly"},
		ChoiceDetails: []string{"One day", "Seven days"}, ChoiceCursor: 0,
	}
}

func inputState(kind string, count int) controllerpkg.State {
	state := controllerpkg.State{Kind: kind}
	for i := 0; i < count; i++ {
		state.Inputs = append(state.Inputs, textinput.New())
	}
	return controllerpkg.SyncDialogFocus(state)
}

func TestExportDialogs(t *testing.T) {
	cases := []struct {
		c    formCase
		want []string
	}{
		{formCase{name: "export_report", state: listState("export_report"), title: "Export Report"}, []string{"▶ Daily", "enter choose"}},
		{formCase{name: "export_report_category", state: listState("export_report_category"), title: "Export Report"}, []string{"▶ Daily", "enter choose category"}},
		{formCase{name: "export_preset", state: listState("export_preset"), title: "Choose Report Style"}, []string{"Preview", "enter use style"}},
		{formCase{name: "export_calendar_repo", state: listState("export_calendar_repo"), title: "Calendar Export Repo"}, []string{"▶ Daily", "enter export"}},
		{formCase{name: "export_summary_range", state: inputState("export_summary_range", 2), title: "Custom Summary Range"}, []string{"Start", "End", "enter continue"}},
		{formCase{name: "edit_export_reports_dir", state: inputState("edit_export_reports_dir", 1), title: "Export Reports Directory"}, []string{"Path", "ctrl+s save"}},
		{formCase{name: "edit_export_ics_dir", state: inputState("edit_export_ics_dir", 1), title: "ICS Export Directory"}, []string{"Path", "ctrl+s save"}},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
	}
}

func TestDetailDialogs(t *testing.T) {
	entity := controllerpkg.State{Kind: "view_entity", ViewTitle: "Issue", ViewName: "Fix alerts", ViewMeta: "OSS / cli", ViewBody: "Body text"}
	momentum := controllerpkg.State{Kind: "view_momentum_detail", ViewTitle: "Momentum Detail", ViewName: "Recovery Mix", ViewBody: "Summary", ChoiceItems: []string{"Run", "Read"}}
	bundle := controllerpkg.State{Kind: "support_bundle_result", ViewTitle: "Support Bundle Ready", ViewName: "bundle.zip", ViewBody: "Saved"}
	for _, c := range []formCase{
		{name: "view_entity", state: entity, title: "Issue"},
		{name: "view_momentum_detail", state: momentum, title: "Momentum Detail"},
		{name: "support_bundle_result", state: bundle, title: "Support Bundle Ready"},
	} {
		out := ansi.Strip(Render(testTheme(), sized(c.state, 120, 36)))
		lines := strings.Split(out, "\n")
		if !strings.HasPrefix(lines[0], "╭─ "+c.title) || strings.Contains(out, "[enter/esc]") {
			t.Errorf("%s: not on the chrome:\n%s", c.name, out)
		}
		if !strings.Contains(out, "esc close") {
			t.Errorf("%s: footer should end esc close:\n%s", c.name, out)
		}
	}
}
