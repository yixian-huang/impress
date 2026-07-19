package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// mockSiteRepository is a test double for repository.SiteRepository
type mockSiteRepository struct {
	byDomain  map[string]*model.Site
	bySubPath map[string]*model.Site
}

func newMockSiteRepo() *mockSiteRepository {
	return &mockSiteRepository{
		byDomain:  make(map[string]*model.Site),
		bySubPath: make(map[string]*model.Site),
	}
}

func (m *mockSiteRepository) Create(_ context.Context, _ *model.Site) error { return nil }
func (m *mockSiteRepository) FindByID(_ context.Context, _ uint) (*model.Site, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSiteRepository) FindByDomain(_ context.Context, domain string) (*model.Site, error) {
	if s, ok := m.byDomain[domain]; ok {
		return s, nil
	}
	return nil, errors.New("site not found")
}
func (m *mockSiteRepository) FindBySubPath(_ context.Context, subPath string) (*model.Site, error) {
	if s, ok := m.bySubPath[subPath]; ok {
		return s, nil
	}
	return nil, errors.New("site not found")
}
func (m *mockSiteRepository) Update(_ context.Context, _ *model.Site) error { return nil }
func (m *mockSiteRepository) Delete(_ context.Context, _ uint) error        { return nil }
func (m *mockSiteRepository) List(_ context.Context, _ string) ([]*model.Site, error) {
	return nil, nil
}
func (m *mockSiteRepository) AddUser(_ context.Context, _ *model.SiteUser) error { return nil }
func (m *mockSiteRepository) RemoveUser(_ context.Context, _, _ uint) error      { return nil }
func (m *mockSiteRepository) ListUsers(_ context.Context, _ uint) ([]*model.SiteUser, error) {
	return nil, nil
}
func (m *mockSiteRepository) FindUserRole(_ context.Context, _, _ uint) (*model.SiteUser, error) {
	return nil, errors.New("not implemented")
}

func TestSiteResolverByDomain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMockSiteRepo()
	repo.byDomain["example.com"] = &model.Site{ID: 1, Domain: "example.com", Name: "Example"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/some/path", nil)
	c.Request.Host = "example.com"

	var resolvedSite *model.Site
	SiteResolver(repo)(c)
	resolvedSite = GetSiteContext(c)

	if resolvedSite == nil {
		t.Fatal("expected site to be resolved, got nil")
	}
	if resolvedSite.ID != 1 {
		t.Errorf("expected site ID 1, got %d", resolvedSite.ID)
	}
}

func TestSiteResolverBySubPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMockSiteRepo()
	repo.bySubPath["/blog"] = &model.Site{
		ID:      2,
		Domain:  "example.com",
		SubPath: "/blog",
		Mode:    model.SiteModeSubpath,
		Name:    "Blog",
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/blog/post-1", nil)
	c.Request.Host = "example.com"

	SiteResolver(repo)(c)
	site := GetSiteContext(c)

	if site == nil {
		t.Fatal("expected site to be resolved from subpath, got nil")
	}
	if site.ID != 2 {
		t.Errorf("expected site ID 2, got %d", site.ID)
	}
}

func TestSiteResolverNoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMockSiteRepo()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "unknown.example.com"

	SiteResolver(repo)(c)
	site := GetSiteContext(c)

	if site != nil {
		t.Errorf("expected no site resolved, got %+v", site)
	}
}

func TestSiteResolverStripsPort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMockSiteRepo()
	repo.byDomain["localhost"] = &model.Site{ID: 3, Domain: "localhost", Name: "Local"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "localhost:8088"

	SiteResolver(repo)(c)
	site := GetSiteContext(c)

	if site == nil {
		t.Fatal("expected site to be resolved after stripping port, got nil")
	}
	if site.ID != 3 {
		t.Errorf("expected site ID 3, got %d", site.ID)
	}
}

func TestRequireSiteContextAborts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RequireSiteContext()(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRequireSiteContextPassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(string(SiteContextKey), &model.Site{ID: 1})

	called := false
	c.Set("next", func() { called = true })
	RequireSiteContext()(c)

	// The middleware should not abort; status stays 200 (default)
	if w.Code == http.StatusNotFound {
		t.Errorf("did not expect 404 when site context is present")
	}
	_ = called
}
