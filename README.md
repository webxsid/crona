# Crona

Crona is a local-first work kernel for developers. It combines a local daemon, a terminal UI, and a CLI into one workflow for planning work, tracking focus sessions, and exporting structured artifacts.

The repository is a Go monorepo with four main modules:
- `kernel`: local daemon, SQLite store, timer, IPC, update checks
- `tui`: Bubble Tea terminal UI
- `cli`: scriptable commands and kernel control flows
- `shared`: shared types, config, protocol, and utilities

## Quick Start

See the full installation guide in [docs/install.md](docs/install.md).

Runtime notes:
- local alerts are emitted by the kernel, not the TUI process
- scheduled reminders only fire while the local kernel is running
- the TUI owns the terminal tab title while it is running and shows active session context when focused
- PDF export depends on local renderer tooling; see [docs/install.md](docs/install.md)

Launch the TUI:

```bash
crona
```

Inspect the local kernel from the CLI:

```bash
crona kernel attach --json
crona kernel status --json
crona kernel info --json
```

Generate shell completions:

```bash
crona completion zsh
crona completion bash
crona completion fish
```

## Release Channels

- `stable` is the preferred channel for general users.
- `beta` is for pre-release testing and faster iteration.
- `v1.0.0-beta.3` is the current `1.0.0` prerelease build for tester validation before the first stable release.

## Documentation

- [Docs Index](docs/README.md)
- [Concepts](docs/concepts.md)
- [Install](docs/install.md)
- [Development](docs/development.md)
- [Contributing](docs/contributing.md)
- [Socket API](docs/api/socket.md)
- [Roadmap](docs/roadmap.md)
- [Changelog](docs/changelog.md)
- [Feature Design](docs/feature-design.md)

Operational references:
- [Notification and alert behavior](docs/install.md#notifications-and-alerts)
- [PDF rendering support](docs/install.md#pdf-rendering)

## Support And Updates

Public support surfaces live on GitHub:

- Bugs: [Issues](https://github.com/webxsid/crona/issues)
- Help and ideas: [Discussions](https://github.com/webxsid/crona/discussions)
- Release updates: [Releases](https://github.com/webxsid/crona/releases)
- Roadmap reading: [docs/roadmap.md](docs/roadmap.md)

Generate a support bundle from the TUI Support view before filing a bug when possible.

## License

[MIT](LICENSE)
