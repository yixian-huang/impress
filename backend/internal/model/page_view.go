package model

import "time"

// PageView records a single page visit
type PageView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PageKey   string    `gorm:"not null;size:50;index:idx_pv_lookup,priority:1" json:"pageKey"`
	Locale    string    `gorm:"not null;size:5" json:"locale"`
	ViewedAt  time.Time `gorm:"not null;index:idx_pv_lookup,priority:2;autoCreateTime" json:"viewedAt"`
	VisitorID string    `gorm:"size:64;index" json:"visitorId"`
	Referer   string    `gorm:"size:500" json:"referer"`
}
