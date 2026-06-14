#!/usr/bin/env bash
# One-time host prep for impress artifact deploy on hk (run as root via QB preDeployScript).
set -euo pipefail

RELEASE_ROOT="${QB_RELEASE_ROOT:-/opt/impress}"

mkdir -p "${RELEASE_ROOT}/data" \
         "${RELEASE_ROOT}/uploads" \
         "${RELEASE_ROOT}/backend/versions" \
         "${RELEASE_ROOT}/frontend/versions" \
         /var/lib/quickbox/incoming \
         /var/lib/quickbox/staging

if ! id impress >/dev/null 2>&1; then
  useradd --system --home "${RELEASE_ROOT}" --shell /usr/sbin/nologin impress 2>/dev/null || true
fi

chown -R impress:impress "${RELEASE_ROOT}/data" "${RELEASE_ROOT}/uploads" 2>/dev/null || true

echo "impress host bootstrap ok: ${RELEASE_ROOT}"
