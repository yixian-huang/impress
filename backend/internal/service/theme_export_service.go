package service

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

type ThemeExportService struct {
	templateRepo   repository.PageTemplateRepository
	siteConfigRepo repository.SiteConfigRepository
}

func NewThemeExportService(templateRepo repository.PageTemplateRepository, siteConfigRepo repository.SiteConfigRepository) *ThemeExportService {
	return &ThemeExportService{templateRepo: templateRepo, siteConfigRepo: siteConfigRepo}
}

// Export assembles a theme JSON package with tokens and page templates
func (s *ThemeExportService) Export(ctx context.Context, name string) (model.JSONMap, error) {
	// 1. Fetch theme SiteConfig for tokens
	themeConfig, err := s.siteConfigRepo.FindByKey(ctx, "theme")
	var tokens model.JSONMap
	if err == nil && themeConfig.DraftConfig != nil {
		tokens = themeConfig.DraftConfig
	}

	// 2. Fetch all page templates
	templates, err := s.templateRepo.List(ctx, "")
	if err != nil {
		return nil, err
	}

	// Build template array for export
	tmplList := make([]map[string]interface{}, 0, len(templates))
	for _, t := range templates {
		tmplList = append(tmplList, map[string]interface{}{
			"key":           t.Key,
			"nameZh":        t.NameZh,
			"nameEn":        t.NameEn,
			"descriptionZh": t.DescriptionZh,
			"descriptionEn": t.DescriptionEn,
			"config":        t.Config,
			"thumbnail":     t.Thumbnail,
		})
	}

	return model.JSONMap{
		"name":          name,
		"version":       "1.0",
		"tokens":        tokens,
		"pageTemplates": tmplList,
	}, nil
}

// Import parses a theme package and creates templates + applies tokens
func (s *ThemeExportService) Import(ctx context.Context, themePackage model.JSONMap) error {
	// Apply tokens to theme SiteConfig
	if tokens, ok := themePackage["tokens"].(map[string]interface{}); ok && len(tokens) > 0 {
		existing, err := s.siteConfigRepo.FindByKey(ctx, "theme")
		if err != nil {
			// Create new
			sc := &model.SiteConfig{
				Key:          "theme",
				DraftConfig:  model.JSONMap(tokens),
				DraftVersion: 1,
			}
			if upsertErr := s.siteConfigRepo.Upsert(ctx, sc); upsertErr != nil {
				return upsertErr
			}
		} else {
			existing.DraftConfig = model.JSONMap(tokens)
			existing.DraftVersion++
			if err := s.siteConfigRepo.Update(ctx, existing); err != nil {
				return err
			}
		}
	}

	// Create page templates from package
	if tmplsRaw, ok := themePackage["pageTemplates"].([]interface{}); ok {
		for _, raw := range tmplsRaw {
			tmplMap, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			key, _ := tmplMap["key"].(string)
			nameZh, _ := tmplMap["nameZh"].(string)
			nameEn, _ := tmplMap["nameEn"].(string)
			descZh, _ := tmplMap["descriptionZh"].(string)
			descEn, _ := tmplMap["descriptionEn"].(string)
			thumbnail, _ := tmplMap["thumbnail"].(string)
			config := model.JSONMap{}
			if c, ok := tmplMap["config"].(map[string]interface{}); ok {
				config = model.JSONMap(c)
			}
			tmpl := &model.PageTemplate{
				Key:           key,
				NameZh:        nameZh,
				NameEn:        nameEn,
				DescriptionZh: descZh,
				DescriptionEn: descEn,
				Category:      model.TemplateCategoryTheme,
				Config:        config,
				Thumbnail:     thumbnail,
			}
			s.templateRepo.Create(ctx, tmpl) // ignore dups
		}
	}

	return nil
}

// ListInstalledThemes returns theme-category templates grouped as themes
func (s *ThemeExportService) ListInstalledThemes(ctx context.Context) ([]model.JSONMap, error) {
	templates, err := s.templateRepo.List(ctx, "theme")
	if err != nil {
		return nil, err
	}
	var themes []model.JSONMap
	for _, t := range templates {
		themes = append(themes, model.JSONMap{
			"id":        t.ID,
			"key":       t.Key,
			"nameZh":    t.NameZh,
			"nameEn":    t.NameEn,
			"thumbnail": t.Thumbnail,
		})
	}
	return themes, nil
}

// ApplyTheme extracts tokens from a theme template and applies to SiteConfig
func (s *ThemeExportService) ApplyTheme(ctx context.Context, themeTemplateID uint) error {
	tmpl, err := s.templateRepo.FindByID(ctx, themeTemplateID)
	if err != nil {
		return err
	}
	tokens, ok := tmpl.Config["tokens"].(map[string]interface{})
	if !ok || tokens == nil {
		return nil
	}
	themeConfig, err := s.siteConfigRepo.FindByKey(ctx, "theme")
	if err != nil {
		// Create new
		sc := &model.SiteConfig{
			Key:          "theme",
			DraftConfig:  model.JSONMap(tokens),
			DraftVersion: 1,
		}
		return s.siteConfigRepo.Upsert(ctx, sc)
	}
	themeConfig.DraftConfig = model.JSONMap(tokens)
	themeConfig.DraftVersion++
	return s.siteConfigRepo.Update(ctx, themeConfig)
}
