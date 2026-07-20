package analytics

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

const analyticsSummaryCacheKey = cache.PrefixAnalytics + "summary"
const analyticsSummaryTTL = 45 * time.Second

// Handler handles analytics-related HTTP requests
type Handler struct {
	pvRepo repository.PageViewRepository
	cache  *cache.Cache
}

// NewHandler creates a new analytics handler
func NewHandler(pvRepo repository.PageViewRepository) *Handler {
	return &Handler{pvRepo: pvRepo}
}

// WithCache enables short-TTL caching for expensive summary queries.
func (h *Handler) WithCache(c *cache.Cache) *Handler {
	h.cache = c
	return h
}

// GetSummary handles GET /admin/analytics/summary
// @Summary      Get analytics summary
// @Description  Returns page view statistics including today, last 7 days, last 30 days, and unique visitors
// @Tags         Analytics (Admin)
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{pages=[]object,totals=object}
// @Failure      500 {object} object{error=string}
// @Router       /admin/analytics/summary [get]
func (h *Handler) GetSummary(c *gin.Context) {
	if h.cache != nil {
		if cached, ok := h.cache.Get(analyticsSummaryCacheKey); ok {
			c.Header("X-Cache", "HIT")
			c.Header("Cache-Control", "private, max-age=30")
			c.JSON(200, cached)
			return
		}
	}

	now := time.Now()
	stats, err := h.pvRepo.GetSummary(c.Request.Context(), now)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch analytics"})
		return
	}

	// Calculate totals
	var totalToday, totalLast7d, totalLast30d, totalUV int64
	for _, s := range stats {
		totalToday += s.Today
		totalLast7d += s.Last7d
		totalLast30d += s.Last30d
		totalUV += s.UniqueVisitors
	}

	payload := gin.H{
		"pages": stats,
		"totals": gin.H{
			"today":          totalToday,
			"last7d":         totalLast7d,
			"last30d":        totalLast30d,
			"uniqueVisitors": totalUV,
		},
	}

	if h.cache != nil {
		h.cache.SetWithTTL(analyticsSummaryCacheKey, payload, analyticsSummaryTTL)
	}

	c.Header("X-Cache", "MISS")
	c.Header("Cache-Control", "private, max-age=30")
	c.JSON(200, payload)
}
