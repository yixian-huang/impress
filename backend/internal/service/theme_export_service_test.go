package service_test

import (
	"context"
	"testing"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/service"
)

func TestThemeExportService_Export(t *testing.T) {
	db := setupServiceTestDB(t)
	templateRepo := repository.NewGormPageTemplateRepository(db)
	siteConfigRepo := repository.NewGormSiteConfigRepository(db)
	svc := service.NewThemeExportService(templateRepo, siteConfigRepo)
	ctx := context.Background()

	// Seed a theme config
	siteConfigRepo.Upsert(ctx, &model.SiteConfig{
		Key:          "theme",
		DraftConfig:  model.JSONMap{"colors": map[string]interface{}{"primary": "#111"}},
		DraftVersion: 1,
	})

	// Seed a template
	templateRepo.Create(ctx, &model.PageTemplate{
		Key:    "hero-tpl",
		NameZh: "英雄模板",
		NameEn: "Hero Template",
		Config: model.JSONMap{"sections": []interface{}{}},
	})

	result, err := svc.Export(ctx, "test-theme")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if result["name"] != "test-theme" {
		t.Errorf("expected name=test-theme, got %v", result["name"])
	}
	if result["version"] != "1.0" {
		t.Errorf("expected version=1.0, got %v", result["version"])
	}
	if result["tokens"] == nil {
		t.Error("expected tokens to be non-nil")
	}
	tmpls, ok := result["pageTemplates"].([]map[string]interface{})
	if !ok || len(tmpls) == 0 {
		t.Errorf("expected at least one pageTemplate, got %v", result["pageTemplates"])
	}
	if tmpls[0]["key"] != "hero-tpl" {
		t.Errorf("expected key=hero-tpl, got %v", tmpls[0]["key"])
	}
}

func TestThemeExportService_Import(t *testing.T) {
	db := setupServiceTestDB(t)
	templateRepo := repository.NewGormPageTemplateRepository(db)
	siteConfigRepo := repository.NewGormSiteConfigRepository(db)
	svc := service.NewThemeExportService(templateRepo, siteConfigRepo)
	ctx := context.Background()

	pkg := model.JSONMap{
		"name":    "imported-theme",
		"version": "1.0",
		"tokens":  map[string]interface{}{"colors": map[string]interface{}{"primary": "#222"}},
		"pageTemplates": []interface{}{
			map[string]interface{}{
				"key":    "imported-tpl",
				"nameZh": "导入模板",
				"nameEn": "Imported Template",
				"config": map[string]interface{}{"sections": []interface{}{}},
			},
		},
	}

	err := svc.Import(ctx, pkg)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	// Check tokens were saved
	sc, err := siteConfigRepo.FindByKey(ctx, "theme")
	if err != nil {
		t.Fatalf("find theme config: %v", err)
	}
	if sc.DraftConfig == nil {
		t.Error("expected draft config to be set")
	}

	// Check template was created with theme category
	tmpls, err := templateRepo.List(ctx, "theme")
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(tmpls) != 1 {
		t.Fatalf("expected 1 theme template, got %d", len(tmpls))
	}
	if tmpls[0].Key != "imported-tpl" {
		t.Errorf("expected key=imported-tpl, got %s", tmpls[0].Key)
	}
	if tmpls[0].Category != model.TemplateCategoryTheme {
		t.Errorf("expected category=theme, got %s", tmpls[0].Category)
	}
}
