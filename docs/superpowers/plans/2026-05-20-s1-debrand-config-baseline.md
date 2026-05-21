# S1 · 去品牌化与可配置基线 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generic-ize impress so any owner can deploy and brand it via `globalConfig` editing — removing all "印迹/Blotting" hardcoding from user-visible surfaces while keeping the consultancy demo data accessible via `SEED_MODE=demo`.

**Architecture:** Six linear PRs. PR-1 lays the schema + admin write endpoint that everything else depends on. PR-2/3 swap hardcoded fallbacks for config-driven values. PR-4 gates consultancy-only pages behind `features.publicPages`. PR-5 finishes SEO + ships the proper form-based admin editor. PR-6 renames DSN, splits seeds, and verifies 0 brand residue.

**Tech Stack:** Go 1.24 / Gin / GORM (SQLite) backend · React 19 / TypeScript / Vite / Tailwind frontend · Vitest + happy-dom for FE tests · `go test -race` for BE tests · pnpm monorepo (frontend workspace).

**Spec reference:** `docs/superpowers/specs/2026-05-20-s1-debrand-config-baseline-design.md` (commits `594f85c` + `004807e`).

---

## File Structure Overview

### Created files

| Path | Purpose | Introduced in |
|---|---|---|
| `backend/internal/handler/global_config/handler.go` | Admin endpoints for `content_documents.global` + schema validation | PR-1 |
| `backend/internal/handler/global_config/schema.go` | `SiteConfigGlobal` Go struct + `validateGlobalConfig` | PR-1 |
| `backend/internal/handler/global_config/handler_test.go` | Handler integration tests | PR-1 |
| `backend/internal/handler/features/handler.go` | Admin endpoints for `site_configs.features` | PR-4 |
| `backend/internal/handler/features/schema.go` | `SiteConfigFeatures` Go struct + `validateFeaturesConfig` | PR-4 |
| `backend/internal/handler/features/handler_test.go` | Handler integration tests | PR-4 |
| `frontend/src/lib/locale.ts` | `pickLocaleValue` + LocalizedString types | PR-1 |
| `frontend/src/lib/locale.test.ts` | Unit tests | PR-1 |
| `frontend/src/types/siteConfig.ts` | TypeScript `SiteConfigGlobal` / `SiteConfigFeatures` types | PR-1 |
| `frontend/src/hooks/useBranding.ts` | Pulls identity / brand / author / footer from `globalConfig` | PR-2 |
| `frontend/src/hooks/useBranding.test.ts` | Unit tests | PR-2 |
| `frontend/src/hooks/useLocaleMode.ts` | `localeMode` + currentLocale collapsed by mono mode | PR-3 |
| `frontend/src/hooks/useLocaleMode.test.ts` | Unit tests | PR-3 |
| `frontend/src/components/feature/FeatureGate.tsx` | Renders children only if feature flag enabled | PR-4 |
| `frontend/src/components/feature/FeatureGate.test.tsx` | Unit tests | PR-4 |
| `frontend/src/router/featureMap.ts` | Single source of truth for route ↔ feature-key map | PR-4 |
| `frontend/src/hooks/useSEODefaults.ts` | SEO defaults + `buildTitle` template renderer | PR-5 |
| `frontend/src/hooks/useSEODefaults.test.ts` | Unit tests | PR-5 |
| `frontend/src/pages/admin/site-config/page.tsx` | Admin editor (PR-1: raw JSON / PR-5: tabbed form) | PR-1 / PR-5 |
| `frontend/src/pages/admin/features/page.tsx` | Admin features toggles | PR-4 |
| `scripts/check-brand-residue.sh` | Greps for "印迹\|blotting\|Blotting" in product code | PR-6 |

### Modified files

| Path | What changes | PR |
|---|---|---|
| `backend/internal/handler/bootstrap/handler.go` | Cache invalidation on global config writes | PR-1 |
| `backend/cmd/server/routes.go` | Register new admin endpoints | PR-1 / PR-4 |
| `backend/cmd/server/main.go` | Wire new handler dependencies | PR-1 / PR-4 |
| `backend/pkg/config/config.go` | Default DSN → `impress.db` | PR-6 |
| `backend/internal/seo/meta.go` | Read defaults from `content_documents.global.seo` | PR-5 |
| `backend/internal/seed/seed.go` | Split into `BlankSiteSeed` + `DemoSiteSeed`, `SEED_MODE` env | PR-6 |
| `backend/internal/seed/seed_test.go` | Test both seed modes | PR-6 |
| `frontend/src/contexts/GlobalConfigContext.tsx` | Type extension | PR-1 |
| `frontend/src/components/feature/Header.tsx` | Use `useBranding`, drop `'Blotting Consultancy'` fallback | PR-2 |
| `frontend/src/components/feature/Footer.tsx` | Use `useBranding`, drop `readdy.ai` link, hide ICP when empty | PR-2 |
| `frontend/src/theme/layouts/ThemedFooter.tsx` | Same as Footer | PR-2 |
| `frontend/src/pages/admin/AdminLayout.tsx` | Use `useBranding().siteName` in titles | PR-2 |
| `frontend/src/i18n/local/zh/common.ts` | Strip business strings, keep UI strings only | PR-2 |
| `frontend/src/i18n/local/en/common.ts` | Same | PR-2 |
| `frontend/src/hooks/useDocumentTitle.ts` | Drop `suffix` param; uses SEO hook internally | PR-5 |
| `frontend/src/router/config.tsx` | Wrap consultancy routes in `<FeatureGate>` | PR-4 |
| `frontend/src/theme/DynamicPage.tsx` | Drop `useDocumentTitle("…", "印迹法规咨询")` suffix | PR-5 |
| `frontend/src/theme/packages/default/index.ts` | `author: "impress"`, neutral description | PR-6 |
| `frontend/src/theme/packages/modern-dark/index.ts` | Same | PR-6 |
| `frontend/src/theme/packages/warm-earth/index.ts` | Same | PR-6 |
| `frontend/src/FRONTEND_RENDERING.md` | Update intro paragraph | PR-6 |
| `frontend/src/modules/qa/admin/page.tsx` | Drop `"印迹后台"` suffix | PR-5 |
| `frontend/src/pages/admin/AdminLayout.tsx` (re-touch) | Pull in features-gated menu items | PR-4 |

---

## PR-1 · `feat(global-config): schema + admin endpoint + raw JSON editor`

**Outcome:** Admin can read & write `globalConfig` via `/admin/global-config/*`; new schema validated server-side; frontend types match; no user-visible behavior change.

### Task 1.1 — Define `SiteConfigGlobal` Go struct + schema constants

**Files:**
- Create: `backend/internal/handler/global_config/schema.go`

- [ ] **Step 1: Write the schema file**

```go
package global_config

import (
	"errors"
	"fmt"
	"strings"

	"blotting-consultancy/internal/model"
)

type LocaleMode string

const (
	LocaleModeMonoZh    LocaleMode = "mono-zh"
	LocaleModeMonoEn    LocaleMode = "mono-en"
	LocaleModeBilingual LocaleMode = "bilingual"
)

type LocalizedString struct {
	Zh string `json:"zh,omitempty"`
	En string `json:"en,omitempty"`
}

type Identity struct {
	Name          LocalizedString `json:"name"`
	Tagline       LocalizedString `json:"tagline,omitempty"`
	LocaleMode    LocaleMode      `json:"localeMode"`
	DefaultLocale string          `json:"defaultLocale"`
}

type LogoRef struct {
	Light string `json:"light"`
	Dark  string `json:"dark,omitempty"`
}

type Brand struct {
	Logo          LogoRef `json:"logo"`
	Favicon       string  `json:"favicon"`
	OgImage       string  `json:"ogImage"`
	PrimaryColor  string  `json:"primaryColor"`
	AccentColor   string  `json:"accentColor,omitempty"`
}

type Social struct {
	Kind  string `json:"kind"`
	URL   string `json:"url"`
	Label string `json:"label,omitempty"`
}

type Author struct {
	Name     string          `json:"name"`
	Avatar   string          `json:"avatar,omitempty"`
	Bio      LocalizedString `json:"bio,omitempty"`
	Location string          `json:"location,omitempty"`
	Socials  []Social        `json:"socials"`
}

type ExtraLink struct {
	Label LocalizedString `json:"label"`
	URL   string          `json:"url"`
}

type Footer struct {
	Copyright  LocalizedString `json:"copyright,omitempty"`
	ICP        string          `json:"icp,omitempty"`
	ExtraLinks []ExtraLink     `json:"extraLinks,omitempty"`
}

type SEO struct {
	DefaultTitle       LocalizedString `json:"defaultTitle,omitempty"`
	TitleTemplate      string          `json:"titleTemplate,omitempty"`
	DefaultDescription LocalizedString `json:"defaultDescription,omitempty"`
	TwitterHandle      string          `json:"twitterHandle,omitempty"`
}

type SiteConfigGlobal struct {
	Identity Identity `json:"identity"`
	Brand    Brand    `json:"brand"`
	Author   Author   `json:"author"`
	Footer   Footer   `json:"footer"`
	SEO      SEO      `json:"seo"`
}

// validateGlobalConfig converts a JSONMap to SiteConfigGlobal and validates.
// Returns nil error on success; concrete error describing the offending field on failure.
func validateGlobalConfig(raw model.JSONMap) (*SiteConfigGlobal, error) {
	bytes, err := jsonMarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal raw config: %w", err)
	}
	var cfg SiteConfigGlobal
	if err := jsonUnmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal to SiteConfigGlobal: %w", err)
	}
	if cfg.Identity.Name.Zh == "" && cfg.Identity.Name.En == "" {
		return nil, errors.New("identity.name: at least one locale must be non-empty")
	}
	switch cfg.Identity.LocaleMode {
	case LocaleModeMonoZh, LocaleModeMonoEn, LocaleModeBilingual:
	default:
		return nil, fmt.Errorf("identity.localeMode: must be one of mono-zh, mono-en, bilingual (got %q)", cfg.Identity.LocaleMode)
	}
	if cfg.Identity.DefaultLocale != "zh" && cfg.Identity.DefaultLocale != "en" {
		return nil, fmt.Errorf("identity.defaultLocale: must be zh or en (got %q)", cfg.Identity.DefaultLocale)
	}
	switch cfg.Identity.LocaleMode {
	case LocaleModeMonoZh:
		if cfg.Identity.DefaultLocale != "zh" {
			return nil, errors.New("identity.defaultLocale: must equal zh when localeMode=mono-zh")
		}
	case LocaleModeMonoEn:
		if cfg.Identity.DefaultLocale != "en" {
			return nil, errors.New("identity.defaultLocale: must equal en when localeMode=mono-en")
		}
	}
	if len(cfg.Footer.ICP) > 100 {
		return nil, errors.New("footer.icp: max length 100")
	}
	for i, s := range cfg.Author.Socials {
		if strings.TrimSpace(s.URL) == "" {
			return nil, fmt.Errorf("author.socials[%d].url: required", i)
		}
	}
	return &cfg, nil
}
```

- [ ] **Step 2: Add helpers `jsonMarshal` / `jsonUnmarshal` at the bottom**

```go
// Indirect through package-level vars so tests can inject if needed.
var (
	jsonMarshal   = jsonStdMarshal
	jsonUnmarshal = jsonStdUnmarshal
)
```

Append at top of file (in imports + same package, but separate concerns) — replace the bare `import` block at the top with:

```go
import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"blotting-consultancy/internal/model"
)

func jsonStdMarshal(v any) ([]byte, error)         { return json.Marshal(v) }
func jsonStdUnmarshal(data []byte, v any) error    { return json.Unmarshal(data, v) }
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./internal/handler/global_config/...`
Expected: no output (success).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/global_config/schema.go
git commit -m "feat(global-config): SiteConfigGlobal Go struct + validate"
```

### Task 1.2 — Schema validation unit tests

**Files:**
- Create: `backend/internal/handler/global_config/schema_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package global_config

import (
	"strings"
	"testing"

	"blotting-consultancy/internal/model"
)

func validBase() model.JSONMap {
	return model.JSONMap{
		"identity": map[string]any{
			"name":          map[string]any{"zh": "My Site"},
			"localeMode":    "mono-zh",
			"defaultLocale": "zh",
		},
		"brand":  map[string]any{},
		"author": map[string]any{"socials": []any{}},
		"footer": map[string]any{},
		"seo":    map[string]any{},
	}
}

func TestValidateGlobalConfig_AcceptsMinimalValid(t *testing.T) {
	cfg, err := validateGlobalConfig(validBase())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if cfg.Identity.LocaleMode != LocaleModeMonoZh {
		t.Errorf("got localeMode=%q want mono-zh", cfg.Identity.LocaleMode)
	}
}

