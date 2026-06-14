#!/usr/bin/env bash
# Emit artifact-manifest.json for Quick-Box transfer/activate.
# Usage: QB_VERSION=... QB_ARTIFACT_STAGING=... ./scripts/qb-artifact-manifest.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=qb-artifact-common.sh
source "${SCRIPT_DIR}/qb-artifact-common.sh"

STAGING="${QB_ARTIFACT_STAGING:?QB_ARTIFACT_STAGING is required}"
VERSION="$(qb_resolve_version "${QB_WORKDIR:-}")"
GIT_SHA="${QB_GIT_COMMIT_SHA:-$(git -C "${QB_WORKDIR:-.}" rev-parse HEAD 2>/dev/null || echo unknown)}"
BUILT_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

backend_tar="${STAGING}/backend-${VERSION}.tar.gz"
frontend_tar="${STAGING}/frontend-${VERSION}.tar.gz"

components_json="["
sep=""
if [[ -f "${backend_tar}" ]]; then
  b_sha="$(qb_sha256_file "${backend_tar}")"
  b_size="$(wc -c <"${backend_tar}" | tr -d ' ')"
  components_json+="${sep}{\"name\":\"backend\",\"path\":\"backend-${VERSION}.tar.gz\",\"sha256\":\"${b_sha}\",\"sizeBytes\":${b_size}}"
  sep=","
fi
if [[ -f "${frontend_tar}" ]]; then
  f_sha="$(qb_sha256_file "${frontend_tar}")"
  f_size="$(wc -c <"${frontend_tar}" | tr -d ' ')"
  components_json+="${sep}{\"name\":\"frontend\",\"path\":\"frontend-${VERSION}.tar.gz\",\"sha256\":\"${f_sha}\",\"sizeBytes\":${f_size}}"
fi
components_json+="]"

bundle_sha=""
if command -v python3 >/dev/null 2>&1; then
  bundle_sha="$(python3 - "${STAGING}" <<'PY'
import hashlib, os, sys
staging = sys.argv[1]
h = hashlib.sha256()
for name in sorted(os.listdir(staging)):
    if name == "artifact-manifest.json":
        continue
    path = os.path.join(staging, name)
    if os.path.isfile(path):
        with open(path, "rb") as f:
            h.update(f.read())
print(h.hexdigest())
PY
)"
else
  qb_log_warn "python3 not found; bundleSha256 omitted"
fi

cat <<EOF
{
  "schemaVersion": 1,
  "version": "${VERSION}",
  "gitCommitSha": "${GIT_SHA}",
  "builtAt": "${BUILT_AT}",
  "bundleSha256": "${bundle_sha}",
  "components": ${components_json}
}
EOF
