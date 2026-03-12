# Phase 3: Plugin Architecture Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable third-party developers to create, install, and manage plugins and themes. Deliverables: plugin runtime with HashiCorp go-plugin, Go/JS plugin SDKs, six official demo plugins, and a theme/plugin marketplace.

**Architecture:** Phase 3 introduces a plugin subsystem as a new layer between the existing Provider/EventBus infrastructure (Phase 2) and external plugin binaries. Plugins run as separate processes managed by HashiCorp go-plugin (gRPC over stdio). Each plugin implements one or more Provider interfaces via gRPC. The frontend gains an ExtensionSlot system for plugin UI injection.

**Tech Stack:** HashiCorp go-plugin + gRPC (plugin runtime), protobuf (plugin protocol), GORM (plugin data), React.lazy + dynamic import (frontend extension slots), JSON Schema (settings UI generation)

**Spec:** `docs/superpowers/specs/2026-03-11-open-source-evolution-design.md` — Phase 3

**Prerequisites:** Phase 2 MUST be complete. Phase 2 provides:
- EventBus (`backend/internal/eventbus/eventbus.go`) — Publish/Subscribe with sync/async handlers
- HookRegistry (`backend/internal/eventbus/hooks.go`) — request-level BeforePublish/AfterPublish/BeforeRender etc.
- Content lifecycle events (`backend/internal/eventbus/events.go`) — ContentCreated/Updated/Published/Deleted
- Provider Registry (`backend/internal/provider/registry.go`) — Register/Get/List/MustGet
- Provider interfaces: StorageProvider, SearchProvider, NotifierProvider, CaptchaProvider
- CLI tool (`backend/cmd/impress/`) with `impress plugin create` stub
- Swagger docs at `/api-docs`

**Go module:** `blotting-consultancy` (in `backend/go.mod`)

**Decision: Plugin Runtime Approach**

Per project guidance, we skip the WASM POC (too limiting for CMS use — no filesystem, no network, no Go stdlib access) and use **HashiCorp go-plugin** as the primary approach. This is battle-tested (Terraform, Vault, Nomad, Packer all use it), provides:
- gRPC over stdio (no port management)
- Automatic process lifecycle (host starts/stops plugin processes)
- Health checking and crash recovery
- Version negotiation
- Mutual TLS for security
- Go and non-Go plugin support via gRPC

We implement a simpler **shared-library fallback** (Go `plugin` package) for trusted, same-binary plugins during development, but go-plugin is the production path.

---

## File Structure Overview

```
backend/
├── internal/plugin/                         (3.1 Plugin Runtime)
│   ├── types.go                             (create - PluginMeta, PluginState, Permission types)
│   ├── types_test.go                        (create)
│   ├── manifest.go                          (create - plugin.yaml parser/validator)
│   ├── manifest_test.go                     (create)
│   ├── lifecycle.go                         (create - Install/Enable/Disable/Uninstall state machine)
│   ├── lifecycle_test.go                    (create)
│   ├── manager.go                           (create - PluginManager: discovery, loading, process mgmt)
│   ├── manager_test.go                      (create)
│   ├── sandbox.go                           (create - permission enforcement, capability checks)
│   ├── sandbox_test.go                      (create)
│   ├── grpc_host.go                         (create - go-plugin host-side implementation)
│   ├── grpc_host_test.go                    (create)
│   └── store.go                             (create - plugin state persistence via GORM)
│   └── store_test.go                        (create)
├── internal/plugin/proto/                   (3.1/3.2 gRPC definitions)
│   ├── plugin.proto                         (create - plugin service protobuf)
│   └── plugin_grpc.pb.go                    (generated)
│   └── plugin.pb.go                         (generated)
├── internal/plugin/shared/                  (3.2 SDK shared types)
│   ├── interface.go                         (create - PluginInterface for go-plugin handshake)
│   └── interface_test.go                    (create)
├── internal/model/plugin.go                 (create - Plugin, PluginSetting GORM models)
├── internal/handler/plugin/                 (3.2.3/3.4 Plugin admin handler)
│   ├── handler.go                           (create - plugin CRUD, enable/disable, settings)
│   └── handler_test.go                      (create)
├── internal/handler/marketplace/            (3.4 Marketplace handler)
│   ├── handler.go                           (create - market API)
│   └── handler_test.go                      (create)
├── cmd/server/main.go                       (modify - wire PluginManager, plugin routes)
├── cmd/impress/cmd_plugin.go               (modify - flesh out plugin create/list/enable/disable)
├── go.mod                                   (modify - add go-plugin, grpc, protobuf deps)
│
├── sdk/                                     (3.2 Go Plugin SDK — separate importable package)
│   ├── go.mod                               (create - module blotting-consultancy/sdk)
│   ├── plugin.go                            (create - SDK entry point: Serve(), RegisterProvider())
│   ├── context.go                           (create - PluginContext: DB, EventBus, Config access)
│   ├── storage.go                           (create - StorageProvider gRPC client/server stubs)
│   ├── search.go                            (create - SearchProvider gRPC client/server stubs)
│   ├── notifier.go                          (create - NotifierProvider gRPC client/server stubs)
│   ├── captcha.go                           (create - CaptchaProvider gRPC client/server stubs)
│   ├── hooks.go                             (create - hook registration helpers)
│   ├── events.go                            (create - event subscription helpers)
│   ├── routes.go                            (create - HTTP route registration)
│   ├── database.go                          (create - plugin-scoped DB helpers)
│   ├── settings.go                          (create - JSON Schema settings definition)
│   └── plugin_test.go                       (create - SDK integration tests)

plugins/                                     (3.3 Official Demo Plugins — top-level dir)
├── s3-storage/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── provider.go                          (create - S3 StorageProvider impl)
│   ├── go.mod                               (create)
│   └── provider_test.go                     (create)
├── email-notifier/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── provider.go                          (create - SMTP NotifierProvider impl)
│   ├── go.mod                               (create)
│   └── provider_test.go                     (create)
├── meilisearch/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── provider.go                          (create - Meilisearch SearchProvider impl)
│   ├── go.mod                               (create)
│   └── provider_test.go                     (create)
├── captcha/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── provider.go                          (create - reCAPTCHA/hCaptcha/Turnstile impl)
│   ├── go.mod                               (create)
│   └── provider_test.go                     (create)
├── webhook/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── handler.go                           (create - webhook dispatch + event subscription)
│   ├── go.mod                               (create)
│   └── handler_test.go                      (create)
├── analytics/
│   ├── plugin.yaml                          (create)
│   ├── main.go                              (create)
│   ├── handler.go                           (create - script injection hook)
│   ├── go.mod                               (create)
│   └── handler_test.go                      (create)

frontend/src/
├── plugin/                                  (3.2.4/3.2.6 Frontend Plugin System)
│   ├── ExtensionSlot.tsx                    (create - renders registered slot components)
│   ├── ExtensionSlot.test.tsx               (create)
│   ├── PluginRegistry.ts                    (create - register/unregister slot components)
│   ├── PluginRegistry.test.ts               (create)
│   ├── PluginLoader.ts                      (create - dynamic import of plugin JS bundles)
│   ├── PluginLoader.test.ts                 (create)
│   ├── PluginSettingsForm.tsx               (create - JSON Schema-driven settings panel)
│   ├── PluginSettingsForm.test.tsx           (create)
│   ├── types.ts                             (create - FrontendPlugin, ExtensionPoint types)
│   └── index.ts                             (create - barrel export)
├── pages/admin/plugins/                     (3.4.5 Marketplace UI)
│   ├── page.tsx                             (create - plugin list/marketplace page)
│   ├── PluginCard.tsx                       (create)
│   ├── PluginDetailModal.tsx                (create)
│   └── page.test.tsx                        (create)
├── theme/packages/types.ts                  (modify - add manifest fields for 3.4.1)
├── theme/packages/registry.ts               (modify - add install/uninstall from URL)
```

---

## Chunk 1: Plugin Metadata & Directory Structure (Tasks 3.1.1 — 3.1.2)

### Task 1: Define plugin types and metadata model (3.1.1)

**Files:**
- Create: `backend/internal/plugin/types.go`
- Create: `backend/internal/plugin/types_test.go`
- Create: `backend/internal/model/plugin.go`

**Steps:**
- [ ] Create `backend/internal/plugin/` directory
- [ ] Define core types in `types.go`:

```go
package plugin

// PluginState represents the lifecycle state of a plugin.
type PluginState string

const (
    StateInstalled  PluginState = "installed"
    StateEnabled    PluginState = "enabled"
    StateDisabled   PluginState = "disabled"
    StateFailed     PluginState = "failed"
)

// Permission represents a capability a plugin requests.
type Permission string

const (
    PermDatabaseRead    Permission = "database:read"
    PermDatabaseWrite   Permission = "database:write"
    PermFileSystemRead  Permission = "filesystem:read"
    PermFileSystemWrite Permission = "filesystem:write"
    PermNetworkOutbound Permission = "network:outbound"
    PermEventSubscribe  Permission = "event:subscribe"
    PermEventPublish    Permission = "event:publish"
    PermHookRegister    Permission = "hook:register"
    PermRouteRegister   Permission = "route:register"
    PermFrontendInject  Permission = "frontend:inject"
)

// AllPermissions returns the full list of valid permissions.
func AllPermissions() []Permission {
    return []Permission{
        PermDatabaseRead, PermDatabaseWrite,
        PermFileSystemRead, PermFileSystemWrite,
        PermNetworkOutbound,
        PermEventSubscribe, PermEventPublish,
        PermHookRegister, PermRouteRegister,
        PermFrontendInject,
    }
}

// PluginMeta represents the parsed content of plugin.yaml.
type PluginMeta struct {
    ID           string            `yaml:"id" json:"id"`
    Name         string            `yaml:"name" json:"name"`
    NameZh       string            `yaml:"nameZh" json:"nameZh"`
    Version      string            `yaml:"version" json:"version"`
    Description  string            `yaml:"description" json:"description"`
    Author       string            `yaml:"author" json:"author"`
    License      string            `yaml:"license" json:"license"`
    Homepage     string            `yaml:"homepage" json:"homepage"`
    MinAppVersion string           `yaml:"minAppVersion" json:"minAppVersion"`
    Dependencies []Dependency      `yaml:"dependencies" json:"dependencies"`
    Permissions  []Permission      `yaml:"permissions" json:"permissions"`
    Providers    []ProviderDecl    `yaml:"providers" json:"providers"`
    Routes       []RouteDecl       `yaml:"routes" json:"routes"`
    FrontendEntry string           `yaml:"frontendEntry" json:"frontendEntry"`
    SettingsSchema map[string]any  `yaml:"settingsSchema" json:"settingsSchema"`
}

// Dependency declares a required plugin.
type Dependency struct {
    PluginID   string `yaml:"pluginId" json:"pluginId"`
    MinVersion string `yaml:"minVersion" json:"minVersion"`
}

// ProviderDecl declares which provider interface a plugin implements.
type ProviderDecl struct {
    Type string `yaml:"type" json:"type"` // "storage", "search", "notifier", "captcha"
    Name string `yaml:"name" json:"name"` // registration key in Provider Registry
}

// RouteDecl declares an API route the plugin wants to register.
type RouteDecl struct {
    Method string `yaml:"method" json:"method"`
    Path   string `yaml:"path" json:"path"`
}
```