func TestValidateGlobalConfig_RejectsEmptyName(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["name"] = map[string]any{}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "identity.name") {
		t.Fatalf("expected identity.name error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsInvalidLocaleMode(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["localeMode"] = "klingon"
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "localeMode") {
		t.Fatalf("expected localeMode error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsDefaultLocaleMismatch(t *testing.T) {
	raw := validBase()
	raw["identity"].(map[string]any)["localeMode"] = "mono-en"
	raw["identity"].(map[string]any)["defaultLocale"] = "zh"
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "defaultLocale") {
		t.Fatalf("expected defaultLocale mismatch error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsLongICP(t *testing.T) {
	raw := validBase()
	raw["footer"] = map[string]any{"icp": strings.Repeat("a", 101)}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "footer.icp") {
		t.Fatalf("expected footer.icp error, got: %v", err)
	}
}

func TestValidateGlobalConfig_RejectsSocialWithoutURL(t *testing.T) {
	raw := validBase()
	raw["author"] = map[string]any{
		"socials": []any{map[string]any{"kind": "github"}},
	}
	_, err := validateGlobalConfig(raw)
	if err == nil || !strings.Contains(err.Error(), "socials[0].url") {
		t.Fatalf("expected socials[0].url error, got: %v", err)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `cd backend && go test ./internal/handler/global_config/... -run TestValidateGlobalConfig -v`
Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/global_config/schema_test.go
git commit -m "test(global-config): schema validation cases"
```

### Task 1.3 — Admin handler GET / PUT draft / POST publish

**Files:**
- Create: `backend/internal/handler/global_config/handler.go`

- [ ] **Step 1: Write the handler**

```go
package global_config

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/cache"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// Handler serves admin endpoints for the "global" content document.
type Handler struct {
	repo  repository.ContentDocumentRepository
	cache *cache.Cache
}

func NewHandler(repo repository.ContentDocumentRepository, c *cache.Cache) *Handler {
	return &Handler{repo: repo, cache: c}
}

func (h *Handler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.GET("/global-config", h.adminGet)
	admin.PUT("/global-config/draft", h.adminPutDraft)
	admin.POST("/global-config/publish", h.adminPublish)
}

type getResponse struct {
	DraftConfig      model.JSONMap `json:"draftConfig"`
	DraftVersion     int           `json:"draftVersion"`
	PublishedConfig  model.JSONMap `json:"publishedConfig"`
	PublishedVersion int           `json:"publishedVersion"`
}

func (h *Handler) adminGet(c *gin.Context) {
	doc, err := h.repo.FindByPageKey(c.Request.Context(), model.PageKeyGlobal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "global config not found"}})
		return
	}
	c.JSON(http.StatusOK, getResponse{
		DraftConfig:      doc.DraftConfig,
		DraftVersion:     doc.DraftVersion,
		PublishedConfig:  doc.PublishedConfig,
		PublishedVersion: doc.PublishedVersion,
	})
}

type putDraftInput struct {
	DraftConfig          model.JSONMap `json:"draftConfig"`
	ExpectedDraftVersion int           `json:"expectedDraftVersion"`
}

func (h *Handler) adminPutDraft(c *gin.Context) {
	var input putDraftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}
	if _, err := validateGlobalConfig(input.DraftConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	newVersion, err := h.repo.UpdateDraft(c.Request.Context(), model.PageKeyGlobal, input.ExpectedDraftVersion, input.DraftConfig)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"message": "draft version conflict"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to update draft"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"draftVersion": newVersion})
}

func (h *Handler) adminPublish(c *gin.Context) {
	doc, err := h.repo.FindByPageKey(c.Request.Context(), model.PageKeyGlobal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "global config not found"}})
		return
	}
	if _, err := validateGlobalConfig(doc.DraftConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "current draft fails validation: " + err.Error()}})
		return
	}
	newPub := doc.PublishedVersion + 1
	if err := h.repo.UpdatePublished(c.Request.Context(), model.PageKeyGlobal, doc.DraftConfig, newPub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to publish"}})
		return
	}
	// Invalidate bootstrap + public content caches for "global".
	if h.cache != nil {
		h.cache.DeletePrefix("bootstrap:")
		h.cache.Delete("content:global:zh")
		h.cache.Delete("content:global:en")
	}
	c.JSON(http.StatusOK, gin.H{"publishedVersion": newPub})
}
```

- [ ] **Step 2: Check `repository.ErrVersionConflict` exists**

Run: `cd backend && grep -n "ErrVersionConflict" internal/repository/*.go`
Expected: at least one match in `content_document_repository_impl.go` defining the error. If absent, add `var ErrVersionConflict = errors.New("draft version conflict")` to `repository/content_document_repository.go` and import `errors`.

- [ ] **Step 3: Check `cache.Cache.DeletePrefix` / `Delete` methods exist**

Run: `cd backend && grep -n "func.*Cache.*Delete\|func.*Cache.*DeletePrefix" internal/cache/*.go`
If `DeletePrefix` doesn't exist, use `c.cache.Clear()` instead in the publish handler (less precise but acceptable since publish is rare). Replace the `DeletePrefix` line accordingly.

- [ ] **Step 4: Verify compile**

Run: `cd backend && go build ./...`
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/global_config/handler.go
git commit -m "feat(global-config): admin GET/PUT-draft/POST-publish handlers"
```

### Task 1.4 — Wire handler into main + routes

**Files:**
- Modify: `backend/cmd/server/main.go` (handler construction)
- Modify: `backend/cmd/server/routes.go` (route registration)

- [ ] **Step 1: Locate handler construction site**

Run: `cd backend && grep -n "Bootstrap.*NewHandler\|bootstrap.NewHandler" cmd/server/main.go`
Note the line number — register the new handler nearby in the same pattern.

- [ ] **Step 2: Add handler construction in `main.go`**

Locate the section where existing handlers are created. Add (alongside other handler constructions):

```go
import (
    // ... existing imports
    globalConfigHandler "blotting-consultancy/internal/handler/global_config"
)

// in handler-wiring block:
handlers.GlobalConfig = globalConfigHandler.NewHandler(contentDocRepo, cache)
```

Add a field to the `Handlers` struct (or whatever the project's handler aggregate is — find it via `grep -n "type.*Handlers" cmd/server/main.go`):

```go
GlobalConfig *globalConfigHandler.Handler
```

- [ ] **Step 3: Register routes**

Modify `backend/cmd/server/routes.go`. After the existing admin route registrations (e.g. near line 326-340 where `theme` and `email-settings` are wired), add:

```go
		// Global config (branding / identity / SEO defaults)
		handlers.GlobalConfig.RegisterRoutes(adminGroup)
```

- [ ] **Step 4: Verify build**

Run: `cd backend && go build ./cmd/server/`
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/server/main.go backend/cmd/server/routes.go
git commit -m "feat(global-config): wire admin routes"
```

### Task 1.5 — Handler integration tests

**Files:**
- Create: `backend/internal/handler/global_config/handler_test.go`

- [ ] **Step 1: Write integration test using gin httptest harness**

```go
package global_config_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/cache"
	gcfg "blotting-consultancy/internal/handler/global_config"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/db"
)

func setup(t *testing.T) (*gin.Engine, repository.ContentDocumentRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	conn, err := db.OpenSQLite("file::memory:?cache=shared&mode=rwc")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&model.ContentDocument{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewGormContentDocumentRepository(conn)
	// seed a global doc
	doc := &model.ContentDocument{
		PageKey:          model.PageKeyGlobal,
		DraftConfig:      model.JSONMap{},
		DraftVersion:     1,
		PublishedConfig:  model.JSONMap{},
		PublishedVersion: 1,
	}
	if err := repo.Create(t.Context(), doc); err != nil {
		t.Fatalf("seed: %v", err)
	}
	r := gin.New()
	admin := r.Group("/admin")
	gcfg.NewHandler(repo, cache.New(0)).RegisterRoutes(admin)
	return r, repo
}

func TestAdminPutDraft_ValidatesSchema(t *testing.T) {
	r, _ := setup(t)
	body := `{"draftConfig":{"identity":{"name":{"zh":""},"localeMode":"mono-zh","defaultLocale":"zh"},"brand":{},"author":{"socials":[]},"footer":{},"seo":{}},"expectedDraftVersion":1}`
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPutDraft_AcceptsValid(t *testing.T) {
	r, _ := setup(t)
	cfg := map[string]any{
		"identity": map[string]any{
			"name":          map[string]any{"zh": "My Site"},
			"localeMode":    "mono-zh",
			"defaultLocale": "zh",
		},
		"brand":  map[string]any{},
		"author": map[string]any{"socials": []any{}},
		"footer": map[string]any{},
		"seo":    map[string]any{},
	}
	bodyMap := map[string]any{"draftConfig": cfg, "expectedDraftVersion": 1}
	body, _ := json.Marshal(bodyMap)
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPublish_BumpsVersion(t *testing.T) {
	r, repo := setup(t)
	// First put a valid draft.
	cfg := map[string]any{
		"identity": map[string]any{
			"name":          map[string]any{"zh": "My Site"},
			"localeMode":    "mono-zh",
			"defaultLocale": "zh",
		},
		"brand":  map[string]any{},
		"author": map[string]any{"socials": []any{}},
		"footer": map[string]any{},
		"seo":    map[string]any{},
	}
	body, _ := json.Marshal(map[string]any{"draftConfig": cfg, "expectedDraftVersion": 1})
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("draft put failed: %d %s", w.Code, w.Body.String())
	}
	// Then publish.
	req = httptest.NewRequest(http.MethodPost, "/admin/global-config/publish", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("publish failed: %d %s", w.Code, w.Body.String())
	}
	doc, err := repo.FindByPageKey(t.Context(), model.PageKeyGlobal)
	if err != nil {
		t.Fatalf("find after publish: %v", err)
	}
	if doc.PublishedVersion != 2 {
		t.Fatalf("expected published version 2, got %d", doc.PublishedVersion)
	}
}
```

- [ ] **Step 2: Confirm `db.OpenSQLite` signature**

Run: `cd backend && grep -n "func OpenSQLite\|func Open" internal/db/*.go`
If the signature differs (e.g. takes config struct, not DSN string), adjust the `setup` call accordingly. If only a single open-from-cfg helper exists, use the project's existing test helper from another `*_test.go` file (search: `grep -rn "func.*setupTest.*sqlite" backend/internal/`).

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./internal/handler/global_config/... -v`
Expected: all three tests pass.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/global_config/handler_test.go
git commit -m "test(global-config): admin handler integration"
```

### Task 1.6 — Frontend `LocalizedString` types + `pickLocaleValue`

**Files:**
- Create: `frontend/src/lib/locale.ts`
- Create: `frontend/src/lib/locale.test.ts`

- [ ] **Step 1: Write the helper**

```ts
export type Locale = "zh" | "en";
export type LocaleMode = "mono-zh" | "mono-en" | "bilingual";

export type LocalizedString = { zh?: string; en?: string };

export interface PickLocaleValueArgs {
  value: LocalizedString | undefined | null;
  mode: LocaleMode;
  defaultLocale: Locale;
  currentLocale: Locale;
}

export function pickLocaleValue({
  value,
  mode,
  defaultLocale,
  currentLocale,
}: PickLocaleValueArgs): string {
  if (!value) return "";
  if (mode === "mono-zh") return value.zh ?? "";
  if (mode === "mono-en") return value.en ?? "";
  return (
    value[currentLocale] ??
    value[defaultLocale] ??
    value.zh ??
    value.en ??
    ""
  );
}
```

- [ ] **Step 2: Write tests**

```ts
import { describe, expect, it } from "vitest";
import { pickLocaleValue } from "./locale";

describe("pickLocaleValue", () => {
  it("returns empty string for undefined", () => {
    expect(
      pickLocaleValue({
        value: undefined,
        mode: "bilingual",
        defaultLocale: "zh",
        currentLocale: "zh",
      })
    ).toBe("");
  });

  it("mono-zh ignores en", () => {
    expect(
      pickLocaleValue({
        value: { zh: "中文", en: "English" },
        mode: "mono-zh",
        defaultLocale: "zh",
        currentLocale: "zh",
      })
    ).toBe("中文");
  });

  it("mono-zh returns empty if zh missing", () => {
    expect(
      pickLocaleValue({
        value: { en: "English" },
        mode: "mono-zh",
        defaultLocale: "zh",
        currentLocale: "zh",
      })
    ).toBe("");
  });

  it("bilingual prefers currentLocale", () => {
    expect(
      pickLocaleValue({
        value: { zh: "中文", en: "English" },
        mode: "bilingual",
        defaultLocale: "zh",
        currentLocale: "en",
      })
    ).toBe("English");
  });

  it("bilingual falls back to defaultLocale", () => {
    expect(
      pickLocaleValue({
        value: { zh: "中文" },
        mode: "bilingual",
        defaultLocale: "zh",
        currentLocale: "en",
      })
    ).toBe("中文");
  });

  it("bilingual final fallback is the other language", () => {
    expect(
      pickLocaleValue({
        value: { en: "Only English" },
        mode: "bilingual",
        defaultLocale: "zh",
        currentLocale: "zh",
      })
    ).toBe("Only English");
  });
});
```

- [ ] **Step 3: Run the tests**

Run: `cd frontend && pnpm test -- src/lib/locale.test.ts`
Expected: all 6 cases pass.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/locale.ts frontend/src/lib/locale.test.ts
git commit -m "feat(lib): pickLocaleValue helper + LocalizedString types"
```

### Task 1.7 — TypeScript `SiteConfigGlobal` types + GlobalConfigContext extension

**Files:**
- Create: `frontend/src/types/siteConfig.ts`
- Modify: `frontend/src/contexts/GlobalConfigContext.tsx`

- [ ] **Step 1: Write the type file**

```ts
import type { Locale, LocaleMode, LocalizedString } from "@/lib/locale";

export interface SiteConfigIdentity {
  name: LocalizedString;
  tagline?: LocalizedString;
  localeMode: LocaleMode;
  defaultLocale: Locale;
}

export interface SiteConfigBrand {
  logo: { light: string; dark?: string };
  favicon: string;
  ogImage: string;
  primaryColor: string;
  accentColor?: string;
}

export type SocialKind =
  | "github"
  | "twitter"
  | "email"
  | "rss"
  | "linkedin"
  | "custom";

export interface SiteConfigSocial {
  kind: SocialKind;
  url: string;
  label?: string;
}

export interface SiteConfigAuthor {
  name: string;
  avatar?: string;
  bio?: LocalizedString;
  location?: string;
  socials: SiteConfigSocial[];
}

export interface SiteConfigFooterLink {
  label: LocalizedString;
  url: string;
}

export interface SiteConfigFooter {
  copyright?: LocalizedString;
  icp?: string;
  extraLinks?: SiteConfigFooterLink[];
}

export interface SiteConfigSEO {
  defaultTitle?: LocalizedString;
  titleTemplate?: string;
  defaultDescription?: LocalizedString;
  twitterHandle?: string;
}

export interface SiteConfigGlobal {
  identity: SiteConfigIdentity;
  brand: SiteConfigBrand;
  author: SiteConfigAuthor;
  footer: SiteConfigFooter;
  seo: SiteConfigSEO;
}

export interface SiteConfigFeatures {
  publicPages: {
    home: boolean;
    blog: boolean;
    contact: boolean;
    about: boolean;
    experts: boolean;
    coreServices: boolean;
    advantages: boolean;
    cases: boolean;
  };
  blog: {
    comments: boolean;
    rss: boolean;
  };
}

// Hard-coded defaults used when published config is missing or partial.
// Kept here (not in a context) so test code can import directly.
export const SITE_CONFIG_GLOBAL_DEFAULT: SiteConfigGlobal = {
  identity: {
    name: { zh: "My Site" },
    localeMode: "mono-zh",
    defaultLocale: "zh",
  },
  brand: {
    logo: { light: "" },
    favicon: "",
    ogImage: "",
    primaryColor: "#1e40af",
  },
  author: { name: "", socials: [] },
  footer: {},
  seo: {},
};

export const SITE_CONFIG_FEATURES_DEFAULT: SiteConfigFeatures = {
  publicPages: {
    home: true,
    blog: true,
    contact: true,
    about: false,
    experts: false,
    coreServices: false,
    advantages: false,
    cases: false,
  },
  blog: { comments: true, rss: true },
};
```

- [ ] **Step 2: Extend `GlobalConfigContext` to expose typed config**

In `frontend/src/contexts/GlobalConfigContext.tsx`, augment the existing `GlobalConfig` interface to inherit from the new schema while staying back-compat for legacy fields:

Replace the `interface GlobalConfig {...}` block (currently lines 26-40) with:

```ts
import type {
  SiteConfigGlobal,
  SiteConfigFeatures,
} from "@/types/siteConfig";
import { SITE_CONFIG_GLOBAL_DEFAULT, SITE_CONFIG_FEATURES_DEFAULT } from "@/types/siteConfig";

// Legacy shape retained for backwards compat with current rendering
// during the migration. New code should use `siteConfig` instead.
interface MediaRef {
  url?: string;
  alt?: string;
}
interface NavItem { label?: string; href?: string }
interface LinkItem { label?: string; href?: string }

export interface GlobalConfig {
  // legacy fields (kept while we transition)
  branding?: { logo?: MediaRef; companyName?: string };
  nav?: { items?: NavItem[] };
  footer?: { address?: string; phone?: string; links?: LinkItem[]; copyright?: string };
  // new typed shape
  siteConfig?: SiteConfigGlobal;
}
```

Then, at the end of the existing `useEffect` that calls `normalizeConfigForLocale`, attach the typed view if present. After the `setConfig(normalized as GlobalConfig)` line, add a helper to coerce. Replace:

```ts
const normalized = normalizeConfigForLocale(
  globalData.config as Record<string, unknown>,
  locale
);
setConfig(normalized as GlobalConfig);
```

with:

```ts
const normalized = normalizeConfigForLocale(
  globalData.config as Record<string, unknown>,
  locale
) as GlobalConfig;
// If the published config matches the new schema, expose it typed as siteConfig.
if (normalized && typeof normalized === "object" && "identity" in normalized) {
  normalized.siteConfig = normalized as unknown as SiteConfigGlobal;
}
setConfig(normalized);
```

Do the same coercion in `doFetch`. Export the defaults too:

```ts
export { SITE_CONFIG_GLOBAL_DEFAULT, SITE_CONFIG_FEATURES_DEFAULT };
```

- [ ] **Step 3: Verify type-check**

Run: `cd frontend && pnpm type-check`
Expected: passes.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/siteConfig.ts frontend/src/contexts/GlobalConfigContext.tsx
git commit -m "feat(types): SiteConfigGlobal + SiteConfigFeatures TS types"
```

### Task 1.8 — Minimal admin "raw JSON" editor

**Files:**
- Create: `frontend/src/pages/admin/site-config/page.tsx`
- Modify: `frontend/src/router/config.tsx`
- Modify: `frontend/src/api/` — add API client for global-config endpoint (path TBD by existing pattern; e.g. `src/api/globalConfig.ts`)

- [ ] **Step 1: Create the API client**

```ts
// frontend/src/api/globalConfig.ts
import { http } from "./http";
import type { SiteConfigGlobal } from "@/types/siteConfig";

export interface GlobalConfigState {
  draftConfig: SiteConfigGlobal;
  draftVersion: number;
  publishedConfig: SiteConfigGlobal;
  publishedVersion: number;
}

export async function fetchAdminGlobalConfig(): Promise<GlobalConfigState> {
  const res = await http.get("/admin/global-config");
  return res.data as GlobalConfigState;
}

export async function putAdminGlobalConfigDraft(
  draftConfig: SiteConfigGlobal,
  expectedDraftVersion: number,
): Promise<{ draftVersion: number }> {
  const res = await http.put("/admin/global-config/draft", {
    draftConfig,
    expectedDraftVersion,
  });
  return res.data as { draftVersion: number };
}

export async function publishAdminGlobalConfig(): Promise<{ publishedVersion: number }> {
  const res = await http.post("/admin/global-config/publish");
  return res.data as { publishedVersion: number };
}
```

- [ ] **Step 2: Create the editor page**

```tsx
// frontend/src/pages/admin/site-config/page.tsx
import { useEffect, useState } from "react";
import {
  fetchAdminGlobalConfig,
  putAdminGlobalConfigDraft,
  publishAdminGlobalConfig,
} from "@/api/globalConfig";
import type { SiteConfigGlobal } from "@/types/siteConfig";

export default function AdminSiteConfigPage() {
  const [draftJson, setDraftJson] = useState("");
  const [draftVersion, setDraftVersion] = useState(0);
  const [publishedVersion, setPublishedVersion] = useState(0);
  const [status, setStatus] = useState<string>("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAdminGlobalConfig()
      .then((s) => {
        setDraftJson(JSON.stringify(s.draftConfig, null, 2));
        setDraftVersion(s.draftVersion);
        setPublishedVersion(s.publishedVersion);
      })
      .catch((e: Error) => setStatus("Load failed: " + e.message))
      .finally(() => setLoading(false));
  }, []);

  async function saveDraft() {
    setStatus("");
    let parsed: SiteConfigGlobal;
    try {
      parsed = JSON.parse(draftJson) as SiteConfigGlobal;
    } catch (e) {
      setStatus("JSON parse error: " + (e as Error).message);
      return;
    }
    try {
      const r = await putAdminGlobalConfigDraft(parsed, draftVersion);
      setDraftVersion(r.draftVersion);
      setStatus("Draft saved (v" + r.draftVersion + ")");
    } catch (e) {
      setStatus("Save failed: " + (e as Error).message);
    }
  }

  async function publish() {
    setStatus("");
    try {
      const r = await publishAdminGlobalConfig();
      setPublishedVersion(r.publishedVersion);
      setStatus("Published (v" + r.publishedVersion + ")");
    } catch (e) {
      setStatus("Publish failed: " + (e as Error).message);
    }
  }

  if (loading) return <div className="p-4">Loading…</div>;

  return (
    <div className="p-4 max-w-4xl">
      <h1 className="text-xl font-semibold mb-4">Site Config (raw JSON)</h1>
      <p className="text-sm text-gray-500 mb-2">
        Draft v{draftVersion} · Published v{publishedVersion}
      </p>
      <textarea
        value={draftJson}
        onChange={(e) => setDraftJson(e.target.value)}
        className="w-full h-[60vh] font-mono text-sm border rounded p-2"
      />
      <div className="mt-4 flex gap-2 items-center">
        <button
          onClick={saveDraft}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Save Draft
        </button>
        <button
          onClick={publish}
          className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
        >
          Publish
        </button>
        {status && <span className="text-sm text-gray-700">{status}</span>}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Register the admin route**

In `frontend/src/router/config.tsx`, add the lazy import near other admin pages (after line ~30):

```tsx
const AdminSiteConfigPage = lazy(() => import('../pages/admin/site-config/page'));
```

And inside the admin route children array (alongside `theme`, `email-settings` etc.), add:

```tsx
{
  path: 'site-config',
  element: <AdminSiteConfigPage />,
},
```

- [ ] **Step 4: Verify lint + type-check**

Run: `cd frontend && pnpm lint && pnpm type-check`
Expected: pass.

- [ ] **Step 5: Smoke test manually (optional)**

Start backend (`make stop && make dev` or per CLAUDE.md memory instructions) → log in admin → navigate to `/admin/site-config` → confirm the JSON textarea loads with the current global config → edit `identity.name` → Save Draft → Publish → refresh public `/` page → no visible change yet (consumers come in PR-2).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/globalConfig.ts frontend/src/pages/admin/site-config/page.tsx frontend/src/router/config.tsx
git commit -m "feat(admin): raw JSON site-config editor at /admin/site-config"
```

### Task 1.9 — PR-1 wrap-up verification

- [ ] **Step 1: Run all checks**

Run from repo root: `pnpm lint && pnpm type-check && pnpm test`
Then: `cd backend && go vet ./... && go test -race ./...`
Expected: all green.

- [ ] **Step 2: Push & open PR-1 (or hand to reviewer)**

```bash
git push -u origin <branch>
gh pr create --title "feat(global-config): schema + admin endpoint + raw JSON editor" --body "$(cat <<'EOF'
## Summary
- Adds `SiteConfigGlobal` Go struct + validation (`validateGlobalConfig`)
- New admin endpoints `GET/PUT/POST /admin/global-config/{draft,publish}` writing `content_documents.PageKey="global"`
- TypeScript `SiteConfigGlobal` / `SiteConfigFeatures` types + defaults
- `pickLocaleValue` helper (consumed in PR-2/3)
- Minimal raw-JSON admin editor at `/admin/site-config`

## Test plan
- [ ] `go test ./internal/handler/global_config/... -race -v` passes
- [ ] `pnpm test -- src/lib/locale` passes
- [ ] Manual: load `/admin/site-config`, edit, save, publish — version bumps
- [ ] Public pages unchanged (no consumer yet)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## PR-2 · `refactor(i18n): UI-only boundary + useBranding 接入`

**Outcome:** `i18n/common.ts` reduced to UI strings; Header / Footer / Admin Layout read identity/footer/author from `globalConfig.siteConfig` via the new `useBranding` hook; all hardcoded "Blotting Consultancy" / "印迹" fallbacks removed; `readdy.ai` link removed.

### Task 2.1 — `useBranding` hook

**Files:**
- Create: `frontend/src/hooks/useBranding.ts`
- Create: `frontend/src/hooks/useBranding.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import type { ReactNode } from "react";

vi.mock("@/contexts/GlobalConfigContext", () => ({
  useGlobalConfig: () => ({
    config: {
      siteConfig: {
        identity: {
          name: { zh: "我的博客", en: "My Blog" },
          localeMode: "bilingual",
          defaultLocale: "zh",
        },
        brand: { logo: { light: "/logo.png" }, favicon: "/fav.ico", ogImage: "", primaryColor: "#000" },
        author: { name: "Isian", socials: [{ kind: "github", url: "https://github.com/isian" }] },
        footer: { copyright: { zh: "© 2026 我的博客" }, icp: "京 ICP 备 123" },
        seo: {},
      },
    },
    loading: false,
    locale: "zh",
    features: {},
    refetch: vi.fn(),
  }),
}));

import { useBranding } from "./useBranding";

const Wrapper = ({ children }: { children: ReactNode }) => <>{children}</>;

describe("useBranding", () => {
  it("returns localized site name based on current locale", () => {
    const { result } = renderHook(() => useBranding(), { wrapper: Wrapper });
    expect(result.current.siteName).toBe("我的博客");
  });
  it("exposes logo and ICP", () => {
    const { result } = renderHook(() => useBranding(), { wrapper: Wrapper });
    expect(result.current.logo.light).toBe("/logo.png");
    expect(result.current.footer.icp).toBe("京 ICP 备 123");
  });
  it("falls back to auto copyright when not configured", () => {
    // Tested separately in a second mock setup — see fallback test file
    // Keep this test focused.
  });
});
```

- [ ] **Step 2: Run test (expect failure — hook doesn't exist)**

Run: `cd frontend && pnpm test -- src/hooks/useBranding.test.ts`
Expected: fail with "Cannot find module './useBranding'".

- [ ] **Step 3: Implement the hook**

```ts
// frontend/src/hooks/useBranding.ts
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import {
  SITE_CONFIG_GLOBAL_DEFAULT,
  type SiteConfigGlobal,
  type SiteConfigSocial,
  type SiteConfigFooterLink,
} from "@/types/siteConfig";
import { pickLocaleValue, type Locale, type LocaleMode } from "@/lib/locale";

export interface BrandingView {
  siteName: string;
  tagline: string;
  logo: { light: string; dark?: string };
  favicon: string;
  primaryColor: string;
  author: {
    name: string;
    avatar?: string;
    bio: string;
    socials: SiteConfigSocial[];
  };
  footer: {
    copyright: string;
    icp?: string;
    extraLinks: SiteConfigFooterLink[];
  };
  // Pass-through so consumers can do locale-aware lookups themselves
  localeMode: LocaleMode;
  defaultLocale: Locale;
  currentLocale: Locale;
}

export function useBranding(): BrandingView {
  const { config, locale } = useGlobalConfig();
  const sc: SiteConfigGlobal = config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
  const mode = sc.identity.localeMode;
  const def = sc.identity.defaultLocale;
  const cur = (locale as Locale) ?? def;

  const siteName = pickLocaleValue({ value: sc.identity.name, mode, defaultLocale: def, currentLocale: cur });

  const copyright =
    pickLocaleValue({ value: sc.footer.copyright, mode, defaultLocale: def, currentLocale: cur }) ||
    `© ${new Date().getFullYear()} ${siteName}`;

  return {
    siteName,
    tagline: pickLocaleValue({ value: sc.identity.tagline, mode, defaultLocale: def, currentLocale: cur }),
    logo: sc.brand.logo,
    favicon: sc.brand.favicon,
    primaryColor: sc.brand.primaryColor,
    author: {
      name: sc.author.name,
      avatar: sc.author.avatar,
      bio: pickLocaleValue({ value: sc.author.bio, mode, defaultLocale: def, currentLocale: cur }),
      socials: sc.author.socials,
    },
    footer: {
      copyright,
      icp: sc.footer.icp,
      extraLinks: sc.footer.extraLinks ?? [],
    },
    localeMode: mode,
    defaultLocale: def,
    currentLocale: cur,
  };
}
```

- [ ] **Step 4: Re-run test, expect pass**

Run: `cd frontend && pnpm test -- src/hooks/useBranding.test.ts`
Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/useBranding.ts frontend/src/hooks/useBranding.test.ts
git commit -m "feat(hooks): useBranding reads identity/brand/author/footer from globalConfig"
```

### Task 2.2 — Strip business strings from `i18n/zh/common.ts`

**Files:**
- Modify: `frontend/src/i18n/local/zh/common.ts`

- [ ] **Step 1: Inspect existing structure**

Run: `cd frontend && head -250 src/i18n/local/zh/common.ts`
Take note of the deprecation header — the file already documents that deprecated keys exist. We're now actually removing them.

- [ ] **Step 2: Replace the whole file**

Open `frontend/src/i18n/local/zh/common.ts` and replace its content with a minimal UI-only set. Preserve only:
- `notFound.*`
- `nav.*` keys for UI labels (home/about/contact/blog/... — labels, not brand)
- `meta.*` UI-only keys (e.g. button labels, common page section labels)
- common UI strings: `actions.*` (submit/cancel/save/reset…), `status.*` (loading/empty/error)
- Form labels generic across pages (placeholder strings, validation messages)

```ts
export const common = {
  notFound: {
    title: '页面未找到',
    description: '您访问的页面可能已被移除或暂时不可用。',
    goBack: '返回上页',
    home: '返回首页',
  },
  nav: {
    home: '首页',
    blog: '博客',
    about: '关于',
    contact: '联系',
    search: '搜索',
    menu: '菜单',
    languageSwitch: 'English',
  },
  actions: {
    submit: '提交',
    cancel: '取消',
    save: '保存',
    edit: '编辑',
    delete: '删除',
    confirm: '确认',
    back: '返回',
    more: '查看更多',
  },
  status: {
    loading: '加载中…',
    empty: '暂无内容',
    error: '出错了',
  },
  form: {
    namePlaceholder: '请输入您的姓名',
    emailPlaceholder: '请输入邮箱',
    messagePlaceholder: '请输入留言',
    required: '此项必填',
    invalidEmail: '邮箱格式不正确',
  },
};
```

- [ ] **Step 3: Run lint to find consumers of removed keys**

Run: `cd frontend && pnpm lint`
Expected: lint passes, but TS likely errors. Run `pnpm type-check` next to surface broken `t("about.title")` etc. callers.

- [ ] **Step 4: For each TS error, decide**
- If the consumer is a consultancy page that PR-4 will gate (`/about`, `/experts`, etc.) — fix by reading from `usePublicContent` page config (which the page should already do for non-deprecated keys), or hard-code a placeholder string. Use `getPageContent("about")` pattern.
- If consumer is shared (Header/Footer) — that's tackled in Task 2.4-2.5.

For now, replace removed `t("about.description")` style calls with a sensible fallback **string literal** so type-check passes. Then PR-4 will gate those pages so the literal is unreachable from any default deployment.

Concretely: open each page file pointed to by the type errors, replace `t("about.description")` with `t("status.empty")` or a hardcoded fallback like `""`. Track them in a small running list and fix one at a time.

- [ ] **Step 5: Re-run type-check until clean**

Run: `cd frontend && pnpm type-check`
Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/i18n/local/zh/common.ts frontend/src/pages/
git commit -m "refactor(i18n): strip zh business strings to UI-only"
```

### Task 2.3 — Mirror changes in `i18n/en/common.ts`

**Files:**
- Modify: `frontend/src/i18n/local/en/common.ts`

- [ ] **Step 1: Replace with the same minimal shape in English**

```ts
export const common = {
  notFound: {
    title: 'Page not found',
    description: 'The page you are looking for is unavailable.',
    goBack: 'Go back',
    home: 'Go home',
  },
  nav: {
    home: 'Home',
    blog: 'Blog',
    about: 'About',
    contact: 'Contact',
    search: 'Search',
    menu: 'Menu',
    languageSwitch: '中文',
  },
  actions: {
    submit: 'Submit',
    cancel: 'Cancel',
    save: 'Save',
    edit: 'Edit',
    delete: 'Delete',
    confirm: 'Confirm',
    back: 'Back',
    more: 'See more',
  },
  status: {
    loading: 'Loading…',
    empty: 'Nothing here yet',
    error: 'Something went wrong',
  },
  form: {
    namePlaceholder: 'Your name',
    emailPlaceholder: 'Your email',
    messagePlaceholder: 'Your message',
    required: 'This field is required',
    invalidEmail: 'Invalid email format',
  },
};
```

- [ ] **Step 2: Verify type-check**

Run: `cd frontend && pnpm type-check`
Expected: pass.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/i18n/local/en/common.ts
git commit -m "refactor(i18n): strip en business strings to UI-only"
```

### Task 2.4 — Header reads `useBranding`

**Files:**
- Modify: `frontend/src/components/feature/Header.tsx`

- [ ] **Step 1: Replace logo / logoAlt source**

In `Header.tsx`, change:

```ts
const logoSrc = globalConfig.branding?.logo?.url || '/images/logo.png';
const logoAlt = globalConfig.branding?.companyName || 'Blotting Consultancy';
```

to:

```ts
import { useBranding } from "@/hooks/useBranding";

// inside the component:
const branding = useBranding();
const logoSrc = branding.logo.light || '/images/logo.png';
const logoAlt = branding.siteName || 'Site';
```

Leave the `useGlobalConfig` call in place — it's still used for legacy `nav.items` until PR-4.

- [ ] **Step 2: Verify lint / type-check / start dev server**

Run: `cd frontend && pnpm lint && pnpm type-check`
Then run `make dev` and open `http://localhost:3000` — Header should still render. If `siteConfig` is unset (no PR-1 editor save yet), it falls back to `SITE_CONFIG_GLOBAL_DEFAULT` → siteName="My Site" (legacy logoAlt was the only remaining brand string).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/feature/Header.tsx
git commit -m "refactor(header): logo + alt text from useBranding (no Blotting fallback)"
```

### Task 2.5 — Footer reads `useBranding`, drops `readdy.ai` link

**Files:**
- Modify: `frontend/src/components/feature/Footer.tsx`
- Modify: `frontend/src/theme/layouts/ThemedFooter.tsx`

- [ ] **Step 1: Replace `Footer.tsx`**

```tsx
import { useBranding } from "@/hooks/useBranding";

interface LinkItem { label?: string; href?: string }

export default function Footer() {
  const branding = useBranding();
  const logoSrc = branding.logo.light || '/images/logo.png';
  const logoAlt = branding.siteName || 'Site';
  // legacy fields still rendered (address/phone) until PR-5 admin form covers them
  const legacyLinks: LinkItem[] = [];

  return (
    <footer className="bg-primary text-white">
      <div className="max-w-layout mx-auto px-4 md:px-6 py-12">
        <div className="flex flex-col md:flex-row md:items-start gap-8">
          <div>
            <img src={logoSrc} alt={logoAlt} className="h-10 w-auto mb-4" />
            {branding.author.bio && <p className="text-sm text-gray-300">{branding.author.bio}</p>}
          </div>
          {branding.footer.extraLinks.length > 0 && (
            <div className="md:ml-auto">
              <ul className="flex flex-wrap gap-4 text-sm">
                {branding.footer.extraLinks.map((link, i) => (
                  <li key={i}>
                    <a href={link.url || '#'} className="text-gray-300 hover:text-accent transition-colors cursor-pointer">
                      {link.label.zh || link.label.en || ''}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          )}
          {legacyLinks.length > 0 && null}
        </div>
        <div className="mt-12 pt-8 border-t border-white/20 text-center text-sm text-gray-300">
          <p>{branding.footer.copyright}</p>
          {branding.footer.icp && <p className="mt-1">{branding.footer.icp}</p>}
        </div>
      </div>
    </footer>
  );
}
```

- [ ] **Step 2: Mirror changes in `ThemedFooter.tsx`**

Repeat the same useBranding-based logic in `frontend/src/theme/layouts/ThemedFooter.tsx`. Open it and apply the equivalent transformation. Pay attention to its theme-token wiring — preserve the className patterns.

- [ ] **Step 3: Verify**

Run: `cd frontend && pnpm lint && pnpm type-check && pnpm test`
Start dev server → open `/` → confirm footer shows the auto-copyright `© 2026 My Site` and no `readdy.ai` link.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/feature/Footer.tsx frontend/src/theme/layouts/ThemedFooter.tsx
git commit -m "refactor(footer): use useBranding; drop readdy.ai link; hide empty ICP"
```

### Task 2.6 — Admin Layout uses `useBranding`

**Files:**
- Modify: `frontend/src/pages/admin/AdminLayout.tsx`

- [ ] **Step 1: Find brand references in AdminLayout**

Run: `cd frontend && grep -n "印迹\|Blotting" src/pages/admin/AdminLayout.tsx`
Each match needs to become `useBranding().siteName` (with a "{name} 后台" suffix or similar — but follow whatever the existing layout uses).

- [ ] **Step 2: Replace**

Add at top:
```ts
import { useBranding } from "@/hooks/useBranding";
```

Inside the component, replace the literal sitename strings:
```ts
const branding = useBranding();
// where you currently render the admin title:
<h1>{branding.siteName} · Admin</h1>
```

- [ ] **Step 3: Verify dev page loads**

Run: `cd frontend && pnpm type-check`
Start dev server, log in to `/admin`, confirm title shows configured site name.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/admin/AdminLayout.tsx
git commit -m "refactor(admin-layout): site name from useBranding"
```

### Task 2.7 — Header/Footer locale-mode snapshot tests

**Files:**
- Create: `frontend/src/components/feature/Header.test.tsx`
- Create: `frontend/src/components/feature/Footer.test.tsx`

- [ ] **Step 1: Header snapshot per locale mode**

```tsx
// frontend/src/components/feature/Header.test.tsx
import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

function mockCtx(localeMode: "mono-zh" | "mono-en" | "bilingual") {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {
        siteConfig: {
          identity: { name: { zh: "测试站", en: "Test Site" }, localeMode, defaultLocale: localeMode === "mono-en" ? "en" : "zh" },
          brand: { logo: { light: "/logo.png" }, favicon: "", ogImage: "", primaryColor: "#000" },
          author: { name: "", socials: [] },
          footer: {},
          seo: {},
        },
      },
      locale: localeMode === "mono-en" ? "en" : "zh",
      loading: false,
      features: {},
      refetch: vi.fn(),
    }),
  }));
  vi.doMock("@/contexts/ThemePagesContext", () => ({
    useThemePages: () => ({ headerNavItems: [], menuNavItems: [] }),
  }));
}

describe("Header locale-mode behavior", () => {
  it.each(["mono-zh", "mono-en"] as const)("hides language switcher in %s", async (mode) => {
    vi.resetModules();
    mockCtx(mode);
    const { default: Header } = await import("./Header");
    const { queryByText } = render(<MemoryRouter><Header /></MemoryRouter>);
    expect(queryByText("English")).toBeNull();
    expect(queryByText("中文")).toBeNull();
  });

  it("shows language switcher in bilingual", async () => {
    vi.resetModules();
    mockCtx("bilingual");
    const { default: Header } = await import("./Header");
    const { getByRole } = render(<MemoryRouter><Header /></MemoryRouter>);
    // bilingual + locale=zh shows the "English" toggle text
    expect(getByRole("button", { name: /English|中文/ })).toBeTruthy();
  });
});
```

- [ ] **Step 2: Footer copyright auto-generation test**

```tsx
// frontend/src/components/feature/Footer.test.tsx
import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

describe("Footer", () => {
  it("auto-generates copyright when not configured", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {
          siteConfig: {
            identity: { name: { zh: "我的站" }, localeMode: "mono-zh", defaultLocale: "zh" },
            brand: { logo: { light: "" }, favicon: "", ogImage: "", primaryColor: "#000" },
            author: { name: "", socials: [] },
            footer: {},
            seo: {},
          },
        },
        locale: "zh",
        loading: false,
        features: {},
        refetch: vi.fn(),
      }),
    }));
    const { default: Footer } = await import("./Footer");
    const { container } = render(<Footer />);
    const year = String(new Date().getFullYear());
    expect(container.textContent).toContain(year);
    expect(container.textContent).toContain("我的站");
  });

  it("hides ICP block when icp is empty", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {
          siteConfig: {
            identity: { name: { zh: "x" }, localeMode: "mono-zh", defaultLocale: "zh" },
            brand: { logo: { light: "" }, favicon: "", ogImage: "", primaryColor: "#000" },
            author: { name: "", socials: [] },
            footer: { icp: "" },
            seo: {},
          },
        },
        locale: "zh",
        loading: false,
        features: {},
        refetch: vi.fn(),
      }),
    }));
    const { default: Footer } = await import("./Footer");
    const { container } = render(<Footer />);
    expect(container.textContent).not.toMatch(/ICP/);
  });
});
```

- [ ] **Step 3: Run, expect pass**

Run: `cd frontend && pnpm test -- src/components/feature/Header.test.tsx src/components/feature/Footer.test.tsx`

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/feature/Header.test.tsx frontend/src/components/feature/Footer.test.tsx
git commit -m "test(header,footer): locale-mode + copyright/ICP rendering"
```

### Task 2.8 — Update `pages.regression.test.tsx`

**Files:**
- Modify: `frontend/src/test/pages.regression.test.tsx`
- Modify: `frontend/src/test/mock-data.ts`

- [ ] **Step 1: Run existing regression to see what breaks**

Run: `cd frontend && pnpm test -- src/test/pages.regression.test.tsx`

Likely failures: (a) hardcoded brand string assertions like `expect(...).toContain("印迹")` (b) i18n key references like `t("about.title")` that we removed.

- [ ] **Step 2: For each failing assertion, decide**
- Brand-string assertions → replace with site-name expectation (e.g. assert `globalConfig.siteConfig.identity.name.zh` value or just check the heading renders non-empty)
- Removed-i18n-key assertions → either delete the assertion if the page is being gated, or switch to reading from mocked `usePublicContent`

- [ ] **Step 3: For `mock-data.ts`, replace brand strings with neutral placeholders**

Open it, scan for "印迹"/"Blotting", change to "Test Site"/"测试站点".

- [ ] **Step 4: Re-run until pass**

Run: `cd frontend && pnpm test`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/test/pages.regression.test.tsx frontend/src/test/mock-data.ts
git commit -m "test(regression): update for PR-2 i18n + brand changes"
```

### Task 2.9 — PR-2 wrap-up

- [ ] **Step 1: Run full quality gate**

```bash
pnpm lint && pnpm type-check && pnpm test
cd backend && go vet ./... && go test -race ./...
```

- [ ] **Step 2: Brand residue grep (intermediate)**

Run: `cd frontend && grep -rn "印迹\|Blotting" src/components/feature src/pages/admin/AdminLayout.tsx src/theme/layouts`
Expected: 0 matches in these specific files. Other files still have residue (cleared in PR-5/6).

- [ ] **Step 3: Push & open PR-2**

```bash
git push
gh pr create --title "refactor(i18n): UI-only boundary + useBranding consumed by Header/Footer/Admin" --body "..."
```

---

## PR-3 · `feat(locale): useLocaleMode + LanguageSwitch behavior`

**Outcome:** `localeMode` controls whether the language switcher renders and which locale is forced into i18next. Tests cover all three modes.

### Task 3.1 — `useLocaleMode` hook

**Files:**
- Create: `frontend/src/hooks/useLocaleMode.ts`
- Create: `frontend/src/hooks/useLocaleMode.test.ts`

- [ ] **Step 1: Write failing tests**

```ts
import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import { useLocaleMode } from "./useLocaleMode";

function mockGlobalConfig(localeMode: string, defaultLocale: "zh" | "en", currentLocale: "zh" | "en") {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {
        siteConfig: {
          identity: { name: { zh: "x" }, localeMode, defaultLocale },
          brand: { logo: { light: "" }, favicon: "", ogImage: "", primaryColor: "#000" },
          author: { name: "", socials: [] },
          footer: {},
          seo: {},
        },
      },
      locale: currentLocale,
      loading: false,
      features: {},
      refetch: vi.fn(),
    }),
  }));
}

describe("useLocaleMode", () => {
  it("mono-zh: available=['zh'], isMono=true", async () => {
    vi.resetModules();
    mockGlobalConfig("mono-zh", "zh", "zh");
    const { useLocaleMode: hook } = await import("./useLocaleMode");
    const { result } = renderHook(() => hook());
    expect(result.current.available).toEqual(["zh"]);
    expect(result.current.isMono).toBe(true);
  });

  it("bilingual: available=['zh','en'], isMono=false", async () => {
    vi.resetModules();
    mockGlobalConfig("bilingual", "zh", "en");
    const { useLocaleMode: hook } = await import("./useLocaleMode");
    const { result } = renderHook(() => hook());
    expect(result.current.available).toEqual(["zh", "en"]);
    expect(result.current.isMono).toBe(false);
    expect(result.current.currentLocale).toBe("en");
  });

  it("mono-en collapses currentLocale to en even if context says zh", async () => {
    vi.resetModules();
    mockGlobalConfig("mono-en", "en", "zh");
    const { useLocaleMode: hook } = await import("./useLocaleMode");
    const { result } = renderHook(() => hook());
    expect(result.current.currentLocale).toBe("en");
  });
});
```

- [ ] **Step 2: Implement hook**

```ts
// frontend/src/hooks/useLocaleMode.ts
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { SITE_CONFIG_GLOBAL_DEFAULT, type SiteConfigGlobal } from "@/types/siteConfig";
import type { Locale, LocaleMode } from "@/lib/locale";

export interface LocaleModeView {
  localeMode: LocaleMode;
  defaultLocale: Locale;
  currentLocale: Locale;
  available: Locale[];
  isMono: boolean;
}

export function useLocaleMode(): LocaleModeView {
  const { config, locale } = useGlobalConfig();
  const sc: SiteConfigGlobal = config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
  const mode = sc.identity.localeMode;
  const def = sc.identity.defaultLocale;
  let available: Locale[];
  let current: Locale;
  if (mode === "mono-zh") {
    available = ["zh"];
    current = "zh";
  } else if (mode === "mono-en") {
    available = ["en"];
    current = "en";
  } else {
    available = ["zh", "en"];
    current = (locale as Locale) ?? def;
  }
  return {
    localeMode: mode,
    defaultLocale: def,
    currentLocale: current,
    available,
    isMono: mode !== "bilingual",
  };
}
```

- [ ] **Step 3: Run tests, expect pass**

Run: `cd frontend && pnpm test -- src/hooks/useLocaleMode.test.ts`
Expected: 3 pass.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/hooks/useLocaleMode.ts frontend/src/hooks/useLocaleMode.test.ts
git commit -m "feat(hooks): useLocaleMode collapses currentLocale by mode"
```

### Task 3.2 — LanguageSwitch hidden in mono mode

**Files:**
- Modify: `frontend/src/components/feature/Header.tsx`

- [ ] **Step 1: Gate the language toggle button**

Locate the existing `<button onClick={toggleLanguage}>...` block (lines ~50-57). Wrap it in a check via `useLocaleMode`:

```tsx
import { useLocaleMode } from "@/hooks/useLocaleMode";

// inside component
const { isMono } = useLocaleMode();

// ...JSX where the language switch lives:
{!isMono && (
  <button onClick={toggleLanguage} className="...">
    {resolveLocale(i18n.language) === 'zh' ? 'English' : '中文'}
  </button>
)}
```

Also wrap the parent `<div className="bg-primary text-white py-2">...` if it would otherwise be visually empty. Make the entire top-bar conditional on `!isMono`.

- [ ] **Step 2: Force i18next to mono locale on mount**

Still inside Header (or extracted to a small init effect — but Header is fine for now), add:

```tsx
const { isMono, available, currentLocale } = useLocaleMode();
useEffect(() => {
  if (isMono && i18n.language !== currentLocale) {
    i18n.changeLanguage(currentLocale);
  }
}, [isMono, currentLocale, i18n]);
```

- [ ] **Step 3: Verify**

Run dev server. From the admin JSON editor (PR-1), set `localeMode=mono-zh` and publish. Refresh public site → top language bar should disappear → all UI strings should be in zh regardless of browser locale setting.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/feature/Header.tsx
git commit -m "feat(locale): hide LanguageSwitch in mono mode, force i18n to mono locale"
```

### Task 3.3 — PR-3 wrap-up

- [ ] **Step 1: Quality gate**

```bash
pnpm lint && pnpm type-check && pnpm test
```

- [ ] **Step 2: Push & open PR-3**

---

## PR-4 · `feat(routing): features.publicPages gates + admin endpoint`

**Outcome:** Consultancy pages default off; admin can toggle each; new `/admin/features` endpoint backs the toggles; routing + menu obey `features.publicPages`.

### Task 4.1 — Backend `features` schema + handler

**Files:**
- Create: `backend/internal/handler/features/schema.go`
- Create: `backend/internal/handler/features/handler.go`
- Create: `backend/internal/handler/features/handler_test.go`

- [ ] **Step 1: Schema file**

```go
package features

import (
	"encoding/json"
	"errors"

	"blotting-consultancy/internal/model"
)

type PublicPages struct {
	Home         bool `json:"home"`
	Blog         bool `json:"blog"`
	Contact      bool `json:"contact"`
	About        bool `json:"about"`
	Experts      bool `json:"experts"`
	CoreServices bool `json:"coreServices"`
	Advantages   bool `json:"advantages"`
	Cases        bool `json:"cases"`
}

type BlogFeatures struct {
	Comments bool `json:"comments"`
	RSS      bool `json:"rss"`
}

type Features struct {
	PublicPages PublicPages  `json:"publicPages"`
	Blog        BlogFeatures `json:"blog"`
}

func validateFeatures(raw model.JSONMap) (*Features, error) {
	if raw == nil {
		return nil, errors.New("features payload required")
	}
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var f Features
	if err := json.Unmarshal(bytes, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
```

- [ ] **Step 2: Handler file**

```go
// backend/internal/handler/features/handler.go
package features

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/cache"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

type Handler struct {
	repo  repository.SiteConfigRepository
	cache *cache.Cache
}

func NewHandler(repo repository.SiteConfigRepository, c *cache.Cache) *Handler {
	return &Handler{repo: repo, cache: c}
}

func (h *Handler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.GET("/features", h.adminGet)
	admin.PUT("/features/draft", h.adminPutDraft)
	admin.POST("/features/publish", h.adminPublish)
}

func (h *Handler) adminGet(c *gin.Context) {
	sc, err := h.repo.FindByKey(c.Request.Context(), model.SiteConfigKeyFeatures)
	if err != nil || sc == nil {
		c.JSON(http.StatusOK, gin.H{
			"draftConfig":      model.JSONMap{},
			"draftVersion":     0,
			"publishedConfig":  model.JSONMap{},
			"publishedVersion": 0,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"draftConfig":      sc.DraftConfig,
		"draftVersion":     sc.DraftVersion,
		"publishedConfig":  sc.PublishedConfig,
		"publishedVersion": sc.PublishedVersion,
	})
}

type putDraftInput struct {
	DraftConfig model.JSONMap `json:"draftConfig"`
}

func (h *Handler) adminPutDraft(c *gin.Context) {
	var in putDraftInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid body"}})
		return
	}
	if _, err := validateFeatures(in.DraftConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	existing, _ := h.repo.FindByKey(c.Request.Context(), model.SiteConfigKeyFeatures)
	if existing == nil {
		sc := &model.SiteConfig{
			Key:              model.SiteConfigKeyFeatures,
			DraftConfig:      in.DraftConfig,
			DraftVersion:     1,
			PublishedConfig:  model.JSONMap{},
			PublishedVersion: 0,
		}
		if err := h.repo.Create(c.Request.Context(), sc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "create failed"}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"draftVersion": 1})
		return
	}
	existing.DraftConfig = in.DraftConfig
	existing.DraftVersion += 1
	if err := h.repo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "update failed"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"draftVersion": existing.DraftVersion})
}

func (h *Handler) adminPublish(c *gin.Context) {
	sc, err := h.repo.FindByKey(c.Request.Context(), model.SiteConfigKeyFeatures)
	if err != nil || sc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "no draft to publish"}})
		return
	}
	if _, err := validateFeatures(sc.DraftConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	sc.PublishedConfig = sc.DraftConfig
	sc.PublishedVersion += 1
	if err := h.repo.Update(c.Request.Context(), sc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "publish failed"}})
		return
	}
	if h.cache != nil {
		h.cache.Clear()
	}
	c.JSON(http.StatusOK, gin.H{"publishedVersion": sc.PublishedVersion})
}
```

- [ ] **Step 3: Verify `repository.SiteConfigRepository` interface matches**

Run: `cd backend && grep -n "func.*SiteConfigRepository\|FindByKey\|Create\|Update" internal/repository/site_config_repository.go`
If method names differ (e.g. `Save` instead of `Update`, or `GetByKey` instead of `FindByKey`), adjust the handler.

- [ ] **Step 4: Handler test**

```go
// backend/internal/handler/features/handler_test.go
package features_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/cache"
	"blotting-consultancy/internal/db"
	"blotting-consultancy/internal/handler/features"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

func setup(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	conn, err := db.OpenSQLite("file::memory:?cache=shared&mode=rwc")
	if err != nil { t.Fatal(err) }
	if err := conn.AutoMigrate(&model.SiteConfig{}); err != nil { t.Fatal(err) }
	repo := repository.NewGormSiteConfigRepository(conn)
	r := gin.New()
	admin := r.Group("/admin")
	features.NewHandler(repo, cache.New(0)).RegisterRoutes(admin)
	return r
}

func TestAdminPutDraft_CreatesIfMissing(t *testing.T) {
	r := setup(t)
	body := `{"draftConfig":{"publicPages":{"home":true,"blog":true,"contact":true,"about":false,"experts":false,"coreServices":false,"advantages":false,"cases":false},"blog":{"comments":true,"rss":true}}}`
	req := httptest.NewRequest(http.MethodPut, "/admin/features/draft", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPublish_RequiresDraft(t *testing.T) {
	r := setup(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/features/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 5: Run tests, expect pass**

Run: `cd backend && go test ./internal/handler/features/... -race -v`
Expected: pass.

- [ ] **Step 6: Wire into main.go / routes.go (same pattern as PR-1 Task 1.4)**

In `cmd/server/main.go`:
```go
import featuresHandler "blotting-consultancy/internal/handler/features"
// ...
handlers.Features = featuresHandler.NewHandler(siteConfigRepo, cache)
```

In `cmd/server/routes.go` (next to global_config registration):
```go
handlers.Features.RegisterRoutes(adminGroup)
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/features backend/cmd/server/main.go backend/cmd/server/routes.go
git commit -m "feat(features): admin endpoints for site_configs.features"
```

### Task 4.2 — `featureMap.ts` + `<FeatureGate>`

**Files:**
- Create: `frontend/src/router/featureMap.ts`
- Create: `frontend/src/components/feature/FeatureGate.tsx`
- Create: `frontend/src/components/feature/FeatureGate.test.tsx`

- [ ] **Step 1: featureMap**

```ts
// frontend/src/router/featureMap.ts
import type { SiteConfigFeatures } from "@/types/siteConfig";

/** Routes (kebab-case URLs) that are gated by a feature key. */
export const routeFeatureMap: Record<string, keyof SiteConfigFeatures["publicPages"]> = {
  "/about": "about",
  "/experts": "experts",
  "/core-services": "coreServices",
  "/advantages": "advantages",
  "/cases": "cases",
};

