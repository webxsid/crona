# TUI Shell Swap (Layout Redesign Step 1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the sidebar/header/help-bar shell with a tab line, optional sub-tab line, and bottom status line, mapping the existing 18 views onto 5 tabs unchanged, and lower the minimum terminal size to 60×20.

**Architecture:** A new `state.Tab` groups existing `View`s as sub-tabs. The model filters each tab's views through the existing availability rules (timer lock, protected rest) and passes the visible tabs/sub-tabs into `layout.State`. `layout.Render` draws the new chrome around the unchanged `viewrenderer.RenderContent`, which now gets the full terminal width. Digits `1`–`5` become global tab keys, `[`/`]` cycle sub-tabs within a tab, and per-view digit bindings are removed.

**Tech Stack:** Go, Bubble Tea, lipgloss, `charmbracelet/x/ansi`.

**Spec:** `docs/superpowers/specs/2026-10-17-tui-layout-design.md` (this plan covers Migration step 1 only).

## Global Constraints

- Tab order and keys: `1 Today`, `2 Issues`, `3 Insights`, `4 Log`, `5 System`.
- Tab → views (sub-tab order): Today = Daily, Summary, Session (`ViewSessionActive`), Away; Issues = Issues (`ViewDefault`), Meta; Insights = Rollup, Momentum, Wellbeing; Log = Sessions (`ViewSessionHistory`), Habits (`ViewHabitHistory`), Ops; System = Settings, Alerts, Reports, Config, Updates, Support.
- `MinWidth = 60`, `MinHeight = 20`.
- Chrome rows: tab line 1, sub-tab line 1 only when the current tab has ≥ 2 visible sub-tabs, status line 1. No header, no sidebar, no help bar, no floating toast.
- Content width is the full terminal width.
- Timer-lock and protected-rest availability rules are unchanged (`jumpAvailableViews`); a tab is visible only if it has ≥ 1 available view.
- Run Go commands from `tui/` with `GOCACHE=/tmp/crona-go-cache`. Run `make fmt` before each commit.
- Commit on the `brainstorm/tui-layout` branch.

## Review Focus

1. Digits typed into a dialog's text input (e.g. an estimate of `25`) must reach the input, not switch tabs. Covered by Task 2's `TestDigitsInDialogDoNotSwitchTabs`.
2. While a timer runs, pressing a locked tab's digit (e.g. `3`) is a no-op, and the tab line shows only Today and Log. Covered by Task 2's `TestTabKeyIgnoresLockedTab` and Task 3's `TestRenderTimerActiveShowsOnlyAllowedTabs`.
3. During protected rest the tab line shows Today (Away), Log, System, and `[`/`]` never lands on a hidden view. Covered by Task 2's `TestProtectedRestTabViews`.
4. A long status message at 60 columns must not wrap or push the frame past the terminal height. Covered by Task 3's `TestRenderLongStatusFitsOneLine`.
5. Every view still renders inside 80×24 without exceeding the frame (views were tuned for ≥ 113 columns). Covered by Task 3's `TestRenderAllViewsFitNarrowFrame`.

---

### Task 1: Tab model in `state`

**Files:**
- Create: `tui/internal/tui/state/tabs.go`
- Test: `tui/internal/tui/state/tabs_test.go`

**Interfaces:**
- Produces:
  - `type Tab string` with `TabToday = "today"`, `TabIssues = "issues"`, `TabInsights = "insights"`, `TabLog = "log"`, `TabSystem = "system"`.
  - `func TabOrder() []Tab`: the five tabs in key order (copy).
  - `func TabViews(tab Tab) []View`: sub-tab views in the order from Global Constraints (copy).
  - `func TabForView(view View) Tab`: owning tab; `TabToday` for unknown views.
  - `func TabLabel(tab Tab) string`: `"Today"`, `"Issues"`, `"Insights"`, `"Log"`, `"System"`.
  - `func ViewLabel(view View) string`: Daily→`"Daily"`, Summary→`"Summary"`, SessionActive→`"Session"`, Away→`"Away"`, Default→`"Issues"`, Meta→`"Meta"`, Rollup→`"Rollup"`, Momentum→`"Momentum"`, Wellbeing→`"Wellbeing"`, SessionHistory→`"Sessions"`, HabitHistory→`"Habits"`, Ops→`"Ops"`, Settings→`"Settings"`, Alerts→`"Alerts"`, Reports→`"Reports"`, Config→`"Config"`, Updates→`"Updates"`, Support→`"Support"`.

- [ ] **Step 1: Write the failing tests**

