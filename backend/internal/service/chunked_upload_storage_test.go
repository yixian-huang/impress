package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

type fakeChunkedUploadRepo struct {
	upload  *model.ChunkedUpload
	updated *model.ChunkedUpload
}

func (r *fakeChunkedUploadRepo) Create(ctx context.Context, upload *model.ChunkedUpload) error {
	r.upload = upload
	return nil
}

func (r *fakeChunkedUploadRepo) FindByID(ctx context.Context, id string) (*model.ChunkedUpload, error) {
	return r.upload, nil
}

func (r *fakeChunkedUploadRepo) Update(ctx context.Context, upload *model.ChunkedUpload) error {
	cp := *upload
	r.updated = &cp
	return nil
}

func (r *fakeChunkedUploadRepo) Delete(ctx context.Context, id string) error { return nil }

type fakeMediaRepo struct {
	created *model.Media
}

func (r *fakeMediaRepo) Create(ctx context.Context, media *model.Media) error {
	cp := *media
	r.created = &cp
	return nil
}

func (r *fakeMediaRepo) FindByID(ctx context.Context, id uint) (*model.Media, error) {
	return nil, nil
}

func (r *fakeMediaRepo) List(ctx context.Context, offset, limit int, mimePrefix string) ([]*model.Media, int64, error) {
	return nil, 0, nil
}

func (r *fakeMediaRepo) Count(ctx context.Context) (int64, error) { return 0, nil }

func (r *fakeMediaRepo) Delete(ctx context.Context, id uint) error { return nil }
func (r *fakeMediaRepo) Update(ctx context.Context, media *model.Media) error {
	return nil
}
func (r *fakeMediaRepo) FindUsages(ctx context.Context, mediaURL string) ([]repository.MediaUsage, error) {
	return nil, nil
}

func TestChunkedUploadCompleteSavesThroughStorageRuntime(t *testing.T) {
	tempDir := t.TempDir()
	uploadDir := t.TempDir()
	uploadTempDir := filepath.Join(tempDir, "upload-1")
	if err := os.MkdirAll(uploadTempDir, 0755); err != nil {
		t.Fatalf("mkdir upload temp: %v", err)
	}
	if err := os.WriteFile(filepath.Join(uploadTempDir, "chunk_000000"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write chunk: %v", err)
	}
	if err := os.WriteFile(filepath.Join(uploadTempDir, "chunk_000001"), []byte("world"), 0644); err != nil {
		t.Fatalf("write chunk: %v", err)
	}

	storage := &fakeStorageProvider{}
	runtime := NewStorageRuntimeService(nil, provider.NewRegistry(), storage, nil)
	chunkRepo := &fakeChunkedUploadRepo{upload: &model.ChunkedUpload{
		ID:          "upload-1",
		Filename:    "example.txt",
		MimeType:    "text/plain",
		TotalSize:   10,
		TotalChunks: 2,
		ChunkSize:   5,
		Status:      model.ChunkedUploadUploading,
		TempDir:     uploadTempDir,
		ExpiresAt:   time.Now().Add(time.Hour),
	}}
	mediaRepo := &fakeMediaRepo{}
	svc := NewChunkedUploadServiceWithStorage(chunkRepo, mediaRepo, tempDir, uploadDir, "", runtime)

	media, err := svc.CompleteUpload(context.Background(), "upload-1")
	if err != nil {
		t.Fatalf("CompleteUpload failed: %v", err)
	}
	if string(storage.savedData) != "helloworld" {
		t.Fatalf("saved data = %q, want helloworld", string(storage.savedData))
	}
	if media.StorageProvider != "local" {
		t.Fatalf("storage provider = %q, want local", media.StorageProvider)
	}
	if media.StorageKey == "" {
		t.Fatal("storage key was not persisted on media")
	}
	if media.URL != "/uploads/"+media.StorageKey {
		t.Fatalf("url = %q, want local uploads URL", media.URL)
	}
	if chunkRepo.updated == nil || chunkRepo.updated.Status != model.ChunkedUploadCompleted {
		t.Fatal("upload status was not marked completed")
	}
	if mediaRepo.created == nil {
		t.Fatal("media record was not created")
	}
}
