#!/usr/bin/env bash
# deploy-run.sh — 解析 deploy-workflow.json 执行生产部署
# 文件传输通过 rclone + pCloud 中转（dev→pCloud→prod）
# Usage:
#   ./scripts/deploy-run.sh                  # 默认: backend-only
#   ./scripts/deploy-run.sh backend-only     # 仅后端
#   ./scripts/deploy-run.sh frontend-only    # 仅前端
#   ./scripts/deploy-run.sh full             # 全量部署

set -euo pipefail

# Bypass local proxy (Privoxy) for all deploy network calls
unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY no_proxy NO_PROXY

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKFLOW_FILE="$SCRIPT_DIR/deploy-workflow.json"
CONF_FILE="$ROOT_DIR/.prod_server"

# ── pCloud config ────────────────────────────────────
PCLOUD_REMOTE="pcloud:/deploy/inkless"
PCLOUD_API="https://eapi.pcloud.com"
PCLOUD_USER="${PCLOUD_USER:-}"
PCLOUD_PASS="${PCLOUD_PASS:-}"
PCLOUD_AUTH=""  # filled by pcloud_auth()

# ── Colors ──────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
step_ok()   { echo -e "  ${GREEN}✓${NC} $*"; }
step_fail() { echo -e "  ${RED}✗${NC} $*"; }
step_info() { echo -e "  ${CYAN}→${NC} $*"; }
header()    { echo -e "\n${YELLOW}[$1]${NC} $2"; }

# ── Parse server config ────────────────────────────
parse_server_conf() {
  if [ ! -f "$CONF_FILE" ]; then
    echo -e "${RED}Server config not found: $CONF_FILE${NC}" >&2
    echo -e "${YELLOW}Copy .prod_server.example to .prod_server and set credentials.${NC}" >&2
    exit 1
  fi
  SERVER_IP=$(grep '^ip=' "$CONF_FILE" | cut -d= -f2 | cut -d: -f1)
  SERVER_PORT=$(grep '^ip=' "$CONF_FILE" | cut -d: -f2)
  SERVER_USER=$(grep '^user=' "$CONF_FILE" | cut -d= -f2)
  SERVER_PASS=$(grep '^passwd=' "$CONF_FILE" | cut -d= -f2)
  REMOTE_BASE=$(grep '^path=' "$CONF_FILE" | cut -d= -f2)
  NGINX_PORT=$(grep '^port=' "$CONF_FILE" | cut -d= -f2)
}

# ── SSH helper ──────────────────────────────────────
remote_exec() {
  sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no -p "$SERVER_PORT" \
    "$SERVER_USER@$SERVER_IP" "$@"
}

# ── pCloud helpers ──────────────────────────────────

# Authenticate with pCloud API, store auth token
pcloud_auth() {
  if [[ -z "${PCLOUD_USER}" || -z "${PCLOUD_PASS}" ]]; then
    step_fail "pCloud credentials are required via PCLOUD_USER and PCLOUD_PASS"
    exit 1
  fi
  PCLOUD_AUTH=$(curl -sf "${PCLOUD_API}/userinfo?getauth=1&username=${PCLOUD_USER}&password=${PCLOUD_PASS}" \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['auth'])")
  if [ -z "$PCLOUD_AUTH" ]; then
    step_fail "pCloud 认证失败"; exit 1
  fi
}

# Upload local file to pCloud via rclone
pcloud_upload() {
  local src="$1" dst_name="$2"
  rclone copyto "$src" "${PCLOUD_REMOTE}/${dst_name}" --progress 2>&1 | tail -3
}

