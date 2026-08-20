---
hosted: true
title: Getting Started
description: Install Crona, create one issue, and complete your first focused session.
order: 1
---

# Your first workday

This guide takes you from an empty install to a completed focus session. It uses one small example so you can learn Crona by doing something useful.

## 1. Open Crona

Run:

```bash
crona
```

The command starts the local daemon when needed and opens the TUI. Your data stays on this machine.

## 2. Create a place for the work

Create a repository for the kind of work you are doing, such as `work` or `personal`. Inside it, create a stream such as `crona-site` or `this-week`.

In a planning view, press `a`, choose **Repository**, submit it, then press `a` again to add a **Stream**. Press `c` on the stream to make it your active context.

Repositories are broad buckets. Streams are the lanes inside them. You can learn the full model in [Concepts](concepts.md).

## 3. Add one issue

Create an issue for the next concrete thing you want to move forward. Give it a short title, an estimate if you know one, and a planning date if it belongs to a particular day.

Press `a`, choose **Issue**, fill in the title, and submit with `ctrl+s`. Select the issue when it appears.

Start small: `tighten the docs navigation` is more useful than `work on the website`.

## 4. Start focus

Select the issue and start a focus session. Crona ties the timer to the issue, so the session already knows what you intended to work on.

Press `f` on the selected issue. When you are done, press `x`, write the summary, and submit it with `ctrl+s`.

Use a Pomodoro profile when you want structured breaks. Use the no-break timer when you want one countdown.

## 5. Finish with a record

When you stop, write a short summary of what changed:

```text
Removed the old docs links and pointed the site at docs.crona.work.
```

That summary becomes part of the session history. Tomorrow, you can see what moved without reconstructing the day from memory.

## What to read next

- [Concepts](concepts.md) explains repositories, streams, issues, contexts, and sessions.
- [Issues and Planning](../workflows/issues-and-planning.md) covers planning dates and daily work.
- [Focus Sessions](../workflows/focus-sessions.md) covers interruptions, stashes, and session history.
- [CLI and Local Engine](../reference/cli-and-local-engine.md) covers scriptable commands.
