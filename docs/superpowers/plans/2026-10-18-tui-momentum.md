# TUI Momentum Sub-tab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Momentum's bordered card stack with a Review-style strip plus a `List │ Detail` layout. Inactive momentums are dimmed and each row has a sparkline.

**Architecture:**
- The cards are ordered active-first once, in dispatch, when they load. The view, the cursor and the selection snapshot then all agree on the index without any remapping.
- `views/momentumview` is split into `strip.go`, `list.go`, `detail.go` and `render.go`. The existing chart helpers (`renderMomentumSeries` and below) stay in `chart.go`, renamed from `view.go`.
- The dead `MomentumTab` plumbing goes last, as a pure deletion.

**Tech Stack:** Go, Bubble Tea, lipgloss, x/ansi, `viewchrome` (`RenderSection`, `JoinColumns`, `KeyHint`, `FitHints`, `ListWindow`, `SelectionCursor`).

**Spec:** `docs/superpowers/specs/2026-10-18-tui-momentum-design.md`

## Global Constraints

- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`.
- Never run `make fmt`; run `gofmt -w` on touched files only.
- UI-only. No daemon, API or details-dialog changes.
- Width classes come from `viewhelpers.ClassFor`: narrow shows the list only, medium and wide show `List │ Detail`.
- `listWidth = 40`.
- The sparkline is 8 cells, `▁▂▃▄▅▆▇█`, scaled by `min(count/target, 1)`. A skipped bucket is a dim `·`, and short series are left-padded with spaces.
- Marks: `✓` met, `◐` near, `✗` missed, `·` skipped, `paused` for inactive momentums.
- Strip copy: `Active N  inactive N`, `Latest  met N  near N  missed N`, `Best streak  <name> Nd`, `Longest ever  <name> Nd`, and `–` when a value is missing.
- Empty-state copy: `No momentums yet  [a] new momentum`.
- Commit on `brainstorm/tui-layout`.

## Review Focus

1. A card with an empty `Series` must not panic in the sparkline, the mark or the "Latest" counts. Its row shows a blank sparkline and no mark. Covered by a test in Task 2.
2. A target of 0 must not divide by zero; it scales as full (met). Covered by a test in Task 2.
3. A cursor past the end after a reload (for example after a delete) is still clamped by `ClampFiltered`. Ordering must not bypass that. Covered by a test in Task 1.
4. Very short heights (`Height` < 6) must render without a negative-height panic. Covered by a test in Task 3.
5. An active filter on `momentum_cards` (`FilteredIndices`): the list renders the filtered rows only, and the cursor indexes into them. Covered by a test in Task 3.

---

### Task 1: Order cards active-first on load

**Files:**
- Modify: `dispatch/messages.go` (the `commands.MomentumRangeLoadedMsg` case, around line 212)
- Test: `dispatch/messages_test.go`

**Interfaces:**
- Produces: `state.MomentumCards` is a stable order with every `Definition.Enabled == true` card before the disabled ones.

- [ ] **Step 1: Write the failing tests.**
  - `TestMomentumCardsLoadActiveFirst`: load cards `[off A, on B, off C, on D]`. Expect the names `[B, D, A, C]`.
  - `TestMomentumCardsLoadStillClamps`: set a cursor of 9 and load 2 cards. Expect `ClampFiltered` to be called for `PaneMomentumCards`.
- [ ] **Step 2: Run the tests and confirm they fail.**
  - Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/dispatch -run MomentumCardsLoad`
  - Expected: the ordering test FAILs.
- [ ] **Step 3: Implement the ordering.** Copy the cards and use `slices.SortStableFunc` on `!Enabled`, so the loaded message's slice is not mutated.
- [ ] **Step 4: Run the tests and confirm they pass.** Expected: PASS.
- [ ] **Step 5: Commit.** `fix(tui): list active momentums before inactive ones`

### Task 2: Strip and list rows

**Files:**
- Rename: `views/momentumview/view.go` to `views/momentumview/chart.go`. Delete everything above `renderMomentumSeries` that only the card layout uses (Task 3 removes the rest).
- Create: `views/momentumview/strip.go` and `views/momentumview/list.go`
- Test: `views/momentumview/list_test.go`

