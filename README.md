# Crona

Crona is a local-first work kernel for developers. It combines a local daemon, a terminal UI, and a CLI into one workflow for planning work, tracking focus sessions, and exporting structured artifacts.

The repository is a Go monorepo with four main modules:
- `kernel`: local daemon, SQLite store, timer, IPC, update checks
- `tui`: Bubble Tea terminal UI
- `cli`: scriptable commands and kernel control flows
- `shared`: shared types, config, protocol, and utilities

## Quick Start

See the full installation guide in [docs/install.md](docs/install.md).

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

- `stable` is the preferred channel for general users once stable releases begin.
- `beta` is for pre-release testing and faster iteration.
- `1.0.0` will be the first stable milestone. Until then, published builds are beta-tagged and intended for testers.

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

## Support And Updates

Public support surfaces live on GitHub:

- Bugs: [Issues](https://github.com/webxsid/crona/issues)
- Help and ideas: [Discussions](https://github.com/webxsid/crona/discussions)
- Release updates: [Releases](https://github.com/webxsid/crona/releases)
- Roadmap reading: [docs/roadmap.md](docs/roadmap.md)

Generate a support bundle from the TUI Support view before filing a bug when possible.

## License

[MIT](LICENSE)
