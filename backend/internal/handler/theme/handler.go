package theme

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

// Handler handles theme-related HTTP requests
type Handler struct {
	siteConfigRepo repository.SiteConfigRepository
	cache          *cache.Cache
}

// NewHandler creates a new theme handler
func NewHandler(siteConfigRepo repository.SiteConfigRepository, c *cache.Cache) *Handler {
	return &Handler{siteConfigRepo: siteConfigRepo, cache: c}
}

func (h *Handler) invalidateBootstrapCache() {
	if h.cache != nil {
		h.cache.DeletePrefix("bootstrap:")
	}
}

// defaultThemeConfig returns the default theme token values
func defaultThemeConfig() model.JSONMap {
	return model.JSONMap{
		"colors": map[string]interface{}{
			"primary":     "#1a5f8f",
			"primaryDark": "#26548b",
			"accent":      "#8bc34a",
			"accentHover": "#7cb342",
			"surface":     "#ffffff",
			"surfaceAlt":  "#f9fafb",
		},
		"fonts": map[string]interface{}{
			"sans":    "system-ui, -apple-system, sans-serif",
			"heading": "system-ui, -apple-system, sans-serif",
		},
		"layout": map[string]interface{}{
			"maxWidth":       "1200px",
			"borderRadius":   "0.5rem",
			"contentPadding": "1.5rem",
			"sectionSpacing": "5rem",
			"contentGap":     "2rem",
		},
	}
}

// PublicGet returns the active (published) theme tokens.
// @Summary      Get public theme tokens
// @Description  Returns the published theme design tokens (colors, fonts, layout)
// @Tags         Theme Tokens
// @Produce      json
// @Success      200 {object} object
// @Router       /public/theme [get]
func (h *Handler) PublicGet(c *gin.Context) {
	doc, err := h.siteConfigRepo.FindByKey(c.Request.Context(), "theme")
	if err != nil {
		// Return default theme if not configured
		c.JSON(http.StatusOK, defaultThemeConfig())
		return
	}

	// Return published config; fall back to default if empty
	config := doc.PublishedConfig
	if len(config) == 0 {
		config = defaultThemeConfig()
	}

	c.JSON(http.StatusOK, config)
}

// AdminGet returns the theme settings for editing.
// @Summary      Get theme settings (admin)
// @Description  Returns draft and published theme token configurations
// @Tags         Theme Tokens
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object
// @Router       /admin/theme [get]
func (h *Handler) AdminGet(c *gin.Context) {
	doc, err := h.siteConfigRepo.FindByKey(c.Request.Context(), "theme")
	if err != nil {
		// Return default with version info
		c.JSON(http.StatusOK, gin.H{
			"draftConfig":      defaultThemeConfig(),
			"draftVersion":     0,
			"publishedConfig":  defaultThemeConfig(),
			"publishedVersion": 0,
		})
		return
	}

	draftConfig := doc.DraftConfig
	if len(draftConfig) == 0 {
		draftConfig = defaultThemeConfig()
	}

	publishedConfig := doc.PublishedConfig
	if len(publishedConfig) == 0 {
		publishedConfig = defaultThemeConfig()
	}

	c.JSON(http.StatusOK, gin.H{
		"draftConfig":      draftConfig,
		"draftVersion":     doc.DraftVersion,
		"publishedConfig":  publishedConfig,
		"publishedVersion": doc.PublishedVersion,
	})
}

// updateInput is the JSON body for updating theme settings
type updateInput struct {
	Config       model.JSONMap `json:"config"`
	DraftVersion int           `json:"draftVersion"`
}

// AdminUpdate updates the theme settings.
// @Summary      Update theme settings
// @Description  Update theme design tokens with optimistic locking
// @Tags         Theme Tokens
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body object true "Theme token configuration"
// @Success      200 {object} object
// @Failure      409 {object} object{error=string}
// @Router       /admin/theme [put]
func (h *Handler) AdminUpdate(c *gin.Context) {
	var input updateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的请求数据"}})
		return
	}

	if input.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "config is required"}})
		return
	}

	// Try to find existing theme config
	existing, err := h.siteConfigRepo.FindByKey(c.Request.Context(), "theme")
	if err != nil {
		// Create new theme config if it doesn't exist
		sc := &model.SiteConfig{
			Key:              "theme",
			DraftConfig:      input.Config,
			DraftVersion:     1,
			PublishedConfig:  input.Config,
			PublishedVersion: 1,
		}
		if createErr := h.siteConfigRepo.Upsert(c.Request.Context(), sc); createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "保存主题设置失败"}})
			return
		}
		h.invalidateBootstrapCache()
		c.JSON(http.StatusOK, gin.H{
			"draftConfig":  sc.DraftConfig,
			"draftVersion": sc.DraftVersion,
			"message":      "主题设置已创建",
		})
		return
	}

	// Optimistic locking: check draft version matches
	if existing.DraftVersion != input.DraftVersion {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"message": "版本冲突，请刷新后重试"}})
		return
	}

	// Theme has no separate publish step — save draft and publish in one go
	existing.DraftConfig = input.Config
	existing.DraftVersion = input.DraftVersion + 1
	existing.PublishedConfig = input.Config
	existing.PublishedVersion = input.DraftVersion + 1
	if err := h.siteConfigRepo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "保存主题设置失败"}})
		return
	}

	h.invalidateBootstrapCache()

	c.JSON(http.StatusOK, gin.H{
		"draftConfig":  existing.DraftConfig,
		"draftVersion": existing.DraftVersion,
		"message":      "主题设置已更新",
	})
}
