package model

import (
	"errors"
	"time"
)

// Media represents an uploaded media file
type Media struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	URL             string    `gorm:"not null;size:500" json:"url"`
	Filename        string    `gorm:"not null;size:255" json:"filename"`
	MimeType        string    `gorm:"not null;size:100" json:"mimeType"`
	Size            int64     `gorm:"not null" json:"size"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	StorageKey      string    `gorm:"size:500" json:"storageKey,omitempty"`
	StorageProvider string    `gorm:"size:50" json:"storageProvider,omitempty"`
	// FolderID associates media with a media folder (null = unfiled / root).
	FolderID        *uint     `gorm:"index" json:"folderId,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName overrides the default table name
func (Media) TableName() string {
	return "media"
}

// Validate validates the media model
func (m *Media) Validate() error {
	if m.URL == "" {
		return errors.New("url is required")
	}
	if m.Filename == "" {
		return errors.New("filename is required")
	}
	if m.MimeType == "" {
		return errors.New("mimeType is required")
	}
	if m.Size <= 0 {
		return errors.New("size must be positive")
	}
	return nil
}
