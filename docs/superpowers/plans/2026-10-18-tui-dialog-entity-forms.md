# Dialog Entity Forms Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Draw the 12 entity-form dialogs, and New Issue, through one shared form builder
on the dialog chrome.

**Architecture:**
- `dialogs/form.go` adds `renderForm`. It lays out a list of `FormRow`s with `formBody`,
  builds a padded footer with `formFooter`, and hands both to `Frame`.
- Each dialog becomes a short `FormSpec`.
- New Issue moves onto the builder first, so its existing tests prove the builder.

**Tech Stack:** Go, Bubble Tea (bubbles textinput and textarea), lipgloss, x/ansi.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-dialog-entity-forms-design.md` (builds on
`2026-10-18-tui-dialog-chrome-design.md`)

## Global Constraints

- Paths below are relative to `tui/internal/tui/`.
- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`.
- Run `gofmt` on touched files only. Never run `make fmt`.
- Rendering only: no change to `dialogs/controller` update logic, keys or focus order.
- Field focus follows the controller's convention:
  - the textarea is focused at `state.DescriptionIndex`
  - input `k` is focused at `k`, or at `k+1` when `k >= DescriptionIndex` (dialogs without a
    textarea use `k`)
  - each opener in `dialogs/controller/controller_openers_entities.go` confirms its values
- Labels have no `(Optional)`. Optional fields use the placeholder `optional`.
- Border colours: create forms are cyan, edit forms and the status note are yellow, and
  Habit Log is green.
- Footer order:
  - the primary hint first (the submit chord plus `create`, `save` or `set`)
  - then `tab next`
  - then `enter newline`, only while the textarea is focused
  - then the focused field's extra keys
  - `esc cancel` last
- Commit on `brainstorm/tui-layout`.

## Review Focus

1. **A validation error on a short screen:** it stays visible above the footer, through
   `withError`.
2. **Narrow screens (inner width under 56):** pairs stack, and focus on the second field of a
   stacked pair keeps that line visible.
3. **Edit dialogs with no repo or stream names:** no `›` and no empty context in the border.
4. **A configured submit chord other than ctrl+s:** the primary hint shows it, through
   `dialogSubmitChord`.
5. **Long `ViewName` or issue titles in the border context:** they are dropped by `Frame`
   when they don't fit, and never wrap the top border.

---

### Task 1: Form builder and New Issue on it

**Files:**
- Create: `dialogs/form.go`, `dialogs/form_test.go`
- Modify:
  - `dialogs/frame.go`: add `Border lipgloss.Color` to `FrameSpec`. Zero means cyan;
    `Danger` still wins.
  - `dialogs/chrome_forms.go`: rewrite `renderIssueCreate` onto `renderForm`

**Interfaces:**
- Produces:
  - `type FormRow struct{ Lines []string; Focus map[int]int }`
    - `Focus` maps a FocusIdx to the line offset inside `Lines`
  - `fieldRow(focus int, line string) FormRow`
  - `pairRow(left, right string, leftFocus, rightFocus, inner int) FormRow`
    - it uses `FieldPair`
    - when the pair stacks, `rightFocus` maps to offset 1
  - `staticRow(lines ...string) FormRow`
    - `Focus` is nil
    - `staticRow("")` is a spacer
  - `textareaRow(focus int, lines []string) FormRow`
  - `formBody(rows []FormRow, focusIdx int) (body []string, focusLine int)`
    - `focusLine` is -1 when no row owns `focusIdx`
  - `formFooter(theme Theme, primary string, hints func(focus int) []string, focus, focusCount, inner int) []string`
    - renders `Footer(theme, primary, hints(focus), inner)`
    - padded with blank lines to the tallest over `0..focusCount-1`
  - `type FormSpec struct{ Title, Context string; Border lipgloss.Color; Rows []FormRow; Primary string; Hints func(focus int) []string; FocusCount int }`
  - `renderForm(theme Theme, state controllerpkg.State, spec FormSpec) string`
    - the footer goes through `withError`
    - `Inner` is `formInner(state.Width)`
  - `formInner(width int) int`
    - returns `FrameInner(width, []string{strings.Repeat(" ", 72)})`
  - `formColumns(inner, column int) (full, half int)`
    - `full = inner - column - 3`
    - `half = inner/2 - column - 6`, or `full` when `inner < 56`
  - `plainHint(hint string) string`
    - strips the brackets from `[k] action`
    - the primary hint is `plainHint(dialogSubmitHint(state, label))`

