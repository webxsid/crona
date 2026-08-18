# Crona Docs

Starlight documentation site for Crona. The source of truth is the repository-level `docs/` directory; the build copies and adapts it into Starlight’s generated content directory.

Do not edit `src/content/docs/` directly. It is generated and ignored. Edit the root `docs/` files, then run the sync step.

## Build

```sh
pnpm install --frozen-lockfile
pnpm build
```

The site requires Node.js 22.12 or newer and pnpm 10.30.3 or newer. Every Markdown file under `docs/` must declare `hosted: true` or `hosted: false`; only hosted documents are copied into the site. The build creates the site homepage from the hosted `docs/README.md`, rewrites repository source links to GitHub, runs the frontmatter and Astro checks, and then builds the static site.

For local development:

```sh
pnpm dev
```

Deploy this directory to `docs.crona.work` with Node.js 22.12 or newer.
