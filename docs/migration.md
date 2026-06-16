# Migration Guide

Use this guide when switching Crona install methods or release channels.

Use this guide when moving between Homebrew, winget, the legacy install script, or the `crona-beta`/stable channels. Crona's Updates view points here once the migration banner appears.

The detailed guides live here:

- [Legacy to Homebrew](migration/legacy-to-brew.md)
- [Legacy to Go](migration/legacy-to-go.md)
- [Legacy to Winget](migration/legacy-to-winget.md)

The shared migration flow is the same in every destination guide:

1. Stop every running Crona process.
2. Download the latest beta release installer script from GitHub Releases.
3. Make the installer executable and run it, then choose to overwrite the existing install.
4. Run `crona backup`.
5. Remove the runtime directory.
6. Remove the installed binaries.
7. Install the destination package manager version.
8. Run `crona restore <path-to-backup>`.

If you are not sure which destination to choose, start with the package manager you want to keep long term:

- Homebrew on macOS and Linux
- Go source installs when you want to keep using `go install`
- Winget on Windows
