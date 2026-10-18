package layout

import (
	"strings"
	"testing"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	dialogstate "crona/tui/internal/tui/dialogs/controller"
	uistate "crona/tui/internal/tui/state"
	viewtypes "crona/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func shellStateForTest(view uistate.View, width, height int) State {
	cursors := map[string]int{}
	filters := map[string]string{}
	for _, tab := range uistate.TabOrder() {
		for _, v := range uistate.TabViews(tab) {
			for _, pane := range uistate.ViewPanes(v) {
				cursors[string(pane)] = 0
				filters[string(pane)] = ""
			}
		}
	}
	state := State{
		Width:      width,
		Height:     height,
		View:       view,
		Pane:       uistate.DefaultPane(view),
		RepoName:   "crona",
		StreamName: "tui",
		Tabs:       uistate.TabOrder(),
		SubTabs:    uistate.TabViews(uistate.TabForView(view)),
		ContentState: viewtypes.ContentState{
			View:    string(view),
			Pane:    string(uistate.DefaultPane(view)),
			Width:   width,
			Cursors: cursors,
			Filters: filters,
		},
	}
	state.ContentState.Height = ContentHeight(state)
	return state
}

func plainLines(rendered string) []string {
	return strings.Split(ansi.Strip(rendered), "\n")
}

func TestRenderTabLineMarksActiveTab(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 120, 36)
	state.SubTabs = []uistate.View{uistate.ViewDaily, uistate.ViewSessionActive}
	lines := plainLines(Render(state))
	if !strings.Contains(lines[0], "[1] Today") || !strings.Contains(lines[0], "[5] System") {
		t.Fatalf("expected tab line with all tabs, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "crona › tui") {
		t.Fatalf("expected repo › stream on tab line, got %q", lines[0])
	}
	rendered := strings.Join(lines, "\n")
	for _, gone := range []string{"DASHBOARD", "version:", "stop daemon"} {
		if strings.Contains(rendered, gone) {
			t.Fatalf("expected old chrome %q to be gone", gone)
		}
	}
}

func TestRenderSubTabLineOnlyForMultiViewTabs(t *testing.T) {
	state := shellStateForTest(uistate.ViewReview, 120, 36)
	lines := plainLines(Render(state))
	if !strings.Contains(lines[1], "Momentum") || !strings.Contains(lines[1], "Review") {
		t.Fatalf("expected sub-tab line, got %q", lines[1])
	}
	if got := ContentHeight(state); got != 36-4 {
		t.Fatalf("expected content height 32 with sub-tabs and the date row, got %d", got)
	}
	ops := shellStateForTest(uistate.ViewOps, 120, 36)
	ops.SubTabs = []uistate.View{uistate.ViewOps}
	if got := ContentHeight(ops); got != 36-2 {
		t.Fatalf("expected content height 34 without sub-tabs, got %d", got)
	}
}

func datedShellState(view uistate.View, width, height int) State {
	state := shellStateForTest(view, width, height)
	state.ContentState.DashboardDate = "2026-05-27"
	state.ContentState.SummaryDate = "2026-05-27"
	state.ContentState.MomentumDate = "2026-05-27"
	state.ContentState.WellbeingDate = "2026-05-27"
	state.ContentState.RollupStartDate = "2026-05-21"
	state.ContentState.RollupEndDate = "2026-05-27"
	return state
}

func TestRenderTabLineNumbersTabs(t *testing.T) {
	line := plainLines(Render(shellStateForTest(uistate.ViewDaily, 120, 36)))[0]
	for _, want := range []string{"[1] Today", "[2] Issues", "[5] System"} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected %q on the tab line, got %q", want, line)
		}
	}
}

func TestRenderSubTabLineFramedByKeys(t *testing.T) {
	line := strings.TrimRight(plainLines(Render(datedShellState(uistate.ViewDaily, 120, 36)))[1], " ")
	if !strings.HasPrefix(strings.TrimSpace(line), "[[] Daily") || !strings.HasSuffix(line, "[]]") {
		t.Fatalf("expected the sub-tabs framed by [ and ], got %q", line)
	}
	if strings.Contains(line, "2026-05-27") {
		t.Fatalf("expected the date off the sub-tab line, got %q", line)
	}
}

