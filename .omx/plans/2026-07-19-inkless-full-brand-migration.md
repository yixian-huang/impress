# Inkless 全量品牌迁移任务

> 日期：2026-07-19
> 目标版本：一次发布完成 Canonical Brand 切换
> 运营域名：`https://inkless.run`
> 状态：仓库内发布收口分支完成；计划整体仍受外部 GitHub、DNS/TLS、生产数据和 Linux CI 闸门阻塞
> 收口基线：`origin/main@a8b62902514ad45d38415192ca74ac21939ae1ac`
> 收口分支：`codex/inkless-release-closure`

## 0. 2026-07-20 发布收口证据矩阵

状态说明：

- **已验证**：实现和对应本地验证证据均存在。
- **缺外部闸门**：仓库配置已就绪，但需要外部写操作、生产数据或远端 CI 才能完成。
- **缺人工验收**：自动化与静态资产检查通过，仍需发布候选截图或人工视觉确认。

| 任务 | 状态 | 证据与剩余闸门 |
| --- | --- | --- |
| T01 品牌常量与残留门禁 | 已验证 | `bash scripts/check-brand-residue.sh` 通过；临时非 allowlist `Impress` 探针被正确拒绝；Quality Gate 已接入门禁。 |
| T02 视觉资产与浏览器品牌 | 缺人工验收 | PNG 尺寸为 32×32、180×180、1200×630，全部 SVG 通过 XML 校验；前端测试、构建和 E2E 通过；仍需发布候选的浅色/深色/移动端截图确认。 |
| T03 安装、后台与 i18n | 已验证 | 前端 34 个测试文件、154 个测试通过；TypeScript、ESLint、构建和 admin release-chain E2E 通过。 |
| T04 后端可见品牌与种子 | 已验证 | `go test -race ./...` 通过；Swagger 用 CI 锁定的 `swag@v1.16.6` 重生成后 clean diff。 |
| T05 CLI 与制品命名 | 已验证 | `go build ./cmd/inkless` 通过；`inkless --help` 只显示 Inkless 命令与说明。 |
| T06 module、Proto、package、仓库链接 | 缺外部闸门 | Go module、imports、Proto 和 npm workspace 已统一；Buf 重生成 clean diff；当前 remote 仍为 `yixian-huang/impress`，必须先完成 GitHub 仓库重命名，目标链接才可对外使用。 |
| T07 环境变量、globals、浏览器存储 | 已验证 | 前端兼容测试与后端 `brandcompat/config/storage` 测试在完整测试集中通过。 |
| T08 插件、Webhook、搜索索引 | 已验证 | canonical/legacy 插件 fixture、Webhook 双 headers/HMAC、Meilisearch legacy reindex 测试均在 `go test -race ./...` 中通过。 |
| T09 数据库默认值与生产数据 | 缺外部闸门 | 新默认值和 legacy DSN 测试通过；生产 PublishedConfig、数据库和 uploads 尚未备份、迁移或核对。 |
| T10 Docker、systemd、QuickBox、回滚 | 缺外部闸门 | 两套 `docker compose config` 通过；QuickBox env/原子回滚测试通过；双实例真实 backup/restore/upgrade/rollback smoke 通过；基础 systemd unit 已放行 data、uploads、backups、plugins 与 plugin data，跨平台门禁本机静态校验通过，Linux `systemd-analyze` 留给 CI；旧 pCloud 脚本硬编码凭证已移除，历史凭证待外部轮换。 |
| T11 `inkless.run` 与 Nginx | 缺外部闸门 | Nginx、BASE_URL、CORS 配置已在仓库就绪；实时只读核验显示 NS 为 Cloudflare，但 apex 无 A/AAAA、www 无 A/CNAME，HTTPS 无法解析；未写 DNS、未签发证书、未切生产流量。 |
| T12 当前文档与社区文件 | 已验证 | `pnpm --dir docs-site build` 通过；品牌残留门禁覆盖当前文档与社区文件。 |
| T13 全链路门禁 | 缺外部闸门 | 本地 lint/typecheck/154 tests/build/E2E、Go mod/vet/race tests/build、Swagger/Proto、Compose、QuickBox、双实例 smoke 全部通过；远端 Linux CI、生产 smoke、视觉截图仍未执行。 |

本轮仓库内新增收口：

