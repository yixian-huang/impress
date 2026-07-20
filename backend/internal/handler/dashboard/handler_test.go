package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

type stubArticleRepo struct {
	repository.ArticleRepository
	count int64
	err   error
}

func (s *stubArticleRepo) Count(ctx context.Context, status string) (int64, error) {
	return s.count, s.err
}

type stubPageRepo struct {
	repository.UnifiedPageRepository
	count int64
	err   error
}

func (s *stubPageRepo) Count(ctx context.Context) (int64, error) {
	return s.count, s.err
}

type stubMediaRepo struct {
	repository.MediaRepository
	count int64
	err   error
}

func (s *stubMediaRepo) Count(ctx context.Context) (int64, error) {
	return s.count, s.err
}

type stubPVRepo struct {
	repository.PageViewRepository
	count int64
	err   error
}

func (s *stubPVRepo) CountSince(ctx context.Context, since time.Time) (int64, error) {
	return s.count, s.err
}

func (s *stubPVRepo) Create(ctx context.Context, pv *model.PageView) error { return nil }
func (s *stubPVRepo) CreateBatch(ctx context.Context, views []*model.PageView) error {
	return nil
}
func (s *stubPVRepo) GetSummary(ctx context.Context, now time.Time) ([]repository.PageViewStats, error) {
	return nil, nil
}
func (s *stubPVRepo) CountByPageKey(ctx context.Context, pageKey string) (int64, error) {
	return 0, nil
}

func TestSummaryReturnsAggregatedCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(
		&stubArticleRepo{count: 12},
		&stubPageRepo{count: 4},
		&stubMediaRepo{count: 30},
		&stubPVRepo{count: 99},
	)

	r := gin.New()
	r.GET("/admin/dashboard/summary", h.Summary)

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.EqualValues(t, 99, body["todayVisits"])
	require.EqualValues(t, 4, body["pagesCount"])
	require.EqualValues(t, 12, body["articlesCount"])
	require.EqualValues(t, 30, body["mediaCount"])
}
