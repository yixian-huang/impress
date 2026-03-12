package content

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/pkg/apierror"
)

// VersionListItem represents a single version in the list response
type VersionListItem struct {
	Version    int    `json:"version"`
	ChangeNote string `json:"changeNote"`
	Operator   string `json:"operator"`
	CreatedAt  string `json:"createdAt"`
}

// GetVersionsResponse represents the response for GET /admin/content/{pageKey}/versions
type GetVersionsResponse struct {
	Items []VersionListItem `json:"items"`
	Total int64             `json:"total"`
}

// GetVersions handles GET /admin/content/{pageKey}/versions
// @Summary      List content versions
// @Description  Returns a paginated list of published versions for a page key
// @Tags         Content (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        pageKey  path  string true  "Page key"
// @Param        page     query int    false "Page number"    default(1)
// @Param        pageSize query int    false "Items per page" default(20)
// @Success      200 {object} GetVersionsResponse
// @Failure      400 {object} object{error=string}
// @Router       /admin/content/{pageKey}/versions [get]
func (h *Handler) GetVersions(c *gin.Context) {
	pageKeyStr := c.Param("pageKey")
	pageKey := model.PageKey(pageKeyStr)

	// Validate page key
	if !isValidPageKey(pageKey) {
		c.JSON(http.StatusBadRequest, apierror.BadRequest("Invalid page key"))
		return
	}

	// Parse pagination parameters with defaults
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch versions with pagination
	versions, total, err := h.versionRepo.ListByPageKey(c.Request.Context(), pageKey, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apierror.InternalServerError("Failed to fetch versions"))
		return
	}

	// Build response
	items := make([]VersionListItem, len(versions))
	for i, v := range versions {
		items[i] = VersionListItem{
			Version:    v.Version,
			ChangeNote: "", // ChangeNote not in model yet, leaving empty
			Operator:   "", // Operator username not in model, leaving empty
			CreatedAt:  v.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := GetVersionsResponse{
		Items: items,
		Total: total,
	}

	c.JSON(http.StatusOK, response)
}
