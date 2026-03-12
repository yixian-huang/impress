package model

import (
	"time"

	"gorm.io/gorm"
)

// MarketplaceItemType represents whether an item is a plugin or theme
type MarketplaceItemType string

const (
	MarketplaceItemTypePlugin MarketplaceItemType = "plugin"
	MarketplaceItemTypeTheme  MarketplaceItemType = "theme"
)

// MarketplaceItemStatus represents the publication status of a marketplace item
type MarketplaceItemStatus string

const (
	MarketplaceItemStatusActive     MarketplaceItemStatus = "active"
	MarketplaceItemStatusDeprecated MarketplaceItemStatus = "deprecated"
	MarketplaceItemStatusDraft      MarketplaceItemStatus = "draft"
)

// MarketplaceItem represents a plugin or theme available in the marketplace
type MarketplaceItem struct {
	ID          uint                  `gorm:"primaryKey" json:"id"`
	Type        MarketplaceItemType   `gorm:"size:20;not null;index" json:"type"`
	Name        string                `gorm:"size:200;not null" json:"name"`
	NameZh      string                `gorm:"size:200" json:"nameZh"`
	Slug        string                `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Description string                `gorm:"size:2000" json:"description"`
	Author      string                `gorm:"size:200" json:"author"`
	Version     string                `gorm:"size:50;not null" json:"version"`
	IconURL     string                `gorm:"size:500" json:"iconUrl"`
	PreviewURL  string                `gorm:"size:500" json:"previewUrl"`
	DownloadURL string                `gorm:"size:500" json:"downloadUrl"`
	Downloads   int64                 `gorm:"default:0" json:"downloads"`
	Category    string                `gorm:"size:100;index" json:"category"`
	Tags        JSONStringSlice       `gorm:"type:text" json:"tags"`
	Status      MarketplaceItemStatus `gorm:"size:20;not null;default:'active';index" json:"status"`
	CreatedAt   time.Time             `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time             `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"-"`

	// HasOne relationship for versions
	Versions []MarketplaceVersion `gorm:"foreignKey:ItemID" json:"versions,omitempty"`
}

// TableName overrides the default table name
func (MarketplaceItem) TableName() string {
	return "marketplace_items"
}

// MarketplaceVersion represents a specific version of a marketplace item
type MarketplaceVersion struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ItemID        uint      `gorm:"not null;index" json:"itemId"`
	Version       string    `gorm:"size:50;not null" json:"version"`
	Changelog     string    `gorm:"type:text" json:"changelog"`
	DownloadURL   string    `gorm:"size:500" json:"downloadUrl"`
	MinAppVersion string    `gorm:"size:50" json:"minAppVersion"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName overrides the default table name
func (MarketplaceVersion) TableName() string {
	return "marketplace_versions"
}
