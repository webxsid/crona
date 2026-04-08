# Changelog

All notable changes to **Crona** are documented here.

Release channel policy:

- `stable` is the preferred channel for general users.
- `beta` is the testing channel for pre-release validation and faster iteration.
- `v1.0.0-beta.3` is the current `1.0.0` prerelease build for tester validation before the first stable release.

## [1.0.0-beta.3] - 2026-04-08

### Added
- Kernel info now exposes an independent `protocolVersion` so future GUIs can validate IPC compatibility without relying on the app release version.
- Daily report exports now surface plan accountability and failure signals more explicitly, including failed-count, accountability score, delayed/high-risk issue metrics, and failed-plan issue details in the default templates.
- The main TUI header now shows the running app version on the right side.

### Changed
- Alert and PDF renderer behavior is now documented more clearly across install, development, concepts, and socket API docs.
- The Alerts view now focuses on the active backend and capability support instead of advertising backend fallback chains.

### Removed
- macOS alert delivery no longer includes the NotifiCLI-specific backend path; supported notification helpers are now documented and implemented as `terminal-notifier` with `osascript` fallback.

## [1.0.0-beta.2] - 2026-04-06

### Added
- The wellbeing dashboard now exposes selectable summary and trends panes, so larger datasets can be scrolled with `j/k` and the arrow keys instead of clipping in place.

### Changed
- The TUI view layer has been fully refactored into first-class `views/<view>/` packages, with each migrated view using a dedicated `render.go` entrypoint and smaller per-view support files.
- Shared view concerns now live in dedicated packages such as `views/chrome`, `views/runtime`, `views/issuecore`, `views/sessionmeta`, `views/settingsmeta`, and `views/contextmeta`, reducing the old root `views` package to a thin type/export surface.
- Wellbeing large-screen rendering now treats summary and trends as independent panes with active-pane borders and scroll indicators for overflow content.

### Fixed
- Wellbeing long-form sections no longer lose content behind pane-box clipping when the selected date has dense check-in, accountability, heatmap, or burnout data.

## [1.0.0-beta.1] - 2026-04-05

### Added
- Running release awareness is now exposed across the kernel, updater, and TUI, including the distinction between the running release channel, the configured update channel, and the latest fetched release kind.
- Beta builds now expose a global `[f9]` support menu for quick bug reporting, releases, roadmap access, diagnostics copy, and support-bundle generation.
- A global `[v]` view-jump chooser now provides direct mnemonic navigation across the available TUI views.
- Socket API documentation and the new docs tree now provide clearer open-source reference material for contributors and advanced users.

### Changed
- The docs layout now lives under `docs/`, with the root `README.md` reduced to a cleaner landing page and direct doc links.
- View-jump choices are now context-aware, so unavailable views like `Away` or the active session view are hidden when they cannot be entered.
- Support diagnostics and GitHub issue prefills now include the running release channel information.
- `v1.0.0-beta.1` is positioned as the pre-stable validation release for beta testers rather than another incremental `0.4.x` beta.

### Fixed
- Kernel shutdown now cancels active IPC event streams correctly, preventing the process from lingering after `kernel.json` is removed and avoiding duplicate kernel launches on the next TUI start.
- The TUI now waits for the kernel to actually stop after `K` instead of quitting immediately on the shutdown acknowledgment.
- The updater and release metadata now track beta/stable release kind explicitly instead of inferring everything from the configured update channel.

## [0.4.0-beta.3] - 2026-04-03

### Added
- Dev-only local updater simulation flow, including `F8` local-release preparation and a loopback-served release source for end-to-end self-update testing.
- Dev-only local updater preparation API and local release scanning that can target a specific release directory through `CRONA_DEV_UPDATE_RELEASE_DIR`.

### Changed
- Self-update install now exits the TUI and hands control to the real installer in the terminal instead of rendering a dedicated in-app install takeover screen.
- Installer scripts now support a release-base override so local/dev update simulations can reuse the production installer flow against loopback-hosted release assets.
- Local updater simulations now install into isolated temp runtime/install roots so dev testing does not touch the user’s real Crona install.

