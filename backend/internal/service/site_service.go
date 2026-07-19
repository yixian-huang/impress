package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"

	"gorm.io/gorm"
)

// SiteService handles business logic for multi-site management
type SiteService struct {
	db       *gorm.DB
	siteRepo repository.SiteRepository
}

// NewSiteService creates a new SiteService
func NewSiteService(db *gorm.DB, siteRepo repository.SiteRepository) *SiteService {
	return &SiteService{
		db:       db,
		siteRepo: siteRepo,
	}
}

// SiteScope returns a GORM scope function that filters records by site_id.
// Apply it with db.Scopes(SiteScope(siteID)) when building queries.
func SiteScope(siteID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("site_id = ?", siteID)
	}
}

// Create validates and persists a new site
func (s *SiteService) Create(ctx context.Context, site *model.Site) error {
	if site.Mode == "" {
		site.Mode = model.SiteModeSubdomain
	}
	if site.Status == "" {
		site.Status = model.SiteStatusActive
	}
	if site.Locale == "" {
		site.Locale = "zh"
	}
	if site.Settings == nil {
		site.Settings = model.SiteSettings{}
	}
	return s.siteRepo.Create(ctx, site)
}

// Update validates and persists site changes
func (s *SiteService) Update(ctx context.Context, site *model.Site) error {
	return s.siteRepo.Update(ctx, site)
}

// Delete removes a site
func (s *SiteService) Delete(ctx context.Context, id uint) error {
	return s.siteRepo.Delete(ctx, id)
}

// GetByID returns a site by ID
func (s *SiteService) GetByID(ctx context.Context, id uint) (*model.Site, error) {
	return s.siteRepo.FindByID(ctx, id)
}

// List returns all sites, optionally filtered by status
func (s *SiteService) List(ctx context.Context, status string) ([]*model.Site, error) {
	return s.siteRepo.List(ctx, status)
}

// AssignUser assigns a user to a site with the given role
func (s *SiteService) AssignUser(ctx context.Context, siteID, userID, roleID uint) error {
	if siteID == 0 || userID == 0 || roleID == 0 {
		return errors.New("siteID, userID, and roleID are required")
	}
	su := &model.SiteUser{
		SiteID: siteID,
		UserID: userID,
		RoleID: roleID,
	}
	return s.siteRepo.AddUser(ctx, su)
}

// UnassignUser removes a user from a site
func (s *SiteService) UnassignUser(ctx context.Context, siteID, userID uint) error {
	return s.siteRepo.RemoveUser(ctx, siteID, userID)
}

// ListUsers returns all user-role assignments for a site
func (s *SiteService) ListUsers(ctx context.Context, siteID uint) ([]*model.SiteUser, error) {
	return s.siteRepo.ListUsers(ctx, siteID)
}

// --- Export / Import ---

// SiteExportPayload is the JSON structure for a single-site data export
type SiteExportPayload struct {
	ExportedAt time.Time   `json:"exportedAt"`
	Site       *model.Site `json:"site"`
}

// ExportSite serialises site configuration to JSON bytes
func (s *SiteService) ExportSite(ctx context.Context, siteID uint) ([]byte, error) {
	site, err := s.siteRepo.FindByID(ctx, siteID)
	if err != nil {
		return nil, err
	}
	payload := SiteExportPayload{
		ExportedAt: time.Now().UTC(),
		Site:       site,
	}
	return json.MarshalIndent(payload, "", "  ")
}

// ImportSite deserialises a site payload and upserts the site
func (s *SiteService) ImportSite(ctx context.Context, data []byte) (*model.Site, error) {
	var payload SiteExportPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, errors.New("invalid site export payload: " + err.Error())
	}
	if payload.Site == nil {
		return nil, errors.New("export payload missing site data")
	}

	site := payload.Site
	// Reset ID so we create a new record (or caller can set it externally)
	site.ID = 0

	if err := s.siteRepo.Create(ctx, site); err != nil {
		return nil, err
	}
	return site, nil
}
