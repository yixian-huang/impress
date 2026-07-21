# Runbook：blottingconsultancy.com 升级到最新 Inkless

**目标机**：`47.93.134.202`（NoPanel 别名同 IP）  
**站点**：印迹 / Blotting Consultancy · `blottingconsultancy.com`  
**当前形态**：early impress 手工部署（约 2026-04-20）  
**生产树**：`/home/app/impress`  
**systemd**：`impress-backend.service`  
**端口**：后端 `8088` · nginx 站点 `18090` · OpenResty 对外 80/443  

本 runbook **不**使用 yx.ink / inkless.run 的路径或 unit。  
运维通道优先 **`npc`**（勿裸 SSH 塞密钥）。  
**默认不自动执行生产切换**——完成预检与并行试跑后，需人工确认再切换。

---

## 0. 风险与成功标准（先读）

### 风险摘要

| 风险 | 说明 |
|------|------|
| Schema 向前迁移 | 新二进制启动会跑 AutoMigrate + data migrations + goose；**难以把 schema 退回** 4 月形态 |
| 部署形态旧 | 无 `versions/current/previous` 时，本 runbook **现场搭** 可回滚布局 |
| 静态资源布局 | 新前端可能是 `out/` 结构；nginx 现为 `frontend/assets` + SPA fallback 经后端 |
| 权限 | `editor`/`sa` 已绑 `site_admin`；升级后应仍可看访问统计 |
| 客户感知 | Admin UI / 文案可能带 Inkless 品牌，内容数据应保留在 DB |

### 成功标准（切换后 15 分钟内）

1. `https://blottingconsultancy.com/`（或客户实际入口）首页 200，主题仍为 **corporate-classic** 相关展示正常  
2. `/public/bootstrap` 返回站点品牌与页面列表，无 5xx  
3. 后台可用：`admin` / `editor` 或 `sa` 登录成功  
4. `GET /admin/analytics/summary` 对有权限账号 **200**（非 403/500）  
5. 旧 `/uploads/...` 图片可访问  
6. 出问题可在 **15 分钟内** 执行 §7 回滚到升级前二进制 + DB 备份  

### 维护窗建议

- 低峰 **30–60 分钟**  
- 并行试跑（§5）可在维护窗外做，**不切流量**  
- 切换（§6）建议维护窗内完成  

---

## 1. 环境与约定

### 1.1 路径常量

在**操作员机器**与**目标机**脚本中统一使用：

```bash
# 目标机
export SERVER_REF="47.93.134.202"
export APP_ROOT="/home/app/impress"
export PROD_PORT="8088"
export CANARY_PORT="18088"          # 并行试跑后端端口（勿占用 8088/18090）
export UNIT="impress-backend.service"
export DB_PATH="${APP_ROOT}/data/blotting.db"
export UPLOAD_DIR="${APP_ROOT}/uploads"
export FRONTEND_DIR="${APP_ROOT}/frontend"
export BACKEND_BIN="${APP_ROOT}/backend/server"
export RELEASE_ROOT="${APP_ROOT}/releases"   # 本 runbook 新建
export BACKUP_ROOT="${APP_ROOT}/data/upgrades"
```

### 1.2 版本号

在**本地仓库**（操作员机）取版本：

```bash
cd /path/to/impress   # 本 monorepo
export VERSION="$(git describe --tags --always --dirty)"
export GIT_SHA="$(git rev-parse HEAD)"
echo "VERSION=${VERSION} SHA=${GIT_SHA}"
```

后续所有目录、备份名带同一 `VERSION`。

### 1.3 权限

- 目标机当前进程以 **root** 跑 backend（unit 中 `User=root`）  
- `npc` API key 需具备该 server 的 `commands:execute` 与文件上传权限  

---

## 2. 阶段 A — 备份（生产只读 + 一致性拷贝）

> **切换前必做。** 并行试跑前也建议做一次「热备」；**切换前再做一次「冷备」**（短停或至少 checkpoint）。

### 2.1 远程创建备份目录

```bash
npc server exec command "${SERVER_REF}" --timeout 60 -- "
set -euo pipefail
mkdir -p ${BACKUP_ROOT} ${RELEASE_ROOT}
date -u +%Y-%m-%dT%H:%M:%SZ
df -h ${APP_ROOT}
"
```