func TestRenderDateRowOnDatedViews(t *testing.T) {
	for view, keys := range map[uistate.View]string{
		uistate.ViewDaily:    "[C] calendar",
		uistate.ViewMomentum: "[g] today",
		uistate.ViewReview:   "[E] calendar",
	} {
		line := plainLines(Render(datedShellState(view, 140, 36)))[2]
		if !strings.Contains(line, "2026-05-27") || !strings.Contains(line, "next [.]") || !strings.Contains(line, keys) {
			t.Fatalf("%s: expected the date row with %q, got %q", view, keys, line)
		}
		if view == uistate.ViewReview && !strings.Contains(line, "2026-05-21") {
			t.Fatalf("review: expected its range start on the date row, got %q", line)
		}
	}
	if line := plainLines(Render(shellStateForTest(uistate.ViewSettings, 140, 36)))[2]; strings.Contains(line, "next [.]") {
		t.Fatalf("expected no date row on Settings, got %q", line)
	}
}

func TestRenderDateRowKeepsDateWhenNarrow(t *testing.T) {
	if line := plainLines(Render(datedShellState(uistate.ViewDaily, 70, 30)))[2]; !strings.Contains(line, "2026-05-27") {
		t.Fatalf("expected the date kept on a narrow date row, got %q", line)
	}
}

func TestContentHeightCountsDateRow(t *testing.T) {
	if got := ContentHeight(datedShellState(uistate.ViewDaily, 120, 36)); got != 36-4 {
		t.Fatalf("expected tabs, sub-tabs, date row and status to take 4 rows, got content %d", got)
	}
}

func TestRenderTimerActiveShowsOnlyAllowedTabs(t *testing.T) {
	state := shellStateForTest(uistate.ViewSessionActive, 120, 36)
	state.TimerActive = true
	state.Tabs = []uistate.Tab{uistate.TabToday, uistate.TabLog}
	state.SubTabs = []uistate.View{uistate.ViewSessionActive}
	lines := plainLines(Render(state))
	if !strings.Contains(lines[0], "[1] Today") || !strings.Contains(lines[0], "[4] Log") {
		t.Fatalf("expected today and log tabs, got %q", lines[0])
	}
	if strings.Contains(lines[0], "Insights") {
		t.Fatalf("expected locked tabs hidden, got %q", lines[0])
	}
}

func TestRenderLongStatusFitsOneLine(t *testing.T) {
	state := shellStateForTest(uistate.ViewOps, 60, 24)
	state.SubTabs = []uistate.View{uistate.ViewOps}
	state.StatusMsg = strings.Repeat("x", 200)
	rendered := Render(state)
	if got := lipgloss.Height(rendered); got != 24 {
		t.Fatalf("expected frame height 24, got %d", got)
	}
	lines := plainLines(rendered)
	last := lines[len(lines)-1]
	if ansi.StringWidth(last) != 60 || !strings.Contains(last, "xxx") {
		t.Fatalf("expected status message on 60-col last line, got %q", last)
	}
}

func TestRenderMinimumSize(t *testing.T) {
	for _, size := range [][2]int{{59, 24}, {80, 19}} {
		rendered := Render(shellStateForTest(uistate.ViewDaily, size[0], size[1]))
		if !strings.Contains(rendered, "Terminal Too Small") {
			t.Fatalf("expected minimum-size warning at %dx%d", size[0], size[1])
		}
	}
	rendered := Render(shellStateForTest(uistate.ViewDaily, 60, 20))
	if !strings.Contains(ansi.Strip(rendered), "[1] Today") {
		t.Fatal("expected shell at 60x20")
	}
}