**Interfaces:**
- Consumes: `momentumStatusForPoint` and `momentumStatusStyle` from `chart.go`.
- Produces:
  - `StripRows(theme types.Theme, state types.ContentState, width int) []string`. It always returns 2 rows, each exactly `width` cells. Groups drop right-first, as in `review.StripRows`.
  - `sparkline(theme types.Theme, series []sharedtypes.MomentumSeriesPoint, enabled bool) string`. Always 8 cells.
  - `statusMark(theme types.Theme, card sharedtypes.MomentumCard) string`. It returns `""` for an empty series.
  - `listRow(theme types.Theme, card sharedtypes.MomentumCard, width int) string`. Layout: the name, padded or truncated; then two spaces; the sparkline; two spaces; the streak right-aligned in 4 cells (`momentumhelpers.FormatLength(card.Current, unit)` in compact form, `27d`); a space; the mark. The whole row is `StyleDim` when the momentum is disabled.
  - `renderList(theme types.Theme, state types.ContentState, width, height int) string`

- [ ] **Step 1: Write the failing tests.**
  - `TestSparklineScalesByTarget`: counts `0,1,2,4` against target 4 give `"    ▁▃▅█"` (stripped, left-padded to 8; level = `round(ratio*7)`).
  - `TestSparklineSkippedAndZeroTarget`: a skipped point gives `·`, and target 0 gives `█`.
  - `TestSparklineEmptySeries`: 8 spaces.
  - `TestStatusMarks`: met `✓`, near (3 of 4) `◐`, missed `✗`, skipped `·`, a disabled momentum gives `paused`, and an empty series gives `""`.
  - `TestListRowFitsWidth`: a long name at width 38. Expect `ansi.StringWidth == 38`, containing `…`, the sparkline, `27d` and `✓`.
  - `TestStripRowsCounts`: 3 active momentums (latest statuses met, near, missed) and 1 inactive. Row 1 contains `Active 3  inactive 1` and `met 1  near 1  missed 1`. Row 2 contains `Best streak` and `Longest ever` with the right names. Both rows are exactly 120 cells.
  - `TestStripRowsDropRightFirst`: at width 30, both rows are 30 cells wide and contain no `Latest`.
  - `TestStripRowsEmpty`: no cards gives `Active 0` and `–` in row 2.
- [ ] **Step 2: Run the tests and confirm they fail.**
  - Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/views/momentumview -run 'Sparkline|StatusMarks|ListRow|StripRows'`
  - Expected: a build failure (undefined functions).
- [ ] **Step 3: Implement `strip.go` and the row helpers in `list.go`.**
  - The sparkline level is `round(ratio*7)` into `▁…█`, coloured with `momentumStatusStyle(momentumStatusForPoint(p, enabled))`.
  - `Best streak` is the active card with the largest `Current`. `Longest ever` is the card with the largest `Longest`.
- [ ] **Step 4: Run the tests and confirm they pass.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): add Momentum strip and sparkline list rows`

### Task 3: List pane, detail pane and Render

**Files:**
- Modify: `views/momentumview/list.go` (`renderList`)
- Create: `views/momentumview/detail.go` and `views/momentumview/render.go`
- Modify: `views/momentumview/chart.go` (delete the card modes, `renderCard*`, `momentumCardWindow`, `visibleMomentumCardWindow`, `momentumHeader`, `renderMomentumTopSection` and `renderMomentumEmptyState`)
- Rewrite: `views/momentumview/view_test.go`. Keep the chart tests (bucket timeline, fixed y-axis ticks, squares, distribution, skipped and prorated metadata, shared scale, wrapping by width) by asserting on `Render` at 200x50, where the detail pane shows the chart. Delete the card-mode and overflow-hint tests.

**Interfaces:**
- Consumes: everything Task 2 produces.
- Produces:
  - `Render(theme types.Theme, state types.ContentState) string`. Its signature is unchanged, so `views/renderer/main.go` keeps working.
  - `DetailLines(theme types.Theme, card sharedtypes.MomentumCard, width, height int) []string`

- [ ] **Step 1: Write the failing tests.**
  - `TestRenderNarrowListOnly`: at 60x20, no `Cadence`, `[enter]` appears in the list meta, and every line is ≤ 60 cells.
  - `TestRenderMediumAndWideShowDetail`: at 120x36 and 200x50, the output contains `Cadence`, `Target`, `Streak` and the selected name as the detail title.
  - `TestRenderInactiveDimmedAndPaused`: a disabled card's row contains `paused`, and the detail meta contains `[x] enable`. An enabled card's meta contains `[x] disable`.
  - `TestRenderListCreateLineAndEmpty`: the list's first line contains `[a] new momentum`. No cards gives `No momentums yet  [a] new momentum`.
  - `TestRenderNoBordersOrHeader`: no `╭` or `│ Momentum ` header, and no `Momentum  2026-` header line.
  - `TestRenderTinyHeight`: Height 4 and Height 1 don't panic, and the output has ≤ `Height` lines.
  - `TestRenderUsesFilteredRows`: with filter `"deep"` on `momentum_cards`, only matching names appear. Use `selection.FilteredIndices` semantics, so check how `ContentState` exposes filtered cards. If `ContentState` has `Filters`, filter there; otherwise this test asserts on the cursor's card only. Ledger which.
  - `TestDetailLinesWrapDescription`: an 80-character description at width 30 wraps, and every line is ≤ 30 cells.
