# TUI System Tab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redraw System's six sub-tabs as borderless sections with a grouped list │ detail
(Settings, Alerts, Config), a day list │ detail (Reports), status │ notes (Updates) and
actions │ diagnostics (Support).

**Architecture:**
- A shared `viewchrome.GroupList` and `viewchrome.HelpLines` give Settings, Alerts and Config
  one list and one detail renderer.
- Each sub-tab owns a help table (description, options and keys per row).
- Support gains a cursor (`SupportCursor`) that is plumbed the way `UpdateDiagnosticsExpanded`
  already is.

**Tech Stack:** Go, Bubble Tea, lipgloss, x/ansi.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-system-design.md`

## Global Constraints

- Paths below are relative to `tui/internal/tui/`.
- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`.
- Run `gofmt` on touched files only; never run `make fmt`.
- Panes use `viewchrome.RenderSection`, `JoinColumns` and `LogPaneWidths`.
- There is no `RenderPaneBox`, title line or in-pane action line in any System view.
- The existing keys, cursors, filters and dialogs keep their meaning.
- Commit on `brainstorm/tui-layout`.

## Review Focus

1. **Filters with groups:** a filtered Settings list hides a group's divider when none of its
   rows match. The cursor still counts only the rows that match.
2. **No ExportAssets yet:** while `ExportAssets` is nil (still loading), Config and Reports
   show a dim `Loading…` and don't panic.
3. **Alert reminder rows:** these are dynamic rows. Their help falls back by kind (check-in
   or daily plan), so the coverage test does not fail when a reminder is added.
4. **Release notes longer than the pane:** they truncate with `↓ more` (RenderSection) and
   never overflow the height.
5. **Support cursor clamping:** the cursor stays within the 5 actions with `j`/`k` at both
   ends.

---

### Task 1: Shared group list and help lines

**Files:**
- Create: `views/chrome/grouplist.go`, `views/chrome/grouplist_test.go`
- Modify: `views/chrome/daylist.go`. DayList and GroupList share one windowing core,
  `dividedList`.

**Interfaces:**
- Produces:
  - `type GroupRow struct{ Group, Line string; Danger bool }`
  - `GroupList(theme Theme, rows []GroupRow, cursor, width, height int) []string`: a dim
    `── <Group> ─…` divider before each new group (red when `Danger`), cursor rows rendered
    as in DayList, and the group's divider kept in view
  - `type RowHelp struct{ Description string; Options []string; Keys []string }`
  - `HelpLines(theme Theme, help RowHelp, current string, width int) []string`, which renders:
    - the wrapped description
    - a blank line
    - for options: one `● opt` or `○ opt` line each, with `●` on the option equal to
      `current` (case-insensitive)
    - for no options: `current`, wrapped whole
    - a blank line
    - the keys joined by two spaces, dim
  - `LabelColumn(labels []string, max int) int`: the longest label, capped at `max`

- [ ] **Step 1: Write the failing tests.**
  - `TestGroupListDividesGroups`: 3 rows in 2 groups give 2 dividers, and the cursor row
    carries `▶`.
  - `TestGroupListKeepsDividerInView`: with height 3 and the cursor on a group's first row,
    the window starts with that group's divider.
  - `TestHelpLinesMarksCurrent`: options `Monday`, `Sunday` with current `sunday` give
    `○ Monday` and `● Sunday`.
  - `TestHelpLinesWrapsValueWithoutOptions`: a 60-char path at width 20 is split over
    lines, each 20 cells or fewer, and joined back it equals the path.
  - `DayList`'s existing tests still pass.
