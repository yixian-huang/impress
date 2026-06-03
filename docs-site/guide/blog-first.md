# Blog-first 模式

Blog-first 使用 **blog-first** 主题：首页 `/` 展示作者介绍与最近文章，适合个人博客站点。

## 启用方式

在后台 **Admin → Theme** 激活 **blog-first** 主题。站点呈现（首页、Header/Footer、布局宽度）完全由当前激活主题决定，不再使用 Features 里的 site mode 开关。

## 内容从哪里来

| 区域 | 配置位置 |
|------|----------|
| 作者名、头像、简介 | 后台 **站点配置 → Author** |
| 站点名、标语 | **站点配置 → Identity** |
| 文章列表 | **文章** 管理（已发布文章） |
| RSS | Features → Blog → RSS（`/feed.xml`） |

**不要**在「页面 → home」里编辑博客首页：blog-first 主题不会渲染企业首页的 CMS 区块。

## 页面管理

- **Composable 页**（`/p/{slug}`）：照常创建与发布，用于 About、Projects 等自定义页。
- **静态博客路由**：`/blog` 归档、`/blog/:slug` 详情，由文章数据驱动。
- **切换回企业站**：Admin → Theme 激活 **corporate-classic**，企业首页 CMS 内容恢复展示。

## 导航

blog-first 主题仅注册 `home` 路由。Blog 链接由 Features → Public pages → `blog` 开关与静态 `/blog` 路由控制。

Header/Footer 由 **blog-first** 主题的 layout chrome 渲染，详见 [Theme layout & chrome](./theme-layout.md)。

## Header 配置

1. **主题设置**（Admin → Theme → 主题设置）：默认品牌区（text/logo/avatar）、RSS、社交链接
2. **Site Config → Header**：覆盖品牌模式与顶栏开关
3. **Site Config → Author / Brand**：作者名、头像、Logo URL
4. **Menus**：主导航链接

## 空白站默认

新站 seed 会安装并激活 **blog-first** 主题，同时 Features 预设仅开启 home/blog/contact 等博客常用路由（不含 legacy `siteMode` 字段）。