### Fixed
- Local update preparation now skips malformed or incomplete release directories instead of selecting unusable entries like `v0.4.0-`.
- TUI event-stream shutdown now uses a shared guarded stop handle, fixing repeated `close of closed channel` panics during update handoff and quit paths.
- Updater/dispatch state was simplified to match the terminal-handoff install model, removing stale install-screen state and dead renderer paths.

## [0.4.0-beta.2] - 2026-04-03

### Added
- Stable/beta update channels, with beta opt-in from settings and update surfaces.
- Destructive runtime management actions in Settings for wiping runtime data and uninstalling Crona.
- Dedicated Support workspace view with GitHub Issues, Discussions, Releases, roadmap links, copied diagnostics, and full redacted support-bundle generation.
- Support bundles now include recent ops plus recent TUI and kernel errors for easier bug reporting.
- Expanded CLI automation for current command families, including context switching helpers, context-based starts, kernel restart/wipe, and broader export/report commands.

### Changed
- Release packaging now publishes compressed platform bundles plus the shared assets archive instead of treating raw binaries as first-class release artifacts.
- Installers now show visible per-download progress and use the platform bundle archives during install.
- Scratchpads no longer use Markdown rendering; they now rely on `$EDITOR` editing and OS-level open behavior.
- GitHub is now the explicit public support surface in-app, with support copy steering users toward Issues, Discussions, Releases, and `docs/roadmap.md`.
- Release docs now match the bundled installer/release layout and the `crona` default launcher flow.

### Fixed
- TUI and kernel dashboard/loading paths now avoid a number of repeated selection, plan, and session aggregation recomputations.
- Core settings reads and update-status persistence now do less redundant work while preserving the same public behavior.
- Support bundle dialogs now carry their rendered state correctly and expose usable follow-up actions.
- Support-focused regression coverage now lives under the TUI testsuite alongside the rest of the behavior-level tests.

## [0.4.0-beta.1] - 2026-04-01

### Added
- Manual session logging from issue-level workflows, including automatic `in_progress` transition when the target issue is focusable but not yet started.
- Dedicated protected-day shell for away mode and configured rest days, with non-work views hidden until the break ends.
- Built-in daily/weekly narrative report presets plus separate HTML/CSS-based PDF rendering for narrative exports.
- Dedicated `Rollup` dashboard with a default weekly range, adjustable start/end dates, calendar-based range picking, and drill-down day details.
- Shared dashboard summary APIs for execution, focus, distribution, goal progress, and estimate-bias rollups.
- Estimate-bias metrics in rollup summaries so over- and under-estimation can be inspected across worked issues.
- Dev-mode seed data now generates a deliberate 7-day scenario window with carry-over, missed days, blocked work, over/under estimates, and varied wellbeing signals.

### Changed
- Settings and Config are now grouped into clearer categories instead of one long flat list.
- Wellbeing now surfaces compact visual cards and a terminal-friendly recent-activity heatmap instead of relying on long text blocks alone.
- Report generation now uses a guided two-step export flow so users choose a category before the concrete report/export type.
- Public release bundles now include the `crona` CLI, and running `crona` with no args now opens the TUI while `crona-tui` remains available for compatibility.
- Dialogs now render their own validation/runtime errors, use consistent `Ctrl+S` submit behavior for forms, and keep field-aware footer hints.
- Self-update now uses a dedicated install takeover screen with phase-based progress and a quiet relaunch handoff.
- Daily and weekly PDF narrative exports now use dedicated HTML/CSS templates instead of reusing the Markdown pipeline.
- Daily dashboard no longer carries opaque weekly glyph summaries; range analytics now live in the dedicated Rollup view.
- Rollup status lines and focus/progress summaries now use stronger color coding for faster at-a-glance reading.

### Fixed
- Session history detail opens correctly again after the session-source additions.
- Away mode toggling now reacts immediately in the TUI and the protected shell updates without waiting for a settings reload.
- Report preset selection and narrative export completion now return from the dialog flow correctly instead of opening empty/stuck states.

## [0.3.1] - 2026-03-27

### Fixed
- Settings view notification toggles now round-trip through the public core-settings API shape correctly after patch and reload.
- Boundary sound and boundary notifications now work independently, so sound-only timer alerts are respected.
- Added regressions covering settings payload decoding and timer-boundary notification dispatch behavior.

## [0.3.0] - 2026-03-26