- [ ] Define GORM model in `backend/internal/model/plugin.go`:

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// Plugin stores installed plugin state in the database.
type Plugin struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    PluginID    string         `gorm:"uniqueIndex;size:100;not null" json:"pluginId"`
    Name        string         `gorm:"size:200;not null" json:"name"`
    NameZh      string         `gorm:"size:200" json:"nameZh"`
    Version     string         `gorm:"size:50;not null" json:"version"`
    Description string         `gorm:"size:1000" json:"description"`
    Author      string         `gorm:"size:200" json:"author"`
    License     string         `gorm:"size:100" json:"license"`
    Homepage    string         `gorm:"size:500" json:"homepage"`
    State       string         `gorm:"size:20;not null;default:'installed'" json:"state"`
    Source      string         `gorm:"size:20;not null;default:'local'" json:"source"` // local, marketplace
    BinaryPath  string         `gorm:"size:500" json:"binaryPath"`
    Permissions JSONStringSlice `gorm:"type:text" json:"permissions"`
    Settings    JSONMap        `gorm:"type:text" json:"settings"`
    ErrorMsg    string         `gorm:"size:2000" json:"errorMsg,omitempty"`
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Plugin) TableName() string { return "plugins" }

// PluginSetting stores per-plugin configuration key-value pairs.
type PluginSetting struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    PluginID string `gorm:"index;size:100;not null" json:"pluginId"`
    Key      string `gorm:"size:200;not null" json:"key"`
    Value    string `gorm:"type:text" json:"value"`
}

func (PluginSetting) TableName() string { return "plugin_settings" }
```

- [ ] Write unit tests in `types_test.go` verifying AllPermissions() returns expected count and values
- [ ] Verify: `cd backend && go build ./internal/plugin/... && go test -v -race ./internal/plugin/...`

### Task 2: Define plugin.yaml spec and parser (3.1.1, 3.1.2)

**Files:**
- Create: `backend/internal/plugin/manifest.go`
- Create: `backend/internal/plugin/manifest_test.go`

**Steps:**
- [ ] Add `gopkg.in/yaml.v3` to go.mod (already present as indirect dep — promote to direct)
- [ ] Implement `LoadManifest(dir string) (*PluginMeta, error)` — reads `plugin.yaml` from a plugin directory
- [ ] Implement `ValidateManifest(meta *PluginMeta) error` — validates required fields:
  - `id` must match `^[a-z][a-z0-9-]{2,49}$`
  - `version` must be valid semver
  - `name` must be non-empty
  - `permissions` must all be from AllPermissions()
  - `providers[].type` must be one of: storage, search, notifier, captcha
- [ ] Document the standard plugin directory structure:

```
my-plugin/
├── plugin.yaml          # required: metadata manifest
├── main.go              # required: plugin entry point (go-plugin Serve)
├── go.mod               # required: Go module
├── frontend/            # optional: frontend extension bundle
│   ├── index.ts         # entry point registered via SDK
│   └── dist/            # built JS bundle (shipped with plugin)
│       └── index.js
├── static/              # optional: static assets served at /plugins/{id}/static/
└── README.md            # optional: documentation
```

- [ ] Write tests:
  - Parse valid plugin.yaml with all fields
  - Parse minimal plugin.yaml (only required fields)
  - Reject invalid id format
  - Reject invalid semver
  - Reject unknown permission
  - Reject missing required fields
  - Handle missing plugin.yaml file gracefully
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

### Task 3: Write a sample plugin.yaml for reference (3.1.2)

**Files:**
- Create: `docs/plugin-spec.yaml` (example/reference, not a real plugin)

**Steps:**
- [ ] Create a fully-documented example `plugin.yaml`:

```yaml
# plugin.yaml — Impress CMS Plugin Manifest
# This file defines a plugin's identity, capabilities, and requirements.

id: s3-storage
name: S3 Storage
nameZh: S3 对象存储
version: 1.0.0
description: Store media files in Amazon S3, MinIO, or Alibaba Cloud OSS
author: Impress Team
license: MIT
homepage: https://github.com/user/impress-plugin-s3

# Minimum Impress version required
minAppVersion: 1.0.0

# Other plugins this plugin depends on
dependencies: []

# Permissions this plugin requires (user confirms at install time)
permissions:
  - network:outbound
  - filesystem:read

# Provider interfaces implemented
providers:
  - type: storage
    name: s3-storage

# Custom API routes (served under /api/v1/plugins/{id}/...)
routes:
  - method: GET
    path: /buckets
  - method: POST
    path: /test-connection

# Frontend JS bundle entry (relative path in plugin dir)
frontendEntry: frontend/dist/index.js

# Settings schema (JSON Schema) for auto-generated settings UI
settingsSchema:
  type: object
  required:
    - endpoint
    - bucket
    - accessKeyId
    - secretAccessKey
  properties:
    endpoint:
      type: string
      title: S3 Endpoint
      description: S3-compatible endpoint URL
    bucket:
      type: string
      title: Bucket Name
    region:
      type: string
      title: Region
      default: us-east-1
    accessKeyId:
      type: string
      title: Access Key ID
    secretAccessKey:
      type: string
      title: Secret Access Key
      format: password
    pathStyle:
      type: boolean
      title: Use Path-Style URLs
      default: false
```

- [ ] No automated test needed for this doc file

---

## Chunk 2: Plugin Lifecycle State Machine (Task 3.1.3)

### Task 4: Implement plugin lifecycle manager (3.1.3)

**Files:**
- Create: `backend/internal/plugin/lifecycle.go`
- Create: `backend/internal/plugin/lifecycle_test.go`
- Create: `backend/internal/plugin/store.go`
- Create: `backend/internal/plugin/store_test.go`

**Steps:**
- [ ] Implement state machine in `lifecycle.go`:

```go
package plugin

import "fmt"

// Valid state transitions:
//   installed -> enabled
//   installed -> disabled  (skip enable, keep installed but off)
//   enabled   -> disabled
//   disabled  -> enabled
//   disabled  -> uninstalled (removed)
//   installed -> uninstalled
//   failed    -> disabled   (retry after fixing)
//   failed    -> uninstalled
//   *         -> failed     (any state can transition to failed on error)

var validTransitions = map[PluginState][]PluginState{
    StateInstalled: {StateEnabled, StateDisabled, StateFailed},
    StateEnabled:   {StateDisabled, StateFailed},
    StateDisabled:  {StateEnabled, StateFailed},
    StateFailed:    {StateDisabled, StateFailed},
}

// CanTransition checks if a state transition is valid.
func CanTransition(from, to PluginState) bool {
    targets, ok := validTransitions[from]
    if !ok {
        return false
    }
    for _, t := range targets {
        if t == to {
            return true
        }
    }
    return false
}

// Transition validates and returns the new state or an error.
func Transition(from, to PluginState) error {
    if !CanTransition(from, to) {
        return fmt.Errorf("invalid plugin state transition: %s -> %s", from, to)
    }
    return nil
}
```

- [ ] Implement `store.go` — GORM-backed persistence for Plugin model:

```go
package plugin

import (
    "context"
    "blotting-consultancy/internal/model"
    "gorm.io/gorm"
)

type Store struct {
    db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
    return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, p *model.Plugin) error
func (s *Store) GetByID(ctx context.Context, pluginID string) (*model.Plugin, error)
func (s *Store) List(ctx context.Context) ([]model.Plugin, error)
func (s *Store) UpdateState(ctx context.Context, pluginID string, state PluginState, errMsg string) error
func (s *Store) UpdateSettings(ctx context.Context, pluginID string, settings map[string]any) error
func (s *Store) Delete(ctx context.Context, pluginID string) error
```

- [ ] Write lifecycle tests:
  - All valid transitions succeed
  - Invalid transitions (enabled->installed, disabled->installed) return error
  - Failed state can transition to disabled (retry path)
  - Any state can transition to failed
- [ ] Write store tests (using in-memory SQLite):
  - Create + GetByID round-trip
  - List returns all plugins
  - UpdateState changes state
  - Delete removes plugin
  - GetByID returns error for non-existent plugin
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

---

## Chunk 3: Plugin Runtime with HashiCorp go-plugin (Tasks 3.1.4, 3.1.5b, 3.1.6, 3.1.7)

> Per project guidance: skip WASM POC (3.1.4), skip standalone gRPC sidecar POC (3.1.5). Go directly to HashiCorp go-plugin as the chosen runtime. Tasks 3.1.4/3.1.5/3.1.5b/3.1.6 are collapsed into this chunk.

### Task 5: Add go-plugin and gRPC dependencies (3.1.5b)

**Files:**
- Modify: `backend/go.mod`

**Steps:**
- [ ] Add dependencies:
```bash
cd backend && go get github.com/hashicorp/go-plugin
cd backend && go get google.golang.org/grpc
cd backend && go get google.golang.org/protobuf
```
- [ ] Verify: `cd backend && go build ./...`

### Task 6: Define protobuf service for plugin communication (3.1.5b)

**Files:**
- Create: `backend/internal/plugin/proto/plugin.proto`

**Steps:**
- [ ] Install protoc and Go gRPC plugins if not present:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
- [ ] Define the protobuf service. This is the core contract between host and plugin:

```protobuf
syntax = "proto3";
package plugin;
option go_package = "blotting-consultancy/internal/plugin/proto";

