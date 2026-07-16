# Wave 2 收尾与外部插件闭环

## 目标

1. 达到 R2「配置即行为」：AI、翻译、远端存储的持久配置真实改变运行时行为。
2. 达到 R3 的插件部分：至少一个外部插件可以安装、启用、调用、禁用、重启恢复和卸载。

## 当前事实

- Wave 2 已完成：AI 与 Storage secret 加密持久化，启动恢复和运行时热切换已接入 server。
- QA、Wizard、Translation 按请求读取当前 AI provider；未配置时返回结构化 503。
- Wizard 写入 `unified_pages` composable 草稿，前后端请求/响应契约及 section 数据结构已对齐。
- StorageConfig 使用 camelCase，保存前真实探测远端；普通与分片上传统一走 StorageRuntime。
- Plugin Manager 已接入 server 启停、恢复和健康检查；真实 protobuf/gRPC client、受控 zip 安装和 provider 回滚已完成。
- Marketplace 下载分发、管理 UI 和插件升级仍未完成；当前真实生命周期通过独立 `/admin/plugins` 管理 API 提供。

## 进度

- [x] R2：配置即行为
- [x] 外部 Notifier 插件安装
- [x] enable/start/register/invoke
- [x] disable/stop/unregister
- [x] server restart restore
- [x] uninstall/cleanup
- [x] R3 插件黑盒验证

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
- 公开 `pkg/pluginproto` 与 `pkg/pluginsdk`，生成真实 protobuf/gRPC glue；独立 Go module 已验证可引用 SDK。
- 增加插件包校验、临时解包、原子安装目录、DB 记录、启停、provider 恢复、卸载文件清理和启动恢复。
- 提供 `file-notifier` 示例插件源码和 `plugin.yaml`，以及管理 API。

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

## 完成证据（2026-07-17）

- `/admin/plugins` 支持列表、zip 安装、启用、停用、卸载、设置更新和 notifier 测试。
- 外部插件运行时默认关闭；仅 `system:manage` 可操作，并需显式设置 `ENABLE_EXTERNAL_PLUGINS=true`。
- `plugins` / `plugin_settings` 已加入 server 与 CLI migration 模型；`PLUGIN_DIR`、`PLUGIN_DATA_DIR` 可配置。
- 插件启用会启动独立 go-plugin 子进程；停用、崩溃重启和 server 关闭会按实例安全恢复被替换的内置 provider。
- 停机先拒绝新 RPC 并有限等待在途调用；健康检查与 StopAll 通过 stopping 状态和 wait group 协调，关闭后不会重新拉起插件。
- `file-notifier` 黑盒测试完成 install → enable → invoke → disable → enable → server restart restore → invoke → uninstall。
- zip-slip、绝对/越界路径和 symlink 被拒绝；安装使用临时目录和原子 rename，失败会清理 staging/final 目录。
- 启用态卸载只停止运行态，DB 在删除提交前始终保持 enabled；安装/数据目录随后原子重命名，失败会恢复文件、进程与 provider，进程崩溃遗留的 uninstall staging 会在 manager 初始化时按 DB 真相恢复或清理。
- zip 安装同时校验声明大小和实际解压写入字节，实际总量超过 100 MiB 会立即失败并清理 staging。
- 独立临时 Go module 成功导入公开 SDK 并完成构建，证明不依赖仓库 `internal/` 边界。

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
- 外部插件二进制是可信服务端代码；manifest permission 只做声明一致性校验，不提供 OS 沙箱。
- 当前拒绝含 secret setting 的外部插件；开放此类配置前仍需补加密持久化与 API 脱敏。
- Beta 仅允许 canonical notifier/search/captcha provider；external storage、dependencies、custom routes 和 frontend entry 在所有权/交付链完成前拒绝安装或启用。