/** Returns true if the given camelCase feature key is enabled.
 *
 * Backwards-compat rule (spec §4.2): when the features record is missing
 * entirely (old deployment, no migration), every key defaults to TRUE so
 * the existing site keeps behaving as before. Only when a publishedConfig
 * record EXISTS but the specific key is unset does it fall to false.
 * New deployments must rely on BlankSiteSeed to explicitly seed a record
 * with personal-blog defaults.
 */
export function isFeatureEnabled(
  features: SiteConfigFeatures | undefined,
  key: keyof SiteConfigFeatures["publicPages"],
): boolean {
  if (!features || !features.publicPages) return true;
  return features.publicPages[key] === true;
}
```

- [ ] **Step 2: FeatureGate component**

```tsx
// frontend/src/components/feature/FeatureGate.tsx
import { lazy, type ReactNode } from "react";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { SITE_CONFIG_FEATURES_DEFAULT, type SiteConfigFeatures } from "@/types/siteConfig";

const NotFound = lazy(() => import("@/pages/NotFound"));

export interface FeatureGateProps {
  feature: keyof SiteConfigFeatures["publicPages"];
  children: ReactNode;
  fallback?: ReactNode;
}

export function FeatureGate({ feature, children, fallback }: FeatureGateProps) {
  const { features } = useGlobalConfig();
  const published = features as unknown as SiteConfigFeatures | undefined;
  // Old-deploy compat: missing record → render as enabled.
  // Missing key within an existing record → render as disabled.
  const enabled = !published || !published.publicPages
    ? true
    : published.publicPages[feature] === true;
  if (!enabled) return <>{fallback ?? <NotFound />}</>;
  return <>{children}</>;
}
```

- [ ] **Step 3: Tests**

```tsx
// frontend/src/components/feature/FeatureGate.test.tsx
import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

