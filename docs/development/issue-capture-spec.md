---
hosted: false
title: "Issue Capture Spec"
description: "Proposed bulk issue capture: annotated - lines parsed into issues through one review step, from an I dialog, notes, or the CLI."
order: 7.3
---

## Issue Capture

Status: proposal. Not scheduled. This document describes what bulk issue capture would be if built; it is not an implementation plan.

### Goal

Turn a block of text into issues in one step. Users type or paste `-` lines, each carrying optional annotations, review the parsed result once, and create everything together. The same parser and review serve three entry points: the `I` capture dialog, `:::issues` blocks in notes (see the Notes spec), and the CLI.

### Non-goals

- No LLM or network calls. Parsing is deterministic and local, so the TUI and CLI always agree.
- No sync back: once created, an issue is edited only through the normal issue flows.
- No statuses or time logging in the text. Lines create open issues; status changes and sessions use the normal flows.

### Task lines

#### Detection

A task line is a top-level `- text` list item.
- Lines indented under a task line become the issue's description.
- Lines inside fenced code blocks are ignored.
- Numbered lists, headings and prose are never tasks.
- In a note, only lines inside an `:::issues` block are considered, and a line already carrying a write-back marker (`→ CRN-142`) is skipped.

#### Inline annotations

All annotations are optional and may appear anywhere in the line. They are removed from the title.

| Annotation | Property |
|---|---|
| `#repo/stream` | Target repo and stream. |
| `#name` | Matched against both repo and stream names. A stream match targets that stream and its repo. A repo match targets that repo; if the repo has exactly one stream it is used, otherwise the stream is left unresolved. A name matching both a repo and a stream in different places is ambiguous. Case-insensitive prefix match; ambiguous, unknown or unresolved targets are flagged, never created silently. |
| `~45m`, `~1h30m` | Estimate. |
| `@today`, `@tomorrow`, `@mon`…`@sun`, `@2026-10-30` | Plan date. `@today` adds the issue to today's plan. |
| `+urgent`, `+client-acme` | Tags. Unknown names are created, as in the dialog Tags field. |

Without a `#` annotation the target is the caller's default: the active context's stream for the `I` dialog, the note's link for notes (see the Notes spec), `--repo`/`--stream` for the CLI, otherwise the Inbox stream.

#### Inbox stream

Crona keeps one built-in `Inbox` repo with a single `Inbox` stream as the default target for lines with no target. It behaves like any other repo and stream; issues can be moved out of it with the existing issue editing flows.

#### Natural phrases

The parser also understands a fixed set of plain phrases and maps them to the same fields. To avoid eating words from titles, a phrase is recognized only at the end of the line, after any annotations are removed, and only one phrase per field.

| Phrase | Field |
|---|---|
| `today`, `tomorrow`, `tonight` | plan date |
| `on friday`, `friday`, `next friday`, `this friday` | plan date (`next` skips the coming one) |
| `in 3 days`, `in 2 weeks`, `next week` (Monday) | plan date |
| `on oct 30`, `on 30 oct`, `on 2026-10-30` | plan date |
| `for 45 min`, `for 1h`, `for 1.5 hours`, `takes 20m` | estimate |

Examples:
- `fix login redirect #api +urgent for 45 min tomorrow` → title "fix login redirect", api stream, urgent, 45m, tomorrow.
- `read 2 min intro` → title unchanged: "2 min" has no `for`/`takes` and is not trailing.
- `call mom friday` → title "call mom", plan Friday.

An explicit annotation wins over a phrase for the same field, and the phrase then stays in the title. The review shows every parsed field, so a wrong guess is visible and editable before anything is created.

### Review

Parsing opens one review list:

```
   title                  repo   stream  est  plan   tags
[x] fix login redirect     crona  api     45m  tmrw   urgent
[x] write release notes    crona  docs    1h   today  —
[?] call bank              Inbox  Inbox   —    —      —
space toggle · ←/→ field · enter edit field · ctrl+s create · esc back
```

- Every field is editable: `←`/`→` moves between fields and `enter` edits the focused one. Title, estimate and plan date edit inline; repo and stream use the existing pickers; tags use the Tags field. Changing the repo clears the stream unless the repo has exactly one.
- All rows start selected except flagged ones (`?`). A flagged row becomes selectable once fixed.
- `ctrl+s` creates the selected issues, their tags and plan entries in one kernel call.

### Entry points

#### `I` capture dialog

`I` (global, outside dialogs) opens a multi-line "Add issues" dialog with the target defaulting to the active context.

- `ctrl+s` parses and opens the review. `esc` in the review returns to the text with its contents intact; `esc` in the text closes the dialog.
- Nothing is saved as a note; the created issues are the only output. Unselected rows are dropped when the dialog closes.
- `I` is unbound today. It pairs with `a` (create one entry); `A` stays reserved for the ad-hoc timer and `ctrl+a` is avoided as a tmux and screen prefix. Inside dialogs `I` is plain text.

#### Notes

Saving a note with new lines in an `:::issues` block runs the same parser and review, then writes `→ ID` markers back and links the note to each issue. Lists outside a block are untouched. The Notes spec covers block syntax, write-back and when triage runs.

#### CLI

- `crona issue add "text"` parses one line; `crona issue add -` reads lines from stdin.
- On a terminal it prints the parsed rows and asks to confirm; `--yes` creates every unflagged row and reports flagged ones; `--dry-run` prints the parse as JSON.
- `--repo`/`--stream` set the default target.

### Kernel and IPC

- parse: text plus a default target → proposed issues with fields and flags (no writes)
- create: an accepted list of proposals (possibly user-edited) → issues, tags and plan entries as one operation, returning created ids per row

The parser lives in the kernel.
