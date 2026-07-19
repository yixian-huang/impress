# Inkless CMS

Inkless CMS is an extensible content-management platform with a React frontend,
a Go API, configurable site branding, themes, plugins, and deployment support
for SQLite or PostgreSQL.

## Quick start

```bash
pnpm install --frozen-lockfile
pnpm dev
```

Build the API and CLI:

```bash
cd backend
go build ./cmd/server
go build ./cmd/inkless
```

Current documentation lives in [`docs/`](docs/) and the VitePress site in
[`docs-site/`](docs-site/). Deployment and compatibility details for the brand
transition are documented in
[`docs/inkless-brand-migration.md`](docs/inkless-brand-migration.md).

## Community

- Website: <https://inkless.run>
- Source: <https://github.com/yixian-huang/inkless>
- Issues: <https://github.com/yixian-huang/inkless/issues>

Inkless CMS is licensed under the [MIT License](LICENSE).
