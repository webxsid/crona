# Today List │ Detail (Layout Redesign Step 2b) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Daily's three peer columns ("kanban") with a htop/lazygit-style layout: a 1–2 row day strip, Tasks as the one growing list, Habits sized to content, and a detail pane for the selection on wide screens. Move section key hints into the status line.

**Architecture:** New `daily/strip.go` (day strip), `daily/detail.go` (detail pane with calendar), rewrite of `daily/layout.go`. The old Day column (`renderSummary`, `summaryBodyLines`, `buildSummaryLines`, the momentum block) is deleted; its signal and streak values move to the strip. The shell status line puts the focused pane's actions first and drops hints the shell already shows.

**Tech Stack:** Go, lipgloss, `charmbracelet/x/ansi`.

**Spec:** `docs/superpowers/specs/2026-10-17-tui-layout-design.md` (Today tab, Responsive layout rules). Builds on `docs/superpowers/plans/2026-10-17-tui-today-sections.md`.

## Global Constraints

- One section absorbs spare height (Tasks). Every other section is sized to its content.
- Columns only for list │ detail: Wide (≥ 160) uses `Columns(w, 3, 2)`; Medium and Narrow have no detail pane.
- Habits height: `header + rows + markers`, capped at `max(3, rest/3)` (`rest/2` when Habits is focused); Tasks keeps at least 5 rows.
- No action-hint rows inside Daily sections.
- Status line left: focused pane actions, then global actions (deduped). `[v] switch views` and `[?] keys` hints are removed (keys stay bound).
- Tab-line context: `All` when no repo; `repo` alone when no stream; never `-`.
- Launch with `CRONA_ENV=Dev make run-tui` only. Go from `tui/` with `GOCACHE=/tmp/crona-go-cache`. `gofmt` touched files only. Commit on `brainstorm/tui-layout`.

## Review Focus

1. Habits focused with many habits: cursor row still visible inside the capped Habits section.
2. Empty states: no tasks, no habits, no signals/streaks (strip collapses to 1 row), no selection (detail pane shows a hint).
3. Wide detail pane follows the focused pane (task vs habit) and the cursor through filters.
4. Every Daily size from 60×18 to 285×73 renders exactly `Height` lines ≤ `Width` cells.
5. Status line still shows the timer when one runs, and the toast on the right.

---

### Task 1: Shell hints and context label

**Files:** `tui/internal/tui/layout/render.go` (`renderStatusLine`, `renderTabLine`), `tui/internal/tui/views/chrome/actions.go` (`GlobalActions`), `tui/internal/tui/model/view.go` (`layoutChromeState` repo/stream defaults). Tests: `layout/render_test.go`, `views/chrome/actions_test.go`.

- [ ] **Step 1: Failing tests.** `TestStatusLineShowsPaneActionsFirst` (PaneActions `[s] set status` appears before `[1-5]`); `TestTabLineContextAll` (empty repo/stream → right side `All`, no `-`); `TestTabLineContextRepoOnly`; `TestGlobalActionsOmitShellHints` (no `[v]`, no `[?]`).
- [ ] **Step 2:** run → FAIL. **Step 3:** implement (model passes `""` for missing names; `contextLabel(repo, stream)` in layout). **Step 4:** module → ok.
- [ ] **Step 5: Commit** `feat(tui): put focused-pane keys in the status line`.

### Task 2: Day strip

**Files:** Create `tui/internal/tui/views/daily/strip.go`; test `daily/strip_test.go` (replaces the Day-column tests in `summary_test.go`).

- Produces `func renderDayStrip(theme types.Theme, state types.ContentState, width int) []string` — 1 or 2 rows of exactly `width` cells.
  - Row 1 left: `Tasks <resolved>/<total> <status bar> <worked>/<est>   Habits <done>/<total> <habit bar>`; right: `<display date> · W<nn>` (week omitted when the date pattern has one) and dim `,/. g`. Bars ≤ 12 cells, dropped first when narrow.
  - Row 2 (only when data): `Mood 4/5  Energy 3/5  Sleep 7h30m` (averages as `avg 3.1/5`) and `Focus 3d (best 10d)  Check-ins 6d (best 9d)`.
- [ ] **Step 1: Failing tests:** `TestDayStripShowsTotalsAndDate`, `TestDayStripUsesDailySummaryWork` (5400s → `1h30m`, not rollup 7200), `TestDayStripSignalsTodayAndAverage`, `TestDayStripOmitsSecondRowWithoutData`, `TestDayStripFitsWidth` (60, 120, 285).
- [ ] **Step 2–4:** FAIL → implement → ok. Delete `renderSummary`, `summaryBodyLines`, `buildSummaryLines`, momentum block renderers and their tests; keep shared helpers still used.
- [ ] **Step 5: Commit** `feat(tui): replace Daily Day column with a day strip`.

### Task 3: Detail pane

**Files:** Create `daily/detail.go`; test `daily/detail_test.go`.

- Produces `func renderDetail(theme types.Theme, state types.ContentState, width, height int) string` — section titled `Selected`; body for the focused pane's cursor item: task (title, status, repo › stream, est/worked, due, pinned, description/notes wrapped) or habit (name, status, duration/target). Month calendar (`renderCalendarForMode`) below when ≥ 10 rows remain.
- [ ] **Step 1: Failing tests:** `TestDetailShowsSelectedTask`, `TestDetailFollowsHabitsFocus`, `TestDetailEmptySelection`, `TestDetailShowsCalendarWhenTall`.
- [ ] **Step 2–4.** **Step 5: Commit** `feat(tui): add Daily detail pane`.

### Task 4: Daily arrangement

**Files:** `daily/layout.go`, `daily/panes.go` (drop action rows; table when section width ≥ 80); tests in `daily/layout_test.go`, `panes_test.go`.

- [ ] **Step 1: Failing tests:** replace `TestDailyArrangementByWidthClass` with `TestDailyStripThenTasksThenHabits` (row order at 99/120/160), `TestDailyWideHasDetailPane` (160: `─ Selected` on the Tasks header row; 159: absent), `TestDailyHabitsSizedToContent` (2 habits at 200×60 → Habits header within 5 rows of the bottom of content, Tasks gets the rest), `TestDailyNoActionRowsInSections`; keep fit/cursor tests.
- [ ] **Step 2–4.** **Step 5: Commit** `feat(tui): lay out Daily as strip, list and detail`.

### Task 5: Live check

- [ ] `CRONA_ENV=Dev make run-tui` in headless tmux at 80×24, 130×36, 285×73; capture Daily; ledger anything cramped as deferred minors.