- `ops/systemd/inkless.service`、`docs/deployment.md`：在 `ProtectSystem=strict` 下完整放行 data、uploads、backups、plugins 与 plugin data 写路径。
- `scripts/check-systemd-unit.sh`：严格验证 Inkless systemd 用户、路径、二进制、安全指令和全部运行时写路径；Linux 环境额外运行 `systemd-analyze verify`。
- `.github/workflows/quality-gate.yml`：执行 systemd unit 门禁。
- `scripts/README.md`：记录新的发布验证命令。
- `scripts/deploy-run.sh`：移除基线中硬编码的 pCloud 凭证，改为强制读取 `PCLOUD_USER`/`PCLOUD_PASS`；已进入 Git 历史的原凭证必须在外部轮换。

明确未授权、未执行的外部动作：GitHub 仓库重命名、DNS 写入、证书签发、生产部署、PublishedConfig/数据库/uploads 写入或迁移、已暴露 pCloud 凭证轮换。

## 1. 目标与完成定义

本次迁移不是单纯替换文案，而是将当前混杂的 `Impress / impress`、`Blotting Consultancy / blotting`、`印迹` 三套标识统一为 Inkless，并同时完成代码、生成物、持久化配置、公开技术接口、部署资产、文档和线上域名的切换。

统一命名契约：

| 场景 | Canonical 值 |
| --- | --- |
| 对外品牌 | `Inkless` |
| 产品全称 | `Inkless CMS` |
| CLI | `inkless` |
| Go/GitHub 仓库路径 | `github.com/yixian-huang/inkless/backend` |
| npm 根包 | `inkless` |
| 前端包 | `@inkless/web` |
| 文档包 | `inkless-docs` |
| 服务、用户、目录、镜像、容器前缀 | `inkless` |
| API 服务标识 | `inkless-api` |
| 新安装 SQLite 文件 | `inkless.db` |
| 生产目录 | `/opt/inkless` |
| 运营域名 | `inkless.run` |

完成定义：

1. 用户可见界面、生成的新配置、CLI 输出、API 文档、当前文档、部署资产中不再出现旧品牌。
2. 新安装、新构建、新部署只产生 Inkless 命名的文件、数据库、容器、服务和浏览器键。
3. 旧名称只允许存在于精确列出的兼容代码和历史归档中；兼容代码必须有自动化测试，且不得成为新输出的默认值。
4. 现有数据库、上传文件、插件、Webhook 消费者和浏览器本地数据在升级后仍可使用。
5. `inkless.run`、HTTPS、SEO、favicon、Open Graph 和后台品牌形成完整闭环。

## 2. 范围边界

### 2.1 纳入范围

- 产品界面、安装流程、管理后台、默认元数据、Logo、favicon 和分享图。
- CLI、Go module、Proto package、npm 包、GitHub 仓库链接。
- 内置主题、插件元数据、邮件、Swagger、健康检查服务名。
- 环境变量、插件握手、Webhook headers、JS globals、localStorage、搜索索引前缀。
- Docker、systemd、QuickBox、构建脚本、发布产物、SQLite/Postgres 示例命名。
- 当前开发文档、部署文档、社区文件和 CI 品牌门禁。
- 线上 Inkless 实例的 PublishedConfig、域名、Nginx、TLS 和数据迁移。

### 2.2 不做错误的全局替换