### Added
- Scriptable `crona` CLI with kernel, context, timer, issue, calendar-export, and dev helper subcommands.
- Shell completion output for `zsh`, `bash`, and `fish`.
- Structured timer-boundary OS notifications with optional audible cues.
- Dedicated configurable ICS export directory for local automation workflows.
- Repo-scoped calendar export that writes stable `issues.ics` and `sessions.ics` bundles.
- Dev-mode binary/runtime split with `crona-dev`, `crona-kernel-dev`, `crona-tui-dev`, and `~/.crona-dev`.
- Dedicated TUI `commands` and `helpers` subpackages plus isolated testsuite support.
- Kernel-owned release update checks with cached release metadata and release notes.
- `crona update notes` plus TUI update-note viewing and dismiss actions.
- Persistent `Updates` workspace view with release notes, install status, and manual-update guidance.

### Changed
- Calendar export no longer inherits the active context/stream scope; it now explicitly targets a repo and defaults from the checked-out repo or repo index `0`.
- TUI calendar export now uses a repo picker and shows written ICS paths on success.
- Reports and calendar exports now use separate output directories, with calendar artifacts excluded from the reports browser.
- The standalone `crona-dev` helper entrypoint was folded into `crona dev ...`.
- Scratchpad rendering now lives under the `views` package with controller logic kept in the app package.
- Phase 5 roadmap planning now includes full CLI CRUD, interactive add/edit flows, and an interactive CLI context picker.
- The roadmap now splits Windows work into Phase 4.1 runtime support and Phase 4.2 installer/self-update support.
- Update prompts now live in the dedicated `Updates` view instead of a temporary header/banner surface, and that view remains accessible even when no update is available.
- Shared update status now carries release-tag and installer/checksum asset metadata for self-update flows.
- Unix runtime data now defaults to native app-data directories instead of `~/.crona`, with `CRONA_HOME` as the explicit override.
- Binary install-location detection is now OS-aware, using `%LocalAppData%\Programs\Crona\bin` as the Windows standard install directory.

### Fixed
- Calendar export now fails clearly when the TUI is talking to a stale kernel that still serves the old response shape.
- Local dev TUI/kernel launch now resolves repo `bin/` binaries correctly instead of requiring them on the shell `PATH`.
- Release installers now stop a running prod kernel before replacing binaries, preventing a newer TUI from attaching to an older still-running kernel.
- Failed update refreshes no longer leave stale cached `update available` prompts visible.
- Prerelease version ordering now follows SemVer rules instead of naive string comparison.
- In-app install is now disabled when Crona is running from a non-standard location, with explicit manual-update guidance shown in the `Updates` view.
- Self-update now validates release installer assets and checksums before replacing binaries and relaunching.
- Unix installs and first kernel start now migrate legacy `~/.crona` runtime data into the new native app-data location automatically.
- Windows releases now ship PowerShell installer assets, `.exe` binaries, and in-app self-update/relaunch support for standard installs.

## [0.2.1] - 2026-03-19

### Added
- Weekly summary, repo, stream, issue-rollup, and CSV exports in the Go kernel and TUI.
- Editable runtime templates for weekly, repo, stream, and issue-rollup narrative reports, with bundled defaults and per-report variable docs.
- Editable JSON CSV export spec plus runtime docs for external-analysis exports.
- Expanded `Config` view asset management for all report templates, docs, and CSV spec files.
- Report browser metadata for report kind, scope, and date-range-aware listing.
- Dedicated kernel and TUI regressions for report asset metadata, Config exposure, and generalized export rendering.
- Report deletion from the `Reports` browser, including removal of sidecar metadata files.

### Changed
- Export assets now use a generalized report-asset model instead of the old daily-only markdown/PDF pair.
- Repo, stream, and issue-rollup reports now include descriptions, issue notes, and per-issue session-note sections.
- Export default output now normalizes legacy `reports/daily` usage back to the shared `reports` root.
- The `Daily Exports` view has been generalized into a broader `Reports` browser in the TUI.
- Bundled report assets are now organized by report kind under `assets/export/{daily,weekly,repo,stream,issue-rollup,csv}`.
- Release/install metadata now targets the `webxsid/crona` repository slug instead of the old `crona-node` slug.
- `make test` now runs `shared`, `kernel`, `tui`, and `cli` tests instead of only the kernel module.
- The `Reports` browser now separates `edit`, `open`, and `delete` actions instead of overloading one open action.

