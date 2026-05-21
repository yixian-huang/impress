# Ops — Quick-Box deployment

This repo is onboarded to Quick-Box. Environments and deploy hooks live in `ops.zoom.ci`; this file just records the URLs.

## Environments

| Environment | Server | Deploy Hook |
|---|---|---|
| `production` | `172.81.57.29` | `POST https://ops.zoom.ci/api/v1/deploy-hooks/bb47ab5c-1e79-4c96-8a9e-2c719d2698e7/production` |

Both `production` env id and project id are recorded in Quick-Box; nothing to track here besides the URL.

## Trigger a deploy

```bash
curl -X POST -H "X-API-Key: $QB_API_KEY" -H "Content-Type: application/json" \
  -d '{"gitRef":"main"}' \
  https://ops.zoom.ci/api/v1/deploy-hooks/bb47ab5c-1e79-4c96-8a9e-2c719d2698e7/production
```

Then poll:

```bash
curl -H "X-API-Key: $QB_API_KEY" https://ops.zoom.ci/api/v1/deployments/<deploymentId>
curl -H "X-API-Key: $QB_API_KEY" https://ops.zoom.ci/api/v1/deployments/<deploymentId>/logs
```

`status: success` ≠ container healthy — always inspect the `healthcheck` step or SSH and `docker ps`.

## Container model

Single container:
- Stage 1 builds the frontend (`pnpm -C frontend build` → `frontend/out`)
- Stage 2 builds the Go binary (`backend/cmd/server`)
- Stage 3 runtime: Alpine + Go binary + bundled SPA
- Container exposes `:8088`; Go server serves both the API and the SPA via `FRONTEND_DIR`

## Env vars (managed in Quick-Box, not in this repo)

Non-secret: `PORT`, `ENV`, `SEED_MODE`, `FRONTEND_DIR`, `UPLOAD_DIR`
Secret (`{__secretRef}` in Quick-Box): `DB_DSN`, `JWT_SECRET`, `JWT_REFRESH_SECRET`

To view (secrets masked):

```bash
curl -H "X-API-Key: $QB_API_KEY" \
  https://ops.zoom.ci/api/v1/projects/bb47ab5c-1e79-4c96-8a9e-2c719d2698e7/environments/6e38641f-9683-4339-bb7a-2b1a6a7dc1b5/variables
```

## Host-side prerequisites on `172.81.57.29`

- Docker installed with rootful daemon (`extraHosts: host.docker.internal:host-gateway` requires that)
- PostgreSQL listening on `127.0.0.1:54321` (database `blog`, role `blog`)
- Volume mount paths `/home/impress/data` and `/home/impress/uploads` writable by docker daemon
