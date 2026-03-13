package service_test

import (
	"context"
	"testing"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&model.UnifiedPage{}, &model.PageVersion{}, &model.PageTemplate{}, &model.SiteConfig{})
	return db
}

func TestUnifiedPageService_Publish(t *testing.T) {
	db := setupServiceTestDB(t)
	pageRepo := repository.NewGormUnifiedPageRepository(db)
	versionRepo := repository.NewGormPageVersionRepository(db)
	svc := service.NewUnifiedPageService(pageRepo, versionRepo)
	ctx := context.Background()

	page := &model.UnifiedPage{
		Slug: "test", Mode: "composable", DraftVersion: 1,
		DraftConfig: model.JSONMap{"sections": []any{}},
	}
	pageRepo.Create(ctx, page)

	err := svc.Publish(ctx, page.ID, 1, 1) // expectedDraftVersion=1, userID=1
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	updated, _ := pageRepo.FindByID(ctx, page.ID)
	if updated.Status != "published" {
		t.Errorf("expected published, got %q", updated.Status)
	}
	if updated.PublishedVersion != 1 {
		t.Errorf("expected publishedVersion 1, got %d", updated.PublishedVersion)
	}

	// Verify version record created
	versions, count, _ := versionRepo.ListByPageID(ctx, page.ID, 0, 10)
	if count != 1 {
		t.Errorf("expected 1 version, got %d", count)
	}
	if versions[0].CreatedBy != 1 {
		t.Errorf("expected createdBy 1, got %d", versions[0].CreatedBy)
	}
}

func TestUnifiedPageService_Publish_VersionConflict(t *testing.T) {
	db := setupServiceTestDB(t)
	pageRepo := repository.NewGormUnifiedPageRepository(db)
	versionRepo := repository.NewGormPageVersionRepository(db)
	svc := service.NewUnifiedPageService(pageRepo, versionRepo)
	ctx := context.Background()

	page := &model.UnifiedPage{
		Slug: "test", Mode: "composable", DraftVersion: 2,
		DraftConfig: model.JSONMap{"sections": []any{}},
	}
	pageRepo.Create(ctx, page)

	err := svc.Publish(ctx, page.ID, 1, 1) // wrong expectedDraftVersion
	if err == nil {
		t.Error("expected version conflict error")
	}
}

func TestUnifiedPageService_Rollback(t *testing.T) {
	db := setupServiceTestDB(t)
	pageRepo := repository.NewGormUnifiedPageRepository(db)
	versionRepo := repository.NewGormPageVersionRepository(db)
	svc := service.NewUnifiedPageService(pageRepo, versionRepo)
	ctx := context.Background()

	page := &model.UnifiedPage{
		Slug: "test", Mode: "composable", DraftVersion: 1,
		DraftConfig: model.JSONMap{"sections": []any{map[string]any{"type": "hero"}}},
	}
	pageRepo.Create(ctx, page)

	// Publish v1 (draftVersion=1)
	if err := svc.Publish(ctx, page.ID, 1, 1); err != nil {
		t.Fatalf("publish v1: %v", err)
	}

	// Modify draft (current draftVersion=1, UpdateDraft increments to 2)
	newDraftVer, err := pageRepo.UpdateDraft(ctx, page.ID, 1, model.JSONMap{"sections": []any{map[string]any{"type": "hero"}, map[string]any{"type": "rich-text"}}})
	if err != nil {
		t.Fatalf("update draft: %v", err)
	}

	// Publish v2 (draftVersion=newDraftVer)
	if err := svc.Publish(ctx, page.ID, newDraftVer, 1); err != nil {
		t.Fatalf("publish v2: %v", err)
	}

	// Rollback to v1 → creates v3
	err = svc.Rollback(ctx, page.ID, 1, 1)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}

	updated, _ := pageRepo.FindByID(ctx, page.ID)
	if updated.PublishedVersion != 3 {
		t.Errorf("expected publishedVersion 3 after rollback, got %d", updated.PublishedVersion)
	}
}

func TestUnifiedPageService_Unpublish(t *testing.T) {
	db := setupServiceTestDB(t)
	pageRepo := repository.NewGormUnifiedPageRepository(db)
	versionRepo := repository.NewGormPageVersionRepository(db)
	svc := service.NewUnifiedPageService(pageRepo, versionRepo)
	ctx := context.Background()

	page := &model.UnifiedPage{
		Slug: "test", Mode: "composable", DraftVersion: 1,
		DraftConfig: model.JSONMap{"sections": []any{}},
	}
	pageRepo.Create(ctx, page)

	// Publish first
	if err := svc.Publish(ctx, page.ID, 1, 1); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Unpublish
	if err := svc.Unpublish(ctx, page.ID); err != nil {
		t.Fatalf("unpublish: %v", err)
	}

	updated, _ := pageRepo.FindByID(ctx, page.ID)
	if updated.Status != "draft" {
		t.Errorf("expected draft, got %q", updated.Status)
	}
	if updated.PublishedConfig != nil {
		t.Error("expected PublishedConfig to be nil after unpublish")
	}
}