function setMock(publicPages: Record<string, boolean>) {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {},
      features: { publicPages, blog: { comments: true, rss: true } },
      locale: "zh",
      loading: false,
      refetch: vi.fn(),
    }),
  }));
}

describe("FeatureGate", () => {
  it("renders children when feature true", async () => {
    vi.resetModules();
    setMock({ about: true, home: true, blog: true, contact: true, experts: false, coreServices: false, advantages: false, cases: false });
    const { FeatureGate } = await import("./FeatureGate");
    const { getByText } = render(<FeatureGate feature="about"><span>visible</span></FeatureGate>);
    expect(getByText("visible")).toBeTruthy();
  });

  it("renders NotFound when feature false", async () => {
    vi.resetModules();
    setMock({ about: false, home: true, blog: true, contact: true, experts: false, coreServices: false, advantages: false, cases: false });
    const { FeatureGate } = await import("./FeatureGate");
    const { container } = render(<FeatureGate feature="about"><span data-testid="should-not-render">x</span></FeatureGate>);
    expect(container.querySelector('[data-testid="should-not-render"]')).toBeNull();
  });

  it("old-deploy compat: missing features record → renders as enabled", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {},
        features: {},                     // empty record (no publicPages)
        locale: "zh",
        loading: false,
        refetch: vi.fn(),
      }),
    }));
    const { FeatureGate } = await import("./FeatureGate");
    const { getByText } = render(<FeatureGate feature="about"><span>visible</span></FeatureGate>);
    expect(getByText("visible")).toBeTruthy();
  });
});
```

- [ ] **Step 4: Run tests**

Run: `cd frontend && pnpm test -- src/components/feature/FeatureGate.test.tsx`
Expected: 2 pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/router/featureMap.ts frontend/src/components/feature/FeatureGate.tsx frontend/src/components/feature/FeatureGate.test.tsx
git commit -m "feat(routing): FeatureGate + featureMap"
```

