package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// ContentDocumentRepository defines the interface for content document data access
type ContentDocumentRepository interface {
	// Create creates a new content document
	Create(ctx context.Context, doc *model.ContentDocument) error

	// FindByPageKey finds a content document by page key
	FindByPageKey(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error)

	// Update updates an existing content document with optimistic locking
	// Returns error if the document has been modified (version mismatch)
	Update(ctx context.Context, doc *model.ContentDocument) error

	// UpdateDraft updates only the draft fields with optimistic locking
	UpdateDraft(ctx context.Context, pageKey model.PageKey, expectedDraftVersion int, draftConfig model.JSONMap) (int, error)

	// UpdatePublished updates only the published fields atomically
	UpdatePublished(ctx context.Context, pageKey model.PageKey, publishedConfig model.JSONMap, publishedVersion int) error

	// List returns all content documents
	List(ctx context.Context) ([]*model.ContentDocument, error)

	// Delete deletes a content document by page key
	Delete(ctx context.Context, pageKey model.PageKey) error
}
