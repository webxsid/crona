---
hosted: true
title: Usage & Diagnostics
description: Telemetry configuration, sanitization rules, and support bundles.
order: 5.9
---

Crona processes diagnostics locally. Users control telemetry and exception submission preferences via the **Privacy & Diagnostics** panel in the TUI Settings.

:::note[Beta diagnostics]
Beta diagnostics include the release channel and protocol version so a report can identify the runtime being tested. They do not include your issues, habits, notes, check-ins, or other work data.
:::

## Telemetry & Exception Tracking

Crona divides reporting into two optional tracks:
1. **Usage Telemetry**: Tracks high-level execution events to help verify features.
2. **Error Reporting**: Captures runtime crashes, panics, and handled errors for bug resolution.

## Shared Data Payloads

When reporting is enabled, only non-identifying application metadata is transmitted:
- **System details**: OS platform, CPU architecture, Go runtime version.
- **Application details**: Binary version, build tags, active release channel (`stable` or `beta`), telemetry consent status.
- **Error reports**: Exception messages, stack traces, and the causing function path.

**No user data** (such as issues, habits, notes, check-ins, or directory paths) is included in automatic reports.

## Local Redacted Support Bundles

For detailed debugging, you can generate a **Support Bundle** from the TUI **Support** view:
- **Payload**: Includes the local configuration structure (keys and secrets redacted), recent error log segments, and the database schema status.
- **Writers**: The bundle is compiled as a compressed archive and saved to your local temp directory.
- **Handoff**: You can review the archive contents before manually attaching it to a GitHub issue or discussion thread.