### 2.2 热备（不中断服务，适合试跑前）

SQLite 先 checkpoint，再拷贝主库与关键配置：

```bash
npc server exec command "${SERVER_REF}" --timeout 120 -- "
set -euo pipefail
TS=\$(date +%Y%m%d-%H%M%S)
DEST=${BACKUP_ROOT}/hot-\${TS}
mkdir -p \"\${DEST}\"

# checkpoint（可能短暂锁库；通常秒级）
python3 - <<'PY'
import sqlite3
db='${DB_PATH}'
c=sqlite3.connect(db, timeout=60)
c.execute('PRAGMA busy_timeout=60000')
try:
    c.execute('PRAGMA wal_checkpoint(TRUNCATE)')
    print('checkpoint ok')
except Exception as e:
    print('checkpoint warn', e)
c.close()
PY

cp -a ${DB_PATH} \"\${DEST}/blotting.db\"
# 若仍存在 wal/shm，一并保留
cp -a ${DB_PATH}-wal \"\${DEST}/\" 2>/dev/null || true
cp -a ${DB_PATH}-shm \"\${DEST}/\" 2>/dev/null || true

cp -a /etc/systemd/system/${UNIT} \"\${DEST}/impress-backend.service\"
cp -a /etc/nginx/sites-enabled/impress \"\${DEST}/nginx-impress.conf\" 2>/dev/null || true
cp -a ${BACKEND_BIN} \"\${DEST}/server.prod-before\"
# 前端体积大：可 tar 压缩
tar -C ${APP_ROOT} -czf \"\${DEST}/frontend-before.tgz\" frontend
# uploads 按需（大则 rsync 到异地，不强制进同盘）
du -sh \"\${DEST}\" ${UPLOAD_DIR}
echo \"HOT_BACKUP=\${DEST}\"
ls -la \"\${DEST}\"
"
```

记下输出中的 `HOT_BACKUP=...`。

### 2.3 冷备（切换前推荐）

短暂停服务 → checkpoint → 拷贝 → 立刻恢复（若只做备份不升级）：

```bash
npc server exec command "${SERVER_REF}" --timeout 180 -- "
set -euo pipefail
TS=\$(date +%Y%m%d-%H%M%S)
DEST=${BACKUP_ROOT}/cold-\${TS}
mkdir -p \"\${DEST}\"

systemctl stop ${UNIT}
sleep 1
python3 -c \"
import sqlite3
c=sqlite3.connect('${DB_PATH}', timeout=60)
c.execute('PRAGMA wal_checkpoint(TRUNCATE)')
c.close()
print('checkpoint ok')
\"
cp -a ${DB_PATH} \"\${DEST}/blotting.db\"
cp -a ${BACKEND_BIN} \"\${DEST}/server.prod-before\"
cp -a /etc/systemd/system/${UNIT} \"\${DEST}/\"
tar -C ${APP_ROOT} -czf \"\${DEST}/frontend-before.tgz\" frontend
# 可选：整库目录
cp -a ${APP_ROOT}/data \"\${DEST}/data-dir\" 2>/dev/null || true

systemctl start ${UNIT}
sleep 2
systemctl is-active ${UNIT}
curl -sf -o /dev/null -w 'health=%{http_code}\n' http://127.0.0.1:${PROD_PORT}/health

echo \"COLD_BACKUP=\${DEST}\"
ls -la \"\${DEST}\"
"
```

### 2.4 可选：拉备份到操作员机

```bash
# 示例：只拉 DB + unit（体积可控）
npc server file pull "${SERVER_REF}" \
  "${BACKUP_ROOT}/cold-YYYYMMDD-HHMMSS/blotting.db" \
  "./backups/blotting-cold.db"
```

---

## 3. 阶段 B — 预检（只读）

### 3.1 生产现状快照

