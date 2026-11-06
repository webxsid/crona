# Security Policy

## Supported Versions

Security fixes land on the latest release (including the current beta line). Please upgrade before reporting.

## Reporting a Vulnerability

Please **do not open a public issue** for security problems.

Report privately through GitHub: go to the
[Security tab](https://github.com/crona-app/crona/security/advisories/new) and choose
**Report a vulnerability**.

Include the Crona version (`crona --version`), your OS and install method (Homebrew, Scoop, Nix, or source), and steps to reproduce.

You can expect an acknowledgement within a few days. Once a fix is released, the advisory will be published with credit unless you prefer otherwise.

## Scope

Crona is local-first: the daemon, TUI, and CLI run on your machine and communicate over local IPC. Issues in that IPC surface, runtime files, exports, or release artifacts are all in scope.