- [ ] **Step 1: Write the failing tests** in `form_test.go`:
  - `TestFormBodyFocusLine`: rows `staticRow("a")`, `fieldRow(0, "x")` and
    `pairRow("l", "r", 1, 2, 50)` (stacked at 50). Then:
    - focus 0 gives line 1
    - focus 2 gives line 3
    - focus 9 gives -1
  - `TestFormFooterPaddedToTallest`: when `hints(1)` returns 6 long hints and `hints(0)`
    returns none, the footer for focus 0 has the same length as for focus 1.
  - `TestFrameBorderColour`:
    - `FrameSpec{Border: theme.ColorYellow}` renders the top border in yellow (compare with
      `lipgloss.NewStyle().Foreground(theme.ColorYellow).Render("╭")`)
    - with `Danger` it is red
- [ ] **Step 2: Run them.**
  `go test ./internal/tui/dialogs/ -run 'TestFormBody|TestFormFooter|TestFrameBorder'`.
  Expected: a build failure.
- [ ] **Step 3: Implement** `form.go` and the `FrameSpec.Border` field.
- [ ] **Step 4: Rewrite `renderIssueCreate` onto `renderForm`.**
  - Keep `issueCreateHint` and `splitHints` for its hints.
  - Its spec sets `FocusCount` to 7 for `create_issue_default` and 4 for
    `create_issue_meta`.
- [ ] **Step 5: Run the full suite.** Expected: PASS. All the `issues_chrome_test.go` tests
  pass unchanged.
- [ ] **Step 6: Commit.** `refactor(tui): build dialog forms through one form builder`

### Task 2: Repo and Stream forms

**Files:**
- Modify: `dialogs/repo_stream.go`, the cases `create_repo`, `edit_repo`, `create_stream`
  and `edit_stream`
- Test: `dialogs/entity_forms_test.go` (new)

**Interfaces:**
- Consumes: Task 1's `renderForm`, `fieldRow`, `textareaRow`, `plainHint`, `formInner` and
  `formColumns`.
- Produces: the test helpers for Tasks 3–4:
  - `type formCase struct{ name string; state controllerpkg.State; title, context string; focusCount int }`
  - `assertFormChrome(t, c formCase)`, which checks:
    - the stripped first line starts with `╭─ <title>`
    - it contains `context` (when that is non-empty)
    - the output has no `(Optional)`
    - the output has no `[` followed by a key and `]`, matching the regex `\[[a-z+/←→↑↓]+\]`
    - the last footer line before the bottom border's padding ends with `esc cancel`
  - `assertHeightStable(t, c formCase)`: for each `FocusIdx` in `0..focusCount-1`, after
    `SyncDialogFocus`, the line count is the same at width 120 and height 36

- **Fields:**
  - Name (`Inputs[0]`, placeholder `required`, focus 0)
  - Description (`textareaRow` at `DescriptionIndex` = 1)
  - `FocusCount` is 2
- **Context:** none for Repo; `RepoName` for Stream.
- **Hints:** `tab next`, plus `enter newline` at focus 1.
- **Border:** cyan for create, yellow for edit.

- [ ] **Step 1: Write the failing tests.** `TestRepoStreamForms` runs `assertFormChrome` and
  `assertHeightStable` over the four kinds, built with `OpenCreateRepo`, `OpenEditRepo`,
  `OpenCreateStream` (with `RepoName` `OSS`) and `OpenEditStream`.
  - The titles are `New Repo`, `Edit Repo`, `New Stream` and `Edit Stream`.
  - The context is `OSS` for both stream kinds.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Update old assertions that pin the old
  layout strings, and list each one in the commit body.
- [ ] **Step 5: Commit.** `feat(tui): draw repo and stream forms with the dialog chrome`

### Task 3: Habit forms and Edit Issue

**Files:**
- Modify:
  - `dialogs/repo_stream.go`: `create_habit` and `edit_habit`; delete `habitDialogHint` if it
    is no longer used
  - `dialogs/issues.go`: `edit_issue`
- Test: `dialogs/entity_forms_test.go`

