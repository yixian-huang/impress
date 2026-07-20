package theme_export

import (
	"net/http"
	"strconv"

	"github.com/yixian-huang/inkless/backend/internal/service"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/apierror"
)

type Handler struct {
	exportSvc *service.ThemeExportService
}

func NewHandler(exportSvc *service.ThemeExportService) *Handler {
	return &Handler{exportSvc: exportSvc}
}

// Export handles POST /admin/themes/export
func (h *Handler) Export(c *gin.Context) {
	name := c.DefaultQuery("name", "my-theme")
	result, err := h.exportSvc.Export(c.Request.Context(), name)
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// Import handles POST /admin/themes/import
func (h *Handler) Import(c *gin.Context) {
	var themePackage map[string]interface{}
	if err := c.ShouldBindJSON(&themePackage); err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.exportSvc.Import(c.Request.Context(), themePackage); err != nil {
		apierror.Message(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "theme imported"})
}

// List handles GET /admin/theme-packages
func (h *Handler) List(c *gin.Context) {
	themes, err := h.exportSvc.ListInstalledThemes(c.Request.Context())
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, themes)
}

// Apply handles PUT /admin/theme-packages/:id/apply
func (h *Handler) Apply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.exportSvc.ApplyTheme(c.Request.Context(), uint(id)); err != nil {
		apierror.Message(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "theme applied"})
}
