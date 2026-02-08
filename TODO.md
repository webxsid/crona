# TODO

This document tracks **open work, decisions, and planned iterations** for the Crona monorepo.  
It is intentionally opinionated and mirrors how the project itself thinks about work.

---

## 🔥 Immediate (Blocking Next Phase)

### Kernel / Core
- [ ] Finalize **scratchpad filesystem contract**
  - directory layout rules
  - rename / move semantics
  - deletion policy
- [ ] Add **scratchpad open / touch** semantics
  - update `lastOpenedAt`
- [ ] Harden scratch variable expansion
  - escaping rules
  - nested paths
- [ ] Enforce **commit message presence** at session end everywhere
- [ ] Add session **amend** command
  - amend last session only (v1)
  - rewrite notes safely
- [ ] Formalize **session history model**
  - ordering guarantees
  - pagination
- [ ] Introduce `session_logs` read model

---

## 🧠 Timer & Sessions

- [ ] Boundary transitions as explicit ops
- [ ] Recover timer state after kernel restart
- [ ] Persist boundary schedule metadata (optional)
- [ ] Long break logic polish
- [ ] Manual override of next boundary
- [ ] Expose remaining time to next boundary
- [ ] Timer drift tolerance tests

---

## 🧳 Stash

- [ ] Decide stash + timer interaction rules definitively
- [ ] Store paused segment snapshots more robustly
- [ ] Support multiple stashes per device
- [ ] Add stash preview (context + elapsed)
- [ ] Stash apply should optionally restart timer

---

## 🗂 Active Context

- [ ] Context switch history
- [ ] Context bookmarks
- [ ] Context-aware defaults for commands
- [ ] Validate cross-device context conflicts

---

## 📝 Scratchpads

- [ ] Editor integration strategy (nano / custom)
- [ ] Multi-buffer support in TUI
- [ ] Read-only preview mode
- [ ] Scratchpad search
- [ ] Scratchpad diff history (optional)
- [ ] Pin ordering
- [ ] Global vs project scratchpad conventions (doc-only)

---

## 🖥 TUI (Ink)

### Layout
- [ ] Lock final panel layout
- [ ] Repo / Stream / Issue browser panes
- [ ] Timer-focused view vs Context view
- [ ] Command palette (:`<command>`)
- [ ] Status bar (kernel, timer, context)

### Interaction
- [ ] Keybinding system
- [ ] Modal input handling
- [ ] Scrollable lists
- [ ] Focus management
- [ ] Mouse support (optional)

### Visuals
- [ ] Color theme system
- [ ] Accessibility (no icon-only semantics)
- [ ] Terminal resize handling

---

## 📦 CLI (Future)

- [ ] Thin CLI client (non-TUI)
- [ ] Scriptable commands
- [ ] JSON output mode
- [ ] Kernel attach/detach

---

## 📤 Export / Review

- [ ] Daily log export (Markdown)
- [ ] Weekly summary export
- [ ] Session → Issue rollups
- [ ] Repo-level reports
- [ ] Timeline view

---

## 🧪 Testing

- [ ] Stress test boundary scheduler
- [ ] Multi-device simulation
- [ ] Kernel crash recovery tests
- [ ] Long-running session tests
- [ ] File-system race tests (scratchpads)

---

## 🧭 Docs & Meta

- [ ] Architecture deep-dive
- [ ] Kernel API reference
- [ ] TUI keybinding reference
- [ ] Philosophy / design principles doc
- [ ] Contribution guide

---

## 🧩 Open Design Questions

- Should session amend allow arbitrary session IDs?
- Should scratchpads ever be auto-created?
- How strict should issue lifecycle transitions be?
- Should kernel ever run headless in background?
- Should ops log be user-visible by default?

---
