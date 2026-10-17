# Today Tab Sections (Layout Redesign Step 2a) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the bordered boxes in the Today tab's views (Daily, Summary, Session) with borderless rule-line sections, and make Daily's arrangement follow width classes (stacked < 100, two columns 100–159, three columns ≥ 160).

**Architecture:** Add `helpers.ClassFor` plus a column splitter to `views/helpers/layout.go`, and `chrome.RenderSection` to `views/chrome/common.go` as the borderless counterpart of `RenderPaneBox`. Migrate the three Today views to them. Other tabs keep `RenderPaneBox` until their own step-2 plans.

**Tech Stack:** Go, lipgloss, `charmbracelet/x/ansi`.

**Spec:** `docs/superpowers/specs/2026-10-17-tui-layout-design.md` (Shell, Responsive layout, Migration step 2). Builds on `docs/superpowers/plans/2026-10-17-tui-shell-swap.md`.

## Global Constraints

- Width classes: Narrow `< 100`, Medium `100–159`, Wide `≥ 160`.
- Section header: `─ TITLE ─────── meta`, exactly one row, `width` cells. The title is `StylePaneTitle` when focused and `StyleDim` bold when not. The rules and meta are `StyleDim`.
- Section body: 1-column left indent, content width `width-2`. The body is clipped to `height-1` rows; when clipped, the last row is a dim `↓ more`. The output is exactly `height` rows of exactly `width` cells.
- Column gutter: 3 cells rendered as dim `" │ "`.
- No `RoundedBorder`, `RenderPaneBox` or `PaneBase.Render` remain in `views/daily`, `views/summary` or `views/sessions/active.go`.
- Placement: `views/helpers` and `views/chrome`, not `layout/` (`layout` imports `views`, so views cannot import `layout`).
- Launch the app for checks with `CRONA_ENV=Dev make run-tui` only. Run Go from `tui/` with `GOCACHE=/tmp/crona-go-cache`. Format with `gofmt` on touched files (not `make fmt`). Commit on `brainstorm/tui-layout`.

## Review Focus

1. Daily at exactly 99, 100, 159 and 160 columns picks the right class: Task 3 `TestDailyArrangementByWidthClass`.
2. A focused section must remain identifiable without a cyan border (title style differs from unfocused): Task 2 `TestRenderSectionFocusedTitleStyle`.
3. Lists longer than the section still show `↑/↓ N more` and the cursor row stays visible at the bottom of a short section: Task 3 `TestDailyIssuesCursorVisibleInShortSection`.
4. Session view timer colour (segment colour) previously came from the border; it must still show: Task 5 `TestActiveSessionHeaderUsesSegmentColour`.
5. Every Today view at 60×18 content and 200×47 content renders exactly `Height` lines, each ≤ `Width` cells: Task 3/4/5 fit tests.

---

### Task 1: Width classes and columns

**Files:**
- Modify: `tui/internal/tui/views/helpers/layout.go`
- Test: `tui/internal/tui/views/helpers/layout_test.go` (create if absent)

**Interfaces:**
- Produces:
  - `type WidthClass int` with `WidthNarrow`, `WidthMedium`, `WidthWide`.
  - `func ClassFor(width int) WidthClass`.
  - `const ColumnGutter = 3`.
  - `func Columns(total int, weights ...int) []int`: splits `total - ColumnGutter*(len(weights)-1)` proportionally to `weights`, gives the remainder to the first column, and never returns a negative width. The sum of results plus gutters equals `total`.

- [ ] **Step 1: Failing tests.** `TestClassForBoundaries` (99→Narrow, 100→Medium, 159→Medium, 160→Wide). `TestColumnsSumToTotal` (`Columns(200, 3, 2, 2)` sums to 194 and `[0] ≥ [1]`; `Columns(100, 3, 2)` → `[59, 38]`; `Columns(5, 1, 1)` has no negatives).
- [ ] **Step 2:** `go test ./internal/tui/views/helpers/` → FAIL (undefined).
- [ ] **Step 3: Implement.**
- [ ] **Step 4:** Same command → ok.
- [ ] **Step 5: Commit** `feat(tui): add width classes and column splitter`.

### Task 2: Borderless section renderer

**Files:**
- Modify: `tui/internal/tui/views/chrome/common.go` (next to `RenderPaneBox`)
- Test: `tui/internal/tui/views/chrome/section_test.go`

**Interfaces:**
- Consumes: Task 1 `ColumnGutter`.
- Produces:
  - `type Section struct { Title, Meta string; Focused bool; Width, Height int; Lines []string }`.
  - `func RenderSection(theme Theme, s Section) string`: the header and body per Global Constraints. When `Meta` doesn't fit, it is dropped before the title is truncated.
  - `func JoinColumns(theme Theme, height int, columns ...string) string`: joins rendered sections horizontally with the dim `" │ "` gutter repeated `height` rows.

