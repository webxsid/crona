---
hosted: true
title: macOS Companion
description: Use Crona for macOS as a native menu-bar client for the shared local runtime.
order: 5.7
---

Crona for macOS is a native client for the same local daemon used by the CLI and terminal UI. It does not create a separate account or copy of your work; issues, sessions, check-ins, day boundaries, and away history remain daemon-owned.

## Before You Install

Install core Crona and start the local daemon first. The Mac app requires macOS 14 or later and connects to a runtime directory containing the daemon discovery metadata and local socket. Current macOS beta releases require daemon protocol `1.5`.

Download the signed DMG from the [Crona for macOS releases](https://github.com/webxsid/crona-macos/releases), move the app to Applications, and open it. If the daemon is unavailable, the app can open but daemon-backed actions remain unavailable until it reconnects.

## Native Daily Workflow

- **Create and plan issues:** Open the menu bar and create an issue with a title, TUI-style estimate (`45m`, `1h30m`), repo, stream, description, and Today planning. Start focus after creation when appropriate.
- **Check in on wellbeing:** Use the Wellbeing tab for today’s mood, energy, stress, sleep quality, and notes. An untouched day is a valid empty state, so the first check-in does not require prior data.
- **Keep focus visible:** Enable the optional floating timer HUD, then choose Compact, Regular, or Spacious sizing and a default screen position. Compact actions appear on hover without leaving the timer surface.
- **Handle breaks and away time:** The daemon owns break-deferral limits, Away Today, historical away dates, and the logical-day boundary. The Mac app acknowledges eligible break actions and refreshes when those shared settings change.

## When to Use the Terminal

The Mac app is for close-at-hand daily actions. Use the TUI and CLI for deeper planning, history editing, reports, automation, and the full set of workspace controls. Both clients are peers of the same local daemon, so switching surfaces does not require synchronization.

## Updates and Support

The app uses Sparkle for beta and stable updates. Beta installs receive beta and stable releases; stable installs receive stable releases only. See the [macOS releases](https://github.com/webxsid/crona-macos/releases) page for the current build and full release notes.
