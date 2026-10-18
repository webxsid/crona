# TUI Issues tab: scope tree, Habits sub-tab, in-pane keys

Status: approved in chat 2026-10-18. Builds on `2026-10-17-tui-layout-design.md` and `2026-10-18-tui-key-hints-design.md`.

## Goal

The Issues tab's Meta view is four bordered boxes (Repos, Streams, Issues, Habits) chained by checkout, and its Issues pane repeats the Issues view. Replace it: repos and streams become a scope tree beside the issue list, habit definitions get their own sub-tab, and both sub-tabs get Daily's treatment: borderless sections, keys inside the panes, globals-only bottom row.

Sub-tabs become `Issues · Habits`. The Meta view is removed.

## Layout

Both sub-tabs share one shape:

```
 Open    4  ███████░░░  in progress 1  blocked 0      Due   today 0  overdue 2
 Closed  1  ██░░░░░░░░  done 1  abandoned 0           Est.  3h over 12 issues
─ Scope ─────────── │ ─ Issues ───────────── 12  [←/→] │ ─ Fix sync conflict ──────────
  All          24   │  Open  Done                       │ status   in progress   effort 1h/2h
▶ crona        18 ● │ [a] new issue  [A] ad hoc         │ context  crona › tui
    tui         9 ● │ ▶ Fix sync conflict  [due 05-28]  │
    daemon      6   │   [enter] open  [s] status  [d] due│ Notes …
  dotfiles      6   │                                   │
 [a] new repo       │   Write release notes             │
                    │   crona › docs | 1h/2h | ready    │
```

- **Summary strip (2 rows max)**: the four summary boxes become a two-row strip like Daily's day strip — labels in one column, values aligned, column groups dropped right-first when narrow.
  - Issues: `Open n ▕bar▏ in progress n  blocked n` over `Closed n ▕bar▏ done n  abandoned n`; then `Due today n  overdue n` over `Est. <time> over n issues`.
  - Habits: `Active n  paused n  due today n` over `Done today n  failed n  best streak <len>`.
  - The strip counts the browse scope (below), so it moves as the tree cursor moves.
- **Scope pane (left, ~24 cols)**: `All`, then each repo with its streams indented beneath it, each row with its open-issue count (Habits sub-tab: habit count). The checked-out repo and stream carry `●`.
- **List pane (middle)**:
  - Issues: an `Open  Done` selector (`[←/→]`), a create line `[a] new issue  [A] ad hoc` (`[A]` only while no timer runs), and two-line cards as on Daily (title + date; context | effort | status).
  - Habits: a create line `[a] new habit`, and two-line cards: name, then `context | schedule | target`. All habit definitions in scope, not only today's.
- **Detail pane (right, wide only)**: the selection's detail, reusing Daily's issue and habit detail lines. No `[S]` (there is no day summary here), so no title meta.
- **Widths**: wide (≥160) shows all three; medium drops the detail pane; narrow (<100) collapses the Scope pane to one focusable row above the list: `Scope ‹ crona › tui ›` (+ `●` when checked out).
- **Bottom row**: globals only, as on Daily.

## Keys

Panes are Scope then the list; `[tab]` cycles them. The detail pane never takes focus.

- **Scope pane**: `[↑/↓]` move (narrow: `[←/→]` step through the same flattened order). Action line on the focused selection: `[c] check out  [e] edit  [D] delete  [enter] open` (`All` offers none of these). Create line under the tree: `[a] new repo`, or `[a] new stream` when the cursor is on a repo or one of its streams.
- **Issues list**: the existing Issues-view keys, shown as on Daily — the focused selected card's meta line becomes `[enter] open  [s] status  [d] due  [P] pin`, plus `[f] focus  [m] log  [e] edit  [D] delete` while no timer runs; ad hoc: `[enter] open`, plus `[e] convert` while no timer runs.
- **Habits list**: Meta's habit keys — `[enter] open  [e] edit  [D] delete`, plus `[m] log` while no timer runs. Pausing stays in the edit dialog.
- Every hint shown must be a key the input layer honours in that state; hints are whole (`FitHints`).

## Browse scope

- New view state, separate from the checked-out context: a repo ID and optional stream ID, or All.
- Set to the checked-out context each time the Issues tab is entered; shared by both sub-tabs.
- Scopes the list and the strip. `[c]` checks the browsed node out (global context, as Meta's `[c]` does today).
- Creating an issue or habit defaults its repo/stream to the browse scope.
- If the browsed repo or stream is deleted, the scope falls back to All.

## Removed

- `ViewMeta`, `views/meta`, its key bindings (including the no-op `y`), its `ContextualActions` cases, and its entries in tabs, view-jump and the `?` sheet. References elsewhere (`dispatch`, `model`, `selection`, `dialogs`) move to the Issues view or the scope pane.
- The Issues view's bordered boxes, in-pane action lines (`RenderPaneActionLine`) and table rows.

## Out of scope

Daemon changes (the tree uses existing repos, streams, issues, habits); archiving or pausing from a key; other tabs.

## Testing

Render tests per pane (tree rows, `●`, counts, create/action lines, focus-only action line, timer gating), strip rows ≤ 2 and width-exact at 60/100/160, narrow scope row, browse-vs-checkout state (move doesn't change context; `[c]` does; entering the tab resets), create defaults, Meta gone from tabs/jump/help, bottom row globals-only.