// ProviderService is the main plugin service interface.
// Each plugin process serves this via go-plugin.
service ProviderService {
  // Lifecycle
  rpc Initialize(InitRequest) returns (InitResponse);
  rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);

  // StorageProvider methods
  rpc StorageSave(StorageSaveRequest) returns (StorageSaveResponse);
  rpc StorageGet(StorageGetRequest) returns (stream StorageChunk);
  rpc StorageDelete(StorageDeleteRequest) returns (StorageDeleteResponse);
  rpc StorageURL(StorageURLRequest) returns (StorageURLResponse);
  rpc StorageExists(StorageExistsRequest) returns (StorageExistsResponse);

  // SearchProvider methods
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc SearchSuggest(SearchSuggestRequest) returns (SearchSuggestResponse);
  rpc SearchIndex(SearchIndexRequest) returns (SearchIndexResponse);
  rpc SearchRemove(SearchRemoveRequest) returns (SearchRemoveResponse);
  rpc SearchRebuild(SearchRebuildRequest) returns (SearchRebuildResponse);

  // NotifierProvider methods
  rpc Notify(NotifyRequest) returns (NotifyResponse);

  // CaptchaProvider methods
  rpc CaptchaVerify(CaptchaVerifyRequest) returns (CaptchaVerifyResponse);

  // Custom HTTP route handling
  rpc HandleHTTP(HTTPRequest) returns (HTTPResponse);
}

message InitRequest {
  map<string, string> settings = 1;
  string data_dir = 2;
  string plugin_id = 3;
}
message InitResponse {
  bool success = 1;
  string error = 2;
}

message ShutdownRequest {}
message ShutdownResponse {}

// Storage messages
message StorageSaveRequest {
  string filename = 1;
  bytes data = 2;
  int64 size = 3;
}
message StorageSaveResponse {
  string path = 1;
  string error = 2;
}
message StorageGetRequest { string path = 1; }
message StorageChunk { bytes data = 1; }
message StorageDeleteRequest { string path = 1; }
message StorageDeleteResponse { string error = 1; }
message StorageURLRequest { string path = 1; }
message StorageURLResponse { string url = 1; }
message StorageExistsRequest { string path = 1; }
message StorageExistsResponse { bool exists = 1; string error = 2; }

// Search messages
message SearchRequest {
  string query = 1;
  string locale = 2;
  string content_type = 3;
  int32 page = 4;
  int32 page_size = 5;
}
message SearchResponse {
  repeated SearchResult results = 1;
  int64 total = 2;
  string error = 3;
}
message SearchResult {
  uint64 id = 1;
  string type = 2;
  string title = 3;
  string snippet = 4;
  string url = 5;
  string locale = 6;
  double score = 7;
}
message SearchSuggestRequest {
  string prefix = 1;
  string locale = 2;
  int32 limit = 3;
}
message SearchSuggestResponse {
  repeated string suggestions = 1;
  string error = 2;
}
message SearchIndexRequest {
  string content_type = 1; // "article" or "page"
  uint64 id = 2;
  string locale = 3;
  string title = 4;
  string body = 5;
  string slug = 6;
}
message SearchIndexResponse { string error = 1; }
message SearchRemoveRequest {
  string content_type = 1;
  uint64 id = 2;
}
message SearchRemoveResponse { string error = 1; }
message SearchRebuildRequest {}
message SearchRebuildResponse { string error = 1; }

// Notifier messages
message NotifyRequest {
  string type = 1;
  string subject = 2;
  string body = 3;
  map<string, string> meta = 4;
}
message NotifyResponse { string error = 1; }

// Captcha messages
message CaptchaVerifyRequest {
  string token = 1;
  string remote_ip = 2;
}
message CaptchaVerifyResponse { string error = 1; }

// HTTP proxy messages (for custom plugin routes)
message HTTPRequest {
  string method = 1;
  string path = 2;
  map<string, string> headers = 3;
  bytes body = 4;
  map<string, string> query_params = 5;
}
message HTTPResponse {
  int32 status_code = 1;
  map<string, string> headers = 2;
  bytes body = 3;
}
```

- [ ] Generate Go code: `cd backend && protoc --go_out=. --go-grpc_out=. internal/plugin/proto/plugin.proto`
- [ ] Verify generated files compile: `cd backend && go build ./internal/plugin/proto/...`

### Task 7: Implement go-plugin host-side integration (3.1.5b, 3.1.7)

**Files:**
- Create: `backend/internal/plugin/grpc_host.go`
- Create: `backend/internal/plugin/grpc_host_test.go`
- Create: `backend/internal/plugin/shared/interface.go`

**Steps:**
- [ ] Define the go-plugin handshake in `shared/interface.go`:

```go
package shared

import (
    "github.com/hashicorp/go-plugin"
)

// Handshake is the handshake config for Impress plugins.
// Changing this breaks all existing plugins — do so only on major versions.
var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "IMPRESS_PLUGIN",
    MagicCookieValue: "impress-cms-v1",
}

// PluginMap is the map of plugin types the host expects.
const ProviderPluginName = "provider"
```

- [ ] Implement `grpc_host.go` — the host-side plugin client:

```go
package plugin

import (
    "context"
    "fmt"
    "os/exec"
    "sync"

    goplugin "github.com/hashicorp/go-plugin"
    "blotting-consultancy/internal/plugin/shared"
    pb "blotting-consultancy/internal/plugin/proto"
    "blotting-consultancy/internal/provider"
)

// GRPCHost manages a single plugin process via go-plugin.
type GRPCHost struct {
    meta     *PluginMeta
    client   *goplugin.Client
    provider pb.ProviderServiceClient
    mu       sync.Mutex
}

// NewGRPCHost creates a host for a plugin binary.
func NewGRPCHost(meta *PluginMeta, binaryPath string) *GRPCHost

// Start launches the plugin process and establishes gRPC connection.
func (h *GRPCHost) Start(settings map[string]string) error

// Stop gracefully shuts down the plugin process.
func (h *GRPCHost) Stop() error

// AsStorageProvider returns a StorageProvider that proxies to the plugin.
func (h *GRPCHost) AsStorageProvider() provider.StorageProvider

// AsSearchProvider returns a SearchProvider that proxies to the plugin.
func (h *GRPCHost) AsSearchProvider() provider.SearchProvider

// AsNotifierProvider returns a NotifierProvider that proxies to the plugin.
func (h *GRPCHost) AsNotifierProvider() provider.NotifierProvider

// AsCaptchaProvider returns a CaptchaProvider that proxies to the plugin.
func (h *GRPCHost) AsCaptchaProvider() provider.CaptchaProvider

// HandleHTTP proxies an HTTP request to the plugin.
func (h *GRPCHost) HandleHTTP(ctx context.Context, req *pb.HTTPRequest) (*pb.HTTPResponse, error)

// IsRunning checks if the plugin process is alive.
func (h *GRPCHost) IsRunning() bool
```

- [ ] Implement the provider proxy types (e.g., `grpcStorageProxy` that implements `provider.StorageProvider` by calling gRPC methods)
- [ ] Write tests using a mock gRPC server (no real plugin binary needed):
  - Start/Stop lifecycle
  - StorageProvider proxy correctly marshals/unmarshals
  - NotifierProvider proxy handles errors
  - HandleHTTP returns correct status codes
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

### Task 8: Implement PluginManager (3.1.7)

**Files:**
- Create: `backend/internal/plugin/manager.go`
- Create: `backend/internal/plugin/manager_test.go`

**Steps:**
- [ ] Implement `PluginManager` — the central coordinator:

```go
package plugin

import (
    "context"
    "path/filepath"
    "sync"

    "blotting-consultancy/internal/eventbus"
    "blotting-consultancy/internal/provider"
)

// ManagerConfig holds configuration for the plugin manager.
type ManagerConfig struct {
    PluginDir string // directory where plugins are installed (default: ./plugins)
    DataDir   string // directory for plugin data storage (default: ./data/plugins)
}

// Manager orchestrates plugin discovery, lifecycle, and provider registration.
type Manager struct {
    config    ManagerConfig
    store     *Store
    registry  *provider.Registry
    bus       eventbus.EventBus
    hosts     map[string]*GRPCHost // pluginID -> running host
    mu        sync.RWMutex
}

func NewManager(cfg ManagerConfig, store *Store, registry *provider.Registry, bus eventbus.EventBus) *Manager

// DiscoverPlugins scans PluginDir for plugin directories with valid plugin.yaml.
func (m *Manager) DiscoverPlugins(ctx context.Context) ([]PluginMeta, error)

// InstallPlugin installs a plugin from a directory path.
// 1. Parse plugin.yaml
// 2. Validate manifest
// 3. Build binary (go build) or verify pre-built binary exists
// 4. Create DB record with state=installed
func (m *Manager) InstallPlugin(ctx context.Context, dir string) (*PluginMeta, error)

// EnablePlugin starts a plugin process and registers its providers.
// 1. Validate state transition (installed/disabled -> enabled)
// 2. Start GRPCHost
// 3. Call Initialize RPC with settings
// 4. Register providers in Provider Registry
// 5. Update DB state to enabled
func (m *Manager) EnablePlugin(ctx context.Context, pluginID string) error

