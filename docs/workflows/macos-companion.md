---
hosted: true
title: macOS Companion
description: Use Crona for macOS as a native menu-bar client for the shared local runtime.
badge: Beta
order: 5.7
---

The macOS companion gives you quick access to the same Crona work from a native desktop surface. It connects to the local daemon, so your issues, sessions, check-ins, and away history remain one shared set of data.

:::caution[Beta compatibility]
The current companion integration requires a daemon that supports protocol `1.5`. Install matching beta builds when using the beta companion; a stable daemon may not expose every companion action or synchronization event.
:::

## Before You Install

Install core Crona and start the local daemon first. The Mac app requires macOS 14 or later and connects to the daemon through its local runtime metadata. Keep the app and daemon on matching release tracks; current beta builds require protocol `1.5`.

Download the signed DMG from the [Crona for macOS releases](https://github.com/webxsid/crona-macos/releases), move the app to Applications, and open it. If the daemon is unavailable, the app can open but daemon-backed actions remain unavailable until it reconnects.

## A daily workflow

- **Create and plan issues:** Open the menu bar and create an issue with a title, TUI-style estimate (`45m`, `1h30m`), repo, stream, description, and Today planning. Start focus after creation when appropriate.
- **Check in on wellbeing:** Use the Wellbeing tab for today’s mood, energy, stress, sleep quality, and notes. An untouched day is a valid empty state, so the first check-in does not require prior data.
- **Keep focus visible:** Enable the optional floating timer HUD, then choose Compact, Regular, or Spacious sizing and a default screen position. Compact actions appear on hover without leaving the timer surface.
- **Handle breaks and away time:** The daemon owns break-deferral limits, Away Today, historical away dates, and the logical-day boundary. The Mac app acknowledges eligible break actions and refreshes when those shared settings change.

## When the terminal is better

The Mac app is for close-at-hand daily actions. Use the TUI and CLI for deeper planning, history editing, reports, automation, and the full set of workspace controls. Both clients are peers of the same local daemon, so switching surfaces does not require synchronization.

## Updates

The app uses Sparkle for beta and stable updates. Beta installs receive beta and stable releases; stable installs receive stable releases only. See the [macOS releases](https://github.com/webxsid/crona-macos/releases) page for the current build and full release notes.

Next: [Getting Started](../guides/getting-started.md) for the core Crona workflow.