```go
func TestEveryViewBelongsToExactlyOneTab(t *testing.T) {
	all := []View{ViewAway, ViewSummary, ViewDefault, ViewDaily, ViewRollup, ViewMomentum, ViewMeta,
		ViewSessionHistory, ViewHabitHistory, ViewSessionActive, ViewOps, ViewWellbeing,
		ViewReports, ViewConfig, ViewSettings, ViewAlerts, ViewUpdates, ViewSupport}
	seen := map[View]Tab{}
	for _, tab := range TabOrder() {
		for _, v := range TabViews(tab) {
			if prev, dup := seen[v]; dup {
				t.Fatalf("%q in both %q and %q", v, prev, tab)
			}
			seen[v] = tab
		}
	}
	for _, v := range all {
		if seen[v] == "" { t.Fatalf("view %q has no tab", v) }
		if TabForView(v) != seen[v] { t.Fatalf("TabForView(%q)=%q want %q", v, TabForView(v), seen[v]) }
		if ViewLabel(v) == "" { t.Fatalf("view %q has no label", v) }
	}
}

func TestTabOrderAndInsightsViews(t *testing.T) {
	// assert TabOrder() == [today issues insights log system]
	// assert TabViews(TabInsights) == [ViewRollup ViewMomentum ViewWellbeing]
	// assert TabLabel(TabLog) == "Log"
}
```

- [ ] **Step 2: Run to verify failure**

Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/state/`
Expected: FAIL, `undefined: TabOrder`.

- [ ] **Step 3: Implement `tabs.go`** using package-level slices and maps alongside `viewOrder`/`viewPanes` in `types.go`.

- [ ] **Step 4: Run to verify pass.** Same command. Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add tui/internal/tui/state/tabs.go tui/internal/tui/state/tabs_test.go
git commit -m "feat(tui): add tab grouping for views"
```

---

### Task 2: Tab navigation in model and input

**Files:**
- Modify: `tui/internal/tui/model/navigation.go` (replace `nextWorkspaceView`/`nextActiveSessionView` bodies, add tab helpers)
- Modify: `tui/internal/tui/model/input.go:179-184` (deps wiring)
- Modify: `tui/internal/tui/input/input.go` (`Deps`, global `1`–`5`, `[`/`]`)
- Modify: `tui/internal/tui/input/handlers_core.go:16-30` (`handleCycleView`), add `handleJumpToTab`
- Modify: `tui/internal/tui/key_registry/{daily,meta,default,config,settings,reports,scratch}.go`: remove digit registrations and the now-unused handler parameters; update callers in `input/input.go`
- Modify: digit hints in `views/daily/panes.go:320,398`, `views/meta/{repos,streams,issues,habits,layout}.go`, `views/issues/layout.go:43,55`: drop the ` [N]` suffixes; `views/meta/layout.go:15,19` empty-state copy becomes `"No repo checked out — select a repo, then [c]"` / `"No stream checked out — select a stream, then [c]"`
- Test: `tui/internal/tui/model/navigation_test.go`, `tui/internal/tui/input/handlers_core_test.go`

**Interfaces:**
- Consumes: Task 1's `TabOrder`, `TabViews`, `TabForView`.
- Produces:
  - `func (m Model) visibleTabViews(tab uistate.Tab) []View`: `TabViews(tab)` filtered to members of `m.jumpAvailableViews()`, keeping tab order.
  - `func (m Model) visibleTabs() []uistate.Tab`: tabs with ≥ 1 visible view.
  - `func (m Model) tabTargetView(tab uistate.Tab) (View, bool)`: the current view if it's in `tab`, else the first visible view of `tab`; `false` if none.
  - `func (m Model) nextSubTabView(dir int) View`: cycles `visibleTabViews(TabForView(m.view))`; returns `m.view` when the tab has < 2 visible views.
  - `input.Deps.TabTargetView func(State, uistate.Tab) (uistate.View, bool)`; `Deps.NextSubTabView func(State, int) uistate.View`. Remove `NextActiveSessionView` and `NextWorkspaceView`.
  - `func handleJumpToTab(s State, deps Deps, tab uistate.Tab) (tea.Model, tea.Cmd, bool)`: no-op (handled=true, state unchanged) when the target is unavailable; otherwise it sets `ActiveView`, `ActivePane = deps.DefaultPane(view)`, and runs the same `EnterHabitHistoryView` branch as `handleCycleView`.
- Note: `jumpAvailableViews` already returns `[SessionActive, SessionHistory]` while a timer runs, and the protected list during rest. The session-active view only appears through it, so `visibleTabViews` must not use `availableViews`.

- [ ] **Step 1: Write the failing tests**

