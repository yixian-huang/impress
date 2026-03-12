package repository

import (
	"context"

	"blotting-consultancy/internal/model"
)

// GlossaryRepository defines the interface for glossary data access
type GlossaryRepository interface {
	// Create creates a new glossary term
	Create(ctx context.Context, glossary *model.Glossary) error

	// FindByID finds a glossary term by ID
	FindByID(ctx context.Context, id uint) (*model.Glossary, error)

	// Update updates a glossary term
	Update(ctx context.Context, glossary *model.Glossary) error

	// Delete deletes a glossary term by ID
	Delete(ctx context.Context, id uint) error

	// List returns a paginated list of glossary terms with optional language filter
	List(ctx context.Context, offset, limit int, sourceLang, targetLang string) ([]*model.Glossary, int64, error)

	// FindByLangs returns all glossary terms for a given language pair
	FindByLangs(ctx context.Context, sourceLang, targetLang string) ([]*model.Glossary, error)
}
