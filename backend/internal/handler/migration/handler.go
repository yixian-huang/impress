package migration

import (
	"context"
	"encoding/json"
	"errors"
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
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"message": "no articles found in export file",
			},
			"parseErrors": result.Errors,
		})
		return
	}

	// Use a background context so the import continues after the HTTP response
	importCtx := context.Background()
	jobID := h.service.StartImport(importCtx, source, result.Articles, result.Errors)

	c.JSON(http.StatusAccepted, gin.H{
		"jobId":         jobID,
		"source":        sourceStr,
		"totalArticles": len(result.Articles),
		"parseErrors":   result.Errors,
		"message":       "import started; poll GET /admin/migration/jobs/:jobId for progress",
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

// RetryJob handles POST /admin/migration/jobs/:jobId/retry
// Retries failed article imports for a failed migration job.
func (h *Handler) RetryJob(c *gin.Context) {
	jobID := c.Param("jobId")

	progress, err := h.service.RetryJob(context.Background(), jobID)
	if err != nil {
		status := http.StatusBadRequest
		message := err.Error()
		switch {
		case errors.Is(err, migrationPkg.ErrJobNotFound):
			status = http.StatusNotFound
			message = "migration job not found"
		case errors.Is(err, migrationPkg.ErrJobRunning):
			status = http.StatusConflict
			message = "migration job is still running"
		case errors.Is(err, migrationPkg.ErrJobNotFailed):
			status = http.StatusConflict
			message = "only failed migration jobs can be retried"
		case errors.Is(err, migrationPkg.ErrJobNotRetryable):
			status = http.StatusConflict
			message = "migration job has no failed article imports to retry"
		}
		c.JSON(status, gin.H{
			"error": gin.H{"message": message},
		})
		return
	}

	c.JSON(http.StatusAccepted, progress)
}

// StreamProgress handles GET /admin/migration/jobs/:jobId/stream
// Provides Server-Sent Events (SSE) for real-time progress monitoring.
func (h *Handler) StreamProgress(c *gin.Context) {
	jobID := c.Param("jobId")

	// Verify job exists
	progress, ok := h.service.GetProgress(jobID)
	if !ok {
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

	if !writeSSE(c, "progress", progress) || isTerminal(progress.Phase) {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	heartbeat := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	defer heartbeat.Stop()

	lastProgress := progress
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress, ok := h.service.GetProgress(jobID)
			if !ok {
				return
			}

			if progressChanged(lastProgress, progress) {
				lastProgress = progress
				if !writeSSE(c, "progress", progress) {
					return
				}
				if isTerminal(progress.Phase) {
					return
				}
			}
		case <-heartbeat.C:
			fmt.Fprint(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		}
	}
}

func writeSSE(c *gin.Context, event string, progress *provider.MigrationProgress) bool {
	data, err := json.Marshal(progress)
	if err != nil {
		return false
	}
	if event != "" {
		fmt.Fprintf(c.Writer, "event: %s\n", event)
	}
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
	return true
}

func progressChanged(previous, current *provider.MigrationProgress) bool {
	if previous == nil || current == nil {
		return previous != current
	}
	if previous.Phase != current.Phase ||
		previous.Processed != current.Processed ||
		previous.Succeeded != current.Succeeded ||
		previous.Failed != current.Failed ||
		previous.Attempt != current.Attempt ||
		previous.Retryable != current.Retryable {
		return true
	}
	if (previous.FinishedAt == nil) != (current.FinishedAt == nil) {
		return true
	}
	return len(previous.Errors) != len(current.Errors)
}

func isTerminal(phase string) bool {
	return phase == "done" || phase == "failed"
}
