#!/usr/bin/env bash
# Quick-Box script deploy for inkless on quickboxd-managed servers.
# Expects: repo checkout at QB_WORKDIR (default /home/inkless), docker available.
set -euo pipefail

WORKDIR="${QB_WORKDIR:-/home/inkless}"
REPO_URL="${QB_REPO_URL:-https://github.com/yixian-huang/inkless.git}"
GIT_REF="${GIT_REF:-${QB_GIT_REF:-main}}"
IMAGE="${QB_IMAGE:-inkless:latest}"
CONTAINER="${QB_CONTAINER:-inkless}"

mkdir -p /home/inkless/data /home/inkless/uploads

if [[ ! -d "${WORKDIR}/.git" ]]; then
  rm -rf "${WORKDIR}"
  git clone "${REPO_URL}" "${WORKDIR}"
fi

cd "${WORKDIR}"
git fetch --all --prune
git reset --hard "origin/${GIT_REF}"

docker build -f Dockerfile -t "${IMAGE}" .

docker stop "${CONTAINER}" 2>/dev/null || true
docker rm "${CONTAINER}" 2>/dev/null || true

docker run -d --name "${CONTAINER}" \
  --restart unless-stopped \
  -p 8088:8088 \
  -v /home/inkless/data:/app/data \
  -v /home/inkless/uploads:/app/uploads \
  --add-host host.docker.internal:host-gateway \
  -e PORT=8088 \
  -e ENV=production \
  -e SEED_MODE=blank \
  -e SETUP_BOOTSTRAP=true \
  -e FRONTEND_DIR=/app/frontend/out \
  -e UPLOAD_DIR=/app/uploads \
  -e "DB_DSN=file:/app/data/inkless.db?cache=shared&mode=rwc" \
  "${IMAGE}"

curl -sf "http://127.0.0.1:8088/health" >/dev/null
echo "inkless healthy on :8088"
