---
title: Beta Channel
description: Decide whether the beta channel is right for you and keep beta installs compatible.
hosted: true
badge: Beta
order: 1.5
---

# Beta Channel

Crona beta builds contain work that is ready for wider testing but has not reached a stable release yet. They may change between releases, and a beta daemon can expose capabilities that stable clients do not know about.

:::caution[Beta software]
Use beta when you want to try the newest TUI, companion, or API behavior and are comfortable reporting problems. Keep the clients and daemon on the same release track; a stable client should not depend on a beta-only method, field, or event.
:::

## Identify the channel

Run `crona daemon info` to inspect the running version, release channel, and protocol version. A version ending in `-beta.N` and a channel of `beta` identify a beta build. Client developers can read the same metadata through `kernel.info.get`.

## What beta can change

Beta releases can introduce changes in these areas before they reach stable:

- The TUI `F9` support actions are available only in beta builds.
- Focus-score range data and Start of Day synchronization require the current protocol supported by the beta daemon.
- The macOS Companion may require a matching beta daemon and protocol version.
- Client developers should read `kernel.info.get` before using newly documented methods or fields. The [API Reference](../api/) explains the compatibility rules.

## Package channels

Stable and beta installations use separate packages:

| Platform | Stable | Beta |
| --- | --- | --- |
| Homebrew | `crona` | `crona-beta` |
| Scoop | `crona` | `crona-beta` |

Install the beta package explicitly:

```bash
# macOS or Linux
brew install webxsid/tap/crona-beta

# Windows PowerShell
scoop install crona-beta
```

Beta builds receive beta and stable release information. Stable builds receive stable releases only. Before switching an existing installation between channels, back up your local data and follow the package manager's uninstall/install flow. The channel-specific package names above are the important part.

## Reporting a beta problem

When something looks wrong, include the Crona version, release channel, protocol version, operating system, and a short reproduction. The TUI's beta support actions can prepare diagnostics without including your issues, notes, or other work data.