func TestRenderAllViewsFitNarrowFrame(t *testing.T) {
	for _, tab := range uistate.TabOrder() {
		for _, view := range uistate.TabViews(tab) {
			rendered := Render(shellStateForTest(view, 80, 24))
			lines := strings.Split(rendered, "\n")
			if len(lines) != 24 {
				t.Fatalf("%s: expected 24 lines, got %d", view, len(lines))
			}
			for i, line := range lines {
				if got := ansi.StringWidth(line); got > 80 {
					t.Fatalf("%s: line %d width %d > 80", view, i, got)
				}
			}
		}
	}
}

func TestRenderDialogUsesBlankScreen(t *testing.T) {
	rendered := Render(State{
		Width:      140,
		Height:     44,
		View:       uistate.ViewDaily,
		Pane:       uistate.PaneIssues,
		DialogOpen: true,
		DialogState: dialogstate.State{
			Kind:      "view_entity",
			Width:     140,
			ViewTitle: "Keys",
			ViewName:  "Keyboard Shortcuts",
			ViewMeta:  "Hint Press ? or esc to close",
			ViewBody:  "Global\n[v] switch views",
		},
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewDaily),
			Pane:   string(uistate.PaneIssues),
			Width:  118,
			Height: 44,
			Cursors: map[string]int{
				string(uistate.PaneIssues): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneIssues): "",
			},
		},
	})

	if strings.Contains(rendered, "Habit History") || strings.Contains(rendered, "Daily") {
		t.Fatalf("expected dialog render to exclude base screen content, got %q", rendered)
	}
	if !strings.Contains(rendered, "Keyboard Shortcuts") {
		t.Fatalf("expected dialog render to show the modal content, got %q", rendered)
	}
}

func TestRenderMomentumDetailDialogShowsModalContent(t *testing.T) {
	rendered := Render(State{
		Width:      140,
		Height:     44,
		View:       uistate.ViewMomentum,
		Pane:       uistate.PaneMomentumCards,
		DialogOpen: true,
		DialogState: dialogstate.State{
			Kind:      "view_momentum_detail",
			Width:     140,
			ViewTitle: "Momentum Detail",
			ViewName:  "Delivery Context All",
			ViewMeta:  "Kind Context   Match All",
			ViewBody:  "Current Bucket\nStatus Met",
			HabitStreakDraft: sharedtypes.HabitStreakDefinition{
				ID:   "momentum-1",
				Name: "Delivery Context All",
			},
			ChoiceItems:   []string{"2026-06-19  Deep work"},
			ChoiceDetails: []string{"Work/app   1h12m"},
		},
		ContentState: viewtypes.ContentState{
			View:   string(uistate.ViewMomentum),
			Pane:   string(uistate.PaneMomentumCards),
			Width:  118,
			Height: 44,
			Cursors: map[string]int{
				string(uistate.PaneMomentumCards): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneMomentumCards): "",
			},
		},
	})

	for _, want := range []string{"Momentum Detail", "Delivery Context All", "Contributors", "Deep work"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected momentum detail render to contain %q, got %q", want, rendered)
		}
	}
}

func TestRenderPadsFullViewportAcrossDailyFrames(t *testing.T) {
	first := Render(dailyLayoutStateForTest(140, 44, 96, 38, "2026-05-14"))
	second := Render(dailyLayoutStateForTest(140, 44, 96, 38, "2026-02-03"))

	for _, rendered := range []string{first, second} {
		lines := strings.Split(rendered, "\n")
		if len(lines) != 44 {
			t.Fatalf("expected rendered frame height 44, got %d", len(lines))
		}
		for _, line := range lines {
			if got := lipgloss.Width(line); got != 140 {
				t.Fatalf("expected padded line width 140, got %d", got)
			}
		}
	}
}

func dailyLayoutStateForTest(
	width, height, contentWidth, contentHeight int,
	date string,
) State {
	return State{
		Width:  width,
		Height: height,
		View:   uistate.ViewDaily,
		Pane:   uistate.PaneIssues,
		ContentState: viewtypes.ContentState{
			View:          string(uistate.ViewDaily),
			Pane:          string(uistate.PaneIssues),
			Width:         contentWidth,
			Height:        contentHeight,
			DashboardDate: date,
			DailySummary:  &api.DailyIssueSummary{Date: date},
			Cursors: map[string]int{
				string(uistate.PaneIssues): 0,
				string(uistate.PaneHabits): 0,
			},
			Filters: map[string]string{
				string(uistate.PaneIssues): "",
				string(uistate.PaneHabits): "",
			},
		},
	}
}