// DisablePlugin stops a plugin process and unregisters its providers.
func (m *Manager) DisablePlugin(ctx context.Context, pluginID string) error

// UninstallPlugin disables (if needed) and removes a plugin.
func (m *Manager) UninstallPlugin(ctx context.Context, pluginID string) error

// GetPlugin returns the current state of a plugin.
func (m *Manager) GetPlugin(ctx context.Context, pluginID string) (*model.Plugin, error)

// ListPlugins returns all installed plugins.
func (m *Manager) ListPlugins(ctx context.Context) ([]model.Plugin, error)

// UpdateSettings updates a plugin's settings and re-initializes if enabled.
func (m *Manager) UpdateSettings(ctx context.Context, pluginID string, settings map[string]any) error

// HandlePluginHTTP routes an HTTP request to the correct plugin.
func (m *Manager) HandlePluginHTTP(ctx context.Context, pluginID string, req *pb.HTTPRequest) (*pb.HTTPResponse, error)

// StartEnabledPlugins starts all plugins that were enabled before shutdown.
// Called during server startup.
func (m *Manager) StartEnabledPlugins(ctx context.Context) error

// StopAll gracefully stops all running plugins.
// Called during server shutdown.
func (m *Manager) StopAll() error
```

- [ ] Write tests:
  - DiscoverPlugins finds plugins in a temp directory
  - InstallPlugin creates DB record
  - EnablePlugin/DisablePlugin state transitions
  - UninstallPlugin cleans up
  - StartEnabledPlugins re-enables previously enabled plugins
  - StopAll stops all hosts
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

---

## Chunk 4: Plugin Permission & Sandbox (Task 3.1.8)

### Task 9: Implement permission model and sandbox (3.1.8)

**Files:**
- Create: `backend/internal/plugin/sandbox.go`
- Create: `backend/internal/plugin/sandbox_test.go`

**Steps:**
- [ ] Implement permission checking:

```go
package plugin

import "fmt"

// Sandbox enforces permissions declared in a plugin's manifest.
type Sandbox struct {
    pluginID    string
    permissions map[Permission]bool
}

func NewSandbox(pluginID string, perms []Permission) *Sandbox

// Check returns nil if the plugin has the requested permission,
// or an error describing the denied capability.
func (s *Sandbox) Check(perm Permission) error

// CheckAll returns nil if all permissions are granted.
func (s *Sandbox) CheckAll(perms ...Permission) error

// RequiredPermissionsForProvider returns the minimum permissions
// needed to implement a given provider type.
func RequiredPermissionsForProvider(providerType string) []Permission {
    switch providerType {
    case "storage":
        return []Permission{PermNetworkOutbound} // most storage plugins need network
    case "search":
        return []Permission{PermNetworkOutbound}
    case "notifier":
        return []Permission{PermNetworkOutbound}
    case "captcha":
        return []Permission{PermNetworkOutbound}
    default:
        return nil
    }
}

// ValidateManifestPermissions checks that a plugin's declared permissions
// are sufficient for its declared providers and routes.
func ValidateManifestPermissions(meta *PluginMeta) error
```

- [ ] Integrate sandbox checks into `Manager.EnablePlugin()`:
  - Before registering a provider, verify the plugin declared necessary permissions
  - Before registering routes, verify `PermRouteRegister`
  - Before allowing event subscription, verify `PermEventSubscribe`
- [ ] Write tests:
  - Plugin with all required permissions passes validation
  - Plugin missing a required permission is rejected
  - Plugin declaring routes without `PermRouteRegister` is rejected
  - Storage plugin without `PermNetworkOutbound` is rejected (warning only for local storage)
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

---

## Chunk 5: Go Plugin SDK (Tasks 3.2.1 — 3.2.3)

### Task 10: Create SDK Go module (3.2.1)

**Files:**
- Create: `backend/sdk/go.mod`
- Create: `backend/sdk/plugin.go`
- Create: `backend/sdk/context.go`
- Create: `backend/sdk/plugin_test.go`

**Steps:**
- [ ] Create `backend/sdk/` directory with its own go.mod:

```bash
cd backend/sdk && go mod init blotting-consultancy/sdk
```

- [ ] Implement SDK entry point in `plugin.go`:

```go
package sdk

import (
    goplugin "github.com/hashicorp/go-plugin"
    "blotting-consultancy/internal/plugin/shared"
)

// Plugin is the main struct plugin authors embed and configure.
type Plugin struct {
    meta           PluginMeta
    storageImpl    StorageProvider
    searchImpl     SearchProvider
    notifierImpl   NotifierProvider
    captchaImpl    CaptchaProvider
    httpHandler    HTTPHandler
    onInit         func(ctx *PluginContext) error
    onShutdown     func() error
}

// New creates a new plugin builder.
func New(id, name, version string) *Plugin

// WithStorage registers a StorageProvider implementation.
func (p *Plugin) WithStorage(impl StorageProvider) *Plugin

// WithSearch registers a SearchProvider implementation.
func (p *Plugin) WithSearch(impl SearchProvider) *Plugin

// WithNotifier registers a NotifierProvider implementation.
func (p *Plugin) WithNotifier(impl NotifierProvider) *Plugin

// WithCaptcha registers a CaptchaProvider implementation.
func (p *Plugin) WithCaptcha(impl CaptchaProvider) *Plugin

// WithHTTPHandler registers a handler for custom HTTP routes.
func (p *Plugin) WithHTTPHandler(h HTTPHandler) *Plugin

// OnInit registers a callback called when the plugin is initialized.
func (p *Plugin) OnInit(fn func(ctx *PluginContext) error) *Plugin

// OnShutdown registers a callback called when the plugin is shutting down.
func (p *Plugin) OnShutdown(fn func() error) *Plugin

// Serve starts the plugin process and blocks.
// This is the last call in a plugin's main().
func (p *Plugin) Serve() {
    goplugin.Serve(&goplugin.ServeConfig{
        HandshakeConfig: shared.Handshake,
        Plugins: map[string]goplugin.Plugin{
            shared.ProviderPluginName: &GRPCProviderPlugin{Impl: p},
        },
        GRPCServer: goplugin.DefaultGRPCServer,
    })
}
```

- [ ] Implement `PluginContext` in `context.go`:

```go
package sdk

// PluginContext provides access to host capabilities during plugin initialization.
type PluginContext struct {
    PluginID string
    DataDir  string            // writable directory for plugin data
    Settings map[string]string // user-configured settings from plugin.yaml schema
}

// GetSetting returns a setting value or the default.
func (c *PluginContext) GetSetting(key, defaultValue string) string
```

- [ ] Implement the gRPC server side (`GRPCProviderPlugin`) that bridges SDK method calls to protobuf:

```go
// GRPCProviderPlugin implements go-plugin's GRPCPlugin interface.
type GRPCProviderPlugin struct {
    goplugin.Plugin
    Impl *Plugin
}

func (p *GRPCProviderPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error
func (p *GRPCProviderPlugin) GRPCClient(broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error)
```

- [ ] Write tests:
  - Plugin builder creates valid plugin with all providers
  - Serve() panics if no providers are registered (guard)
  - PluginContext.GetSetting returns default for missing keys
- [ ] Verify: `cd backend/sdk && go test -v -race ./...`

### Task 11: Implement plugin database access (3.2.2)

**Files:**
- Create: `backend/sdk/database.go`

**Steps:**
- [ ] Plugin DB access uses a **separate table prefix** approach. Each plugin gets tables prefixed with `plugin_{id}_`. The SDK provides migration helpers:

```go
package sdk

// TableName returns the plugin-scoped table name.
// Example: for plugin "webhook" and table "subscriptions" -> "plugin_webhook_subscriptions"
func (c *PluginContext) TableName(name string) string {
    return fmt.Sprintf("plugin_%s_%s", c.PluginID, name)
}

// DB access is provided via gRPC calls back to the host process.
// The host exposes a DatabaseService that plugins can call.
// This avoids giving plugins direct DB credentials.

type DatabaseService interface {
    // Exec runs a write query (INSERT/UPDATE/DELETE) on plugin-scoped tables.
    Exec(ctx context.Context, query string, args ...interface{}) (int64, error)
    // Query runs a read query on plugin-scoped tables.
    Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
    // Migrate creates/updates plugin tables.
    Migrate(ctx context.Context, tableName string, schema map[string]string) error
}
```

- [ ] NOTE: For the initial implementation, plugins that need DB access should use their own embedded SQLite or connect to an external DB. The host-mediated DB access via gRPC is a future enhancement (marked in the proto as a TODO). This keeps the initial scope manageable.
- [ ] Write tests:
  - TableName generates correct prefixed names
  - Plugin ID with hyphens is sanitized in table names
- [ ] Verify: `cd backend/sdk && go test -v -race ./...`

### Task 12: Implement plugin route registration (3.2.3)

**Files:**
- Create: `backend/sdk/routes.go`
- Modify: `backend/cmd/server/main.go` — add plugin route proxy

**Steps:**
- [ ] Implement route declaration in the SDK:

```go
package sdk

// HTTPHandler handles incoming HTTP requests proxied from the host.
type HTTPHandler interface {
    HandleHTTP(req *HTTPRequest) *HTTPResponse
}

// HTTPRequest mirrors the protobuf HTTPRequest for plugin authors.
type HTTPRequest struct {
    Method      string
    Path        string
    Headers     map[string]string
    Body        []byte
    QueryParams map[string]string
}

// HTTPResponse mirrors the protobuf HTTPResponse.
type HTTPResponse struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
}

// SimpleRouter is a basic HTTP router for plugin authors.
type SimpleRouter struct {
    routes map[string]map[string]func(*HTTPRequest) *HTTPResponse // method -> path -> handler
}