- [ ] **Step 2: Run them.** Expected: a build failure.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): add a shared grouped list and setting help lines`

### Task 2: Settings

**Files:**
- Create: `views/settings/help.go` (`var settingsHelp = map[string]viewchrome.RowHelp`,
  keyed by `settingsmeta.Row.Label`), `views/settings/view_test.go`
- Modify: `views/settings/view.go` (rewrite `renderView`)

**Interfaces:** Consumes Task 1.

- **Layout:**
  - The list section is titled `Settings`, with the count as meta.
  - Rows are `label` padded to `LabelColumn(labels, 24)`, then the value.
  - The order comes from `settingsmeta.FilteredIndices(filter, settings)`, and the cursor
    is `state.Cursors["settings"]`.
  - The detail title is the label. Its lines are `HelpLines(help, row.Value, …)`.
- **Options** come from the current label helpers, so their wording matches the value
  column (for example, `weekStartLabel` gives `Monday` and `Sunday`). Rows with free values
  (Custom Date Format, Date Preview, day boundaries, lists) have no options.
- **Keys:**
  - option rows: `[←/→] change`
  - free rows: `[enter] edit`
  - Danger rows: `[enter] confirm`

- [ ] **Step 1: Write the failing tests.**
  - `TestSettingsEveryRowHasHelp`: every `settingsmeta.Rows(defaultSettings)` label has a
    non-empty description.
  - `TestSettingsCurrentIsAnOption`: for each row with options, the row's value is among
    them (case-insensitive).
  - `TestSettingsLayout`:
    - at 120x36: `── Dates`, a `▶` row, a detail titled with the selected label, and `●`
    - no `╭`, and no `[←/→] change   [enter]` line inside the list
    - at 60x20: no detail
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Update the old Settings tests that
  assert the box.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Settings out as a grouped list and detail`

### Task 3: Alerts

**Files:**
- Create: `views/alerts/help.go` (`alertsHelp map[alertsmeta.RowKey]viewchrome.RowHelp`,
  plus a reminder fallback by `ReminderKind`), `views/alerts/view_test.go`
- Modify: `views/alerts/view.go`

- **Layout:**
  - Backend rows (`Section == "Backend"`) leave the list and become a 2-row
    `viewchrome.StripRows` strip of `label value` pairs, with groups in their row order.
  - The list holds the selectable rows only, but keeps the existing selectable cursor
    indexing (`alertsmeta.FilteredSelectableRows`).
  - The sound credit line moves into the detail of the Sound Preset row.
- **Test rows:** the Test Notification and Test Sound rows have no options; their keys are
  `[enter] run`.

- [ ] **Step 1: Write the failing tests.**
  - `TestAlertsEveryRowHasHelp`: this includes the reminder rows, through the fallback.
  - `TestAlertsBackendInStrip`: `terminal-notifier` (or the fixture backend) shows in the
    first 2 lines, and no `── Backend` divider appears.
  - `TestAlertsLayout`: as in Task 2.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Alerts out as a backend strip, list and detail`

### Task 4: Config

**Files:**
- Create: `views/config/help.go`, `views/config/view_test.go`
- Modify: `views/config/view.go`, `views/config/rows.go` (drop `groupedConfigRows` and
  `visibleRow`)

- **Description:** `configitems.Item.DetailBody` when it is set, otherwise a table entry
  keyed by `Label`.
- **Current value:** `Item.Path` for paths, otherwise `Item.Value`.
- **Options:** the style rows' presets (the values the `c/space` cycle walks). Template rows
  have none.
- **Keys:** built from `Editable`, `Mutable` and `Resettable`, as
  `[e] edit or open`, `[c/space] change` and `[r] reset`. `[R] rescan tools` goes in the list
  title meta.
- **Loading:** a nil `ExportAssets` gives a dim `Loading…` in the list.

- [ ] **Step 1: Write the failing tests.**
  - `TestConfigEveryRowHasHelp`
  - `TestConfigDetailShowsWholePath`: a 90-char reports dir, joined from the detail lines
    at 200x40, equals the path.
  - `TestConfigLayout`
  - `TestConfigLoading`
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Config out as a grouped list and detail`

### Task 5: Reports

**Files:**
- Create: `views/reports/log.go`, `views/reports/log_test.go`
- Modify: `views/reports/view.go`

- **Rows:**
  - grouped by the local date of `ModifiedAt`, through `DayList`, in the order of
    `viewhelpers.FilteredStrings(reportItems, filter)`
  - each row is the kind, the scope (when set), `DateLabel` or `Date`, and the format
- **Strip:**
  - Row 1: `Reports N`, then `newest <YYYY-MM-DD HH:MM>`
  - Row 2: `Dir ` and the reports dir, truncated from the left with a leading `…`
- **Detail:**
  - the title is `filepath.Base(Path)`
  - the meta is `[o] open  [e] edit  [D] delete`
  - the rows are `Kind`, `Period`, `Format`, `Size` (human-readable B, KB or MB),
    `Created`, and `Path`, wrapped whole
- **Empty:** `No exported reports yet`, then a dim `Export from Insights to see reports here`.

