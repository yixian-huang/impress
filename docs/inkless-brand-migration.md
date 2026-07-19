# Inkless Brand Migration

This guide records the supported compatibility surface for the one-time
canonical brand switch to Inkless CMS and the external production steps that
must be executed by an authorized operator.

## Canonical values

| Surface | New value |
| --- | --- |
| Product | Inkless CMS |
| CLI | `inkless` |
| Service | `inkless` |
| API binary | `inkless-api-{version}` and `inkless-api-latest` |
| SQLite database | `inkless.db` |
| PostgreSQL database/user | `inkless` |
| Install root | `/opt/inkless` |
| Domain | `https://inkless.run` |
| Repository | `https://github.com/yixian-huang/inkless` |

## Compatibility matrix

| Legacy surface | Compatibility behavior | New output behavior | Removal gate |
| --- | --- | --- | --- |
| `IMPRESS_ENV_FILE` | Read only when `INKLESS_ENV_FILE` is unset and emit one deprecation warning. | Generated examples and docs use `INKLESS_ENV_FILE`. | Remove after one announced compatibility release. |
| `IMPRESS_SECRET_KEY` | Read only when `INKLESS_SECRET_KEY` is unset and emit one deprecation warning. | Generated examples and docs use `INKLESS_SECRET_KEY`. | Remove after one announced compatibility release. |
| `__IMPRESS_SHARED__` / `__IMPRESS_THEME_REGISTER__` | Runtime aliases keep existing external bundles loadable. | Built-in bundles write only `__INKLESS_SHARED__` and `__INKLESS_THEME_REGISTER__`. | Remove after old plugin fixtures are no longer supported. |
| `impress.setup.step`, `impress.setup.draft`, `impress.comment.guest` | Browser startup migrates valid legacy values to `inkless.*` keys, then deletes the old keys. | New browser sessions write only `inkless.*` keys. | Keep migration until telemetry or support policy confirms old clients have upgraded. |
| `.impress-storage-probe` | Startup may delete only this completed probe file. It must not delete uploads or real user files. | New probes use `.inkless-storage-probe`. | Keep cleanup indefinitely because it is harmless and bounded. |
| `IMPRESS_PLUGIN=impress-cms-v1` | Host preflights compiled binaries and selects this deprecated handshake only when both legacy markers are present and canonical markers are absent. Unknown or broken plugins are started once with the canonical handshake, so runtime errors cannot trigger a second launch. | SDK examples generate `INKLESS_PLUGIN=inkless-cms-v1`. | Remove only after legacy plugin fixture support is explicitly dropped. |
| `X-Impress-*` webhook headers | Transition releases send both old and `X-Inkless-*` headers with matching values. | Documentation and new consumers should use `X-Inkless-Event`, `X-Inkless-Timestamp`, and `X-Inkless-Signature`. | Remove after one announced webhook compatibility window. |
| `impress_` search indexes | The database upgrade writes an explicit `impress_` prefix only into existing Meilisearch records that had no prefix; those records are reindexed before cutover. | New Meilisearch configs default to `inkless_`. | Keep read support until all production indexes have been verified and cut over. |
| `/opt/impress`, `/opt/blotting`, old service units | Treated only as migration sources or rollback references. Scripts do not recursively delete them. | New bootstrap creates `inkless` user, `/opt/inkless`, and `inkless.service`. | Clean up in a separate authorized maintenance task after backup retention expires. |

## Production migration runbook

These steps are intentionally operational instructions, not actions performed by
the repository migration.

1. Freeze deployments and export the current database, PublishedConfig, plugin
   configuration, environment file, Nginx config, systemd unit, and release
   symlink target.
2. Back up uploads and record checksum samples for representative files.
3. Provision `/opt/inkless` with `ops/qb-host-bootstrap.sh` or the equivalent
   commands from `ops/qb-init-hk-artifact.json`.
4. Copy or restore data into `/opt/inkless/data/inkless.db`; do not move or
   delete the old database path during cutover.
5. Deploy Inkless artifacts and install `ops/systemd/inkless.service`.
6. Export the operating site's current PublishedConfig before changing it:

   ```bash
   curl --fail-with-body \
     -H "Authorization: Bearer ${INKLESS_ADMIN_TOKEN}" \
     "${INKLESS_ADMIN_ORIGIN}/admin/global-config" \
     > "published-config-before-$(date -u +%Y%m%dT%H%M%SZ).json"
   ```

   Review `ops/inkless-site-config.example.json`, preserving any intentional
   site-specific fields. Put it as the next draft with the exported
   `draftVersion`, publish it, then GET the config again and verify that
   `publishedVersion` increased by exactly one. Keep both JSON responses as
   rollback inputs. This example targets the operated `inkless.run` site only;
   it is not a global default for user-created sites.
7. Start the service locally and verify `/health`, admin login, content reads,
   uploads, plugins, webhooks, and search before switching public traffic.
8. Point DNS for `inkless.run` and `www.inkless.run` to the production host,
   issue certificates, then install `nginx.conf`.
9. Set `BASE_URL=https://inkless.run` and include the actual frontend origin in
   `CORS_ALLOWED_ORIGINS`.
10. Smoke test HTTPS, canonical redirects, sitemap/RSS/canonical/OG URLs,
   `/admin`, `/api/*`, `/uploads/*`, and client route refreshes.
11. Keep old directories and units available for rollback through at least one
    release observation window.

Rollback restores the saved Nginx config and previous service/symlink/database
backup. Rollback must not delete `/opt/inkless`; leave it available for
diagnosis.
