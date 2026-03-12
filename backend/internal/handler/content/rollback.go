package content

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/middleware"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/apierror"
	"blotting-consultancy/pkg/metrics"
)

// RollbackRequest represents the request for POST /admin/content/{pageKey}/rollback/{version}
type RollbackRequest struct {
	ChangeNote string `json:"changeNote"`
}

// RollbackResponse represents the response for POST /admin/content/{pageKey}/rollback/{version}
type RollbackResponse struct {
	PageKey          string    `json:"pageKey"`
	PublishedVersion int       `json:"publishedVersion"`
	SourceVersion    int       `json:"sourceVersion"`
	PublishedAt      time.Time `json:"publishedAt"`
}

// Rollback handles POST /admin/content/{pageKey}/rollback/{version}
// @Summary      Rollback content
// @Description  Rolls back published content to a previous version
// @Tags         Content (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey path string          true "Page key"
// @Param        version path int             true "Source version to rollback to"
// @Param        body    body RollbackRequest  true "Rollback parameters"
// @Success      200 {object} RollbackResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /admin/content/{pageKey}/rollback/{version} [post]
func (h *Handler) Rollback(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Record metrics attempt
	metrics.Global().RecordRollbackAttempt()
	startTime := time.Now()

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Parse source version parameter
	versionStr := c.Param("version")
	sourceVersion, err := strconv.Atoi(versionStr)
	if err != nil || sourceVersion <= 0 {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid version parameter"))
		return
	}

	// Parse request body
	var req RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid request body"))
		return
	}

	// Extract user context
	user := middleware.GetUserContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, apierror.Unauthorized("User context not found"))
		return
	}

	// Call rollback service
	result, err := h.contentSvc.Rollback(c.Request.Context(), pageKey, sourceVersion, user.UserID)
	if err != nil {
		// Record failure metrics and audit log
		metrics.Global().RecordRollbackFailure()

		if errors.Is(err, service.ErrVersionNotFound) {
			h.auditLog.LogRollbackFailure(pageKeyStr, user.Username, sourceVersion, "version_not_found")
			c.JSON(http.StatusNotFound, apierror.NotFound("Source version not found"))
			return
		}
		if errors.Is(err, service.ErrDocumentNotFound) {
			h.auditLog.LogRollbackFailure(pageKeyStr, user.Username, sourceVersion, "not_found")
			c.JSON(http.StatusNotFound, apierror.NotFound("Content document not found"))
			return
		}
		h.auditLog.LogRollbackFailure(pageKeyStr, user.Username, sourceVersion, "internal_error")
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to rollback content"))
		return
	}

	// Record success metrics and audit log with latency
	latency := time.Since(startTime)
	metrics.Global().RecordRollbackSuccess(latency)
	h.auditLog.LogRollbackSuccess(pageKeyStr, result.PublishedVersion, result.SourceVersion, user.Username)

	// Return rollback result
	response := RollbackResponse{
		PageKey:          string(result.PageKey),
		PublishedVersion: result.PublishedVersion,
		SourceVersion:    result.SourceVersion,
		PublishedAt:      result.PublishedAt,
	}

	c.JSON(http.StatusOK, response)
}
