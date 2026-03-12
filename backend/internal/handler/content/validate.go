package content

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/middleware"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/apierror"
	"blotting-consultancy/pkg/metrics"
)

// ValidateRequest represents the request for POST /admin/content/{pageKey}/validate
type ValidateRequest struct {
	Config model.JSONMap `json:"config" binding:"required"`
}

// ValidateResponse represents the response for POST /admin/content/{pageKey}/validate
type ValidateResponse struct {
	Valid             bool                               `json:"valid"`
	Errors            []service.ValidationError          `json:"errors"`
	TranslationStatus map[string]service.TranslationState `json:"translationStatus"`
}

// Validate handles POST /admin/content/{pageKey}/validate
// @Summary      Validate content config
// @Description  Validates a page config and returns translation status without saving
// @Tags         Content (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey path string          true "Page key"
// @Param        body    body ValidateRequest  true "Config to validate"
// @Success      200 {object} ValidateResponse
// @Failure      400 {object} object{error=string}
// @Router       /admin/content/{pageKey}/validate [post]
func (h *Handler) Validate(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Record metrics attempt
	metrics.Global().RecordValidationAttempt()

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Parse request body
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid request body"))
		return
	}

	// Extract user context for audit logging
	user := middleware.GetUserContext(c)
	actorName := "unknown"
	if user != nil {
		actorName = user.Username
	}

	// Validate config
	result := h.validationSvc.ValidateConfig(pageKey, req.Config)

	// Record metrics and audit log
	if !result.Valid {
		metrics.Global().RecordValidationFailure()
	}

	// Count translation issues (missing/stale states)
	translationIssueCount := 0
	for _, state := range result.TranslationStatus {
		if state == service.TranslationStateMissing || state == service.TranslationStateStale {
			translationIssueCount++
		}
	}
	h.auditLog.LogValidation(pageKeyStr, actorName, result.Valid, len(result.Errors), translationIssueCount)

	// Return validation result
	response := ValidateResponse{
		Valid:             result.Valid,
		Errors:            result.Errors,
		TranslationStatus: result.TranslationStatus,
	}

	c.JSON(http.StatusOK, response)
}
