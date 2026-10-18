# Insights Review Sub-tab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Insights tab's Rollup and Wellbeing sub-tabs with one Review sub-tab: one window, a day list, Trends/Time toggled by `[S]`, and the selected day's detail.

**Architecture:** A new `views/review` package renders the strip, Days list, Trends, Time and Detail as borderless `viewchrome` sections, porting the chart, breakdown, check-in, risk and driver renderers from `views/rollup` and `views/wellbeing`. The model gets one window (`reviewWindowDays` plus a window-end anchor) and one session bool (`reviewShowTime`); input binds Review's keys; the old views, panes, handlers and reload paths are deleted.

**Tech Stack:** Go, Bubble Tea, lipgloss, x/ansi. Module `tui/`; package paths below are relative to `tui/internal/tui/`.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-insights-review-design.md`

## Global Constraints

- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`. Never run `make fmt`; `gofmt -w` only files you touched.
- UI-only; no daemon changes.
- Every hint shown must be a key the input layer honours in that state; hints are whole (`viewchrome.FitHints`, `viewchrome.KeyHint`).
- Width classes: `viewhelpers.ClassFor` — wide ≥160, medium ≥100, narrow <100.
- Output of the view is exactly `state.Height` rows of `state.Width` cells.
- Bottom row: globals only on ViewReview (paneActions nil in `model/view.go`, like Daily).
- Run the app with `CRONA_ENV=Dev make run-tui`.

## Rulings made while planning (not in the spec)

