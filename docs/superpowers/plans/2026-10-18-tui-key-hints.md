# Key Hints Where They Apply Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move key hints to where they act: a date row in the header, numbered tabs, Daily's task/habit/detail keys in their panes, and a slim global bottom row.

**Architecture:**
- `viewchrome.FitHints(hints []string, width int) string` joins whole hints within width.
- The shell (`layout/render.go`) gains a date row, sourced from `viewrenderer.DateRow`, which replaces `SubTabMeta`. `ContentHeight` counts that row.
- Daily renders its own pane hints.
- The status line drops date, window and navigation globals.

**Tech Stack:** Go, Bubble Tea, lipgloss.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-key-hints-design.md`

## Global Constraints

- Test command: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`. gofmt touched files only; never `make fmt`.
- Hint style stays `StyleHeader("[k]") + StyleDim(" label")`. Hints are separated by two spaces and never cut mid-hint.

## Review Focus

1. Row budget: the date row and the new pane lines must not overflow the screen at the minimum size (60x20) or push lists off-screen.
2. The action line shows only on the focused pane's selected card, and goes away when focus moves.
3. Timer running: issue and habit actions that need an idle timer are hidden, as they are today.
4. Historical away date on Daily: no date row keys that do nothing, and no `[S]`.
5. Other views' bottom lines lose only date, window and navigation hints, and nothing else.

---

### Task 1: Shell: numbered tabs, framed sub-tabs, date row

**Files:** `layout/render.go`, `views/renderer/main.go`, `views/calendar/datenav.go`, `views/chrome/hints.go` (new). Test `layout/render_test.go`, `views/chrome/hints_test.go`.

- [ ] Tests:
  - `TestRenderTabLineNumbersTabs`: `[1] Today`.
  - `TestRenderSubTabLineFramedByKeys`: `[[]` / `[]]`, and no date on the line.
  - `TestRenderDateRowOnDatedViews`: line 2 carries the date and `next [.]`, plus `[C] calendar` on Daily and `[+/-/0] window` on Momentum; there is no date row on Settings.
  - `TestContentHeightCountsDateRow`.
  - `TestFitHintsKeepsWholeHints`.
- [ ] Watch them fail. Implement. Fix the existing layout tests that moved (line indices). Run the suite. Commit `feat(tui): number the tabs and give the date its own header row`.

### Task 2: Daily pane hints

**Files:** `views/daily/panes.go`, `views/daily/detail.go`, `views/types/main.go` (a `TimerIdle bool`, if not already derivable). Test `views/daily/panes_test.go`, `detail_test.go`.

- [ ] Tests:
  - `TestTasksPaneOffersNewTaskLine`.
  - `TestSelectedTaskCardShowsActions`: focused card shows `[s] status`, not its meta; an unfocused pane shows the meta.
  - `TestTaskActionsHideTimerKeysWhileRunning`.
  - `TestHabitsPaneOffersNewHabitAndActionLine`.
  - `TestDetailTitleOffersDayKey`.
- [ ] Watch them fail. Implement. Run the suite. Commit `feat(tui): show Daily's task and habit keys in their panes`.

### Task 3: Slim bottom row and the `?` sheet

**Files:** `views/chrome/actions.go`, `layout/render.go`, `model/dialogs.go` (help sheet), `views/renderer/main.go` (narrow `[S] day` on the date row). Tests `views/chrome/actions_test.go`, `layout/render_test.go`, `model` help test.

- [ ] Tests:
  - `TestDailyBottomRowIsGlobalsOnly`: no `[s]`/`[a]`/`[,/.]`/`[1-5]`, but includes `[c] context` and `[?] all keys`.
  - `TestRollupBottomRowDropsDateKeys`.
  - `TestHelpSheetListsNavigationKeys`: `[1-5]` and `[[/]]` appear in the help dialog.
  - `TestNarrowDailyDateRowOffersDayKey`.
- [ ] Watch them fail. Implement. Run the suite. Update the layout spec's status-line notes. Commit `feat(tui): slim the bottom row to global keys`.
