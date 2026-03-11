package model

import "time"

// SearchIndexEntry tracks what content is indexed (for rebuild tracking).
type SearchIndexEntry struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ContentType string    `json:"contentType" gorm:"size:20;index:idx_search_content,unique"`
	ContentID   uint      `json:"contentId" gorm:"index:idx_search_content,unique"`
	Locale      string    `json:"locale" gorm:"size:5"`
	IndexedAt   time.Time `json:"indexedAt" gorm:"autoCreateTime"`
}