func NewRouter() *SimpleRouter
func (r *SimpleRouter) GET(path string, handler func(*HTTPRequest) *HTTPResponse)
func (r *SimpleRouter) POST(path string, handler func(*HTTPRequest) *HTTPResponse)
func (r *SimpleRouter) PUT(path string, handler func(*HTTPRequest) *HTTPResponse)
func (r *SimpleRouter) DELETE(path string, handler func(*HTTPRequest) *HTTPResponse)
func (r *SimpleRouter) HandleHTTP(req *HTTPRequest) *HTTPResponse
```

- [ ] Add plugin route proxy to main.go:

```go
// In main.go, after PluginManager is created:
// Plugin API routes: /api/v1/plugins/:pluginId/*path
pluginGroup := router.Group("/api/v1/plugins")
pluginGroup.Any("/:pluginId/*path", func(c *gin.Context) {
    pluginID := c.Param("pluginId")
    path := c.Param("path")
    // Build HTTPRequest from gin.Context
    // Call pluginManager.HandlePluginHTTP(ctx, pluginID, req)
    // Write HTTPResponse back to gin.Context
})
```

- [ ] Write tests:
  - SimpleRouter correctly dispatches GET/POST/PUT/DELETE
  - Unknown route returns 404
  - Plugin route proxy in main.go correctly forwards requests
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/... && cd backend/sdk && go test -v -race ./...`

---

## Chunk 6: Frontend Extension System (Tasks 3.2.4 — 3.2.6)

### Task 13: Implement frontend ExtensionSlot system (3.2.4)

**Files:**
- Create: `frontend/src/plugin/types.ts`
- Create: `frontend/src/plugin/PluginRegistry.ts`
- Create: `frontend/src/plugin/PluginRegistry.test.ts`
- Create: `frontend/src/plugin/ExtensionSlot.tsx`
- Create: `frontend/src/plugin/ExtensionSlot.test.tsx`
- Create: `frontend/src/plugin/index.ts`

**Steps:**
- [ ] Define extension point types in `types.ts`:

```typescript
export type ExtensionPoint =
  | "admin:navbar"       // items in the admin navigation bar
  | "admin:sidebar"      // items in the admin sidebar
  | "admin:settings"     // tabs in the settings page
  | "admin:dashboard"    // widgets on the admin dashboard
  | "public:head"        // script/style injection in <head>
  | "public:body-end"    // script injection before </body>
  | "public:footer"      // components in the footer
  | "article:after"      // components after article content
  | "page:after";        // components after page content

export interface ExtensionRegistration {
  pluginId: string;
  point: ExtensionPoint;
  component: React.ComponentType<any>;
  priority?: number; // lower = renders first, default 100
  props?: Record<string, unknown>;
}

export interface FrontendPlugin {
  id: string;
  name: string;
  version: string;
  register: (registry: PluginRegistryAPI) => void;
}

export interface PluginRegistryAPI {
  registerExtension(point: ExtensionPoint, component: React.ComponentType<any>, priority?: number): void;
  getSettings(): Record<string, unknown>;
}
```

- [ ] Implement `PluginRegistry.ts`:

```typescript
import type { ExtensionPoint, ExtensionRegistration, FrontendPlugin } from "./types";

class PluginRegistryImpl {
  private extensions = new Map<ExtensionPoint, ExtensionRegistration[]>();
  private plugins = new Map<string, FrontendPlugin>();

  registerPlugin(plugin: FrontendPlugin, settings: Record<string, unknown>): void {
    this.plugins.set(plugin.id, plugin);
    const api: PluginRegistryAPI = {
      registerExtension: (point, component, priority = 100) => {
        const reg: ExtensionRegistration = {
          pluginId: plugin.id,
          point,
          component,
          priority,
        };
        const existing = this.extensions.get(point) || [];
        existing.push(reg);
        existing.sort((a, b) => (a.priority ?? 100) - (b.priority ?? 100));
        this.extensions.set(point, existing);
      },
      getSettings: () => settings,
    };
    plugin.register(api);
  }

  getExtensions(point: ExtensionPoint): ExtensionRegistration[] {
    return this.extensions.get(point) || [];
  }

  unregisterPlugin(pluginId: string): void {
    this.plugins.delete(pluginId);
    for (const [point, regs] of this.extensions) {
      this.extensions.set(point, regs.filter(r => r.pluginId !== pluginId));
    }
  }

  clear(): void {
    this.extensions.clear();
    this.plugins.clear();
  }
}

export const pluginRegistry = new PluginRegistryImpl();
```

- [ ] Implement `ExtensionSlot.tsx`:

```tsx
import { Suspense } from "react";
import type { ExtensionPoint } from "./types";
import { pluginRegistry } from "./PluginRegistry";

interface ExtensionSlotProps {
  point: ExtensionPoint;
  fallback?: React.ReactNode;
  props?: Record<string, unknown>;
}

export function ExtensionSlot({ point, fallback, props }: ExtensionSlotProps) {
  const extensions = pluginRegistry.getExtensions(point);

  if (extensions.length === 0) {
    return fallback ? <>{fallback}</> : null;
  }

  return (
    <Suspense fallback={<div className="animate-pulse h-4 bg-gray-200 rounded" />}>
      {extensions.map((ext, i) => {
        const Component = ext.component;
        return <Component key={`${ext.pluginId}-${i}`} {...props} {...ext.props} />;
      })}
    </Suspense>
  );
}
```

- [ ] Write barrel export in `index.ts`
- [ ] Write tests:
  - PluginRegistry: register a plugin, getExtensions returns its components
  - PluginRegistry: unregisterPlugin removes all its extensions
  - PluginRegistry: priority ordering works
  - ExtensionSlot: renders nothing when no extensions registered
  - ExtensionSlot: renders registered components
  - ExtensionSlot: renders fallback when no extensions
- [ ] Verify: `cd frontend && pnpm test -- src/plugin/`

### Task 14: Implement plugin JS bundle loader (3.2.4)

**Files:**
- Create: `frontend/src/plugin/PluginLoader.ts`
- Create: `frontend/src/plugin/PluginLoader.test.ts`

**Steps:**
- [ ] Implement dynamic plugin JS loading:

```typescript
import type { FrontendPlugin } from "./types";
import { pluginRegistry } from "./PluginRegistry";

interface PluginManifest {
  pluginId: string;
  name: string;
  version: string;
  frontendEntry: string; // URL to the JS bundle
  settings: Record<string, unknown>;
  enabled: boolean;
}

// loadPluginBundle dynamically imports a plugin's JS bundle.
export async function loadPluginBundle(manifest: PluginManifest): Promise<void> {
  if (!manifest.enabled || !manifest.frontendEntry) return;

  try {
    // Dynamic import of the plugin's JS bundle
    // The bundle must export a default FrontendPlugin object
    const module = await import(/* @vite-ignore */ manifest.frontendEntry);
    const plugin: FrontendPlugin = module.default;

    if (!plugin || !plugin.id || !plugin.register) {
      console.error(`Invalid plugin bundle for ${manifest.pluginId}: missing id or register`);
      return;
    }

    pluginRegistry.registerPlugin(plugin, manifest.settings);
    console.log(`Plugin loaded: ${plugin.id}@${plugin.version}`);
  } catch (err) {
    console.error(`Failed to load plugin ${manifest.pluginId}:`, err);
  }
}

// loadAllPlugins fetches the plugin list from the API and loads enabled plugins.
export async function loadAllPlugins(): Promise<void> {
  try {
    const response = await fetch("/api/v1/plugins?enabled=true");
    if (!response.ok) return;
    const plugins: PluginManifest[] = await response.json();

    await Promise.allSettled(
      plugins.filter(p => p.frontendEntry).map(loadPluginBundle)
    );
  } catch (err) {
    console.error("Failed to load plugins:", err);
  }
}
```

- [ ] Write tests:
  - loadPluginBundle handles missing frontendEntry gracefully
  - loadPluginBundle handles import failure gracefully
  - loadAllPlugins fetches and loads enabled plugins
- [ ] Verify: `cd frontend && pnpm test -- src/plugin/`

### Task 15: Implement plugin settings panel (3.2.5)

**Files:**
- Create: `frontend/src/plugin/PluginSettingsForm.tsx`
- Create: `frontend/src/plugin/PluginSettingsForm.test.tsx`

**Steps:**
- [ ] Implement a JSON Schema-driven form generator:

```tsx
import { useState } from "react";

interface JSONSchemaProperty {
  type: "string" | "number" | "boolean" | "integer";
  title?: string;
  description?: string;
  default?: unknown;
  format?: "password" | "email" | "url" | "textarea";
  enum?: string[];
}

interface JSONSchema {
  type: "object";
  required?: string[];
  properties: Record<string, JSONSchemaProperty>;
}

interface PluginSettingsFormProps {
  schema: JSONSchema;
  values: Record<string, unknown>;
  onSave: (values: Record<string, unknown>) => Promise<void>;
  loading?: boolean;
}

export function PluginSettingsForm({ schema, values, onSave, loading }: PluginSettingsFormProps) {
  const [formValues, setFormValues] = useState<Record<string, unknown>>(values);
  const [saving, setSaving] = useState(false);

  // Render form fields based on JSON Schema property types:
  // - string -> <input type="text"> (or textarea/password based on format)
  // - number/integer -> <input type="number">
  // - boolean -> <input type="checkbox">
  // - enum -> <select>
  // Required fields get asterisk and validation
  // ...
}
```

- [ ] Keep the implementation simple — support the basic JSON Schema subset (string, number, boolean, enum). Complex schemas (arrays, nested objects) can be added later.
- [ ] Write tests:
  - Renders text input for string properties
  - Renders checkbox for boolean properties
  - Renders select for enum properties
  - Renders password field for format: password
  - Required fields show validation error when empty
  - onSave is called with form values
- [ ] Verify: `cd frontend && pnpm test -- src/plugin/`

### Task 16: Create JS/TS Plugin SDK documentation and types (3.2.6)

**Files:**
- Create: `frontend/src/plugin/sdk.ts` (re-exports for plugin developers)

**Steps:**
- [ ] Create SDK type exports that plugin authors import:

```typescript
// sdk.ts — public API for frontend plugin development
// Plugin authors install @impress/plugin-sdk (or copy this file) and use these types.

export type { ExtensionPoint, FrontendPlugin, PluginRegistryAPI } from "./types";

// Example plugin (for documentation):
//
// import type { FrontendPlugin } from "@impress/plugin-sdk";
//
// const myPlugin: FrontendPlugin = {
//   id: "my-plugin",
//   name: "My Plugin",
//   version: "1.0.0",
//   register(api) {
//     api.registerExtension("admin:dashboard", MyDashboardWidget);
//     api.registerExtension("public:body-end", MyAnalyticsScript);
//   },
// };
//
// export default myPlugin;
```

- [ ] No separate npm package needed initially — plugins can import types from the main frontend build
- [ ] Verify: `cd frontend && pnpm type-check`

---

## Chunk 7: Plugin Admin Handler & CLI (Tasks 3.2.3, 3.2.7)

### Task 17: Implement plugin admin HTTP handler (3.2.3)

**Files:**
- Create: `backend/internal/handler/plugin/handler.go`
- Create: `backend/internal/handler/plugin/handler_test.go`
- Modify: `backend/cmd/server/main.go` — wire plugin routes

**Steps:**
- [ ] Implement handler:

```go
package plugin

import (
    "github.com/gin-gonic/gin"
    pluginpkg "blotting-consultancy/internal/plugin"
)

type Handler struct {
    manager *pluginpkg.Manager
}

func NewHandler(manager *pluginpkg.Manager) *Handler

// RegisterRoutes sets up admin and public plugin routes.
func (h *Handler) RegisterRoutes(admin *gin.RouterGroup, pluginProxy *gin.RouterGroup) {
    // Admin routes (require auth)
    admin.GET("/plugins", h.List)
    admin.GET("/plugins/:id", h.GetByID)
    admin.POST("/plugins/install", h.Install)
    admin.PUT("/plugins/:id/enable", h.Enable)
    admin.PUT("/plugins/:id/disable", h.Disable)
    admin.DELETE("/plugins/:id", h.Uninstall)
    admin.GET("/plugins/:id/settings", h.GetSettings)
    admin.PUT("/plugins/:id/settings", h.UpdateSettings)

    // Plugin API proxy (public, individual plugins handle their own auth)
    pluginProxy.Any("/:pluginId/*path", h.ProxyToPlugin)
}

// @Summary List installed plugins
// @Tags plugins
// @Security BearerAuth
// @Success 200 {array} model.Plugin
// @Router /admin/plugins [get]
func (h *Handler) List(c *gin.Context)

// @Summary Get plugin by ID
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 200 {object} model.Plugin
// @Router /admin/plugins/{id} [get]
func (h *Handler) GetByID(c *gin.Context)

// @Summary Install a plugin
// @Tags plugins
// @Security BearerAuth
// @Param body body InstallRequest true "Install request"
// @Success 201 {object} model.Plugin
// @Router /admin/plugins/install [post]
func (h *Handler) Install(c *gin.Context)

// @Summary Enable a plugin
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 200 {object} model.Plugin
// @Router /admin/plugins/{id}/enable [put]
func (h *Handler) Enable(c *gin.Context)

// @Summary Disable a plugin
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 200 {object} model.Plugin
// @Router /admin/plugins/{id}/disable [put]
func (h *Handler) Disable(c *gin.Context)

// @Summary Uninstall a plugin
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 204
// @Router /admin/plugins/{id} [delete]
func (h *Handler) Uninstall(c *gin.Context)

// @Summary Get plugin settings
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/plugins/{id}/settings [get]
func (h *Handler) GetSettings(c *gin.Context)

// @Summary Update plugin settings
// @Tags plugins
// @Security BearerAuth
// @Param id path string true "Plugin ID"
// @Success 200 {object} model.Plugin
// @Router /admin/plugins/{id}/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context)

// ProxyToPlugin forwards requests to the appropriate plugin process.
func (h *Handler) ProxyToPlugin(c *gin.Context)

type InstallRequest struct {
    Source string `json:"source"` // "local" or "marketplace"
    Path   string `json:"path"`   // local directory path or marketplace URL
}
```

- [ ] Wire into main.go:
  - Add `Plugin` and `PluginSetting` to AutoMigrate list
  - Create PluginManager after provider registry initialization
  - Call `pluginManager.StartEnabledPlugins(ctx)` during startup
  - Call `pluginManager.StopAll()` during shutdown
  - Register plugin admin routes and API proxy route
- [ ] Write handler tests with httptest
- [ ] Verify: `cd backend && go build ./cmd/server/ && go test -v -race ./internal/handler/plugin/...`

### Task 18: Flesh out CLI plugin commands (3.2.7)

**Files:**
- Modify: `backend/cmd/impress/cmd_plugin.go`

**Steps:**
- [ ] Implement full plugin CLI subcommands:

```go
func pluginCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "plugin",
        Short: "Plugin management commands",
    }

    cmd.AddCommand(pluginCreateCmd())  // generate plugin scaffold
    cmd.AddCommand(pluginListCmd())    // list installed plugins
    cmd.AddCommand(pluginEnableCmd())  // enable a plugin
    cmd.AddCommand(pluginDisableCmd()) // disable a plugin
    cmd.AddCommand(pluginBuildCmd())   // build a plugin binary
    cmd.AddCommand(pluginInstallCmd()) // install a plugin from dir/URL

    return cmd
}
```

- [ ] `impress plugin create <name>` — scaffold a new plugin:
  - Create directory structure: `<name>/plugin.yaml`, `<name>/main.go`, `<name>/go.mod`
  - Generate starter plugin.yaml with the plugin name
  - Generate starter main.go with SDK import and Serve() call
  - Generate go.mod with correct module name and SDK dependency
- [ ] `impress plugin build <dir>` — build a plugin binary:
  - Run `go build -o <dir>/plugin <dir>/` inside the plugin directory
- [ ] `impress plugin list` — list installed plugins via API call to `/admin/plugins`
- [ ] `impress plugin enable <id>` / `impress plugin disable <id>` — via API
- [ ] Write tests for `plugin create` (verify generated files)
- [ ] Verify: `cd backend && go build ./cmd/impress/ && go test -v -race ./cmd/impress/...`

---

## Chunk 8: Official Demo Plugins (Tasks 3.3.1 — 3.3.6)

### Task 19: S3 Storage Plugin (3.3.1)

**Files:**
- Create: `plugins/s3-storage/plugin.yaml`
- Create: `plugins/s3-storage/main.go`
- Create: `plugins/s3-storage/provider.go`
- Create: `plugins/s3-storage/go.mod`
- Create: `plugins/s3-storage/provider_test.go`

**Steps:**
- [ ] Create `plugin.yaml` (use the reference from Task 3 above)
- [ ] Implement `provider.go` — S3StorageProvider implementing `sdk.StorageProvider`:

```go
package main

import (
    "context"
    "io"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "blotting-consultancy/sdk"
)

type S3Storage struct {
    client   *s3.Client
    bucket   string
    endpoint string
    region   string
}

func NewS3Storage(ctx *sdk.PluginContext) (*S3Storage, error) {
    // Read settings: endpoint, bucket, region, accessKeyId, secretAccessKey, pathStyle
    // Initialize AWS SDK v2 S3 client
}

func (s *S3Storage) Save(ctx context.Context, filename string, reader io.Reader, size int64) (string, error)
func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error)
func (s *S3Storage) Delete(ctx context.Context, path string) error
func (s *S3Storage) URL(path string) string
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error)
```

- [ ] Implement `main.go`:

```go
package main

import "blotting-consultancy/sdk"

func main() {
    sdk.New("s3-storage", "S3 Storage", "1.0.0").
        WithStorage(&S3Storage{}).
        OnInit(func(ctx *sdk.PluginContext) error {
            storage, err := NewS3Storage(ctx)
            if err != nil {
                return err
            }
            // Replace the storage impl with initialized version
            return nil
        }).
        Serve()
}
```

- [ ] Write unit tests with a mock S3 client (using interface abstraction)
- [ ] Verify: `cd plugins/s3-storage && go build ./ && go test -v -race ./...`

### Task 20: Email Notifier Plugin (3.3.2)

**Files:**
- Create: `plugins/email-notifier/plugin.yaml`
- Create: `plugins/email-notifier/main.go`
- Create: `plugins/email-notifier/provider.go`
- Create: `plugins/email-notifier/go.mod`
- Create: `plugins/email-notifier/provider_test.go`

**Steps:**
- [ ] plugin.yaml: id=email-notifier, provider type=notifier, settings: smtpHost, smtpPort, username, password, fromAddress, fromName
- [ ] Implement SMTP-based NotifierProvider:

```go
type EmailNotifier struct {
    host     string
    port     int
    username string
    password string
    from     string
    fromName string
}

func (e *EmailNotifier) Notify(ctx context.Context, event sdk.NotifyEvent) error {
    // Build email from event.Subject and event.Body
    // Send via net/smtp with TLS
}
```

- [ ] Write tests with a mock SMTP server
- [ ] Verify: `cd plugins/email-notifier && go build ./ && go test -v -race ./...`

### Task 21: Meilisearch Plugin (3.3.3)

**Files:**
- Create: `plugins/meilisearch/plugin.yaml`
- Create: `plugins/meilisearch/main.go`
- Create: `plugins/meilisearch/provider.go`
- Create: `plugins/meilisearch/go.mod`
- Create: `plugins/meilisearch/provider_test.go`

**Steps:**
- [ ] plugin.yaml: id=meilisearch, provider type=search, settings: host, apiKey, indexPrefix
- [ ] Implement Meilisearch-based SearchProvider using the official Go SDK (`github.com/meilisearch/meilisearch-go`)
- [ ] Map provider.SearchResult <-> Meilisearch hits
- [ ] Write tests with mock HTTP server simulating Meilisearch API
- [ ] Verify: `cd plugins/meilisearch && go build ./ && go test -v -race ./...`

