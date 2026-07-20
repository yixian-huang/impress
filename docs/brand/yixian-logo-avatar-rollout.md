# yx.ink · Logo & Avatar 品牌落地设计

> 来源：Claude Design handoff `Yixian Huang 品牌设计.zip`（Logo & Avatar v2 · 无衬线）  
> 原则：**不硬编码主题、不改主题包**；能配置的走站点配置；缺能力的先设计、后实现。

## 1. 设计包里有什么

| 区块 | 内容 | 性质 |
|------|------|------|
| Primary Logotype | `Yixian Huang` + `黄逸仙` + `yx.ink` 锁头 | 文案 + 排版规则 |
| Monogram YX | 四态：裸标 / 圆框 / 反白 / 反白圆 | 图形资产 |
| Clear space | 保护空间与最小尺寸规则 | 规范（文档） |
| Avatar matrix | Ink / Paper / Initial · 多尺寸 | 社交头像资产 |
| Favicon & App Icon | 96 / 48 / 32 / 16（16px 简化为 **Y**） | 站点图标资产 |
| Tokens | 单色板 + Manrope / Noto Sans SC / IBM Plex Mono | 设计令牌 |

**不是**完整站点主题稿；没有页面布局、组件状态机或营销文案系统。审美关键词：单色、极简、双语、无衬线。

### 核心令牌（可引用，勿写死进主题源码）

| Token | Hex | 用途 |
|-------|-----|------|
| Ink | `#141310` | 主字色、深底 |
| Paper | `#fbfaf8` | 浅底、反白字 |
| Canvas | `#e9e7e3` | 画廊底（可选） |
| Text-muted | `#4a4842` | 中文副文 |
| Hairline | `#ecebe6` | 分隔线 |
| Label | `#a29e96` | 说明/kicker |

字体：Manrope（wordmark/YX）、Noto Sans SC（中文）、IBM Plex Mono（元信息）。

---

## 2. 对照 yx.ink 现有能力：现在就能配

后台 **系统 → 站点配置**（`/admin/site-config`），**不要改主题代码**。

| 设计内容 | 配置入口 | 建议填法 |
|----------|----------|----------|
| 英文站名 / wordmark | 基本信息 · 站点名称（en） | `Yixian Huang` |
| 中文名 | 基本信息 · 站点名称（zh） | `黄逸仙` 或短品牌「一弦」（二选一，全站统一） |
| 域名/副标 | 基本信息 · 标语 | 例如 `yx.ink · open source, photography, notes` 或中文短句 |
| 作者名 | 作者 · 显示名称 | `黄逸仙` / `Yixian Huang` |
| 作者头像 | 作者 · 头像 | 上传 `avatar-yx-ink-132`（主）或 paper 变体 |
| Logo 浅色 | 品牌 · Logo（浅色） | 上传 wordmark 或 compact `YX \| yx.ink` |
| Logo 深色 | 品牌 · Logo（深色） | Paper 反白变体（若有） |
| Favicon | 品牌 · Favicon | `favicon-32`（浏览器）/ 另备 16 用 **Y** 简化版 |
| OG 默认图 | 品牌 · 默认分享图 | `og-default-1200x630` |
| 主色 | 品牌 · 主色 | `#141310`（单色系统；按钮/链接会变墨色） |
| SEO 默认标题 | SEO · 默认 SEO 标题 | `Yixian Huang · yx.ink` / `黄逸仙 · yx.ink` |
| 标题模板 | SEO · 标题模板 | `{page} · {site}` 或 `{page} | yx.ink` |
| 默认描述 | SEO · 默认描述 | 用 handoff 定位句改写：开源 / 摄影 / 想法 |

配套导出（SVG，可直接进媒体库）：见 `docs/brand/assets/`。

> SVG 文字依赖本机/浏览器字体；**生产 favicon 建议再导出一版 outlined PNG**（设计规范要求：导出时转曲，避免依赖字体）。

### 顶栏展示建议

- 页眉 · 品牌展示方式 → **Logo 图片** 或 **头像 + 名称**  
  - Logo：compact 或 wordmark  
  - 头像：Ink 圆形 YX（个人站更贴「Personal Brand」）

---

## 3. 能影响外观、但应走「主题设置」而非硬编码

