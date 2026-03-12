package content

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/pkg/apierror"
)

// GetVersionDetailResponse represents the response for GET /admin/content/{pageKey}/versions/{version}
type GetVersionDetailResponse struct {
	ID          uint          `json:"id"`
	PageKey     string        `json:"pageKey"`
	Version     int           `json:"version"`
	Config      model.JSONMap `json:"config"`
	PublishedAt time.Time     `json:"publishedAt"`
	CreatedBy   uint          `json:"createdBy"`
}

// GetVersionDetail handles GET /admin/content/{pageKey}/versions/{version}
// @Summary      Get version detail
// @Description  Returns the config snapshot for a specific published version
// @Tags         Content (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey path string true "Page key"
// @Param        version path int    true "Version number"
// @Success      200 {object} GetVersionDetailResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /admin/content/{pageKey}/versions/{version} [get]
func (h *Handler) GetVersionDetail(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Parse version parameter
	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid version parameter"))
		return
	}

	// Fetch version detail
	versionRecord, err := h.versionRepo.FindByPageKeyAndVersion(c.Request.Context(), pageKey, version)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, apierror.NotFound("Version not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to fetch version detail"))
		return
	}

	// Build response
	response := GetVersionDetailResponse{
		ID:          versionRecord.ID,
		PageKey:     string(versionRecord.PageKey),
		Version:     versionRecord.Version,
		Config:      versionRecord.Config,
		PublishedAt: versionRecord.PublishedAt,
		CreatedBy:   versionRecord.CreatedBy,
	}

	c.JSON(http.StatusOK, response)
}