```bash
npc server exec command "${SERVER_REF}" --timeout 60 -- "
set -euo pipefail
echo '=== process ==='
systemctl is-active ${UNIT} || true
ss -lntp | grep -E '8088|18090|18088' || true
echo '=== binary ==='
ls -la ${BACKEND_BIN}
stat ${BACKEND_BIN} | sed -n '1,8p'
echo '=== env (secrets redacted) ==='
# 从 unit 读非密钥字段
grep -E '^Environment=' /etc/systemd/system/${UNIT} | sed -E 's/(SECRET|PASSWORD|TOKEN)=.*/\1=***/I'
echo '=== db size ==='
ls -la ${APP_ROOT}/data/
echo '=== bootstrap theme ==='
curl -sf http://127.0.0.1:${PROD_PORT}/public/bootstrap | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d.get(\"activeTheme\"));print(\"pages\",len(d.get(\"themePages\")or[]))'
"
```

### 3.2 遗留多站点 schema 预检（推荐在**备份库副本**上跑）

在**操作员机**用当前仓库代码（需 Go）：

```bash
cd backend
# 将冷备 DB 拷到本地后：
export CHECK_DSN="file:$(pwd)/../backups/blotting-cold.db?mode=ro"
# 若 CLI 支持 DSN 环境变量，以实际 inkless CLI 为准；否则：
go run ./cmd/inkless migrate legacy-site-status --json
# 或：
go run ./cmd/inkless migrate legacy-site-status
```

解读：

- `sites` / `site_users` 行数为 0 → 多站点历史数据风险低  
- `user_roles.site_id` 非空行：本机升级前已用 `site_id NULL` 绑 `site_admin`，应正常  
- **预检命令不得修改库**

### 3.3 RBAC 与统计（生产只读）

```bash
npc server file upload "${SERVER_REF}" /tmp/impress-rbac-ro.py /tmp/impress-rbac-ro.py 2>/dev/null || true

npc server exec command "${SERVER_REF}" --timeout 45 -- "
python3 - <<'PY'
import sqlite3
c=sqlite3.connect('${DB_PATH}', timeout=10)
c.row_factory=sqlite3.Row
print('users', [dict(r) for r in c.execute('select id,username,role,is_super_admin from users')])
print('user_roles', [dict(r) for r in c.execute('''
  select u.username, r.name role, ur.site_id
  from user_roles ur join users u on u.id=ur.user_id join roles r on r.id=ur.role_id
''')])
print('pv', c.execute('select count(*), min(viewed_at), max(viewed_at) from page_views').fetchone())
PY
"
```

期望：`editor` / `sa` 已有 `site_admin`；`page_views` 有数据。

### 3.4 磁盘与依赖

```bash
npc server exec command "${SERVER_REF}" --timeout 30 -- "
df -h / /home
free -h
python3 --version
# 运行新二进制一般只需 glibc；构建应在操作员机完成
ldd --version | head -1
"
```

建议 `/` 或数据盘 **至少预留 2GB+**（备份 + 双份前端 + canary 库）。

**Go / 预检门槛**：`legacy-site-status` 非 0 退出或 `hasLegacyData=true` 且有大量非空站点行 → **停止**，先导出审计再继续。

---

## 4. 阶段 C — 本地构建产物

在**操作员机** monorepo 根目录（`pnpm` + 匹配 `go.mod` 的 Go）：

```bash
cd /path/to/impress
git checkout main
git pull --ff-only
export VERSION="$(git describe --tags --always)"
export GIT_SHA="$(git rev-parse HEAD)"

# 前端
pnpm install
pnpm -C frontend build
# 产物默认：frontend/out/

# 后端 linux/amd64（目标机为 x86_64）
mkdir -p "artifacts/${VERSION}"
(
  cd backend
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_SHA}" \
    -o "../artifacts/${VERSION}/server" \
    ./cmd/server/
)
# 若项目惯用 scripts：
# VERSION="$VERSION" ./scripts/build-backend.sh
# VERSION="$VERSION" ./scripts/build-frontend.sh

# 打包前端（保持 out 内结构）
tar -C frontend/out -czf "artifacts/${VERSION}/frontend-out.tgz" .

ls -la "artifacts/${VERSION}/"
file "artifacts/${VERSION}/server"
```

**构建门禁（建议）**：

```bash
pnpm lint && pnpm type-check
cd backend && go test ./internal/db/... ./internal/handler/analytics/... -count=1
```

---