### Task 22: CAPTCHA Plugin (3.3.4)

**Files:**
- Create: `plugins/captcha/plugin.yaml`
- Create: `plugins/captcha/main.go`
- Create: `plugins/captcha/provider.go`
- Create: `plugins/captcha/go.mod`
- Create: `plugins/captcha/provider_test.go`

**Steps:**
- [ ] plugin.yaml: id=captcha, provider type=captcha, settings: provider (recaptcha/hcaptcha/turnstile), siteKey, secretKey
- [ ] Implement multi-backend CAPTCHA provider:

```go
type CaptchaPlugin struct {
    verifier Verifier
}

type Verifier interface {
    Verify(ctx context.Context, token string, ip string) error
}

// Implementations for reCAPTCHA v2/v3, hCaptcha, Cloudflare Turnstile
type RecaptchaVerifier struct { secretKey string }
type HCaptchaVerifier struct { secretKey string }
type TurnstileVerifier struct { secretKey string }
```

- [ ] Frontend component: add `siteKey` to settings, inject the appropriate script tag via `public:head` extension slot
- [ ] Write tests with mock verification endpoints
- [ ] Verify: `cd plugins/captcha && go build ./ && go test -v -race ./...`

### Task 23: Webhook Plugin (3.3.5)

**Files:**
- Create: `plugins/webhook/plugin.yaml`
- Create: `plugins/webhook/main.go`
- Create: `plugins/webhook/handler.go`
- Create: `plugins/webhook/go.mod`
- Create: `plugins/webhook/handler_test.go`

**Steps:**
- [ ] plugin.yaml: id=webhook, permissions: [event:subscribe, route:register, network:outbound]
- [ ] No provider implementation — this plugin uses event subscription and custom routes
- [ ] Implement webhook configuration via custom routes:
  - `POST /api/v1/plugins/webhook/subscriptions` — create webhook subscription (URL, events, secret)
  - `GET /api/v1/plugins/webhook/subscriptions` — list subscriptions
  - `DELETE /api/v1/plugins/webhook/subscriptions/:id` — remove subscription
- [ ] Implement event-driven webhook dispatch:
  - Subscribe to all content lifecycle events
  - On event, find matching subscriptions
  - POST event payload to each URL with HMAC signature header
  - Retry with exponential backoff (3 attempts)
- [ ] Store webhook subscriptions in plugin-local SQLite (embedded)
- [ ] Write tests:
  - Webhook dispatch sends correct payload
  - HMAC signature is valid
  - Retry logic works on HTTP 5xx
  - Subscriptions CRUD works
- [ ] Verify: `cd plugins/webhook && go build ./ && go test -v -race ./...`

### Task 24: Analytics Injection Plugin (3.3.6)

**Files:**
- Create: `plugins/analytics/plugin.yaml`
- Create: `plugins/analytics/main.go`
- Create: `plugins/analytics/handler.go`
- Create: `plugins/analytics/go.mod`
- Create: `plugins/analytics/frontend/index.ts`
- Create: `plugins/analytics/handler_test.go`

**Steps:**
- [ ] plugin.yaml: id=analytics, permissions: [frontend:inject, hook:register], settings: provider (google/baidu/umami), trackingId, domain (for umami)
- [ ] Backend: register a BeforeRender hook that injects the analytics `<script>` tag into the HTML `<head>` via the SEO renderer
- [ ] Frontend: register an extension at `public:head` that injects the appropriate analytics script:

```typescript
// For Google Analytics:
// <script async src="https://www.googletagmanager.com/gtag/js?id=GA_ID"></script>
// For Baidu Tongji:
// <script>var _hmt = _hmt || []; ...</script>
// For Umami:
// <script async src="https://analytics.example.com/script.js" data-website-id="ID"></script>
```

- [ ] Write tests for script generation with different providers
- [ ] Verify: `cd plugins/analytics && go build ./ && go test -v -race ./...`

---

## Chunk 9: Theme Marketplace Foundation (Tasks 3.4.1 — 3.4.3)

### Task 25: Upgrade theme package spec (3.4.1)

**Files:**
- Modify: `frontend/src/theme/packages/types.ts`
- Modify: `frontend/src/theme/packages/registry.ts`
- Modify: `frontend/src/theme/packages/default/index.ts`
- Modify: `frontend/src/theme/packages/modern-dark/index.ts`
- Modify: `frontend/src/theme/packages/warm-earth/index.ts`

**Steps:**
- [ ] Extend ThemePackage type to include marketplace fields:

```typescript
export interface ThemePackage {
  // Existing fields
  id: string;
  name: string;
  description: string;
  author: string;
  version: string;
  preview: string; // CSS gradient for preview card
  tokens: ThemeTokens;
  sectionOverrides?: Record<string, ComponentType<SectionProps<any>>>;

  // New marketplace fields
  nameZh?: string;
  descriptionZh?: string;
  license?: string;
  homepage?: string;
  screenshots?: string[];  // URLs to full-page screenshots
  tags?: string[];          // e.g. ["minimal", "blog", "portfolio"]
  minAppVersion?: string;
  source?: "built-in" | "marketplace" | "local";
}
```

- [ ] Update existing theme packages (default, modern-dark, warm-earth) to include the new fields:
  - Add `source: "built-in"`
  - Add `nameZh`, `license: "MIT"`, `tags`
- [ ] Verify: `cd frontend && pnpm type-check && pnpm test`

### Task 26: Implement theme online switching (3.4.2)

**Files:**
- Modify: `frontend/src/theme/ThemeProvider.tsx` (if needed)
- Verify existing `installedThemeHandlerInst.AdminActivate` handles this

**Steps:**
- [ ] Review existing theme activation flow:
  - Backend: `PUT /admin/themes/:id/activate` already exists (`installedThemeHandler.AdminActivate`)
  - Frontend: ThemeProvider loads active theme from API
- [ ] Verify the existing flow works end-to-end:
  - Switch from default to modern-dark via API
  - Page content and configuration are preserved
  - Only theme tokens and section overrides change
- [ ] If gaps exist, fix them. Expected state: theme switching already works via the installed_themes system from Phase 1.
- [ ] Write an integration test (API level):
  - Create two themes
  - Activate theme A -> verify public/active-theme returns A
  - Activate theme B -> verify public/active-theme returns B
  - Content pages still render correctly
- [ ] Verify: `cd frontend && pnpm type-check`

### Task 27: Implement theme preview (3.4.3)

**Files:**
- Create: `frontend/src/pages/admin/themes/ThemePreview.tsx`
- Create: `frontend/src/pages/admin/themes/ThemePreview.test.tsx`

**Steps:**
- [ ] Implement a preview component that temporarily applies a theme's tokens without activating it:

```tsx
interface ThemePreviewProps {
  themeId: string;
  onClose: () => void;
}

export function ThemePreview({ themeId, onClose }: ThemePreviewProps) {
  // 1. Load the theme package by ID
  // 2. Apply its tokens to a scoped container (using CSS variables in a wrapper div)
  // 3. Render a preview of the homepage inside an iframe or scoped div
  // 4. Show "Activate" / "Cancel" buttons
}
```

- [ ] The preview renders in an iframe pointing to `/?theme_preview={themeId}` — the ThemeProvider checks for this query param and temporarily overrides the active theme
- [ ] Write tests:
  - Preview component renders with correct theme tokens
  - Cancel button restores original theme
  - Activate button calls the activation API
- [ ] Verify: `cd frontend && pnpm test -- src/pages/admin/themes/ && pnpm type-check`

---

## Chunk 10: Plugin & Theme Marketplace (Tasks 3.4.4 — 3.4.7)

### Task 28: Implement marketplace backend API (3.4.4)

**Files:**
- Create: `backend/internal/handler/marketplace/handler.go`
- Create: `backend/internal/handler/marketplace/handler_test.go`
- Modify: `backend/cmd/server/main.go`

**Steps:**
- [ ] The marketplace works in two modes:
  1. **Embedded registry**: plugins/themes shipped in the `plugins/` directory
  2. **Remote registry**: fetches catalog from a configurable registry URL

- [ ] Implement handler:

