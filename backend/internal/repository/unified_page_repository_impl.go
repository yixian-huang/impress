package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormUnifiedPageRepository struct {
	db *gorm.DB
}

func NewGormUnifiedPageRepository(db *gorm.DB) UnifiedPageRepository {
	return &GormUnifiedPageRepository{db: db}
}

func (r *GormUnifiedPageRepository) Create(ctx context.Context, page *model.UnifiedPage) error {
	return r.db.WithContext(ctx).Create(page).Error
}

func (r *GormUnifiedPageRepository) Update(ctx context.Context, page *model.UnifiedPage) error {
	return r.db.WithContext(ctx).Save(page).Error
}

func (r *GormUnifiedPageRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.UnifiedPage{}, id)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *GormUnifiedPageRepository) FindByID(ctx context.Context, id uint) (*model.UnifiedPage, error) {
	var page model.UnifiedPage
	if err := r.db.WithContext(ctx).First(&page, id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *GormUnifiedPageRepository) FindBySlug(ctx context.Context, slug string) (*model.UnifiedPage, error) {
	var page model.UnifiedPage
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

// unifiedPageListSelectColumns omits large JSON config blobs from list queries.
const unifiedPageListSelectColumns = "id, slug, zh_title, en_title, zh_description, en_description, " +
	"mode, template_id, draft_version, published_version, status, scheduled_at, " +
	"zh_meta_title, en_meta_title, zh_meta_description, en_meta_description, " +
	"zh_meta_keywords, en_meta_keywords, sort_order, show_in_nav, parent_id, " +
	"created_at, updated_at, published_at, deleted_at"

// unifiedPagePublishedSelectColumns includes published_config for public listing
// but omits draft_config / translation_status (often multi-KB drafts).
const unifiedPagePublishedSelectColumns = unifiedPageListSelectColumns + ", published_config"

func (r *GormUnifiedPageRepository) List(ctx context.Context, status string, mode string, parentID *uint) ([]*model.UnifiedPage, error) {
	q := r.db.WithContext(ctx).Model(&model.UnifiedPage{}).Select(unifiedPageListSelectColumns)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if mode != "" {
		q = q.Where("mode = ?", mode)
	}
	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	}
	var pages []*model.UnifiedPage
	err := q.Order("sort_order ASC, created_at DESC").Find(&pages).Error
	return pages, err
}

func (r *GormUnifiedPageRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.UnifiedPage{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *GormUnifiedPageRepository) ListPublished(ctx context.Context) ([]*model.UnifiedPage, error) {
	var pages []*model.UnifiedPage
	err := r.db.WithContext(ctx).
		Select(unifiedPagePublishedSelectColumns).
		Where("status = ?", "published").
		Order("sort_order ASC, created_at DESC").
		Find(&pages).Error
	return pages, err
}

func (r *GormUnifiedPageRepository) UpdateDraft(ctx context.Context, id uint, expectedVersion int, draftConfig model.JSONMap) (int, error) {
	result := r.db.WithContext(ctx).Table("unified_pages").Where("id = ? AND draft_version = ?", id, expectedVersion).Updates(map[string]interface{}{
		"draft_config":  draftConfig,
		"draft_version": gorm.Expr("draft_version + 1"),
	})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errors.New("draft version conflict or page not found")
	}
	var page model.UnifiedPage
	if err := r.db.WithContext(ctx).Select("draft_version").First(&page, id).Error; err != nil {
		return 0, fmt.Errorf("fetch new version: %w", err)
	}
	return page.DraftVersion, nil
}

func (r *GormUnifiedPageRepository) PublishDraft(
	ctx context.Context,
	id uint,
	expectedVersion int,
	userID uint,
	publishedAt time.Time,
	requireSchedule bool,
) (*model.UnifiedPage, int, bool, error) {
	var publishedPage *model.UnifiedPage
	var publishedVersion int
	didPublish := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var page model.UnifiedPage
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&page, id).Error; err != nil {
			return err
		}
		if requireSchedule && page.ScheduledAt == nil && page.Status == "published" {
			publishedPage = &page
			publishedVersion = page.PublishedVersion
			return nil
		}
		if page.DraftVersion != expectedVersion {
			return ErrUnifiedPageDraftVersionConflict
		}

		var latest struct{ Max int }
		if err := tx.Model(&model.PageVersion{}).
			Select("COALESCE(MAX(version), 0) as max").
			Where("page_id = ?", id).
			Scan(&latest).Error; err != nil {
			return err
		}
		publishedVersion = latest.Max + 1
		version := &model.PageVersion{
			PageID:    id,
			Version:   publishedVersion,
			Config:    page.DraftConfig,
			CreatedBy: userID,
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}

		result := tx.Table("unified_pages").
			Where("id = ? AND draft_version = ?", id, expectedVersion).
			Updates(map[string]interface{}{
				"published_config":  page.DraftConfig,
				"published_version": publishedVersion,
				"status":            "published",
				"published_at":      publishedAt,
				"scheduled_at":      nil,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrUnifiedPageDraftVersionConflict
		}
		page.PublishedConfig = model.NullableJSONMap(page.DraftConfig)
		page.PublishedVersion = publishedVersion
		page.Status = "published"
		page.PublishedAt = &publishedAt
		page.ScheduledAt = nil
		publishedPage = &page
		didPublish = true
		return nil
	})
	if err != nil {
		return nil, 0, false, err
	}
	return publishedPage, publishedVersion, didPublish, nil
}

func (r *GormUnifiedPageRepository) UpdatePublished(
	ctx context.Context,
	id uint,
	publishedConfig model.JSONMap,
	publishedVersion int,
	publishedAt time.Time,
) error {
	result := r.db.WithContext(ctx).Table("unified_pages").Where("id = ?", id).Updates(map[string]interface{}{
		"published_config":  publishedConfig,
		"published_version": publishedVersion,
		"status":            "published",
		"published_at":      publishedAt,
		"scheduled_at":      nil,
	})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *GormUnifiedPageRepository) UpdateRollback(
	ctx context.Context,
	id uint,
	draftConfig model.JSONMap,
	draftVersion int,
	publishedConfig model.JSONMap,
	publishedVersion int,
	publishedAt time.Time,
) error {
	result := r.db.WithContext(ctx).Table("unified_pages").Where("id = ?", id).Updates(map[string]interface{}{
		"draft_config":      draftConfig,
		"draft_version":     draftVersion,
		"published_config":  publishedConfig,
		"published_version": publishedVersion,
		"status":            "published",
		"published_at":      publishedAt,
		"scheduled_at":      nil,
	})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *GormUnifiedPageRepository) ClearPublished(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Table("unified_pages").Where("id = ?", id).Updates(map[string]interface{}{
		"published_config": nil,
		"status":           "draft",
		"scheduled_at":     nil,
	})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *GormUnifiedPageRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	result := r.db.WithContext(ctx).Table("unified_pages").Where("id = ?", id).Update("sort_order", sortOrder)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