## 5. 阶段 D — 并行端口试跑（不切公网流量）

思路：用**生产 DB 的拷贝** + **新二进制** 在 `18088` 启动；公网仍走 `8088`/`18090`。

### 5.1 上传产物

```bash
npc server exec command "${SERVER_REF}" --timeout 30 -- "
mkdir -p ${RELEASE_ROOT}/${VERSION}/{backend,frontend,data}
"

npc server file upload "${SERVER_REF}" \
  "artifacts/${VERSION}/server" \
  "${RELEASE_ROOT}/${VERSION}/backend/server"

npc server file upload "${SERVER_REF}" \
  "artifacts/${VERSION}/frontend-out.tgz" \
  "${RELEASE_ROOT}/${VERSION}/frontend-out.tgz"

npc server exec command "${SERVER_REF}" --timeout 120 -- "
set -euo pipefail
chmod +x ${RELEASE_ROOT}/${VERSION}/backend/server
mkdir -p ${RELEASE_ROOT}/${VERSION}/frontend
tar -C ${RELEASE_ROOT}/${VERSION}/frontend -xzf ${RELEASE_ROOT}/${VERSION}/frontend-out.tgz
# 规范：canary 的 FRONTEND_DIR 指向解压根（含 index.html 与 assets/）
ls -la ${RELEASE_ROOT}/${VERSION}/frontend | head
test -f ${RELEASE_ROOT}/${VERSION}/frontend/index.html
"
```

### 5.2 准备 canary 数据（拷贝库，隔离写入）

```bash
npc server exec command "${SERVER_REF}" --timeout 120 -- "
set -euo pipefail
# 再 checkpoint 后拷到 release data
python3 -c \"
import sqlite3
c=sqlite3.connect('${DB_PATH}', timeout=60)
c.execute('PRAGMA wal_checkpoint(TRUNCATE)')
c.close()
\"
cp -a ${DB_PATH} ${RELEASE_ROOT}/${VERSION}/data/blotting-canary.db
# canary 使用独立 uploads 只读挂载或共享只读：默认共享生产 uploads（只读风险：canary 若写媒体会进生产盘）
# 建议 canary 写媒体指向隔离目录：
mkdir -p ${RELEASE_ROOT}/${VERSION}/uploads-canary
# 可选：不复制全量 uploads，仅验证 DB + 静态
echo CANARY_DB=${RELEASE_ROOT}/${VERSION}/data/blotting-canary.db
"
```

> **注意**：canary 若 `UPLOAD_DIR` 指生产 `uploads`，管理端上传会污染生产文件。试跑验收以**读**为主；需要写测时用 `uploads-canary`。

### 5.3 启动 canary 进程（前台 nohup / 临时 unit）

**方式 A — nohup（简单）**：

```bash
npc server exec command "${SERVER_REF}" --timeout 30 -- "
set -euo pipefail
# 从生产 unit 读取 JWT（勿打印到日志聚合；此处仅进程环境）
# 若不便解析，请手工在下面 export 与生产一致的 JWT（操作员本地保管，不要提交 git）

# 停止旧 canary
if [ -f /tmp/impress-canary.pid ]; then kill \$(cat /tmp/impress-canary.pid) 2>/dev/null || true; fi

# 从运行中的生产进程继承非密钥环境较难；推荐显式写 canary env 文件（权限 600）
# 若已有 /home/app/impress/releases/canary.env 则复用：
test -f ${RELEASE_ROOT}/canary.env || {
  echo 'Create canary.env first (see below)' >&2
  exit 1
}

set -a
# shellcheck disable=SC1091
source ${RELEASE_ROOT}/canary.env
set +a

nohup ${RELEASE_ROOT}/${VERSION}/backend/server \
  > ${RELEASE_ROOT}/${VERSION}/canary.log 2>&1 &
echo \$! > /tmp/impress-canary.pid
sleep 2
curl -sf -o /dev/null -w 'canary_health=%{http_code}\n' http://127.0.0.1:${CANARY_PORT}/health || {
  echo 'canary failed'; tail log:' >&2
  tail -50 ${RELEASE_ROOT}/${VERSION}/canary.log >&2
  exit 1
}
echo CANARY_PID=\$(cat /tmp/impress-canary.pid)
"
```

