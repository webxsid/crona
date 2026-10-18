# Dialog Chrome Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the shared dialog chrome (frame, fields, pickers, choices, footer) and move
New Issue and the context switcher onto it.

**Architecture:**
- `dialogs/frame.go` holds pure render helpers.
- The controller `State` gains the repo and stream option labels, so a focused picker can
  list its matches.
- The two proof dialogs are re-rendered with the helpers. Inputs, focus and keys are
  untouched.

**Tech Stack:** Go, Bubble Tea (bubbles textinput and textarea), lipgloss, x/ansi.

**Spec:** `docs/superpowers/specs/2026-10-18-tui-dialog-chrome-design.md`

## Global Constraints

- Paths below are relative to `tui/internal/tui/`.
- Tests: `cd tui && GOCACHE=/tmp/crona-go-cache go test ./...`.
- Run `gofmt` on touched files only.
- Rendering only: no change to `dialogs/controller` update logic, key handling or focus
  order. The one exception is the two new State fields from Task 2.
- Text inputs render through their own `View()` with `Prompt` set to `""` on a copy, so the
  cursor and placeholder behaviour stays native.
- Commit on `brainstorm/tui-layout`.

## Review Focus

1. **Short terminals (height 20):** New Issue must keep the focused field and the footer
   visible. The old `formModal` scrolled to the active input, and `Frame`'s `↓ more` must
   not hide the focused field.
2. **The ad-hoc variants of `create_issue_default`** (`AdhocTimer`, `PromotingAdhoc`):
   these keep their titles, the `ViewName` line and the required session summary field.
3. **No repos:** a picker with zero options shows its placeholder (`Type to search or
   create`) and no empty list.
4. **The "Create New Repo: x" option:** it appears in the picker list like any other.
5. **Error messages:** they stay visible under the field area when the form is long.

---

### Task 1: Frame primitives

**Files:**
- Create: `dialogs/frame.go`, `dialogs/frame_test.go`

**Interfaces:**
- Produces:
  - `type FrameSpec struct{ Title, Context string; Danger bool; Body, Footer []string; ScreenWidth, ScreenHeight int; Focus int }`
    - `Focus` is the body line index that must stay visible when the body truncates (-1
      means none)
  - `Frame(theme Theme, spec FrameSpec) string`
  - `FrameInner(screenWidth int, content []string) int`: the inner width the frame will
    use, clamped as in the spec. Callers lay out fields with it.
  - `Field(theme Theme, label string, column int, value string, focused bool) string`
  - `FieldPair(left, right string, inner int) []string`
  - `Textarea(theme Theme, label string, column int, body string, focused bool, inner int) []string`
  - `type PickerSpec struct{ Label string; Column int; Focused bool; Input string; Current string; Options []string; Selected int; Placeholder string }`
    - `Input` is the rendered filter input
  - `Picker(theme Theme, spec PickerSpec) []string`
  - `type Choice struct{ Key, Label, Help string }`
  - `Choices(theme Theme, choices []Choice, selected, inner int) []string`
  - `Footer(theme Theme, primary string, hints []string, inner int) []string`

- **Frame:**
  - The top border is built by hand as `╭─ Title ─…─ Context ─╮`, with the context
    dropped when it does not fit.
  - The sides are `│`, padded 2 cells.
  - Lines are wrapped to the inner width with `ansi.Wrap`.
  - The border colour is the theme's `ColorCyan`, or `ColorRed` when `Danger`.
  - When `len(body)+len(footer)+4 > ScreenHeight-4`, the body is windowed around `Focus`
    with `viewchrome.ListWindow`, and a dim `↓ more` / `↑ more` marks the cut edge.

- [ ] **Step 1: Write the failing tests.**
  - `TestFrameTitleInBorder`: the first line starts with `╭─ New Issue ` and ends with `╮`.
  - `TestFrameContextDroppedWhenTight`: a 40-char title plus a 30-char context at screen
    width 60 shows the title and not the context.
  - `TestFrameWidthClamp`: tiny content gives an outer width of 44+6. Long content at
    screen 200 gives at most 76+6. At screen 60 it is at most 52.
  - `TestFrameNoTrailingBlank`: the line before the bottom border is the last footer line.
  - `TestFrameKeepsFocusVisible`: a 40-line body at screen height 20 with `Focus` 30 keeps
    body line 30 and the footer, and shows `↑ more`.
  - `TestFieldPadsLabel`: `Field(…, "Due", 8, "x", false)` starts with `Due` followed by 7
    spaces (column 8 plus 2).
  - `TestFieldPairFallsBack`: at inner 50 it gives 2 lines; at 60 it gives 1.
  - `TestPickerListOnlyWhenFocused`: unfocused gives 1 line ending ` ▾`. Focused with 7
    options gives the input line, 5 options and `+2 more`, with `▶` on the selected one.
  - `TestPickerEmptyShowsPlaceholder`: no options and unfocused shows the placeholder.
  - `TestFooterNeverSplits`: no hint is broken across lines; the primary comes first;
    `esc cancel` comes last.
