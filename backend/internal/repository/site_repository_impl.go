package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// GormSiteRepository implements SiteRepository using GORM
type GormSiteRepository struct {
	db *gorm.DB
}

// NewGormSiteRepository creates a new GormSiteRepository
func NewGormSiteRepository(db *gorm.DB) SiteRepository {
	return &GormSiteRepository{db: db}
}

// Create creates a new site
func (r *GormSiteRepository) Create(ctx context.Context, site *model.Site) error {
	if err := site.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(site).Error
}

// FindByID finds a site by ID
func (r *GormSiteRepository) FindByID(ctx context.Context, id uint) (*model.Site, error) {
	var site model.Site
	err := r.db.WithContext(ctx).First(&site, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("site not found")
		}
		return nil, err
	}
	return &site, nil
}

// FindByDomain finds an active site by its domain name
func (r *GormSiteRepository) FindByDomain(ctx context.Context, domain string) (*model.Site, error) {
	var site model.Site
	err := r.db.WithContext(ctx).
		Where("domain = ? AND status = ?", domain, model.SiteStatusActive).
		First(&site).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("site not found")
		}
		return nil, err
	}
	return &site, nil
}

// FindBySubPath finds an active site by its sub-path prefix
func (r *GormSiteRepository) FindBySubPath(ctx context.Context, subPath string) (*model.Site, error) {
	var site model.Site
	err := r.db.WithContext(ctx).
		Where("sub_path = ? AND mode = ? AND status = ?", subPath, model.SiteModeSubpath, model.SiteStatusActive).
		First(&site).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("site not found")
		}
		return nil, err
	}
	return &site, nil
}

// Update updates a site
func (r *GormSiteRepository) Update(ctx context.Context, site *model.Site) error {
	if err := site.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(site).Error
}

// Delete deletes a site by ID
func (r *GormSiteRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Site{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("site not found")
	}
	return nil
}

// List returns all sites with optional status filter
func (r *GormSiteRepository) List(ctx context.Context, status string) ([]*model.Site, error) {
	var sites []*model.Site
	query := r.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at ASC").Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

// AddUser assigns a user to a site with a given role
func (r *GormSiteRepository) AddUser(ctx context.Context, siteUser *model.SiteUser) error {
	return r.db.WithContext(ctx).
		Where(model.SiteUser{SiteID: siteUser.SiteID, UserID: siteUser.UserID}).
		Assign(model.SiteUser{RoleID: siteUser.RoleID}).
		FirstOrCreate(siteUser).Error
}

// RemoveUser removes a user from a site
func (r *GormSiteRepository) RemoveUser(ctx context.Context, siteID, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("site_id = ? AND user_id = ?", siteID, userID).
		Delete(&model.SiteUser{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("site user assignment not found")
	}
	return nil
}

// ListUsers returns all user-role assignments for a site
func (r *GormSiteRepository) ListUsers(ctx context.Context, siteID uint) ([]*model.SiteUser, error) {
	var users []*model.SiteUser
	if err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Preload("User").
		Preload("Role").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindUserRole returns the site-user-role record for a specific user on a site
func (r *GormSiteRepository) FindUserRole(ctx context.Context, siteID, userID uint) (*model.SiteUser, error) {
	var su model.SiteUser
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND user_id = ?", siteID, userID).
		Preload("Role").
		First(&su).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("site user assignment not found")
		}
		return nil, err
	}
	return &su, nil
}
