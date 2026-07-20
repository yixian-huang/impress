package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ArticleVersion stores a historical snapshot of an article for comparison/restore.
type ArticleVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID uint      `gorm:"not null;index;uniqueIndex:idx_av_article_version" json:"articleId"`
	Version   int       `gorm:"not null;uniqueIndex:idx_av_article_version" json:"version"`
	// Snapshot holds the full article field set at save time (JSON).
	Snapshot  JSONMap   `gorm:"type:jsonb;not null" json:"snapshot"`
	// Action is a short label: "save", "publish", "create", etc.
	Action    string    `gorm:"size:40" json:"action"`
	// Summary is a short human-readable note (e.g. title at save time).
	Summary   string    `gorm:"size:500" json:"summary"`
	CreatedBy uint      `gorm:"not null;index" json:"createdBy"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName overrides the default table name.
func (ArticleVersion) TableName() string {
	return "article_versions"
}

// Validate validates the article version model.
func (v *ArticleVersion) Validate() error {
	if v.ArticleID == 0 {
		return errors.New("articleId is required")
	}
	if v.Version < 1 {
		return errors.New("version must be >= 1")
	}
	if v.Snapshot == nil {
		return errors.New("snapshot is required")
	}
	return nil
}

// BeforeSave ensures Snapshot is initialized and validates.
func (v *ArticleVersion) BeforeSave(tx *gorm.DB) error {
	if v.Snapshot == nil {
		v.Snapshot = make(JSONMap)
	}
	return v.Validate()
}
