---
hosted: true
title: Getting Started
description: A quick setup and workflow guide for developers starting with Crona.
order: 1
---

Crona is a local-first work tracker for developers. It runs entirely on your local machine, keeping your task lists, focus timers, wellbeing tracking, and generated reports private and self-contained.

## The Runtime Architecture

Crona consists of three primary components:
- `crona-daemon`: The background local service that owns the SQLite storage, runs focus timers, manages scheduled reminders, checks for updates, and exposes the local IPC socket.
- `crona-tui`: The interactive Terminal User Interface for daily planning, logging, and focus sessions.
- `crona`: The scriptable CLI launcher and daemon controller.

## Quick Start (First 5 Minutes)

To start the terminal interface, run:
```bash
crona
```
This automatically starts `crona-daemon` if it is not already running and launches the TUI.

Follow these steps to track your first task:
1. **Create a Repository**: Press `r` to create a top-level bucket (e.g., `work`, `personal`).
2. **Create a Stream**: Press `s` to subdivide a repository (e.g., `main`, `refactor`).
3. **Add an Issue**: Press `a` to create a task. Set a title, estimate, and choose a timer profile.
4. **Checkout Context**: Press `c` (or use the arrow keys) to set the active repository and stream context.
5. **Start Focus**: Select your issue and press `Enter` to start the focus session timer.