- **New Habit** (focus order: Repo 0, Stream 1, Name 2, Description 3, Schedule 4, Target 5):
  - fields: Name (`Inputs[2]`); Description; a pair of Schedule (`Inputs[3]`, placeholder
    `daily, weekdays, mon,wed`) and Target (`Inputs[4]`, placeholder `optional`)
  - after a spacer, the Repo and Stream pickers, built exactly as in `renderIssueCreate`
  - `FocusCount` is 6
  - hints:
    - `type filter` and `←/→ choose` at focus 0–1
    - `enter newline` at focus 3
  - context: `RepoName › StreamName` when both are set
- **Edit Habit** (Name 0, Description 1, Schedule 2, Target 3):
  - the same fields, with no pickers
  - `FocusCount` is 4
  - yellow border
- **Edit Issue** (Title 0, Description 1, Estimate 2, Due 3):
  - fields: Title; Description; a pair of Estimate and Due, with placeholders `45m, 1h30m`
    and `YYYY-MM-DD`
  - `FocusCount` is 4
  - yellow border
  - hints:
    - the extra keys at focus 3 come from `issueDialogHint`, through `splitHints`, keeping
      `g today`
    - `enter newline` at focus 1
  - context: `RepoName › StreamName` when both are set

- [ ] **Step 1: Write the failing tests.**
  - `TestHabitAndEditIssueForms`: chrome and height for the three kinds.
  - `TestEditIssueDueKeepsTodayHint`: at focus 3 the footer contains `g today`.
  - `TestNewHabitShortTerminal`: at height 20 and width 60 with focus 5 (Target), both the
    Target line and `esc cancel` are visible.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Update and list any old assertions.
- [ ] **Step 5: Commit.** `feat(tui): draw habit and edit issue forms with the dialog chrome`

### Task 4: Status note, Check-In, Habit Log, Reminder

**Files:**
- Modify:
  - `dialogs/issues.go`: `issue_status_note`
  - `dialogs/utility.go`: `create_checkin`/`edit_checkin`, `complete_habit` and
    `create_alert_reminder`/`edit_alert_reminder`
- Test: `dialogs/entity_forms_test.go`

- **Status note:**
  - title from today's map
  - context `ViewName`
  - one field labelled `StatusLabel` (fall back to `Note`) on `Inputs[0]`
  - primary `set`
  - `FocusCount` is 1
  - yellow border
- **Check-In** (inputs 0–5, with no textarea):
  - context: the display date the renderer already computes
  - rows:
    - a pair of `Mood` and `Energy`
    - a pair of `Sleep` and `Sleep score`
    - `Screen time`
    - `Notes`
  - placeholders: `1-5`, `1-5`, `7h30m`, `0-100`, `2h`, `optional`
  - `FocusCount` is 6
  - border: cyan for New, yellow for Edit
- **Habit Log** (Duration 0, Notes 1):
  - title: `ViewTitle`, else `Habit Log`
  - context: `CheckInDate`
  - fields: Duration (`Inputs[0]`, `optional`); Notes as a textarea labelled `Notes`
  - `FocusCount` is 2
  - green border
  - `enter newline` at focus 1
- **Reminder:**
  - today's titles
  - fields: a pair of Schedule (`Inputs[0]`, placeholder `daily`) and Time (`Inputs[1]`,
    placeholder `HH:MM`)
  - then a spacer and the two help lines as a dim `staticRow`
  - `FocusCount` is 2
  - border: cyan for create, yellow for edit
  - render the inputs with `bareInput`, not `dialogInputView`

- [ ] **Step 1: Write the failing tests.**
  - `TestRemainingEntityForms`: chrome and height for:
    - `issue_status_note` (built with `OpenIssueStatusNote`)
    - the two check-in kinds
    - `complete_habit` (built with `OpenHabitCompletion`)
    - the two reminder kinds
  - `TestCheckInShortTerminal`: at height 20 and width 60 with focus 5 (Notes), both `Notes`
    and `esc cancel` are visible.
  - `TestEntityFormErrorVisible`: a check-in with `ErrorMessage` `Mood must be 1-5` at height
    20 shows that message.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Update and list any old assertions.
- [ ] **Step 5: Commit.**
  `feat(tui): draw status note, check-in, habit log and reminder forms with the dialog chrome`

### Task 5: Visual check

- [ ] **Step 1: Capture the forms at 60x20, 120x36 and 200x50.**
  - Start `CRONA_ENV=Dev make run-tui` in tmux.
  - Open each reachable form: repo, stream and habit create/edit, edit issue, a status
    change, a check-in, a habit log and a reminder.
  - Tab through the fields.
- [ ] **Step 2: Fix what the captures show.** Each fix gets a failing test first.
