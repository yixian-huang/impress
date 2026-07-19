package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// SiteRepository defines the interface for site data access
type SiteRepository interface {
	// Create creates a new site
	Create(ctx context.Context, site *model.Site) error

	// FindByID finds a site by ID
	FindByID(ctx context.Context, id uint) (*model.Site, error)

	// FindByDomain finds an active site by its domain name
	FindByDomain(ctx context.Context, domain string) (*model.Site, error)

	// FindBySubPath finds an active site by its sub-path prefix
	FindBySubPath(ctx context.Context, subPath string) (*model.Site, error)

	// Update updates a site
	Update(ctx context.Context, site *model.Site) error

	// Delete deletes a site by ID
	Delete(ctx context.Context, id uint) error

	// List returns all sites with optional status filter
	List(ctx context.Context, status string) ([]*model.Site, error)

	// AddUser assigns a user to a site with a given role
	AddUser(ctx context.Context, siteUser *model.SiteUser) error

	// RemoveUser removes a user from a site
	RemoveUser(ctx context.Context, siteID, userID uint) error

	// ListUsers returns all user-role assignments for a site
	ListUsers(ctx context.Context, siteID uint) ([]*model.SiteUser, error)

	// FindUserRole returns the site-user-role record for a specific user on a site
	FindUserRole(ctx context.Context, siteID, userID uint) (*model.SiteUser, error)
}
