# Wave 2 收尾与外部插件闭环

## 目标

1. 达到 R2「配置即行为」：AI、翻译、远端存储的持久配置真实改变运行时行为。
2. 达到 R3 的插件部分：至少一个外部插件可以安装、启用、调用、禁用、重启恢复和卸载。

## 当前事实

- Wave 2 已完成：AI 与 Storage secret 加密持久化，启动恢复和运行时热切换已接入 server。
- QA、Wizard、Translation 按请求读取当前 AI provider；未配置时返回结构化 503。
- Wizard 写入 `unified_pages` composable 草稿，前后端请求/响应契约及 section 数据结构已对齐。
- StorageConfig 使用 camelCase，保存前真实探测远端；普通与分片上传统一走 StorageRuntime。
- Plugin Manager、Store、manifest 和生命周期状态机已存在，但未接入 server；go-plugin gRPC client 仍是占位实现，Marketplace 安装只返回下载 URL。

## 进度

- [x] R2：配置即行为
- [ ] 外部 Notifier 插件安装
- [ ] enable/start/register/invoke
- [ ] disable/stop/unregister
- [ ] server restart restore
- [ ] uninstall/cleanup
- [ ] R3 插件黑盒验证

## 实施决定

### Wave 2

- 增加共享 AES-GCM secret cipher，以服务端 JWT secret 派生加密密钥；数据库只保存带版本前缀的密文。
- AIConfig 和 StorageConfig 通过独立运行时服务统一负责：读取、解密、构建 provider、健康检查、持久化、原子切换与启动恢复。
- Provider Registry 增加 typed getter/setter/unregister；请求处理时解析当前 provider，不在启动时捕获实例。
- Wizard 写入 `unified_pages`；QA、Wizard、Translation 共用当前 AI provider。
- Translation 未配置时返回 503；单条/批量只预览，文章覆盖必须显式指定字段和覆盖策略。
- Media 与 chunked upload 统一调用当前 StorageProvider；媒体记录保存 storage key/provider，切换只影响新上传。
- 远端配置保存前执行真实 HEAD 探测；失败不持久化、不切换，继续使用最后一个健康 provider。

### 外部插件

- 选择 Notifier 作为第一个外部插件：接口小、无内容迁移、可通过测试通知明确证明调用链。
- 生成真实 protobuf/gRPC glue，修复 go-plugin host/client。
- 增加插件包下载、校验、临时解包、原子安装目录、DB 记录、启停、卸载文件清理和启动恢复。
- 提供独立示例插件二进制和 `plugin.yaml`，以及管理 API/CLI。

## 验收标准

### R2

- AI 配置保存后无需重启即可被 AI API、QA、Wizard、Translation 使用；重启后恢复。
- API key/secret key 的数据库值不是明文；GET 只返回 `hasApiKey` / `hasSecretKey`。
- 未配置 AI 时 AI、Wizard、Translation 返回结构化 503，不写入空译文。
- Wizard 应用结果出现在统一页面列表，可由统一页面编辑器发布。
- 远端存储保存会访问目标 endpoint；失败时 active provider 不变。
- 普通与分片上传均落到 active provider，删除使用媒体记录中的 storage key。

### 外部插件

- 从 zip/package 安装外部 notifier 插件后产生 installed 记录和原子安装目录。
- enable 启动真实子进程并注册 notifier；测试通知由插件处理。
- disable 后 provider 被注销且进程退出；再次 enable 可恢复。
- server 重启后 enabled 插件自动启动。
- uninstall 删除 DB 记录、安装目录和插件数据目录，不留下 enabled provider。
- 安装失败不留下 DB 记录、临时目录或半安装目录。

## 验证

- 后端 unit/integration/race tests。
- 前端 API/page tests、typecheck、lint、build。
- Playwright 覆盖 AI 配置状态、Translation 503/成功、Wizard 生成统一页面、Storage 保存失败回退。
- 外部插件黑盒测试构建示例插件，并完成 install → enable → invoke → disable → restart restore → uninstall。
- 独立 code-reviewer 与 architect 门禁。

## 风险与约束

- JWT secret 变化会导致已加密配置不可解密：启动时保持安全默认并暴露明确错误，不回退为明文使用。
- 存储切换不迁移历史对象；历史媒体保留原 URL 和 provider 标识，删除需要对应 provider 可用。
- 插件安装只接受限定大小的 zip，拒绝绝对路径、软链接和 zip-slip。
- 插件 settings 中声明为 secret 的字段必须加密或脱敏，管理 API 不返回明文。