**首次创建 `canary.env`（在目标机，0600，JWT 与生产 unit 一致）**：

```bash
npc server exec command "${SERVER_REF}" --timeout 30 -- "
set -euo pipefail
# 从生产 unit 提取（注意：会短暂出现在会话审计中；完成后可改用 secret 文件）
grep '^Environment=' /etc/systemd/system/${UNIT} | sed 's/^Environment=//' > /tmp/prod.env.raw
# 人工确认后生成 canary.env：
# 注意：DB_DSN 禁止裸写 &mode=rwc —— bash source 会把 & 拆成后台任务，导致 DB_DSN 丢失、连到空库。
# 推荐：去掉 & 参数，或整值单引号包裹。
cat > ${RELEASE_ROOT}/canary.env <<EOF
PORT=${CANARY_PORT}
ENV=production
DB_DSN=file:${RELEASE_ROOT}/${VERSION}/data/blotting-canary.db?cache=shared
JWT_SECRET=<<<从生产 unit 复制>>>
JWT_REFRESH_SECRET=<<<从生产 unit 复制>>>
UPLOAD_DIR=${RELEASE_ROOT}/${VERSION}/uploads-canary
FRONTEND_DIR=${RELEASE_ROOT}/${VERSION}/frontend
EOF
# 将 <<<>>> 替换为真实值后：
chmod 600 ${RELEASE_ROOT}/canary.env
"
```

启动后立刻核对进程环境（必须有完整 `DB_DSN=file:...`）：

```bash
tr '\\0' '\\n' < /proc/\$(cat /tmp/impress-canary.pid)/environ | grep '^DB_DSN='
```

更稳妥：操作员在本地写好 `canary.env`（不含提交），再：

```bash
npc server file upload "${SERVER_REF}" ./canary.env "${RELEASE_ROOT}/canary.env"
npc server exec command "${SERVER_REF}" -- -- "chmod 600 ${RELEASE_ROOT}/canary.env"
```

### 5.4 Canary 验收清单

在目标机本机 curl（或 `npc exec`）：

```bash
npc server exec command "${SERVER_REF}" --timeout 60 -- "
set -euo pipefail
B=http://127.0.0.1:${CANARY_PORT}
echo health: \$(curl -sf -o /dev/null -w '%{http_code}' \$B/health)
echo bootstrap:
curl -sf \$B/public/bootstrap | python3 -c '
import sys,json
d=json.load(sys.stdin)
t=d.get(\"activeTheme\") or {}
print(\"theme\", t.get(\"themeId\"), t.get(\"source\"))
print(\"pages\", len(d.get(\"themePages\") or []))
g=(d.get(\"globalConfig\") or {}).get(\"config\") or {}
brand=((g.get(\"branding\") or {}).get(\"companyName\") or {})
print(\"brand_zh\", brand.get(\"zh\"))
'
# 未登录 analytics 应为 401（证明路由在）
echo analytics_noauth: \$(curl -s -o /dev/null -w '%{http_code}' \$B/admin/analytics/summary)
# 首页 HTML
echo home: \$(curl -sf -o /dev/null -w '%{http_code}' \$B/)
# 迁移是否写了 canary 库
ls -la ${RELEASE_ROOT}/${VERSION}/data/
tail -30 ${RELEASE_ROOT}/${VERSION}/canary.log
"
```

**人工浏览器试跑（可选）**：SSH 隧道

```bash
# 操作员机
ssh -L 18088:127.0.0.1:18088 root@47.93.134.202
# 浏览器打开 http://127.0.0.1:18088/ 与 /admin/login
```

验收打勾：

- [ ] health 200  
- [ ] bootstrap 主题 `corporate-classic`（或等价，非空白）  
- [ ] 品牌中文仍为印迹相关  
- [ ] 主要栏目页 200  
- [ ] 登录 admin / editor  
- [ ] 访问统计 200（editor/sa）  
- [ ] canary 日志无 migration panic  

**失败**：停 canary，**不要**进入 §6。

```bash
npc server exec command "${SERVER_REF}" --timeout 20 -- "
kill \$(cat /tmp/impress-canary.pid) 2>/dev/null || true
rm -f /tmp/impress-canary.pid
"
```

