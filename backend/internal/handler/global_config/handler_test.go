package global_config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/model"
)

// MockContentDocumentRepository — minimal in-memory mock with per-method override functions.
type MockContentDocumentRepository struct {
	FindByPageKeyFunc   func(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error)
	UpdateDraftFunc     func(ctx context.Context, pageKey model.PageKey, expected int, draft model.JSONMap) (int, error)
	UpdatePublishedFunc func(ctx context.Context, pageKey model.PageKey, published model.JSONMap, version int) error
}

func (m *MockContentDocumentRepository) Create(ctx context.Context, doc *model.ContentDocument) error {
	return nil
}
func (m *MockContentDocumentRepository) FindByPageKey(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error) {
	if m.FindByPageKeyFunc != nil {
		return m.FindByPageKeyFunc(ctx, pageKey)
	}
	return nil, errors.New("not implemented")
}
func (m *MockContentDocumentRepository) Update(ctx context.Context, doc *model.ContentDocument) error {
	return nil
}
func (m *MockContentDocumentRepository) UpdateDraft(ctx context.Context, pageKey model.PageKey, expected int, draft model.JSONMap) (int, error) {
	if m.UpdateDraftFunc != nil {
		return m.UpdateDraftFunc(ctx, pageKey, expected, draft)
	}
	return expected + 1, nil
}
func (m *MockContentDocumentRepository) UpdatePublished(ctx context.Context, pageKey model.PageKey, published model.JSONMap, version int) error {
	if m.UpdatePublishedFunc != nil {
		return m.UpdatePublishedFunc(ctx, pageKey, published, version)
	}
	return nil
}
func (m *MockContentDocumentRepository) List(ctx context.Context) ([]*model.ContentDocument, error) {
	return nil, nil
}
func (m *MockContentDocumentRepository) Delete(ctx context.Context, pageKey model.PageKey) error {
	return nil
}

func newRouter(repo *MockContentDocumentRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin")
	NewHandler(repo, cache.New(0*time.Second)).RegisterRoutes(admin)
	return r
}

func validGlobalConfig() map[string]any {
	return map[string]any{
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

func TestAdminPutDraft_RejectsInvalidSchema(t *testing.T) {
	repo := &MockContentDocumentRepository{}
	r := newRouter(repo)
	body := `{"draftConfig":{"identity":{"name":{},"localeMode":"mono-zh","defaultLocale":"zh"},"brand":{},"author":{"socials":[]},"footer":{},"seo":{}},"expectedDraftVersion":1}`
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPutDraft_AcceptsValid(t *testing.T) {
	called := false
	repo := &MockContentDocumentRepository{
		UpdateDraftFunc: func(ctx context.Context, pageKey model.PageKey, expected int, draft model.JSONMap) (int, error) {
			called = true
			if pageKey != model.PageKeyGlobal {
				t.Errorf("expected PageKeyGlobal, got %q", pageKey)
			}
			return expected + 1, nil
		},
	}
	r := newRouter(repo)
	body, _ := json.Marshal(map[string]any{"draftConfig": validGlobalConfig(), "expectedDraftVersion": 1})
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatalf("UpdateDraft was not invoked")
	}
}

func TestAdminPutDraft_VersionConflictReturns409(t *testing.T) {
	repo := &MockContentDocumentRepository{
		UpdateDraftFunc: func(ctx context.Context, pageKey model.PageKey, expected int, draft model.JSONMap) (int, error) {
			return 0, errors.New("draft version conflict or document not found")
		},
	}
	r := newRouter(repo)
	body, _ := json.Marshal(map[string]any{"draftConfig": validGlobalConfig(), "expectedDraftVersion": 1})
	req := httptest.NewRequest(http.MethodPut, "/admin/global-config/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPublish_BumpsVersion(t *testing.T) {
	cfgMap := validGlobalConfig()
	// Convert to model.JSONMap for the mock doc
	cfgJsonMap := model.JSONMap(cfgMap)
	publishedCalled := false
	repo := &MockContentDocumentRepository{
		FindByPageKeyFunc: func(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error) {
			return &model.ContentDocument{
				PageKey:          model.PageKeyGlobal,
				DraftConfig:      cfgJsonMap,
				DraftVersion:     2,
				PublishedConfig:  model.JSONMap{},
				PublishedVersion: 1,
			}, nil
		},
		UpdatePublishedFunc: func(ctx context.Context, pageKey model.PageKey, published model.JSONMap, version int) error {
			publishedCalled = true
			if version != 2 {
				t.Errorf("expected published version 2, got %d", version)
			}
			return nil
		},
	}
	r := newRouter(repo)
	req := httptest.NewRequest(http.MethodPost, "/admin/global-config/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !publishedCalled {
		t.Fatalf("UpdatePublished was not invoked")
	}
}

func TestAdminGet_ReturnsBothDraftAndPublished(t *testing.T) {
	repo := &MockContentDocumentRepository{
		FindByPageKeyFunc: func(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error) {
			return &model.ContentDocument{
				PageKey:          model.PageKeyGlobal,
				DraftConfig:      model.JSONMap{"identity": map[string]any{"name": map[string]any{"zh": "draft"}}},
				DraftVersion:     3,
				PublishedConfig:  model.JSONMap{"identity": map[string]any{"name": map[string]any{"zh": "published"}}},
				PublishedVersion: 1,
			}, nil
		},
	}
	r := newRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/admin/global-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if int(resp["draftVersion"].(float64)) != 3 {
		t.Errorf("draftVersion not 3: %v", resp["draftVersion"])
	}
	if int(resp["publishedVersion"].(float64)) != 1 {
		t.Errorf("publishedVersion not 1: %v", resp["publishedVersion"])
	}
}
