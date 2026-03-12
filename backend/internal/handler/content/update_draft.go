package content

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/pkg/apierror"
)

// UpdateDraftRequest represents the request for PUT /admin/content/{pageKey}/draft
type UpdateDraftRequest struct {
	Config     model.JSONMap `json:"config" binding:"required"`
	ChangeNote string        `json:"changeNote"`
}

// UpdateDraftResponse represents the response for PUT /admin/content/{pageKey}/draft
type UpdateDraftResponse struct {
	PageKey   string    `json:"pageKey"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UpdateDraft handles PUT /admin/content/{pageKey}/draft
// @Summary      Update draft content
// @Description  Updates the draft config for a page key with optimistic locking via If-Match header
// @Tags         Content (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey  path   string            true "Page key"
// @Param        If-Match header int               true "Expected draft version for optimistic locking"
// @Param        body     body   UpdateDraftRequest true "Draft content data"
// @Success      200 {object} UpdateDraftResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Router       /admin/content/{pageKey}/draft [put]
func (h *Handler) UpdateDraft(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Extract If-Match header for optimistic locking
	ifMatchHeader := c.GetHeader("If-Match")
	if ifMatchHeader == "" {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("If-Match header is required"))
		return
	}

	expectedVersion, err := strconv.Atoi(ifMatchHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("If-Match header must be a valid integer"))
		return
	}

	// Parse request body
	var req UpdateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid request body"))
		return
	}

	// Update draft with optimistic locking
	newVersion, err := h.docRepo.UpdateDraft(c.Request.Context(), pageKey, expectedVersion, req.Config)
	if err != nil {
		// Check for version conflict error
		if err.Error() == "draft version conflict or document not found" {
			c.JSON(http.StatusConflict, apierror.New(http.StatusConflict, "CONFLICT_VERSION", "Draft version conflict"))
			return
		}
		if err.Error() == "content document not found" {
			c.JSON(http.StatusNotFound, apierror.NotFound("Content document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to update draft"))
		return
	}

	// Return updated document info
	response := UpdateDraftResponse{
		PageKey:   string(pageKey),
		Version:   newVersion,
		UpdatedAt: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
