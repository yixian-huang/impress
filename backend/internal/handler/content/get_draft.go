package content

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/pkg/apierror"
)

// GetDraftResponse represents the response for GET /admin/content/{pageKey}/draft
type GetDraftResponse struct {
	PageKey          string         `json:"pageKey"`
	Version          int            `json:"version"`
	Config           model.JSONMap  `json:"config"`
	PublishedVersion int            `json:"publishedVersion"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

// GetDraft handles GET /admin/content/{pageKey}/draft
// @Summary      Get draft content
// @Description  Returns the draft config for a given page key
// @Tags         Content (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey path string true "Page key (e.g. home, about)"
// @Success      200 {object} GetDraftResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /admin/content/{pageKey}/draft [get]
func (h *Handler) GetDraft(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Fetch document
	doc, err := h.docRepo.FindByPageKey(c.Request.Context(), pageKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, apierror.NotFound("Content document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to fetch draft"))
		return
	}

	// Return draft config
	response := GetDraftResponse{
		PageKey:          string(doc.PageKey),
		Version:          doc.DraftVersion,
		Config:           doc.DraftConfig,
		PublishedVersion: doc.PublishedVersion,
		UpdatedAt:        doc.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// isValidPageKey validates if the page key is one of the supported values
func isValidPageKey(pageKey model.PageKey) bool {
	switch pageKey {
	case model.PageKeyHome, model.PageKeyAbout, model.PageKeyAdvantages,
		model.PageKeyCoreServices, model.PageKeyCases, model.PageKeyExperts,
		model.PageKeyContact, model.PageKeyGlobal:
		return true
	}
	return false
}
