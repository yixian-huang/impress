package model

import (
	"errors"
	"time"
)

// APIKey is a long-lived personal access token for machine clients (PicGo, CLI).
// Only the hash is stored; plaintext is shown once at creation.
type APIKey struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	UserID      uint        `gorm:"not null;index" json:"userId"`
	Name        string      `gorm:"not null;size:64" json:"name"`
	TokenPrefix string      `gorm:"not null;size:16" json:"tokenPrefix"` // ink_ + first 4 of secret for display
	TokenHash   string      `gorm:"not null;uniqueIndex;size:64" json:"-"`
	Scopes      StringSlice `gorm:"type:text;not null" json:"scopes"` // e.g. ["media:create"]
	LastUsedAt  *time.Time  `json:"lastUsedAt,omitempty"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"createdAt"`
}

func (APIKey) TableName() string { return "api_keys" }

func (k *APIKey) Validate() error {
	if k.UserID == 0 {
		return errors.New("user id is required")
	}
	if k.Name == "" {
		return errors.New("name is required")
	}
	if k.TokenHash == "" {
		return errors.New("token hash is required")
	}
	if len(k.Scopes) == 0 {
		return errors.New("scopes are required")
	}
	return nil
}
