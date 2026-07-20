package analytics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

type stubPV struct {
	calls int
}

func (s *stubPV) Create(ctx context.Context, pv *model.PageView) error { return nil }
func (s *stubPV) GetSummary(ctx context.Context, now time.Time) ([]repository.PageViewStats, error) {
	s.calls++
	return []repository.PageViewStats{{PageKey: "home", Today: 3, Last7d: 10, Last30d: 20, UniqueVisitors: 5}}, nil
}
func (s *stubPV) CountByPageKey(ctx context.Context, pageKey string) (int64, error) { return 0, nil }
func (s *stubPV) CountSince(ctx context.Context, since time.Time) (int64, error)    { return 0, nil }

func TestGetSummaryUsesShortTTLCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pv := &stubPV{}
	c := cache.New(time.Minute)
	defer c.Stop()
	h := NewHandler(pv).WithCache(c)

	r := gin.New()
	r.GET("/admin/analytics/summary", h.GetSummary)

	// Miss
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "MISS", w.Header().Get("X-Cache"))
	require.Equal(t, 1, pv.calls)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	totals := body["totals"].(map[string]interface{})
	require.EqualValues(t, 3, totals["today"])

	// Hit
	req2 := httptest.NewRequest(http.MethodGet, "/admin/analytics/summary", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)
	require.Equal(t, "HIT", w2.Header().Get("X-Cache"))
	require.Equal(t, 1, pv.calls, "cached response must not re-query page views")
}
