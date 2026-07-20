# Extension Points

Inkless has three extension layers. This page summarizes the **shipped** contracts and trust boundaries.

## Trust model (read first)

| Surface | Trust level | Default |
|---------|-------------|---------|
| **External backend plugins** (`pkg/pluginsdk`, go-plugin/gRPC) | **Trusted server-side code** — manifest “sandbox” checks are capability declarations only, not OS isolation | **Off** (`ENABLE_EXTERNAL_PLUGINS=true` required) |
| **External UMD themes** (`ThemeManager.loadExternal`) | **Trusted script origin** — loads arbitrary JS into the admin/public SPA | Only install themes from sources you control |
| **Compile-time modules** (`internal/modules`) | First-party code shipped with the binary | Always linked |

Do **not** run untrusted third-party plugin binaries or theme URLs on production hosts.

See also: `backend/pkg/pluginsdk/README.md`, `docs/theme-contract.md`.

## Backend

### Providers (`internal/provider`)

Central `Registry` holds named providers (AI, storage, search, notifier, captcha, …). Core services and optional external plugins register here.

External beta plugins may register canonical names only: `notifier`, `search`, `captcha` (see pluginsdk README). Storage and custom routes are not part of the external beta surface.

### Modules (`internal/module`)

Feature modules implement:

```go
type Module interface {
  Name() string
  Init(deps Dependencies) error
  RegisterRoutes(public, admin *gin.RouterGroup)
}
```

Registered at startup in `internal/app/wire_handlers.go` (comment, form_submission, qa, backup). Third-party installable modules are not supported yet.

### EventBus & hooks (`internal/eventbus`)

- **EventBus**: in-process pub/sub for content lifecycle (publish, delete, …). Used for cache invalidation and logging.
- **HookRegistry**: ordered request/content hooks exist in code and tests; they are **not** fully wired into every publish path for external plugins yet.

### External plugins

- SDK: `backend/pkg/pluginsdk`
- Host: `backend/internal/plugin` (Manager, manifest, gRPC host)
- Example: `backend/examples/plugins/file-notifier`

## Frontend themes

- **Contract**: `@inkless/theme-host` / `frontend/src/theme-host` (`THEME_CONTRACT_VERSION`)
- **Registration**: `ThemeManager.registerBuiltIn` / `registerExternal` / `loadExternal`
- **Plugin shape**: `frontend/src/plugins/types.ts` → `ThemePlugin` (pages, layoutChrome, sections, tokens)
- **UMD peers**: `frontend/src/plugins/externals.ts` (`window.InklessThemeHost`, React globals)

Themes must import host APIs only via the theme-host facade (or UMD globals), not deep `@/` paths.

## Frontend modules

Comment and QA UIs live under `frontend/src/modules/*` and are wired into host routes/layout. There is no generic third-party widget slot manager yet.
