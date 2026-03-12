package analytics

import (
	"time"

	"blotting-consultancy/internal/repository"

	"github.com/gin-gonic/gin"
)

// Handler handles analytics-related HTTP requests
type Handler struct {
	pvRepo repository.PageViewRepository
}

// NewHandler creates a new analytics handler
func NewHandler(pvRepo repository.PageViewRepository) *Handler {
	return &Handler{pvRepo: pvRepo}
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

	c.JSON(200, gin.H{
		"pages": stats,
		"totals": gin.H{
			"today":          totalToday,
			"last7d":         totalLast7d,
			"last30d":        totalLast30d,
			"uniqueVisitors": totalUV,
		},
	})
}
