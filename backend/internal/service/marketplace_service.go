package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// MarketplaceService provides marketplace business logic
type MarketplaceService struct {
	repo repository.MarketplaceRepository
}

// NewMarketplaceService creates a new MarketplaceService
func NewMarketplaceService(repo repository.MarketplaceRepository) *MarketplaceService {
	return &MarketplaceService{repo: repo}
}

// SearchItems searches marketplace items using the provided filter
func (s *MarketplaceService) SearchItems(ctx context.Context, filter repository.MarketplaceFilter) ([]*model.MarketplaceItem, int64, error) {
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	return s.repo.List(ctx, filter)
}

// GetItemDetails returns details for a marketplace item by slug, including all versions
func (s *MarketplaceService) GetItemDetails(ctx context.Context, slug string) (*model.MarketplaceItem, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.GetBySlug(ctx, slug)
}

// InstallItem records the installation of a marketplace item and increments the download counter.
// It returns the download URL so the caller can perform the actual download.
func (s *MarketplaceService) InstallItem(ctx context.Context, slug string) (*model.MarketplaceItem, error) {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	if item.Status != model.MarketplaceItemStatusActive {
		return nil, fmt.Errorf("item %q is not available for installation (status: %s)", slug, item.Status)
	}

	if err := s.repo.IncrementDownloads(ctx, item.ID); err != nil {
		return nil, fmt.Errorf("failed to record download: %w", err)
	}

	// Refresh to get updated download count
	item, err = s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateItem records a marketplace item update (increments downloads and returns updated item)
func (s *MarketplaceService) UpdateItem(ctx context.Context, slug string) (*model.MarketplaceItem, error) {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	if item.Status != model.MarketplaceItemStatusActive {
		return nil, fmt.Errorf("item %q is not available for updates (status: %s)", slug, item.Status)
	}

	if err := s.repo.IncrementDownloads(ctx, item.ID); err != nil {
		return nil, fmt.Errorf("failed to record update download: %w", err)
	}

	// Refresh to get updated download count
	item, err = s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// ListInstalled returns all marketplace items (exposed for admin installed-item management)
func (s *MarketplaceService) ListInstalled(ctx context.Context) ([]*model.MarketplaceItem, error) {
	items, _, err := s.repo.List(ctx, repository.MarketplaceFilter{
		Page:     1,
		PageSize: 100,
		Status:   string(model.MarketplaceItemStatusActive),
	})
	return items, err
}

// RegisterItem creates a new marketplace item (for use by marketplace registry)
func (s *MarketplaceService) RegisterItem(ctx context.Context, item *model.MarketplaceItem) error {
	if item.Slug == "" {
		return errors.New("slug is required")
	}
	if item.Name == "" {
		return errors.New("name is required")
	}
	if item.Type == "" {
		return errors.New("type is required")
	}
	if item.Type != model.MarketplaceItemTypePlugin && item.Type != model.MarketplaceItemTypeTheme {
		return fmt.Errorf("type must be %q or %q", model.MarketplaceItemTypePlugin, model.MarketplaceItemTypeTheme)
	}
	if item.Version == "" {
		return errors.New("version is required")
	}
	if item.Status == "" {
		item.Status = model.MarketplaceItemStatusActive
	}
	return s.repo.Create(ctx, item)
}

// UpdateRegistryItem updates an existing marketplace item's metadata
func (s *MarketplaceService) UpdateRegistryItem(ctx context.Context, slug string, updates *model.MarketplaceItem) (*model.MarketplaceItem, error) {
	existing, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.NameZh != "" {
		existing.NameZh = updates.NameZh
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Version != "" {
		existing.Version = updates.Version
	}
	if updates.IconURL != "" {
		existing.IconURL = updates.IconURL
	}
	if updates.PreviewURL != "" {
		existing.PreviewURL = updates.PreviewURL
	}
	if updates.DownloadURL != "" {
		existing.DownloadURL = updates.DownloadURL
	}
	if updates.Category != "" {
		existing.Category = updates.Category
	}
	if len(updates.Tags) > 0 {
		existing.Tags = updates.Tags
	}
	if updates.Status != "" {
		existing.Status = updates.Status
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// AddVersion adds a new version record to an existing marketplace item
func (s *MarketplaceService) AddVersion(ctx context.Context, slug string, version *model.MarketplaceVersion) error {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	version.ItemID = item.ID
	return s.repo.CreateVersion(ctx, version)
}

// UninstallItem soft-deletes a marketplace item by slug
func (s *MarketplaceService) UninstallItem(ctx context.Context, slug string) error {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}
	return s.repo.Delete(ctx, item.ID)
}

// FetchManifest is a helper for downloading a remote manifest/package.
// It returns the raw bytes of the download or an error if the request fails.
// A short timeout is applied to avoid blocking indefinitely.
func FetchManifest(downloadURL string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(downloadURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return data, nil
}
