# Issues Tab: Scope Tree and Habits Sub-tab — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Meta view with a scope tree beside the Issues list and a Habits sub-tab, and give both Daily's in-pane key hints.

**Architecture:** A new `scope` package owns the tree (nodes from repos + all streams) and the browse scope (a repo/stream pair, zero = All). The model keeps `browseScope` beside the checked-out context; selection filters Issues and habit definitions by it. Both sub-tabs render a shared scope pane, a list pane and Daily's detail pane, under a two-row strip.

**Tech Stack:** Go, Bubble Tea, lipgloss, x/ansi.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-issues-tab-design.md`

## Global Constraints

- Sub-tabs are `Issues · Habits`; `ViewMeta` and `views/meta` are deleted.
- Summary strip: at most 2 rows; column groups drop right-first when narrow; every row exactly `width` cells.
- Widths: wide ≥160 shows scope │ list │ detail; medium drops detail; narrow <100 collapses scope to one row `Scope ‹ <node> ›`.
- Every hint shown is a key the input layer honours in that state; hints are whole (`viewchrome.FitHints`, `viewchrome.KeyHint`).
- Bottom row: globals only (as Daily). No `[S]` on these views.
- No daemon changes. UI commits only. gofmt touched files; never `make fmt`.
- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`.

## Review Focus

1. Browsed repo/stream deleted (or streams reloaded) → scope falls back to All and the scope cursor stays on a valid node.
2. All streams arrive after the tree first renders → the cursor follows the browse scope's node, not a stale index.
3. Switching Open/Done, or a scope change, leaves the Issues cursor in range (empty Done list, filter active).
4. No repos at all → tree shows only `All` and `[a] new repo`; Issues list still works.
5. `[c]` on the `All` node does nothing visible and offers no hint; `[a]` on `All` creates a repo.

---

### Task 1: Scope data — all streams, browse scope, scoped lists

**Files:**
- Create: `tui/internal/tui/scope/scope.go`, `scope/scope_test.go`
- Modify: `tui/internal/tui/commands/load.go`, `commands/messages.go` (all-streams loader), `dispatch/messages.go` + `dispatch/events.go` (store; reload where `LoadRepos` reloads), `model/core.go`, `model/update.go`, `model/layout_state.go`, `selection/snapshot.go`, `selection/snapshot_scoped.go`, `views/types/main.go`
- Test: `scope/scope_test.go`, `selection/*_test.go`, `model/*_test.go`

**Interfaces:**
- Produces (package `scope`):
  - `type Scope struct{ RepoID, StreamID int64 }` — zero value is All; `StreamID != 0` implies `RepoID != 0`. `func (s Scope) IsAll() bool`.
  - `func (s Scope) Contains(repoID, streamID int64) bool`.
  - `type Node struct{ Scope Scope; Name string; Depth int }` — Depth 0 for All and repos, 1 for streams.
  - `func Nodes(repos []api.Repo, streams []api.Stream) []Node` — `All`, then repos in given order, each followed by its streams.
  - `func IndexOf(nodes []Node, s Scope) int` — exact match, else the repo's node, else 0.
  - `func FromContext(ctx *api.ActiveContext) Scope`.