- [ ] **Step 1: Failing tests.**
  - `TestRenderSectionHeaderFormat`: width 40, title `Tasks`, meta `3/5` → `ansi.Strip` of line 0 starts with `"─ Tasks "`, ends with `" 3/5"`, and has width 40.
  - `TestRenderSectionExactSize`: 10 lines into height 4, width 30 → 4 rows, each 30 cells, last row contains `↓ more`.
  - `TestRenderSectionFocusedTitleStyle`: the focused and unfocused renders of the same section differ in line 0 (with a theme where `StylePaneTitle` sets a foreground; build it from `lipgloss.NewStyle().Foreground(lipgloss.Color("14"))`).
  - `TestJoinColumnsGutter`: two 10-wide, 3-high sections → 3 rows of 23 cells, each containing `│`.
- [ ] **Step 2:** `go test ./internal/tui/views/chrome/` → FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4:** → ok.
- [ ] **Step 5: Commit** `feat(tui): add borderless section renderer`.

### Task 3: Daily view

**Files:**
- Modify: `tui/internal/tui/views/daily/layout.go` (`renderView`), `daily/summary.go:22-31` (`renderSummary`), `daily/panes.go` (`renderIssues`, `renderHabits`: build title/meta separately and return `chrome.RenderSection`)
- Test: `tui/internal/tui/views/daily/layout_test.go` (new); update `panes_test.go` and `summary_test.go` where they assume the box insets

**Interfaces:**
- Consumes: Tasks 1–2.
- Section titles and meta:
  - Summary: title `Day`, meta the dashboard date.
  - Issues: title from `dailyTaskTitle` (plain text, no `[←/→]` suffix), meta `"<n>  [←/→]"`.
  - Habits: title `Habits Due`, meta `"<done>/<total>"`.
  - Scope labels (`contextmeta.DefaultScopeLabel`) move out of every section; the tab line already shows repo › stream.
- Arrangement (`renderView`):
  - Narrow: Day (height from the existing `SplitVertical(h, 10, 8, h*2/5)`), then Issues and Habits stacked (split the rest 3:2).
  - Medium: Day on top (same split), then `JoinColumns(Issues, Habits)` with `Columns(w, 3, 2)`.
  - Wide: `JoinColumns(Issues, Habits, Day)` with `Columns(w, 3, 2, 2)`, full height.
- Body width: each builder that used `width-6` (padding + border) now uses `width-2`.

- [ ] **Step 1: Failing tests** in `layout_test.go`, using a `dailyState(width, height)` helper with three issues and two habits:
  - `TestDailyHasNoBorders`: output has no `╭` or `╰` at 80, 120 and 200.
  - `TestDailyArrangementByWidthClass`: at 99 the `Habits Due` header row index > the `Tasks`/`Planned` header row index (stacked). At 100 and 159 they are on the same row. At 160 the `Day` header is on row 0 alongside both.
  - `TestDailyIssuesCursorVisibleInShortSection`: 30 issues, cursor 29, height 18, width 80 → the output contains issue 29's title and `↑`.
  - `TestDailyFitsFrame`: 60×18 and 200×47 → exactly `height` lines, each ≤ `width` cells.
- [ ] **Step 2:** `go test ./internal/tui/views/daily/` → FAIL.
- [ ] **Step 3: Implement** `renderSummary`, `renderIssues`, `renderHabits`, `renderView`.
- [ ] **Step 4:** Run the package, then the whole module: `go test ./...` → ok. Update the existing assertions that depended on box padding, keeping their intent.
- [ ] **Step 5: Commit** `feat(tui): render Daily as borderless responsive sections`.

### Task 4: Summary view

**Files:**
- Modify: `tui/internal/tui/views/summary/render.go:25-57`
- Test: `tui/internal/tui/views/summary/render_test.go`

- One section: title `Summary`, meta `state.SummaryDate`, body = the existing lines at `innerWidth = width-2`, scrolled by the existing `scrollSummaryLines` with height `Height-1`.
- [ ] **Step 1: Failing tests:** `TestSummaryHasNoBorders`, plus `TestSummaryFitsFrame` at 60×18 and 200×47.
- [ ] **Step 2:** → FAIL. **Step 3:** implement. **Step 4:** package + module → ok.
- [ ] **Step 5: Commit** `feat(tui): render Summary as a borderless section`.

### Task 5: Active session view

**Files:**
- Modify: `tui/internal/tui/views/sessions/active.go:18-173`
- Test: `tui/internal/tui/views/sessions/active_test.go`

- Two sections: `Session` (meta: segment label from `sessionActionSegmentLabel`) over `Issue`. Narrow and Medium stack them; Wide puts them side by side with `Columns(w, 3, 2)`.
- The timer section's title uses `activeTimerColor` instead of `StylePaneTitle` (it replaces the coloured border).
- [ ] **Step 1: Failing tests:** `TestActiveSessionHasNoBorders`, `TestActiveSessionHeaderUsesSegmentColour` (focus segment vs break segment → line 0 differs), `TestActiveSessionFitsFrame` (60×18, 200×47).
- [ ] **Step 2:** → FAIL. **Step 3:** implement. **Step 4:** package + module → ok.
- [ ] **Step 5: Commit** `feat(tui): render active session as borderless sections`.

### Task 6: Live check

- [ ] Run `CRONA_ENV=Dev make run-tui` in a headless tmux pane at 80×24, 130×36 and 200×50. Capture Daily, Summary, and Session (start a timer: the dev environment is safe for that), then stop the timer.
- [ ] Record anything cramped in the ledger as a deferred minor for the next step-2 plan. No commit unless a fix is needed (TDD applies to any fix).
