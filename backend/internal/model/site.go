package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// SiteStatus represents the operational status of a site
type SiteStatus string

const (
	SiteStatusActive   SiteStatus = "active"
	SiteStatusInactive SiteStatus = "inactive"
)

// SiteMode represents how a site is addressed
type SiteMode string

const (
	SiteModeSubdomain SiteMode = "subdomain"
	SiteModeSubpath   SiteMode = "subpath"
)

// SiteSettings holds JSON-encoded per-site configuration
type SiteSettings map[string]interface{}

// Value implements driver.Valuer for database storage
func (s SiteSettings) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner for database retrieval
func (s *SiteSettings) Scan(value interface{}) error {
	if value == nil {
		*s = SiteSettings{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("unsupported type for SiteSettings")
	}
	return json.Unmarshal(bytes, s)
}

// Site represents a managed website in the multi-site system
type Site struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// Domain is the primary hostname (e.g. "example.com" or "sub.example.com")
	Domain string `gorm:"uniqueIndex;not null;size:255" json:"domain"`
	// SubPath is the URL prefix when running in subpath mode (e.g. "/blog")
	SubPath string `gorm:"size:100;default:''" json:"subPath"`
	Name    string `gorm:"not null;size:200" json:"name"`
	// Locale is the default locale for this site (e.g. "zh", "en")
	Locale string `gorm:"not null;size:10;default:'zh'" json:"locale"`
	// ThemeID references the active installed theme for this site
	ThemeID string `gorm:"size:100" json:"themeId"`
	// Mode controls how the site is addressed: subdomain or subpath
	Mode SiteMode `gorm:"not null;size:20;default:'subdomain'" json:"mode"`
	// Settings stores arbitrary JSON config (SEO defaults, feature flags, etc.)
	Settings  SiteSettings `gorm:"type:text" json:"settings"`
	Status    SiteStatus   `gorm:"not null;size:20;default:'active';index" json:"status"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName overrides the default table name
func (Site) TableName() string {
	return "sites"
}

// Validate validates the site model
func (s *Site) Validate() error {
	if strings.TrimSpace(s.Domain) == "" {
		return errors.New("domain is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	if s.Locale == "" {
		return errors.New("locale is required")
	}
	if s.Mode != SiteModeSubdomain && s.Mode != SiteModeSubpath {
		return errors.New("mode must be 'subdomain' or 'subpath'")
	}
	if s.Status != SiteStatusActive && s.Status != SiteStatusInactive {
		return errors.New("status must be 'active' or 'inactive'")
	}
	if s.Mode == SiteModeSubpath && strings.TrimSpace(s.SubPath) == "" {
		return errors.New("subPath is required when mode is 'subpath'")
	}
	return nil
}
