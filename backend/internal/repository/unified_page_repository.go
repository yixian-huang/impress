package repository

import (
	"context"
	"errors"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"time"
)

var ErrUnifiedPageDraftVersionConflict = errors.New("draft version conflict")

type UnifiedPageRepository interface {
	Create(ctx context.Context, page *model.UnifiedPage) error
	Update(ctx context.Context, page *model.UnifiedPage) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.UnifiedPage, error)
	FindBySlug(ctx context.Context, slug string) (*model.UnifiedPage, error)
	// List returns admin list rows without draft/published JSON configs.
	List(ctx context.Context, status string, mode string, parentID *uint) ([]*model.UnifiedPage, error)
	ListPublished(ctx context.Context) ([]*model.UnifiedPage, error)
	// Count returns total unified pages (all statuses unless filtered later).
	Count(ctx context.Context) (int64, error)
	UpdateDraft(ctx context.Context, id uint, expectedVersion int, draftConfig model.JSONMap) (int, error)
	PublishDraft(ctx context.Context, id uint, expectedVersion int, userID uint, publishedAt time.Time, requireSchedule bool) (*model.UnifiedPage, int, bool, error)
	UpdatePublished(ctx context.Context, id uint, publishedConfig model.JSONMap, publishedVersion int, publishedAt time.Time) error
	UpdateRollback(ctx context.Context, id uint, draftConfig model.JSONMap, draftVersion int, publishedConfig model.JSONMap, publishedVersion int, publishedAt time.Time) error
	ClearPublished(ctx context.Context, id uint) error
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error
}
