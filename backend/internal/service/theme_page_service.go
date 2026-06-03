package service

import (
	"context"
	"log"
	"strings"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// ThemePageSeedDef defines a single page to seed for a theme
type ThemePageSeedDef struct {
	Slug       string
	ContentKey string
	RenderMode string
	Title      model.JSONMap
	SortOrder  int
	NavConfig  model.JSONMap
}

// BuiltInThemePages maps theme IDs to their page definitions
var BuiltInThemePages = map[string][]ThemePageSeedDef{
	"corporate-classic": {
		{
			Slug: "home", ContentKey: "home", RenderMode: "hardcoded", SortOrder: 0,
			Title:     model.JSONMap{"zh": "首页", "en": "Home"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "about", ContentKey: "about", RenderMode: "hardcoded", SortOrder: 1,
			Title:     model.JSONMap{"zh": "关于我们", "en": "About"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "advantages", ContentKey: "advantages", RenderMode: "hardcoded", SortOrder: 2,
			Title:     model.JSONMap{"zh": "我们的优势", "en": "Advantages"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "core-services", ContentKey: "core-services", RenderMode: "hardcoded", SortOrder: 3,
			Title:     model.JSONMap{"zh": "核心服务", "en": "Services"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "cases", ContentKey: "cases", RenderMode: "hardcoded", SortOrder: 4,
			Title:     model.JSONMap{"zh": "案例展示", "en": "Cases"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "experts", ContentKey: "experts", RenderMode: "hardcoded", SortOrder: 5,
			Title:     model.JSONMap{"zh": "专家团队", "en": "Experts"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
		{
			Slug: "contact", ContentKey: "contact", RenderMode: "hardcoded", SortOrder: 6,
			Title:     model.JSONMap{"zh": "联系我们", "en": "Contact"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": true},
		},
	},
	"blog-first": {
		{
			Slug: "home", ContentKey: "home", RenderMode: "hardcoded", SortOrder: 0,
			Title:     model.JSONMap{"zh": "首页", "en": "Home"},
			NavConfig: model.JSONMap{"showInHeader": true, "showInFooter": false},
		},
	},
}

// ThemePageService handles seeding theme pages into the Page table
type ThemePageService struct {
	pageRepo repository.PageRepository
}

// NewThemePageService creates a new ThemePageService
func NewThemePageService(pageRepo repository.PageRepository) *ThemePageService {
	return &ThemePageService{pageRepo: pageRepo}
}

// SeedThemePages creates page records for a theme, skipping already-existing ones (by contentKey)
func (s *ThemePageService) SeedThemePages(ctx context.Context, themeID string) error {
	defs, ok := BuiltInThemePages[themeID]
	if !ok {
		log.Printf("No built-in page definitions for theme %s, skipping seed", themeID)
		return nil
	}

	for _, def := range defs {
		// Check if already exists (dedup by themeID + contentKey)
		existing, err := s.pageRepo.FindByThemeIDAndContentKey(ctx, themeID, def.ContentKey)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
		if existing != nil {
			log.Printf("Theme page %s/%s already exists, skipping", themeID, def.ContentKey)
			continue
		}

		page := &model.Page{
			Slug:        def.Slug,
			ThemeID:     themeID,
			ContentKey:  def.ContentKey,
			RenderMode:  def.RenderMode,
			IsThemePage: true,
			Title:       def.Title,
			SortOrder:   def.SortOrder,
			NavConfig:   def.NavConfig,
			Status:      model.PageStatusPublished,
		}

		if err := s.pageRepo.Create(ctx, page); err != nil {
			return err
		}
		log.Printf("Created theme page: %s/%s (slug=%s)", themeID, def.ContentKey, def.Slug)
	}

	return nil
}