# Get direct download URL for a file on pCloud
pcloud_get_link() {
  local filepath="$1"
  local resp
  resp=$(curl -sf "${PCLOUD_API}/getfilelink?auth=${PCLOUD_AUTH}&path=/deploy/inkless/${filepath}")
  local url
  if ! url=$(printf '%s\n' "$resp" | python3 -c "
import sys, json
d = json.load(sys.stdin)
if 'hosts' not in d:
    print('ERROR:' + d.get('error', 'unknown'), file=sys.stderr)
    sys.exit(1)
print('https://' + d['hosts'][0] + d['path'])
"); then
    step_fail "获取 pCloud 下载链接失败"
    exit 1
  fi
  if [[ -z "$url" ]]; then
    step_fail "获取 pCloud 下载链接失败"
    exit 1
  fi
  echo "$url"
}


# ── Parse workflow JSON ────────────────────────────
get_backend_env() {
  python3 -c "
import json
with open('$WORKFLOW_FILE') as f:
    wf = json.load(f)
env = wf['env']['BACKEND_ENV']
for k, v in env.items():
    print(f'{k}={v}')
"
}

get_workflow_jobs() {
  local workflow_name="$1"
  python3 -c "
import json
with open('$WORKFLOW_FILE') as f:
    wf = json.load(f)
jobs = wf['workflows']['$workflow_name']['jobs']
print(' '.join(jobs))
"
}

# ── Job implementations ───────────────────────────

job_lint_check() {
  header "LINT" "代码检查"
  cd "$ROOT_DIR"

  step_info "ESLint..."
  if pnpm lint > /dev/null 2>&1; then
    step_ok "ESLint 通过"
  else
    step_fail "ESLint 失败"; exit 1
  fi

  step_info "TypeScript 类型检查..."
  if pnpm type-check > /dev/null 2>&1; then
    step_ok "类型检查通过"
  else
    step_fail "类型检查失败"; exit 1
  fi
}

job_build_backend() {
  header "BUILD" "构建后端"
  cd "$ROOT_DIR"

  step_info "编译 Go 二进制..."
  (cd backend && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server/)
  LOCAL_MD5=$(md5sum backend/server | awk '{print $1}')
  LOCAL_SIZE=$(du -h backend/server | awk '{print $1}')
  step_ok "编译完成 (${LOCAL_SIZE}, md5: ${LOCAL_MD5:0:8}...)"
}

job_build_frontend() {
  header "BUILD" "构建前端"
  cd "$ROOT_DIR"

  step_info "安装依赖..."
  pnpm install > /dev/null 2>&1
  step_ok "依赖已安装"

  step_info "构建生产包..."
  pnpm build > /dev/null 2>&1
  step_ok "前端构建完成"
}

job_deploy_backend() {
  header "DEPLOY" "部署后端 (via pCloud)"

  step_info "获取 pCloud 认证..."
  pcloud_auth
  step_ok "pCloud 认证成功"

  step_info "上传二进制到 pCloud..."
  pcloud_upload "$ROOT_DIR/backend/server" "server"
  step_ok "上传至 pCloud 完成"

  step_info "生产服务器从 pCloud 下载二进制..."
  remote_exec "
    RESP=\$(curl -sk '${PCLOUD_API}/getfilelink?auth=${PCLOUD_AUTH}&path=/deploy/inkless/server')
    URL=\$(echo \"\$RESP\" | python3 -c \"import sys,json; d=json.load(sys.stdin); print('https://'+d['hosts'][0]+d['path'])\")
    curl -sk -o $REMOTE_BASE/backend/server.new \"\$URL\"
  "
  # Verify downloaded file size
  local remote_size
  remote_size=$(remote_exec "stat -c%s $REMOTE_BASE/backend/server.new 2>/dev/null || echo 0")
  if [ "$remote_size" -lt 1000000 ]; then
    step_fail "下载文件异常 (${remote_size} bytes)"; exit 1
  fi
  step_ok "下载完成 ($(numfmt --to=iec "$remote_size" 2>/dev/null || echo "${remote_size}B"))"

  step_info "替换二进制并重启服务..."
  remote_exec "cd $REMOTE_BASE/backend && mv server.new server && chmod +x server && systemctl restart inkless-backend"
  sleep 3
  step_ok "服务已重启"

  step_info "健康检查..."
  local retries=5
  for i in $(seq 1 $retries); do
    if remote_exec "curl -sf http://127.0.0.1:8088/public/pages > /dev/null" 2>/dev/null; then
      step_ok "健康检查通过"
      return 0
    fi
    [ "$i" -lt "$retries" ] && sleep 2
  done
  step_fail "健康检查失败（请检查 systemctl status inkless-backend）"
  exit 1
}

job_deploy_frontend() {
  header "DEPLOY" "部署前端 (via pCloud)"

  step_info "获取 pCloud 认证..."
  [ -z "$PCLOUD_AUTH" ] && pcloud_auth
  step_ok "pCloud 认证成功"

  step_info "打包前端产物..."
  tar czf /tmp/frontend-out.tar.gz -C "$ROOT_DIR/frontend/out" .
  step_ok "打包完成"

  step_info "上传到 pCloud..."
  pcloud_upload /tmp/frontend-out.tar.gz "frontend-out.tar.gz"
  rm -f /tmp/frontend-out.tar.gz
  step_ok "上传至 pCloud 完成"

  step_info "生产服务器从 pCloud 下载前端包..."
  remote_exec "
    RESP=\$(curl -sk '${PCLOUD_API}/getfilelink?auth=${PCLOUD_AUTH}&path=/deploy/inkless/frontend-out.tar.gz')
    URL=\$(echo \"\$RESP\" | python3 -c \"import sys,json; d=json.load(sys.stdin); print('https://'+d['hosts'][0]+d['path'])\")
    curl -sk -o $REMOTE_BASE/frontend-out.tar.gz \"\$URL\"
  "
  step_ok "下载完成"

  step_info "解压并替换..."
  remote_exec "cd $REMOTE_BASE && rm -rf frontend/* && tar xzf frontend-out.tar.gz -C frontend/ && rm frontend-out.tar.gz" 2>/dev/null
  step_ok "前端文件已同步"

  step_info "重载 Nginx..."
  remote_exec "nginx -t && systemctl reload nginx" 2>/dev/null
  step_ok "Nginx 已重载"

  step_info "重启后端（刷新 index.html 模板缓存）..."
  remote_exec "systemctl restart inkless-backend" 2>/dev/null
  sleep 3
  if remote_exec "curl -sf http://127.0.0.1:8088/public/pages > /dev/null" 2>/dev/null; then
    step_ok "后端已重启"
  else
    step_fail "后端重启后健康检查失败"
    exit 1
  fi
}

# ── Main ──────────────────────────────────────────

main() {
  local workflow="${1:-backend-only}"

  parse_server_conf

  echo -e "${CYAN}══════════════════════════════════════════${NC}"
  echo -e "${CYAN}  Inkless CMS · 生产部署${NC}"
  echo -e "${CYAN}══════════════════════════════════════════${NC}"
  echo -e "  工作流: ${YELLOW}${workflow}${NC}"
  echo -e "  目标:   ${SERVER_USER}@${SERVER_IP} → ${REMOTE_BASE}"
  echo -e "  端口:   Nginx ${NGINX_PORT} → Backend 8088"
  echo -e "${CYAN}══════════════════════════════════════════${NC}"

  # Get job list from workflow
  local jobs
  jobs=$(get_workflow_jobs "$workflow") || {
    echo -e "${RED}Unknown workflow: $workflow${NC}"
    echo "Available: backend-only, frontend-only, full"
    exit 1
  }

  # Execute jobs (respecting simple dependency order from JSON)
  for job in $jobs; do
    local fn="job_${job//-/_}"
    if declare -f "$fn" > /dev/null 2>&1; then
      "$fn"
    else
      echo -e "${RED}Unknown job: $job${NC}"; exit 1
    fi
  done

  echo ""
  echo -e "${GREEN}══════════════════════════════════════════${NC}"
  echo -e "${GREEN}  部署完成!${NC}"
  echo -e "${GREEN}══════════════════════════════════════════${NC}"
}

main "$@"
