---
hosted: true
title: TUI Keymap Reference
description: Quick-reference keybindings for global navigation, dialog controls, and view-specific commands.
badge: Beta
order: 6.4
---

Crona's Terminal User Interface is keyboard-driven and context-aware. Use the footer of each pane for active key guides.

:::caution[Beta-only support actions]
The `F9` support surface is available only in beta builds. The `F6` and `F7` shortcuts remain restricted to development builds.
:::

## Global Bindings

These shortcuts remain active across most standard views:

| Key | Description |
| --- | --- |
| `v` | Open the view selection jump menu |
| `[` / `]` | Cycle through views (restricted when timer is active) |
| `tab` / `shift+tab` | Cycle pane focus in the active view |
| `up` / `down` | Navigate the current selection up / down |
| `?` | Toggle help overlay modal |
| `ctrl+c` | Quit TUI |
| `K` | Request local daemon shutdown |

After opening view jump with `v`, press a view mnemonic directly. For example, `g` opens **Summary** and `y` opens **History**. Use the arrow keys and `enter` to choose from the menu instead.

### Development and beta hotkeys
- `F9`: Open the support dialog (beta builds only)
- `F6`: Seed mock wellbeing and focus data (development builds only)
- `F7`: Clear the local database (development builds only)

## Common Work Bindings

These apply to items in most planning tables or lists:
- `a`: Create a new entry (issue, habit, repo, stream).
- `c`: Checkout selection context or open checkout modal.
- `e`: Edit the selected item.
- `d`: Change the selected issue's planning due date.
- `D`: Delete the selected item.
- `f`: Start focus session for the selected issue.
- `m`: Log manual work session or habit completion.
- `s`: Open the issue status menu, including **Abandoned**.
- `w`: Open the daily wellbeing check-in in Daily or Wellbeing.
- `W`: Toggle away mode in Daily, Wellbeing, or Away.

## View-Specific Shortcuts

### Summary
- `,` / `.`: Move the summary date backward / forward.
- `g`: Reset the summary date to today.
- `up` / `down`: Scroll a long summary.

### Daily Dashboard
- `,` / `.`: Page calendar date backward / forward.
- `g`: Reset view date to today.
- `x`: Toggle selected habit completion.
- `F`: Mark selected habit as failed.

### Active Focus Timer
- `p`: Pause running session.
- `r`: Resume paused session.
- `x`: Stop session (opens summary prompt).
- `i`: Focus active issue details.

### Calendar Range Picker (Rollup View)
- `S`: Set range start date.
- `E`: Set range end date.
- `g`: Reset range to current week.

## Dialog & Text Form Controls

- `esc`: Cancel, dismiss, or go back.
- `ctrl+s`: Submit form data.
- `tab`: Shift focus to the next input field.
- `ctrl+e` / `ctrl+y`: Open calendar picker in date fields.
- `g`: Set date field to today.
- `c`: Clear text in selector fields.
- `left` / `right` / `enter`: Cycle options and select in popups.