- 公共站点的可配置站名继续由站点配置控制；不能把所有用户网站默认名强制设为 Inkless。动态站点品牌来自 [useBranding.ts](../../frontend/src/hooks/useBranding.ts#L32) 和 [siteConfig.ts](../../frontend/src/types/siteConfig.ts#L103)。
- `.omx/plans/`、`docs/superpowers/` 等历史记录保留原始语境，不作为产品品牌残留失败条件。
- 已有生产数据库文件、搜索索引和上传目录先迁移再切换，不允许直接覆盖、删除或无备份重命名。
- 生成文件不手改：Swagger 从 [backend/cmd/server/main.go](../../backend/cmd/server/main.go#L92) 重新生成；Proto 从 [plugin.proto](../../backend/pkg/pluginproto/plugin.proto#L3) 重新生成。

## 3. 迁移策略

采用“一次发布切换 Canonical Brand，旧接口只读兼容”的策略：

- 新代码、新配置、新文档和新部署全部只输出 Inkless。
- 旧环境变量和浏览器键允许读取，读取成功后迁移到新键；不再写回旧键。
- Webhook 在一个过渡版本内同时发送 `X-Inkless-*` 与旧 `X-Impress-*` headers，文档只推荐新 headers。
- 插件宿主优先使用 Inkless 握手，失败时对旧插件执行受控的 legacy fallback；SDK 新版本只生成 Inkless 握手配置。
- 现有 Meilisearch 配置缺少显式前缀时按 legacy prefix 处理，避免搜索失效；新建配置默认 `inkless_`，迁移工具负责重建并切换索引。
- 兼容标识集中在有注释、有测试的 compatibility surface 中，不允许散落在产品代码。

## 4. 实施任务

### T01 — 建立品牌常量与残留基线

**目标**：避免后续在各处重新硬编码品牌和域名。

改动：

- 新增前端产品品牌常量，例如 `frontend/src/config/productBrand.ts`，包含 `Inkless`、`Inkless CMS`、`inkless.run` 和默认产品描述；站点自身品牌仍从 GlobalConfig 读取。
- 新增后端产品元数据常量，例如 `backend/pkg/brand/brand.go`，供 CLI、Swagger、邮件、插件作者和服务标识使用。
- 重写 [scripts/check-brand-residue.sh](../../scripts/check-brand-residue.sh#L1)：同时检查 `Impress|impress|IMPRESS|Blotting|blotting|印迹`，覆盖 frontend、backend、docs-site、docs、ops、scripts、CI 和根配置。
- 品牌门禁采用精确 allowlist：历史目录、legacy compatibility 常量、迁移说明；禁止整目录跳过 `scripts/`、`ops/` 或 Swagger。
- 在 [quality-gate.yml](../../.github/workflows/quality-gate.yml#L184) 中加入品牌残留检查。

验收：

- 门禁能对任一产品文件中新增的 `Impress` 或 `Blotting` 返回非零退出码。
- 门禁输出的每个旧品牌命中都对应一条带原因的 allowlist 规则。
- 产品品牌常量与用户站点品牌类型保持不同模块，禁止互相覆盖。

### T02 — 替换视觉资产与浏览器品牌闭环

**目标**：浏览器、登录、后台和分享渠道完整显示 Inkless。

改动：

- 删除/替换旧 [frontend/public/images/logo.png](../../frontend/public/images/logo.png)，建立 `frontend/public/brand/`：
  - `inkless-wordmark.svg`
  - `inkless-mark.svg`
  - `favicon.svg`
  - `favicon-32x32.png`
  - `apple-touch-icon.png`
  - `og-default.png`（1200×630）
- 在 [frontend/index.html](../../frontend/index.html#L8) 中将静态 fallback title、description、Open Graph 和 Twitter metadata 改为 Inkless，并同步重命名内部 SEO markers。
- 在应用启动时把 `branding.favicon` 写入或更新 DOM 的 `<link rel="icon">`；未配置站点 favicon 时回退到 Inkless favicon。配置字段当前存在于 [site-config/page.tsx](../../frontend/src/pages/admin/site-config/page.tsx#L152)，但运行时尚未消费。
- 登录页增加 Inkless wordmark；后台侧栏将旧“印”标识替换为 Inkless mark，相关位置见 [AdminSidebar.tsx](../../frontend/src/pages/admin/components/AdminSidebar.tsx#L313)。
- 所有图片补齐 `alt="Inkless"`，验证浅色、深色、折叠侧栏和移动端显示。

验收：

- 16×16、32×32、180×180、512×512 场景下无旧 Logo、无透明边界裁切问题。
- 无站点 favicon 时显示 Inkless favicon；配置自定义 favicon 后无需刷新配置即可更新。
- 页面源码和浏览器 DevTools 中的 title、description、OG、Twitter、favicon 都符合 Inkless 或当前站点配置。
- 后台登录页、展开侧栏、折叠侧栏和移动端均通过截图检查。

### T03 — 统一安装流程、管理后台和国际化文案

**目标**：所有产品级用户界面只显示 Inkless。

改动：

- 更新中英文 setup 文案，当前英文入口见 [frontend/src/i18n/local/en/setup.ts](../../frontend/src/i18n/local/en/setup.ts#L2)。
- 更新 setup 页面数据库默认值、提示、成功页和后端重启指令，当前默认值位于 [frontend/src/pages/setup/page.tsx](../../frontend/src/pages/setup/page.tsx#L61)。
- 更新管理后台登录、导航、标题、错误提示和产品署名；站点名称继续显示用户配置值。
- 内置主题和前端插件的作者统一为 `Inkless CMS`；演示站作者、公司名等内容改为与产品品牌无关的中性示例。
- 所有中英文词条做键值一致性检查，不新增仅单语可见的品牌字符串。

验收：

- setup 全流程中不存在旧品牌；成功页给出的命令为 `inkless serve`。
- 管理后台产品壳显示 Inkless，公共站点 header/footer 仍可显示任意用户站点名称。
- 中英文构建均无缺失翻译键，浏览器控制台无 i18n fallback 警告。

### T04 — 清理后端用户可见品牌和种子数据

**目标**：API、邮件、插件、主题和新建数据只生成 Inkless 或中性站点内容。

改动：

- 更新 Swagger title/description 和健康检查 `service` 字段，入口位于 [backend/cmd/server/main.go](../../backend/cmd/server/main.go#L92)。
- 更新邮件测试主题，当前硬编码位于 [email_service.go](../../backend/internal/service/email_service.go#L360)。
- 更新内置 analytics、captcha、email、meilisearch、s3storage、webhook 插件元数据作者。
- 更新内置主题作者；将 [seed.go](../../backend/internal/seed/seed.go#L464) 中“印迹咨询 / Blotting Consultancy”、旧邮箱和旧 Logo alt 改为中性演示站内容。
- 空白安装继续使用中性 `My Site`/对应中文默认值，不将用户网站默认品牌绑定为 Inkless。
- 重新生成 `backend/docs/swagger/{docs.go,swagger.json,swagger.yaml}`。

验收：

- `/api-docs/index.html` 显示 Inkless CMS。
- `/health` 或对应健康接口返回 `inkless-api`。
- 测试邮件主题、内置插件和主题作者均为 Inkless CMS。
- `SEED_MODE=demo` 生成的内容没有 Impress、Blotting 或印迹品牌，且仍能完整渲染。
- Swagger 重新生成后 `git diff --exit-code backend/docs/swagger/` 通过。

### T05 — 重命名 CLI、构建入口和导出物

**目标**：唯一 Canonical CLI 为 `inkless`。

改动：

- 将 `backend/cmd/impress/` 移至 `backend/cmd/inkless/`，Cobra root `Use`、Short、Long、examples、初始化输出全部切换为 Inkless，当前入口见 [backend/cmd/impress/main.go](../../backend/cmd/impress/main.go#L18)。
- 导出默认文件名改为 `inkless-export-TIMESTAMP.json`，但导入器继续接受任意旧导出文件名和旧数据 schema。
- [Makefile](../../Makefile#L51)、CI、seed scripts、开发文档改为构建/调用 `inkless`。
- 服务端二进制统一为 `inkless-api-{version}`，稳定 symlink 为 `inkless-api-latest`。
- 不再生成新的 `impress` 或 `blotting-api` 二进制；线上回滚所需旧二进制保留在旧 release 目录，不纳入新制品。

验收：

- `make build-cli` 仅生成 `backend/inkless`。
- `./backend/inkless --help`、`init`、`migrate`、`seed`、`export`、`import`、`serve` 命令测试通过。
- CI 和脚本中不存在对 `./impress`、`cmd/impress` 或 `blotting-api-*` 的新调用。
- 旧版本生成的 export JSON 可由 `inkless import` 成功导入。

### T06 — 统一 Go module、Proto、npm 包和仓库链接

**目标**：源码身份与公开仓库身份统一。

改动：

- 将 [backend/go.mod](../../backend/go.mod#L1) module 改为 `github.com/yixian-huang/inkless/backend`，机械更新约 268 个 Go 文件中的内部 imports。
- 将 [plugin.proto](../../backend/pkg/pluginproto/plugin.proto#L3) 的 `go_package` 改为新 module path，使用 `buf generate` 重新生成 `plugin.pb.go` 和 `plugin_grpc.pb.go`。
- 根 [package.json](../../package.json#L2) 改为 `inkless`；`frontend/package.json` 改为 `@inkless/web`；`docs-site/package.json` 改为 `inkless-docs`，更新 lockfile。
- GitHub 仓库重命名为 `yixian-huang/inkless` 后，更新 remote、clone commands、Issues、源码链接、QuickBox repo URL 和 VitePress 导航。
- 新增/更新根 README、LICENSE 引用、贡献指南和社区模板，统一使用 Inkless CMS。

验收：

- `go list ./...` 输出的本地包均以 `github.com/yixian-huang/inkless/backend` 开头。
- `buf generate` 后生成文件无 diff，生成文件内不含 `blotting-consultancy`。
- `pnpm install --frozen-lockfile` 成功，workspace 包名唯一且不再使用 `react` 作为本项目包名。
- 文档中的 GitHub 链接全部可访问并指向新仓库。

### T07 — 迁移环境变量、JS globals 和浏览器存储

**目标**：新安装只使用 Inkless 技术标识，现有客户端无数据丢失。

改动：

- `INKLESS_ENV_FILE` 替代 [envfile.go](../../backend/pkg/config/envfile.go#L38) 中的 `IMPRESS_ENV_FILE`；读取顺序为新变量优先、旧变量 fallback，并对旧变量记录一次 deprecation warning。
- `INKLESS_SECRET_KEY` 替代 [storage_runtime.go](../../backend/internal/service/storage_runtime.go#L410) 中的 `IMPRESS_SECRET_KEY`，采用相同优先级策略。
- `__INKLESS_SHARED__`、`__INKLESS_THEME_REGISTER__` 替代 [externals.ts](../../frontend/src/plugins/externals.ts#L10) 和 [ThemeManager.ts](../../frontend/src/plugins/ThemeManager.ts#L84) 中的 globals；外部 bundle 加载期间对旧 global 提供受控 alias。
- `inkless.setup.step`、`inkless.setup.draft`、`inkless.comment.guest` 替代旧 localStorage keys；集中实现 `migrateLegacyBrandStorage()`：新键缺失时读旧键、校验结构、写新键、删除旧键。
- `.inkless-storage-probe` 替代 `.impress-storage-probe`；启动时清理成功完成的旧 probe 文件，但不得删除真实上传内容。
- 新生成的 `.env` 注释和示例只显示 Inkless。

验收：

- 新旧环境变量同时存在时始终使用 `INKLESS_*`。
- 仅设置旧环境变量时应用仍启动并产生明确 deprecation 日志。
- 预置旧 localStorage 后升级，setup 草稿和评论访客身份保持不变，迁移完成后只剩新键。
- 外部主题分别使用新/旧注册 global 时都能加载，产品自带 bundle 只使用新 global。
- 新建部署和浏览器会话不会写出任何旧键。

### T08 — 迁移插件握手、Webhook 和搜索索引

**目标**：公开技术契约完成品牌切换，同时保护现有集成。

改动：

- 在 [pluginsdk/sdk.go](../../backend/pkg/pluginsdk/sdk.go#L14) 定义 Canonical Inkless handshake：`INKLESS_PLUGIN=inkless-cms-v1`；旧 handshake 移入带 `Deprecated` 注释的 legacy compatibility 模块。
- 在 [grpc_host.go](../../backend/internal/plugin/grpc_host.go#L53) 实现确定性的 canonical-first、legacy-fallback 启动策略；fallback 只对握手失败生效，插件自身运行错误不得误判为品牌兼容问题。
- Webhook User-Agent 改为 `Inkless-CMS-Webhook/1.0`，新增 `X-Inkless-Event/Timestamp/Signature`；一个过渡版本内同步发送旧 headers，当前发送位置见 [webhook/plugin.go](../../backend/internal/plugins/webhook/plugin.go#L233)。
- Meilisearch 新配置默认 prefix 从 [plugin.go](../../backend/internal/plugins/meilisearch/plugin.go#L94) 的 `impress_` 改为 `inkless_`。
- 为既有插件配置增加显式 legacy-prefix 判定；新增可重复执行的 reindex/cutover 操作，验证文档数一致后才删除或保留旧索引。
- 更新插件 SDK README、示例、契约测试和 Webhook 文档。

验收：

- 新 SDK 插件通过 Inkless handshake 加载。
- 使用旧 SDK 编译的 fixture 插件仍能由新宿主加载；日志明确标注 legacy fallback。
- 错误插件不会因 fallback 被启动两次或掩盖真实错误。
- Webhook 新 headers 的 event、timestamp、HMAC 验证通过；过渡期旧 headers 与新 headers 值一致。
- 全新 Meilisearch 配置创建 `inkless_*` 索引；旧配置升级后搜索结果数和语言分区不变。

### T09 — 统一数据库默认值与生产站点数据

**目标**：新安装以 Inkless 命名，现有数据安全升级。

改动：

- 将 [backend/pkg/config/config.go](../../backend/pkg/config/config.go#L25)、[dsn.go](../../backend/pkg/config/dsn.go#L34)、setup、CLI、Makefile 和 Compose 的新安装默认数据库改为 `inkless.db` / `inkless` Postgres 用户与库名。
- 所有测试 fixture 同步使用 Inkless 默认值，但增加至少一个 legacy DSN 测试证明旧路径仍可显式连接。
- 不在应用启动时自动重命名生产数据库；部署切换流程先备份，再复制/验证到 `/opt/inkless/data/inkless.db`，最后原子切换服务。
- 使用现有 GlobalConfig API 或经过审计的一次性运维脚本，把 `inkless.run` 运营实例的 PublishedConfig 更新为 Inkless 名称、Logo、favicon、SEO 描述和社交分享图。
- 数据更新前导出 PublishedConfig；更新后比较 revision/version 并保留回滚文件。

验收：

- 全新 SQLite setup 只创建 `data/inkless.db`。
- 全新 Postgres setup 默认显示 `inkless` 命名，且可自定义覆盖。
- 显式传入旧 `impress.db` 或 `blotting.db` DSN 时仍能启动，不会自动移动或删除文件。
- 生产迁移前后用户、文章、页面、媒体、配置记录数一致；上传文件 checksum 抽样一致。
- `inkless.run` PublishedConfig 发布版本增加且可用导出的旧版本回滚。

### T10 — 统一容器、systemd、QuickBox 和发布制品

**目标**：所有新部署资产只使用 Inkless。

改动：

- [docker-compose.yml](../../docker-compose.yml#L1) 和 `docker-compose.sqlite.yml`：project、container、network、volume、DB/user、image 全部切换为 `inkless-*`。
- [artifact-manifest.json](../../ops/artifact-manifest.json#L3)：project、binary 和 symlink 切换为 Inkless。
- 将 `ops/systemd/impress.service` 重命名为 `ops/systemd/inkless.service`，User/Group、WorkingDirectory、EnvironmentFile、ExecStart、ReadWritePaths 改为 `/opt/inkless`，当前旧值见 [impress.service](../../ops/systemd/impress.service#L8)。
- 更新 `qb-host-bootstrap.sh`、`qb-artifact-*`、`qb-docker-deploy.sh`、`deploy-run.sh`、rollback/build scripts 和 QuickBox JSON 中的 repo、workdir、service、user、path、container 与日志名。
- 新部署只创建 `inkless` 系统用户和 `/opt/inkless`；旧 `/opt/impress`、`/opt/blotting` 只作为迁移源，不由脚本递归删除。
- 发布制品清单记录品牌迁移版本、数据源路径、目标路径和 rollback unit。

验收：

- `docker compose config` 成功且输出中无旧品牌。
- systemd unit 通过 `systemd-analyze verify`；只引用 `/opt/inkless` 和 `inkless-api-latest`。
- artifact build/activate/rollback 在临时目录完成一次闭环，symlink 能在新旧两个 Inkless release 间切换。
- 全新宿主 bootstrap 后不存在 `impress`/`blotting` 用户、服务、容器或新目录。
- 从旧部署迁移时，旧目录保持可恢复，回滚不依赖已删除文件。

### T11 — 切换 `inkless.run`、Nginx、SEO 与安全配置

**目标**：运营域名成为唯一 Canonical Origin。

改动：

- 将 [nginx.conf](../../nginx.conf#L1) 更新为 Inkless，`server_name inkless.run www.inkless.run`，确定唯一 canonical host，并将另一 host 301 重定向。
- 生产 `BASE_URL=https://inkless.run`；`CORS_ALLOWED_ORIGINS` 至少包含实际前端 origin，读取逻辑位于 [bootstrap.go](../../backend/pkg/config/bootstrap.go#L91)。
- 配置 TLS、HSTS（确认全站 HTTPS 后启用）、静态缓存、favicon/OG 缓存策略和上传目录。
- Sitemap、RSS、canonical、Open Graph URL 和结构化数据不得指向旧域名。
- 健康检查、API、后台登录、上传资源和前端路由在反向代理下逐项 smoke test。

验收：

- `https://inkless.run` 返回 200 且证书主机名、有效期和链路正确。
- HTTP 和非 canonical host 一次 301 到 `https://inkless.run`，无重定向环。
- HTML 中 canonical、OG URL、favicon、sitemap、RSS 均使用 HTTPS Inkless 域名或用户站点自身配置。
- 非 allowlist Origin 的 CORS 预检被拒绝；`inkless.run` 正常通过。
- `/admin`、`/api/*`、`/uploads/*` 和客户端路由刷新均正常。

### T12 — 更新当前文档、社区文件和迁移说明

**目标**：所有仍代表当前产品状态的文档统一为 Inkless。

改动：

- 更新 [CONTRIBUTING.md](../../CONTRIBUTING.md#L1)、docs-site 首页、VitePress 配置、getting started、theme/plugin guide、deployment、developer guide 和 migration 文档。
- 更新 Issue 模板、workflow 说明、README badges、GitHub clone/Issues/源码链接。
- 新增 `docs/inkless-brand-migration.md`：记录旧变量、旧 headers、旧 localStorage、旧索引、旧目录的兼容和弃用策略。
- 历史文件不改写正文，但在历史目录 README 中说明旧品牌仅代表当时项目名称。
- 文档示例默认使用 `inkless`, `inkless.db`, `/opt/inkless`, `inkless-api`, `inkless.run`。

验收：

- VitePress build 成功且站点 title、导航和 GitHub 链接均为 Inkless。
- 当前文档中没有旧命令、旧路径、旧数据库或旧域名示例。
- 旧技术标识只在迁移文档的 compatibility table 和历史归档中出现。

### T13 — 全链路验证和发布门禁

**目标**：以自动化证据证明迁移完成，而不是只依赖搜索替换。

验证命令：

```bash
bash scripts/check-brand-residue.sh
pnpm install --frozen-lockfile
pnpm lint
pnpm type-check
pnpm test
pnpm build
pnpm test:e2e
pnpm --dir docs-site build

cd backend
go fmt ./...
go vet ./...
go test ./...
go build ./cmd/server
go build ./cmd/inkless
swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal
cd pkg/pluginproto && buf generate && cd ../..
git diff --exit-code pkg/pluginproto docs/swagger

cd ..
docker compose config
```

补充验证：

- Fresh install E2E：SQLite setup → admin login → 创建并发布内容 → 上传媒体 → 重启 → 数据仍在。
- Upgrade E2E：复制旧品牌 fixture 数据/配置/localStorage → 新版本启动 → 验证兼容迁移 → 导出 → 回滚。
- Plugin integration：新 SDK fixture + 旧 SDK fixture。
- Webhook integration：新旧 headers 与签名。
- Search integration：旧索引升级和新索引 fresh install。
- Deployment smoke：systemd、Docker、QuickBox artifact 三条支持路径至少各验证配置；生产实际使用路径执行完整发布/回滚演练。
- Visual QA：登录、setup、后台展开/折叠、首页、favicon、暗色模式、移动端、OG preview。

最终门禁：

```bash
rg -n --hidden \
  --glob '!frontend/out/**' \
  --glob '!**/node_modules/**' \
  --glob '!.git/**' \
  'Impress|impress|IMPRESS|Blotting|blotting|印迹' .
```

输出必须只包含：

1. `.omx/plans/`、`docs/superpowers/` 等明确历史档案；
2. `docs/inkless-brand-migration.md` 的迁移说明；
3. 有单元/集成测试覆盖的 legacy compatibility 常量；
4. 本品牌门禁脚本自身的匹配模式。

## 5. 依赖顺序与并行边界

必须按顺序完成：

1. T01 命名契约和门禁先落地。
2. T06 module/package/Proto rename 完成后再大规模修改后端 imports 和 SDK 文档。
3. T07/T08 兼容层完成并有测试后，才能切换 T05 CLI 和 T10 部署制品。
4. T09 数据备份与迁移演练通过后，才能执行 T11 生产域名切换。
5. T13 全部门禁通过后才能发布。

可并行执行：

- T02 + T03：前端视觉与产品壳。
- T04 + T06：后端品牌内容与 module/生成物，需协调生成文件。
- T07 + T08：兼容协议，但共享 compatibility 命名时必须由同一负责人最终整合。
- T10 + T12：部署资产与文档。

## 6. 风险与缓解

| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 全局替换破坏用户站点品牌 | 所有用户网站被错误显示为 Inkless | 产品品牌常量与 GlobalConfig 分层；增加自定义站名 E2E |
| 插件握手直接改名 | 已编译外部插件全部无法启动 | canonical-first + legacy fallback；旧 SDK fixture 集成测试 |
| 索引前缀直接切换 | 线上搜索返回空结果 | 显式 legacy prefix；reindex 后校验文档数再切换 |
| 默认 DB/目录直接改名 | 数据丢失或服务启动空库 | 禁止应用自动移动；备份、checksum、原子服务切换、保留旧路径 |
| localStorage 键改名 | setup 草稿或评论身份丢失 | 一次性结构校验迁移；新键优先；迁移测试 |
| 只改源文件未改生成物 | Swagger/Proto/构建输出继续暴露旧品牌 | CI 重生成并检查 clean diff |
| systemd/脚本命名不一致 | 发布成功但服务无法启动或回滚 | artifact/activate/rollback 临时目录闭环测试 |
| 浏览器缓存旧 Logo/OG | 上线后仍传播旧品牌 | 文件指纹、缓存刷新策略、实际抓取验证 |

## 7. 发布与回滚标准

发布前：

- 导出数据库、PublishedConfig、插件配置、环境文件和部署清单。
- 备份 uploads，并至少完成 checksum 抽样。
- 保存旧 systemd unit、旧 release symlink 目标和旧 Nginx 配置。
- 在 staging 使用生产数据脱敏副本完成升级和回滚。

发布顺序：

1. 上传 Inkless 制品到新目录，不覆盖旧 release。
2. 迁移/复制数据并验证记录数、文件可读性与配置版本。
3. 启动 `inkless.service`，仅在本机端口执行健康检查。
4. 切换 Nginx 到 Inkless upstream 和域名配置。
5. 执行 smoke/E2E，确认 SEO、登录、内容、上传、插件、Webhook、搜索。
6. 保留旧服务和目录但停止运行，观察一个发布窗口后再另立清理任务。

触发回滚的明确条件：

- 健康检查连续 3 次失败；
- admin 登录、内容读取、上传或搜索任一核心链路失败；
- 数据记录数不一致或出现 migration error；
- 插件宿主 crash loop；
- 5xx 比例明显高于发布前基线。

回滚：恢复 Nginx 配置和旧 service/symlink，使用迁移前数据库与 PublishedConfig 备份；不在回滚过程中删除 Inkless 目录，以便事后诊断。

## 8. 总体验收清单

- [x] UI、邮件、API docs、CLI、文档和部署资产只显示 Inkless。
- [ ] Logo、favicon、OG、暗色/浅色和移动端全部验证。（静态资产与自动化已通过，待发布候选截图）
- [x] 新安装默认使用 `inkless.db`、`inkless` CLI 和 Inkless 技术标识。
- [x] 用户公共站点仍支持完全自定义品牌。
- [x] 旧环境变量、localStorage、插件、Webhook 和搜索索引兼容测试通过。
- [ ] Go module、Proto、npm packages、GitHub links 全部统一。（仓库内完成，待 GitHub 仓库重命名）
- [x] Docker、systemd、QuickBox、artifact、rollback 的仓库实现和本地门禁统一。
- [ ] `inkless.run` HTTPS、canonical、CORS、SEO 和反向代理验证通过。（DNS/TLS/生产流量未切换）
- [ ] 生产数据与 uploads 验证无损，PublishedConfig 可回滚。（生产写操作未授权）
- [x] lint、typecheck、build、Go test/vet、Swagger、Proto、docs build 全部通过。
- [x] 品牌残留扫描只剩受控 compatibility 和历史归档。

## 9. 停止条件

只有当第 8 节所有条目有实际命令输出、测试记录、截图或生产 smoke evidence 时，任务才可标记完成。任何“旧品牌只剩内部使用”的说法都必须对应 allowlist 和兼容测试，否则视为未完成。
