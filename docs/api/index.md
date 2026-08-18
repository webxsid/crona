---
title: API Reference
description: The local Crona daemon API for TUI, CLI, companion, and automation clients.
hosted: true
order: 7.0
---

# API Reference

Crona clients communicate with the local daemon through a request/response socket API. This reference is organized by responsibility so you can find the contract for a feature without reading one large method catalogue.

## Start here

1. Read [Transport and Envelopes](./transport.md) to connect and decode messages.
2. Call `kernel.info.get` and compare its `protocolVersion` with the client-supported version.
3. Subscribe with `events.subscribe` before rendering live state.
4. Use the domain references below for request names, payloads, results, and side effects.

## Domain references

| Area | Covers |
| --- | --- |
| [Transport and Envelopes](./transport.md) | Local endpoints, request/response envelopes, errors, and compatibility. |
| [Runtime and Operations](./runtime.md) | Health, daemon lifecycle, updates, alerts, and operations history. |
| [Work Management](./work-management.md) | Repositories, streams, issues, habits, and daily planning. |
| [Focus and Wellbeing](./focus-and-wellbeing.md) | Sessions, timers, context, check-ins, momentum, dashboards, and metrics. |
| [Exports and Settings](./exports-and-settings.md) | Reports, calendars, settings, and export configuration. |
| [Events](./events.md) | Subscription behavior and events emitted as state changes. |

## Contract rules

- The current local IPC protocol is **1.5**. It is independent of the Crona release version.
- The canonical method names are defined in [`shared/protocol/methods.go`](../../shared/protocol/methods.go).
- Request and response DTOs are defined in [`shared/dto`](../../shared/dto/requests.go).
- Before `1.0.0`, generated clients should prefer the shared Go types over assumptions from prose.
- This is a local API, not a remotely exposed network service.

For the complete legacy overview, see [Socket API](../reference/socket-api.md).