---

## 6. 阶段 E — 生产切换

> 仅在 canary 验收通过 + 冷备完成 + 维护窗内执行。

### 6.1 切换策略（本机推荐）

保留旧二进制/前端为 `previous`，新版本进 `releases/$VERSION`，再原子切换目录内容或 symlink。

当前 nginx 写死：

- `alias /home/app/impress/frontend/assets/`  
- `proxy_pass http://127.0.0.1:8088`  
- 后端 `FRONTEND_DIR=/home/app/impress/frontend`

故切换时：**保持路径不变**，用 `releases` 作暂存 + `previous` 备份拷贝。

### 6.2 切换脚本（生产）

```bash
npc server exec command "${SERVER_REF}" --timeout 180 -- "
set -euo pipefail
VERSION='${VERSION}'   # 由操作员展开为实际版本号
APP_ROOT=/home/app/impress
UNIT=impress-backend.service
RELEASE_ROOT=\${APP_ROOT}/releases
PREV=\${RELEASE_ROOT}/previous-prod
NEW=\${RELEASE_ROOT}/\${VERSION}

test -x \${NEW}/backend/server
test -f \${NEW}/frontend/index.html

# 停 canary
if [ -f /tmp/impress-canary.pid ]; then kill \$(cat /tmp/impress-canary.pid) 2>/dev/null || true; fi

# 停生产
systemctl stop \${UNIT}

# 冷备 previous（目录）
rm -rf \${PREV}
mkdir -p \${PREV}
cp -a \${APP_ROOT}/backend/server \${PREV}/server
tar -C \${APP_ROOT} -czf \${PREV}/frontend.tgz frontend
python3 -c \"
import sqlite3
c=sqlite3.connect('\${APP_ROOT}/data/blotting.db', timeout=60)
c.execute('PRAGMA wal_checkpoint(TRUNCATE)')
c.close()
\"
cp -a \${APP_ROOT}/data/blotting.db \${PREV}/blotting.db

# 换二进制
cp -a \${NEW}/backend/server \${APP_ROOT}/backend/server
chmod +x \${APP_ROOT}/backend/server

# 换前端：先备份现网目录名，再 rsync 新 out
rm -rf \${APP_ROOT}/frontend.prev-dir
mv \${APP_ROOT}/frontend \${APP_ROOT}/frontend.prev-dir
mkdir -p \${APP_ROOT}/frontend
# 若新产物根即含 assets/ + index.html：
cp -a \${NEW}/frontend/. \${APP_ROOT}/frontend/
# nginx alias 需要 assets/；确认：
test -d \${APP_ROOT}/frontend/assets
test -f \${APP_ROOT}/frontend/index.html

# 启动（DB_DSN 仍指向生产 blotting.db → 首次启动跑 migration）
systemctl start \${UNIT}
sleep 3
systemctl is-active \${UNIT}
curl -sf -o /dev/null -w 'health=%{http_code}\n' http://127.0.0.1:8088/health
curl -sf -o /dev/null -w 'bootstrap=%{http_code}\n' http://127.0.0.1:8088/public/bootstrap
echo SWITCH_OK version=\${VERSION}
"
```

**注意**：上面脚本里的 `VERSION='${VERSION}'` 在 `npc` 远端单引号环境中**不会**自动替换。请改为实际值，例如 `VERSION='v0.x.x-githash'`，或在本地生成完整脚本再上传执行。

### 6.3 切换后验收（生产）

```bash
npc server exec command "${SERVER_REF}" --timeout 60 -- "
set -euo pipefail
B=http://127.0.0.1:8088
curl -sf \$B/health
curl -sf \$B/public/bootstrap | python3 -c '
import sys,json
d=json.load(sys.stdin)
print(d.get(\"activeTheme\"))
print(\"brand\", (((d.get(\"globalConfig\") or {}).get(\"config\") or {}).get(\"branding\") or {}).get(\"companyName\"))
'
# nginx 层
curl -sf -o /dev/null -w 'nginx_local=%{http_code}\n' -H 'Host: blottingconsultancy.com' http://127.0.0.1:18090/ || true
# 日志
tail -40 /home/app/impress/logs/backend.log
journalctl -u impress-backend.service -n 30 --no-pager
"
```