- **Window anchor.** The spec has the window end on the shared date and the Days cursor set the date; together every cursor step would slide the window. So the window ends on `reviewWindowEnd`, set to the shared date when the date changes from the date row, calendar or `g`, and on entering Review; `↑/↓` change the shared date only within `[start, reviewWindowEnd]` and do not move the window. Cost if wrong: one field and the `↑/↓` handler.
- **Presets.** One window: presets `{7, 14, 30, 90}`, default 7 (Rollup's). `commands.LoadWellbeingWindow`'s 30-day cap is raised to 90.
- **No Days cursor state.** The selected row is the shared date's index in the window; there is no `cursor[PaneReviewDays]`.

## Review Focus

1. Window of 90 days: the Days list scrolls to keep the selected row visible; charts squeeze 90 points into the slot without panicking (width < points).
2. Days with no data (no check-in, no rollup row, metrics missing): rows render with blanks, charts skip gaps, strip shows `–`, never a zero that reads as a real score.
3. Date moved outside the window by `,`/`.`/calendar: the window re-anchors and reloads; the selected row is the new date (last row).
4. Late messages: a rollup or metrics message for a stale range (window changed mid-load) must not overwrite newer data — keep the existing range check in `dispatch/messages.go` and key it to the Review range.
5. `[S]` on narrow opens a dialog; resizing from narrow to wide while that dialog is open leaves a usable screen (dialog closes or stays; no stuck mode).

Tests pinning these are listed in Tasks 1, 2 and 3.

---

### Task 1: `views/review` rendering

**Files:**
- Create: `views/review/render.go` (`Render`), `views/review/strip.go`, `views/review/days.go`, `views/review/trends.go` (ported from `views/wellbeing/trends.go` + `views/rollup/focus_graph.go`), `views/review/time.go` (ported from `views/rollup/summary.go`, `distribution.go`, `layout.go` breakdown lines), `views/review/detail.go` (ported from `views/wellbeing/burnout.go`, `risk_snapshot.go`, `accountability.go` check-in lines)
- Modify: `views/types/main.go` (add fields)
- Test: `views/review/review_test.go`, `views/review/fixtures_test.go`

**Interfaces:**
- Produces:
  - `func Render(theme types.Theme, state types.ContentState) string`
  - `func StripRows(theme types.Theme, state types.ContentState, width int) []string` — exactly 2 rows, each exactly `width` cells.
  - `func DetailLines(theme types.Theme, state types.ContentState, width int) []string` — day summary + check-in + risk + drivers (Task 3's narrow dialog uses it).
  - `func TimeLines(theme types.Theme, state types.ContentState, width int) []string` — the breakdown (Task 3's narrow dialog uses it).
  - `func WindowDates(state types.ContentState) []string` — ISO dates `start..end`, oldest first.
  - New `ContentState` fields: `ReviewWindowDays int`, `ReviewWindowEnd string`, `ReviewShowTime bool`.

- [ ] **Step 1: Write failing tests** in `views/review/review_test.go`, fixture with a 7-day window ending `2026-10-14`, shared date `2026-10-14`, rollup rows, metrics and a check-in:
  - `TestStripTwoRowsWidthExact`: for widths 60, 100, 160 `StripRows` returns 2 rows, `ansi.StringWidth(row) == width`; row 0 contains `Worked`, row 1 contains `Mood` at 160; at 60 the last column group of each row is dropped.
  - `TestStripBlankWithoutData`: empty metrics → row 1 shows `–` for Mood, not `0`.
  - `TestPanesByWidth`: 200x50 output contains section titles `Days`, `Trends`, and the date title `Tue 14 Oct`; 120x36 contains `Days` and `Tue 14 Oct`, not `Trends`; 80x30 contains `Days` only (no `Trends`, no date title).
  - `TestShowTimeSwapsSlot`: 200x50 with `ReviewShowTime` → `Time` title present, `Trends` absent, meta `[S] trends`; without → meta `[S] time`. 120x36 with `ReviewShowTime` → `Time` present, date title absent, meta `[S] day`.
  - `TestDaysRowsAndSelection`: 7 rows oldest first; the `▶` row is `Tue 14`; a day without check-in has no mood mark.
  - `TestDaysScrollsLongWindow`: 90-day window at height 20 with the date 40 days back → the selected row is visible.
  - `TestShortHeightFoldsTrends`: 200x22 → no `Trends` title (falls back to medium layout); 200x28 → `Trends` present with fewer charts than at 200x50.
  - `TestChartsSqueezeManyPoints`: 90 days into a 30-cell slot renders without panic.
  - `TestDetailTitleMetaHints`: wide detail meta contains `[w] check-in` and `[W] away`.
  - `TestHistoricalAwayDate`: `HistoricalAwayDate` set → output equals `viewaway.RenderHistorical(theme, state)`.
  - `TestRenderExactSize`: 60x20, 120x36, 200x50 → `state.Height` lines, each `state.Width` cells.

- [ ] **Step 2: Run** `cd tui && GOCACHE=/tmp/crona-go-cache go test ./internal/tui/views/review/` — expected FAIL (undefined `Render`, `StripRows`).

- [ ] **Step 3: Implement.**
  - Strip row 0 columns: `Worked <h>` (from `DashboardWindow`), `Focus <score> <bar>` (`WeeklyFocusScore`/rollup focus), `Plan <pct>%`, `done n  failed n`; row 1: `Mood`, `Energy`, `Sleep` (averages over `MetricsRange`), `Check-ins n/days`, `Risk <level>` (`latestBurnout`). Drop groups right-first; pad to width (copy `views/daily/strip.go` `padCells`).
  - Layout: class from `viewhelpers.ClassFor(width)`; height < 24 treats wide as medium. Wide: Days ≈ 30 cols │ middle │ detail ≈ 40 cols via `viewchrome.JoinColumns`. Trends draws focus, mood, energy, recovery top to bottom while each chart keeps ≥ 4 rows; at least one.
  - Days row: `Mon 13  ✓ 4/5  6h ▲` — status mark and counts from the rollup day, worked from metrics, mood mark from mood vs window mean (`▲` above, `▬` within 0.25, `▼` below, blank without check-in). Selected row prefixed `▶`; scroll window keeps it visible.
  - Detail: `viewsummary.DayLines(theme, state, w)` then a blank line, then check-in, risk snapshot, top drivers (ported). Title `Mon 2 Jan` format of the shared date.
  - Port only what these sections use; restyle as borderless lines. Old packages stay until Task 2 deletes them.

- [ ] **Step 4: Run** the package tests — expected PASS. Then the full suite — PASS.

- [ ] **Step 5: Commit**

```bash
git add tui/internal/tui/views/review tui/internal/tui/views/types/main.go
git commit -m "feat(tui): add the Review view with strip, days, trends, time and detail"
```

---

### Task 2: Wire `ViewReview`, one window, and remove Rollup and Wellbeing

**Files:**
- Modify: `state/types.go`, `state/tabs.go`, `views/renderer/main.go`, `model/core.go`, `model/layout_state.go`, `model/navigation.go`, `model/update.go`, `model/input.go`, `model/overlays.go`, `model/view.go`, `model/terminal_title.go`, `model/selection_actions.go`, `input/input.go`, `input/handlers_dates.go`, `input/handlers_core.go`, `input/handlers_settings.go`, `dispatch/messages.go`, `commands/load.go`, `testsuite/support/support.go`
- Delete: `views/rollup/`, `views/wellbeing/`
- Test: `input/review_test.go`, `model/review_test.go`; update `state/tabs_test.go`, `input/tabs_test.go`, `model/tabs_test.go`, `input/dates_test.go`, `model/navigation_test.go`, `model/bottom_row_test.go`, `layout/render_test.go`, `testsuite/dialog_support_test.go`

**Interfaces:**
- Consumes: Task 1's `review.Render`, `ContentState.ReviewWindowDays/ReviewWindowEnd/ReviewShowTime`.
- Produces:
  - `uistate.ViewReview View = "review"` (label `Review`), `uistate.PaneReviewDays Pane = "review_days"`; `TabInsights: {ViewReview, ViewMomentum}`.
  - Model fields `reviewWindowDays int`, `reviewWindowEnd string`, `reviewShowTime bool`; input `State` fields `ReviewWindowDays`, `ReviewWindowEnd`, `ReviewShowTime`.
  - `func (m Model) reviewRange() (start, end string)`; `func (m Model) reloadReview() tea.Cmd` — batch of `commands.LoadRollupSummaries(c, start, end)` and `commands.LoadWellbeingWindow(c, end, days)`.
  - `handleShiftReviewWindow(s State, deps Deps, dir int)`, `handleResetReviewWindow(s State, deps Deps)`, `handleReviewStepDay(s State, deps Deps, dir int)`, `handleToggleReviewTime(s State, deps Deps)`.

- [ ] **Step 1: Write failing tests.**
  - `input/review_test.go`:
    - `TestReviewWindowPresets`: `+` from 7 → 14 → 30 → 90 → 90; `-` back; `0` → 7.
    - `TestReviewStepDayStaysInWindow`: window end `2026-10-14`, 7 days, date `2026-10-08`; `up` keeps `2026-10-08` (first row) ; `down` → `2026-10-09`; window end unchanged.
    - `TestReviewShiftDateReanchors`: `.` / `,` / `g` set `ReviewWindowEnd` to the new date.
    - `TestReviewToggleTime`: `S` flips `ReviewShowTime` (wide).
    - `TestReviewCheckInKeys`: `w` opens the check-in, `W` toggles away on ViewReview; `E` opens the date dialog (not export).
  - `model/review_test.go`:
    - `TestReviewShowTimeSurvivesTabSwitch`: set `reviewShowTime`, switch to Today, back to Review → still true.
    - `TestEnteringReviewAnchorsWindow`: entering Review sets `reviewWindowEnd` to the shared date.
    - `TestReviewReloadUsesOneRange`: after `+`, the rollup start date and wellbeing window both equal `reviewRange()`'s start.
    - `TestStaleReviewMessageIgnored`: a rollup message for an older range does not replace current data.
    - `TestReviewBottomRowGlobalsOnly`: bottom row on ViewReview has no pane actions.
  - Update tab/jump/help tests: Insights is `Review · Momentum`; `rollup`/`wellbeing` absent.
- [ ] **Step 2: Run** the full suite — expected FAIL (undefined `ViewReview`, …).
- [ ] **Step 3: Implement.**
  - Replace `ViewRollup`/`ViewWellbeing` and their panes with `ViewReview`/`PaneReviewDays` everywhere (`grep -rln 'ViewRollup\|ViewWellbeing\|PaneRollup\|PaneWellbeing' tui/internal/tui`); delete pane line counts and cursors for the old panes in `model/core.go`, `model/layout_state.go`, `model/update.go`.
  - `rollupWindowDays`/`wellbeingWindowDays` → `reviewWindowDays` (default 7, presets `{7, 14, 30, 90}` in `input/handlers_dates.go`); `currentRollupStartDate` → `reviewRange`.
  - `reloadDateSurfaces` calls `reloadReview` instead of the two Insights loaders; `RollupRefreshRequestedMsg` handling uses `reviewRange`. `LoadWellbeingWindow` cap 30 → 90.
  - Date row/calendar/`g` and entering Review set `reviewWindowEnd = currentDate()`; `↑/↓` on ViewReview call `handleReviewStepDay` (clamped to the window, `ReloadDateSurfaces` without re-anchoring).
  - Keys: `keyregistry.RegisterRollup` on ViewReview with the review window handlers; `E` = date dialog; `w`, `W` as Wellbeing's; `S` = `handleToggleReviewTime`. Remove `ViewWellbeing` from the `E` export list.
  - `views/renderer/main.go`: `"review"` → `review.Render`; delete `views/rollup`, `views/wellbeing`, and their testsuite support wrappers.
- [ ] **Step 4: Run** the full suite and `go vet ./...` — PASS.
- [ ] **Step 5: Commit**

```bash
git commit -am "feat(tui): replace Rollup and Wellbeing with the Review sub-tab"
```

---

### Task 3: Narrow dialogs, help, docs, visual check

**Files:**
- Modify: `input/input.go`, `model/dialogs.go` (or the overlay that hosts Daily's narrow day summary), `views/review/render.go`, help sheet source, `docs/superpowers/specs/2026-10-17-tui-layout-design.md`
- Test: `model/review_test.go`, `views/review/review_test.go`

**Interfaces:**
- Consumes: `review.DetailLines`, `review.TimeLines`, `handleToggleReviewTime`.
- Produces: narrow `[enter]` → day dialog, narrow `[S]` → Time dialog.

- [ ] **Step 1: Write failing tests.**
  - `TestNarrowEnterOpensDayDialog`: width 80, ViewReview, `enter` → a dialog whose body contains `DetailLines` output; at width 200 `enter` is not handled.
  - `TestNarrowSOpensTimeDialog`: width 80, `S` → Time dialog; `ReviewShowTime` unchanged.
  - `TestNarrowDaysHints`: width 80 render shows `[enter] day  [S] time` in the Days title meta; width 200 shows neither `[enter]` hint.
  - `TestResizeWithDialogOpen`: open the Time dialog at 80, resize to 200 → render is valid size and `esc` closes the dialog.
  - `TestHelpListsReviewKeys`: `?` sheet on ViewReview lists `S`, `w`, `W`, `+/-/0`.
- [ ] **Step 2: Run** — expected FAIL.
- [ ] **Step 3: Implement** using the same dialog path Daily's narrow `[S]` day summary uses; the narrow Days title meta is `[enter] day  [S] time` fitted with `FitHints`. Update the layout-design spec's Insights notes to `Review · Momentum`.
- [ ] **Step 4: Run** the full suite — PASS. Visual check with `CRONA_ENV=Dev make run-tui` at 60x20, 120x36, 200x50: strip, `[S]` in each width, `+/-/0`, `↑/↓`, an away date.
- [ ] **Step 5: Commit**

```bash
git commit -am "feat(tui): open Review's day and time views as dialogs on narrow screens"
```
