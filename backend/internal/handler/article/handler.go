package article

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/eventbus"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// Handler handles article-related HTTP requests
type Handler struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
	eventBus     eventbus.EventBus
}

// NewHandler creates a new article handler
func NewHandler(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	eventBus eventbus.EventBus,
) *Handler {
	return &Handler{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		eventBus:     eventBus,
	}
}

// --- Public endpoints ---

// PublicList returns a paginated list of published articles
// GET /public/articles?page=1&pageSize=10&category=&tag=
func (h *Handler) PublicList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	categorySlug := c.Query("category")
	tagSlug := c.Query("tag")
	offset := (page - 1) * pageSize

	items, total, err := h.articleRepo.ListPublished(c.Request.Context(), offset, pageSize, categorySlug, tagSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "查询文章失败"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// PublicGetBySlug returns a single published article by slug
// GET /public/articles/:slug
func (h *Handler) PublicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.articleRepo.FindBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	if article.Status != model.ArticleStatusPublished {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	c.JSON(http.StatusOK, article)
}

// --- Admin endpoints ---

// AdminList returns a paginated list of articles (all statuses)
// GET /admin/articles?page=1&pageSize=10&status=
func (h *Handler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	status := c.Query("status")
	offset := (page - 1) * pageSize

	items, total, err := h.articleRepo.List(c.Request.Context(), offset, pageSize, status, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "查询文章失败"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AdminGetByID returns a single article by ID
// GET /admin/articles/:id
func (h *Handler) AdminGetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的 ID"}})
		return
	}

	article, err := h.articleRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	c.JSON(http.StatusOK, article)
}

// createUpdateInput is the JSON body for creating/updating articles
type createUpdateInput struct {
	Slug              string  `json:"slug"`
	Status            string  `json:"status"`
	ZhTitle           string  `json:"zhTitle"`
	EnTitle           string  `json:"enTitle"`
	ZhBody            string  `json:"zhBody"`
	EnBody            string  `json:"enBody"`
	CoverImage        string  `json:"coverImage"`
	ZhSeoTitle        string  `json:"zhSeoTitle"`
	EnSeoTitle        string  `json:"enSeoTitle"`
	ZhMetaDescription string  `json:"zhMetaDescription"`
	EnMetaDescription string  `json:"enMetaDescription"`
	OgImage           string  `json:"ogImage"`
	CategoryID        *uint   `json:"categoryId"`
	TagIDs            []uint  `json:"tagIds"`
}

// AdminCreate creates a new article
// POST /admin/articles
func (h *Handler) AdminCreate(c *gin.Context) {
	var input createUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的请求数据"}})
		return
	}

	status := model.ArticleStatus(input.Status)
	if status == "" {
		status = model.ArticleStatusDraft
	}

	article := &model.Article{
		Slug:              input.Slug,
		Status:            status,
		ZhTitle:           input.ZhTitle,
		EnTitle:           input.EnTitle,
		ZhBody:            input.ZhBody,
		EnBody:            input.EnBody,
		CoverImage:        input.CoverImage,
		ZhSeoTitle:        input.ZhSeoTitle,
		EnSeoTitle:        input.EnSeoTitle,
		ZhMetaDescription: input.ZhMetaDescription,
		EnMetaDescription: input.EnMetaDescription,
		OgImage:           input.OgImage,
		CategoryID:        input.CategoryID,
	}

	if status == model.ArticleStatusPublished {
		now := time.Now()
		article.PublishedAt = &now
	}

	// Resolve tags
	if len(input.TagIDs) > 0 {
		tags, err := h.resolveTagIDs(c, input.TagIDs)
		if err != nil {
			return // error already written to response
		}
		article.Tags = tags
	}

	if err := h.articleRepo.Create(c.Request.Context(), article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	// Re-fetch with preloads
	created, err := h.articleRepo.FindByID(c.Request.Context(), article.ID)
	if err != nil {
		c.JSON(http.StatusCreated, article)
		return
	}

	// Publish content created event
	if h.eventBus != nil {
		h.eventBus.Publish(eventbus.Event{
			Type: eventbus.ContentCreated,
			Payload: eventbus.ContentEventPayload{
				ContentType: "article",
				ContentID:   article.ID,
				Slug:        article.Slug,
				Action:      eventbus.ContentCreated,
			},
		})
	}

	c.JSON(http.StatusCreated, created)
}

// AdminUpdate updates an existing article
// PUT /admin/articles/:id
func (h *Handler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的 ID"}})
		return
	}

	existing, err := h.articleRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	var input createUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的请求数据"}})
		return
	}

	// Update fields
	existing.Slug = input.Slug
	existing.ZhTitle = input.ZhTitle
	existing.EnTitle = input.EnTitle
	existing.ZhBody = input.ZhBody
	existing.EnBody = input.EnBody
	existing.CoverImage = input.CoverImage
	existing.ZhSeoTitle = input.ZhSeoTitle
	existing.EnSeoTitle = input.EnSeoTitle
	existing.ZhMetaDescription = input.ZhMetaDescription
	existing.EnMetaDescription = input.EnMetaDescription
	existing.OgImage = input.OgImage
	existing.CategoryID = input.CategoryID

	if input.Status != "" {
		newStatus := model.ArticleStatus(input.Status)
		// Set publishedAt when transitioning to published
		if newStatus == model.ArticleStatusPublished && existing.Status != model.ArticleStatusPublished {
			now := time.Now()
			existing.PublishedAt = &now
		}
		existing.Status = newStatus
	}

	// Resolve tags
	if input.TagIDs != nil {
		tags, err := h.resolveTagIDs(c, input.TagIDs)
		if err != nil {
			return
		}
		existing.Tags = tags
	}

	if err := h.articleRepo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	// Re-fetch with preloads
	updated, err := h.articleRepo.FindByID(c.Request.Context(), existing.ID)
	if err != nil {
		c.JSON(http.StatusOK, existing)
		return
	}

	// Publish content updated event
	if h.eventBus != nil {
		h.eventBus.Publish(eventbus.Event{
			Type: eventbus.ContentUpdated,
			Payload: eventbus.ContentEventPayload{
				ContentType: "article",
				ContentID:   existing.ID,
				Slug:        existing.Slug,
				Action:      eventbus.ContentUpdated,
			},
		})
	}

	c.JSON(http.StatusOK, updated)
}

// AdminDelete deletes an article
// DELETE /admin/articles/:id
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的 ID"}})
		return
	}

	if err := h.articleRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	// Publish content deleted event
	if h.eventBus != nil {
		h.eventBus.Publish(eventbus.Event{
			Type: eventbus.ContentDeleted,
			Payload: eventbus.ContentEventPayload{
				ContentType: "article",
				ContentID:   uint(id),
				Action:      eventbus.ContentDeleted,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// resolveTagIDs looks up tags by their IDs and returns them
func (h *Handler) resolveTagIDs(c *gin.Context, tagIDs []uint) ([]model.Tag, error) {
	var tags []model.Tag
	for _, tagID := range tagIDs {
		t, err := h.tagRepo.FindByID(c.Request.Context(), tagID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "标签 ID " + strconv.FormatUint(uint64(tagID), 10) + " 不存在"}})
			return nil, err
		}
		tags = append(tags, *t)
	}
	return tags, nil
}