公网：

- 打开 `https://blottingconsultancy.com/`  
- `/admin/login` → 访问统计  
- 抽查 2–3 个栏目与一张旧图  

### 6.4 切换失败立刻回滚

见 §7，**不要**在 migration 中途反复 restart 指望自愈而不看日志。

---

## 7. 阶段 F — 回滚

### 7.1 应用回滚（二进制 + 前端，保留**升级后** DB —— 仅当新 schema 仍兼容旧二进制时）

多数情况下 **旧 4 月二进制无法读懂新 migration 后的库**。  
因此 **标准回滚 = previous 二进制/前端 + previous（升级前）DB**。

```bash
npc server exec command "${SERVER_REF}" --timeout 180 -- "
set -euo pipefail
APP_ROOT=/home/app/impress
UNIT=impress-backend.service
PREV=\${APP_ROOT}/releases/previous-prod

test -f \${PREV}/server
test -f \${PREV}/blotting.db
test -f \${PREV}/frontend.tgz

systemctl stop \${UNIT}

cp -a \${PREV}/server \${APP_ROOT}/backend/server
chmod +x \${APP_ROOT}/backend/server

rm -rf \${APP_ROOT}/frontend
mkdir -p \${APP_ROOT}/frontend
tar -C \${APP_ROOT}/frontend -xzf \${PREV}/frontend.tgz
# 若 tar 内顶层是 frontend/ 目录：
if [ ! -f \${APP_ROOT}/frontend/index.html ] && [ -f \${APP_ROOT}/frontend/frontend/index.html ]; then
  mv \${APP_ROOT}/frontend/frontend/* \${APP_ROOT}/frontend/ && rmdir \${APP_ROOT}/frontend/frontend || true
fi

# 恢复升级前 DB（会丢失升级后写入的 PV/内容变更）
cp -a \${PREV}/blotting.db \${APP_ROOT}/data/blotting.db
rm -f \${APP_ROOT}/data/blotting.db-wal \${APP_ROOT}/data/blotting.db-shm

systemctl start \${UNIT}
sleep 2
systemctl is-active \${UNIT}
curl -sf -o /dev/null -w 'health=%{http_code}\n' http://127.0.0.1:8088/health
echo ROLLBACK_OK
"
```

若切换时保留了 `frontend.prev-dir`：

```bash
# 可选：用目录级恢复前端
# rm -rf frontend && mv frontend.prev-dir frontend
```

### 7.2 从冷备目录回滚

```bash
# 将 COLD_BACKUP 换成实际路径
npc server exec command "${SERVER_REF}" --timeout 180 -- "
set -euo pipefail
COLD=/home/app/impress/data/upgrades/cold-YYYYMMDD-HHMMSS
systemctl stop impress-backend.service
cp -a \${COLD}/server.prod-before /home/app/impress/backend/server
cp -a \${COLD}/blotting.db /home/app/impress/data/blotting.db
rm -f /home/app/impress/data/blotting.db-wal /home/app/impress/data/blotting.db-shm
# 前端
rm -rf /home/app/impress/frontend
mkdir -p /home/app/impress/frontend
tar -C /home/app/impress -xzf \${COLD}/frontend-before.tgz
systemctl start impress-backend.service
curl -sf http://127.0.0.1:8088/health
"
```

### 7.3 回滚后验证

- health / bootstrap / 首页  
- admin 登录  
- 确认 `page_views` 时间戳停在备份点附近（预期）  

---

## 8. 切换后观察（24h）

```bash
# 错误与 5xx
npc server exec command "${SERVER_REF}" --timeout 30 -- "
grep -E 'ERROR|panic|FATAL' /home/app/impress/logs/backend.log | tail -30
grep analytics/summary /home/app/impress/logs/backend.log | tail -15
df -h /home
"
```

关注：

- analytics 是否再出现大量 403/500  
- 磁盘（备份 + releases 是否占满）  
- 客户反馈导航/表单  

清理（稳定 7 天后）：旧 `releases/*` canary 库、过期 hot 备份；**至少保留 1 份 cold 备份异地**。

---

## 9. 一页检查表（执行时勾选）

