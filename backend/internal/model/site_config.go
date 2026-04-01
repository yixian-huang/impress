package model

import (
	"errors"
	"time"
	"gorm.io/gorm"
)

const (
	SiteConfigKeyGlobal   = "global"
	SiteConfigKeyTheme    = "theme"
	SiteConfigKeyEmail    = "email"
	SiteConfigKeyFeatures = "features"
)

type SiteConfig struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Key              string    `gorm:"uniqueIndex;size:50;not null" json:"key"`
	DraftConfig      JSONMap   `gorm:"type:text" json:"draftConfig"`
	DraftVersion     int       `gorm:"not null;default:1" json:"draftVersion"`
	PublishedConfig  JSONMap   `gorm:"type:text" json:"publishedConfig"`
	PublishedVersion int       `gorm:"not null;default:0" json:"publishedVersion"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (SiteConfig) TableName() string { return "site_configs" }

func (sc *SiteConfig) Validate() error {
	if sc.Key != SiteConfigKeyGlobal && sc.Key != SiteConfigKeyTheme && sc.Key != SiteConfigKeyEmail && sc.Key != SiteConfigKeyFeatures {
		return errors.New("key must be 'global', 'theme', 'email', or 'features'")
	}
	return nil
}

func (sc *SiteConfig) BeforeSave(tx *gorm.DB) error {
	return sc.Validate()
}
