package public

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/apierror"
	"blotting-consultancy/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// Handler handles public content-related HTTP requests
type Handler struct {
	docRepo  repository.ContentDocumentRepository
	pvRepo   repository.PageViewRepository
	pageRepo repository.UnifiedPageRepository
}

// NewHandler creates a new public content handler
func NewHandler(
	docRepo repository.ContentDocumentRepository,
	pvRepo repository.PageViewRepository,
	pageRepo repository.UnifiedPageRepository,
) *Handler {
	return &Handler{
		docRepo:  docRepo,
		pvRepo:   pvRepo,
		pageRepo: pageRepo,
	}
}

// GetPublicContent handles GET /public/content/{pageKey}?locale=zh|en
// Returns published-only content with locale support.
// Reads from unified_pages first (new system); falls back to content_documents (legacy).
func (h *Handler) GetPublicContent(c *gin.Context) {
	// Record metrics attempt and start timer
	metrics.Global().RecordPublicGetAttempt()
	startTime := time.Now()

	// Parse page key
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	if !pageKey.IsValid() {
		metrics.Global().RecordPublicGetFailure()
		c.JSON(400, apierror.BadRequest("invalid pageKey"))
		return
	}

	// Parse locale parameter (default to zh)
	locale := c.DefaultQuery("locale", "zh")
	if locale != "zh" && locale != "en" {
		metrics.Global().RecordPublicGetFailure()
		c.JSON(400, apierror.BadRequest("locale must be zh or en"))
		return
	}

	// Try unified_pages first (slug == pageKey for the 7 builtin pages)
	if h.pageRepo != nil {
		page, err := h.pageRepo.FindBySlug(c.Request.Context(), pageKeyStr)
		if err == nil && len(page.PublishedConfig) > 0 {
			// Convert sections-based config back to flat content doc format
			publishedMap := model.JSONMap(page.PublishedConfig)
			flatConfig := service.ConvertSectionsToContentDoc(pageKeyStr, publishedMap)

			latency := time.Since(startTime)
			metrics.Global().RecordPublicGetSuccess(latency)

			h.recordPageViewAsync(pageKeyStr, locale, c)

			c.JSON(200, gin.H{
				"pageKey": pageKeyStr,
				"version": page.PublishedVersion,
				"locale":  locale,
				"config":  flatConfig,
			})
			return
		}
	}

	// Fallback: read from legacy content_documents
	doc, err := h.docRepo.FindByPageKey(c.Request.Context(), pageKey)
	if err != nil {
		metrics.Global().RecordPublicGetFailure()
		c.JSON(404, apierror.NotFound("page not found"))
		return
	}

	// Record success with latency
	latency := time.Since(startTime)
	metrics.Global().RecordPublicGetSuccess(latency)

	h.recordPageViewAsync(pageKeyStr, locale, c)

	// Return published-only data (never expose draft fields)
	c.JSON(200, gin.H{
		"pageKey": doc.PageKey.String(),
		"version": doc.PublishedVersion,
		"locale":  locale,
		"config":  doc.PublishedConfig,
	})
}

// recordPageViewAsync records a page view in a background goroutine.
func (h *Handler) recordPageViewAsync(pageKey, locale string, c *gin.Context) {
	clientIP := c.ClientIP()
	referer := c.GetHeader("Referer")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hash := sha256.Sum256([]byte(clientIP))
		visitorID := fmt.Sprintf("%x", hash[:])[:16]
		if err := h.pvRepo.Create(ctx, &model.PageView{
			PageKey:   pageKey,
			Locale:    locale,
			VisitorID: visitorID,
			Referer:   referer,
		}); err != nil {
			slog.Error("failed to record page view", "pageKey", pageKey, "error", err)
		}
	}()
}
