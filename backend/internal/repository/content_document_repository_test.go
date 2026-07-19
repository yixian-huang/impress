package repository

import (
	"context"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestContentDocumentRepository_CRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewGormContentDocumentRepository(database.DB)
	ctx := context.Background()

	doc := &model.ContentDocument{
		PageKey:          model.PageKeyHome,
		DraftConfig:      model.JSONMap{"title": "Home Page"},
		DraftVersion:     1,
		PublishedConfig:  model.JSONMap{"title": "Published Home"},
		PublishedVersion: 1,
	}

	// Test Create
	err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to create content document: %v", err)
	}

	// Test FindByPageKey
	found, err := repo.FindByPageKey(ctx, model.PageKeyHome)
	if err != nil {
		t.Fatalf("Failed to find content document: %v", err)
	}

	if found.PageKey != model.PageKeyHome {
		t.Errorf("Expected page key %s, got %s", model.PageKeyHome, found.PageKey)
	}

	// Test Update
	doc.DraftConfig = model.JSONMap{"title": "Updated Home Page"}
	doc.DraftVersion = 2

	err = repo.Update(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to update content document: %v", err)
	}

	found, _ = repo.FindByPageKey(ctx, model.PageKeyHome)
	if found.DraftVersion != 2 {
		t.Errorf("Expected draft version 2, got %d", found.DraftVersion)
	}

	// Test List
	docs, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list content documents: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("Expected 1 document, got %d", len(docs))
	}

	// Test Delete
	err = repo.Delete(ctx, model.PageKeyHome)
	if err != nil {
		t.Fatalf("Failed to delete content document: %v", err)
	}

	_, err = repo.FindByPageKey(ctx, model.PageKeyHome)
	if err == nil {
		t.Error("Expected error when finding deleted document")
	}
}

func TestContentDocumentRepository_UpdateDraft(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewGormContentDocumentRepository(database.DB)
	ctx := context.Background()

	doc := &model.ContentDocument{
		PageKey:          model.PageKeyHome,
		DraftConfig:      model.JSONMap{"title": "Home Page"},
		DraftVersion:     1,
		PublishedConfig:  model.JSONMap{"title": "Published Home"},
		PublishedVersion: 1,
	}
	_ = repo.Create(ctx, doc)

	// Test UpdateDraft with correct version
	newDraftConfig := model.JSONMap{"title": "Updated Home Page", "content": "New content"}
	newVersion, err := repo.UpdateDraft(ctx, model.PageKeyHome, 1, newDraftConfig)
	if err != nil {
		t.Fatalf("Failed to update draft: %v", err)
	}

	if newVersion != 2 {
		t.Errorf("Expected new version 2, got %d", newVersion)
	}

	found, _ := repo.FindByPageKey(ctx, model.PageKeyHome)
	if found.DraftVersion != 2 {
		t.Errorf("Expected draft version 2, got %d", found.DraftVersion)
	}

	// Test UpdateDraft with wrong version (optimistic locking)
	_, err = repo.UpdateDraft(ctx, model.PageKeyHome, 1, newDraftConfig)
	if err == nil {
		t.Error("Expected error when updating with wrong expected version")
	}
}

func TestContentDocumentRepository_UpdatePublished(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewGormContentDocumentRepository(database.DB)
	ctx := context.Background()

	doc := &model.ContentDocument{
		PageKey:          model.PageKeyHome,
		DraftConfig:      model.JSONMap{"title": "Home Page"},
		DraftVersion:     2,
		PublishedConfig:  model.JSONMap{"title": "Published Home"},
		PublishedVersion: 1,
	}
	_ = repo.Create(ctx, doc)

	newPublishedConfig := model.JSONMap{"title": "Updated Published Home"}
	err := repo.UpdatePublished(ctx, model.PageKeyHome, newPublishedConfig, 2)
	if err != nil {
		t.Fatalf("Failed to update published: %v", err)
	}

	found, _ := repo.FindByPageKey(ctx, model.PageKeyHome)
	if found.PublishedVersion != 2 {
		t.Errorf("Expected published version 2, got %d", found.PublishedVersion)
	}

	if found.PublishedConfig["title"] != "Updated Published Home" {
		t.Error("Published config was not updated")
	}
}
