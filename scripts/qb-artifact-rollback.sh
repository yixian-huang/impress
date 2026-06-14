#!/usr/bin/env bash
# Quick-Box artifact rollback (runs on deploy server).
# Env: QB_RELEASE_ROOT, COMPONENT (backend|frontend|all), TARGET_VERSION (previous|explicit)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=qb-artifact-common.sh
source "${SCRIPT_DIR}/qb-artifact-common.sh"

RELEASE_ROOT="$(qb_release_root)"
COMPONENT="${COMPONENT:-all}"
TARGET_VERSION="${TARGET_VERSION:-previous}"
FRONTEND_PATH="${RELEASE_ROOT}/frontend"
BACKEND_PATH="${RELEASE_ROOT}/backend"

rollback_frontend() {
  qb_log_info "rollback frontend target=${TARGET_VERSION}"
  local target_path
  if [[ "${TARGET_VERSION}" == "previous" ]]; then
    [[ -L "${FRONTEND_PATH}/previous" ]] || { qb_log_error "no previous frontend"; return 1; }
    target_path="$(readlink "${FRONTEND_PATH}/previous")"
  else
    target_path="${FRONTEND_PATH}/versions/${TARGET_VERSION}"
    [[ -d "${target_path}" ]] || { qb_log_error "frontend version not found: ${TARGET_VERSION}"; return 1; }
  fi
  qb_atomic_symlink "${target_path}" "${FRONTEND_PATH}/current"
}

rollback_backend() {
  qb_log_info "rollback backend target=${TARGET_VERSION}"
  local target_path
  if [[ "${TARGET_VERSION}" == "previous" ]]; then
    [[ -L "${BACKEND_PATH}/previous" ]] || { qb_log_error "no previous backend"; return 1; }
    target_path="$(readlink "${BACKEND_PATH}/previous")"
  else
    target_path="${BACKEND_PATH}/versions/${TARGET_VERSION}"
    [[ -d "${target_path}" ]] || { qb_log_error "backend version not found: ${TARGET_VERSION}"; return 1; }
  fi

  if [[ "$(qb_runtime_type)" == "systemd" ]] && command -v systemctl >/dev/null 2>&1; then
    systemctl stop "$(qb_systemd_unit)" || true
  fi

  qb_atomic_symlink "${target_path}" "${BACKEND_PATH}/current"

  export FRONTEND_DIR="${FRONTEND_DIR:-${RELEASE_ROOT}/frontend/current}"
  qb_write_env_file "${RELEASE_ROOT}/backend/.env" "${RELEASE_ROOT}"
  qb_restart_runtime "${RELEASE_ROOT}"
  qb_health_check
}

case "${COMPONENT}" in
  frontend) rollback_frontend ;;
  backend) rollback_backend ;;
  all)
    rollback_frontend
    rollback_backend
    ;;
  *)
    qb_log_error "invalid COMPONENT=${COMPONENT}"
    exit 1
    ;;
esac

qb_log_info "rollback complete"