- [ ] **Step 1: Write the failing tests.**
  - `TestReportsGroupsByDay`
  - `TestReportsDetail` (with `1.2 KB` for 1234 bytes)
  - `TestReportsStripKeepsDirTail`: at width 40, row 2 ends with `/reports`.
  - `TestReportsEmpty`
  - `TestReportsLoading`
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Reports out as a strip, day list and detail`

### Task 6: Updates

**Files:**
- Modify: `views/updates/pane.go` (rewrite `renderView`; keep `statusCardLines`,
  `releaseNotesExcerptLines`, `diagnosticsSectionLines`, the migration checklist and the
  deprecation banner as section bodies)
- Test: `views/updates/pane_test.go`

- **Wide and medium:**
  - Left (`LogPaneWidths`' list width) holds a `Status` section (meta
    `[r] check again  [d] diagnostics`) over a `Diagnostics` section when expanded. The
    status section takes its content height, and diagnostics take the rest.
  - Right is `Release Notes`, at full height.
- **Narrow:** the sections stack.
- **Banner and checklist:** the install-script deprecation banner and the migration
  checklist go at the top of the Status body when they apply.

- [ ] **Step 1: Write the failing tests.**
  - `TestUpdatesColumns`: at 160x40, `Release Notes` and `Status` sit on the same line.
  - `TestUpdatesDiagnosticsToggle`: diagnostics show only when expanded.
  - `TestUpdatesNarrowStacks`
  - `TestUpdatesNoBox`
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Updates out as status and release notes`

### Task 7: Support

**Files:**
- Modify:
  - `views/support/view.go` (rewrite)
  - `views/types/main.go` (`SupportCursor int`)
  - `input/input.go` (`SupportCursor` on `State`; register `j`, `k`, `down`, `up` and
    `enter` on `ViewSupport`)
  - `model/core.go`, `model/input.go`, `model/layout_state.go` (plumb `supportCursor` like
    `updateDiagnosticsExpanded`)
- Test: `views/support/view_test.go`, `input/support_test.go`

**Interfaces:**
- Produces `supportActions` in `input`, an ordered slice of `{key string; run func(State, Deps) tea.Cmd}`:
  1. `o` feedback
  2. `r` releases
  3. `g` docs
  4. `c` copy
  5. `b` bundle

  The letter handlers and `enter` both call through it.
- Produces `support.Actions() []support.Action{Key, Label, URL string}` in views, in the
  same order. A test asserts the two orders match by key.

- **Left:** the `Actions` section:
  - one dim line: `Report bugs, share feedback, or build a support bundle.`
  - then rows of `[o] Feedback & roadmap  <url dim>`, with the cursor on `SupportCursor`
- **Right:** the `Diagnostics` section: `helperpkg.SupportDiagnosticsSummary`, wrapped, then
  `Transport`.

- [ ] **Step 1: Write the failing tests.**
  - `TestSupportEnterRunsSelectedAction`: for each cursor 0..4, `enter` returns a non-nil
    cmd, and the letter key at the same index does too. The fake deps record which was
    called; the two must match.
  - `TestSupportCursorClamps`: `k` at 0 stays 0, and `j` at 4 stays 4.
  - `TestSupportLayout`: the actions and diagnostics are side by side at 120; at 60 they
    stack, with no box.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): lay System's Support out as actions and diagnostics`

### Task 8: Bottom rows

**Files:**
- Modify: `model/view.go` (add `ViewSettings`, `ViewAlerts`, `ViewReports`, `ViewConfig`,
  `ViewUpdates` and `ViewSupport` to the globals-only case)
- Test: `model/bottom_row_test.go`

- [ ] **Step 1: Write the failing test.** `TestSystemBottomRowsGlobalsOnly`: for each of the
  6 views, the bottom row has `[?] all keys` and none of `[enter]`, `[e]`, `[r]`, `[o]` or
  `[←/→]`.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Also check that the help sheet still
  lists those keys for each view.
- [ ] **Step 5: Commit.** `fix(tui): keep System's bottom row global`

### Task 9: Visual check

- [ ] **Step 1: Capture each sub-tab at 60x20, 120x36 and 200x50.** Run
  `tmux new-session -d -s crsys -x W -y H "CRONA_ENV=Dev make run-tui"`, press `5`, then `]`
  between sub-tabs, and capture each with `capture-pane -p`.
  - Expected: no box, the grouped dividers, the details filled, whole paths, the Updates
    columns, and the Support cursor moving.
- [ ] **Step 2: Fix what the captures show.** Each fix gets a failing test first.
