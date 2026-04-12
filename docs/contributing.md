# Contributing

Thanks for contributing to Crona.

## Before You Start

- Read [concepts.md](concepts.md) for the product model.
- Read [development.md](development.md) for build and test workflows.
- Read [api/socket.md](api/socket.md) if your change touches local engine IPC.
- Check [roadmap.md](roadmap.md) before starting larger feature work.

## Workflow Expectations

- Keep changes local-first and local-engine-centric.
- Preserve the current command/repository and engine/client ownership boundaries unless there is a strong reason to change them.
- Prefer small focused refactors over broad rewrites.
- Keep docs and tests in sync with behavioral changes.
- Avoid introducing wrapper-only files or dead abstraction layers.

## Code Quality

- Add or update tests for meaningful behavior changes.
- Keep higher-level TUI behavior checks in the TUI testsuite, and keep pure helper/parser tests local to the owning package.
- Use `make ci` before release-facing changes.
- Use `make test-e2e` when touching local engine IPC startup, shutdown, runtime paths, or protocol behavior.
- Preserve public JSON and IPC contracts unless the change is intentional and documented.
- Prefer clear, low-risk decomposition over framework-style rewrites.

## Pull Request Guidance

- Explain the user-visible outcome first.
- Call out protocol, storage, updater, or installer risk explicitly.
- Note any tests you ran.
- If a change defers or drops an earlier approach, update the docs so the repository does not advertise stale direction.

## Where To Discuss Things

- Bugs: [GitHub Issues](https://github.com/webxsid/crona/issues)
- Questions and ideas: [GitHub Discussions](https://github.com/webxsid/crona/discussions)
