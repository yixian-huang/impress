package comment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubCommentRepo struct {
	lastPage     int
	lastPageSize int
}

func (s *stubCommentRepo) Create(context.Context, *Comment) error { return nil }
func (s *stubCommentRepo) FindByID(context.Context, uint) (*Comment, error) {
	return nil, nil
}
func (s *stubCommentRepo) Update(context.Context, *Comment) error  { return nil }
func (s *stubCommentRepo) Delete(context.Context, uint) error      { return nil }
func (s *stubCommentRepo) CountByContent(context.Context, string, uint) (int64, error) {
	return 0, nil
}
func (s *stubCommentRepo) UpdateStatus(context.Context, uint, CommentStatus) error { return nil }
func (s *stubCommentRepo) SetPinned(context.Context, uint, bool) error             { return nil }

func (s *stubCommentRepo) ListByContent(_ context.Context, _ string, _ uint, _ CommentStatus, page, pageSize int) ([]*Comment, int64, error) {
	s.lastPage = page
	s.lastPageSize = pageSize
	return []*Comment{}, 0, nil
}

func (s *stubCommentRepo) ListAll(_ context.Context, _ string, page, pageSize int) ([]*Comment, int64, error) {
	s.lastPage = page
	s.lastPageSize = pageSize
	return []*Comment{}, 0, nil
}

func TestPublicListClampsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCommentRepo{}
	h := &Handler{repo: repo}

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/public/comments", h.PublicList)

	// page=0 and huge pageSize must clamp via handlerutil.ParsePagination
	req := httptest.NewRequest(http.MethodGet, "/public/comments?contentType=article&contentId=1&page=0&pageSize=9999", nil)
	r.ServeHTTP(w, req)
	_ = c

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if repo.lastPage != 1 {
		t.Fatalf("expected page clamped to 1, got %d", repo.lastPage)
	}
	if repo.lastPageSize != 100 {
		t.Fatalf("expected pageSize clamped to 100, got %d", repo.lastPageSize)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["page"] != float64(1) {
		t.Fatalf("response page=%v", body["page"])
	}
	if body["pageSize"] != float64(100) {
		t.Fatalf("response pageSize=%v", body["pageSize"])
	}
}

func TestPublicListRejectsMissingContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{repo: &stubCommentRepo{}}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/public/comments", h.PublicList)
	req := httptest.NewRequest(http.MethodGet, "/public/comments?page=-1&pageSize=20", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
