package repository

import (
	"context"
	"github.com/yixian-huang/inkless/backend/internal/model"
)

type PageTemplateRepository interface {
	Create(ctx context.Context, tmpl *model.PageTemplate) error
	Update(ctx context.Context, tmpl *model.PageTemplate) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.PageTemplate, error)
	FindByKey(ctx context.Context, key string) (*model.PageTemplate, error)
	List(ctx context.Context, category string) ([]*model.PageTemplate, error)
}
