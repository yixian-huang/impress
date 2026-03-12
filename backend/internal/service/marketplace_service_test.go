package service

import (
	"context"
	"errors"
	"testing"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// mockMarketplaceRepo is a simple in-memory implementation for tests
type mockMarketplaceRepo struct {
	items    map[string]*model.MarketplaceItem
	versions []*model.MarketplaceVersion
	nextID   uint
}

func newMockMarketplaceRepo() *mockMarketplaceRepo {
	return &mockMarketplaceRepo{
		items:  make(map[string]*model.MarketplaceItem),
		nextID: 1,
	}
}

func (m *mockMarketplaceRepo) List(_ context.Context, filter repository.MarketplaceFilter) ([]*model.MarketplaceItem, int64, error) {
	var result []*model.MarketplaceItem
	for _, item := range m.items {
		if filter.Type != "" && string(item.Type) != filter.Type {
			continue
		}
		if filter.Status != "" && string(item.Status) != filter.Status {
			continue
		}
		result = append(result, item)
	}
	return result, int64(len(result)), nil
}

func (m *mockMarketplaceRepo) GetBySlug(_ context.Context, slug string) (*model.MarketplaceItem, error) {
	item, ok := m.items[slug]
	if !ok {
		return nil, errors.New("marketplace item not found")
	}
	return item, nil
}

func (m *mockMarketplaceRepo) GetByID(_ context.Context, id uint) (*model.MarketplaceItem, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, errors.New("marketplace item not found")
}

func (m *mockMarketplaceRepo) Create(_ context.Context, item *model.MarketplaceItem) error {
	if item.Slug == "" {
		return errors.New("slug required")
	}
	if _, exists := m.items[item.Slug]; exists {
		return errors.New("duplicate slug")
	}
	item.ID = m.nextID
	m.nextID++
	// Store a copy
	cp := *item
	m.items[item.Slug] = &cp
	return nil
}

func (m *mockMarketplaceRepo) Update(_ context.Context, item *model.MarketplaceItem) error {
	if _, ok := m.items[item.Slug]; !ok {
		return errors.New("not found")
	}
	cp := *item
	m.items[item.Slug] = &cp
	return nil
}

func (m *mockMarketplaceRepo) Delete(_ context.Context, id uint) error {
	for slug, item := range m.items {
		if item.ID == id {
			delete(m.items, slug)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockMarketplaceRepo) IncrementDownloads(_ context.Context, id uint) error {
	for _, item := range m.items {
		if item.ID == id {
			item.Downloads++
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockMarketplaceRepo) CreateVersion(_ context.Context, v *model.MarketplaceVersion) error {
	m.versions = append(m.versions, v)
	return nil
}

func (m *mockMarketplaceRepo) ListVersions(_ context.Context, itemID uint) ([]*model.MarketplaceVersion, error) {
	var result []*model.MarketplaceVersion
	for _, v := range m.versions {
		if v.ItemID == itemID {
			result = append(result, v)
		}
	}
	return result, nil
}

// --- Tests ---

func newTestMarketplaceService() (*MarketplaceService, *mockMarketplaceRepo) {
	repo := newMockMarketplaceRepo()
	svc := NewMarketplaceService(repo)
	return svc, repo
}

func seedItem(repo *mockMarketplaceRepo, slug string, status model.MarketplaceItemStatus) *model.MarketplaceItem {
	item := &model.MarketplaceItem{
		Type:        model.MarketplaceItemTypePlugin,
		Name:        "Plugin " + slug,
		Slug:        slug,
		Version:     "1.0.0",
		Status:      status,
		DownloadURL: "https://example.com/" + slug + ".zip",
	}
	_ = repo.Create(context.Background(), item)
	return repo.items[slug]
}

func TestMarketplaceService_SearchItems(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "plugin-a", model.MarketplaceItemStatusActive)
	seedItem(repo, "plugin-b", model.MarketplaceItemStatusActive)

	items, total, err := svc.SearchItems(ctx, repository.MarketplaceFilter{})
	if err != nil {
		t.Fatalf("SearchItems failed: %v", err)
	}
	if total != 2 {
		t.Errorf("Expected 2 items, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items in result, got %d", len(items))
	}
}

func TestMarketplaceService_GetItemDetails(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "detail-plugin", model.MarketplaceItemStatusActive)

	item, err := svc.GetItemDetails(ctx, "detail-plugin")
	if err != nil {
		t.Fatalf("GetItemDetails failed: %v", err)
	}
	if item.Slug != "detail-plugin" {
		t.Errorf("Expected slug %q, got %q", "detail-plugin", item.Slug)
	}
}

func TestMarketplaceService_GetItemDetails_EmptySlug(t *testing.T) {
	svc, _ := newTestMarketplaceService()

	_, err := svc.GetItemDetails(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty slug")
	}
}

func TestMarketplaceService_InstallItem(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seeded := seedItem(repo, "install-me", model.MarketplaceItemStatusActive)

	item, err := svc.InstallItem(ctx, "install-me")
	if err != nil {
		t.Fatalf("InstallItem failed: %v", err)
	}
	if item.Downloads != 1 {
		t.Errorf("Expected downloads=1 after install, got %d", item.Downloads)
	}
	_ = seeded
}

func TestMarketplaceService_InstallItem_Deprecated(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "old-plugin", model.MarketplaceItemStatusDeprecated)

	_, err := svc.InstallItem(ctx, "old-plugin")
	if err == nil {
		t.Error("Expected error when installing deprecated item")
	}
}

func TestMarketplaceService_InstallItem_NotFound(t *testing.T) {
	svc, _ := newTestMarketplaceService()

	_, err := svc.InstallItem(context.Background(), "doesnt-exist")
	if err == nil {
		t.Error("Expected error for nonexistent item")
	}
}

func TestMarketplaceService_UpdateItem(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "update-me", model.MarketplaceItemStatusActive)

	item, err := svc.UpdateItem(ctx, "update-me")
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}
	if item.Downloads != 1 {
		t.Errorf("Expected downloads=1 after update, got %d", item.Downloads)
	}
}

