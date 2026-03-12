package model

import (
	"errors"
	"time"
)

// MediaFolder represents a folder for organizing media files
type MediaFolder struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"not null;size:255" json:"name"`
	ParentID  *uint         `gorm:"index" json:"parentId,omitempty"`
	Path      string        `gorm:"not null;size:1000;index" json:"path"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	Parent    *MediaFolder  `gorm:"foreignKey:ParentID" json:"-"`
	Children  []MediaFolder `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName overrides the default table name
func (MediaFolder) TableName() string {
	return "media_folders"
}

// Validate validates the media folder model
func (f *MediaFolder) Validate() error {
	if f.Name == "" {
		return errors.New("name is required")
	}
	if len(f.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	return nil
}
