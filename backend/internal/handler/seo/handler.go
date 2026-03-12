package seo

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SiteSetting stores key-value site settings in the database.
type SiteSetting struct {
	Key       string    `gorm:"primaryKey;size:100"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	h := &Handler{db: db}
	// Ensure the site_settings table exists
	_ = db.AutoMigrate(&SiteSetting{})
	return h
}

func defaultRobotsTxt() string {
	return "User-agent: *\nAllow: /\n\nSitemap: /sitemap.xml\n"
}

func (h *Handler) getRobotsTxt() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var setting SiteSetting
	if err := h.db.WithContext(ctx).Where("`key` = ?", "robots_txt").First(&setting).Error; err != nil {
		return defaultRobotsTxt()
	}
	return setting.Value
}

// GetRobotsTxt returns the robots.txt content.
// @Summary      Get robots.txt
// @Description  Returns the current robots.txt content
// @Tags         SEO
// @Produce      text/plain
// @Success      200 {string} string
// @Router       /public/robots.txt [get]
func (h *Handler) GetRobotsTxt(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(h.getRobotsTxt()))
}

// AdminGetRobotsTxt returns robots.txt for editing.
// @Summary      Get robots.txt (admin)
// @Description  Returns the robots.txt content for admin editing
// @Tags         SEO
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{content=string}
// @Router       /admin/seo/robots [get]
func (h *Handler) AdminGetRobotsTxt(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"content": h.getRobotsTxt()})
}

// AdminUpdateRobotsTxt updates the robots.txt content.
// @Summary      Update robots.txt
// @Description  Update the robots.txt content
// @Tags         SEO
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body object true "Robots.txt content"
// @Success      200 {object} object{content=string}
// @Router       /admin/seo/robots [put]
func (h *Handler) AdminUpdateRobotsTxt(c *gin.Context) {
	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	setting := SiteSetting{Key: "robots_txt", Value: input.Content}
	result := h.db.Where("`key` = ?", "robots_txt").Assign(SiteSetting{Value: input.Content}).FirstOrCreate(&setting)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
		return
	}
	// Update if already existed
	if result.RowsAffected == 0 {
		h.db.Model(&setting).Update("value", input.Content)
	}

	c.JSON(http.StatusOK, gin.H{"content": input.Content})
}

func (h *Handler) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/robots.txt", h.GetRobotsTxt)
	admin.GET("/seo/robots", h.AdminGetRobotsTxt)
	admin.PUT("/seo/robots", h.AdminUpdateRobotsTxt)
}