In `model/navigation_test.go`:
```go
func TestVisibleTabsWhenIdle(t *testing.T)        // visibleTabs() == all 5; visibleTabViews(TabToday) == [Daily, Summary]
func TestTabKeyIgnoresLockedTab(t *testing.T)     // timer State "running": visibleTabs() == [today, log]; tabTargetView(TabInsights) ok == false
func TestProtectedRestTabViews(t *testing.T)      // RestSpecificDates=[today]: visibleTabs()==[today, log, system]; visibleTabViews(TabSystem)==[Settings, Reports]; from ViewReports nextSubTabView(1)==ViewSettings
func TestNextSubTabViewWraps(t *testing.T)        // view=Wellbeing: nextSubTabView(1)==Rollup, nextSubTabView(-1)==Momentum
```
Replace `TestAvailableViewsFollowSidebarOrder` with nothing; its order is now covered by Task 1.

In `input/handlers_core_test.go`:
```go
func TestDigitSwitchesTab(t *testing.T)               // Handle(State{ActiveView: Daily}, key "3", deps with TabTargetView returning Rollup) → ActiveView==Rollup, ActivePane==PaneRollupDays
func TestBracketCyclesSubTab(t *testing.T)            // key "]" calls NextSubTabView(s, 1) and applies its result
func TestIssuesDigitsNoLongerSwitchSection(t *testing.T) // ActiveView Default, key "2" → TabTargetView called with TabIssues, DefaultIssueSection unchanged
```
In `model` (full routing through `Model.Update`):
```go
func TestDigitsInDialogDoNotSwitchTabs(t *testing.T) // open create_issue_default dialog, send "2"; m.view unchanged and m.dialog still set
```

- [ ] **Step 2: Run to verify failure**

Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/model/ ./internal/tui/input/`
Expected: FAIL, undefined `visibleTabs`, `TabTargetView`, etc.

- [ ] **Step 3: Implement the model helpers** in `navigation.go` and delete `nextWorkspaceView` and `nextActiveSessionView`.

- [ ] **Step 4: Implement input routing.** Register `"1"`…`"5"` in `RegisterGlobal` mapping to `TabOrder()[i]` via `handleJumpToTab`, which ignores hidden tabs. Make `handleCycleView` call `deps.NextSubTabView(s, dir)` regardless of timer state. Wire both deps in `model/input.go` through `m.applyInputState(state)`.

- [ ] **Step 5: Remove digit bindings** from the key_registry files and their call sites, and update the view hint strings listed under Files.

- [ ] **Step 6: Run the whole TUI module**

Run: `cd tui && GOCACHE=/tmp/crona-go-cache go build ./... && GOCACHE=/tmp/crona-go-cache go test ./...`
Expected: all `ok`. Fix any existing test that asserted digit bindings or `[N]` labels (`testsuite/input_meta_test.go`, `testsuite/default_scope_test.go`) to use `tab` instead.

- [ ] **Step 7: Commit**

```bash
git add tui/internal/tui
git commit -m "feat(tui): navigate with tab keys and sub-tab cycling"
```

---

### Task 3: New shell rendering

**Files:**
- Modify: `tui/internal/tui/layout/render.go`: replace `renderHeader`, `renderSidebar`, `renderSidebarItem`, `renderHelpBar`, `renderBody`, `renderStatusToast` and the status-overlay branch of `Render`; update `ContentHeight`, `contentHeightFromChrome`, constants; delete `SidebarWidth`
- Modify: `tui/internal/tui/model/view.go`: fill the new `State` fields; `mainContentWidth()` returns `m.width`
- Modify: `tui/internal/tui/model/dialogs.go:604` (`openHelpDialog`): add `"Version " + versionpkg.Current()` (plus `"  · Dev"` when `m.isDevMode()`) as the help dialog's meta prefix
- Test: `tui/internal/tui/layout/render_test.go` (rewrite the four sidebar tests), `tui/internal/tui/testsuite/compact_modes_test.go`

**Interfaces:**
- Consumes: Task 1's `TabLabel`, `ViewLabel`, `TabForView`; Task 2's `visibleTabs`, `visibleTabViews`.
- Produces:
  - `layout.State` gains `Tabs []uistate.Tab` and `SubTabs []uistate.View`; loses nothing else (keep `GlobalActions`, `HeaderState`, `StatusMsg`, `StatusErr`).
  - `func renderTabLine(state State) string`: `" 1 Today  2 Issues …"`. The active tab uses `chrome.StyleCursor`, others `chrome.StyleDim`. Each tab shows its key from `TabOrder()` position, so a hidden tab's number is skipped but numbers don't shift. Right-aligned: `repo › stream`, plus `" · dev"` when `IsDevMode`. Truncate the right side first when the line overflows.
  - `func renderSubTabLine(state State) string`: `" Rollup · Momentum · Wellbeing"`, active in `chrome.StyleHeader`, others dim, joined by dim `" · "`; returns `""` when `len(SubTabs) < 2`.
  - `func renderStatusLine(state State) string`: exactly one row of `Width` cells. Left: `viewchrome.HeaderSessionLine(ViewTheme(), state.HeaderState)` when `TimerActive`, else `GlobalActions` joined by three spaces. Right: the status message (`chrome.StyleError` if `StatusErr`, else `chrome.StyleSelected`) when `StatusMsg != ""`, else dim `"? help  ctrl+c quit"`. The right side wins; the left is truncated to fit with ≥ 1 space gap.
  - `func chromeRows(state State) int`: `2 + (1 if len(SubTabs) ≥ 2)`; `ContentHeight(state) = max(4, Height - chromeRows(state))`.
  - Protected mode uses the same chrome (its tab/sub-tab lists already come pre-filtered).

- [ ] **Step 1: Write the failing tests** in `layout/render_test.go`, using a helper `baseState(view, width, height)`:

```go
func TestRenderTabLineMarksActiveTab(t *testing.T)            // Daily at 120x36, Tabs=all 5 → first line contains "1 Today" and "5 System"; no "DASHBOARD", no "version:"
func TestRenderSubTabLineOnlyForMultiViewTabs(t *testing.T)   // Rollup with SubTabs=[Rollup,Momentum,Wellbeing] → line 2 contains "Momentum"; Ops with SubTabs=[Ops] → ContentHeight == Height-2
func TestRenderTimerActiveShowsOnlyAllowedTabs(t *testing.T)  // Tabs=[today,log], TimerActive → first line contains "1 Today" and "4 Log", not "Insights"
func TestRenderLongStatusFitsOneLine(t *testing.T)            // Width 60, StatusMsg = strings.Repeat("x", 200) → lipgloss.Height(out)==Height and last line width == 60
func TestRenderMinimumSize(t *testing.T)                      // 59x24 and 80x19 → minimum-size warning; 60x20 → contains "1 Today"
func TestRenderAllViewsFitNarrowFrame(t *testing.T)           // for every view in every tab at 80x24: every line's ansi.StringWidth ≤ 80 and line count == 24
```
Keep `TestRenderDialogUsesBlankScreen`, `TestRenderMomentumDetailDialogShowsModalContent` and `TestRenderPadsFullViewportAcrossDailyFrames`, updating only their widths.

- [ ] **Step 2: Run to verify failure**

Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/layout/`
Expected: FAIL on the new tests (no `Tabs` field).

