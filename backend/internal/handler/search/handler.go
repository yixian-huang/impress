package search

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/provider"
)

type Handler struct {
	search provider.SearchProvider
}

func NewHandler(search provider.SearchProvider) *Handler {
	return &Handler{search: search}
}

func (h *Handler) PublicSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})
		return
	}
	locale := c.DefaultQuery("locale", "zh")
	contentType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	resp, err := h.search.Search(c.Request.Context(), query, locale, contentType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) PublicSuggest(c *gin.Context) {
	prefix := c.Query("q")
	if prefix == "" {
		c.JSON(http.StatusOK, []string{})
		return
	}
	locale := c.DefaultQuery("locale", "zh")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	suggestions, err := h.search.Suggest(c.Request.Context(), prefix, locale, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggest failed"})
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

func (h *Handler) AdminRebuildIndex(c *gin.Context) {
	if err := h.search.RebuildIndex(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rebuild failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "index rebuilt successfully"})
}

func (h *Handler) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/search", h.PublicSearch)
	public.GET("/search/suggest", h.PublicSuggest)
	admin.POST("/search/rebuild", h.AdminRebuildIndex)
}