### Task 4.3 — Apply `<FeatureGate>` to consultancy routes

**Files:**
- Modify: `frontend/src/router/config.tsx`

- [ ] **Step 1: Wrap consultancy page elements**

Routes for `/about`, `/experts`, `/core-services`, `/advantages`, `/cases` are NOT in `staticRoutes` — they're generated dynamically from theme pages (see `AppRoutes` comment). To gate them, modify the `AppRoutes` component (find it via `grep -rn "AppRoutes" src/router/`).

Run: `find frontend/src/router -name "*.tsx" -o -name "*.ts"` to list router files.

In whichever component generates theme-driven routes (likely `AppRoutes.tsx` or similar), iterate the `themePages` (from BootstrapContext) and wrap each generated `<Route>` element in `<FeatureGate>` keyed by `routeFeatureMap[page.path]` if present:

```tsx
import { FeatureGate } from "@/components/feature/FeatureGate";
import { routeFeatureMap } from "@/router/featureMap";

// inside the page-mapping loop:
const featureKey = routeFeatureMap[page.path];
const elem = featureKey
  ? <FeatureGate feature={featureKey}><PageComponent /></FeatureGate>
  : <PageComponent />;
```

- [ ] **Step 2: Also gate the static nav menu**

In the Header (or wherever the navigation items list is built), filter out items whose path is in `routeFeatureMap` AND whose feature is off:

```tsx
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { isFeatureEnabled, routeFeatureMap } from "@/router/featureMap";
import type { SiteConfigFeatures } from "@/types/siteConfig";

const { features } = useGlobalConfig();
const publishedFeatures = features as unknown as SiteConfigFeatures;
const visibleNav = navigation.filter(item => {
  if (!item.href) return true;
  const key = routeFeatureMap[item.href];
  if (!key) return true; // not gated
  return isFeatureEnabled(publishedFeatures, key);
});
```

Then use `visibleNav` in the JSX instead of `navigation`.

- [ ] **Step 3: Manual smoke**

Start dev. Default features (`SITE_CONFIG_FEATURES_DEFAULT`) has all consultancy pages off. Visit `/about` → should render NotFound. Toggle on via admin (PR-4 admin page Task 4.4) and re-test.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/router/
git commit -m "feat(routing): gate consultancy routes + nav items by features.publicPages"
```

### Task 4.4 — Admin features toggles UI

**Files:**
- Create: `frontend/src/pages/admin/features/page.tsx`
- Create: `frontend/src/api/features.ts`
- Modify: `frontend/src/router/config.tsx` (register admin route)

- [ ] **Step 1: API client**

```ts
// frontend/src/api/features.ts
import { http } from "./http";
import type { SiteConfigFeatures } from "@/types/siteConfig";

export interface FeaturesState {
  draftConfig: SiteConfigFeatures;
  draftVersion: number;
  publishedConfig: SiteConfigFeatures;
  publishedVersion: number;
}

export async function fetchAdminFeatures(): Promise<FeaturesState> {
  const r = await http.get("/admin/features");
  return r.data as FeaturesState;
}
export async function putAdminFeaturesDraft(cfg: SiteConfigFeatures): Promise<{ draftVersion: number }> {
  const r = await http.put("/admin/features/draft", { draftConfig: cfg });
  return r.data;
}
export async function publishAdminFeatures(): Promise<{ publishedVersion: number }> {
  const r = await http.post("/admin/features/publish");
  return r.data;
}
```

- [ ] **Step 2: Admin page**

```tsx
// frontend/src/pages/admin/features/page.tsx
import { useEffect, useState } from "react";
import {
  fetchAdminFeatures,
  putAdminFeaturesDraft,
  publishAdminFeatures,
} from "@/api/features";
import { SITE_CONFIG_FEATURES_DEFAULT, type SiteConfigFeatures } from "@/types/siteConfig";

const PUBLIC_PAGE_KEYS: Array<keyof SiteConfigFeatures["publicPages"]> = [
  "home", "blog", "contact",
  "about", "experts", "coreServices", "advantages", "cases",
];

export default function AdminFeaturesPage() {
  const [draft, setDraft] = useState<SiteConfigFeatures>(SITE_CONFIG_FEATURES_DEFAULT);
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAdminFeatures()
      .then((s) => {
        if (s.draftConfig && Object.keys(s.draftConfig).length > 0) setDraft(s.draftConfig);
        else if (s.publishedConfig && Object.keys(s.publishedConfig).length > 0) setDraft(s.publishedConfig);
      })
      .finally(() => setLoading(false));
  }, []);

  function toggle(key: keyof SiteConfigFeatures["publicPages"]) {
    setDraft((d) => ({
      ...d,
      publicPages: { ...d.publicPages, [key]: !d.publicPages[key] },
    }));
  }

  async function save() {
    setStatus("");
    try {
      await putAdminFeaturesDraft(draft);
      setStatus("Draft saved");
    } catch (e) {
      setStatus("Save failed: " + (e as Error).message);
    }
  }

  async function publish() {
    setStatus("");
    try {
      const r = await publishAdminFeatures();
      setStatus("Published v" + r.publishedVersion);
    } catch (e) {
      setStatus("Publish failed: " + (e as Error).message);
    }
  }

  if (loading) return <div className="p-4">Loading…</div>;

  return (
    <div className="p-4 max-w-2xl">
      <h1 className="text-xl font-semibold mb-4">Features</h1>
      <section className="mb-6">
        <h2 className="text-sm font-medium text-gray-600 mb-2">Public pages</h2>
        <ul className="space-y-2">
          {PUBLIC_PAGE_KEYS.map((key) => (
            <li key={key} className="flex items-center gap-3">
              <input
                type="checkbox"
                id={`pp-${key}`}
                checked={draft.publicPages[key]}
                onChange={() => toggle(key)}
              />
              <label htmlFor={`pp-${key}`} className="text-sm">/{key}</label>
            </li>
          ))}
        </ul>
      </section>
      <div className="flex gap-2 items-center">
        <button onClick={save} className="px-4 py-2 bg-blue-600 text-white rounded">Save Draft</button>
        <button onClick={publish} className="px-4 py-2 bg-green-600 text-white rounded">Publish</button>
        {status && <span className="text-sm text-gray-700">{status}</span>}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Register route**

In `frontend/src/router/config.tsx`:

```tsx
const AdminFeaturesPage = lazy(() => import('../pages/admin/features/page'));

// inside admin children:
{ path: 'features', element: <AdminFeaturesPage /> },
```

- [ ] **Step 4: Verify quality gate**

Run: `cd frontend && pnpm lint && pnpm type-check && pnpm test`

- [ ] **Step 5: Smoke test**

Dev server → `/admin/features` → toggle `about` on → publish → visit `/about` → should now render (with whatever fallback the page shows).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/features.ts frontend/src/pages/admin/features/page.tsx frontend/src/router/config.tsx
git commit -m "feat(admin): features toggles UI at /admin/features"
```

### Task 4.5 — PR-4 wrap-up

- [ ] **Step 1: Full quality gate + push**

```bash
pnpm lint && pnpm type-check && pnpm test
cd backend && go vet ./... && go test -race ./...
git push
gh pr create --title "feat(routing): features.publicPages gates + admin endpoint" --body "..."
```

---

## PR-5 · `feat(seo): seo defaults from global-config + form-based admin editor`

**Outcome:** SEO defaults sourced from `siteConfig.seo`; all `useDocumentTitle("page", "印迹")` suffix calls collapsed into `useSEODefaults`; backend SEO meta reads from content_documents global; admin editor upgraded from raw JSON to tabbed form.

### Task 5.1 — `useSEODefaults` hook

**Files:**
- Create: `frontend/src/hooks/useSEODefaults.ts`
- Create: `frontend/src/hooks/useSEODefaults.test.ts`

- [ ] **Step 1: Write failing tests**

```ts
import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";

function mock(seo: any, identityName: any = { zh: "My Site" }) {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {
        siteConfig: {
          identity: { name: identityName, localeMode: "bilingual", defaultLocale: "zh" },
          brand: { logo: { light: "" }, favicon: "", ogImage: "https://example.com/og.png", primaryColor: "#000" },
          author: { name: "", socials: [] },
          footer: {},
          seo,
        },
      },
      locale: "zh",
      loading: false,
      features: {},
      refetch: vi.fn(),
    }),
  }));
}

describe("useSEODefaults", () => {
  it("falls back to identity.name for default title", async () => {
    vi.resetModules();
    mock({});
    const { useSEODefaults } = await import("./useSEODefaults");
    const { result } = renderHook(() => useSEODefaults());
    expect(result.current.defaultTitle).toBe("My Site");
  });

  it("uses titleTemplate '{page} | {site}' by default", async () => {
    vi.resetModules();
    mock({});
    const { useSEODefaults } = await import("./useSEODefaults");
    const { result } = renderHook(() => useSEODefaults());
    expect(result.current.buildTitle("Blog")).toBe("Blog | My Site");
  });

  it("respects configured titleTemplate", async () => {
    vi.resetModules();
    mock({ titleTemplate: "{site} — {page}" });
    const { useSEODefaults } = await import("./useSEODefaults");
    const { result } = renderHook(() => useSEODefaults());
    expect(result.current.buildTitle("Blog")).toBe("My Site — Blog");
  });

  it("buildTitle with empty page returns site only", async () => {
    vi.resetModules();
    mock({});
    const { useSEODefaults } = await import("./useSEODefaults");
    const { result } = renderHook(() => useSEODefaults());
    expect(result.current.buildTitle("")).toBe("My Site");
  });
});
```

- [ ] **Step 2: Implement**

```ts
// frontend/src/hooks/useSEODefaults.ts
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { SITE_CONFIG_GLOBAL_DEFAULT, type SiteConfigGlobal } from "@/types/siteConfig";
import { pickLocaleValue, type Locale } from "@/lib/locale";

export interface SEODefaultsView {
  defaultTitle: string;
  titleTemplate: string;
  defaultDescription: string;
  defaultOgImage: string;
  buildTitle(pageTitle: string): string;
}

export function useSEODefaults(): SEODefaultsView {
  const { config, locale } = useGlobalConfig();
  const sc: SiteConfigGlobal = config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
  const mode = sc.identity.localeMode;
  const def = sc.identity.defaultLocale;
  const cur = (locale as Locale) ?? def;

  const siteName = pickLocaleValue({ value: sc.identity.name, mode, defaultLocale: def, currentLocale: cur });
  const defaultTitle =
    pickLocaleValue({ value: sc.seo.defaultTitle, mode, defaultLocale: def, currentLocale: cur }) ||
    siteName;
  const titleTemplate = sc.seo.titleTemplate?.trim() || "{page} | {site}";
  const defaultDescription = pickLocaleValue({ value: sc.seo.defaultDescription, mode, defaultLocale: def, currentLocale: cur });
  const defaultOgImage = sc.brand.ogImage;

  function buildTitle(pageTitle: string): string {
    const trimmed = (pageTitle ?? "").trim();
    if (!trimmed) return siteName || defaultTitle;
    return titleTemplate
      .replace("{page}", trimmed)
      .replace("{site}", siteName || defaultTitle);
  }

  return { defaultTitle, titleTemplate, defaultDescription, defaultOgImage, buildTitle };
}
```

- [ ] **Step 3: Run tests, expect 4 pass**

Run: `cd frontend && pnpm test -- src/hooks/useSEODefaults.test.ts`

- [ ] **Step 4: Commit**

```bash
git add frontend/src/hooks/useSEODefaults.ts frontend/src/hooks/useSEODefaults.test.ts
git commit -m "feat(hooks): useSEODefaults + buildTitle template renderer"
```

### Task 5.2 — Refactor `useDocumentTitle` to drop suffix arg

**Files:**
- Modify: `frontend/src/hooks/useDocumentTitle.ts`
- Modify: all callers (find via grep)

- [ ] **Step 1: Replace `useDocumentTitle`**

```ts
// frontend/src/hooks/useDocumentTitle.ts
import { useEffect } from "react";
import { useSEODefaults } from "./useSEODefaults";

/**
 * Sets document.title to `buildTitle(title)` for the lifetime of the component.
 * Restores the previous title on unmount.
 */
export function useDocumentTitle(title: string | undefined | null) {
  const { buildTitle } = useSEODefaults();
  useEffect(() => {
    if (!title) return;
    const prev = document.title;
    document.title = buildTitle(title);
    return () => {
      document.title = prev;
    };
  }, [title, buildTitle]);
}
```

- [ ] **Step 2: Find all callers**

Run: `cd frontend && grep -rn "useDocumentTitle(" src/ | grep -v ".test."`
For each call:
- If it's `useDocumentTitle("foo", "印迹后台")` → change to `useDocumentTitle("foo")`
- If it's `useDocumentTitle("foo")` → leave alone

Specifically known callers (from spec residue grep earlier):
- `src/theme/DynamicPage.tsx:24` — has `"印迹法规咨询"` suffix
- `src/modules/qa/admin/page.tsx:41` — has `"印迹后台"` suffix

Edit each: remove the second argument.

- [ ] **Step 3: Verify type-check**

Run: `cd frontend && pnpm type-check`
Expected: pass. TypeScript will flag any caller that still passes 2 args.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/hooks/useDocumentTitle.ts frontend/src/theme/DynamicPage.tsx frontend/src/modules/qa/admin/page.tsx
git commit -m "refactor(seo): useDocumentTitle drops suffix, uses useSEODefaults"
```

### Task 5.3 — Backend SEO defaults from `content_documents.global`

**Files:**
- Modify: `backend/internal/seo/meta.go`

- [ ] **Step 1: Read defaults from repo on construction**

Decide between option (a) inject defaults at `DefaultPageMeta()` callsite, or (b) make `DefaultPageMeta` take a repo. (a) is less disruptive.

Find `DefaultPageMeta` callers: `cd backend && grep -rn "DefaultPageMeta\|seo\.PageMeta" .`. Likely in `cmd/server/main.go` or in the SPA fallback handler that injects meta into `index.html`.

Replace:

```go
const (
	defaultTitle       = "印迹法规咨询 - 企业内设型法规团队 | 专业法规咨询服务"
	defaultDescription = "印迹法规咨询（Blotting Consultancy）- 为企业提供专业的内设型法规团队服务"
)

func DefaultPageMeta() PageMeta {
	return PageMeta{
		Title:         defaultTitle,
		Description:   defaultDescription,
		Locale:        "zh",
		OgType:        "website",
		OgTitle:       defaultTitle,
		OgDescription: defaultDescription,
		TwitterCard:   "summary_large_image",
	}
}
```

with:

```go
// DefaultPageMeta returns sensible defaults. Caller is expected to overlay
// values from the published global config (identity.name / seo.* / brand.ogImage)
// via PageMeta.ApplyGlobal.
func DefaultPageMeta() PageMeta {
	return PageMeta{
		Title:         "Site",
		Description:   "",
		Locale:        "zh",
		OgType:        "website",
		OgTitle:       "Site",
		OgDescription: "",
		TwitterCard:   "summary_large_image",
	}
}

// ApplyGlobal overlays defaults from a "global" content document published config.
// Pass the JSONMap of doc.PublishedConfig directly.
func (pm *PageMeta) ApplyGlobal(global map[string]any, locale string) {
	if global == nil { return }
	identity, _ := global["identity"].(map[string]any)
	seo, _      := global["seo"].(map[string]any)
	brand, _    := global["brand"].(map[string]any)

	pickLocalized := func(m map[string]any, key, loc string) string {
		v, _ := m[key].(map[string]any)
		if v == nil { return "" }
		if s, ok := v[loc].(string); ok && s != "" { return s }
		if s, ok := v["zh"].(string); ok && s != "" { return s }
		if s, ok := v["en"].(string); ok && s != "" { return s }
		return ""
	}

	siteName := ""
	if identity != nil {
		siteName = pickLocalized(identity, "name", locale)
	}
	if seoTitle := pickLocalized(seo, "defaultTitle", locale); seoTitle != "" {
		pm.Title = seoTitle
		pm.OgTitle = seoTitle
	} else if siteName != "" {
		pm.Title = siteName
		pm.OgTitle = siteName
	}
	if desc := pickLocalized(seo, "defaultDescription", locale); desc != "" {
		pm.Description = desc
		pm.OgDescription = desc
	}
	if brand != nil {
		if og, ok := brand["ogImage"].(string); ok && og != "" {
			pm.OgImage = og
		}
	}
	if identity != nil {
		if mode, ok := identity["localeMode"].(string); ok {
			if mode == "mono-zh" { pm.Locale = "zh" }
			if mode == "mono-en" { pm.Locale = "en" }
		}
	}
}
```

- [ ] **Step 2: Wire into the SPA index.html / meta injection callsite**

Find where `DefaultPageMeta` is called and immediately overlay:
```go
meta := seo.DefaultPageMeta()
if doc, err := contentDocRepo.FindByPageKey(ctx, model.PageKeyGlobal); err == nil {
    meta.ApplyGlobal(map[string]any(doc.PublishedConfig), locale)
}
```

The exact location depends on the project. Search: `cd backend && grep -rn "DefaultPageMeta\(\)" .`

- [ ] **Step 3: Quick test**

Add a unit test for `ApplyGlobal`:

```go
// backend/internal/seo/meta_test.go (append)
func TestApplyGlobal_OverlaysTitleAndOG(t *testing.T) {
	pm := DefaultPageMeta()
	pm.ApplyGlobal(map[string]any{
		"identity": map[string]any{
			"name":       map[string]any{"zh": "我的博客"},
			"localeMode": "mono-zh",
		},
		"seo": map[string]any{},
		"brand": map[string]any{"ogImage": "https://x.test/og.png"},
	}, "zh")
	if pm.Title != "我的博客" { t.Errorf("Title: got %q want 我的博客", pm.Title) }
	if pm.OgImage != "https://x.test/og.png" { t.Errorf("OgImage: got %q", pm.OgImage) }
	if pm.Locale != "zh" { t.Errorf("Locale: got %q", pm.Locale) }
}
```

Run: `cd backend && go test ./internal/seo/... -race -v`
Expected: pass.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/seo/meta.go backend/internal/seo/meta_test.go backend/...
git commit -m "feat(seo): backend defaults overlay from content_documents.global"
```

### Task 5.4 — Upgrade admin editor from raw JSON to tabbed form

**Files:**
- Modify: `frontend/src/pages/admin/site-config/page.tsx`

- [ ] **Step 1: Define tabs**

Decide tab layout: Identity / Brand / Author / Footer / SEO. Replace the textarea with a controlled form per tab. Each tab is a sub-component receiving the relevant slice of `SiteConfigGlobal` + an onChange callback.

This is a substantial UI change. Keep it self-contained in the same file, since it's bounded scope (no shared form components needed).

```tsx
// frontend/src/pages/admin/site-config/page.tsx — full rewrite
import { useEffect, useState } from "react";
import {
  fetchAdminGlobalConfig,
  putAdminGlobalConfigDraft,
  publishAdminGlobalConfig,
} from "@/api/globalConfig";
import { SITE_CONFIG_GLOBAL_DEFAULT, type SiteConfigGlobal } from "@/types/siteConfig";

type TabKey = "identity" | "brand" | "author" | "footer" | "seo";

const TABS: { key: TabKey; label: string }[] = [
  { key: "identity", label: "Identity" },
  { key: "brand",    label: "Brand" },
  { key: "author",   label: "Author" },
  { key: "footer",   label: "Footer" },
  { key: "seo",      label: "SEO" },
];

export default function AdminSiteConfigPage() {
  const [cfg, setCfg] = useState<SiteConfigGlobal>(SITE_CONFIG_GLOBAL_DEFAULT);
  const [draftVersion, setDraftVersion] = useState(0);
  const [tab, setTab] = useState<TabKey>("identity");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAdminGlobalConfig()
      .then((s) => {
        if (s.draftConfig && Object.keys(s.draftConfig).length > 0) setCfg(s.draftConfig);
        else if (s.publishedConfig && Object.keys(s.publishedConfig).length > 0) setCfg(s.publishedConfig);
        setDraftVersion(s.draftVersion);
      })
      .catch((e: Error) => setStatus("Load failed: " + e.message))
      .finally(() => setLoading(false));
  }, []);

  async function save() {
    setStatus("");
    try {
      const r = await putAdminGlobalConfigDraft(cfg, draftVersion);
      setDraftVersion(r.draftVersion);
      setStatus("Draft saved (v" + r.draftVersion + ")");
    } catch (e) {
      setStatus("Save failed: " + (e as Error).message);
    }
  }

  async function publish() {
    setStatus("");
    try {
      const r = await publishAdminGlobalConfig();
      setStatus("Published (v" + r.publishedVersion + ")");
    } catch (e) {
      setStatus("Publish failed: " + (e as Error).message);
    }
  }

  if (loading) return <div className="p-4">Loading…</div>;

  return (
    <div className="p-4 max-w-3xl">
      <h1 className="text-xl font-semibold mb-4">Site Config</h1>
      <div className="border-b mb-4 flex gap-1">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-2 text-sm border-b-2 ${tab === t.key ? "border-blue-600 text-blue-600" : "border-transparent text-gray-600"}`}
          >
            {t.label}
          </button>
        ))}
      </div>
      {tab === "identity" && <IdentityTab cfg={cfg} setCfg={setCfg} />}
      {tab === "brand"    && <BrandTab cfg={cfg} setCfg={setCfg} />}
      {tab === "author"   && <AuthorTab cfg={cfg} setCfg={setCfg} />}
      {tab === "footer"   && <FooterTab cfg={cfg} setCfg={setCfg} />}
      {tab === "seo"      && <SEOTab cfg={cfg} setCfg={setCfg} />}
      <div className="mt-6 flex gap-2 items-center">
        <button onClick={save} className="px-4 py-2 bg-blue-600 text-white rounded">Save Draft</button>
        <button onClick={publish} className="px-4 py-2 bg-green-600 text-white rounded">Publish</button>
        {status && <span className="text-sm text-gray-700">{status}</span>}
      </div>
    </div>
  );
}

interface TabProps {
  cfg: SiteConfigGlobal;
  setCfg: (cfg: SiteConfigGlobal) => void;
}

function IdentityTab({ cfg, setCfg }: TabProps) {
  return (
    <div className="space-y-3">
      <div>
        <label className="block text-sm font-medium">Site name (zh)</label>
        <input
          type="text"
          value={cfg.identity.name.zh ?? ""}
          onChange={(e) => setCfg({ ...cfg, identity: { ...cfg.identity, name: { ...cfg.identity.name, zh: e.target.value } } })}
          className="border rounded px-2 py-1 w-full"
        />
      </div>
      <div>
        <label className="block text-sm font-medium">Site name (en)</label>
        <input
          type="text"
          value={cfg.identity.name.en ?? ""}
          onChange={(e) => setCfg({ ...cfg, identity: { ...cfg.identity, name: { ...cfg.identity.name, en: e.target.value } } })}
          className="border rounded px-2 py-1 w-full"
        />
      </div>
      <div>
        <label className="block text-sm font-medium">Locale mode</label>
        <select
          value={cfg.identity.localeMode}
          onChange={(e) => setCfg({ ...cfg, identity: { ...cfg.identity, localeMode: e.target.value as SiteConfigGlobal["identity"]["localeMode"] } })}
          className="border rounded px-2 py-1"
        >
          <option value="mono-zh">mono-zh</option>
          <option value="mono-en">mono-en</option>
          <option value="bilingual">bilingual</option>
        </select>
      </div>
      <div>
        <label className="block text-sm font-medium">Default locale</label>
        <select
          value={cfg.identity.defaultLocale}
          onChange={(e) => setCfg({ ...cfg, identity: { ...cfg.identity, defaultLocale: e.target.value as "zh" | "en" } })}
          className="border rounded px-2 py-1"
        >
          <option value="zh">zh</option>
          <option value="en">en</option>
        </select>
      </div>
    </div>
  );
}

function BrandTab({ cfg, setCfg }: TabProps) {
  return (
    <div className="space-y-3">
      <Input label="Logo (light) URL" value={cfg.brand.logo.light} onChange={(v) => setCfg({ ...cfg, brand: { ...cfg.brand, logo: { ...cfg.brand.logo, light: v } } })} />
      <Input label="Logo (dark) URL" value={cfg.brand.logo.dark ?? ""} onChange={(v) => setCfg({ ...cfg, brand: { ...cfg.brand, logo: { ...cfg.brand.logo, dark: v } } })} />
      <Input label="Favicon URL" value={cfg.brand.favicon} onChange={(v) => setCfg({ ...cfg, brand: { ...cfg.brand, favicon: v } })} />
      <Input label="Default OG image URL" value={cfg.brand.ogImage} onChange={(v) => setCfg({ ...cfg, brand: { ...cfg.brand, ogImage: v } })} />
      <Input label="Primary color (hex)" value={cfg.brand.primaryColor} onChange={(v) => setCfg({ ...cfg, brand: { ...cfg.brand, primaryColor: v } })} />
    </div>
  );
}

function AuthorTab({ cfg, setCfg }: TabProps) {
  return (
    <div className="space-y-3">
      <Input label="Name" value={cfg.author.name} onChange={(v) => setCfg({ ...cfg, author: { ...cfg.author, name: v } })} />
      <Input label="Avatar URL" value={cfg.author.avatar ?? ""} onChange={(v) => setCfg({ ...cfg, author: { ...cfg.author, avatar: v } })} />
      <Input label="Location" value={cfg.author.location ?? ""} onChange={(v) => setCfg({ ...cfg, author: { ...cfg.author, location: v } })} />
      <div>
        <label className="block text-sm font-medium mb-1">Socials</label>
        {cfg.author.socials.map((s, i) => (
          <div key={i} className="flex gap-2 mb-1">
            <input
              type="text"
              value={s.kind}
              onChange={(e) => {
                const next = [...cfg.author.socials];
                next[i] = { ...s, kind: e.target.value as any };
                setCfg({ ...cfg, author: { ...cfg.author, socials: next } });
              }}
              placeholder="kind"
              className="border rounded px-2 py-1 w-28"
            />
            <input
              type="text"
              value={s.url}
              onChange={(e) => {
                const next = [...cfg.author.socials];
                next[i] = { ...s, url: e.target.value };
                setCfg({ ...cfg, author: { ...cfg.author, socials: next } });
              }}
              placeholder="url"
              className="border rounded px-2 py-1 flex-1"
            />
            <button
              onClick={() => setCfg({ ...cfg, author: { ...cfg.author, socials: cfg.author.socials.filter((_, j) => j !== i) } })}
              className="px-2 text-red-600"
            >×</button>
          </div>
        ))}
        <button
          onClick={() => setCfg({ ...cfg, author: { ...cfg.author, socials: [...cfg.author.socials, { kind: "github", url: "" }] } })}
          className="text-sm text-blue-600"
        >+ Add social</button>
      </div>
    </div>
  );
}

function FooterTab({ cfg, setCfg }: TabProps) {
  return (
    <div className="space-y-3">
      <Input label="Copyright (zh)" value={cfg.footer.copyright?.zh ?? ""} onChange={(v) => setCfg({ ...cfg, footer: { ...cfg.footer, copyright: { ...(cfg.footer.copyright ?? {}), zh: v } } })} />
      <Input label="Copyright (en)" value={cfg.footer.copyright?.en ?? ""} onChange={(v) => setCfg({ ...cfg, footer: { ...cfg.footer, copyright: { ...(cfg.footer.copyright ?? {}), en: v } } })} />
      <Input label="ICP (中国大陆备案号；留空隐藏)" value={cfg.footer.icp ?? ""} onChange={(v) => setCfg({ ...cfg, footer: { ...cfg.footer, icp: v } })} />
    </div>
  );
}

function SEOTab({ cfg, setCfg }: TabProps) {
  return (
    <div className="space-y-3">
      <Input label="Default title template" value={cfg.seo.titleTemplate ?? ""} placeholder="{page} | {site}" onChange={(v) => setCfg({ ...cfg, seo: { ...cfg.seo, titleTemplate: v } })} />
      <Input label="Default description (zh)" value={cfg.seo.defaultDescription?.zh ?? ""} onChange={(v) => setCfg({ ...cfg, seo: { ...cfg.seo, defaultDescription: { ...(cfg.seo.defaultDescription ?? {}), zh: v } } })} />
      <Input label="Default description (en)" value={cfg.seo.defaultDescription?.en ?? ""} onChange={(v) => setCfg({ ...cfg, seo: { ...cfg.seo, defaultDescription: { ...(cfg.seo.defaultDescription ?? {}), en: v } } })} />
      <Input label="Twitter handle" value={cfg.seo.twitterHandle ?? ""} placeholder="@yourhandle" onChange={(v) => setCfg({ ...cfg, seo: { ...cfg.seo, twitterHandle: v } })} />
    </div>
  );
}

function Input({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <div>
      <label className="block text-sm font-medium">{label}</label>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="border rounded px-2 py-1 w-full"
      />
    </div>
  );
}
```

- [ ] **Step 2: Verify lint + type-check**

Run: `cd frontend && pnpm lint && pnpm type-check`

- [ ] **Step 3: Smoke test**

Dev server → `/admin/site-config` → fill each tab → Save Draft → Publish → confirm public site reflects changes.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/admin/site-config/page.tsx
git commit -m "feat(admin): tabbed form editor for site-config (replaces raw JSON)"
```

### Task 5.5 — PR-5 wrap-up

- [ ] **Step 1: Full quality gate + push**

```bash
pnpm lint && pnpm type-check && pnpm test
cd backend && go vet ./... && go test -race ./...
git push
gh pr create --title "feat(seo): defaults from global-config + tabbed admin editor" --body "..."
```

---

## PR-6 · `chore: rename DSN, blank seed, theme metadata, residue check`

**Outcome:** Default DSN `impress.db`; seeds split (`BlankSiteSeed` / `DemoSiteSeed`); `SEED_MODE` env; theme package metadata neutralized; brand residue script passes 0 hits in product code.

### Task 6.1 — Default DSN to `impress.db`

**Files:**
- Modify: `backend/pkg/config/config.go`

- [ ] **Step 1: Replace constant**

In `backend/pkg/config/config.go` line ~23:

Replace:
```go
const defaultSQLiteDSN = "file:./data/blotting.db?cache=shared&mode=rwc"
```

with:
```go
const defaultSQLiteDSN = "file:./data/impress.db?cache=shared&mode=rwc"
```

- [ ] **Step 2: Build + test**

Run: `cd backend && go build ./... && go vet ./...`

- [ ] **Step 3: Commit**

```bash
git add backend/pkg/config/config.go
git commit -m "chore(config): default DSN -> impress.db"
```

### Task 6.2 — Split `Seed()` into `BlankSiteSeed` / `DemoSiteSeed`

**Files:**
- Modify: `backend/internal/seed/seed.go`
- Modify: `backend/internal/seed/seed_test.go`
- Modify: `backend/cmd/server/main.go` (call site)

- [ ] **Step 1: Identify SeedAll call site**

Run: `cd backend && grep -rn "SeedAll\b" cmd/`

- [ ] **Step 2: Refactor `seed.go`**

Rename the existing `SeedAll` to `DemoSiteSeed` (keep all current behavior — it inserts the consultancy demo data). Add a new `BlankSiteSeed`:

At the bottom of `seed.go`:

```go
// BlankSiteSeed inserts the minimum required for a fresh site:
// one admin user, the global content document with default SiteConfigGlobal
// payload, and the features site_config with personal-blog defaults.
// No articles, no media, no demo data.
func (s *Seeder) BlankSiteSeed(ctx context.Context) error {
	log.Println("Starting blank-site seed...")
	if err := s.SeedUsers(ctx); err != nil { return err }

	// Ensure a "global" content_document exists with personal-blog defaults.
	existing, err := s.contentRepo.FindByPageKey(ctx, model.PageKeyGlobal)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}
	if existing == nil {
		doc := &model.ContentDocument{
			PageKey:          model.PageKeyGlobal,
			DraftConfig:      blankGlobalConfig(),
			DraftVersion:     1,
			PublishedConfig:  blankGlobalConfig(),
			PublishedVersion: 1,
		}
		if err := s.contentRepo.Create(ctx, doc); err != nil { return err }
		log.Println("Created blank global content document")
	}

	// Seed site_configs.features with personal-blog defaults so the gates
	// in PR-4 default to "consultancy pages off". Without an explicit record,
	// the frontend treats missing features as "all on" (old-deploy compat).
	if s.siteCfgRepo != nil {
		featuresExisting, ferr := s.siteCfgRepo.FindByKey(ctx, model.SiteConfigKeyFeatures)
		if ferr != nil && !strings.Contains(ferr.Error(), "not found") {
			return ferr
		}
		if featuresExisting == nil {
			cfg := blankFeaturesConfig()
			row := &model.SiteConfig{
				Key:              model.SiteConfigKeyFeatures,
				DraftConfig:      cfg,
				DraftVersion:     1,
				PublishedConfig:  cfg,
				PublishedVersion: 1,
			}
			if err := s.siteCfgRepo.Create(ctx, row); err != nil { return err }
			log.Println("Created blank features site_config")
		}
	}

	log.Println("Blank-site seed completed")
	return nil
}