- [ ] **Step 2: Run the tests and confirm they fail.**
  - Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/views/momentumview`
  - Expected: the new tests FAIL.
- [ ] **Step 3: Implement.**
  - The detail rows use `viewsummary`'s label/value style. Export `summary.Row(theme, label, value, width)` from `views/summary/render.go`, renaming `row`.
  - The chart takes the height left after the rows and before the footnote, at `width-2`.
  - `Render` lays out `StripRows`, then the panes, sized as Review's `Render` does. The list pane is `listWidth`, and the detail pane gets the rest minus 3.
- [ ] **Step 4: Run the tests and confirm they pass.**
  - Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`
  - Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay Momentum out as a strip, list and detail`

### Task 4: Keys, enter and hints

**Files:**
- Modify: `views/chrome/actions.go` (the `state.View == "momentum"` ContextualActions block, around line 202)
- Test: `views/chrome/actions_test.go`, `input/*_test.go` (the existing input test file for momentum)

**Interfaces:**
- Consumes: nothing new.

- [ ] **Step 1: Write the failing tests.**
  - `TestMomentumActionsOmitDateKeys`: the momentum `ContextualActions` (the help sheet's source) contain `[D] delete` and no `[,/.]`, `[g]` or `[+/-/0]`, because the date row owns those.
  - `TestMomentumBottomRowGlobalOnly`: `GlobalActions(View:"momentum")` contains no `[e]`, `[x]` or `[D]`.
  - `TestMomentumEnterOpensDialogAtAnyWidth`: enter on `ViewMomentum` with a card selected calls the existing momentum details dialog opener, at width 60 and at width 200.
- [ ] **Step 2: Run the tests and confirm they fail.**
  - Run: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/views/chrome ./internal/tui/input -run Momentum`
  - Expected: the first test FAILs. If the other two pass already, they are regression pins; note that in the ledger.
- [ ] **Step 3: Implement.** Drop the three date keys from the momentum block.
- [ ] **Step 4: Run the tests and confirm they pass.** Expected: PASS.
- [ ] **Step 5: Commit.** `fix(tui): keep Momentum's date keys on the date row`

### Task 5: Remove the dead MomentumTab plumbing

**Files:**
- Modify:
  - `model/core.go`: the `MomentumTab` type and consts, the `momentumTab` field and initialiser, and the `listLen` branches.
  - `model/navigation.go:614-622`
  - `model/input.go:103,143-146`
  - `model/update.go:272,370`
  - `model/layout_state.go:45`
  - `model/view.go:44,196`
  - `model/dialogs.go:656`
  - `model/selection_actions.go` (5 guards become `m.view == ViewMomentum`)
  - `input/input.go:32`
  - `input/handlers_session.go:39` (becomes `return true`)
  - `dispatch/messages.go:45,287-295` (drop the `focus` branch and the `MomentumHistoryY` clamp if it is unused)
  - `views/types/main.go:57`
  - `views/chrome` `ActionsState.MomentumTab`
- Test: the existing suite.

- [ ] **Step 1: Delete the code.** `grep -rn 'MomentumTab\|momentumTab' tui/` must return nothing. If `momentumHistoryCursor` or `MomentumHistoryY` become unused, delete them too.
- [ ] **Step 2: Build and test.**
  - Run: `cd tui && go vet ./... && GOCACHE=/tmp/crona-go-cache go test ./...`
  - Expected: PASS. Fix tests that set `MomentumTab: "custom"` by dropping the field.
- [ ] **Step 3: Commit.** `refactor(tui): drop Momentum's unused focus and wellbeing tabs`

### Task 6: Visual check

- [ ] **Step 1: Capture each size.** For each of 60x20, 120x36 and 200x50, start `tmux new-session -d -s crmom -x W -y H "CRONA_ENV=Dev make run-tui"`, navigate to Insights › Momentum, run `capture-pane -p`, then `kill-session`.
  - Expected: the strip has 2 rows and nothing wraps; inactive rows are dimmed and last; the detail chart fills the pane; narrow shows the list only.
- [ ] **Step 2: Fix what the captures show.** Each fix gets a failing test first. Commit as `fix(tui): …`.