func TestMarketplaceService_RegisterItem(t *testing.T) {
	svc, _ := newTestMarketplaceService()
	ctx := context.Background()

	item := &model.MarketplaceItem{
		Type:    model.MarketplaceItemTypeTheme,
		Name:    "My Theme",
		Slug:    "my-theme",
		Version: "1.0.0",
	}

	if err := svc.RegisterItem(ctx, item); err != nil {
		t.Fatalf("RegisterItem failed: %v", err)
	}

	// Verify default status applied
	if item.Status != model.MarketplaceItemStatusActive {
		t.Errorf("Expected status active, got %s", item.Status)
	}
}

func TestMarketplaceService_RegisterItem_Validation(t *testing.T) {
	svc, _ := newTestMarketplaceService()
	ctx := context.Background()

	tests := []struct {
		name    string
		item    *model.MarketplaceItem
		wantErr bool
	}{
		{
			name:    "missing slug",
			item:    &model.MarketplaceItem{Type: model.MarketplaceItemTypePlugin, Name: "X", Version: "1.0.0"},
			wantErr: true,
		},
		{
			name:    "missing name",
			item:    &model.MarketplaceItem{Type: model.MarketplaceItemTypePlugin, Slug: "x", Version: "1.0.0"},
			wantErr: true,
		},
		{
			name:    "missing type",
			item:    &model.MarketplaceItem{Slug: "x", Name: "X", Version: "1.0.0"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			item:    &model.MarketplaceItem{Type: "widget", Slug: "x", Name: "X", Version: "1.0.0"},
			wantErr: true,
		},
		{
			name:    "missing version",
			item:    &model.MarketplaceItem{Type: model.MarketplaceItemTypePlugin, Slug: "x", Name: "X"},
			wantErr: true,
		},
		{
			name:    "valid plugin",
			item:    &model.MarketplaceItem{Type: model.MarketplaceItemTypePlugin, Slug: "valid-plugin", Name: "X", Version: "1.0.0"},
			wantErr: false,
		},
		{
			name:    "valid theme",
			item:    &model.MarketplaceItem{Type: model.MarketplaceItemTypeTheme, Slug: "valid-theme", Name: "Y", Version: "1.0.0"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RegisterItem(ctx, tc.item)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMarketplaceService_UninstallItem(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "uninstall-me", model.MarketplaceItemStatusActive)

	if err := svc.UninstallItem(ctx, "uninstall-me"); err != nil {
		t.Fatalf("UninstallItem failed: %v", err)
	}

	_, err := svc.GetItemDetails(ctx, "uninstall-me")
	if err == nil {
		t.Error("Expected error after uninstall")
	}
}

func TestMarketplaceService_AddVersion(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seeded := seedItem(repo, "versioned-plugin", model.MarketplaceItemStatusActive)

	version := &model.MarketplaceVersion{
		Version:   "2.0.0",
		Changelog: "Major update",
	}

	if err := svc.AddVersion(ctx, "versioned-plugin", version); err != nil {
		t.Fatalf("AddVersion failed: %v", err)
	}

	if version.ItemID != seeded.ID {
		t.Errorf("Expected ItemID=%d, got %d", seeded.ID, version.ItemID)
	}
}

func TestMarketplaceService_UpdateRegistryItem(t *testing.T) {
	svc, repo := newTestMarketplaceService()
	ctx := context.Background()

	seedItem(repo, "update-registry", model.MarketplaceItemStatusActive)

	updates := &model.MarketplaceItem{
		Name:    "Updated Name",
		Version: "2.0.0",
	}

	updated, err := svc.UpdateRegistryItem(ctx, "update-registry", updates)
	if err != nil {
		t.Fatalf("UpdateRegistryItem failed: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Expected name %q, got %q", "Updated Name", updated.Name)
	}
	if updated.Version != "2.0.0" {
		t.Errorf("Expected version %q, got %q", "2.0.0", updated.Version)
	}
}