- [ ] **Step 3: Implement the render functions and constants** in `layout/render.go`; `Render` composes `tabLine`, `subTabLine` (if any), `RenderContent` at full width and `ContentHeight`, then `statusLine`. Dialog and session overlays keep their current `lipgloss.Place` behaviour. Remove `renderOverlay`/`overlayHeight` if they become unused.

- [ ] **Step 4: Wire the model** in `model/view.go` `layoutState()`: `Tabs: m.visibleTabs()`, `SubTabs: m.visibleTabViews(uistate.TabForView(state.View))` (after the Away substitution, so rest mode shows Away's tab). Pass the same fields into the `contentHeight()` call so content height matches. Add the version to `openHelpDialog`.

- [ ] **Step 5: Run the whole module**

Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`
Expected: all `ok`. Update `testsuite/compact_modes_test.go` and other tests that assert on sidebar or header text to assert on the tab line instead.

- [ ] **Step 6: Visual check**

Run `make run-tui` in an 80×24 tmux pane and a 200×50 window. Cycle every tab with `1`–`5` and `[`/`]`, start and stop a timer, and open `?`. Note views that look cramped at 80 columns (expected; step 2 of the migration fixes them) in the commit body.

- [ ] **Step 7: Commit**

```bash
git add tui/internal/tui
git commit -m "feat(tui): replace sidebar shell with tab and status lines"
```

---

### Task 4: Keymap documentation

**Files:**
- Modify: `docs/reference/tui-keymap-reference.md`

- [ ] **Step 1: Update Global Bindings:** add `1`–`5` (switch tab: Today, Issues, Insights, Log, System); `[`/`]` becomes "cycle sub-tabs within the current tab"; `tab`/`shift+tab` unchanged. Add a short "Tabs" table listing each tab's sub-tabs from Global Constraints. Remove mentions of the sidebar; note that Meta and Daily panes switch with `tab` and Issues open/completed with `tab`.

- [ ] **Step 2: Verify:** `make release-check`. Expected: passes (docs metadata intact).

- [ ] **Step 3: Commit**

```bash
git add docs/reference/tui-keymap-reference.md
git commit -m "docs: document tab navigation keys"
```
