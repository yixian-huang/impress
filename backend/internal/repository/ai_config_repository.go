package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// AIConfigRepository defines singleton AI config persistence.
type AIConfigRepository interface {
	Get(ctx context.Context) (*model.AIConfig, error)
	Upsert(ctx context.Context, config *model.AIConfig) error
}
