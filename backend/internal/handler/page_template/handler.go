package page_template

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

// Handler handles page template HTTP requests.
type Handler struct {
	tmplRepo repository.PageTemplateRepository
}

// NewHandler creates a new page template handler.
func NewHandler(tmplRepo repository.PageTemplateRepository) *Handler {
	return &Handler{tmplRepo: tmplRepo}
}

func parseID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

// List returns all templates with optional category filter.
func (h *Handler) List(c *gin.Context) {
	category := c.Query("category")
	templates, err := h.tmplRepo.List(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": templates})
}

type createInput struct {
	Key           string        `json:"key" binding:"required"`
	NameZh        string        `json:"nameZh" binding:"required"`
	NameEn        string        `json:"nameEn"`
	DescriptionZh string        `json:"descriptionZh"`
	DescriptionEn string        `json:"descriptionEn"`
	Category      string        `json:"category"`
	Config        model.JSONMap `json:"config" binding:"required"`
	Thumbnail     string        `json:"thumbnail"`
}

// Create creates a new page template.
func (h *Handler) Create(c *gin.Context) {
	var input createInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl := &model.PageTemplate{
		Key:           input.Key,
		NameZh:        input.NameZh,
		NameEn:        input.NameEn,
		DescriptionZh: input.DescriptionZh,
		DescriptionEn: input.DescriptionEn,
		Category:      input.Category,
		Config:        input.Config,
		Thumbnail:     input.Thumbnail,
	}

	if err := h.tmplRepo.Create(c.Request.Context(), tmpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

// Update updates an existing page template. Builtin templates cannot be modified.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	existing, err := h.tmplRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	if existing.Category == model.TemplateCategoryBuiltin {
		c.JSON(http.StatusForbidden, gin.H{"error": "builtin templates cannot be modified"})
		return
	}

	var input createInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Key = input.Key
	existing.NameZh = input.NameZh
	existing.NameEn = input.NameEn
	existing.DescriptionZh = input.DescriptionZh
	existing.DescriptionEn = input.DescriptionEn
	if input.Category != "" {
		existing.Category = input.Category
	}
	existing.Config = input.Config
	existing.Thumbnail = input.Thumbnail

	if err := h.tmplRepo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// Delete deletes a page template. Builtin templates cannot be deleted.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	existing, err := h.tmplRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	if existing.Category == model.TemplateCategoryBuiltin {
		c.JSON(http.StatusForbidden, gin.H{"error": "builtin templates cannot be deleted"})
		return
	}

	if err := h.tmplRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Duplicate copies a template with "-copy" key suffix and custom category.
func (h *Handler) Duplicate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	existing, err := h.tmplRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	dup := &model.PageTemplate{
		Key:           fmt.Sprintf("%s-copy", existing.Key),
		NameZh:        existing.NameZh + "(副本)",
		NameEn:        existing.NameEn + "(Copy)",
		DescriptionZh: existing.DescriptionZh,
		DescriptionEn: existing.DescriptionEn,
		Category:      model.TemplateCategoryCustom,
		Config:        existing.Config,
		Thumbnail:     existing.Thumbnail,
	}

	if err := h.tmplRepo.Create(c.Request.Context(), dup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to duplicate template: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dup)
}