func TestRenderTabLineKeepsTabsWithLongContext(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 60, 24)
	state.IsDevMode = true
	state.RepoName = "crona-monorepo"
	state.StreamName = "feature/tui-layout-redesign-long-name"
	line := plainLines(Render(state))[0]
	if !strings.Contains(line, "[5] System") {
		t.Fatalf("expected all tabs to survive a long context, got %q", line)
	}
	// The numbered tabs fill the minimum width; the context shows once there is room.
	state.Width = 80
	if line := plainLines(Render(state))[0]; !strings.Contains(line, "crona") {
		t.Fatalf("expected truncated context to remain visible at 80, got %q", line)
	}
}

func TestStatusLineShowsPaneActionsFirst(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 160, 30)
	state.GlobalActions = []string{"[1-5] tabs", "[[/]] sub-tabs"}
	state.PaneActions = []string{"[s] set status"}
	line := plainLines(Render(state))[29]
	pane, global := strings.Index(line, "[s] set status"), strings.Index(line, "[1-5]")
	if pane < 0 || global < 0 || pane > global {
		t.Fatalf("expected pane actions before global actions, got %q", line)
	}
}

func TestTabLineContextAll(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 120, 30)
	state.RepoName, state.StreamName = "", ""
	line := plainLines(Render(state))[0]
	if !strings.HasSuffix(strings.TrimSpace(line), "All") || strings.Contains(line, "-") {
		t.Fatalf("expected All context, got %q", line)
	}
}

func TestTabLineContextRepoOnly(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 120, 30)
	state.StreamName = ""
	line := strings.TrimSpace(plainLines(Render(state))[0])
	if !strings.HasSuffix(line, "crona") || strings.Contains(line, "›") {
		t.Fatalf("expected repo-only context, got %q", line)
	}
}

func TestDailyDateRowLeavesDayKeyOff(t *testing.T) {
	// The bottom row offers [S] when no detail pane shows; the date row never does.
	if line := plainLines(Render(datedShellState(uistate.ViewDaily, 120, 36)))[2]; strings.Contains(line, "[S]") {
		t.Fatalf("expected [S] off the date row, got %q", line)
	}
	if line := plainLines(Render(datedShellState(uistate.ViewDaily, 200, 36)))[2]; strings.Contains(line, "[S]") {
		t.Fatalf("expected [S] left to the detail pane when it shows, got %q", line)
	}
	away := datedShellState(uistate.ViewDaily, 120, 36)
	away.ContentState.HistoricalAwayDate = "2026-05-27"
	if line := plainLines(Render(away))[2]; strings.Contains(line, "[S]") {
		t.Fatalf("expected no [S] on a historical away date, got %q", line)
	}
}

func TestDateRowNavSitsRightInKeyOrder(t *testing.T) {
	for view, calendar := range map[uistate.View]string{uistate.ViewDaily: "[C] calendar", uistate.ViewReview: "[E] calendar"} {
		line := plainLines(Render(datedShellState(view, 160, 36)))[2]
		at := -1
		for _, want := range []string{calendar, "[g] today", "[,] prev", "2026-05-27", "next [.]"} {
			next := strings.Index(line, want)
			if next <= at {
				t.Fatalf("%s: expected %q after the previous key, got %q", view, want, line)
			}
			at = next
		}
		if !strings.HasSuffix(line, "next [.] ") {
			t.Fatalf("%s: expected the navigation one cell from the right edge, got %q", view, line)
		}
	}
	// The window keys live on the strip below, at its right.
	lines := plainLines(Render(datedShellState(uistate.ViewReview, 160, 36)))
	if strings.Contains(lines[2], "[+/-/0]") || !strings.Contains(lines[3], "[+/-/0]") {
		t.Fatalf("expected the window key under the date nav, got %q / %q", lines[2], lines[3])
	}
}
