package media

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

type fakeMediaRepository struct {
	media   *model.Media
	deleted bool
}

func (r *fakeMediaRepository) Create(ctx context.Context, media *model.Media) error { return nil }

func (r *fakeMediaRepository) FindByID(ctx context.Context, id uint) (*model.Media, error) {
	return r.media, nil
}

func (r *fakeMediaRepository) List(ctx context.Context, offset, limit int, mimePrefix string) ([]*model.Media, int64, error) {
	return nil, 0, nil
}

func (r *fakeMediaRepository) Delete(ctx context.Context, id uint) error {
	r.deleted = true
	return nil
}

func (r *fakeMediaRepository) Update(ctx context.Context, media *model.Media) error { return nil }

func (r *fakeMediaRepository) FindUsages(ctx context.Context, mediaURL string) ([]repository.MediaUsage, error) {
	return nil, nil
}

func TestDeleteReturnsConflictWhenRecordedProviderUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeMediaRepository{
		media: &model.Media{
			ID:              1,
			URL:             "https://cdn.test/file.png",
			Filename:        "file.png",
			MimeType:        "image/png",
			Size:            10,
			StorageKey:      "file.png",
			StorageProvider: "s3",
		},
	}
	handler := NewHandler(repo, t.TempDir(), "")

	router := gin.New()
	router.DELETE("/media/:id", handler.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/media/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if repo.deleted {
		t.Fatal("database record was deleted despite unavailable storage provider")
	}
}
