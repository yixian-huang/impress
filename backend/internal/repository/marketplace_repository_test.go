package repository

import (
	"context"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func setupMarketplaceTestDB(t *testing.T) *GormMarketplaceRepository {
	t.Helper()

	database := setupTestDB(t)
	t.Cleanup(func() { database.Close() })

	// Migrate marketplace tables
	if err := database.DB.AutoMigrate(&model.MarketplaceItem{}, &model.MarketplaceVersion{}); err != nil {
		t.Fatalf("Failed to migrate marketplace tables: %v", err)
	}

	return &GormMarketplaceRepository{db: database.DB}
}

func sampleItem(slug string) *model.MarketplaceItem {
	return &model.MarketplaceItem{
		Type:        model.MarketplaceItemTypePlugin,
		Name:        "Test Plugin",
		NameZh:      "测试插件",
		Slug:        slug,
		Description: "A test plugin",
		Author:      "tester",
		Version:     "1.0.0",
		DownloadURL: "https://example.com/plugin.zip",
		Category:    "utility",
		Tags:        model.JSONStringSlice{"test", "demo"},
		Status:      model.MarketplaceItemStatusActive,
	}
}

func TestMarketplaceRepository_Create(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("test-plugin")
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if item.ID == 0 {
		t.Error("Expected item ID to be set after create")
	}
}

func TestMarketplaceRepository_GetBySlug(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("find-me")
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetBySlug(ctx, "find-me")
	if err != nil {
		t.Fatalf("GetBySlug failed: %v", err)
	}
	if found.Slug != "find-me" {
		t.Errorf("Expected slug %q, got %q", "find-me", found.Slug)
	}
	if found.Name != "Test Plugin" {
		t.Errorf("Expected name %q, got %q", "Test Plugin", found.Name)
	}
}

func TestMarketplaceRepository_GetBySlug_NotFound(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	_, err := repo.GetBySlug(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent slug")
	}
}

func TestMarketplaceRepository_List_Filter(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	plugin := sampleItem("a-plugin")
	plugin.Type = model.MarketplaceItemTypePlugin
	plugin.Category = "utility"
	_ = repo.Create(ctx, plugin)

	theme := sampleItem("a-theme")
	theme.Type = model.MarketplaceItemTypeTheme
	theme.Category = "ui"
	_ = repo.Create(ctx, theme)

	// Filter by type
	items, total, err := repo.List(ctx, MarketplaceFilter{Type: "plugin", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 plugin, got %d", total)
	}
	if items[0].Type != model.MarketplaceItemTypePlugin {
		t.Errorf("Expected type plugin, got %s", items[0].Type)
	}

	// Filter by category
	items, total, err = repo.List(ctx, MarketplaceFilter{Category: "ui", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List by category failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 ui item, got %d", total)
	}
	_ = items
}

func TestMarketplaceRepository_List_Search(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("searchable-plugin")
	item.Name = "Awesome Search Plugin"
	_ = repo.Create(ctx, item)

	other := sampleItem("other-plugin")
	other.Name = "Other Widget"
	_ = repo.Create(ctx, other)

	items, total, err := repo.List(ctx, MarketplaceFilter{Search: "awesome", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List with search failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 result for search 'awesome', got %d", total)
	}
	if items[0].Slug != "searchable-plugin" {
		t.Errorf("Expected slug %q, got %q", "searchable-plugin", items[0].Slug)
	}
}

func TestMarketplaceRepository_Update(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("update-me")
	_ = repo.Create(ctx, item)

	item.Name = "Updated Name"
	item.Version = "2.0.0"
	if err := repo.Update(ctx, item); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetBySlug(ctx, "update-me")
	if found.Name != "Updated Name" {
		t.Errorf("Expected name %q, got %q", "Updated Name", found.Name)
	}
	if found.Version != "2.0.0" {
		t.Errorf("Expected version %q, got %q", "2.0.0", found.Version)
	}
}

func TestMarketplaceRepository_Delete(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("delete-me")
	_ = repo.Create(ctx, item)

	if err := repo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetBySlug(ctx, "delete-me")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

func TestMarketplaceRepository_Delete_NotFound(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, 99999)
	if err == nil {
		t.Error("Expected error for nonexistent ID")
	}
}

func TestMarketplaceRepository_IncrementDownloads(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("popular-plugin")
	_ = repo.Create(ctx, item)

	if err := repo.IncrementDownloads(ctx, item.ID); err != nil {
		t.Fatalf("IncrementDownloads failed: %v", err)
	}
	if err := repo.IncrementDownloads(ctx, item.ID); err != nil {
		t.Fatalf("IncrementDownloads failed: %v", err)
	}

	found, _ := repo.GetBySlug(ctx, "popular-plugin")
	if found.Downloads != 2 {
		t.Errorf("Expected downloads=2, got %d", found.Downloads)
	}
}

func TestMarketplaceRepository_IncrementDownloads_NotFound(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	err := repo.IncrementDownloads(ctx, 99999)
	if err == nil {
		t.Error("Expected error for nonexistent item")
	}
}

func TestMarketplaceRepository_Versions(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	item := sampleItem("versioned-plugin")
	_ = repo.Create(ctx, item)

	v1 := &model.MarketplaceVersion{
		ItemID:      item.ID,
		Version:     "1.0.0",
		Changelog:   "Initial release",
		DownloadURL: "https://example.com/v1.zip",
	}
	v2 := &model.MarketplaceVersion{
		ItemID:      item.ID,
		Version:     "1.1.0",
		Changelog:   "Bug fixes",
		DownloadURL: "https://example.com/v2.zip",
	}

	if err := repo.CreateVersion(ctx, v1); err != nil {
		t.Fatalf("CreateVersion v1 failed: %v", err)
	}
	if err := repo.CreateVersion(ctx, v2); err != nil {
		t.Fatalf("CreateVersion v2 failed: %v", err)
	}

	versions, err := repo.ListVersions(ctx, item.ID)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	// Verify preload in GetBySlug
	found, _ := repo.GetBySlug(ctx, "versioned-plugin")
	if len(found.Versions) != 2 {
		t.Errorf("Expected 2 preloaded versions, got %d", len(found.Versions))
	}
}

func TestMarketplaceRepository_List_Pagination(t *testing.T) {
	repo := setupMarketplaceTestDB(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		item := sampleItem("paged-plugin-" + string(rune('a'+i)))
		_ = repo.Create(ctx, item)
	}

	// First page
	items, total, err := repo.List(ctx, MarketplaceFilter{Page: 1, PageSize: 3})
	if err != nil {
		t.Fatalf("List page 1 failed: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected total=5, got %d", total)
	}
	if len(items) != 3 {
		t.Errorf("Expected 3 items on page 1, got %d", len(items))
	}

	// Second page
	items, _, err = repo.List(ctx, MarketplaceFilter{Page: 2, PageSize: 3})
	if err != nil {
		t.Fatalf("List page 2 failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items on page 2, got %d", len(items))
	}
}