- [ ] **Step 2: Run them.** Expected: a build failure.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): add shared dialog chrome`

### Task 2: Picker options in the render state

**Files:**
- Modify:
  - `dialogs/controller/types.go`: `RepoOptionLabels, StreamOptionLabels []string`
  - `model/layout_state.go` (`dialogRenderState`): for `create_issue_default`,
    `create_habit` and `checkout_context`, fill both from the same option functions the
    labels use (`DefaultRepoOptions` / `DefaultStreamOptions`, and the Checkout variants)
- Test: `model/dialog_options_test.go`

- [ ] **Step 1: Write the failing test.** `TestDialogRenderStateCarriesPickerOptions`:
  - a model on `checkout_context` with repos `OSS` and `Work` gives
    `RepoOptionLabels == [OSS Work]`
  - after typing `wo` into input 0, it gives `[Work, Create New Repo: wo]`
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): pass picker options to dialog rendering`

### Task 3: New Issue on the new chrome

**Files:**
- Modify: `dialogs/issues.go`, the `create_issue_meta` and `create_issue_default` cases
- Test: `dialogs/issues_chrome_test.go`

- **Fields:**
  - Title; a Description textarea; Estimate and Due as a `FieldPair`
  - `create_issue_default`: Repo and Stream pickers, focused when `FocusIdx` is 0 or 1
  - `create_issue_meta`: Repo and Stream as plain fields, from `RepoName` and `StreamName`
  - ad-hoc: the `ViewName` line, plus `Summary` (Inputs[5]) when `AdhocTimer`
- **Context:** `RepoName › StreamName` when both are set.
- **Footer:** the primary is `ctrl+s save` (or the existing submit label), then the
  existing `issueCreateHint` actions, rewritten as `key action`, then `tab next` and
  `esc cancel`.
- **Errors:** `ErrorMessage` goes under the last field.
- **Focus:** the body line of the focused input.

- [ ] **Step 1: Write the failing tests.**
  - `TestNewIssueChrome`: at width 120 there is no `[ > `, no `⌕` and no `Description
    (Optional)`. It has `╭─ New Issue`, `Title`, `Estimate` and `Due` on one line, and the
    footer `ctrl+s save`.
  - `TestNewIssuePickerFocused`: with `FocusIdx` 0 and 3 repo options, the options are
    listed.
  - `TestAdhocIssueKeepsSummary`: with `AdhocTimer`, `Summary` is present and the title is
    `Create Issue & Finish Session`.
  - `TestNewIssueShortTerminal`: at height 20 with focus on Due, the Due line and the footer
    are visible.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS. Update old assertions that pin the old
  layout strings, and list each one in the commit body.
- [ ] **Step 5: Commit.** `feat(tui): draw New Issue with the dialog chrome`

### Task 4: Context switcher on the new chrome

**Files:**
- Modify: `dialogs/repo_stream.go`, the `checkout_context` case and `checkoutHint`
- Test: `dialogs/checkout_chrome_test.go`

- **Pickers:** the Repo picker is focused when `FocusIdx` is 0, and the Stream picker when
  it is 1.
- **Footer:** `enter set context`, `c clear`, `↑/↓/tab move`, `esc cancel`.

- [ ] **Step 1: Write the failing tests.**
  - `TestCheckoutChrome`:
    - the title is `Checkout Context`
    - no `[ > `
    - the Repo picker lists options when focused
    - the Stream line ends with `▾`
    - the footer is not split across lines at width 60
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run the full suite.** Expected: PASS.
- [ ] **Step 5: Commit.** `feat(tui): draw the context switcher with the dialog chrome`

### Task 5: Visual check

- [ ] **Step 1: Capture both dialogs at 60x20, 120x36 and 200x50.** Start
  `CRONA_ENV=Dev make run-tui` in tmux. Open New Issue with `2` then `a`, and the context
  switcher with `c`. Tab through the fields and type into the repo filter.
- [ ] **Step 2: Fix what the captures show.** Each fix gets a failing test first.
