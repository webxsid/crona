# Fold Summary into Daily Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the Summary sub-tab by moving its unique content into Daily: strip headline column, a Day summary detail mode (dialog when narrow) and the scored calendar on `C`.

**Architecture:** `views/summary` exports `DayLines`, the Day summary sections at a width. Daily's strip and detail pane read the summary snapshot already in `ContentState`. The model holds a `dailyDetailDay` flag, reset when the cursor or pane moves.

**Tech Stack:** Go, Bubble Tea, lipgloss.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-daily-summary-design.md`

## Global Constraints

- Test command: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`. gofmt touched files only; never `make fmt`.
- Keys on Daily: `S` Day summary, `C` calendar. `c` stays context checkout.
- The snapshot counts only when `SummarySnapshot.Date == SummaryDate` (shared date).

## Review Focus

1. Snapshot for another date (stale or loading): strip omits Score/Plan, Day summary says "Loading summary…".
2. `S` toggles back off; cursor/pane moves reset to the selection.
3. Narrow screens: `S` opens a dialog, never toggles an invisible pane.
4. Nothing still routes to the removed Summary view (jump menu, tab digits, `[`/`]` cycling, terminal title).
5. Strip width: the new column never pushes a row past the width.

---

### Task 1: Score / Plan strip column

**Files:** Modify `tui/internal/tui/views/daily/strip.go`; Test `views/daily/strip_test.go`

- [ ] Test `TestDayStripShowsScoreOverPlan`: with a snapshot dated the shown date (focus score 72, accountability 80), row 0 contains `Score 72/100` and row 1 `Plan 80%` in one column; with the snapshot dated another day neither appears.
- [ ] Run; it fails. Implement as a last `dayGridColumns` column. Run the suite. Commit `feat(tui): show the day's score and plan in Daily's strip`.

### Task 2: Day summary pane and dialog

**Files:** Modify `views/summary/render.go`, `views/daily/detail.go`, `views/types/main.go`, `model/*`, `input/input.go`; Tests in `views/daily/detail_test.go`, `model/navigation_test.go`

**Interfaces:** `summary.DayLines(theme, state, width) []string`; `ContentState.DailyDetailDay bool`; `input.State.DailyDetailDay bool`; `input.Deps.OpenDaySummary func(*State) bool` (returns true when it opened a dialog).

- [ ] Tests:
  - `TestDetailShowsDaySummaryWhenToggled`: title `Day`, contains `Focus 72/100` and `Accountability`.
  - `TestDaySummaryKeyTogglesOnWideAndOpensDialogWhenNarrow` (model): wide `S` flips the flag; narrow `S` opens `view_entity` titled `Day`.
  - `TestCursorMoveReturnsDetailToSelection` (model).
- [ ] Run; they fail. Implement. Run the suite. Commit `feat(tui): show the day summary in Daily's detail pane or a dialog`.

### Task 3: Calendar on C, drop the Summary sub-tab

**Files:** Modify `state/tabs.go`, `state/types.go`, `input/input.go`, `key_registry/daily.go`, `dialogs/controller/support.go`, `views/renderer/main.go`, `views/chrome/actions.go`, `views/summary/render.go`, `model/*`; Tests across touched packages.

- [ ] Tests: `TestTodayTabHasNoSummary` (state), `TestDailyCOpensCalendar` (input), jump menu has no Summary.
- [ ] Run; they fail. Implement; remove the full-page Summary render, its keys and hints; update tests that used ViewSummary. Run the suite. Update the layout spec's tab table. Commit `feat(tui): fold Summary into Daily and drop its sub-tab`.