| 设计内容 | 现状 | 建议 |
|----------|------|------|
| 全局字体 Manrope / Noto / Mono | 主题 tokens + 字体预设（主题管理） | 在**当前激活主题**的字体设置里选/配 Manrope、Noto Sans SC；**不要**写死进 blog-first 源码 |
| surface / border 接近 Paper/Hairline | 主题色板 | 主题管理里微调 surface=`#fbfaf8`、border=`#ecebe6`、onSurface=`#141310`（可选，会改变全站气质） |
| 主色=Ink | 站点配置 primaryColor 与主题 primary 可能双源 | 统一约定：**品牌主色以站点配置为准，或主题为准**，避免两套打架（见 §5 缺口） |

---

## 4. 缺能力：先设计、后做（不改主题）

### 4.1 品牌资产管线（高优先）

**问题**：设计包是 HTML 参考 + 无转曲资源；站点要 favicon/avatar/OG 的 **稳定位图/outline SVG**。

**目标形态**：

```
品牌源（字标规则）
  → 导出流水线（outline SVG / PNG 多尺寸）
  → 媒体库
  → 站点配置字段引用 URL
```

**功能设计（不必现在写代码）**：

1. **Admin · 品牌资产**（可挂在站点配置「品牌与图片」下）  
   - 一键生成/上传：favicon 16/32、app 96、avatar ink/paper、og  
   - 预览 clear-space 最小尺寸提示（文案即可）  
2. **导出清单校验**：缺 16px「Y」简化版时警告  
3. **不绑主题**：只产出 URL，写入现有 `brand.*` / `author.avatar`

### 4.2 字标锁头组件（中优先）

**问题**：真正的 logotype 是「EN 名 + 中文 + yx.ink」三层，不是一张图能完全替代全部场景。

**设计**：

- 可选 `header.brandMode = "wordmark"`（新枚举）  
- 渲染规则读站点配置：`identity.name` + `identity.tagline` 或固定 `yx.ink`  
- 样式用 **CSS 变量**（来自主题），默认 tracking 贴近 handoff；**不引入新主题包**

### 4.3 单色品牌预设（低优先）

**问题**：handoff 是严格 monochrome；现主题偏彩色 primary。

**设计**：主题管理增加「应用品牌色板」按钮：把 Ink/Paper/Hairline 写入 **已发布主题 draft tokens**，仍走主题发布流程，而不是 fork 主题代码。

### 4.4 品牌规范页（可选）

公开路由 `/brand` 或关于页区块：展示 clear space / 变体说明（内部用或开源品牌页）。内容 CMS 可配，**不进主题**。

---

## 5. 建议落地顺序（yx.ink）

### 本周可完成（零代码）

1. 导出/上传 `docs/brand/assets/*` 到媒体库（或先用 SVG）  
2. 站点配置填名称、标语、作者、头像、logo、favicon、OG  
3. 主色设为 `#141310`  
4. SEO 标题/描述对齐个人定位句  
5. 页眉改为 Logo 或 头像+名  

### 下一迭代（有设计再开发）

1. **品牌资产导出**：outline PNG 流水线 + 站点配置引导  
2. **`wordmark` 顶栏模式**（配置驱动）  
3. **主题「应用 monochrome 色板」**（写 tokens，不硬编码）  
4. 可选：公开 Brand kit 页  

### 明确不做

- 把 HTML handoff 当生产页面拷进前端  
- 为 yx.ink 单独硬编码一套 logo React 组件写死文案  
- 新建主题包「yixian-brand」除非长期多站复用  

---

## 6. 文案/身份统一建议

| 场景 | 建议 |
|------|------|
| 正式英文 | Yixian Huang |
| 正式中文 | 黄逸仙 |
| 短标 / 域名 | yx.ink |
| 字母标 | YX（≥24px）；16px 仅用 **Y** |
| 定位一句话（可作 slogan） | 开发者 · 开源 / 摄影 / 想法（中英各一版，写入标语字段） |

站名「一弦」与「黄逸仙」二选一作 `identity.name.zh`，避免顶栏与文章署名冲突；另一个放 `author.name` 或 slogan。

---

## 7. 本仓库附带文件

| 路径 | 说明 |
|------|------|
| `docs/brand/assets/*.svg` | 可上传的临时矢量资产（字体依赖，生产建议转曲 PNG） |
| `docs/brand/yixian-logo-avatar-rollout.md` | 本文 |

原始 zip 仅作设计源，**勿提交生产运行时依赖**。
