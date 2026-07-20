# PicGo 对接预研（Inkless CMS）

日期：2026-07-21 · 状态：**预研 + API Key 已实现**（见 `docs/picgo.md`）  
对照：图床产品 **img.li / imgli** 正式短文 `imgli/docs/ops/picgo.md`。

## 1. 目标

让作者用 PicGo（或同类客户端）一键上传图片到 **某台 Inkless 站点的媒体库**，返回可写入文章的 URL，与后台「媒体库」同源。

## 2. 现有能力（代码事实）

| 项 | 现状 |
|---|---|
| 上传路由 | `POST /admin/media/upload`（admin 路由组，见 `routes_admin.go`） |
| 鉴权 | **JWT** 或长期 **API Key**（`ink_…` Bearer；见 `docs/picgo.md`） |
| 权限 | RBAC `media:create` ∩ API Key scope `media:create` |
| 表单字段 | multipart **`file`**（与 imgli 相同） |
| 成功响应 | **HTTP 201**，**裸 `Media` JSON**（无 `status/data` 信封） |
| URL 字段 | **`url`**（`json:"url"`） |
| 其它字段 | `id`, `filename`, `mimeType`, `size`, `width`, `height`, `storageKey`, `storageProvider`, `createdAt` |
| 存储 | `StorageRuntimeService`：local 默认；可接 S3 类 provider |
| 公网访问 | 本地：`router.Static("/uploads", cfg.UploadDir)`；S3 则 URL 常为对象/CDN 地址 |

响应示例（形状）：

```json
{
  "id": 12,
  "url": "https://your-site.example/uploads/1716...-shot.png",
  "filename": "shot.png",
  "mimeType": "image/png",
  "size": 12345,
  "width": 800,
  "height": 600,
  "storageKey": "...",
  "storageProvider": "local",
  "createdAt": "..."
}
```

**PicGo JSON Path 预研值**：`url`（不是 `data.links.url`）。

## 3. 与 img.li 的差异（对接时最容易踩坑）

| | **img.li (imgli)** | **Inkless** |
|---|---|---|
| 产品定位 | 图床 / 外链 | CMS 媒体库（进站内内容） |
| 路径 | `/api/v1/upload` | `/admin/media/upload` |
| 鉴权 | 长期 **API Token**（设置页） | 长期 **API Key**（`ink_…`）或 JWT |
| 信封 | `{status,data:{links:{url}}}` | 直接 `Media` |
| JSON Path | `data.links.url` | `url` |
| HTTP 码 | 200 | **201** |
| 游客 | 可选 | **无**公开上传（必须 admin 权限） |
| 限速 | 用户组三档 | 登录 + RBAC；另有 public rate limit 不覆盖 admin 上传 |

结论：**同一套 PicGo「Web 图床」插件可复用**，但 **URL / Header / JSON Path 必须分两套配置**，不能共用一份。

## 4. 推荐对接方式（不改代码即可试点）

### 4.1 手工流程（验证 API）

```bash
# 1) 登录拿 access token（字段名以实际 /auth/login 响应为准）
TOKEN=$(curl -sS -X POST 'https://YOUR_HOST/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"...","password":"..."}' | jq -r '.accessToken // .token // .data.accessToken')

# 2) 上传
curl -sS -X POST 'https://YOUR_HOST/admin/media/upload' \
  -H "Authorization: Bearer $TOKEN" \
  -F 'file=@./test.png' | jq .
```

### 4.2 PicGo 配置要点

| 配置项 | Inkless |
|---|---|
| API URL | `https://YOUR_HOST/admin/media/upload` |
| 文件字段 | `file` |
| Header | `Authorization: Bearer <JWT>` |
| JSON Path | `url` |

### 4.3 JWT 过期问题（主风险）

PicGo **不会**自动调 `/auth/refresh`。  
预研可选方案（按成本排序）：

| 方案 | 工作量 | 说明 |
|---|---|---|
| **A. 文档约定「用长会话 + 过期重登」** | 文档 | 个人站勉强可用，体验差 |
| **B. 新增「媒体上传 API Key」** | 中 | ✅ 已实现：`ink_` 前缀、`/admin/api-keys`、scope `media:create` |
| **C. 专用 PicGo 插件** | 高 | 内置 login/refresh；维护成本高 |
| **D. 网关侧 Basic/固定密钥反代到 admin upload** | 运维 | 安全边界要慎做 |

**预研建议**：若要认真做「作者用 PicGo 写 Inkless 博客」，优先 **B（API Key）**；短期试点可用 **A + 手动换 JWT**。

## 5. 安全与产品边界

1. **不要**对公网开放无鉴权的 `/admin/media/upload`。  
2. 上传即进入 **CMS 媒体库**（可能被文章引用）；与图床「随便外链」心理模型不同，文案要写清。  
3. 已有 S3/storage provider 时，返回的 `url` 必须是 **浏览器可访问的绝对 URL**（含正确 `baseURL` / CDN），否则 PicGo 复制的链不可用。  
4. CORS：PicGo 是桌面客户端直连，**一般不依赖浏览器 CORS**；浏览器内插件才要查 CORS。  
5. RBAC：仅 `media:create` 角色可传；编辑角色是否默认可传需产品确认。

## 6. 与 img.li 的协同（可选架构）

| 模式 | 含义 |
|---|---|
| **分离** | 写作配 img.li（外链图床）；Inkless 只存正文 Markdown 里的外链 |
| **统一进 CMS** | PicGo → Inkless media（本文对接） |
| **双写** | 不推荐：复杂度高、URL 漂移 |

对「博客主题 / product-first」：站内配图更适合 **Inkless media**；公开分享图床更适合 **img.li**。

## 7. 建议落地清单（Inkless 后续）

| # | 项 | 优先级 |
|---|---|---|
| 1 | 文档：`docs/picgo.md`（照抄 imgli 结构，字段换成上表） | P0 |
| 2 | 登录响应字段名写死到文档（`accessToken` 等）并加 curl 示例进 CI/手册 | P0 |
| 3 | **API Key / personal access token** for media upload | P1 |
| 4 | 可选：成功响应增加兼容字段 `links.url` 方便与 imgli 共用配置模板 | P2（非必须） |
| 5 | Admin 设置页「复制 PicGo 配置」按钮 | P2 |

## 8. 结论

- **技术上可对接**：字段 `file` + Bearer + JSON Path `url` 即可。  
- **体验上的最大缺口**：JWT 短命 vs PicGo 期望「配一次 Token 用很久」→ 应用 **API Key** 补齐。  
- **与 img.li 可共用插件，不可共用配置**；JSON Path / 路径 / 鉴权模型不同。

参考实现短文（图床侧）：`https://github.com/yixian-huang/imgli` → `docs/ops/picgo.md`。
