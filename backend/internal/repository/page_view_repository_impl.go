package repository

import (
	"context"
	"time"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormPageViewRepository implements PageViewRepository using GORM
type GormPageViewRepository struct {
	db *gorm.DB
}

// NewGormPageViewRepository creates a new GormPageViewRepository
func NewGormPageViewRepository(db *gorm.DB) PageViewRepository {
	return &GormPageViewRepository{db: db}
}

// Create records a new page view
func (r *GormPageViewRepository) Create(ctx context.Context, pv *model.PageView) error {
	return r.db.WithContext(ctx).Create(pv).Error
}

// GetSummary returns aggregated view stats for all pages within the last 30 days.
// Uses standard SQL-92 compatible with both SQLite and PostgreSQL.
func (r *GormPageViewRepository) GetSummary(ctx context.Context, now time.Time) ([]PageViewStats, error) {
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := startOfToday.AddDate(0, 0, -6)
	thirtyDaysAgo := startOfToday.AddDate(0, 0, -29)

	var stats []PageViewStats
	err := r.db.WithContext(ctx).
		Model(&model.PageView{}).
		Select(`page_key,
			SUM(CASE WHEN viewed_at >= ? THEN 1 ELSE 0 END) as today,
			SUM(CASE WHEN viewed_at >= ? THEN 1 ELSE 0 END) as last7d,
			COUNT(*) as last30d,
			COUNT(DISTINCT visitor_id) as unique_visitors`,
			startOfToday, sevenDaysAgo).
		Where("viewed_at >= ?", thirtyDaysAgo).
		Group("page_key").
		Order("last30d DESC").
		Find(&stats).Error

	return stats, err
}
