package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/yixian-huang/inkless/backend/internal/builtinthemes"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
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

type builtinPageJSON struct {
	Slug       string        `json:"slug"`
	ContentKey string        `json:"contentKey"`
	RenderMode string        `json:"renderMode"`
	SortOrder  int           `json:"sortOrder"`
	Title      model.JSONMap `json:"title"`
	NavConfig  model.JSONMap `json:"navConfig"`
}

// BuiltInThemePages maps theme IDs to their page definitions (loaded from embedded JSON).
var BuiltInThemePages map[string][]ThemePageSeedDef

func init() {
	BuiltInThemePages = loadBuiltInThemePages()
}

func loadBuiltInThemePages() map[string][]ThemePageSeedDef {
	out := make(map[string][]ThemePageSeedDef)
	var raw map[string][]builtinPageJSON
	if err := json.Unmarshal(builtinthemes.PagesJSON, &raw); err != nil {
		log.Printf("Warning: failed to parse builtin_theme_pages.json: %v", err)
		return out
	}
	for themeID, pages := range raw {
		defs := make([]ThemePageSeedDef, 0, len(pages))
		for _, p := range pages {
			defs = append(defs, ThemePageSeedDef{
				Slug:       p.Slug,
				ContentKey: p.ContentKey,
				RenderMode: p.RenderMode,
				Title:      p.Title,
				SortOrder:  p.SortOrder,
				NavConfig:  p.NavConfig,
			})
		}
		out[themeID] = defs
	}
	return out
}

// ThemePageService handles seeding theme pages into the Page table
type ThemePageService struct {
	pageRepo repository.PageRepository
}

// NewThemePageService creates a new ThemePageService
func NewThemePageService(pageRepo repository.PageRepository) *ThemePageService {
	return &ThemePageService{pageRepo: pageRepo}
}

func pageFromDef(themeID string, def ThemePageSeedDef) *model.Page {
	return &model.Page{
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
}

func applyDefToPage(page *model.Page, themeID string, def ThemePageSeedDef) {
	page.ThemeID = themeID
	page.ContentKey = def.ContentKey
	page.RenderMode = def.RenderMode
	page.IsThemePage = true
	page.Title = def.Title
	page.SortOrder = def.SortOrder
	page.NavConfig = def.NavConfig
	page.Status = model.PageStatusPublished
}

// SeedThemePages creates or updates page records for a theme.
// Slugs are globally unique; activating a new theme reassigns shared slugs (e.g. home).
func (s *ThemePageService) SeedThemePages(ctx context.Context, themeID string) error {
	defs, ok := BuiltInThemePages[themeID]
	if !ok {
		log.Printf("No built-in page definitions for theme %s, skipping seed", themeID)
		return nil
	}

	for _, def := range defs {
		existing, err := s.pageRepo.FindByThemeIDAndContentKey(ctx, themeID, def.ContentKey)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
		if existing != nil {
			log.Printf("Theme page %s/%s already exists, skipping", themeID, def.ContentKey)
			continue
		}

		bySlug, slugErr := s.pageRepo.FindBySlug(ctx, def.Slug)
		if slugErr != nil && !strings.Contains(slugErr.Error(), "not found") {
			return slugErr
		}
		if bySlug != nil {
			applyDefToPage(bySlug, themeID, def)
			if err := s.pageRepo.Update(ctx, bySlug); err != nil {
				return err
			}
			log.Printf("Reassigned theme page slug=%s to theme %s (contentKey=%s)", def.Slug, themeID, def.ContentKey)
			continue
		}

		page := pageFromDef(themeID, def)
		if err := s.pageRepo.Create(ctx, page); err != nil {
			return err
		}
		log.Printf("Created theme page: %s/%s (slug=%s)", themeID, def.ContentKey, def.Slug)
	}

	return nil
}
