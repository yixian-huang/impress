package dashboard

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

const dashboardSummaryCacheKey = cache.PrefixDashboard + "summary"
const dashboardSummaryTTL = 20 * time.Second

// Handler serves aggregated admin dashboard stats in one round-trip.
type Handler struct {
	articleRepo repository.ArticleRepository
	pageRepo    repository.UnifiedPageRepository
	mediaRepo   repository.MediaRepository
	pvRepo      repository.PageViewRepository
	cache       *cache.Cache
}

// NewHandler creates a dashboard handler.
func NewHandler(
	articleRepo repository.ArticleRepository,
	pageRepo repository.UnifiedPageRepository,
	mediaRepo repository.MediaRepository,
	pvRepo repository.PageViewRepository,
) *Handler {
	return &Handler{
		articleRepo: articleRepo,
		pageRepo:    pageRepo,
		mediaRepo:   mediaRepo,
		pvRepo:      pvRepo,
	}
}

// WithCache enables short-TTL caching for dashboard summary.
func (h *Handler) WithCache(c *cache.Cache) *Handler {
	h.cache = c
	return h
}

// Summary handles GET /admin/dashboard/summary
// @Summary      Admin dashboard summary
// @Description  Returns today visits, pages/articles/media counts for the dashboard
// @Tags         Dashboard (Admin)
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{todayVisits=int,pagesCount=int,articlesCount=int,mediaCount=int}
// @Failure      401 {object} object{error=string}
// @Router       /admin/dashboard/summary [get]
func (h *Handler) Summary(c *gin.Context) {
	if h.cache != nil {
		if cached, ok := h.cache.Get(dashboardSummaryCacheKey); ok {
			c.Header("X-Cache", "HIT")
			c.Header("Cache-Control", "private, max-age=15")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	ctx := c.Request.Context()
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var (
		todayVisits   int64
		pagesCount    int64
		articlesCount int64
		mediaCount    int64
		mu            sync.Mutex
		errs          = map[string]bool{}
	)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		if h.pvRepo == nil {
			mu.Lock()
			errs["todayVisits"] = true
			mu.Unlock()
			return
		}
		n, err := h.pvRepo.CountSince(ctx, startOfToday)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs["todayVisits"] = true
			return
		}
		todayVisits = n
	}()

	go func() {
		defer wg.Done()
		if h.pageRepo == nil {
			mu.Lock()
			errs["pagesCount"] = true
			mu.Unlock()
			return
		}
		n, err := h.pageRepo.Count(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs["pagesCount"] = true
			return
		}
		pagesCount = n
	}()

	go func() {
		defer wg.Done()
		if h.articleRepo == nil {
			mu.Lock()
			errs["articlesCount"] = true
			mu.Unlock()
			return
		}
		n, err := h.articleRepo.Count(ctx, "")
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs["articlesCount"] = true
			return
		}
		articlesCount = n
	}()

	go func() {
		defer wg.Done()
		if h.mediaRepo == nil {
			mu.Lock()
			errs["mediaCount"] = true
			mu.Unlock()
			return
		}
		n, err := h.mediaRepo.Count(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs["mediaCount"] = true
			return
		}
		mediaCount = n
	}()

	wg.Wait()

	payload := gin.H{
		"todayVisits":   todayVisits,
		"pagesCount":    pagesCount,
		"articlesCount": articlesCount,
		"mediaCount":    mediaCount,
		"errors":        errs,
	}
	if h.cache != nil {
		h.cache.SetWithTTL(dashboardSummaryCacheKey, payload, dashboardSummaryTTL)
	}

	c.Header("X-Cache", "MISS")
	c.Header("Cache-Control", "private, max-age=15")
	c.JSON(http.StatusOK, payload)
}
