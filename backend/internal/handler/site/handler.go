package site

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/internal/service"
)

// Handler handles site management HTTP requests
type Handler struct {
	siteSvc  *service.SiteService
	siteRepo repository.SiteRepository
}

// NewHandler creates a new site management handler
func NewHandler(siteSvc *service.SiteService, siteRepo repository.SiteRepository) *Handler {
	return &Handler{
		siteSvc:  siteSvc,
		siteRepo: siteRepo,
	}
}

// --- Admin CRUD ---

// AdminList returns all sites.
// GET /admin/sites
func (h *Handler) AdminList(c *gin.Context) {
	status := c.Query("status")
	sites, err := h.siteSvc.List(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to list sites"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": sites, "total": len(sites)})
}

// AdminGetByID returns a single site by ID.
// GET /admin/sites/:id
func (h *Handler) AdminGetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	site, err := h.siteSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "site not found"}})
		return
	}
	c.JSON(http.StatusOK, site)
}

// createUpdateInput is the JSON body for creating/updating sites
type createUpdateInput struct {
	Domain   string                 `json:"domain"`
	SubPath  string                 `json:"subPath"`
	Name     string                 `json:"name"`
	Locale   string                 `json:"locale"`
	ThemeID  string                 `json:"themeId"`
	Mode     model.SiteMode         `json:"mode"`
	Settings map[string]interface{} `json:"settings"`
	Status   model.SiteStatus       `json:"status"`
}

// AdminCreate creates a new site.
// POST /admin/sites
func (h *Handler) AdminCreate(c *gin.Context) {
	var input createUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}

	site := &model.Site{
		Domain:   input.Domain,
		SubPath:  input.SubPath,
		Name:     input.Name,
		Locale:   input.Locale,
		ThemeID:  input.ThemeID,
		Mode:     input.Mode,
		Settings: model.SiteSettings(input.Settings),
		Status:   input.Status,
	}

	if err := h.siteSvc.Create(c.Request.Context(), site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, site)
}

// AdminUpdate updates an existing site.
// PUT /admin/sites/:id
func (h *Handler) AdminUpdate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}

	existing, err := h.siteSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "site not found"}})
		return
	}

	var input createUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}

	if input.Domain != "" {
		existing.Domain = input.Domain
	}
	if input.SubPath != "" {
		existing.SubPath = input.SubPath
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Locale != "" {
		existing.Locale = input.Locale
	}
	if input.ThemeID != "" {
		existing.ThemeID = input.ThemeID
	}
	if input.Mode != "" {
		existing.Mode = input.Mode
	}
	if input.Settings != nil {
		existing.Settings = model.SiteSettings(input.Settings)
	}
	if input.Status != "" {
		existing.Status = input.Status
	}

	if err := h.siteSvc.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// AdminDelete deletes a site.
// DELETE /admin/sites/:id
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	if err := h.siteSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "site not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "site deleted"})
}

// --- Site user management ---

type assignUserInput struct {
	UserID uint `json:"userId"`
	RoleID uint `json:"roleId"`
}

// AdminAssignUser assigns a user to a site.
// POST /admin/sites/:id/users
func (h *Handler) AdminAssignUser(c *gin.Context) {
	siteID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	var input assignUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}
	if err := h.siteSvc.AssignUser(c.Request.Context(), siteID, input.UserID, input.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user assigned"})
}

// AdminUnassignUser removes a user from a site.
// DELETE /admin/sites/:id/users/:userId
func (h *Handler) AdminUnassignUser(c *gin.Context) {
	siteID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	userIDStr := c.Param("userId")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid user ID"}})
		return
	}
	if err := h.siteSvc.UnassignUser(c.Request.Context(), siteID, uint(userID64)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user unassigned"})
}

// AdminListUsers lists all user-role assignments for a site.
// GET /admin/sites/:id/users
func (h *Handler) AdminListUsers(c *gin.Context) {
	siteID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	users, err := h.siteSvc.ListUsers(c.Request.Context(), siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to list site users"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": len(users)})
}

// --- Export / Import ---

// AdminExport exports a single site's configuration as JSON.
// GET /admin/sites/:id/export
func (h *Handler) AdminExport(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid site ID"}})
		return
	}
	data, err := h.siteSvc.ExportSite(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=site-export.json")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// AdminImport imports a site from a JSON payload.
// POST /admin/sites/import
func (h *Handler) AdminImport(c *gin.Context) {
	const maxSize = 10 * 1024 * 1024 // 10 MB
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxSize+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "failed to read request body"}})
		return
	}
	if len(body) > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"message": "request body too large"}})
		return
	}

	site, err := h.siteSvc.ImportSite(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, site)
}

// parseID extracts and validates the :id URL parameter as a uint
func parseID(c *gin.Context) (uint, error) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	return uint(id64), err
}