### Fixed
- Daily habit deletion is now exposed from the Daily view action line and dialog flow.
- Repo and stream cascade delete/restore now include habits in addition to issues.
- Habit creation now reuses existing repo/stream selections more reliably by normalizing names and selector inputs.
- TUI Config now visibly lists the generalized report templates/specs instead of showing only the legacy daily export rows.
- Legacy flat user report-template paths now migrate into the nested report-kind asset layout.

## [0.2.0-beta.1] - 2026-03-19

### Added
- Wellbeing tracking flow with daily check-ins for mood, energy, sleep hours, sleep score, screen time, and notes.
- Bubble Tea `Wellbeing` view with per-day check-in details, rolling metrics, streak summaries, and burnout status.
- Habit management across kernel and TUI, including create/edit/delete flows, due-by-date queries, and completion history.
- Daily dashboard habit lane with completion, failure, and optional duration logging.
- Kernel metrics APIs for date-range rollups, burnout indicators, and focus/check-in streak summaries.
- Kernel e2e coverage for daily check-ins, metrics, and persisted sort settings.
- Daily export system with user-editable Handlebars templates, runtime asset management, and generated Markdown reports.
- `Config` view for export templates, template docs, report-directory management, and PDF renderer visibility.
- `Daily Exports` sidebar section and report browser for generated `.md` and `.pdf` files.
- Dedicated PDF export template plus optional PDF file generation through runtime renderer detection.

### Changed
- Daily dashboard now combines planned issues with due habits for the selected date.
- TUI dashboards now use explicit compact layouts at small terminal heights, including minimum-size guarding, wrapped pane hotkeys, and height-aware compact modes for Daily, Default, and Wellbeing views.
- Repo, stream, and issue ordering is now user-configurable through persisted sort settings in core settings.
- Default issue scoping and create/checkout dialogs now prefill from the active repo/stream context when available.
- Roadmap documentation now reflects the implemented Phase 2 check-ins, metrics, and habit work present in the current branch.
- Daily export markdown now uses a glanceable snapshot-first layout with grouped issues/habits, derived highlights/risks, and formatted durations.
- Active-session navigation now keeps session history accessible, scopes that history to the active issue, and reduces the sidebar to session-only views while focused.
- Export configuration now persists a custom reports directory and refreshes export browser state after report generation.

### API / Core
- Added kernel RPC methods for habit CRUD, habit completion/uncompletion/history, daily check-in CRUD/range, and metrics range/rollup/streak queries.
- Added daily check-in persistence plus habit and habit-completion repositories to the SQLite kernel store.
- Added shared domain types and DTOs for habits, check-ins, metrics rollups, streaks, and burnout indicators.
- Added persisted `repoSort`, `streamSort`, and `issueSort` core settings that drive kernel list ordering.
- Added kernel export RPC methods and shared contracts for export assets, template reset by format, report-directory updates, report listing, and format-aware daily export.

### Verification
- `make build` passes for the current workspace.
- `make test` passes for `kernel`.
- `go test ./internal/tui/...` passes for `tui`.

## [0.1.0-beta.2] - 2026-03-14

### Added
- Go monorepo workspace with `kernel`, `tui`, `cli`, and `shared`.
- Go TUI workspace with Default, Meta, Session History, Active Session, Scratchpads, Ops, Settings, and Daily Dashboard views.
- Session-focused workflow from issue panes with auto-context checkout, session lock, stash/end prompts, and scratchpad access during active sessions.
- Session detail overlay in Session History, with richer kernel-backed session context and amend entrypoint.
- Daily Dashboard with date navigation, planned-task list, worked-vs-estimate stats, and resolved-task progress.
- UI-local filtering across repos, streams, issues, scratchpads, and ops.
- Searchable repo and stream selectors in the Default issue-create dialog.
- Optional due date on issue creation, with a calendar picker in the Go TUI dialogs.
- Issue due-date picker action from issue tables/lists, backed by a date-aware todo API.
- Kernel shutdown hotkey from the Go TUI.
- Idle-only stash dialog in the TUI with stash pop/apply.
- Root `.env`-driven runtime mode plus dev-only seed / clear workflows.
- Root `Makefile` and helper scripts for workspace tasks and dev data management.
- Release builder and TUI installer flow for shipping standalone `crona-tui` and `crona-kernel` binaries.
- Go end-to-end tests under `kernel/e2e`.