| # | 步骤 | 负责人 | 完成 |
|---|------|--------|------|
| 1 | 约定 VERSION / 维护窗 | | ☐ |
| 2 | 热备 | | ☐ |
| 3 | 预检（bootstrap / RBAC / disk / legacy-site-status） | | ☐ |
| 4 | 本地 build linux/amd64 + frontend out | | ☐ |
| 5 | 上传 releases/\$VERSION | | ☐ |
| 6 | canary.env（JWT 与生产一致，PORT=18088，独立 DB） | | ☐ |
| 7 | canary 验收通过 | | ☐ |
| 8 | 冷备 | | ☐ |
| 9 | 生产切换 | | ☐ |
| 10 | 生产验收（公网+统计+上传） | | ☐ |
| 11 | 失败则 §7 回滚 | | ☐ |
| 12 | 停 canary / 记变更 | | ☐ |

---

## 10. 明确不做的事

- 不在本机跑 `yx.ink` / `inkless-ops` 的 unit 或 `/opt/inkless*` 路径  
- 不改生产 `JWT_SECRET`（除非同时接受全员重新登录）  
- 不用 canary 库覆盖生产库  
- 不在未备份时 `DROP` 任何 `sites` / 遗留列  
- 不把 JWT 写进 git 或聊天明文长期存档  

---

## 11. 附录：当前生产 unit 参考（升级时保持 DB/JWT/路径）

切换**默认保留**现有 unit，仅换二进制与前端目录内容：

```ini
[Service]
WorkingDirectory=/home/app/impress/backend
Environment=PORT=8088
Environment=DB_DSN=file:/home/app/impress/data/blotting.db?cache=shared&mode=rwc
Environment=ENV=production
Environment=UPLOAD_DIR=/home/app/impress/uploads
Environment=FRONTEND_DIR=/home/app/impress/frontend
Environment=JWT_SECRET=***
Environment=JWT_REFRESH_SECRET=***
ExecStart=/home/app/impress/backend/server
```

nginx `sites-enabled/impress` 监听 `18090`，静态 `assets/` + 反代 `8088`。若新前端 `assets` 哈希文件名变化，只需确保 `index.html` 与 `assets/` 同次发布即可。

---

## 12. 附录：与 artifact 流水线的关系

长期可将该机迁到 `/opt/inkless` + `qb-artifact-activate.sh` 布局；**本次 runbook 刻意兼容现网 `/home/app/impress`**，降低一次改造两件事（产品升级 + 部署现代化）的耦合风险。  
部署现代化可作为升级稳定 1–2 周后的独立变更。

## 13. 路径 B：corporate-classic 外置包（同车发布）

主题已抽到 monorepo 包 **`@inkless/theme-corporate-classic`**（`packages/theme-corporate-classic`），**theme id 仍为 `corporate-classic`**，DB 无需迁移。

| 项 | 说明 |
|----|------|
| Host 接线 | `frontend/src/plugins/themes/corporate-classic/index.ts` 调用 `createCorporateClassicTheme({ loaders })`，页面组件仍在 `frontend/src/pages/*` |
| 依赖 | `frontend/package.json` → `"@inkless/theme-corporate-classic": "workspace:*"` |
| theme-host 增量 | `useHeaderScroll` / `useThemePages` / `CORPORATE_DEFAULT_LAYOUT` 已导出 |

**Canary 重建**（路径 B 与 host 同 artifact）：

```bash
# 操作员机 monorepo（含 packages/theme-corporate-classic）
pnpm install
pnpm -C frontend build
# Docker CGO 构建 backend（同 §4）
# 上传 releases/$VERSION 后重启 canary（同 §5）
# 验收：bootstrap themeId=corporate-classic + 7 页 200
```

独立 GitHub 仓（对齐 blog-first）可后续从 `packages/theme-corporate-classic` 迁出；**不阻塞** 本机升级。

---

**文档版本**：2026-07-21  
**适用**：`47.93.134.202` / `blottingconsultancy.com` / `/home/app/impress`  
**相关**：`docs/upgrading-single-site-convergence.md` · `docs/ops-lessons-yx-ink-vs-inkless-run.md` · `OPS.md`
