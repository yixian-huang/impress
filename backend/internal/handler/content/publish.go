package content

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/middleware"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/apierror"
	"blotting-consultancy/pkg/metrics"
)

// PublishRequest represents the request for POST /admin/content/{pageKey}/publish
type PublishRequest struct {
	ExpectedDraftVersion int    `json:"expectedDraftVersion" binding:"required"`
	ChangeNote           string `json:"changeNote"`
}

// PublishResponse represents the response for POST /admin/content/{pageKey}/publish
type PublishResponse struct {
	PageKey          string    `json:"pageKey"`
	PublishedVersion int       `json:"publishedVersion"`
	PublishedAt      time.Time `json:"publishedAt"`
}

// Publish handles POST /admin/content/{pageKey}/publish
// @Summary      Publish content
// @Description  Promotes the current draft to published state with optimistic locking
// @Tags         Content (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey path string         true "Page key"
// @Param        body    body PublishRequest  true "Publish parameters"
// @Success      200 {object} PublishResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Failure      422 {object} object{error=string}
// @Router       /admin/content/{pageKey}/publish [post]
func (h *Handler) Publish(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Record metrics attempt
	metrics.Global().RecordPublishAttempt()

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Parse request body
	var req PublishRequest
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

	// Call publish service
	result, err := h.contentSvc.Publish(c.Request.Context(), pageKey, req.ExpectedDraftVersion, user.UserID)
	if err != nil {
		// Record failure metrics and audit log
		metrics.Global().RecordPublishFailure()

		if errors.Is(err, service.ErrVersionMismatch) {
			h.auditLog.LogPublishFailure(pageKeyStr, user.Username, "version_mismatch", map[string]interface{}{
				"expected_version": req.ExpectedDraftVersion,
			})
			c.JSON(http.StatusConflict, apierror.New(http.StatusConflict, "CONFLICT_VERSION", "Draft version mismatch"))
			return
		}
		if errors.Is(err, service.ErrCannotPublish) {
			h.auditLog.LogPublishFailure(pageKeyStr, user.Username, "validation_failed", nil)
			c.JSON(http.StatusUnprocessableEntity, apierror.New(http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Publish blocked by missing or stale translations"))
			return
		}
		if errors.Is(err, service.ErrDocumentNotFound) {
			h.auditLog.LogPublishFailure(pageKeyStr, user.Username, "not_found", nil)
			c.JSON(http.StatusNotFound, apierror.NotFound("Content document not found"))
			return
		}
		h.auditLog.LogPublishFailure(pageKeyStr, user.Username, "internal_error", nil)
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to publish content"))
		return
	}

	// Record success metrics and audit log
	metrics.Global().RecordPublishSuccess()
	h.auditLog.LogPublishSuccess(pageKeyStr, result.PublishedVersion, user.Username, req.ExpectedDraftVersion)

	// Return publish result
	response := PublishResponse{
		PageKey:          string(result.PageKey),
		PublishedVersion: result.PublishedVersion,
		PublishedAt:      result.PublishedAt,
	}

	c.JSON(http.StatusOK, response)
}