```go
package marketplace

import "github.com/gin-gonic/gin"

type Handler struct {
    registryURL string // remote registry URL (empty = embedded only)
    pluginDir   string // local plugin directory
}

func NewHandler(registryURL, pluginDir string) *Handler

func (h *Handler) RegisterRoutes(admin *gin.RouterGroup) {
    market := admin.Group("/marketplace")
    market.GET("/plugins", h.ListPlugins)       // browse available plugins
    market.GET("/plugins/:id", h.GetPlugin)     // plugin detail + versions
    market.GET("/themes", h.ListThemes)         // browse available themes
    market.GET("/themes/:id", h.GetTheme)       // theme detail
    market.POST("/install", h.Install)          // install from marketplace
    market.POST("/check-updates", h.CheckUpdates) // check for available updates
}

// @Summary List available plugins in the marketplace
// @Tags marketplace
// @Security BearerAuth
// @Param category query string false "Filter by category"
// @Param search query string false "Search query"
// @Success 200 {array} MarketplaceItem
// @Router /admin/marketplace/plugins [get]
func (h *Handler) ListPlugins(c *gin.Context)

type MarketplaceItem struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    NameZh      string   `json:"nameZh"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Version     string   `json:"version"`
    Downloads   int64    `json:"downloads"`
    Category    string   `json:"category"`
    Tags        []string `json:"tags"`
    Preview     string   `json:"preview"` // screenshot URL
    Installed   bool     `json:"installed"`
    UpdateAvail bool     `json:"updateAvailable"`
}
```

- [ ] For the initial implementation, the embedded registry reads `plugin.yaml` files from the local `plugins/` directory. Remote registry support is stubbed but not implemented (returns empty results).
- [ ] Wire into main.go admin routes
- [ ] Write handler tests
- [ ] Verify: `cd backend && go build ./cmd/server/ && go test -v -race ./internal/handler/marketplace/...`

### Task 29: Implement marketplace frontend page (3.4.5)

**Files:**
- Create: `frontend/src/pages/admin/plugins/page.tsx`
- Create: `frontend/src/pages/admin/plugins/PluginCard.tsx`
- Create: `frontend/src/pages/admin/plugins/PluginDetailModal.tsx`
- Create: `frontend/src/pages/admin/plugins/page.test.tsx`
- Modify: `frontend/src/router/config.tsx` — add plugin management route

**Steps:**
- [ ] Implement the plugin management page with two tabs: "Installed" and "Marketplace"
- [ ] "Installed" tab:
  - List all installed plugins with status (enabled/disabled/failed)
  - Enable/Disable toggle for each plugin
  - Settings button (opens PluginSettingsForm)
  - Uninstall button with confirmation
- [ ] "Marketplace" tab:
  - Grid of available plugins (PluginCard component)
  - Search and category filter
  - Click to open PluginDetailModal with description, version, permissions, install button
- [ ] PluginCard component:
  - Plugin name, author, version, short description
  - "Install" / "Installed" / "Update" button
  - Download count badge
- [ ] PluginDetailModal:
  - Full description, screenshots, permissions list
  - Version history
  - "Install" button that shows permission confirmation dialog
- [ ] Add route to `config.tsx`:
```typescript
{
  path: "plugins",
  lazy: () => import("@/pages/admin/plugins/page"),
}
```
- [ ] Write tests:
  - Plugin list renders installed plugins
  - Enable/disable toggle calls correct API
  - Marketplace tab renders available plugins
  - Install button shows permission confirmation
- [ ] Verify: `cd frontend && pnpm test -- src/pages/admin/plugins/ && pnpm type-check`

### Task 30: Implement one-click install/update (3.4.6)

**Files:**
- Modify: `backend/internal/plugin/manager.go` — add InstallFromMarketplace
- Modify: `backend/internal/handler/marketplace/handler.go` — add Install endpoint logic

**Steps:**
- [ ] Implement `Manager.InstallFromMarketplace(ctx, marketplaceID)`:
  1. Fetch plugin archive from registry (for now, copy from local `plugins/` dir)
  2. Extract to `{PluginDir}/{pluginID}/`
  3. Validate plugin.yaml
  4. Build binary (`go build`)
  5. Create DB record
  6. Return plugin metadata
- [ ] Implement `Manager.UpdatePlugin(ctx, pluginID)`:
  1. Disable plugin if enabled
  2. Download new version
  3. Replace files
  4. Rebuild binary
  5. Re-enable if was previously enabled
- [ ] Write tests:
  - Install from local directory creates correct file structure
  - Update preserves settings
  - Update re-enables previously enabled plugin
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

### Task 31: Stub marketplace registry service (3.4.7)

**Files:**
- Create: `backend/internal/plugin/registry_client.go`

**Steps:**
- [ ] Implement a registry client that can fetch plugin catalogs from a remote URL:

```go
package plugin

import (
    "context"
    "encoding/json"
    "net/http"
)

// RegistryClient fetches plugin/theme catalogs from a remote registry.
type RegistryClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewRegistryClient(baseURL string) *RegistryClient

// ListPlugins fetches available plugins from the registry.
func (c *RegistryClient) ListPlugins(ctx context.Context) ([]PluginMeta, error)

// ListThemes fetches available themes from the registry.
func (c *RegistryClient) ListThemes(ctx context.Context) ([]ThemeMeta, error)

// DownloadPlugin downloads a plugin archive.
func (c *RegistryClient) DownloadPlugin(ctx context.Context, pluginID, version string) ([]byte, error)
```

- [ ] For now this is a stub — the `baseURL` defaults to empty, and all methods return `nil, nil` when no registry is configured. A real registry service would be a separate deployment (out of scope for Phase 3).
- [ ] Write tests verifying graceful behavior when no registry is configured
- [ ] Verify: `cd backend && go test -v -race ./internal/plugin/...`

---

## Chunk 11: Integration Wiring & Final Verification

### Task 32: Wire everything into main.go

**Files:**
- Modify: `backend/cmd/server/main.go`

**Steps:**
- [ ] Add Plugin and PluginSetting to the AutoMigrate call:

```go
migrator.AutoMigrate(
    // ... existing models ...
    &model.Plugin{},
    &model.PluginSetting{},
)
```

- [ ] Initialize PluginManager after provider registry:

```go
// Initialize plugin system
pluginStore := plugin.NewStore(database.DB)
pluginManager := plugin.NewManager(plugin.ManagerConfig{
    PluginDir: filepath.Join(cfg.UploadDir, "..", "plugins"),
    DataDir:   filepath.Join(cfg.UploadDir, "..", "data", "plugins"),
}, pluginStore, registry, bus)

// Start previously enabled plugins
if err := pluginManager.StartEnabledPlugins(context.Background()); err != nil {
    log.Warn("Some plugins failed to start", "error", err)
    // Non-fatal: continue server startup
}
```

- [ ] Add plugin routes to admin group:

```go
pluginHandlerInst := pluginHandler.NewHandler(pluginManager)
pluginHandlerInst.RegisterRoutes(adminGroup, router.Group("/api/v1/plugins"))

marketplaceHandlerInst := marketplace.NewHandler("", filepath.Join(cfg.UploadDir, "..", "plugins"))
marketplaceHandlerInst.RegisterRoutes(adminGroup)
```

- [ ] Add graceful shutdown:

```go
// In shutdown section:
pluginManager.StopAll()
```

- [ ] Update NoRoute handler to exclude plugin API paths:

```go
!strings.HasPrefix(path, "/api/v1/plugins/") &&
```

- [ ] Verify: `cd backend && go build ./cmd/server/`

### Task 33: End-to-end integration test

**Steps:**
- [ ] Create `backend/internal/plugin/integration_test.go`:
  1. Start an in-memory SQLite DB
  2. Create a PluginManager
  3. Create a minimal test plugin binary (compile a Go plugin that implements NotifierProvider)
  4. Install the test plugin
  5. Enable it — verify provider is registered in the Registry
  6. Call the provider through the Registry — verify it responds
  7. Disable it — verify provider is unregistered
  8. Uninstall — verify cleanup
- [ ] Verify: `cd backend && go test -v -race -timeout 60s ./internal/plugin/...`

### Task 34: Run full verification suite

**Steps:**
- [ ] Run backend compilation and tests:
```bash
cd backend && go build ./... && go test -v -race ./...
```
- [ ] Run backend static analysis:
```bash
cd backend && go vet ./...
```
- [ ] Run frontend linting and type checking:
```bash
pnpm lint && pnpm type-check
```
- [ ] Run frontend tests:
```bash
pnpm test
```
- [ ] Verify Swagger regeneration (if swag is installed):
```bash
cd backend && swag init -g cmd/server/main.go -o docs/swagger
```

---

## Dependency Graph

```
Chunk 1 (types, manifest)
    ↓
Chunk 2 (lifecycle, store)
    ↓
Chunk 3 (go-plugin, gRPC, GRPCHost, Manager)  ←── Chunk 4 (sandbox)
    ↓
Chunk 5 (Go SDK)
    ↓                           ↓
Chunk 7 (handler, CLI)    Chunk 6 (frontend extension system)
    ↓                           ↓
Chunk 8 (demo plugins)    Chunk 9 (theme marketplace)
    ↓                           ↓
         Chunk 10 (marketplace UI + API)
                    ↓
         Chunk 11 (integration wiring)
```

Chunks 4, 6, 9 can be worked in parallel once their dependencies are met.
Chunk 8 (demo plugins) can start as soon as Chunk 5 (SDK) is ready.

---

## Estimated Timeline

| Chunk | Description | Tasks | Est. Days |
|-------|-------------|-------|-----------|
| 1 | Plugin types & manifest | 1-3 | 1.5 |
| 2 | Lifecycle state machine | 4 | 1 |
| 3 | go-plugin runtime | 5-8 | 3 |
| 4 | Permission sandbox | 9 | 0.5 |
| 5 | Go SDK | 10-12 | 2 |
| 6 | Frontend extensions | 13-16 | 2 |
| 7 | Admin handler & CLI | 17-18 | 1.5 |
| 8 | Demo plugins (6x) | 19-24 | 3 |
| 9 | Theme marketplace | 25-27 | 1.5 |
| 10 | Marketplace API + UI | 28-31 | 2 |
| 11 | Integration wiring | 32-34 | 1 |
| **Total** | | **34 tasks** | **~19 days** |

---

## Key Design Decisions

1. **HashiCorp go-plugin over WASM/raw gRPC**: Battle-tested, handles process lifecycle automatically, supports Go natively and other languages via gRPC. WASM is too restrictive for CMS plugins that need filesystem/network/DB access.

2. **Plugin-per-process over shared-library**: Each plugin runs in its own process for isolation. Crash in a plugin does not crash the host. Security boundary is real (process-level). go-plugin handles all the complexity.

3. **Provider interface as plugin contract**: Plugins implement the same Provider interfaces already defined in Phase 2. No new abstraction layer — plugins just provide remote implementations of StorageProvider, SearchProvider, etc. via gRPC proxying.

4. **Frontend plugins via dynamic import**: Plugin JS bundles are loaded at runtime via `import()`. They register React components into named extension slots. No build-time coupling between host and plugins.

5. **Embedded marketplace first**: The initial marketplace reads from the local `plugins/` directory. A remote registry is stubbed but not implemented. This avoids building a separate registry service in Phase 3.

6. **Plugin data isolation**: Plugins that need persistence use their own embedded SQLite database (webhook plugin) or connect to external services (Meilisearch, S3). Host-mediated DB access is a future enhancement.

7. **Settings via JSON Schema**: Plugin settings are defined as JSON Schema in plugin.yaml. The frontend auto-generates a settings form from this schema. No custom UI needed for basic configuration.
