# TUI Unified Date Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the TUI's five per-view dates with one shared date anchored on the daemon's logical today. Every dated view navigates it with the same keys.

**Architecture:**
- The model keeps one `date` plus per-view window lengths.
- The input, message and dialog states carry `Date`/`RollupWindowDays` instead of per-view dates.
- `ContentState` keeps its per-view date fields, filled from the one date.
- One handler pair shifts or resets the date and reloads all date-bound surfaces.

**Tech Stack:** Go, Bubble Tea, lipgloss.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-unified-date-design.md`

## Global Constraints

- Test command: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`. Run `gofmt` on touched files only; never `make fmt`.
- "Today" = `currentLogicalDate()` (daemon `health.CurrentDate`, else `time.Now()`); no other `time.Now()` date fallback in the model's date accessors.
- Window presets:
  - Momentum: 7/14/30/90, default 30.
  - Wellbeing: 7/14/21/30, default 7.
  - Rollup: 7/14/30/90, default 7.
- Rollup range = `[date − (window − 1), date]`.

## Review Focus

1. **Midnight rollover.** If the shared date is on the old today, it moves to the new today. Otherwise it stays.
2. **One move reloads every surface.** Shifting the date on one view must reload every surface: Daily, Summary, Momentum, Wellbeing and Rollup show the new date with no stale data.
3. **Dialogs default to the shared date.** The check-in dialog, the issue todo-date default, habit status and momentum definition edits pass the shared date, not a stale per-view one.
4. **Rollup window edges.** The window clamps at its presets, and `0` resets to 7.
5. **Unset date.** Before health loads, `""` resolves to `time.Now()` without panicking, and every surface still loads.

---

### Task 1: One date in the model and state plumbing

**Files:**
- Modify:
  - `tui/internal/tui/model/core.go`, `navigation.go`, `input.go`, `update.go`, `overlays.go`, `dialogs.go`, `selection_actions.go`, `layout_state.go`, `support.go`
  - `tui/internal/tui/input/input.go`, `handlers_dates.go`, `handlers_settings.go`
  - `tui/internal/tui/dispatch/messages.go`
  - `tui/internal/tui/dialog_runtime/actions.go`
- Test: `tui/internal/tui/model/tabs_test.go`

**Interfaces:**
- Produces:
  - `Model.date string`, `Model.rollupWindowDays int`
  - `func (m Model) currentDate() string`
  - `func (m Model) currentRollupWindowDays() int` (default 7)
  - `input.State.Date`, `input.State.RollupWindowDays`
  - `dispatch.MessageState.Date`, `dispatch.MessageState.RollupWindowDays`
  - `dialog_runtime` state `Date`
  - `input.Deps.CurrentDate func(State) string` replaces the five `Current*Date` deps.
- Removes: `dashboardDate`, `summaryDate`, `momentumDate`, `wellbeingDate`, `rollupStartDate`, `rollupEndDate` and their accessors.

- [ ] Step 1: Write `TestContentStateSharesOneDate` in `model/tabs_test.go`. Set `m.date = "2026-05-01"` and build content state for Daily, Summary, Momentum, Wellbeing and Rollup. Assert:
  - `DashboardDate`, `SummaryDate`, `MomentumDate`, `WellbeingDate` and `RollupEndDate` all equal `2026-05-01`.
  - `RollupStartDate == "2026-04-25"` at the default window.
  - It fails to compile, because `m.date` doesn't exist yet.
- [ ] Step 2: Collapse the fields (compiler-guided). Every former per-view accessor call becomes `currentDate()`, and Rollup loads use `shiftISODate(date, -(window-1))`.
- [ ] Step 3: Run the full suite. Fix tests that set per-view dates by switching them to `Date`.
- [ ] Step 4: Commit `refactor(tui): hold one shared date in the model`.

### Task 2: Shared date keys and eager reload

**Files:**
- Modify: `tui/internal/tui/input/handlers_dates.go`, `input/input.go`, `key_registry/rollup.go`, `model/input.go`
- Test: `tui/internal/tui/input/summary_dates_test.go` (renamed `dates_test.go`)

**Interfaces:**
- Produces:
  - `handleShiftDate(s State, deps Deps, dir int)`
  - `handleResetDate(s State, deps Deps)`
  - `handleShiftRollupWindow(s State, deps Deps, dir int)`
  - `handleResetRollupWindow(s State, deps Deps)`
  - `Deps.ReloadDateSurfaces func(State) tea.Cmd`
  - `Deps.CurrentLogicalDate func(State) string`

- [ ] Step 1: Tests:
  - `TestShiftDateMovesSharedDateOnEveryDatedView`: for each of Daily, Summary, Momentum, Wellbeing and Rollup, pressing `,` moves `Date` back one day and calls `ReloadDateSurfaces` once.
  - `TestResetDateGoesToLogicalToday`
  - `TestRollupWindowPresets`: 7→14→30→90 clamps, and `0` resets to 7.
- [ ] Step 2: Run them; they fail.
- [ ] Step 3: Implement.
  - Register `,`/`.`/`g` for all five views.
  - Rollup: `+`/`-`/`0` window; remove ←/→, `S`, and the start/end handlers and dialogs. `E` opens the date picker for the shared date.
  - `ReloadDateSurfaces` batches the Daily summary/plan/due habits/streaks, summary snapshot, momentum range, wellbeing and rollup loads.
- [ ] Step 4: Run the suite and commit `feat(tui): navigate one shared date from every dated view`.

### Task 3: Midnight rollover on the shared date

**Files:** Modify `tui/internal/tui/dispatch/messages.go`; Test `dispatch/messages_test.go`

- [ ] Step 1: Tests in `TestLogicalDateRollover`:
  - When `Date` equals the previous today, it becomes the new date.
  - When `Date` is another day, it stays.
  - When `Date` is empty, it is set to the new date.
- [ ] Step 2: Run them; they fail or no longer compile.
- [ ] Step 3: Implement. Make `applyLogicalDate` act on `Date` only and reload via the same surface batch.
- [ ] Step 4: Run the suite and commit `fix(tui): roll the shared date over at midnight`.

### Task 4: Date navigation on every dated view

**Files:** Modify `tui/internal/tui/views/renderer/main.go`, `views/daily/strip.go` (export a generic helper); Test `layout/render_test.go`

**Interfaces:** Produces `viewchrome.DateNavMeta(theme, rawDate string, settings, weekStart, room int) string`, moved from `daily.SubTabMeta`. Rollup adds a `start → end` range prefix when room allows.

- [ ] Step 1: Test `TestRenderSubTabLineCarriesDateOnDatedViews`: for Summary, Momentum, Wellbeing and Rollup, the sub-tab line ends with `next [.]` and contains the shared date. Rollup's line also contains its start date.
- [ ] Step 2: Run it; it fails.
- [ ] Step 3: Implement.
- [ ] Step 4: Run the suite, update the layout spec's key notes, and commit `feat(tui): show the shared date on every dated view`.