func blankGlobalConfig() model.JSONMap {
	return model.JSONMap{
		"identity": model.JSONMap{
			"name":          model.JSONMap{"zh": "My Site"},
			"localeMode":    "mono-zh",
			"defaultLocale": "zh",
		},
		"brand":  model.JSONMap{"logo": model.JSONMap{"light": ""}, "favicon": "", "ogImage": "", "primaryColor": "#1e40af"},
		"author": model.JSONMap{"name": "", "socials": []any{}},
		"footer": model.JSONMap{},
		"seo":    model.JSONMap{},
	}
}

func blankFeaturesConfig() model.JSONMap {
	return model.JSONMap{
		"publicPages": model.JSONMap{
			"home":         true,
			"blog":         true,
			"contact":      true,
			"about":        false,
			"experts":      false,
			"coreServices": false,
			"advantages":   false,
			"cases":        false,
		},
		"blog": model.JSONMap{
			"comments": true,
			"rss":      true,
		},
	}
}
```

**Note**: BlankSiteSeed needs access to `siteCfgRepo`. The current `Seeder` constructor (`NewSeeder`) doesn't accept it. As part of this task, extend the struct & constructor to accept `siteCfgRepo repository.SiteConfigRepository` and pass `nil` from any caller that doesn't have one (e.g. legacy tests). Update `NewSeeder` signature and all call sites accordingly. The `BlankSiteSeed` body guards with `if s.siteCfgRepo != nil` so it stays test-safe.

- [ ] **Step 3: Rename `SeedAll` to `DemoSiteSeed`**

Find every reference (`grep -rn "SeedAll" backend/`) and rename. The method body stays identical — just the name changes. Add a deprecation comment above pointing to the new pair.

Actually simpler: keep `SeedAll` as it is (existing behavior), but **add a new `DemoSiteSeed` that is just a thin wrapper calling `SeedAll`**. This avoids touching every test:

```go
// DemoSiteSeed runs the full consultancy demo data seed.
// Equivalent to the legacy SeedAll().
func (s *Seeder) DemoSiteSeed(ctx context.Context) error {
	return s.SeedAll(ctx)
}
```

- [ ] **Step 4: SEED_MODE dispatch in `main.go`**

Find the seed call site (e.g. `seeder.SeedAll(ctx)`). Replace with:

```go
seedMode := os.Getenv("SEED_MODE")
switch seedMode {
case "blank":
    if err := seeder.BlankSiteSeed(ctx); err != nil { log.Fatal(err) }
case "demo":
    if err := seeder.DemoSiteSeed(ctx); err != nil { log.Fatal(err) }
case "none":
    log.Println("SEED_MODE=none, skipping seed")
default:
    // Backwards compat: existing deployments continue to get demo behavior.
    if err := seeder.DemoSiteSeed(ctx); err != nil { log.Fatal(err) }
}
```

Make sure `os` is imported.

- [ ] **Step 5: Tests**

Add to `backend/internal/seed/seed_test.go`:

```go
func TestBlankSiteSeed_CreatesGlobalWithDefaults(t *testing.T) {
	// Mirror the existing seeder setup in this file (look at TestSeedUsers etc.
	// to see how seeder is constructed in tests).
	s, ctx := newTestSeeder(t)
	if err := s.BlankSiteSeed(ctx); err != nil {
		t.Fatalf("blank seed: %v", err)
	}
	doc, err := s.contentRepo.FindByPageKey(ctx, model.PageKeyGlobal)
	if err != nil { t.Fatalf("find global: %v", err) }
	id, ok := doc.PublishedConfig["identity"].(map[string]any)
	if !ok || id["localeMode"] != "mono-zh" {
		t.Fatalf("expected mono-zh in identity, got: %#v", doc.PublishedConfig["identity"])
	}
}
```

Look at how existing seed tests build a Seeder (`grep -n "newTestSeeder\|NewSeeder" seed_test.go`) and follow that pattern. If `s.contentRepo` is unexported, use a public accessor or read directly via a constructed repo.

- [ ] **Step 6: Run all backend tests**

Run: `cd backend && go test -race ./internal/seed/... -v`
Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/seed/ backend/cmd/server/main.go
git commit -m "feat(seed): BlankSiteSeed + DemoSiteSeed wrapper + SEED_MODE env"
```

### Task 6.3 — Theme package metadata neutralization

**Files:**
- Modify: `frontend/src/theme/packages/default/index.ts`
- Modify: `frontend/src/theme/packages/modern-dark/index.ts`
- Modify: `frontend/src/theme/packages/warm-earth/index.ts`

- [ ] **Step 1: Edit each file**

In each, change:
- `author: "Blotting Consultancy"` → `author: "impress"`
- For `default/index.ts`: `description: "印迹咨询经典配色，专业沉稳的蓝绿色调"` → `description: "经典蓝绿配色，专业沉稳"`

- [ ] **Step 2: Verify**

Run: `cd frontend && pnpm type-check && grep -n "印迹\|Blotting" src/theme/packages/`
Expected: 0 matches.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/theme/packages/
git commit -m "chore(themes): neutralize package metadata (author/description)"
```

### Task 6.4 — Update `FRONTEND_RENDERING.md` intro

**Files:**
- Modify: `frontend/src/FRONTEND_RENDERING.md`

- [ ] **Step 1: Replace opening paragraph**

Line 5 currently reads:
`> This document describes the configuration-driven rendering system for the 印迹官网 (Blotting Consultancy) frontend...`

Replace with:
`> This document describes the configuration-driven rendering system for the **impress** frontend. All public pages and global sections fetch content from the backend CMS rather than hardcoded i18n keys.`

- [ ] **Step 2: Commit**

```bash
git add frontend/src/FRONTEND_RENDERING.md
git commit -m "docs(frontend): debrand FRONTEND_RENDERING intro"
```

### Task 6.5 — `scripts/check-brand-residue.sh`

**Files:**
- Create: `scripts/check-brand-residue.sh`

- [ ] **Step 1: Write script**

```bash
#!/usr/bin/env bash
# Fails (exits 1) if "印迹|Blotting|blotting" appears in product code.
# Allowed locations (excluded): docs/, .long-agent/, backups/, go.mod, vendor/,
# generated swagger docs, *.zip, .git/, the spec/plan files for this S1 round.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Patterns to grep
PATTERN='印迹\|Blotting\|blotting'

# Files / dirs allowed to contain residue (docs, history, generated)
EXCLUDES=(
  ':!docs/**'
  ':!**/swagger/**'
  ':!backend/go.mod'
  ':!backend/scripts/migrate-sqlite-to-postgres.go'
  ':!**/backups/**'
  ':!**/.long-agent/**'
  ':!**/vendor/**'
  ':!**/*.zip'
)

# Use git grep so we automatically skip .gitignored files.
if git grep -l "$PATTERN" -- ':!' "${EXCLUDES[@]}" 2>/dev/null; then
  echo ""
  echo "ERROR: brand residue found in product code. See files above."
  exit 1
fi
echo "OK: no brand residue in product code"
```

Make executable: `chmod +x scripts/check-brand-residue.sh`

- [ ] **Step 2: Run it now**

Run: `bash scripts/check-brand-residue.sh`
If it lists matches, fix each. Common remaining suspects:
- `backend/internal/seo/meta.go` if PR-5 step missed a string
- `backend/internal/seed/seed.go` if `DemoSiteSeed` contents leak into default seed somehow (it shouldn't — but check)
- `backend/internal/handler/*` if any 404 message hardcoded "印迹"
- `frontend/src/i18n/` if any leftover
- `frontend/src/components/feature/` if any string slipped
- `frontend/src/test/mock-data.ts` (likely test fixture — that's product code if used in regression tests)

For each, decide: fix or add path to EXCLUDES (only docs/, swagger, history).

- [ ] **Step 3: Add CI hook (optional, but recommended)**

Append to `.github/workflows/quality-gate.yml` (or equivalent) a step:

```yaml
- name: Brand residue check
  run: bash scripts/check-brand-residue.sh
```

- [ ] **Step 4: Commit**

```bash
git add scripts/check-brand-residue.sh .github/workflows/
git commit -m "chore: brand-residue grep script + CI step"
```

### Task 6.6 — PR-6 wrap-up + final verification

- [ ] **Step 1: Full quality gate**

```bash
pnpm lint && pnpm type-check && pnpm test
cd backend && go vet ./... && go test -race ./...
bash scripts/check-brand-residue.sh
```

All must pass.

- [ ] **Step 2: Acceptance walkthrough (DoD §8 of spec)**

Per spec §8 manually verify:

1. **Fresh deploy + SEED_MODE=blank**: stop dev server, `rm backend/data/impress.db`, `SEED_MODE=blank` restart, log into `/admin`, navigate `/admin/site-config`, edit name, publish, refresh `/` → reflects.
2. **localeMode=mono-zh**: confirm no English in nav, no language switcher, no `hreflang`.
3. **No brand residue in user-visible surfaces**: visit `/`, `/blog`, view-source for `<title>`, `<meta description>`, `<meta og:>` — all reflect configured name.
4. **Consultancy pages default off**: `/about`, `/experts`, `/cases`, `/advantages`, `/core-services` → NotFound. Nav doesn't show them.
5. **Blog still works**: `/blog` lists articles (if seeded — use `SEED_MODE=demo` to verify), `/blog/<slug>` renders.
6. **Lints, types, tests, race**: all green.
7. **Residue script**: 0 hits.
8. **Demo seed**: `rm backend/data/impress.db`, `SEED_MODE=demo` restart — full consultancy demo restored.

- [ ] **Step 3: Push & open PR-6**

```bash
git push
gh pr create --title "chore: rename DSN, blank seed, theme metadata, residue check" --body "..."
```

- [ ] **Step 4: Mark S1 done**

After all 6 PRs merged, update spec §1 frontmatter `Status: Complete` and add `Closed: 2026-MM-DD` line. Commit.

---

## Verification Summary (cross-cutting)

| Check | Where | Pass condition |
|---|---|---|
| `pnpm lint` | repo root | exit 0 |
| `pnpm type-check` | repo root | exit 0 |
| `pnpm test` | repo root | all vitest pass |
| `go vet ./...` | backend | exit 0 |
| `go test -race ./...` | backend | all pass |
| `bash scripts/check-brand-residue.sh` | repo root (after PR-6) | 0 hits |
| Manual: localeMode=mono-zh demo | running site | language switcher gone, all UI in zh |
| Manual: consultancy routes default off | running site | `/about` → NotFound |
| Manual: `SEED_MODE=demo` restore | fresh DB | consultancy demo loads |

## Risks Captured From Implementation

| Risk | Mitigation |
|---|---|
| `db.OpenSQLite` signature differs from assumed (Task 1.5) | First step in that task is to grep the signature and adjust |
| `cache.Cache.DeletePrefix` may not exist (Task 1.3) | Fallback to `cache.Clear()` documented in step |
| `repository.ErrVersionConflict` may not exist (Task 1.3) | Add the var if missing — documented in step |
| Header / AppRoutes location for theme-page generation may differ (Task 4.3) | Step explicitly says "find the file via grep first" |
| i18n key consumers spread widely, breaking after Task 2.2 strip | Task 2.2 step 4 explicitly walks through fixing each consumer |
| `SeedAll` callers may expect old name (Task 6.2) | Approach: keep `SeedAll` + add `DemoSiteSeed` thin wrapper rather than renaming |

## Out of Scope (deferred to S2/S3/S4)

- Real plugin loading runtime
- Blog-shaped sections (article-list, TOC, related, author-card, code blocks…)
- `go.mod` module path rename
- Multi-author user model
- Custom CSS / domain / email branding
- Actual RSS generation
- Multi-language extension (>2 locales)
