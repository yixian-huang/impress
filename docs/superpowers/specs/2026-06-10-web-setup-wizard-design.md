# Web Setup Wizard — Phase 1 Design

**Date:** 2026-06-10  
**Status:** Phase 1 + Phase 2 shipped (hardened)  
**Scope:** Web wizard for first-run install including bootstrap DB/JWT via browser.

## Goals

- First-time deploys complete initialization in the browser at `/setup`.
- No hardcoded `admin/admin123` on fresh installs (unless `SEED_MODE=demo` for dev).
- Existing deployments with super-admin users continue working unchanged.

## Phase 2 (Bootstrap)

When `SETUP_BOOTSTRAP=true` or JWT env vars are missing:

- Server starts with ephemeral in-memory JWT secrets.
- `GET /setup/status` adds `bootstrapMode`, `needsEnvConfig`, `envFilePath`.
- `POST /setup/test-database` — test SQLite/PostgreSQL connectivity.
- `POST /setup/save-env` — write `.env` with DSN + generated JWT secrets; `restartRequired: true`.
- `POST /setup/complete` blocked until persisted JWT secrets are loaded in-process (`envSecretsLoaded`).
- `POST /setup/test-database` / `save-env` only when `!installed && needsEnvConfig`.
- `Complete` runs in a DB transaction (seed → site name → admin → install record → RBAC).
- Frontend requires DB test before save; polls `/setup/status` after restart.

## Install State

Stored in `site_configs` with key `system`:

```json
{
  "installed": true,
  "installedAt": "2026-06-10T12:00:00Z",
  "seedMode": "blank",
  "installerVersion": "1"
}
```

**`IsInstalled` logic:**

1. `site_configs.system.publishedConfig.installed === true`, or
2. `users` table has at least one `is_super_admin = true` row (backward compat).

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/setup/status` | None | `{ installed, databaseType }` |
| POST | `/setup/complete` | None | One-shot install (rate-limited) |

`POST /setup/complete` body:

```json
{
  "admin": { "username": "admin", "password": "..." },
  "site": {
    "name": { "zh": "我的站点", "en": "My Site" },
    "defaultLocale": "zh"
  },
  "seedMode": "blank"
}
```

Server flow: validate → reject if installed → create super admin → run content seed (no default users) → apply site name → write `system` config → `SeedRBAC`.

## Startup Seed Behavior

| State | `SEED_MODE` | Behavior |
|-------|-------------|----------|
| Installed | any | Existing logic (`demo` default) |
| Not installed | `demo` | Full demo seed (dev) |
| Not installed | `blank` | Blank seed via CLI path |
| Not installed | unset/`none` | Skip seed — await `/setup` |

`make dev` sets `SEED_MODE=demo` for developer convenience.

## Frontend

- Route: `/setup` (standalone page, 4 steps).
- `/admin/login` redirects to `/setup` when not installed.
- `/setup` redirects to `/admin/login` when installed.

## Security

- `/setup/*` uses login-rate-limit preset (5 req/min/IP).
- `POST /setup/complete` disabled after install (`403`).
- Password: min 8 chars, letters + digits.
