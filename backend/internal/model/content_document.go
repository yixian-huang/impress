package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// PageKey represents a valid page key enum
type PageKey string

const (
	PageKeyHome         PageKey = "home"
	PageKeyAbout        PageKey = "about"
	PageKeyAdvantages   PageKey = "advantages"
	PageKeyCoreServices PageKey = "core-services"
	PageKeyCases        PageKey = "cases"
	PageKeyExperts      PageKey = "experts"
	PageKeyContact      PageKey = "contact"
	PageKeyGlobal       PageKey = "global"
	PageKeyTheme        PageKey = "theme"
)

// ValidPageKeys contains all valid page key values
var ValidPageKeys = []PageKey{
	PageKeyHome,
	PageKeyAbout,
	PageKeyAdvantages,
	PageKeyCoreServices,
	PageKeyCases,
	PageKeyExperts,
	PageKeyContact,
	PageKeyGlobal,
	PageKeyTheme,
}

// IsValid checks if a page key value is valid
func (pk PageKey) IsValid() bool {
	for _, valid := range ValidPageKeys {
		if pk == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the page key
func (pk PageKey) String() string {
	return string(pk)
}

// ContentDocument represents a page configuration document
type ContentDocument struct {
	PageKey          PageKey   `gorm:"primaryKey;size:50"`
	DraftConfig      JSONMap   `gorm:"type:jsonb"`
	DraftVersion     int       `gorm:"not null;default:0"`
	PublishedConfig  JSONMap   `gorm:"type:jsonb"`
	PublishedVersion int       `gorm:"not null;default:0"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// TableName overrides the default table name
func (ContentDocument) TableName() string {
	return "content_documents"
}

// Validate validates the content document model
func (cd *ContentDocument) Validate() error {
	if !cd.PageKey.IsValid() {
		return errors.New("pageKey must be one of: home, about, advantages, core-services, cases, experts, contact, global, theme")
	}
	if cd.DraftVersion < 0 {
		return errors.New("draftVersion cannot be negative")
	}
	if cd.PublishedVersion < 0 {
		return errors.New("publishedVersion cannot be negative")
	}
	return nil
}

// BeforeSave hook to ensure JSON fields are initialized
func (cd *ContentDocument) BeforeSave(tx *gorm.DB) error {
	if cd.DraftConfig == nil {
		cd.DraftConfig = make(JSONMap)
	}
	if cd.PublishedConfig == nil {
		cd.PublishedConfig = make(JSONMap)
	}
	return nil
}
