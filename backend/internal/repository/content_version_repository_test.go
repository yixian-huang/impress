package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestContentVersionRepository_CRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	versionRepo := NewGormContentVersionRepository(database.DB)
	ctx := context.Background()

	// Create user first
	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	version := &model.ContentVersion{
		PageKey:     model.PageKeyHome,
		Version:     1,
		Config:      model.JSONMap{"title": "Home Page V1"},
		PublishedAt: time.Now(),
		CreatedBy:   user.ID,
	}

	// Test Create
	err := versionRepo.Create(ctx, version)
	if err != nil {
		t.Fatalf("Failed to create content version: %v", err)
	}

	if version.ID == 0 {
		t.Error("Expected version ID to be set after creation")
	}

	// Test FindByID
	found, err := versionRepo.FindByID(ctx, version.ID)
	if err != nil {
		t.Fatalf("Failed to find content version: %v", err)
	}

	if found.PageKey != model.PageKeyHome {
		t.Errorf("Expected page key %s, got %s", model.PageKeyHome, found.PageKey)
	}

	// Test FindByPageKeyAndVersion
	found2, err := versionRepo.FindByPageKeyAndVersion(ctx, model.PageKeyHome, 1)
	if err != nil {
		t.Fatalf("Failed to find content version: %v", err)
	}

	if found2.Version != 1 {
		t.Errorf("Expected version 1, got %d", found2.Version)
	}

	// Test Delete
	err = versionRepo.Delete(ctx, version.ID)
	if err != nil {
		t.Fatalf("Failed to delete version: %v", err)
	}

	_, err = versionRepo.FindByID(ctx, version.ID)
	if err == nil {
		t.Error("Expected error when finding deleted version")
	}
}

func TestContentVersionRepository_ListByPageKey(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	versionRepo := NewGormContentVersionRepository(database.DB)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	// Create multiple versions for the same page
	for i := 1; i <= 5; i++ {
		version := &model.ContentVersion{
			PageKey:     model.PageKeyHome,
			Version:     i,
			Config:      model.JSONMap{"title": "Home Page V" + string(rune('0'+i))},
			PublishedAt: time.Now(),
			CreatedBy:   user.ID,
		}
		_ = versionRepo.Create(ctx, version)
	}

	// Create version for another page
	version := &model.ContentVersion{
		PageKey:     model.PageKeyAbout,
		Version:     1,
		Config:      model.JSONMap{"title": "About Page V1"},
		PublishedAt: time.Now(),
		CreatedBy:   user.ID,
	}
	_ = versionRepo.Create(ctx, version)

	versions, total, err := versionRepo.ListByPageKey(ctx, model.PageKeyHome, 0, 10)
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected 5 versions, got %d", total)
	}

	if len(versions) != 5 {
		t.Errorf("Expected 5 versions in list, got %d", len(versions))
	}

	// Verify versions are ordered descending
	if versions[0].Version < versions[len(versions)-1].Version {
		t.Error("Expected versions to be ordered by version descending")
	}

	// Test pagination
	versions, _, _ = versionRepo.ListByPageKey(ctx, model.PageKeyHome, 0, 3)
	if len(versions) != 3 {
		t.Errorf("Expected 3 versions in first page, got %d", len(versions))
	}
}

func TestContentVersionRepository_GetLatestVersion(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	versionRepo := NewGormContentVersionRepository(database.DB)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	// Test with no versions
	latestVersion, err := versionRepo.GetLatestVersion(ctx, model.PageKeyHome)
	if err != nil {
		t.Fatalf("Failed to get latest version: %v", err)
	}

	if latestVersion != 0 {
		t.Errorf("Expected 0 for page with no versions, got %d", latestVersion)
	}

	// Create versions
	for i := 1; i <= 5; i++ {
		version := &model.ContentVersion{
			PageKey:     model.PageKeyHome,
			Version:     i,
			Config:      model.JSONMap{"title": "Home Page V" + string(rune('0'+i))},
			PublishedAt: time.Now(),
			CreatedBy:   user.ID,
		}
		_ = versionRepo.Create(ctx, version)
	}

	latestVersion, err = versionRepo.GetLatestVersion(ctx, model.PageKeyHome)
	if err != nil {
		t.Fatalf("Failed to get latest version: %v", err)
	}

	if latestVersion != 5 {
		t.Errorf("Expected latest version 5, got %d", latestVersion)
	}
}

func TestContentVersionRepository_UniqueConstraint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	versionRepo := NewGormContentVersionRepository(database.DB)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	// Create first version
	version1 := &model.ContentVersion{
		PageKey:     model.PageKeyHome,
		Version:     1,
		Config:      model.JSONMap{"title": "Home Page V1"},
		PublishedAt: time.Now(),
		CreatedBy:   user.ID,
	}
	err := versionRepo.Create(ctx, version1)
	if err != nil {
		t.Fatalf("Failed to create first version: %v", err)
	}

	// Try to create duplicate (same page key and version)
	version2 := &model.ContentVersion{
		PageKey:     model.PageKeyHome,
		Version:     1,
		Config:      model.JSONMap{"title": "Duplicate Version"},
		PublishedAt: time.Now(),
		CreatedBy:   user.ID,
	}
	err = versionRepo.Create(ctx, version2)
	if err == nil {
		t.Error("Expected error when creating duplicate page key and version")
	}

	// Create version with same version number but different page key (should succeed)
	version3 := &model.ContentVersion{
		PageKey:     model.PageKeyAbout,
		Version:     1,
		Config:      model.JSONMap{"title": "About Page V1"},
		PublishedAt: time.Now(),
		CreatedBy:   user.ID,
	}
	err = versionRepo.Create(ctx, version3)
	if err != nil {
		t.Fatalf("Failed to create version for different page: %v", err)
	}
}
