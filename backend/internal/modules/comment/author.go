package comment

import (
	"context"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

const (
	AuthorRoleGuest  = "guest"
	AuthorRoleAuthor = "author"
)

func authorNameFromGlobalConfig(cfg model.JSONMap) string {
	if cfg == nil {
		return ""
	}
	if author, ok := cfg["author"].(map[string]interface{}); ok {
		if name, ok := author["name"].(string); ok && name != "" {
			return name
		}
	}
	if identity, ok := cfg["identity"].(map[string]interface{}); ok {
		if name, ok := identity["name"].(map[string]interface{}); ok {
			if zh, ok := name["zh"].(string); ok && zh != "" {
				return zh
			}
			if en, ok := name["en"].(string); ok && en != "" {
				return en
			}
		}
	}
	return ""
}

func resolveSiteAuthorName(ctx context.Context, contentDoc repository.ContentDocumentRepository) string {
	if contentDoc == nil {
		return "Author"
	}
	doc, err := contentDoc.FindByPageKey(ctx, model.PageKeyGlobal)
	if err != nil || doc == nil || doc.PublishedConfig == nil {
		return "Author"
	}
	if name := authorNameFromGlobalConfig(doc.PublishedConfig); name != "" {
		return name
	}
	return "Author"
}