- Produces: `commands.LoadAllStreams(c) tea.Cmd` → `commands.AllStreamsLoadedMsg{Streams []api.Stream}` (lists repos, then each repo's streams). Model field `allStreams`; `ContentState.AllStreams`, `ContentState.AllHabits []api.HabitWithMeta`, `ContentState.BrowseScope scope.Scope`. Snapshot fields `BrowseScope`, `AllStreams`, `AllHabits`, `DefaultIssueSection`.
- Produces: `selection.ScopedHabits(s Snapshot) []api.HabitWithMeta` (AllHabits within BrowseScope); `DefaultScopedIssues` filters by `BrowseScope` (no longer by Context).
- Produces: model `browseScope scope.Scope`; `func (m *Model) enterIssuesTab()` sets it from the context and the scope cursor to `IndexOf`.

- [ ] **Step 1: Failing tests**
  - `TestNodesNestStreamsUnderRepos`: repos `crona`(1), `dotfiles`(3); streams `tui`(2,repo1), `daemon`(4,repo1) → names `All, crona, tui, daemon, dotfiles`, depths `0,0,1,1,0`.
  - `TestIndexOfFallsBackToRepoThenAll`: stream 9 of repo 1 missing → index of `crona`; repo 7 missing → 0.
  - `TestScopeContains`: All contains anything; `{1,0}` contains `(1,2)` not `(3,4)`; `{1,2}` contains `(1,2)` not `(1,4)`.
  - `TestDefaultScopedIssuesFollowBrowseScope`: context stream 2, BrowseScope `{1,4}` → only stream-4 issues.
  - `TestScopedHabitsFollowBrowseScope`.
  - `TestEnteringIssuesTabResetsBrowseScope`: model with context stream 2, browseScope `{3,0}`; press `2` → browseScope `{1,2}`.
  - `TestAllStreamsLoadedStored` (dispatch).
- [ ] **Step 2:** Run; expect FAIL (undefined `scope`, fields).
- [ ] **Step 3:** Implement. `enterIssuesTab` runs from the tab switch (digit and `[`/`]`) when the target tab is `TabIssues` and the previous view was not in it. Startup and repo/stream events also issue `LoadAllStreams`.
- [ ] **Step 4:** Run suite; expect PASS.
- [ ] **Step 5:** Commit `feat(tui): track a browse scope over all repos and streams`.

### Task 2: Scope pane — tree, keys, narrow row

**Files:**
- Create: `tui/internal/tui/views/scopepane/scopepane.go`, `scopepane_test.go`
- Modify: `state/types.go` (`PaneScope Pane = "scope"`; `ViewPanes[ViewDefault] = {PaneScope, PaneIssues}`; default pane Issues), `input/input.go`, `input/handlers_core.go`, `model/selection_actions.go`, `model/navigation.go`, `model/input.go`, `selection/snapshot.go`
- Test: `views/scopepane/scopepane_test.go`, `input/*_test.go` or `testsuite`, `model/*_test.go`

**Interfaces:**
- Consumes: Task 1's `scope` package, `ContentState.AllStreams/BrowseScope`.
- Produces: `scopepane.Render(theme, state types.ContentState, counts map[scope.Scope]int, width, height int) string` (tree section titled `Scope`); `scopepane.RenderRow(theme, state, width int) string` (narrow one-row form); `scopepane.IssueCounts(issues []api.IssueWithMeta) map[scope.Scope]int` (open issues per All/repo/stream); `scopepane.HabitCounts(habits []api.HabitWithMeta) map[scope.Scope]int`.
- Produces: `selection.SelectedScopeNode(s Snapshot) (scope.Node, bool)` (PaneScope cursor in ViewDefault or ViewHabitList). Replaces `SelectedMetaRepo`/`SelectedMetaStream` for checkout, create, edit, delete and details.

- [ ] **Step 1: Failing tests**
  - `TestScopeTreeRows`: rows `All 24`, `▶ crona 18 ●`, `    tui 9 ●`, …; `●` on the checked-out repo and stream only.
  - `TestScopeActionLineOnlyWhenFocused`: focused cursor on `crona` → line under it `[c] check out  [e] edit  [D] delete  [enter] open`; unfocused → none; on `All` → none.
  - `TestScopeCreateLine`: cursor on `All` → `[a] new repo`; on `crona` or `tui` → `[a] new stream`.
  - `TestScopeRowNarrow`: `RenderRow` at 80 → `Scope ‹ crona › tui ›` with `●`.
  - `TestScopeCursorMovesBrowseScope`: PaneScope, `down` → browseScope is the next node; context unchanged.
  - `TestScopeCheckoutUsesNode`: `c` on a stream node returns the stream checkout cmd; on `All` → not handled.
  - `TestScopeCreateOpensRepoOrStream`: `a` on `All` → `create_repo` dialog; on `tui` → `create_stream` with repo `crona`.
  - `TestScopeCursorFollowsScopeWhenStreamsLoad` (Review Focus 2) and `TestDeletedBrowsedStreamFallsBackToAll` (Review Focus 1).
- [ ] **Step 2:** Run; expect FAIL.
- [ ] **Step 3:** Implement. On PaneScope, `up/down` (narrow: `left/right`) move the cursor and set `browseScope = nodes[cursor].Scope`, then clamp the list pane's cursor. After `AllStreamsLoadedMsg`/`ReposLoadedMsg`, set the scope cursor to `IndexOf(nodes, browseScope)` and, if the browsed node is gone, `browseScope = scope.Scope{}`. `e`/`D`/`enter` on a node open the existing repo/stream edit, delete and details dialogs.
- [ ] **Step 4:** Run suite; expect PASS.
- [ ] **Step 5:** Commit `feat(tui): browse repos and streams in a scope tree`.

### Task 3: Issues sub-tab

**Files:**
- Modify: `views/issues/layout.go`, `views/issues/panes.go`, `views/issues/summary.go`, `views/issues/render.go` (rewrite), `input/handlers_core.go` (`handleCyclePane` loses the ViewDefault special case; `handleCycleIssueSection` handles ViewDefault Open/Done), `model/navigation.go` (`setDefaultIssueSection` resets cursor to 0), `selection/snapshot.go` (ViewDefault PaneIssues indices are the section's only)
- Create: `views/issues/hints.go`
- Test: `views/issues/panes_test.go` (replace), `input`/`model` tests

**Interfaces:**
- Consumes: `scopepane.Render/RenderRow/IssueCounts`, Daily's detail (`daily` must export `IssueDetailLines(theme, state, issue api.Issue, width int) []string`; move it to `views/issuecore` if import cycles appear), `viewchrome.KeyHint/FitHints/RenderSection/JoinColumns`.
- Produces: `issues.StripRows(theme, state, width int) []string` (≤2 rows).

- [ ] **Step 1: Failing tests**
  - `TestIssuesStripTwoRows`: rows start `Open 4` / `Closed 1`; contain `in progress 1`, `done 1`, `today 0`, `overdue 2`; exactly 2 rows; each `width` cells at 60/100/160; at 60 no `Due`.
  - `TestIssuesListSelectorAndCreateLine`: `Open  Done` selector, then `[a] new issue  [A] ad hoc`; `[A]` absent while a timer runs.
  - `TestIssuesCardActionLine`: focused selected card's second line has `[s] status`, `[d] due`, `[e] edit`, not its meta; unfocused shows meta; timer running hides `[f]`, `[m]`, `[e]`, `[D]`.
  - `TestIssuesLayoutByWidth`: 200 → `Scope`, list, detail title; 120 → no detail; 80 → `Scope ‹` row, no tree.
  - `TestArrowsSwitchOpenDoneAndTabCyclesPanes`: on PaneIssues `right` → section completed, cursor 0; `tab` → PaneScope.
  - `TestDoneSectionListsOnlyClosed` (Review Focus 3): cursor in range after switching to an empty Done.
  - `TestCreateIssueDefaultsToBrowseScope`: context stream `tui`, browse scope `crona › daemon`; `a` on PaneIssues → create dialog's repo/stream labels are `crona` / `daemon` (seed the dialog from an `ActiveContext` built from the browse scope).
- [ ] **Step 2:** Run; expect FAIL.
- [ ] **Step 3:** Implement. Cards reuse Daily's two-line shape (title + date; `context | worked/est | status`). Delete the bordered boxes, `RenderPaneActionLine` use and table rows.
- [ ] **Step 4:** Run suite; expect PASS.
- [ ] **Step 5:** Commit `feat(tui): rebuild the Issues view with a scope tree and in-pane keys`.

### Task 4: Habits sub-tab and Meta removal

**Files:**
- Create: `views/habitlist/` (`view.go`, `render.go`, `render_test.go`)
- Modify: `state/types.go`, `state/tabs.go` (`ViewHabitList View = "habit_list"`, label `Habits`, panes `{PaneScope, PaneHabits}`, default PaneHabits; `TabIssues: {ViewDefault, ViewHabitList}`), `views/renderer/main.go`, `input/input.go` (Meta's PaneHabits bindings move to ViewHabitList; `y` dropped), `selection/snapshot.go` (`SelectedHabit` reads `ScopedHabits` in ViewHabitList), `model/*`, `dispatch/*`, `dialogs/*`, `tui.go`, `views/chrome/actions.go` — every `ViewMeta` / `"meta"` reference
- Delete: `views/meta/`
- Test: `views/habitlist/render_test.go`, `state/tabs_test.go`, `testsuite/input_meta_test.go` (port to the scope pane or delete)

**Interfaces:**
- Consumes: Tasks 1–2 (`ScopedHabits`, `scopepane.Render/RenderRow/HabitCounts`), Daily's habit detail (export `HabitDetailLines`, taking `api.HabitWithMeta`-compatible input; adapt via a small converter if needed).
- Produces: `habitlist` view registered in `views/renderer`.

- [ ] **Step 1: Failing tests**
  - `TestIssuesTabSubTabs`: `TabViews(TabIssues) == {ViewDefault, ViewHabitList}`; label `Habits`.
  - `TestHabitStripTwoRows`: `Active 5  paused 1  due today 3` / `Done today 2  failed 0  best streak 12d` (counts from fixture), ≤2 rows, width-exact.
  - `TestHabitListCardsAndActions`: `[a] new habit`; card `Walk` then `crona › tui | weekdays | 30m`; focused selected card's second line `[enter] open  [e] edit  [D] delete  [m] log`; `[m]` and `[D]` hidden while a timer runs.
  - `TestCreateHabitDefaultsToBrowseScope`: as for issues, via the `create_habit` dialog.
  - `TestHabitListKeysActOnScopedHabit`: `e` on the selected habit opens its edit dialog with that habit's ID.
  - `TestNoMetaView`: `grep`-equivalent — `uistate.ViewPanes` has no `"meta"` key; view-jump list and help sheet have no `Meta`.
- [ ] **Step 2:** Run; expect FAIL.
- [ ] **Step 3:** Implement; delete `views/meta`. Run `grep -rn 'ViewMeta\|"meta"' tui/internal` → no matches (except unrelated `"meta"` strings, which the ledger notes).
- [ ] **Step 4:** Run suite; expect PASS.
- [ ] **Step 5:** Commit `feat(tui): add the Habits sub-tab and drop the Meta view`.

### Task 5: Bottom row, help and spec notes

**Files:** `views/chrome/actions.go` (ContextualActions: `issues` pane on ViewDefault, `habits` on ViewHabitList and `scope` return nil for the status line — keys live in panes), `model/view.go` (paneActions nil on ViewDefault and ViewHabitList), `model/dialogs.go` (help "Current View" lists the scope/list keys for these views), `docs/superpowers/specs/2026-10-17-tui-layout-design.md` (Issues tab section).

- [ ] **Step 1: Failing tests** — `TestIssuesBottomRowIsGlobalsOnly` and `TestHabitsBottomRowIsGlobalsOnly` (no `[s]`, `[c] check out`, `[a]`; has `[?] all keys`); `TestHelpSheetListsScopeKeys` on ViewDefault (`[c] check out`, `[a] new repo`).
- [ ] **Step 2:** Run; expect FAIL.
- [ ] **Step 3:** Implement; update the layout spec's Issues tab notes.
- [ ] **Step 4:** Run suite; expect PASS. Run the app at 60x20, 120x36, 200x50 with `CRONA_ENV=Dev make run-tui` and check both sub-tabs by eye.
- [ ] **Step 5:** Commit `feat(tui): keep the Issues tab's bottom row to globals`.