### Changed
- Repo, stream, and issue public IDs now use numeric IDs.
- The entire local runtime path moved from the old HTTP prototype to Go/Unix socket IPC.
- Kernel auto-launch now prefers an adjacent Go kernel binary and falls back to repo-local `go run` when developing from source.
- Scratchpad reading now stays confined to its pane instead of taking over the full screen.
- Scratchpad editing now opens the real file under the kernel scratch directory, with `.md` fallback when metadata paths omit the extension.
- Scratchpad previews render markdown again after fixing the reload path.
- Pane sizing now uses fixed sidebar/content budgeting and deterministic vertical splits instead of letting overlays and narrow terminals distort row math.
- Default issues are now prioritized by due/open work, split into active vs completed panes, and support direct `1/2`/`tab` section switching.
- Meta issues now show lifecycle status inline.
- Ops moved from a plain list to a table and now load newest-first.
- Ops fetch size is user-adjustable instead of fixed.
- View navigation moved from a top bar to a grouped left sidebar.
- Header was simplified back to a stable context row plus an active-session row.
- Issue lifecycle actions now follow the core transition rules, with one cycle key and explicit abandon behavior.
- Session progress uses cumulative worked time for the active issue based on kernel session history.
- Focus-session start/end now drive issue status transitions through the kernel timer flow.
- Direct issue-status changes are now blocked while the same issue has an active focus session; the end-session transition flow now stops the timer before applying terminal statuses.
- Session amend is now exposed in the TUI as a commit-message rewrite flow for ended sessions.
- Status colors are applied consistently across issue lists and dashboard indicators.
- Release packaging now treats TUI and kernel as independent deliverables instead of bundling them together.

### Fixed
- Footer/status errors now render as transient toast overlays instead of permanently consuming layout space.
- `?` key help now opens as an overlay modal instead of expanding the footer and breaking small-screen layouts.
- Daily and Settings panes no longer overflow unpredictably on small terminals because row-height and list-window calculations now match the rendered layout.
- Session detail and help overlays now match the rest of the pane styling, and session-detail actions stay visible in a fixed footer.
- Dev seed data now follows the current issue lifecycle rules.
- Stash restore no longer intermittently fails with `SQLITE_BUSY` under overlapping local kernel activity.
- Stash apply now fails cleanly while another focus session is active, without mutating context or consuming the stash.
- Focus-session auto-transition to `in_progress` now bypasses the active-session status guard used for manual changes.
- Structured timer boundaries are now recovered when the kernel restarts with an active session still persisted.
- Session timer acceleration caused by overlapping local tick loops.
- Meta issue switching now updates issue context correctly.
- Scratchpad editor saves now reload properly in the Go TUI.
- Go client repo creation now uses the correct kernel route.
- Go client ops loading now uses the kernel's latest-ops endpoint.
- Todo-for-date clearing now actually removes the stored date.
- Issue completion and abandonment timestamps are persisted for dashboard reporting.
- Commit-message dialogs no longer treat typed confirmation characters as submit/cancel.
- Focus-session start no longer races separate issue-status and timer writes in the TUI.

### API / Core
- Added shared Go contracts for domain types, DTOs, and Unix socket IPC envelopes.
- Added daily summary by arbitrary date in the kernel issue summary flow.
- Added kernel shutdown IPC support for TUI-triggered shutdown.
- Added session history and stash IPC consumption in the Go TUI.
- Added kernel session-detail IPC for the Session History overlay.
- Added `kernel.dev.seed` and `kernel.dev.clear` dev-only IPC methods guarded by `CRONA_ENV=Dev`.
- Migrated kernel storage, commands, timer, stash, scratchpad, and settings flows from TypeScript to Go.
- Switched the TUI from HTTP/SSE to Unix socket IPC.

### Verification
- `go build ./...` passes for `shared`, `kernel`, `tui`, and `cli`.
- `go test ./...` passes for `kernel`.
