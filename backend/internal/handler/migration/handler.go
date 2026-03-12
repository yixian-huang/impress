package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	migrationPkg "blotting-consultancy/internal/migration"
	"blotting-consultancy/internal/provider"
)

// Handler handles data migration HTTP requests.
type Handler struct {
	service   *migrationPkg.Service
	providers map[provider.MigrationSource]provider.MigrationProvider
}

// NewHandler creates a new migration handler with all registered providers.
func NewHandler(service *migrationPkg.Service) *Handler {
	h := &Handler{
		service:   service,
		providers: make(map[provider.MigrationSource]provider.MigrationProvider),
	}

	// Register built-in providers
	wp := migrationPkg.NewWordPressProvider()
	halo := migrationPkg.NewHaloProvider()
	md := migrationPkg.NewMarkdownProvider()

	h.providers[wp.Source()] = wp
	h.providers[halo.Source()] = halo
	h.providers[md.Source()] = md

	return h
}

// Import handles POST /admin/migration/import
// Accepts multipart form with:
//   - source: "wordpress" | "halo" | "markdown"
//   - file: the export file (WXR XML, JSON, or ZIP of .md files)
func (h *Handler) Import(c *gin.Context) {
	sourceStr := c.PostForm("source")
	source := provider.MigrationSource(sourceStr)

	p, ok := h.providers[source]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("unsupported migration source: %q (supported: wordpress, halo, markdown)", sourceStr),
			},
		})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "file is required"},
		})
		return
	}
	defer file.Close()

	// Parse the source data synchronously (typically fast)
	parseCtx, parseCancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer parseCancel()

	result, err := p.Parse(parseCtx, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("failed to parse %s export: %v", sourceStr, err),
			},
		})
		return
	}

	if len(result.Articles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":     "no articles found in export file",
			"parseErrors": result.Errors,
		})
		return
	}

	// Generate job ID and start async import
	jobID := fmt.Sprintf("mig-%s-%d", sourceStr, time.Now().UnixMilli())

	// Use a background context so the import continues after the HTTP response
	importCtx := context.Background()
	h.service.ImportArticles(importCtx, jobID, source, result.Articles, result.Errors)

	c.JSON(http.StatusAccepted, gin.H{
		"jobId":        jobID,
		"source":       sourceStr,
		"totalArticles": len(result.Articles),
		"parseErrors":  result.Errors,
		"message":      "import started; poll GET /admin/migration/jobs/:jobId for progress",
	})
}

// GetJob handles GET /admin/migration/jobs/:jobId
// Returns the current progress of a migration job.
func (h *Handler) GetJob(c *gin.Context) {
	jobID := c.Param("jobId")

	progress, ok := h.service.GetProgress(jobID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "migration job not found"},
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// ListJobs handles GET /admin/migration/jobs
// Returns all known migration jobs.
func (h *Handler) ListJobs(c *gin.Context) {
	jobs := h.service.ListJobs()
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}

// StreamProgress handles GET /admin/migration/jobs/:jobId/stream
// Provides Server-Sent Events (SSE) for real-time progress monitoring.
func (h *Handler) StreamProgress(c *gin.Context) {
	jobID := c.Param("jobId")

	// Verify job exists
	if _, ok := h.service.GetProgress(jobID); !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "migration job not found"},
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()

	// Poll and stream progress updates
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastProcessed int
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress, ok := h.service.GetProgress(jobID)
			if !ok {
				return
			}

			// Only send update if something changed
			if progress.Processed != lastProcessed || progress.Phase == "done" || progress.Phase == "failed" {
				lastProcessed = progress.Processed

				data, _ := json.Marshal(progress)
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				c.Writer.Flush()

				// Stop streaming once complete
				if progress.Phase == "done" || progress.Phase == "failed" {
					return
				}
			}
		}
	}
}
