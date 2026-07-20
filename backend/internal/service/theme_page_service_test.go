package service

import (
	"context"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

type stubPageRepo struct {
	byThemeContent map[string]*model.Page
	bySlug         map[string]*model.Page
	created        []*model.Page
	updated        []*model.Page
}

func (s *stubPageRepo) key(themeID, contentKey string) string {
	return themeID + ":" + contentKey
}

func (s *stubPageRepo) Create(_ context.Context, page *model.Page) error {
	s.created = append(s.created, page)
	if s.bySlug == nil {
		s.bySlug = map[string]*model.Page{}
	}
	s.bySlug[page.Slug] = page
	return nil
}

func (s *stubPageRepo) Update(_ context.Context, page *model.Page) error {
	s.updated = append(s.updated, page)
	if s.bySlug == nil {
		s.bySlug = map[string]*model.Page{}
	}
	s.bySlug[page.Slug] = page
	return nil
}

func (s *stubPageRepo) Delete(context.Context, uint) error { return nil }
func (s *stubPageRepo) FindByID(context.Context, uint) (*model.Page, error) {
	return nil, nil
}
func (s *stubPageRepo) FindBySlug(_ context.Context, slug string) (*model.Page, error) {
	if p, ok := s.bySlug[slug]; ok {
		return p, nil
	}
	return nil, errNotFound("page not found")
}
func (s *stubPageRepo) List(context.Context, string, *uint) ([]*model.Page, error) {
	return nil, nil
}
func (s *stubPageRepo) FindByThemeIDAndContentKey(_ context.Context, themeID, contentKey string) (*model.Page, error) {
	if s.byThemeContent == nil {
		return nil, errNotFound("page not found")
	}
	if p, ok := s.byThemeContent[s.key(themeID, contentKey)]; ok {
		return p, nil
	}
	return nil, errNotFound("page not found")
}
func (s *stubPageRepo) ListByThemeID(context.Context, string, string) ([]*model.Page, error) {
	return nil, nil
}
func (s *stubPageRepo) ListPublishedByThemeID(context.Context, string) ([]*model.Page, error) {
	return nil, nil
}
func (s *stubPageRepo) ListPublished(context.Context) ([]*model.Page, error) { return nil, nil }
func (s *stubPageRepo) UpdateSortOrder(context.Context, uint, int) error     { return nil }

type errNotFound string

func (e errNotFound) Error() string { return string(e) }

func TestSeedThemePages_ReassignsSlugWhenThemeSwitches(t *testing.T) {
	repo := &stubPageRepo{
		bySlug: map[string]*model.Page{
			"home": {
				Slug:       "home",
				ThemeID:    "corporate-classic",
				ContentKey: "home",
				RenderMode: "hardcoded",
				Status:     model.PageStatusPublished,
			},
		},
	}
	svc := NewThemePageService(repo)

	if err := svc.SeedThemePages(context.Background(), "blog-first"); err != nil {
		t.Fatalf("SeedThemePages: %v", err)
	}
	// Shared slug "home" is reassigned; theme-only pages (e.g. author) are created.
	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update (home reassign), got %d", len(repo.updated))
	}
	if repo.updated[0].ThemeID != "blog-first" {
		t.Fatalf("expected theme blog-first, got %s", repo.updated[0].ThemeID)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 create (author), got %d", len(repo.created))
	}
	if repo.created[0].Slug != "author" || repo.created[0].ThemeID != "blog-first" {
		t.Fatalf("expected created author page for blog-first, got slug=%s theme=%s",
			repo.created[0].Slug, repo.created[0].ThemeID)
	}
}
